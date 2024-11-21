---
Title: Writing YAML Command Files
Slug: writing-yaml-commands
Short: A comprehensive guide on writing YAML files to define glazed commands
Topics:
  - YAML
  - Commands
  - Configuration
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Writing YAML Command Files in Glazed

This guide explains how to write YAML files to define commands in the Glazed framework. YAML configuration provides a declarative way to create commands without writing Go code.

## Basic Structure

A basic YAML command file contains these essential elements:

```yaml
name: command-name
short: Short description of the command
long: |
  A longer description that can span
  multiple lines and provide more detail
flags:
  - name: flag-name
    type: string
    help: Description of the flag
arguments:
  - name: arg-name
    type: string
    required: true
    help: Description of the argument
```

## Parameter Types

Glazed supports these parameter types for both flags and arguments:

### Basic Types
- `string`: Text values
- `int`: Integer values
- `float`: Floating-point numbers
- `bool`: Boolean true/false values
- `date`: Date values
- `choice`: Single selection from predefined choices
- `choiceList`: Multiple selections from predefined choices

### List Types
- `stringList`: List of strings
- `intList`: List of integers
- `floatList`: List of floating-point numbers

### File-Related Types
- `file`: Single file input, provides detailed file metadata and content
- `fileList`: List of files with metadata and content
- `stringFromFile`: Load string content from a file (prefix with @)
- `stringFromFiles`: Load string content from multiple files
- `stringListFromFile`: Load list of strings from a file
- `stringListFromFiles`: Load list of strings from multiple files

### Object Types
- `objectFromFile`: Load and parse structured data from a file
- `objectListFromFile`: Load and parse list of objects from a file
- `objectListFromFiles`: Load and parse lists of objects from multiple files

### Key-Value Types
- `keyValue`: Parse key-value pairs from string (comma-separated) or file (with @ prefix)

## Defining Flags

Flags are optional parameters that modify command behavior. Here's a comprehensive example:

```yaml
flags:
  - name: output-format
    shortFlag: f  # Optional short form
    type: string
    default: "json"
    help: "Output format (json, yaml, or table)"
    choices: ["json", "yaml", "table"]
    required: false

  - name: limit
    type: int
    default: 10
    help: "Maximum number of items to process"

  - name: tags
    type: stringList
    help: "List of tags to filter by"
    default: ["default"]
```

### Flag Properties

- `name`: The flag name (required)
- `shortFlag`: Optional single-character alias
- `type`: Parameter type (required)
- `default`: Default value if flag is not specified
- `help`: Help text describing the flag
- `choices`: List of valid values
- `required`: Whether the flag must be specified

## Defining Arguments

Arguments are positional parameters. They're defined similarly to flags:

```yaml
arguments:
  - name: source
    type: file
    help: "Source file to process"
    required: true

  - name: destinations
    type: stringList
    help: "List of destination paths"
    default: ["./output"]
```

### Argument Properties

Arguments support the same properties as flags, except for `shortFlag`.

## Using Layers

Layers help organize related parameters. Define layers in your YAML:

```yaml
layers:
  - name: Output Configuration
    slug: output
    description: "Configure output formatting"
    prefix: "output-"
    flags:
      - name: format
        type: string
        default: "json"
      - name: pretty
        type: bool
        default: true

  - name: Processing Options
    slug: processing
    description: "Control processing behavior"
    flags:
      - name: threads
        type: int
        default: 4
```

## Best Practices

1. **Naming Conventions**
   - Use kebab-case for command and flag names
   - Make names descriptive but concise
   - Use consistent prefixes for related flags

3. **Organization**
   - Group related parameters using layers
   - Use consistent parameter ordering
   - Keep command files focused and single-purpose

4. **Validation**
   - Use `choices` to restrict valid values
   - Set appropriate default values
   - Mark parameters as required when necessary

## Examples

### Simple Query Command

```yaml
name: list-items
short: List items from database
flags:
  - name: limit
    type: int
    default: 10
    help: Maximum number of items to return

  - name: format
    type: string
    default: table
    choices: [table, json, yaml]
    help: Output format
```

### Complex Data Processing Command

```yaml
name: process-data
short: Process data files with multiple options
layers:
  - name: Input Options
    slug: input
    flags:
      - name: source
        type: file
        required: true
        help: Source data file

      - name: format
        type: string
        default: csv
        choices: [csv, json, xml]
        help: Input file format

  - name: Processing Options
    slug: processing
    flags:
      - name: batch-size
        type: int
        default: 1000
        help: Number of records to process at once

      - name: threads
        type: int
        default: 4
        help: Number of processing threads

  - name: Output Options
    slug: output
    flags:
      - name: destination
        type: directory
        required: true
        help: Output directory

      - name: compress
        type: bool
        default: false
        help: Compress output files
```

## Conclusion

YAML command files provide a powerful way to define Glazed commands declaratively. By following these guidelines and patterns, you can create clear, maintainable, and user-friendly command definitions that take full advantage of Glazed's capabilities.
