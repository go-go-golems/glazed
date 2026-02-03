---
Title: Glazed Middlewares Guide
Slug: middlewares-guide
Short: Learn how to use Glazed's middleware system to load field values from various sources
Topics:
- middlewares
- fields
- configuration
Commands:
- ExecuteMiddlewares
- SetFromDefaults
- UpdateFromEnv
- LoadFieldsFromFile
- ParseFromCobraCommand
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Middlewares Guide: Loading Field Values

## Overview

Glazed provides a flexible middleware system for loading field values from various sources. This guide explains how to use these middlewares effectively to populate your command fields from different locations like environment variables, config files, and command line arguments.

## Key Concepts

### Middleware Function Signature

A middleware function in the Glazed framework has the following signature:

```go
type HandlerFunc func(sections *schema.Schema, parsedSections *values.Values) error
type Middleware func(next HandlerFunc) HandlerFunc
```

### Relationship between Schema and Values

- **Schema**: These are collections of field definitions. They define the structure and metadata of fields, such as their names, types, and default values.

- **Values**: These are collections of parsed field values. They store the actual values obtained from various sources like command-line arguments, environment variables, or configuration files.

Middlewares operate on these two structures to manage and transform field values.

### Purpose of Middlewares

Middlewares in the Glazed framework serve several purposes:

1. **Modular Field Handling**: They allow for modular and reusable field processing logic. Each middleware can focus on a specific source or transformation of field values.

2. **Logging and Tracking**: Each middleware can log its actions, providing a trace of how field values were derived.

### Adding Information to Parsed Fields

Each middleware can add information to the parsed fields by:

- Setting default values if no value exists.
- Overriding existing values with those from a more specific source.
- Logging the source and transformation steps for each field value.

### Middleware Structure

Each middleware follows a consistent pattern:
- It receives a `next` handler function
- It can process fields before and/or after calling `next`
- It works with two main structures:
    - `Schema`: Contains field definitions
    - `Values`: Stores the actual field values

### Order of Execution

Middlewares are executed in reverse order of how they're provided to `ExecuteMiddlewares`. For example:

```go
ExecuteMiddlewares(sections, parsedSections,
    SetFromDefaults(),
    UpdateFromEnv("APP"),
    LoadFieldsFromFile("config.yaml"),
)
```

Will execute in this order:
1. LoadFieldsFromFile
2. UpdateFromEnv
3. SetFromDefaults

## Common Middlewares

### 1. Setting Default Values

Use `SetFromDefaults` to populate fields with their default values:

```go
middleware := sources.FromDefaults(
    sources.WithSource("defaults"),
)
```

This middleware reads the default values specified in field definitions and sets them if no value exists.

### 2. Environment Variables

Use `UpdateFromEnv` to load values from environment variables:

```go
middleware := sources.FromEnv("APP", 
    sources.WithSource("env"),
)
```

This will look for environment variables with the specified prefix. For example:
- Field `port` becomes `APP_PORT`
- Field `db_host` becomes `APP_DB_HOST`

### 3. Configuration Files

Load fields from JSON or YAML files using `LoadFieldsFromFile`:

```go
middleware := sources.FromFile("config.yaml",
    middlewares.WithParseOptions(
        sources.WithSource("config"),
    ),
)
```

By default, `LoadFieldsFromFile` expects the config file to have this structure:
```yaml
sectionName:
  fieldName: value
  anotherField: value
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
    
    // Map flat keys to section fields
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
middleware := sources.FromFile(
    "config.yaml",
    middlewares.WithConfigFileMapper(mapper),
    middlewares.WithParseOptions(
        sources.WithSource("config"),
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

The mapper handles both structures and maps them to the standard section format. This allows you to:
- Support legacy config file formats
- Adapt to existing configuration structures
- Transform nested JSON/YAML hierarchies into section fields
- Implement custom key mapping logic

### 4. Command Line Arguments

For CLI applications using Cobra:

```go
middleware := sources.FromCobra(cmd,
    sources.WithSource("flags"),
)
```

For positional arguments (from command line):
```go
middleware := sources.FromArgs(args,
    sources.WithSource("args"),
)
```

### 5. Config Files (recommended)

Load one or more config files using built-in middlewares:

```go
// Single file
sources.FromFile("config.yaml",
    middlewares.WithParseOptions(
        sources.WithSource("config"),
    ),
)

// Multiple files (low -> high precedence)
sources.FromFiles([]string{
    "base.yaml", "env.yaml", "local.yaml",
}, middlewares.WithParseOptions(
    sources.WithSource("config"),
))
```

### 6. Custom Configuration Files

Load fields from specific config files using built-in file middlewares:

```go
// Load from a specific config file (standard format)
middleware := sources.FromFile(
    "/path/to/custom-config.yaml",
    middlewares.WithParseOptions(
        sources.WithSource("config"),
    ),
)

// Load multiple config files with overlay precedence (low -> high)
middleware := sources.FromFiles(
    []string{"base.yaml", "env.yaml", "local.yaml"},
    middlewares.WithParseOptions(
        sources.WithSource("config"),
    ),
)

// Load with custom config structure mapper
mapper := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
    // Transform your custom config structure to section map format
    // ...
}
middleware := sources.FromFile(
    "custom-structure.yaml",
    middlewares.WithConfigFileMapper(mapper),
    middlewares.WithParseOptions(
        sources.WithSource("config"),
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
middleware := sources.FromMapAsDefault(values,
    sources.WithSource("defaults"),
)
```

### 8. Section Manipulation

Glazed provides several middlewares for manipulating parsed sections directly:

#### Replacing Sections

Replace a single section:
```go
// Replace the "config" section with a new one
middleware := middlewares.ReplaceSectionValues("config", newSection)
```

Replace multiple sections at once:
```go
// Replace multiple sections with new ones
middleware := middlewares.ReplaceValues(newSections)
```

#### Merging Sections

Merge a single section:
```go
// Merge a section into the "config" section
middleware := middlewares.MergeSectionValues("config", sectionToMerge)
```

Merge multiple sections:
```go
// Merge multiple sections into existing ones
middleware := middlewares.MergeValues(sectionsToMerge)
```

#### Selective Section Operations

For more fine-grained control, you can use selective middlewares that only operate on specific sections:

```go
// Replace only specific sections
middleware := middlewares.ReplaceValuesSelective(newSections, []string{"config", "env"})

// Merge only specific sections
middleware := middlewares.MergeValuesSelective(sectionsToMerge, []string{"user", "profile"})
```

These selective middlewares are useful when you want to:
- Update only certain configuration sections while preserving others
- Merge specific profiles while keeping others untouched
- Apply partial configuration updates
- Handle targeted configuration overrides

Example using selective operations:
```go
sources.Execute(sections, parsedSections,
    // Replace only the base configuration sections
    middlewares.ReplaceValuesSelective(baseConfig, []string{"system", "defaults"}),
    
    // Merge only user-specific settings
    middlewares.MergeValuesSelective(userConfig, []string{"preferences", "history"}),
    
    // Apply full environment config
    middlewares.ReplaceValues(envConfig),
)
```

These section manipulation middlewares are useful when you need to:
- Override configuration from different sources
- Combine multiple configuration profiles
- Apply temporary field changes
- Handle dynamic configuration updates

Example combining multiple operations:
```go
sources.Execute(sections, parsedSections,
    // Replace base configuration
    middlewares.ReplaceSectionValues("base", baseConfig),
    // Merge environment-specific settings
    middlewares.MergeSectionValues("env", envSettings),
    // Apply user preferences
    middlewares.MergeSectionValues("user", userPrefs),
)
```

## Advanced Usage

### 1. Chaining Middlewares

Use `Chain` to combine multiple middlewares:

```go
combined := middlewares.Chain(
    sources.FromDefaults(),
    sources.FromEnv("APP"),
    sources.FromFile("config.yaml"),
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
    sources.FromCobra(cmd),
    ConditionalMiddleware(enableConfigFile,
        sources.FromFile("config.yaml")),
    sources.FromDefaults(),
}
```

### 2. Section Filtering

Restrict middleware operation to specific sections:

```go
// Only apply to specified sections
middleware := middlewares.WrapWithWhitelistedSections(
    []string{"config", "api"},
    sources.FromEnv("APP"),
)

// Exclude specific sections
middleware := middlewares.WrapWithBlacklistedSections(
    []string{"internal"},
    sources.FromEnv("APP"),
)
```

### 3. Map-based Updates

Update values directly from a map:

```go
values := map[string]map[string]interface{}{
    "section1": {
        "field1": "value1",
        "field2": 42,
    },
}

middleware := sources.FromMap(values,
    sources.WithSource("map"),
)
```

## Best Practices

1. **Source Tracking**: Always specify the source using `WithParseStepSource` to track where values came from.

2. **Order Matters**: Arrange middlewares so that more specific sources override more general ones:
   ```go
   ExecuteMiddlewares(sections, parsedSections,
       SetFromDefaults(),           // Most general
       UpdateFromEnv("APP"),        // More specific
       LoadFieldsFromFile(),    // More specific
       ParseFromCobraCommand(),     // Most specific
   )
   ```

3. **Error Handling**: Always check for errors returned by `ExecuteMiddlewares`:
   ```go
   err := sources.Execute(sections, parsedSections, 
       // ... middlewares ...
   )
   if err != nil {
       return err
   }
   ```

4. **Section Organization**: Group related fields into logical sections for easier management and filtering.

## Common Patterns

### Configuration Loading Hierarchy

A typical configuration loading pattern:

```go
sources.Execute(sections, parsedSections,
    // Base defaults
    sources.FromDefaults(),
    
    // Configuration files (base overlays)
    sources.FromFiles([]string{
        "config.yaml", "config.local.yaml",
    }),
    
    // Environment overrides
    sources.FromEnv("APP"),
    
    // Command-line flags (highest priority)
    sources.FromCobra(cmd),
)
```

### Profile-based Configuration

Load different configurations based on profiles:

```go
sources.Execute(sections, parsedSections,
    sources.FromDefaults(),
    middlewares.GatherFlagsFromProfiles(
        defaultProfileFile, // default profile file (commonly ~/.config/<app>/profiles.yaml)
        profileFile,        // selected profile file (can be overridden)
        profileName,        // selected profile name
        defaultProfileName, // configured default profile name (typically "default")
    ),
)
```

Notes:

- `GatherFlagsFromProfiles` takes `(defaultProfileFile, profileFile, profileName, defaultProfileName)`.
- If `profileName` / `profileFile` themselves can come from env/config/flags, make sure you **resolve them first**
  (bootstrap parse of `profile-settings`) before instantiating the profiles middleware. See `topics/15-profiles.md`.

### Custom Profile Sources

Load profiles from custom locations or other applications:

```go
// Load from a specific profile file
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "production",
    middlewares.WithProfileFile("/etc/app/custom-profiles.yaml"),
    middlewares.WithProfileParseOptions(sources.WithSource("custom-profiles")),
)

// Load from another app's profiles
middleware := middlewares.GatherFlagsFromCustomProfiles(
    "shared-config",
    middlewares.WithProfileAppName("central-config"),
    middlewares.WithProfileParseOptions(sources.WithSource("shared-profiles")),
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
    sections := createTestSections()
    parsedSections := values.New()

    middleware := sources.FromDefaults()

    err := sources.Execute(sections, parsedSections, middleware)
    require.NoError(t, err)

    // Verify default values were set
    value, exists := parsedSections.GetField("default", "param1")
    assert.True(t, exists)
    assert.Equal(t, "default-value", value)
}
```

### Integration Testing Middleware Chains

```go
func TestMiddlewareChain(t *testing.T) {
    sections := createTestSections()
    parsedSections := values.New()

    // Set up test environment
    os.Setenv("APP_PARAM1", "env-value")
    defer os.Unsetenv("APP_PARAM1")

    mws := []middlewares.Middleware{
        sources.FromEnv("APP"),
        sources.FromDefaults(),
    }

    err := sources.Execute(sections, parsedSections, mws...)
    require.NoError(t, err)

    // Environment should override defaults
    value, _ := parsedSections.GetField("default", "param1")
    assert.Equal(t, "env-value", value)
}
```

### Testing Custom Middlewares

```go
func TestCustomValidationMiddleware(t *testing.T) {
    sections := createTestSections()
    parsedSections := values.New()

    // Add a value that should fail validation
    parsedSections.SetField("default", "email", "invalid-email")

    middleware := ValidateEmailMiddleware()

    err := sources.Execute(sections, parsedSections, middleware)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid email format")
}
```

## Debugging Tips

1. Use logging middleware to track field changes:
```go
middleware := func(next middlewares.HandlerFunc) middlewares.HandlerFunc {
    return func(l *schema.Schema, pl *values.Values) error {
        // Log before
        err := next(l, pl)
        // Log after
        return err
    }
}
```

2. Inspect parsed values:
```go
parsedSections.ForEach(func(section string, params *fields.FieldValues) {
    params.ForEach(func(name string, value interface{}) {
        fmt.Printf("%s.%s = %v\n", section, name, value)
    })
})
```

Remember that middlewares are a powerful tool for managing field values, but with that power comes the need for careful organization and consideration of precedence rules.

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
        sources.FromCobra(cmd),
        sources.FromDefaults(),
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
            ShortHelpSections: []string{"default", "helpers"},
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
        sources.FromCobra(cmd),
        sources.FromArgs(args),
    }

    // Optional config file
    if commandSettings.LoadFieldsFromFile != "" {
        middlewares_ = append(middlewares_,
            sources.FromFile(commandSettings.LoadFieldsFromFile))
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
			commandSettings.DefaultProfileName,
		),
    // Env config for specific sections (if needed)
    middlewares.WrapWithWhitelistedSections(
        []string{"api", "client"},
        sources.FromEnv("APP"),
    ),
        // Defaults (lowest priority)
        sources.FromDefaults(),
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

## Section-Specific Configuration

Restrict middleware to specific sections:

```go
middlewares.WrapWithWhitelistedSections(
    []string{"api", "client"},
        sources.FromEnv("APP_"),
)
```

## Tutorial: Practical Implementation

This section provides concrete examples of implementing and using the middleware system. While the previous sections explained the architectural concepts, here we'll see how these concepts translate into working code.

### 1. Basic Setup

The foundation of Glazed's field system is the `Section`. Before we can use middlewares, we need to define our field structure. This example shows how to create a section that matches the architectural concepts discussed earlier:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

func main() {
    // Create a new field section
    section, err := schema.NewSection(
        "config",
        "Configuration Options",
        schema.WithFields(
            fields.New(
                "host",
                fields.TypeString,
                fields.WithDefault("localhost"),
                fields.WithHelp("Server hostname"),
            ),
            fields.New(
                "port",
                fields.TypeInteger,
                fields.WithDefault(8080),
                fields.WithHelp("Server port"),
            ),
        ),
    )
    if err != nil {
        panic(err)
    }

    // Create field sections container
    schema_ := schema.NewSchema(
        sections.WithSections(section),
    )
}
```

This setup demonstrates several key concepts:
- Field definitions with types, defaults, and help text
- Section organization with meaningful names and descriptions
- Error handling for section creation
- Container structure for managing multiple sections

### 2. Using Individual Middlewares

Now that we understand the middleware signature and execution order, let's see how to implement specific middlewares. These examples show how the middleware chain processes fields in practice.

#### SetFromDefaults Middleware

The `SetFromDefaults` middleware demonstrates the basic middleware pattern of processing fields after the next handler:

```go
func useDefaultsMiddleware() {
    // Create empty parsed sections
    parsedSections := values.New()

    // Create and execute the middleware
    middleware := sources.FromDefaults(
        sources.WithSource(sources.SourceDefaults),
    )

    err := sources.Execute(
        schema_,
        parsedSections,
        middleware,
    )
    if err != nil {
        panic(err)
    }

    // Access the parsed values
    configSection, _ := parsedSections.Get("config")
    hostValue, _ := configSection.GetField("host")
    // hostValue will be "localhost"
}
```

This example shows:
- Creation of empty Values to store results
- Source tracking using ParseStepSource
- Middleware execution order
- Access to parsed values
- Error handling in middleware chains

#### UpdateFromMap Middleware

The `UpdateFromMap` middleware shows how to override values from an external source:

```go
func useMapMiddleware() {
    parsedSections := values.New()

    // Define the update map
    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "example.com",
            "port": 9090,
        },
    }

    err := sources.Execute(
        schema_,
        parsedSections,
        sources.FromMap(updateMap),
    )
    if err != nil {
        panic(err)
    }
}
```

This demonstrates:
- Structured data input through maps
- Section-specific updates
- Value override patterns
- Integration with the middleware chain

### 3. Accessing Parsed Values

After middlewares process the fields, there are several ways to access the results. These patterns align with different use cases in the architecture:

```go
func accessParsedValues(parsedSections *values.Values) {
    // 1. Direct access through section
    configSection, _ := parsedSections.Get("config")
    hostValue, _ := configSection.GetField("host")

    // 2. Get all fields as a map
    dataMap := parsedSections.GetDataMap()
    host := dataMap["host"]

    // 3. Initialize a struct
    type Config struct {
        Host string `glazed:"host"`
        Port int    `glazed:"port"`
    }
    
    var config Config
    err := parsedSections.DecodeSectionInto("config", &config)
    if err != nil {
        panic(err)
    }
}
```

These access patterns support:
- Direct section access for fine-grained control
- Map-based access for dynamic field handling
- Struct initialization for type-safe field usage
- Integration with Go's type system through struct tags

### 4. Tracking Field History

One of the key features of Glazed's middleware system is its ability to track field changes. This helps debug field processing and understand value origins:

```go
func checkFieldHistory(parsedSections *values.Values) {
    configSection, _ := parsedSections.Get("config")
    hostParam, _ := configSection.Fields.Get("host")

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
- Debugging support for field processing

### 5. Complex Middleware Chaining

This example demonstrates how multiple middlewares work together in the chain, following the execution order principles discussed earlier:

```go
func chainMiddlewares() {
    parsedSections := values.New()

    // Define different field sources
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
    err := sources.Execute(
        schema_,
        parsedSections,
        sources.FromMapAsDefault(defaultMap),  // Lowest precedence
        sources.FromDefaults(
            sources.WithSource(sources.SourceDefaults),
        ),
        sources.FromMap(configMap),  // Highest precedence
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

### 6. Working with Restricted Sections

Section restriction is a powerful feature that implements the modular field handling concept discussed in the architecture:

```go
func useRestrictedSections() {
    parsedSections := values.New()

    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "restricted.com",
        },
    }

    // Only apply to whitelisted sections
    whitelistedMiddleware := middlewares.WrapWithWhitelistedSections(
        []string{"config"},
        sources.FromMap(updateMap),
    )

    // Or blacklist specific sections
    blacklistedMiddleware := middlewares.WrapWithBlacklistedSections(
        []string{"other-section"},
        sources.FromMap(updateMap),
    )

    err := sources.Execute(
        schema_,
        parsedSections,
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
- Section isolation
- Field scope control
- Middleware composition
- Complex configuration scenarios

### Integration with Larger Systems

These examples can be combined to create sophisticated field handling systems. For instance, a typical application might:

1. Define multiple field sections for different concerns
2. Set up a chain of middlewares to handle various input sources
3. Use section restrictions to manage field scope
4. Track field history for debugging
5. Access parsed values through the most appropriate pattern
