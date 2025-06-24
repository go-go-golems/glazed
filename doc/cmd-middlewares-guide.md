# Glazed Command Middlewares: Complete Guide

## Table of Contents
1. [The Big Picture](#the-big-picture)
2. [Mental Model & Why Use This Approach](#mental-model--why-use-this-approach)
3. [Introduction](#introduction)
4. [Understanding the Middleware System](#understanding-the-middleware-system)
5. [Core Middleware Concepts](#core-middleware-concepts)
6. [Available Middleware Types](#available-middleware-types)
7. [Middleware Execution Order](#middleware-execution-order)
8. [Practical Examples](#practical-examples)
9. [Advanced Usage Patterns](#advanced-usage-patterns)
10. [Best Practices](#best-practices)
11. [Testing Middlewares](#testing-middlewares)

## The Big Picture

Imagine you're building a CLI application that needs to handle configuration from multiple sources:
- Command-line flags (highest priority)
- Configuration files (medium priority)  
- Environment variables (lower priority)
- Default values (lowest priority)

Traditional approaches often result in scattered, repetitive code that's hard to test and maintain. You end up with functions that look like:

```go
// The traditional approach - brittle and hard to extend
func loadConfig() (*Config, error) {
    config := &Config{}
    
    // Load defaults
    config.setDefaults()
    
    // Override with environment variables
    if val := os.Getenv("APP_HOST"); val != "" {
        config.Host = val
    }
    if val := os.Getenv("APP_PORT"); val != "" {
        config.Port = parsePort(val)
    }
    
    // Override with config file
    if configFile := getConfigFile(); configFile != "" {
        fileConfig, err := loadConfigFile(configFile)
        if err != nil {
            return nil, err
        }
        config.merge(fileConfig)
    }
    
    // Override with command-line flags
    if *hostFlag != "" {
        config.Host = *hostFlag
    }
    if *portFlag != 0 {
        config.Port = *portFlag
    }
    
    return config, nil
}
```

This approach has several problems:
- **Repetitive**: Each parameter needs custom handling for each source
- **Brittle**: Adding new parameters or sources requires touching multiple places
- **Hard to test**: Configuration logic is scattered and coupled
- **No flexibility**: Priority order is hard-coded and can't be changed
- **No traceability**: Can't easily see where a value came from

## Mental Model & Why Use This Approach

### The Pipeline Mental Model

The middleware system treats parameter processing as a **pipeline** where each middleware is a **stage** that can:
1. **Transform** the parameter definitions (what parameters are available)
2. **Add values** from different sources (files, environment, CLI args)
3. **Validate** or **filter** the results

Think of it like an assembly line where each worker (middleware) does one specific job:

```
[CLI Args] → [Config Files] → [Environment] → [Defaults] → [Final Config]
    ↓            ↓               ↓              ↓
Middleware1  Middleware2   Middleware3   Middleware4
```

### Key Benefits of This Approach

#### 1. **Composability**
You can mix and match different configuration sources:
```go
// Different configurations for different environments
devMiddlewares := []Middleware{cobra, localConfig, envVars, defaults}
prodMiddlewares := []Middleware{cobra, vaultConfig, k8sSecrets, defaults}
```

#### 2. **Explicit Priority Control**
The order of middlewares determines priority - no hidden behavior:
```go
// CLI flags override everything, then config file, then environment
middlewares := []Middleware{
    ParseFromCobraCommand(cmd),  // Highest priority
    LoadFromConfigFile("app.yaml"),
    LoadFromEnvironment("APP"),
    SetDefaults(),              // Lowest priority
}
```

#### 3. **Testability**
Each middleware can be tested independently:
```go
func TestEnvironmentMiddleware(t *testing.T) {
    os.Setenv("APP_HOST", "test-host")
    // Test just the environment middleware
    middleware := LoadFromEnvironment("APP")
    // ... test logic
}
```

#### 4. **Extensibility**
Adding new configuration sources is trivial:
```go
// Add HashiCorp Vault support
middlewares = append(middlewares, LoadFromVault("secret/app"))
```

#### 5. **Traceability**
Each middleware can tag where values came from:
```go
// You can see: "host=prod-server (source: vault, step: secrets)"
middleware := LoadFromVault("secret/app", 
    WithParseStepSource("vault"),
    WithParseStepMetadata(map[string]interface{}{
        "step": "secrets",
    }))
```

#### 6. **Conditional Logic**
You can conditionally apply middlewares:
```go
var middlewares []Middleware
if isProduction {
    middlewares = append(middlewares, LoadFromVault("secret/app"))
} else {
    middlewares = append(middlewares, LoadFromFile("dev-config.yaml"))
}
```

### Why Not Just Use Viper?

Viper is great for basic config management, but the middleware system provides:
- **Fine-grained control** over priority and sources
- **Layer-based organization** (group related parameters)
- **Composable validation and transformation**
- **Better testing** through isolated middleware units
- **Custom processing logic** for complex scenarios

### Real-World Example

Consider a database connection that can be configured via:
1. Command-line flags (for quick testing)
2. Kubernetes secrets (in production)
3. Local config file (in development)
4. Environment variables (fallback)
5. Defaults (last resort)

With middlewares:
```go
middlewares := []Middleware{
    ParseFromCobraCommand(cmd),                    // --db-host flag
    LoadFromK8sSecret("db-credentials"),           // Kubernetes secret
    LoadFromFile("~/.myapp/config.yaml"),          // Local config
    LoadFromEnvironment("MYAPP"),                  // MYAPP_DB_HOST env var
    SetDefaults(),                                 // db-host: localhost
}
```

Each middleware is:
- **Single responsibility**: Does one thing well
- **Testable**: Can be tested in isolation
- **Reusable**: Can be used in different combinations
- **Traceable**: Records where values came from

## Introduction

Glazed's command middleware system provides this powerful and flexible approach to parameter processing in CLI applications. Located in [`pkg/cmds/middlewares`](file:///Users/kball/git/go-go-golems/glazed/pkg/cmds/middlewares), this system enables:
- Flexible parameter source prioritization
- Cross-cutting concerns like validation and transformation
- Modular configuration loading
- Custom parameter processing logic

## Understanding the Middleware System

### Core Types

```go
// HandlerFunc processes parameter layers
type HandlerFunc func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error

// Middleware wraps a HandlerFunc to add functionality
type Middleware func(HandlerFunc) HandlerFunc
```

### Key Components

1. **ParameterLayers**: Define what parameters are available and their configuration
2. **ParsedLayers**: Store the actual parameter values after processing
3. **Middleware**: Functions that process and transform parameters
4. **ExecuteMiddlewares**: Orchestrates the middleware chain execution

## Core Middleware Concepts

### Middleware Execution Flow

```go
// ExecuteMiddlewares processes middlewares in reverse order
func ExecuteMiddlewares(
    layers_ *layers.ParameterLayers, 
    parsedLayers *layers.ParsedLayers, 
    middlewares ...Middleware
) error
```

**Important**: Middlewares execute in **reverse order** of how they're provided:
- `[f1, f2, f3]` executes as `f1(f2(f3(handler)))`
- First middleware listed has the highest priority

### Middleware Pattern

```go
func SampleMiddleware(options ...Option) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
            // Option 1: Process before calling next (modify layers)
            // ... preprocessing logic ...
            
            err := next(layers_, parsedLayers)
            if err != nil {
                return err
            }
            
            // Option 2: Process after calling next (modify parsed values)
            // ... postprocessing logic ...
            
            return nil
        }
    }
}
```

## Available Middleware Types

### 1. Command-Line Integration

#### ParseFromCobraCommand
Parses parameters from Cobra command flags and options.

```go
middleware := middlewares.ParseFromCobraCommand(cmd, 
    parameters.WithParseStepSource("cobra"))
```

#### GatherArguments
Handles positional command-line arguments.

```go
middleware := middlewares.GatherArguments(args, 
    parameters.WithParseStepSource("arguments"))
```

### 2. Configuration Sources

#### GatherFlagsFromViper
Loads configuration from Viper (config files, environment variables).

```go
middleware := middlewares.GatherFlagsFromViper(
    parameters.WithParseStepSource("viper"))
```

#### GatherFlagsFromCustomViper
Loads configuration from custom Viper instances.

```go
// From specific config file
middleware := middlewares.GatherFlagsFromCustomViper(
    middlewares.WithConfigFile("/path/to/config.yaml"),
    middlewares.WithParseOptions(parameters.WithParseStepSource("custom-config")),
)

// From another app's config
middleware := middlewares.GatherFlagsFromCustomViper(
    middlewares.WithAppName("other-app"),
    middlewares.WithParseOptions(parameters.WithParseStepSource("other-app-config")),
)
```

### 3. Profile Management

#### GatherFlagsFromProfiles
Loads configuration from YAML profile files.

```go
middleware := middlewares.GatherFlagsFromProfiles(
    defaultProfileFile,
    profileFile,
    profileName,
    parameters.WithParseStepSource("profiles"),
)
```

#### GatherFlagsFromCustomProfiles
Loads profiles from custom sources.

```go
// From specific profile file
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "production",
    middlewares.WithProfileFile("/path/to/profiles.yaml"),
    middlewares.WithProfileRequired(true),
)

// From another app's profiles
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "shared-profile",
    middlewares.WithProfileAppName("other-app"),
)
```

### 4. Environment and Defaults

#### UpdateFromEnv
Loads parameters from environment variables.

```go
middleware := middlewares.UpdateFromEnv("APP", 
    parameters.WithParseStepSource("env"))
```

#### SetFromDefaults
Sets default values from parameter definitions.

```go
middleware := middlewares.SetFromDefaults(
    parameters.WithParseStepSource("defaults"))
```

### 5. File-Based Configuration

#### LoadParametersFromFile
Loads parameters from JSON or YAML files.

```go
middleware := middlewares.LoadParametersFromFile("config.yaml",
    parameters.WithParseStepSource("file"))
```

### 6. Direct Value Updates

#### UpdateFromMap
Updates parameters from a map structure.

```go
values := map[string]map[string]interface{}{
    "default": {
        "param1": "value1",
        "param2": "value2",
    },
}
middleware := middlewares.UpdateFromMap(values,
    parameters.WithParseStepSource("map"))
```

#### UpdateFromMapAsDefault
Updates parameters only if they haven't been set already.

```go
middleware := middlewares.UpdateFromMapAsDefault(defaultValues,
    parameters.WithParseStepSource("defaults"))
```

### 7. Layer Management

#### Layer Filtering
Control which layers are processed.

```go
// Whitelist specific layers
middleware := middlewares.WhitelistLayers([]string{"api", "database"})

// Blacklist layers
middleware := middlewares.BlacklistLayers([]string{"debug"})

// Whitelist specific parameters within layers
middleware := middlewares.WhitelistLayerParameters(map[string][]string{
    "api": {"endpoint", "timeout"},
    "database": {"host", "port"},
})
```

#### Layer Manipulation
Replace or merge layer contents.

```go
// Replace a layer entirely
middleware := middlewares.ReplaceParsedLayer("config", newLayer)

// Merge layers
middleware := middlewares.MergeParsedLayers(layersToMerge)
```

### 8. Advanced Composition

#### WrapWithWhitelistedLayers
Apply middlewares only to specific layers.

```go
middleware := middlewares.WrapWithWhitelistedLayers(
    []string{"api", "database"},
    middlewares.UpdateFromEnv("API"),
    middlewares.LoadParametersFromFile("api-config.yaml"),
)
```

## Middleware Execution Order

### Priority Rules

1. **Highest Priority**: First middleware listed
2. **Lowest Priority**: Last middleware listed
3. **Defaults**: Always processed last

### Typical Order Pattern

```go
middlewares := []middlewares.Middleware{
    // Highest priority: Command-line arguments
    middlewares.ParseFromCobraCommand(cmd),
    middlewares.GatherArguments(args),
    
    // Medium priority: Configuration files
    middlewares.LoadParametersFromFile("config.yaml"),
    middlewares.GatherFlagsFromViper(),
    
    // Lower priority: Environment variables
    middlewares.UpdateFromEnv("APP"),
    
    // Lowest priority: Defaults
    middlewares.SetFromDefaults(),
}
```

## Practical Examples

### Basic CLI Application

```go
func setupMiddlewares(cmd *cobra.Command, args []string) []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd, 
            parameters.WithParseStepSource("cobra")),
        middlewares.GatherArguments(args, 
            parameters.WithParseStepSource("arguments")),
        middlewares.GatherFlagsFromViper(
            parameters.WithParseStepSource("viper")),
        middlewares.SetFromDefaults(
            parameters.WithParseStepSource("defaults")),
    }
}

// Usage
err := middlewares.ExecuteMiddlewares(
    cmd.Description().Layers, 
    parsedLayers, 
    setupMiddlewares(cmd, args)...
)
```

### Multi-Profile Configuration

```go
func setupProfileMiddlewares(profile string) []middlewares.Middleware {
    return []middlewares.Middleware{
        // Command-line overrides everything
        middlewares.ParseFromCobraCommand(cmd),
        
        // Profile-specific overrides
        middlewares.GatherFlagsFromCustomViper(
            middlewares.WithConfigFile(fmt.Sprintf("/etc/app/%s.yaml", profile)),
        ),
        
        // Shared profile configuration
        middlewares.GatherFlagsFromCustomProfiles(profile,
            middlewares.WithProfileFile("/etc/app/profiles.yaml"),
        ),
        
        // Environment variables
        middlewares.UpdateFromEnv("APP"),
        
        // Defaults
        middlewares.SetFromDefaults(),
    }
}
```

### Development vs Production Configuration

```go
func setupEnvironmentMiddlewares(env string) []middlewares.Middleware {
    middlewares := []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherArguments(args),
    }
    
    if env == "development" {
        middlewares = append(middlewares,
            middlewares.LoadParametersFromFile("dev-config.yaml"),
            middlewares.UpdateFromEnv("DEV"),
        )
    } else {
        middlewares = append(middlewares,
            middlewares.GatherFlagsFromCustomProfiles("production",
                middlewares.WithProfileFile("/etc/app/production-profiles.yaml"),
                middlewares.WithProfileRequired(true),
            ),
            middlewares.UpdateFromEnv("PROD"),
        )
    }
    
    middlewares = append(middlewares, 
        middlewares.SetFromDefaults())
    
    return middlewares
}
```

### Selective Layer Processing

```go
func setupLayerFilteredMiddlewares() []middlewares.Middleware {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        
        // API configuration only from environment
        middlewares.WrapWithWhitelistedLayers([]string{"api"},
            middlewares.UpdateFromEnv("API"),
        ),
        
        // Database configuration from secure config file
        middlewares.WrapWithWhitelistedLayers([]string{"database"},
            middlewares.LoadParametersFromFile("/secure/db-config.yaml"),
        ),
        
        // Everything else from defaults
        middlewares.SetFromDefaults(),
    }
}
```

## Advanced Usage Patterns

### Custom Middleware Creation

```go
func ValidateRequiredParameters(required []string) middlewares.Middleware {
    return func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
        return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
            err := next(layers_, parsedLayers)
            if err != nil {
                return err
            }
            
            // Validate required parameters after processing
            for _, param := range required {
                if !parsedLayers.HasParameter("default", param) {
                    return fmt.Errorf("required parameter %s is missing", param)
                }
            }
            
            return nil
        }
    }
}
```

### Conditional Middleware Application

```go
func ConditionalMiddleware(condition bool, middleware middlewares.Middleware) middlewares.Middleware {
    if condition {
        return middleware
    }
    return func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
        return next // Pass through without modification
    }
}

// Usage
middlewares := []middlewares.Middleware{
    middlewares.ParseFromCobraCommand(cmd),
    ConditionalMiddleware(enableViper, 
        middlewares.GatherFlagsFromViper()),
    middlewares.SetFromDefaults(),
}
```

### Middleware Chaining Helpers

```go
func ChainMiddlewares(middlewares ...middlewares.Middleware) middlewares.Middleware {
    return middlewares.Chain(middlewares...)
}

// Create reusable middleware chains
var defaultChain = ChainMiddlewares(
    middlewares.ParseFromCobraCommand(cmd),
    middlewares.GatherFlagsFromViper(),
    middlewares.SetFromDefaults(),
)
```

## Best Practices

### 1. Order Matters
Always arrange middlewares by priority:
```go
middlewares := []middlewares.Middleware{
    // Highest priority first
    middlewares.ParseFromCobraCommand(cmd),
    middlewares.GatherArguments(args),
    
    // Configuration sources
    middlewares.LoadParametersFromFile("config.yaml"),
    middlewares.GatherFlagsFromViper(),
    
    // Environment
    middlewares.UpdateFromEnv("APP"),
    
    // Defaults last
    middlewares.SetFromDefaults(),
}
```

### 2. Error Handling
Always check for errors from ExecuteMiddlewares:
```go
err := middlewares.ExecuteMiddlewares(layers, parsedLayers, middlewares...)
if err != nil {
    return fmt.Errorf("failed to process parameters: %w", err)
}
```

### 3. Source Tracking
Use parse step sources for debugging:
```go
middleware := middlewares.ParseFromCobraCommand(cmd,
    parameters.WithParseStepSource("cobra"),
    parameters.WithParseStepMetadata(map[string]interface{}{
        "command": cmd.Name(),
    }),
)
```

### 4. Layer Isolation
Use layer filtering for security-sensitive configuration:
```go
// Only load database config from secure sources
middleware := middlewares.WrapWithWhitelistedLayers([]string{"database"},
    middlewares.LoadParametersFromFile("/secure/db-config.yaml"),
)
```

### 5. Reusable Patterns
Create middleware factories for common patterns:
```go
func CreateStandardMiddlewares(cmd *cobra.Command, args []string, configFile string) []middlewares.Middleware {
    middlewares := []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherArguments(args),
    }
    
    if configFile != "" {
        middlewares = append(middlewares,
            middlewares.LoadParametersFromFile(configFile))
    }
    
    middlewares = append(middlewares,
        middlewares.GatherFlagsFromViper(),
        middlewares.SetFromDefaults(),
    )
    
    return middlewares
}
```

## Testing Middlewares

### Unit Testing Individual Middlewares

```go
func TestSetFromDefaults(t *testing.T) {
    layers := createTestLayers()
    parsedLayers := layers.NewParsedLayers()
    
    middleware := middlewares.SetFromDefaults()
    
    err := middlewares.ExecuteMiddlewares(layers, parsedLayers, middleware)
    require.NoError(t, err)
    
    // Verify default values were set
    value, exists := parsedLayers.GetParameter("default", "param1")
    assert.True(t, exists)
    assert.Equal(t, "default-value", value)
}
```

### Integration Testing Middleware Chains

```go
func TestMiddlewareChain(t *testing.T) {
    layers := createTestLayers()
    parsedLayers := layers.NewParsedLayers()
    
    // Set up test environment
    os.Setenv("APP_PARAM1", "env-value")
    defer os.Unsetenv("APP_PARAM1")
    
    middlewares := []middlewares.Middleware{
        middlewares.UpdateFromEnv("APP"),
        middlewares.SetFromDefaults(),
    }
    
    err := middlewares.ExecuteMiddlewares(layers, parsedLayers, middlewares...)
    require.NoError(t, err)
    
    // Environment should override defaults
    value, _ := parsedLayers.GetParameter("default", "param1")
    assert.Equal(t, "env-value", value)
}
```

### Testing Custom Middlewares

```go
func TestCustomValidationMiddleware(t *testing.T) {
    layers := createTestLayers()
    parsedLayers := layers.NewParsedLayers()
    
    // Add a value that should fail validation
    parsedLayers.SetParameter("default", "email", "invalid-email")
    
    middleware := ValidateEmailMiddleware()
    
    err := middlewares.ExecuteMiddlewares(layers, parsedLayers, middleware)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid email format")
}
```

## Conclusion

The Glazed middleware system provides a powerful foundation for building flexible, configurable CLI applications. By understanding the middleware patterns and execution order, you can create sophisticated parameter processing pipelines that handle complex configuration scenarios with ease.

Key takeaways:
- Middlewares execute in reverse order of specification
- Use appropriate sources and priorities for different configuration types
- Leverage layer filtering for security and organization
- Create reusable middleware patterns for common use cases
- Always handle errors and track parameter sources for debugging

This system enables you to build CLI applications that gracefully handle configuration from multiple sources while maintaining clear precedence rules and providing excellent user experience.
