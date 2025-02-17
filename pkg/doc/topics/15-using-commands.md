---
Title: Using Commands in Glazed
Slug: using-commands
Short: Comprehensive guide on creating, configuring, running, and loading commands in Glazed.
Topics:
- commands
- flags
- arguments
- loaders
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

DISCLAIMER: This was generated with o1-preview and was not validated yet.

## Overview

The `glazed` framework offers a powerful and flexible system for defining and managing commands in your Go applications. This guide provides a comprehensive overview of how to create commands, add flags and arguments, execute commands using various run methods, and load commands from YAML configurations. By the end of this guide, you'll have a solid understanding of leveraging `glazed` to build robust command-line interfaces.

## Creating a Command

Creating a new command in `glazed` involves defining a command structure, initializing it with necessary configurations, and implementing the required interfaces.

### 1. Define the Command Structure

Start by creating a new struct that embeds `CommandDescription`. This struct will hold all configurations for your command.

```go
package cmds

import (
    "github.com/go-go-golems/glazed/pkg/cmds"
)

type MyNewCommand struct {
    *cmds.CommandDescription
}
```

### 2. Initialize the Command

Create a constructor function to initialize your command, defining its name, description, flags, and arguments.

```go
package cmds

import (
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func NewMyNewCommand() (*MyNewCommand, error) {
    return &MyNewCommand{
        CommandDescription: cmds.NewCommandDescription(
            "mynewcommand",
            cmds.WithShort("Description of my new command"),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "verbose",
                    parameters.ParameterTypeBool,
                    parameters.WithHelp("Enable verbose output"),
                    parameters.WithDefault(false),
                ),
            ),
            cmds.WithArguments(
                parameters.NewParameterDefinition(
                    "input",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Input file"),
                    parameters.WithRequired(true),
                ),
            ),
        ),
    }, nil
}
```

## Adding Flags/Arguments to Commands

Flags and arguments allow users to pass additional data to commands. `glazed` provides a declarative way to define these parameters.

### Defining Flags

Flags are optional parameters that modify the behavior of a command.

```go
flags := parameters.NewParameterDefinitions(
    parameters.WithParameterDefinitions(
        parameters.NewParameterDefinition(
            "count",
            parameters.ParameterTypeInteger,
            parameters.WithHelp("Number of items to process"),
            parameters.WithDefault(5),
        ),
    ),
)
```

### Defining Arguments

Arguments are positional parameters required by the command.

```go
arguments := parameters.NewParameterDefinitions(
    parameters.WithParameterDefinitions(
        parameters.NewParameterDefinition(
            "filename",
            parameters.ParameterTypeString,
            parameters.WithHelp("Name of the file to process"),
            parameters.WithRequired(true),
        ),
    ),
)
```

### Adding Flags and Arguments to a Command

Integrate the defined flags and arguments into your command's description.

```go
cmd := cmds.NewCommandDescription(
    "process",
    cmds.WithShort("Process a file"),
    cmds.WithFlags(flags.ToList()...),
    cmds.WithArguments(arguments.ToList()...),
)
```

## Running a Command

`glazed` supports multiple run methods to execute commands based on the desired output and interaction.

### 1. BareCommand

Use `BareCommand` for commands that perform actions without producing direct output.

```go
type MyBareCommand struct {
    *cmds.CommandDescription
}

func (c *MyBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Implement command logic here
    return nil
}
```

### 2. WriterCommand

Use `WriterCommand` when your command needs to write output to a provided writer.

```go
type MyWriterCommand struct {
    *cmds.CommandDescription
}

func (c *MyWriterCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
    // Implement command logic and write output to w
    _, err := w.Write([]byte("Output from WriterCommand"))
    return err
}
```

### 3. GlazeCommand

Use `GlazeCommand` for commands that emit structured data using `GlazeProcessor`.

```go
type MyGlazeCommand struct {
    *cmds.CommandDescription
}

func (c *MyGlazeCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    // Create and add rows to the processor
    row := types.NewRow(
        types.MRP("id", 1),
        types.MRP("name", "Example"),
    )
    return gp.AddRow(ctx, row)
}
```

## Using a Loader to Load Commands from YAML

`glazed` allows you to define commands in YAML files, enabling easier configuration and management.

### 1. Define Commands in YAML

Create a YAML file defining your command, including its flags and arguments.

```yaml
---
name: example
short: Example command loaded from YAML
flags:
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
arguments:
  - name: input
    type: string
    help: Input file
    required: true
---
```

### 2. Load Commands from YAML

Use the `CommandLoader` to load commands from the YAML file.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/loaders"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "os"
)

func main() {
    loader := loaders.NewYAMLCommandLoader()
    commands, err := loader.LoadCommands(os.DirFS("."), "commands.yaml", nil, nil)
    if err != nil {
        panic(err)
    }

    // Integrate loaded commands into your application
    for _, cmd := range commands {
        // Add cmd to your command hierarchy
    }
}
```

### 3. Integrate Loaded Commands

Once loaded, integrate the commands into your application's command hierarchy, typically using a root command with subcommands.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "Root command",
    }

    // Assume commands are loaded into a slice called loadedCommands
    for _, cmd := range loadedCommands {
        cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
        if err != nil {
            panic(err)
        }
        rootCmd.AddCommand(cobraCmd)
    }

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Running Commands Programmatically

The `glazed/pkg/cmds/runner` package provides a flexible way to run commands programmatically without using Cobra. This is useful when embedding Glazed commands in your own applications or when you need more control over parameter parsing and command execution.

### Basic Command Execution

The simplest way to run a command is using the `ParseAndRun` function:

```go
import (
    "context"
    "github.com/go-go-golems/glazed/pkg/cmds/runner"
)

func executeCommand(cmd cmds.Command) error {
    ctx := context.Background()
    return runner.ParseAndRun(ctx, cmd, nil, nil)
}
```

### Configuring Parameter Parsing

The runner package provides several options to customize how parameters are parsed:

```go
// Set up parsing options
parseOptions := []runner.ParseOption{
    // Load from environment variables with prefix
    runner.WithEnvMiddleware("APP_"),
    
    // Load from Viper configuration
    runner.WithViper(),
    
    // Provide values for specific layers
    runner.WithValuesForLayers(map[string]map[string]interface{}{
        "config": {
            "host": "localhost",
            "port": 8080,
        },
        "api": {
            "timeout": 30,
            "retries": 3,
        },
    }),
    
    // Add custom middleware
    runner.WithAdditionalMiddlewares(customMiddleware),
}

// Parse parameters
parsedLayers, err := runner.ParseCommandParameters(cmd, parseOptions...)
if err != nil {
    return err
}
```

### Configuring Command Execution

You can customize how commands are executed using run options:

```go
// Set up run options
runOptions := []runner.RunOption{
    // Specify output writer for WriterCommand
    runner.WithWriter(customWriter),
    
    // Provide custom processor for GlazeCommand
    runner.WithGlazeProcessor(customProcessor),
}

// Run the command with parsed parameters
err = runner.RunCommand(ctx, cmd, parsedLayers, runOptions...)
```

### Parameter Loading Order

When using the default middleware chain, parameters are loaded in the following order (later sources override earlier ones):

1. Base parameter defaults (from parameter definitions)
2. Layer values (via `WithValuesForLayers`)
3. Environment variables (if enabled via `WithEnvMiddleware`)
4. Viper configuration (if enabled via `WithViper`)
5. Custom middlewares (via `WithAdditionalMiddlewares`)

### Separate Parsing and Execution

For more control, you can separate parameter parsing from command execution:

```go
func executeWithCustomParsing(cmd cmds.Command) error {
    ctx := context.Background()
    
    // Parse parameters
    parsedLayers, err := runner.ParseCommandParameters(cmd,
        runner.WithEnvMiddleware("APP_"),
        runner.WithValuesForLayers(defaultValues),
    )
    if err != nil {
        return err
    }
    
    // Modify parsed parameters if needed
    // ...
    
    // Execute command
    return runner.RunCommand(ctx, cmd, parsedLayers,
        runner.WithWriter(os.Stdout),
    )
}
```

### Example: Complete Command Execution

Here's a complete example showing how to run a command with custom configuration:

```go
func runCustomCommand(cmd cmds.Command) error {
    ctx := context.Background()
    
    // Define default values
    values := map[string]map[string]interface{}{
        "config": {
            "environment": "development",
            "logLevel": "debug",
        },
    }
    
    // Setup parsing options
    parseOptions := []runner.ParseOption{
        runner.WithValuesForLayers(values),
        runner.WithEnvMiddleware("APP_"),
        runner.WithViper(),
    }
    
    // Setup run options
    runOptions := []runner.RunOption{
        runner.WithWriter(os.Stdout),
    }
    
    // Parse and run in one step
    return runner.ParseAndRun(ctx, cmd, parseOptions, runOptions)
}
```

This approach provides a flexible way to run Glazed commands programmatically while maintaining control over parameter parsing and command execution.

## JSON Schema Generation

Commands in Glazed can generate JSON Schema representations of their parameters using the `ToJsonSchema` method. This is useful for:

- Documenting command parameters in a standardized format
- Enabling client-side validation
- Supporting auto-completion and parameter hints in tools
- Integrating with external tools that understand JSON Schema

### Using ToJsonSchema

The `ToJsonSchema` method returns a `CommandJsonSchema` struct that follows the JSON Schema specification:

```go
schema, err := cmd.Description().ToJsonSchema()
if err != nil {
    // Handle error
}

// Pretty print the schema
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
if err := encoder.Encode(schema); err != nil {
    // Handle error
}
```

The generated schema includes:
- Parameter types and descriptions
- Required vs optional parameters
- Default values
- Enum choices for choice parameters
- Array types for list parameters
- Nested object structures for complex parameters

### Schema Structure

The generated JSON Schema follows this structure:

```json
{
  "type": "object",
  "description": "Command description",
  "properties": {
    "parameterName": {
      "type": "string|number|boolean|array|object",
      "description": "Parameter help text",
      "default": "Default value if specified",
      "enum": ["choice1", "choice2"] // For choice parameters
    }
  },
  "required": ["required_parameter_names"]
}
```

### Parameter Type Mapping

Glazed parameter types are mapped to JSON Schema types as follows:

Basic Types:
- `ParameterTypeString` → `"type": "string"`
- `ParameterTypeInteger` → `"type": "integer"`
- `ParameterTypeFloat` → `"type": "number"`
- `ParameterTypeBool` → `"type": "boolean"`
- `ParameterTypeDate` → `"type": "string"` with `"format": "date"`

List Types:
- `ParameterTypeStringList` → `"type": "array"` with string items
- `ParameterTypeIntegerList` → `"type": "array"` with integer items
- `ParameterTypeFloatList` → `"type": "array"` with number items

Choice Types:
- `ParameterTypeChoice` → `"type": "string"` with enum values
- `ParameterTypeChoiceList` → `"type": "array"` with string items and enum values

File Types:
- `ParameterTypeFile` → `"type": "object"` with path and content properties
- `ParameterTypeFileList` → `"type": "array"` with file objects as items

Key-Value Type:
- `ParameterTypeKeyValue` → `"type": "object"` with key and value string properties

File-Based Types:
- `ParameterTypeStringFromFile` → `"type": "string"`
- `ParameterTypeStringFromFiles` → `"type": "array"` with string items
- `ParameterTypeObjectFromFile` → `"type": "object"` with additional string properties
- `ParameterTypeObjectListFromFile` → `"type": "array"` with object items
- `ParameterTypeObjectListFromFiles` → `"type": "array"` with object items
- `ParameterTypeStringListFromFile` → `"type": "array"` with string items
- `ParameterTypeStringListFromFiles` → `"type": "array"` with string items

## Summary

By following this guide, you can effectively create, configure, run, and manage
commands in the `glazed` framework. Whether you're building simple utilities or
complex applications, `glazed` provides the tools necessary to streamline your
command-line interface development.
