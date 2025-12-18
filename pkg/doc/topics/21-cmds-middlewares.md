---
Title: Glazed Middlewares Guide
Slug: middlewares-guide
Short: Learn how to use Glazed's middleware system to load parameter values from various sources
Topics:
- middlewares
- parameters
- configuration
Commands:
- ExecuteMiddlewares
- SetFromDefaults
- UpdateFromEnv
- LoadParametersFromFile
- ParseFromCobraCommand
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Middlewares Guide: Loading Parameter Values

## Overview

Glazed provides a flexible middleware system for loading parameter values from various sources. This guide explains how to use these middlewares effectively to populate your command parameters from different locations like environment variables, config files, and command line arguments.

## Key Concepts

### Middleware Function Signature

A middleware function in the Glazed framework has the following signature:

```go
type HandlerFunc func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error
type Middleware func(next HandlerFunc) HandlerFunc
```

### Relationship between ParameterLayers and ParsedLayers

- **ParameterLayers**: These are collections of parameter definitions. They define the structure and metadata of parameters, such as their names, types, and default values.

- **ParsedLayers**: These are collections of parsed parameter values. They store the actual values obtained from various sources like command-line arguments, environment variables, or configuration files.

Middlewares operate on these two structures to manage and transform parameter values.

### Purpose of Middlewares

Middlewares in the Glazed framework serve several purposes:

1. **Modular Parameter Handling**: They allow for modular and reusable parameter processing logic. Each middleware can focus on a specific source or transformation of parameter values.

2. **Logging and Tracking**: Each middleware can log its actions, providing a trace of how parameter values were derived.

### Adding Information to Parsed Parameters

Each middleware can add information to the parsed parameters by:

- Setting default values if no value exists.
- Overriding existing values with those from a more specific source.
- Logging the source and transformation steps for each parameter value.

### Middleware Structure

Each middleware follows a consistent pattern:
- It receives a `next` handler function
- It can process parameters before and/or after calling `next`
- It works with two main structures:
    - `ParameterLayers`: Contains parameter definitions
    - `ParsedLayers`: Stores the actual parameter values

### Order of Execution

Middlewares are executed in reverse order of how they're provided to `ExecuteMiddlewares`. For example:

```go
ExecuteMiddlewares(layers, parsedLayers,
    SetFromDefaults(),
    UpdateFromEnv("APP"),
    LoadParametersFromFile("config.yaml"),
)
```

Will execute in this order:
1. LoadParametersFromFile
2. UpdateFromEnv
3. SetFromDefaults

## Common Middlewares

### 1. Setting Default Values

Use `SetFromDefaults` to populate parameters with their default values:

```go
middleware := middlewares.SetFromDefaults(
    parameters.WithParseStepSource("defaults"),
)
```

This middleware reads the default values specified in parameter definitions and sets them if no value exists.

### 2. Environment Variables

Use `UpdateFromEnv` to load values from environment variables:

```go
middleware := middlewares.UpdateFromEnv("APP", 
    parameters.WithParseStepSource("env"),
)
```

This will look for environment variables with the specified prefix. For example:
- Parameter `port` becomes `APP_PORT`
- Parameter `db_host` becomes `APP_DB_HOST`

### 3. Configuration Files

Load parameters from JSON or YAML files using `LoadParametersFromFile`:

```go
middleware := middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)
```

By default, `LoadParametersFromFile` expects the config file to have this structure:
```yaml
layerName:
  parameterName: value
  anotherParameter: value
```

#### Custom Config File Structures

To use config files with different structures (flat, nested, or any custom format), provide a `ConfigFileMapper`:

```go
// Define a mapper function that transforms your config structure
mapper := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
    configMap := rawConfig.(map[string]interface{})
    result := map[string]map[string]interface{}{
        "demo": make(map[string]interface{}),
    }
    
    // Map flat keys to layer parameters
    if apiKey, ok := configMap["api_key"]; ok {
        result["demo"]["api-key"] = apiKey
    }
    
    // Handle nested structures
    if app, ok := configMap["app"].(map[string]interface{}); ok {
        if settings, ok := app["settings"].(map[string]interface{}); ok {
            if api, ok := settings["api"].(map[string]interface{}); ok {
                if key, ok := api["key"]; ok {
                    result["demo"]["api-key"] = key
                }
            }
        }
    }
    
    return result, nil
}

// Use the mapper when loading the config file
middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigFileMapper(mapper),
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)
```

Example config file with custom structure:
```yaml
# Flat structure
api_key: "secret-from-flat-config"
threshold: 42

# Triple-nested structure
app:
  settings:
    api:
      key: "secret-from-triple-nested"
```

The mapper handles both structures and maps them to the standard layer format. This allows you to:
- Support legacy config file formats
- Adapt to existing configuration structures
- Transform nested JSON/YAML hierarchies into layer parameters
- Implement custom key mapping logic

### 4. Command Line Arguments

For CLI applications using Cobra:

```go
middleware := middlewares.ParseFromCobraCommand(cmd,
    parameters.WithParseStepSource("flags"),
)
```

For positional arguments (from command line):
```go
middleware := middlewares.GatherArguments(args,
    parameters.WithParseStepSource("args"),
)
```

### 5. Config Files (recommended)

Load one or more config files using built-in middlewares:

```go
// Single file
middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)

// Multiple files (low -> high precedence)
middlewares.LoadParametersFromFiles([]string{
    "base.yaml", "env.yaml", "local.yaml",
}, middlewares.WithParseOptions(
    parameters.WithParseStepSource("config"),
))
```

### 6. Custom Configuration Files

Load parameters from specific config files using built-in file middlewares:

```go
// Load from a specific config file (standard format)
middleware := middlewares.LoadParametersFromFile(
    "/path/to/custom-config.yaml",
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)

// Load multiple config files with overlay precedence (low -> high)
middleware := middlewares.LoadParametersFromFiles(
    []string{"base.yaml", "env.yaml", "local.yaml"},
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)

// Load with custom config structure mapper
mapper := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
    // Transform your custom config structure to layer map format
    // ...
}
middleware := middlewares.LoadParametersFromFile(
    "custom-structure.yaml",
    middlewares.WithConfigFileMapper(mapper),
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)
```

These middlewares are useful for:
- Loading configuration from explicit file paths
- Applying overlays and environment-specific configurations
- Tracking parse steps with source and metadata
- Supporting custom config file formats via mappers

### 7. Default Map Updates

Set values only if they haven't been set already:

```go
middleware := middlewares.UpdateFromMapAsDefault(values,
    parameters.WithParseStepSource("defaults"),
)
```

### 8. Layer Manipulation

Glazed provides several middlewares for manipulating parsed layers directly:

#### Replacing Layers

Replace a single layer:
```go
// Replace the "config" layer with a new one
middleware := middlewares.ReplaceParsedLayer("config", newLayer)
```

Replace multiple layers at once:
```go
// Replace multiple layers with new ones
middleware := middlewares.ReplaceParsedLayers(newLayers)
```

#### Merging Layers

Merge a single layer:
```go
// Merge a layer into the "config" layer
middleware := middlewares.MergeParsedLayer("config", layerToMerge)
```

Merge multiple layers:
```go
// Merge multiple layers into existing ones
middleware := middlewares.MergeParsedLayers(layersToMerge)
```

#### Selective Layer Operations

For more fine-grained control, you can use selective middlewares that only operate on specific layers:

```go
// Replace only specific layers
middleware := middlewares.ReplaceParsedLayersSelective(newLayers, []string{"config", "env"})

// Merge only specific layers
middleware := middlewares.MergeParsedLayersSelective(layersToMerge, []string{"user", "profile"})
```

These selective middlewares are useful when you want to:
- Update only certain configuration layers while preserving others
- Merge specific profiles while keeping others untouched
- Apply partial configuration updates
- Handle targeted configuration overrides

Example using selective operations:
```go
middlewares.ExecuteMiddlewares(layers, parsedLayers,
    // Replace only the base configuration layers
    middlewares.ReplaceParsedLayersSelective(baseConfig, []string{"system", "defaults"}),
    
    // Merge only user-specific settings
    middlewares.MergeParsedLayersSelective(userConfig, []string{"preferences", "history"}),
    
    // Apply full environment config
    middlewares.ReplaceParsedLayers(envConfig),
)
```

These layer manipulation middlewares are useful when you need to:
- Override configuration from different sources
- Combine multiple configuration profiles
- Apply temporary parameter changes
- Handle dynamic configuration updates

Example combining multiple operations:
```go
middlewares.ExecuteMiddlewares(layers, parsedLayers,
    // Replace base configuration
    middlewares.ReplaceParsedLayer("base", baseConfig),
    // Merge environment-specific settings
    middlewares.MergeParsedLayer("env", envSettings),
    // Apply user preferences
    middlewares.MergeParsedLayer("user", userPrefs),
)
```

## Advanced Usage

### 1. Chaining Middlewares

Use `Chain` to combine multiple middlewares:

```go
combined := middlewares.Chain(
    middlewares.SetFromDefaults(),
    middlewares.UpdateFromEnv("APP"),
    middlewares.LoadParametersFromFile("config.yaml"),
)
```

### 2. Conditional Middleware Application

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
    ConditionalMiddleware(enableConfigFile,
        middlewares.LoadParametersFromFile("config.yaml")),
    middlewares.SetFromDefaults(),
}
```

### 2. Layer Filtering

Restrict middleware operation to specific layers:

```go
// Only apply to specified layers
middleware := middlewares.WrapWithWhitelistedLayers(
    []string{"config", "api"},
    middlewares.UpdateFromEnv("APP"),
)

// Exclude specific layers
middleware := middlewares.WrapWithBlacklistedLayers(
    []string{"internal"},
    middlewares.UpdateFromEnv("APP"),
)
```

### 3. Map-based Updates

Update values directly from a map:

```go
values := map[string]map[string]interface{}{
    "layer1": {
        "param1": "value1",
        "param2": 42,
    },
}

middleware := middlewares.UpdateFromMap(values,
    parameters.WithParseStepSource("map"),
)
```

## Best Practices

1. **Source Tracking**: Always specify the source using `WithParseStepSource` to track where values came from.

2. **Order Matters**: Arrange middlewares so that more specific sources override more general ones:
   ```go
   ExecuteMiddlewares(layers, parsedLayers,
       SetFromDefaults(),           // Most general
       UpdateFromEnv("APP"),        // More specific
       LoadParametersFromFile(),    // More specific
       ParseFromCobraCommand(),     // Most specific
   )
   ```

3. **Error Handling**: Always check for errors returned by `ExecuteMiddlewares`:
   ```go
   err := middlewares.ExecuteMiddlewares(layers, parsedLayers, 
       // ... middlewares ...
   )
   if err != nil {
       return err
   }
   ```

4. **Layer Organization**: Group related parameters into logical layers for easier management and filtering.

## Common Patterns

### Configuration Loading Hierarchy

A typical configuration loading pattern:

```go
middlewares.ExecuteMiddlewares(layers, parsedLayers,
    // Base defaults
    middlewares.SetFromDefaults(),
    
    // Configuration files (base overlays)
    middlewares.LoadParametersFromFiles([]string{
        "config.yaml", "config.local.yaml",
    }),
    
    // Environment overrides
    middlewares.UpdateFromEnv("APP"),
    
    // Command-line flags (highest priority)
    middlewares.ParseFromCobraCommand(cmd),
)
```

### Profile-based Configuration

Load different configurations based on profiles:

```go
middlewares.ExecuteMiddlewares(layers, parsedLayers,
    middlewares.SetFromDefaults(),
    middlewares.GatherFlagsFromProfiles(
        defaultProfileFile, // default profile file (commonly ~/.config/<app>/profiles.yaml)
        profileFile,        // selected profile file (can be overridden)
        profileName,        // selected profile name
    ),
)
```

Notes:

- `GatherFlagsFromProfiles` takes `(defaultProfileFile, profileFile, profileName)`.
- If `profileName` / `profileFile` themselves can come from env/config/flags, make sure you **resolve them first**
  (bootstrap parse of `profile-settings`) before instantiating the profiles middleware. See `topics/15-profiles.md`.

### Custom Profile Sources

Load profiles from custom locations or other applications:

```go
// Load from a specific profile file
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "production",
    middlewares.WithProfileFile("/etc/app/custom-profiles.yaml"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
)

// Load from another app's profiles
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "shared-config",
    middlewares.WithProfileAppName("central-config"),
    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("shared-profiles")),
)

// Load with required validation
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "critical-config",
    middlewares.WithProfileFile("/etc/app/critical.yaml"),
    middlewares.WithProfileRequired(true),  // Fail if profile not found
)
```

This is useful for:
- Loading profiles from explicit file paths
- Sharing profile configurations between applications  
- Loading different profile files based on runtime conditions
- Enforcing that critical profiles must exist

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

    mws := []middlewares.Middleware{
        middlewares.UpdateFromEnv("APP"),
        middlewares.SetFromDefaults(),
    }

    err := middlewares.ExecuteMiddlewares(layers, parsedLayers, mws...)
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

## Debugging Tips

1. Use logging middleware to track parameter changes:
```go
middleware := func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
    return func(l *layers.ParameterLayers, pl *layers.ParsedLayers) error {
        // Log before
        err := next(l, pl)
        // Log after
        return err
    }
}
```

2. Inspect parsed values:
```go
parsedLayers.ForEach(func(layer string, params *parameters.ParsedParameters) {
    params.ForEach(func(name string, value interface{}) {
        fmt.Printf("%s.%s = %v\n", layer, name, value)
    })
})
```

Remember that middlewares are a powerful tool for managing parameter values, but with that power comes the need for careful organization and consideration of precedence rules.

# Integrating Middlewares with Glazed Commands

## Basic Command Integration

To add middlewares to a Glazed command:

```go
func BuildCobraCommand(cmd cmds.Command) (*cobra.Command, error) {
    return cli.BuildCobraCommand(cmd,
        cli.WithParserConfig(cli.CobraParserConfig{
            MiddlewaresFunc: GetMiddlewares,
        }),
    )
}

func GetMiddlewares(
    commandSettings *cli.GlazedCommandSettings,
    cmd *cobra.Command,
    args []string,
) ([]middlewares.Middleware, error) {
    return []middlewares.Middleware{
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.SetFromDefaults(),
    }, nil
}
```

## Common Integration Pattern

Here's a typical pattern that handles profiles, config files, and command-line arguments:

```go
func BuildCobraCommandWithMiddlewares(
    cmd cmds.Command,
    opts ...cli.CobraOption,
) (*cobra.Command, error) {
    opts_ := append([]cli.CobraOption{
        cli.WithParserConfig(cli.CobraParserConfig{
            MiddlewaresFunc: GetCommandMiddlewares,
            ShortHelpLayers: []string{"default", "helpers"},
        }),
    }, opts...)
    return cli.BuildCobraCommand(cmd, opts_...)
}

func GetCommandMiddlewares(
    commandSettings *cli.GlazedCommandSettings,
    cmd *cobra.Command,
    args []string,
) ([]middlewares.Middleware, error) {
    middlewares_ := []middlewares.Middleware{
        // Command line args (highest priority)
        middlewares.ParseFromCobraCommand(cmd),
        middlewares.GatherArguments(args),
    }

    // Optional config file
    if commandSettings.LoadParametersFromFile != "" {
        middlewares_ = append(middlewares_,
            middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
    }

    // Profile support
    configDir, _ := os.UserConfigDir()
    defaultProfileFile := fmt.Sprintf("%s/app/profiles.yaml", configDir)
    profileFile := commandSettings.ProfileFile
    if profileFile == "" {
        profileFile = defaultProfileFile
    }

    middlewares_ = append(middlewares_,
        // Profile settings
        middlewares.GatherFlagsFromProfiles(
            defaultProfileFile,
            profileFile,
            commandSettings.Profile,
        ),
    // Env config for specific layers (if needed)
    middlewares.WrapWithWhitelistedLayers(
        []string{"api", "client"},
        middlewares.UpdateFromEnv("APP"),
    ),
        // Defaults (lowest priority)
        middlewares.SetFromDefaults(),
    )

    return middlewares_, nil
}
```

## Key Points

1. Middleware order determines priority (last middleware runs first)
2. Use `WithCobraMiddlewaresFunc` or `CobraParserConfig` to add middleware to commands
3. Common middleware order:
   - Config files (base overlays)
   - Environment
   - Positional arguments
   - Command-line flags (highest)
   - Defaults (lowest)

## Layer-Specific Configuration

Restrict middleware to specific layers:

```go
middlewares.WrapWithWhitelistedLayers(
    []string{"api", "client"},
        middlewares.UpdateFromEnv("APP_"),
)
```

## Tutorial: Practical Implementation

This section provides concrete examples of implementing and using the middleware system. While the previous sections explained the architectural concepts, here we'll see how these concepts translate into working code.

### 1. Basic Setup

The foundation of Glazed's parameter system is the `ParameterLayer`. Before we can use middlewares, we need to define our parameter structure. This example shows how to create a layer that matches the architectural concepts discussed earlier:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

func main() {
    // Create a new parameter layer
    layer, err := layers.NewParameterLayer(
        "config",
        "Configuration Options",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Server hostname"),
            ),
            parameters.NewParameterDefinition(
                "port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(8080),
                parameters.WithHelp("Server port"),
            ),
        ),
    )
    if err != nil {
        panic(err)
    }

    // Create parameter layers container
    parameterLayers := layers.NewParameterLayers(
        layers.WithLayers(layer),
    )
}
```

This setup demonstrates several key concepts:
- Parameter definitions with types, defaults, and help text
- Layer organization with meaningful names and descriptions
- Error handling for layer creation
- Container structure for managing multiple layers

### 2. Using Individual Middlewares

Now that we understand the middleware signature and execution order, let's see how to implement specific middlewares. These examples show how the middleware chain processes parameters in practice.

#### SetFromDefaults Middleware

The `SetFromDefaults` middleware demonstrates the basic middleware pattern of processing parameters after the next handler:

```go
func useDefaultsMiddleware() {
    // Create empty parsed layers
    parsedLayers := layers.NewParsedLayers()

    // Create and execute the middleware
    middleware := middlewares.SetFromDefaults(
        parameters.WithParseStepSource(parameters.SourceDefaults),
    )

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middleware,
    )
    if err != nil {
        panic(err)
    }

    // Access the parsed values
    configLayer, _ := parsedLayers.Get("config")
    hostValue, _ := configLayer.GetParameter("host")
    // hostValue will be "localhost"
}
```

This example shows:
- Creation of empty ParsedLayers to store results
- Source tracking using ParseStepSource
- Middleware execution order
- Access to parsed values
- Error handling in middleware chains

#### UpdateFromMap Middleware

The `UpdateFromMap` middleware shows how to override values from an external source:

```go
func useMapMiddleware() {
    parsedLayers := layers.NewParsedLayers()

    // Define the update map
    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "example.com",
            "port": 9090,
        },
    }

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middlewares.UpdateFromMap(updateMap),
    )
    if err != nil {
        panic(err)
    }
}
```

This demonstrates:
- Structured data input through maps
- Layer-specific updates
- Value override patterns
- Integration with the middleware chain

### 3. Accessing Parsed Values

After middlewares process the parameters, there are several ways to access the results. These patterns align with different use cases in the architecture:

```go
func accessParsedValues(parsedLayers *layers.ParsedLayers) {
    // 1. Direct access through layer
    configLayer, _ := parsedLayers.Get("config")
    hostValue, _ := configLayer.GetParameter("host")

    // 2. Get all parameters as a map
    dataMap := parsedLayers.GetDataMap()
    host := dataMap["host"]

    // 3. Initialize a struct
    type Config struct {
        Host string `glazed.parameter:"host"`
        Port int    `glazed.parameter:"port"`
    }
    
    var config Config
    err := parsedLayers.InitializeStruct("config", &config)
    if err != nil {
        panic(err)
    }
}
```

These access patterns support:
- Direct layer access for fine-grained control
- Map-based access for dynamic parameter handling
- Struct initialization for type-safe parameter usage
- Integration with Go's type system through struct tags

### 4. Tracking Parameter History

One of the key features of Glazed's middleware system is its ability to track parameter changes. This helps debug parameter processing and understand value origins:

```go
func checkParameterHistory(parsedLayers *layers.ParsedLayers) {
    configLayer, _ := parsedLayers.Get("config")
    hostParam, _ := configLayer.Parameters.Get("host")

    // View the parsing history
    for _, step := range hostParam.Log {
        fmt.Printf("Source: %s, Value: %v\n", step.Source, step.Value)
    }
}
```

The history tracking shows:
- Source identification for each update
- Value transformation tracking
- Middleware execution order verification
- Debugging support for parameter processing

### 5. Complex Middleware Chaining

This example demonstrates how multiple middlewares work together in the chain, following the execution order principles discussed earlier:

```go
func chainMiddlewares() {
    parsedLayers := layers.NewParsedLayers()

    // Define different parameter sources
    configMap := map[string]map[string]interface{}{
        "config": {
            "host": "config.com",
            "port": 5000,
        },
    }

    defaultMap := map[string]map[string]interface{}{
        "config": {
            "host": "default.com",
            "port": 8080,
        },
    }

    // Execute middlewares in order (last middleware has highest precedence)
    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middlewares.UpdateFromMapAsDefault(defaultMap),  // Lowest precedence
        middlewares.SetFromDefaults(
            parameters.WithParseStepSource(parameters.SourceDefaults),
        ),
        middlewares.UpdateFromMap(configMap),  // Highest precedence
    )
    if err != nil {
        panic(err)
    }
}
```

This complex example illustrates:
- Priority-based middleware ordering
- Multiple data source handling
- Default value management
- Value override patterns
- Error propagation through the chain

### 6. Working with Restricted Layers

Layer restriction is a powerful feature that implements the modular parameter handling concept discussed in the architecture:

```go
func useRestrictedLayers() {
    parsedLayers := layers.NewParsedLayers()

    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "restricted.com",
        },
    }

    // Only apply to whitelisted layers
    whitelistedMiddleware := middlewares.WrapWithWhitelistedLayers(
        []string{"config"},
        middlewares.UpdateFromMap(updateMap),
    )

    // Or blacklist specific layers
    blacklistedMiddleware := middlewares.WrapWithBlacklistedLayers(
        []string{"other-layer"},
        middlewares.UpdateFromMap(updateMap),
    )

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        whitelistedMiddleware,
        blacklistedMiddleware,
    )
    if err != nil {
        panic(err)
    }
}
```

This demonstrates advanced concepts:
- Selective middleware application
- Layer isolation
- Parameter scope control
- Middleware composition
- Complex configuration scenarios

### Integration with Larger Systems

These examples can be combined to create sophisticated parameter handling systems. For instance, a typical application might:

1. Define multiple parameter layers for different concerns
2. Set up a chain of middlewares to handle various input sources
3. Use layer restrictions to manage parameter scope
4. Track parameter history for debugging
5. Access parsed values through the most appropriate pattern
