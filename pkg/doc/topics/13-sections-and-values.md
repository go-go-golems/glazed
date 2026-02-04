---
Title: Sections and Values
Slug: sections-and-values
Short: |
    Learn how to use sections and values in Glazed to organize and manage field definitions.
Topics:
  - sections
  - middleware
  - configuration
Commands:
Flags:
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Field Sections

Sections in the glazed package provide a way to group and organize field definitions. They allow for better structure and modularity in command-line interfaces and other field-driven systems.

A section is a logical grouping of related field definitions. It consists of several components:
1. **Name**: A human-readable name for the section.
2. **Slug**: A unique identifier for the section, used in code.
3. **Description**: A brief explanation of the section's purpose.
4. **Prefix**: An optional prefix for field names within the section.
5. **Field Definitions**: A collection of field definitions that belong to this section.

### Field Definitions

A `Definition` defines a field's properties, including name, type, default value, choices, and required status.

1. **Section**: An interface that groups field definitions and provides metadata.
2. **Schema**: A collection of Section objects.

## Creating and Working with Field Sections

The `SectionImpl` struct provides a straightforward implementation of the
`Section` interface.

### Creating a Field Section

You can create a new field section using the `NewSection` function:

```go
section, err := NewSection("config", "Configuration",
    WithDescription("Configuration options for the application"),
    WithPrefix("config-"),
    WithFields(
        fields.New("verbose", fields.TypeBool),
        fields.New("output", fields.TypeString),
    ),
)
if err != nil {
    // Handle error
}
```

### Adding Fields to a Section

You can add fields to an existing section using the `AddFields` method:

```go
section.AddFields(
    fields.New("log-level", fields.TypeString),
    fields.New("max-retries", fields.TypeInteger),
)
```

### Initializing Field Defaults from a Struct

You can initialize the default values of fields in a section using a struct:

```go
type Config struct {
    Verbose bool   `glazed:"verbose"`
    Output  string `glazed:"output"`
    LogLevel string `glazed:"log-level"`
    MaxRetries int `glazed:"max-retries"`
}

defaultConfig := Config{
    Verbose: true,
    Output: "stdout",
    LogLevel: "info",
    MaxRetries: 3,
}

err := section.InitializeFieldDefaultsFromStruct(&defaultConfig)
if err != nil {
    // Handle error
}
```

### Initializing Field Defaults from a Map

Alternatively, you can initialize defaults using a map:

```go
defaultValues := map[string]interface{}{
    "verbose": true,
    "output": "stdout",
    "log-level": "info",
    "max-retries": 3,
}

err := section.InitializeFieldDefaultsFromFields(defaultValues)
if err != nil {
    // Handle error
}
```

### Initializing a Struct from Field Defaults

You can also populate a struct with the default values from the field section:

```go
var config Config
err := section.InitializeStructFromFieldDefaults(&config)
if err != nil {
    // Handle error
}
```

### Loading a Field Section from YAML

You can create a field section from a YAML definition:

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

section, err := NewSectionFromYAML(yamlContent)
if err != nil {
    // Handle error
}
```

### Cloning a Field Section

To create a deep copy of a field section:

```go
clonedSection := section.Clone()
```

### Creating Schema

```go
sections := NewSchema(
    WithSections(configSection, outputSection),
)
```

### Registering Sections Under Explicit Slugs on Commands

When creating a `cmds.CommandDescription`, you can register sections under explicit slugs using `cmds.WithSectionsMap`.

```go
// Create sections with internal slugs
cfgSection, _ := schema.NewSection("config", "Configuration")
outSection, _ := schema.NewSection("output", "Output")

// Register them under different command slugs
cmd := cmds.NewCommandDescription(
    "run",
    cmds.WithSectionsMap(map[string]schema.Section{
        "cfg": cfgSection,   // registered as "cfg"
        "out": outSection,   // registered as "out"
    }),
)

// Later, parsed sections will be accessed by these slugs
// parsedSections.DecodeSectionInto("cfg", &myCfg)
```

Note: If the section is a `*schema.SectionImpl` and the key differs from the section's internal slug, the section is cloned and aligned to the registration key to maintain consistency at runtime.

### Accessing Section Information

```go
name := section.GetName()
slug := section.GetSlug()
description := section.GetDescription()
```

### ForEach and ForEachE

Iterate over all field sections:

```go
schema_.ForEach(func(key string, p Section) {
    // Process each section
})

err := schema_.ForEachE(func(key string, p Section) error {
    // Process each section, return error to stop iteration
    return nil
})
```

### Subset

Create a new Schema containing only the specified sections:

```go
subset := schema_.Subset("config", "output")
```

### AppendSections and PrependSections

Add sections to the end or beginning of the collection:

```go
schema_.AppendSections(newSection1, newSection2)
schema_.PrependSections(newSection3, newSection4)
```

### Merge

Merge two Schema collections:

```go
mergedSections := schema_.Merge(otherSchema)
```

### Clone

Create a deep copy of Schema:

```go
clonedSchema := schema_.Clone()
```

### GetAllDefinitions

Get all field definitions across all sections:

```go
allDefinitions := schema_.GetAllDefinitions()
```

## Parsed Sections

A SectionValues is the result of parsing input data (such as command-line
arguments, configuration files, or environment variables) using a Section
specification. It consists of:

1. **Section**: A reference to the original Section used for parsing.
2. **Fields**: A collection of FieldValue objects, each containing:
    - The original Definition
    - The parsed value
    - Metadata about how the value was parsed (e.g., source, parse steps)

Values is a collection of SectionValues objects, typically representing all the sections used in a command or application.

### Usage of Values

Values are primarily used to:

1. Store and organize parsed field values
2. Access parsed values across different sections
3. Initialize structs with parsed values
4. Merge parsed values from different sources
5. Provide a unified interface for accessing all parsed fields in an application


### Creating a SectionValues

```go
parsedSection, err := NewSectionValues(section,
    WithFieldValueValue("verbose", true),
)
if err != nil {
    // Handle error
}
```

### Creating Values

```go
parsedSections := NewValues(
    WithSectionValues("config", parsedConfigSection),
    WithSectionValues("output", parsedOutputSection),
)
```

### Accessing Parsed Values

```go
value, ok := parsedSection.GetField("verbose")
if !ok {
    // Field not found
}
```

### Initializing Structs from Values

```go
type Config struct {
    Verbose bool   `glazed:"verbose"`
    Output  string `glazed:"output"`
}

var config Config
err := parsedSections.DecodeSectionInto("config", &config)
if err != nil {
    // Handle error
}
```

### Merging Values

```go
parsedSections.GetDefaultSection().MergeFields(otherSectionValues)
```

### Getting All Parsed Fields

```go
allParams := parsedSections.GetAllFieldValues()
```


### ForEach and ForEachE

Iterate over all parsed sections:

```go
parsedSections.ForEach(func(k string, v *SectionValues) {
    // Process each section
})

err := parsedSections.ForEachE(func(k string, v *SectionValues) error {
    // Process each section, return error to stop iteration
    return nil
})
```

### GetDataMap

Get a map of all field values across all sections:

```go
dataMap := parsedSections.GetDataMap()
```

### GetOrCreate

Get an existing SectionValues or create a new one if it doesn't exist:

```go
parsedSection := parsedSections.GetOrCreate(someSection)
```

### Clone

Create a deep copy of Values:

```go
clonedValues := parsedSections.Clone()
```


## Middleware Integration

Middlewares in the Glazed framework provide a powerful mechanism to manage field values from various sources such as environment variables, configuration files, and command-line arguments. They allow for flexible and modular field handling in your applications.

### Key Middleware Concepts

- **Middleware Structure**: Each middleware processes fields before and/or after calling the next handler in the chain. They work with `Schema` and `Values` to manage field definitions and values.

- **Order of Execution**: Middlewares are executed in reverse order of how they're provided. This means the last middleware added is executed first.

### Common Middlewares

1. **SetFromDefaults**: Populates fields with their default values if no value exists.
   ```go
   middleware := sources.FromDefaults(
       sources.WithSource("defaults"),
   )
   ```

2. **UpdateFromEnv**: Loads values from environment variables.
   ```go
   middleware := sources.FromEnv("APP", 
       sources.WithSource("env"),
   )
   ```

3. **LoadFieldsFromFile / LoadFieldsFromFiles**: Load fields from JSON or YAML files.
   ```go
   // Single file
   middleware := sources.FromFile("config.yaml",
       middlewares.WithParseOptions(sources.WithSource("config")))

   // Multiple files (low -> high precedence)
   middleware2 := sources.FromFiles([]string{"base.yaml", "local.yaml"},
       middlewares.WithParseOptions(sources.WithSource("config")))
   ```

4. **ParseFromCobraCommand**: Parses field values from a Cobra command, typically used for CLI applications.
   ```go
   middleware := sources.FromCobra(cmd,
       sources.WithSource("flags"),
   )
   ```

### Using Middlewares

To use middlewares, chain them together and execute them with your field sections and parsed sections:


```go
sources.Execute(sections, parsedSections,
    sources.FromDefaults(),
    sources.FromFiles([]string{"config.yaml", "config.local.yaml"}),
    sources.FromEnv("APP"),
    sources.FromCobra(cmd),
)
```

### Advanced Middleware Usage

- **Chaining Middlewares**: Combine multiple middlewares using `Chain` to create a single middleware.
- **Section Filtering**: Apply middlewares to specific sections using whitelisting or blacklisting.
- **Source Tracking**: Use `WithParseStepSource` to track where field values originate.
