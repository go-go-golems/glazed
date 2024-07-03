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

Layers in the glazed package provide a way to group and organize parameter definitions. They allow for better structure
and modularity in command-line interfaces and other parameter-driven systems.

A layer is a logical grouping of related parameter definitions. It consists of several components:
1. **Name**: A human-readable name for the layer.
2. **Slug**: A unique identifier for the layer, used in code.
3. **Description**: A brief explanation of the layer's purpose.
4. **Prefix**: An optional prefix for parameter names within the layer.
5. **Parameter Definitions**: A collection of parameter definitions that belong to this layer.

1. **ParameterLayer**: An interface that groups parameter definitions and provides metadata.
2. **ParameterLayers**: A collection of ParameterLayer objects.

### Creating a ParameterLayer

```go
layer, err := NewParameterLayer("config", "Configuration")
if err != nil {
    // Handle error
}
```

### Adding Parameters to a Layer

```go
layer.AddFlags(
    parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool),
    parameters.NewParameterDefinition("output", parameters.ParameterTypeString),
)
```

### Creating ParameterLayers

```go
layers := NewParameterLayers(
    WithLayers(configLayer, outputLayer),
)
```

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


## Real-world Usage Example

Here's an example of how layers are used in a real-world scenario:

```go
package repo

import (
    "context"
    cmds2 "github.com/go-go-golems/clay/pkg/cmds"
    "github.com/go-go-golems/clay/pkg/repositories"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
)

type ListCommand struct {
    *cmds.CommandDescription
}

func NewListCommand(options ...cmds.CommandDescriptionOption) (*ListCommand, error) {
    glazeParameterLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    options = append(options,
        cmds.WithShort("Import a command directory or individual files into a database"),
        cmds.WithFlags(),
        cmds.WithArguments(
            parameters.NewParameterDefinition(
                "inputs",
                parameters.ParameterTypeStringList,
                parameters.WithHelp("The command directory or individual files to import"),
                parameters.WithRequired(true),
            ),
        ),
        cmds.WithLayersList(glazeParameterLayer),
    )
    
    return &ListCommand{
        CommandDescription: cmds.NewCommandDescription("list", options...),
    }, nil
}

type ListSettings struct {
    Inputs []string `glazed.parameter:"inputs"`
}

func (c *ListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &ListSettings{}
    d := parsedLayers.GetDefaultParameterLayer()
    err := d.InitializeStruct(s)
    if err != nil {
        return err
    }
    
    // Use the parsed parameters
    commands, err := repositories.LoadCommandsFromInputs(cmds2.NewRawCommandLoader(), s.Inputs)
    if err != nil {
        return err
    }
    
    err2 := cmds2.ListCommandsIntoProcessor(ctx, commands, gp)
    if err2 != nil {
        return err2
    }
    
    return nil
}
```
