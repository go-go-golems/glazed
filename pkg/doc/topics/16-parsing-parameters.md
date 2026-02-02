---
Title: Parsing Parameters
Slug: parsing-parameters
Short: Learn how to define and parse parameters in Go applications using the Parameter API.
Topics:
  - Parameter API
  - Go Programming
  - Command-line Tools
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Introduction

The **Parameter API** facilitates parsing and managing parameters in Go applications. It's ideal for applications requiring flexible parameter handling.

## Key Concepts

### ParameterDefinition

A `ParameterDefinition` defines a parameter's properties, including name, type, default value, choices, and required status.

```go
type ParameterDefinition struct {
    Name       string        `yaml:"name"`
    ShortFlag  string        `yaml:"shortFlag,omitempty"`
    Type       ParameterType `yaml:"type"`
    Help       string        `yaml:"help,omitempty"`
    Default    *interface{}  `yaml:"default,omitempty"`
    Choices    []string      `yaml:"choices,omitempty"`
    Required   bool          `yaml:"required,omitempty"`
    IsArgument bool          `yaml:"-"`
}
```

### ParameterDefinitions

`ParameterDefinitions` is an ordered map of `ParameterDefinition` instances, indexed by name.

```go
type ParameterDefinitions struct {
    *orderedmap.OrderedMap[string, *ParameterDefinition]
}
```

### ParsedParameter

A `ParsedParameter` contains the parsed value, its `ParameterDefinition`, and a log of parsing steps.

```go
type ParsedParameter struct {
    Value               interface{}
    ParameterDefinition *ParameterDefinition
    Log                 []ParseStep
}
```

### ParsedParameters

`ParsedParameters` is an ordered map of `ParsedParameter` instances, indexed by parameter names.

```go
type ParsedParameters struct {
    *orderedmap.OrderedMap[string, *ParsedParameter]
}
```

### ParseStep

`ParseStep` records each step in the parsing process, including source, value, and metadata.

```go
type ParseStep struct {
    Source   string
    Value    interface{}
    Metadata map[string]interface{}
}
```

## Getting Started

Follow these steps to use the Parameter API: define parameters and parse them from user input or configuration files.

### Defining Parameters

Define parameters using `ParameterDefinition`, specifying name, type, and options like default values or choices.

```go
import "github.com/go-go-golems/glazed/pkg/cmds/parameters"

// Define a string parameter with a default value
paramName := fields.New(
    "username",
    fields.TypeString,
    fields.WithHelp("The name of the user"),
    fields.WithDefault("guest"),
    fields.WithRequired(false),
)
```

### Parsing Parameters

Parse input values (e.g., from command-line arguments) to obtain `ParsedParameter` instances.

```go
// Assume 'inputs' is a slice of strings received from the command line
inputs := []string{"john_doe"}

parsedParam, err := paramName.ParseParameter(inputs)
if err != nil {
    // Handle parsing error
}

fmt.Println("Parsed Value:", parsedParam.Value)
```

## Accessing Parsed Values

Access parsed parameter values via `ParsedParameters`.

```go
// Create a collection of parameter definitions
paramDefs := fields.News(
    parameters.WithParameterDefinitionList([]*fields.Definition{paramName}),
)

// Parse a parameter value
parsedParams, err := paramDefs.ParseParameters(userInputs)
if err != nil {
    // Handle error
}

// Access the parsed value
username := parsedParams.GetValue("username").(string)
fmt.Println("Username:", username)
```

## Managing Parsed Parameters

Manage parsed parameters using these methods: updating values, merging parameter sets, and cloning parsed data.

### Updating Values

Update a parsed parameter's value, optionally appending a new parsing step.

```go
parsedParam.Update("new_username", sources.WithSource("override"))
```

### Merging Parameters

Merge another `ParsedParameters` instance, combining values and logs.

```go
parsedParams.Merge(otherParsedParams)
```

### Cloning Parameters

Create a deep copy of `ParsedParameters` to avoid unintended mutations.

```go
clonedParams := parsedParams.Clone()
```

## Handling Defaults and Required Parameters

Specify default values and enforce required parameters.

```go
// Define a required integer parameter without a default
ageParam := fields.New(
    "age",
    fields.TypeInteger,
    fields.WithHelp("The age of the user"),
    fields.WithRequired(true),
)
```

During parsing, an error is returned if a required parameter is missing. If an optional parameter is missing, its default value is used.

```go
parsedParam, err := ageParam.ParseParameter([]string{})
if err != nil {
    // Since 'age' is required, an error is expected
}
```

## Advanced Features

Advanced features include logging parsing steps and file format parsing.

### Logging Parsing Steps

Each `ParsedParameter` logs parsing steps, showing how the final value was derived.

```go
for _, step := range parsedParam.Log {
    fmt.Printf("Source: %s, Value: %v, Metadata: %v\n", step.Source, step.Value, step.Metadata)
}
```

### Parsing from Files

Define parameters to accept values from files (e.g., JSON, YAML, CSV).

```go
// Define a parameter that expects a JSON file
fileParam := fields.New(
    "config",
    fields.TypeObjectFromFile,
    fields.WithHelp("Path to the configuration file"),
    fields.WithRequired(true),
)

// Parse the parameter from file input
parsedFileParam, err := fileParam.ParseParameter([]string{"config.json"})
if err != nil {
    // Handle error
}

configData := parsedFileParam.Value.(map[string]interface{})
fmt.Println("Config Data:", configData)
```
