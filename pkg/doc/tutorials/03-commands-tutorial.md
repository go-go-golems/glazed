---
Title: Building Commands with Glazed - A Step-by-Step Tutorial
Slug: commands-tutorial
Topics:
- commands
- tutorial
- flags
- arguments
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

DISCLAIMER: This was generated with o1-preview and was not validated yet.

# Building Commands with Glazed: A Step-by-Step Tutorial

In this tutorial, you'll learn how to create a command-line application using the `glazed` framework. We'll walk through creating a new command, adding flags and arguments, executing the command using different run methods, and loading commands from YAML configurations.

## Prerequisites

- Go installed on your machine.
- Basic knowledge of Go programming.
- Familiarity with command-line interfaces.

## Step 1: Setting Up Your Project

First, set up a new Go project and initialize a Go module.

```bash
mkdir glazed-cli
cd glazed-cli
go mod init github.com/yourusername/glazed-cli
```

## Step 2: Installing Glazed

Add the `glazed` package to your project.

```bash
go get github.com/go-go-golems/glazed
```

## Step 3: Creating a New Command

Let's create a command called `generate` that generates a specified number of user records.

### 3.1 Define the Command Structure

Create a new file `generate.go` in the `cmds` package.

```go
package cmds

import (
    "context"
    "strconv"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
    "math/rand"
)

type GenerateCommand struct {
    *cmds.CommandDescription
}

func NewGenerateCommand() (*GenerateCommand, error) {
    return &GenerateCommand{
        CommandDescription: cmds.NewCommandDescription(
            "generate",
            cmds.WithShort("Generate user records"),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "count",
                    parameters.ParameterTypeInteger,
                    parameters.WithHelp("Number of users to generate"),
                    parameters.WithDefault(10),
                ),
                parameters.NewParameterDefinition(
                    "verbose",
                    parameters.ParameterTypeBool,
                    parameters.WithHelp("Enable verbose output"),
                    parameters.WithDefault(false),
                ),
            ),
            cmds.WithArguments(),
        ),
    }, nil
}
```

## Step 4: Adding Flags and Arguments

In the previous step, we added two flags: `count` and `verbose`. These flags allow users to specify the number of user records to generate and whether to enable verbose output.

- **count** (`int`): Specifies how many user records to generate. Defaults to `10`.
- **verbose** (`bool`): Enables verbose output. Defaults to `false`.

## Step 5: Implementing Run Methods

`glazed` provides different run methods based on how you want to handle command execution and output.

### 5.1 GlazeCommand

For commands that emit structured data using `GlazeProcessor`.

```go
func (c *GenerateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    type GenerateSettings struct {
        Count   int  `glazed.parameter:"count"`
        Verbose bool `glazed.parameter:"verbose"`
    }

    settings := &GenerateSettings{}
    if err := parsedLayers.InitializeStruct("default", settings); err != nil {
        return err
    }

    for i := 1; i <= settings.Count; i++ {
        user := types.NewRow(
            types.MRP("id", i),
            types.MRP("name", "User-"+strconv.Itoa(i)),
            types.MRP("email", "user"+strconv.Itoa(i)+"@example.com"),
        )

        if settings.Verbose {
            user.Set("debug", "Verbose mode enabled")
        }

        if err := gp.AddRow(ctx, user); err != nil {
            return err
        }
    }

    return nil
}
```

### 5.2 BareCommand

For commands that perform actions without producing direct output.

```go
func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Implement command logic here
    return nil
}
```

### 5.3 WriterCommand

For commands that write output to a provided writer.

```go
func (c *GenerateCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
    settings := &GenerateSettings{}
    if err := parsedLayers.InitializeStruct("default", settings); err != nil {
        return err
    }

    for i := 1; i <= settings.Count; i++ {
        output := fmt.Sprintf("User %d: %s <user%d@example.com>\n", i, "User-"+strconv.Itoa(i), i)
        _, err := w.Write([]byte(output))
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Step 6: Loading Commands from YAML

Let's create a simple YAML command loader that leverages existing Glazed structures.

### 6.1 Create `commands.yaml`

```yaml
name: generate
short: Generate user records
long: |
  This command generates a specified number of user records
  with optional verbose output and username prefix.
flags:
  - name: count
    type: int
    help: Number of users to generate
    default: 10
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
  - name: prefix
    type: string
    help: Prefix for usernames
    default: User
arguments: []
```

### 6.2 Create a Simple YAML Command Loader

Create a new file called `yaml_loader.go` in your project:

```go
package main

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"path/filepath"
)

type YAMLCommandLoader struct{}

func NewYAMLCommandLoader() *YAMLCommandLoader {
	return &YAMLCommandLoader{}
}

type YAMLCommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
}

func (l *YAMLCommandLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	file, err := f.Open(entryName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var yamlCmd YAMLCommandDescription
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&yamlCmd)
	if err != nil {
		return nil, err
	}

	cmdOptions := []cmds.CommandDescriptionOption{
		cmds.WithShort(yamlCmd.Short),
		cmds.WithLong(yamlCmd.Long),
		cmds.WithFlags(yamlCmd.Flags...),
		cmds.WithArguments(yamlCmd.Arguments...),
	}
	cmdOptions = append(cmdOptions, options...)

	cmd := cmds.NewCommandDescription(yamlCmd.Name, cmdOptions...)

	// Here you would implement the actual command logic
	// For this example, we'll use a simple placeholder function
	glazedCmd := cmds.NewCommandFromDescription(cmd, func(
		ctx context.Context,
		parsedLayers *layers.ParsedLayers,
		gp glazed.GlobalParameters,
	) error {
		fmt.Println("Executing command:", yamlCmd.Name)
		return nil
	})

	return []cmds.Command{glazedCmd}, nil
}

func (l *YAMLCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	ext := filepath.Ext(fileName)
	return ext == ".yaml" || ext == ".yml"
}
```

This loader now uses a `YAMLCommandDescription` struct that closely mirrors the Glazed `CommandDescription` structure, making it easier to parse YAML files directly into Glazed-compatible structures.

### 6.3 Load Commands in Your Application

Modify your `main.go` to use the YAML loader:

```go
package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "glazed-cli",
		Short: "A CLI application using Glazed",
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	yamlLoader := NewYAMLCommandLoader()

	commands, err := yamlLoader.LoadCommands(os.DirFS("."), "commands.yaml", nil, nil)
	if err != nil {
		fmt.Println("Error loading commands:", err)
		os.Exit(1)
	}

	for _, cmd := range commands {
		cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
		if err != nil {
			fmt.Println("Error building Cobra command:", err)
			os.Exit(1)
		}
		rootCmd.AddCommand(cobraCmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```


## Step 7: Testing Your Command

Run your application and test the `generate` command.

```bash
go run main.go generate --count 5 --verbose
```

You should see output similar to:

```
User 1: User-1 <user1@example.com>
User 2: User-2 <user2@example.com>
User 3: User-3 <user3@example.com>
User 4: User-4 <user4@example.com>
User 5: User-5 <user5@example.com>
```

With verbose mode enabled, additional debug information will be included.

## Step 8: Extending the Command

Enhance your command by adding more flags, arguments, or integrating with other systems like databases or APIs.

### Adding a Prefix Flag

Allow users to specify a prefix for generated usernames.

```go
parameters.NewParameterDefinition(
    "prefix",
    parameters.ParameterTypeString,
    parameters.WithHelp("Prefix for usernames"),
    parameters.WithDefault("User"),
),
```

### Modifying the Run Method

Update the `RunIntoGlazeProcessor` method to use the prefix.

```go
type GenerateSettings struct {
    Count   int    `glazed.parameter:"count"`
    Verbose bool   `glazed.parameter:"verbose"`
    Prefix  string `glazed.parameter:"prefix"`
}

func (c *GenerateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    settings := &GenerateSettings{}
    if err := parsedLayers.InitializeStruct("default", settings); err != nil {
        return err
    }

    for i := 1; i <= settings.Count; i++ {
        user := types.NewRow(
            types.MRP("id", i),
            types.MRP("name", settings.Prefix+"-"+strconv.Itoa(i)),
            types.MRP("email", "user"+strconv.Itoa(i)+"@example.com"),
        )

        if settings.Verbose {
            user.Set("debug", "Verbose mode enabled")
        }

        if err := gp.AddRow(ctx, user); err != nil {
            return err
        }
    }

    return nil
}
```

Update the YAML configuration accordingly.

```yaml
---
name: generate
short: Generate user records
flags:
  - name: count
    type: int
    help: Number of users to generate
    default: 10
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
  - name: prefix
    type: string
    help: Prefix for usernames
    default: User
arguments: []
---
```

## Conclusion

Congratulations! You've successfully created a command-line application using `glazed`. You've learned how to define commands, add flags and arguments, execute commands using different run methods, and load commands from YAML configurations. With this foundation, you can build more complex and feature-rich CLI applications using the `glazed` framework.
