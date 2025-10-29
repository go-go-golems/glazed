---
Title: Parameter Layers and Parsed Layers
Slug: parameter-layers-and-parsed-layers
Short: |
    Learn how to use parameter layers and parsed layers in Glazed to organize and manage parameter definitions.
Topics:
  - layers
  - middleware
  - configuration
Commands:
Flags:
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Parameter Layers

Layers in the glazed package provide a way to group and organize parameter definitions. They allow for better structure and modularity in command-line interfaces and other parameter-driven systems.

A layer is a logical grouping of related parameter definitions. It consists of several components:
1. **Name**: A human-readable name for the layer.
2. **Slug**: A unique identifier for the layer, used in code.
3. **Description**: A brief explanation of the layer's purpose.
4. **Prefix**: An optional prefix for parameter names within the layer.
5. **Parameter Definitions**: A collection of parameter definitions that belong to this layer.

### Parameter Definitions

A `ParameterDefinition` defines a parameter's properties, including name, type, default value, choices, and required status.

1. **ParameterLayer**: An interface that groups parameter definitions and provides metadata.
2. **ParameterLayers**: A collection of ParameterLayer objects.

## Creating and Working with Parameter Layers

The `ParameterLayerImpl` struct provides a straightforward implementation of the
`ParameterLayer` interface.

### Creating a Parameter Layer

You can create a new parameter layer using the `NewParameterLayer` function:

```go
layer, err := NewParameterLayer("config", "Configuration",
    WithDescription("Configuration options for the application"),
    WithPrefix("config-"),
    WithParameterDefinitions(
        parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool),
        parameters.NewParameterDefinition("output", parameters.ParameterTypeString),
    ),
)
if err != nil {
    // Handle error
}
```

### Adding Parameters to a Layer

You can add parameters to an existing layer using the `AddFlags` method:

```go
layer.AddFlags(
    parameters.NewParameterDefinition("log-level", parameters.ParameterTypeString),
    parameters.NewParameterDefinition("max-retries", parameters.ParameterTypeInteger),
)
```

### Initializing Parameter Defaults from a Struct

You can initialize the default values of parameters in a layer using a struct:

```go
type Config struct {
    Verbose bool   `glazed.parameter:"verbose"`
    Output  string `glazed.parameter:"output"`
    LogLevel string `glazed.parameter:"log-level"`
    MaxRetries int `glazed.parameter:"max-retries"`
}

defaultConfig := Config{
    Verbose: true,
    Output: "stdout",
    LogLevel: "info",
    MaxRetries: 3,
}

err := layer.InitializeParameterDefaultsFromStruct(&defaultConfig)
if err != nil {
    // Handle error
}
```

### Initializing Parameter Defaults from a Map

Alternatively, you can initialize defaults using a map:

```go
defaultValues := map[string]interface{}{
    "verbose": true,
    "output": "stdout",
    "log-level": "info",
    "max-retries": 3,
}

err := layer.InitializeParameterDefaultsFromParameters(defaultValues)
if err != nil {
    // Handle error
}
```

### Initializing a Struct from Parameter Defaults

You can also populate a struct with the default values from the parameter layer:

```go
var config Config
err := layer.InitializeStructFromParameterDefaults(&config)
if err != nil {
    // Handle error
}
```

### Loading a Parameter Layer from YAML

You can create a parameter layer from a YAML definition:

```go
yamlContent := []byte(`
name: Configuration
slug: config
description: Configuration options for the application
flags:
  - name: verbose
    type: bool
    help: Enable verbose output
  - name: output
    type: string
    help: Output destination
`)

layer, err := NewParameterLayerFromYAML(yamlContent)
if err != nil {
    // Handle error
}
```

### Cloning a Parameter Layer

To create a deep copy of a parameter layer:

```go
clonedLayer := layer.Clone()
```

### Creating ParameterLayers

```go
layers := NewParameterLayers(
    WithLayers(configLayer, outputLayer),
)
```

### Registering Layers Under Explicit Slugs on Commands

When creating a `cmds.CommandDescription`, you can register layers under explicit slugs using `cmds.WithLayersMap`.

```go
// Create layers with internal slugs
cfgLayer, _ := layers.NewParameterLayer("config", "Configuration")
outLayer, _ := layers.NewParameterLayer("output", "Output")

// Register them under different command slugs
cmd := cmds.NewCommandDescription(
    "run",
    cmds.WithLayersMap(map[string]layers.ParameterLayer{
        "cfg": cfgLayer,   // registered as "cfg"
        "out": outLayer,   // registered as "out"
    }),
)

// Later, parsed layers will be accessed by these slugs
// parsedLayers.InitializeStruct("cfg", &myCfg)
```

Note: If the layer is a `*layers.ParameterLayerImpl` and the key differs from the layer's internal slug, the layer is cloned and aligned to the registration key to maintain consistency at runtime.

### Accessing Layer Information

```go
name := layer.GetName()
slug := layer.GetSlug()
description := layer.GetDescription()
```

### ForEach and ForEachE

Iterate over all parameter layers:

```go
parameterLayers.ForEach(func(key string, p ParameterLayer) {
    // Process each layer
})

err := parameterLayers.ForEachE(func(key string, p ParameterLayer) error {
    // Process each layer, return error to stop iteration
    return nil
})
```

### Subset

Create a new ParameterLayers containing only the specified layers:

```go
subset := parameterLayers.Subset("config", "output")
```

### AppendLayers and PrependLayers

Add layers to the end or beginning of the collection:

```go
parameterLayers.AppendLayers(newLayer1, newLayer2)
parameterLayers.PrependLayers(newLayer3, newLayer4)
```

### Merge

Merge two ParameterLayers collections:

```go
mergedLayers := parameterLayers.Merge(otherParameterLayers)
```

### Clone

Create a deep copy of ParameterLayers:

```go
clonedParameterLayers := parameterLayers.Clone()
```

### GetAllParameterDefinitions

Get all parameter definitions across all layers:

```go
allDefinitions := parameterLayers.GetAllParameterDefinitions()
```

## Parsed Layers

A ParsedLayer is the result of parsing input data (such as command-line
arguments, configuration files, or environment variables) using a ParameterLayer
specification. It consists of:

1. **Layer**: A reference to the original ParameterLayer used for parsing.
2. **Parameters**: A collection of ParsedParameter objects, each containing:
    - The original ParameterDefinition
    - The parsed value
    - Metadata about how the value was parsed (e.g., source, parse steps)

ParsedLayers is a collection of ParsedLayer objects, typically representing all the layers used in a command or application.

### Usage of ParsedLayers

ParsedLayers are primarily used to:

1. Store and organize parsed parameter values
2. Access parsed values across different layers
3. Initialize structs with parsed values
4. Merge parsed values from different sources
5. Provide a unified interface for accessing all parsed parameters in an application


### Creating a ParsedLayer

```go
parsedLayer, err := NewParsedLayer(layer,
    WithParsedParameterValue("verbose", true),
)
if err != nil {
    // Handle error
}
```

### Creating ParsedLayers

```go
parsedLayers := NewParsedLayers(
    WithParsedLayer("config", parsedConfigLayer),
    WithParsedLayer("output", parsedOutputLayer),
)
```

### Accessing Parsed Values

```go
value, ok := parsedLayer.GetParameter("verbose")
if !ok {
    // Parameter not found
}
```

### Initializing Structs from ParsedLayers

```go
type Config struct {
    Verbose bool   `glazed.parameter:"verbose"`
    Output  string `glazed.parameter:"output"`
}

var config Config
err := parsedLayers.InitializeStruct("config", &config)
if err != nil {
    // Handle error
}
```

### Merging ParsedLayers

```go
parsedLayers.GetDefaultParameterLayer().MergeParameters(otherParsedLayer)
```

### Getting All Parsed Parameters

```go
allParams := parsedLayers.GetAllParsedParameters()
```


### ForEach and ForEachE

Iterate over all parsed layers:

```go
parsedLayers.ForEach(func(k string, v *ParsedLayer) {
    // Process each layer
})

err := parsedLayers.ForEachE(func(k string, v *ParsedLayer) error {
    // Process each layer, return error to stop iteration
    return nil
})
```

### GetDataMap

Get a map of all parameter values across all layers:

```go
dataMap := parsedLayers.GetDataMap()
```

### GetOrCreate

Get an existing ParsedLayer or create a new one if it doesn't exist:

```go
parsedLayer := parsedLayers.GetOrCreate(someParameterLayer)
```

### Clone

Create a deep copy of ParsedLayers:

```go
clonedParsedLayers := parsedLayers.Clone()
```


## Middleware Integration

Middlewares in the Glazed framework provide a powerful mechanism to manage parameter values from various sources such as environment variables, configuration files, and command-line arguments. They allow for flexible and modular parameter handling in your applications.

### Key Middleware Concepts

- **Middleware Structure**: Each middleware processes parameters before and/or after calling the next handler in the chain. They work with `ParameterLayers` and `ParsedLayers` to manage parameter definitions and values.

- **Order of Execution**: Middlewares are executed in reverse order of how they're provided. This means the last middleware added is executed first.

### Common Middlewares

1. **SetFromDefaults**: Populates parameters with their default values if no value exists.
   ```go
   middleware := middlewares.SetFromDefaults(
       parameters.WithParseStepSource("defaults"),
   )
   ```

2. **UpdateFromEnv**: Loads values from environment variables.
   ```go
   middleware := middlewares.UpdateFromEnv("APP", 
       parameters.WithParseStepSource("env"),
   )
   ```

3. **LoadParametersFromFile / LoadParametersFromFiles**: Load parameters from JSON or YAML files.
   ```go
   // Single file
   middleware := middlewares.LoadParametersFromFile("config.yaml",
       middlewares.WithParseOptions(parameters.WithParseStepSource("config")))

   // Multiple files (low -> high precedence)
   middleware2 := middlewares.LoadParametersFromFiles([]string{"base.yaml", "local.yaml"},
       middlewares.WithParseOptions(parameters.WithParseStepSource("config")))
   ```

4. **ParseFromCobraCommand**: Parses parameter values from a Cobra command, typically used for CLI applications.
   ```go
   middleware := middlewares.ParseFromCobraCommand(cmd,
       parameters.WithParseStepSource("flags"),
   )
   ```

### Using Middlewares

To use middlewares, chain them together and execute them with your parameter layers and parsed layers:


```go
middlewares.ExecuteMiddlewares(layers, parsedLayers,
    middlewares.SetFromDefaults(),
    middlewares.LoadParametersFromFiles([]string{"config.yaml", "config.local.yaml"}),
    middlewares.UpdateFromEnv("APP"),
    middlewares.ParseFromCobraCommand(cmd),
)
```

### Advanced Middleware Usage

- **Chaining Middlewares**: Combine multiple middlewares using `Chain` to create a single middleware.
- **Layer Filtering**: Apply middlewares to specific layers using whitelisting or blacklisting.
- **Source Tracking**: Use `WithParseStepSource` to track where parameter values originate.
