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

## Summary

By following this guide, you can effectively create, configure, run, and manage commands in the `glazed` framework. Whether you're building simple utilities or complex applications, `glazed` provides the tools necessary to streamline your command-line interface development.
