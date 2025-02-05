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

For positional arguments:
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