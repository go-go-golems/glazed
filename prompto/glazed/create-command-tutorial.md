Here's a clean markdown version of the guide, assuming all flags are set to true and all command types are shown:

# Creating a New Command for `glazed` Framework

The `glazed` framework provides a declarative way to define commands with flags, arguments, and output structured data. This README will guide you step-by-step on how to create a new command for the framework.

## 1. Understanding the Building Blocks

### 1.1 Parameter Types

ParameterType establishes the data type of a parameter. Commands can incorporate multiple flags and arguments, each with a designated parameter type.

```go
import "github.com/go-go-golems/glazed/pkg/cmds/parameters"

ParameterTypeString -> string
ParameterTypeFile -> FileData
ParameterTypeInteger -> int
ParameterTypeStringFromFile -> string
ParameterTypeStringFromFiles -> string
ParameterTypeFileList -> []FileData
ParameterTypeObjectListFromFile -> []map[string]interface{}
ParameterTypeObjectListFromFiles -> []map[string]interface{}
ParameterTypeObjectFromFile -> map[string]interface{}
ParameterTypeStringListFromFile -> []string
ParameterTypeStringListFromFiles -> []string
ParameterTypeKeyValue: -> map[string]string
ParameterTypeFloat -> float64
ParameterTypeBool -> bool
ParameterTypeDate  -> time.Time
ParameterTypeStringList -> []string
ParameterTypeIntegerList -> []int
ParameterTypeFloatList -> []float64
ParameterTypeChoice -> string
ParameterTypeChoiceList -> []string
```

### 1.2 FileData

`FileData` is a structure that provides detailed information about a file. This is useful when your command needs to work with files.

```go
import "github.com/go-go-golems/glazed/pkg/cmds/parameters"

Content: File's string content.
ParsedContent: Parsed version of the file's content (for json and yaml files).
ParseError: Any error that occurred during parsing.
RawContent: File content in byte format.
StringContent: File content as a string.
IsList: Indicates if the content represents a list.
IsObject: Signifies if the content denotes an object.
BaseName: File's base name.
Extension: File's extension.
FileType: File's type.
Path: File's path.
RelativePath: File's relative path.
AbsolutePath: File's absolute path.
Size: File's size in bytes.
LastModifiedTime: Timestamp when the file was last modified.
Permissions: File's permissions.
IsDirectory: Indicates if the file is a directory.
```

### 1.3 CommandDescription

This is a structure that contains the necessary information for registering a command. It has properties like:

- `Name`: Command name.
- `Short`: Short description.
- `Flags`: List of parameter definitions for flags.
- `Arguments`: List of parameter definitions for command arguments.

### 1.4 ParameterDefinition

This structure describes a command-line parameter, whether it's a flag or an argument.

## 2. Steps to Create a New Command

### 2.1 Define the Command Structure

Create a new structure for your command that embeds `CommandDescription`. This structure will contain all the necessary configurations for your command.

Example:

```go
import "github.com/go-go-golems/glazed/pkg/cmds"
import "github.com/go-go-golems/glazed/pkg/cmds/parameters"

type MyNewCommand struct {
    *cmds.CommandDescription
}
```

### 2.2 Initialize the Command

Create a function to initialize your new command:

```go
func NewMyNewCommand() (*MyNewCommand, error) {
    // Command initialization logic here
}
```

### 2.3 Define Flags and Arguments

Utilize the `ParameterDefinition` structure to define flags and arguments for your command.

Example:

```go
flag1 := parameters.NewParameterDefinition(
    "flagName",
    parameters.ParameterTypeString,
    parameters.WithHelp("Help description for the flag"),
    parameters.WithDefault("default_value"),
)

arg1 := parameters.NewParameterDefinition(
    "argName",
    parameters.ParameterTypeInteger,
    parameters.WithHelp("Help description for the argument"),
    parameters.WithDefault(10),
)
```

### 2.4 Assemble the Command

When assembling your command, a notable addition you can include is the `glazedParameterLayer`. This layer adds support for all the glazed structured data layer flags, enriching your command with more capabilities. Here's how you can integrate it:

```go
import 	"github.com/go-go-golems/glazed/pkg/settings"

func NewExampleCommand() (*ExampleCommand, error) {
  glazedParameterLayer, err := settings.NewGlazedParameterLayers()
  if err != nil {
      return nil, errors.Wrap(err, "could not create Glazed parameter layer")
  }

  return &ExampleCommand{
      CommandDescription: cmds.NewCommandDescription(
          "example",
          cmds.WithShort("Example command"),
          cmds.WithFlags(
              parameters.NewParameterDefinition(
                  "count",
                  parameters.ParameterTypeInteger,
                  parameters.WithHelp("Number of rows to output"),
                  parameters.WithDefault(10),
              ),
          ),
          cmds.WithArguments(
              parameters.NewParameterDefinition(
                  "test",
                  parameters.ParameterTypeBool,
                  parameters.WithHelp("Whether to add a test column"),
                  parameters.WithDefault(false),
              ),
          ),
          cmds.WithLayersList(
              glazedParameterLayer,
          ),
      ),
  }, nil
}
```

You can then create a struct to represent those parameters and map them to flags using the glazed.parameter tag. The tag has to match the parameter definition name exactly.

```go
type ExampleSettings struct {
    Count int `glazed.parameter:"count"`
    Test  bool `glazed.parameter:"test"`
}
```

## Running the command

Run or RunIntoWriter is called to actually execute the command.

parsedLayers contains the result of parsing each layer that has been registered with the command description. These layers can be glazed structured data flags, database connection parameters, application specification parameters.

ps is a convenience map containing *all* parsed flags.

### BareCommand

The basic command type that requires a simpler `Run` method signature. The focus here is on providing the parsed layers and parsed flags for your command execution. Output is entirely left to the BareCommand.

```go
package "github.com/go-go-golems/glazed/pkg/cmds"

type BareCommand interface {
    Command
    Run(
        ctx context.Context,
        parsedLayers *layers.ParsedLayers,
    ) error
}
```

### WriterCommand

This command type introduces the capability to write outputs to a provided writer. It's particularly useful when your command needs to print or send data to an external output stream.

```go
package "github.com/go-go-golems/glazed/pkg/cmds"
type WriterCommand interface {
    Command
    RunIntoWriter(
        ctx context.Context,
        ps *layers.ParsedLayers,
        w io.Writer,
    ) error
}
```

### GlazeCommand

The GlazeCommand is a specialized command type that deals with the `glazed` framework's structured data capabilities. Its `Run` method signature not only encompasses parsed layers and parsed flags but also a `GlazeProcessor` which can be utilized to emit rows of data.

```go
package "github.com/go-go-golems/glazed/pkg/cmds"
type GlazeCommand interface {
    Command
    RunIntoGlazeProcessor(
        ctx context.Context,
        parsedLayers *layers.ParsedLayers,
        gp middlewares.Processor,
    ) error
}
```

It's vital to understand the specifics of each command type to ensure your `Run` method aligns with the intended command behavior.

## 2.6 Implementing the `Run` Method for a `GlazeCommand`

Given the specialized nature of a `GlazeCommand`, here's how you can effectively implement the `Run` method for this command type:

```go
import 	"github.com/go-go-golems/glazed/pkg/cmds/layers"

func (c *ExampleCommand) Run(
  ctx context.Context,
  parsedLayers *layers.ParsedLayers,
  gp middlewares.Processor,
) error {
  s := &ExampleSettings{}
  if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
    return err
  }

  for i := 0; i < s.Count; i++ {
    row := types.NewRow(
      types.MRP("id", i),
      types.MRP("name", "foobar-"+strconv.Itoa(i)),
    )
    
    if s.Test {
      row.Set("test", rand.Intn(100)+1)
    }
    
    if err := gp.AddRow(ctx, row); err != nil {
      return err
    }
  }

  return nil
}
```

## 2.7 Creating and Managing Rows in `GlazeCommand` with `types.Row`

Within the `glazed` framework, `types.Row` represents an ordered map, ensuring consistent field order in rows.

### Creating Rows with `types.Row` and `MRP`

- **Direct Initialization**: Use `NewRow` with the `MRP` (MapRowPair) function to directly initialize rows while maintaining the field order:

```go
import "github.com/go-go-golems/glazed/pkg/types"

row := types.NewRow(
  types.MRP("id", i),
  types.MRP("name", "foobar-"+strconv.Itoa(i)),
)
```

`MRP` is a utility function to quickly create a key-value pair for a row.

### Other Methods for Creating Rows

- **From Map**: Use `NewRowFromMap` to create a row from a regular map. It sorts and maintains the order of keys:

```go
data := map[string]interface{}{"id": 1, "name": "Alice"}
row := types.NewRowFromMap(data)
```

- **From Struct**: `NewRowFromStruct` allows you to convert a struct into a row. Optionally, you can specify whether struct field names should be converted to lowercase:

```go
type Person struct {
   ID   int
   Name string
}
person := Person{ID: 1, Name: "Alice"}
row := types.NewRowFromStruct(&person, true)  // true indicates lowercase field names.
```

These functions offer flexibility in how you create and work with rows within the `glazed` framework.

## 3. Integrating `glazed` Commands with Cobra

Now, the core part: adding your `glazed` command to the Cobra root command. This requires converting the `glazed` command into a Cobra command using the provided utility functions.

```go
import "github.com/go-go-golems/glazed/pkg/cli"

func GetVerbsFromCobraCommand(cmd *cobra.Command) []string 
func BuildCobraCommandFromCommandAndFunc(
   s cmds.Command,
   run CobraRunFunc,
   options ...CobraParserOption,
) (*cobra.Command, error)
func BuildCobraCommandFromBareCommand(c cmds.BareCommand, options ...CobraParserOption) (*cobra.Command, error)
func BuildCobraCommandFromWriterCommand(s cmds.WriterCommand, options ...CobraParserOption) (*cobra.Command, error)
func BuildCobraCommandAlias(
   alias *alias.CommandAlias,
   options ...CobraParserOption,
) (*cobra.Command, error)
// findOrCreateParentCommand will create empty commands to anchor the passed in parents.
func findOrCreateParentCommand(rootCmd *cobra.Command, parents []string) *cobra.Command
func BuildCobraCommandFromGlazeCommand(cmd_ cmds.GlazeCommand, options ...CobraParserOption) (*cobra.Command, error)
func BuildCobraCommandFromCommand(
   command cmds.Command,
   options ...CobraParserOption,
) (*cobra.Command, error)
func AddCommandsToRootCommand(
   rootCmd *cobra.Command,
   commands []cmds.Command,
   aliases []*alias.CommandAlias,
   options ...CobraParserOption,
) error
```

```go
import "github.com/go-go-golems/glazed/pkg/cli"

func main() {
  var rootCmd = &cobra.Command{
    Use:   "coommandName",
    Short: "Short Command Description",
  }
  
  helpSystem := help.NewHelpSystem()
  
  helpSystem.SetupCobraRootCommand(rootCmd)

  glazeCmdInstance, err := NewMyGlazeCommand() // Assuming you've created this function
  cobra.CheckErr(err)

  command, err := cli.BuildCobraCommandFromGlazedCommand(glazeCmdInstance)
  cobra.CheckErr(err)

  rootCmd.AddCommand(command)
  
  err = rootCmd.Execute()
  cobra.CheckErr(err)
}
```