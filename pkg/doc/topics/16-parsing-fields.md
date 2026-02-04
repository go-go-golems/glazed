---
Title: Parsing Fields
Slug: parsing-fields
Short: Learn how to define and parse fields in Go applications using the Field API.
Topics:
  - Field API
  - Go Programming
  - Command-line Tools
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Introduction

The **Field API** facilitates parsing and managing fields in Go applications. It's ideal for applications requiring flexible field handling.

## Key Concepts

### Definition

A `Definition` defines a field's properties, including name, type, default value, choices, and required status.

```go
type Definition struct {
    Name       string        `yaml:"name"`
    ShortFlag  string        `yaml:"shortFlag,omitempty"`
    Type       Type `yaml:"type"`
    Help       string        `yaml:"help,omitempty"`
    Default    *interface{}  `yaml:"default,omitempty"`
    Choices    []string      `yaml:"choices,omitempty"`
    Required   bool          `yaml:"required,omitempty"`
    IsArgument bool          `yaml:"-"`
}
```

### Definitions

`Definitions` is an ordered map of `Definition` instances, indexed by name.

```go
type Definitions struct {
    *orderedmap.OrderedMap[string, *Definition]
}
```

### FieldValue

A `FieldValue` contains the parsed value, its `Definition`, and a log of parsing steps.

```go
type FieldValue struct {
    Value               interface{}
    Definition *Definition
    Log                 []ParseStep
}
```

### FieldValues

`FieldValues` is an ordered map of `FieldValue` instances, indexed by field names.

```go
type FieldValues struct {
    *orderedmap.OrderedMap[string, *FieldValue]
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

Follow these steps to use the Field API: define fields and parse them from user input or configuration files.

### Defining Fields

Define fields using `Definition`, specifying name, type, and options like default values or choices.

```go
import "github.com/go-go-golems/glazed/pkg/cmds/fields"

// Define a string field with a default value
paramName := fields.New(
    "username",
    fields.TypeString,
    fields.WithHelp("The name of the user"),
    fields.WithDefault("guest"),
    fields.WithRequired(false),
)
```

### Parsing Fields

Parse input values (e.g., from command-line arguments) to obtain `FieldValue` instances.

```go
// Assume 'inputs' is a slice of strings received from the command line
inputs := []string{"john_doe"}

parsedParam, err := paramName.ParseField(inputs)
if err != nil {
    // Handle parsing error
}

fmt.Println("Parsed Value:", parsedParam.Value)
```

## Accessing Parsed Values

Access parsed field values via `FieldValues`.

```go
// Create a collection of field definitions
paramDefs := fields.News(
    fields.WithDefinitionList([]*fields.Definition{paramName}),
)

// Parse a field value
parsedParams, err := paramDefs.ParseFields(userInputs)
if err != nil {
    // Handle error
}

// Access the parsed value
username := parsedParams.GetValue("username").(string)
fmt.Println("Username:", username)
```

## Managing Parsed Fields

Manage parsed fields using these methods: updating values, merging field sets, and cloning parsed data.

### Updating Values

Update a parsed field's value, optionally appending a new parsing step.

```go
parsedParam.Update("new_username", sources.WithSource("override"))
```

### Merging Fields

Merge another `FieldValues` instance, combining values and logs.

```go
parsedParams.Merge(otherParsedParams)
```

### Cloning Fields

Create a deep copy of `FieldValues` to avoid unintended mutations.

```go
clonedParams := parsedParams.Clone()
```

## Handling Defaults and Required Fields

Specify default values and enforce required fields.

```go
// Define a required integer field without a default
ageParam := fields.New(
    "age",
    fields.TypeInteger,
    fields.WithHelp("The age of the user"),
    fields.WithRequired(true),
)
```

During parsing, an error is returned if a required field is missing. If an optional field is missing, its default value is used.

```go
parsedParam, err := ageParam.ParseField([]string{})
if err != nil {
    // Since 'age' is required, an error is expected
}
```

## Advanced Features

Advanced features include logging parsing steps and file format parsing.

### Logging Parsing Steps

Each `FieldValue` logs parsing steps, showing how the final value was derived.

```go
for _, step := range parsedParam.Log {
    fmt.Printf("Source: %s, Value: %v, Metadata: %v\n", step.Source, step.Value, step.Metadata)
}
```

### Parsing from Files

Define fields to accept values from files (e.g., JSON, YAML, CSV).

```go
// Define a field that expects a JSON file
fileParam := fields.New(
    "config",
    fields.TypeObjectFromFile,
    fields.WithHelp("Path to the configuration file"),
    fields.WithRequired(true),
)

// Parse the field from file input
parsedFileParam, err := fileParam.ParseField([]string{"config.json"})
if err != nil {
    // Handle error
}

configData := parsedFileParam.Value.(map[string]interface{})
fmt.Println("Config Data:", configData)
```
