---
Title: Glazed Middlewares Guide
Slug: glazed-middlewares
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
    GatherFlagsFromViper(),
)
```

Will execute in this order:
1. GatherFlagsFromViper
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
    parameters.WithParseStepSource("config"),
)
```

Configuration file format:
```yaml
layerName:
  parameterName: value
  anotherParameter: value
```

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

### 5. Viper Configuration

To integrate with Viper:

```go
middleware := middlewares.GatherFlagsFromViper(
    parameters.WithParseStepSource("viper"),
)
```

### 6. Default Map Updates

Set values only if they haven't been set already:

```go
middleware := middlewares.UpdateFromMapAsDefault(values,
    parameters.WithParseStepSource("defaults"),
)
```

### 7. Layer Manipulation

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
    middlewares.GatherFlagsFromViper(),
)
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
       SetFromDefaults(),        // Most general
       UpdateFromEnv("APP"),     // More specific
       GatherFlagsFromViper(),   // More specific
       ParseFromCobraCommand(),  // Most specific
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
    
    // Configuration file
    middlewares.LoadParametersFromFile("config.yaml"),
    
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
        "config.yaml",  // default profile file
        "dev.yaml",     // environment-specific profile
        "development",  // profile name
    ),
)
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
    return cli.BuildCobraCommandFromCommand(cmd,
        cli.WithCobraMiddlewaresFunc(GetMiddlewares),
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
    options ...cli.CobraParserOption,
) (*cobra.Command, error) {
    options_ := append([]cli.CobraParserOption{
        cli.WithCobraMiddlewaresFunc(GetCommandMiddlewares),
        cli.WithCobraShortHelpLayers("default", "helpers"),
    }, options...)
    return cli.BuildCobraCommandFromCommand(cmd, options_...)
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
        // Viper config for specific layers
        middlewares.WrapWithWhitelistedLayers(
            []string{"api", "client"},
            middlewares.GatherFlagsFromViper(),
        ),
        // Defaults (lowest priority)
        middlewares.SetFromDefaults(),
    )

    return middlewares_, nil
}
```

## Key Points

1. Middleware order determines priority (last middleware runs first)
2. Use `WithCobraMiddlewaresFunc` to add middleware to commands
3. Common middleware order:
   - Command-line arguments
   - Config files
   - Profiles
   - Viper settings
   - Defaults

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
