# Creating a New Command for `glazed` Framework

The `glazed` framework provides a declarative way to define commands with flags, arguments, and output structured data. This README will guide you step-by-step on how to create a new command for the framework.

## 1. **Understanding the Building Blocks**

Before diving into creating a command, it's crucial to understand the components provided by the `glazed` framework:

### 1.1 Parameter Types

`ParameterType` defines the data type of a parameter. A command can have multiple flags and arguments, each having a specific parameter type. Some examples are:

- `ParameterTypeString`: For simple string input.
- `ParameterTypeFile`: For fetching file data.
- `ParameterTypeInteger`: For integer inputs.

... and many more. Check the provided API documentation for a full list.

### 1.2 FileData

`FileData` is a structure that provides detailed information about a file. This is useful when your command needs to work with files.

### 1.3 CommandDescription

This is a structure that contains the necessary information for registering a command. It has properties like:

- `Name`: Command name.
- `Short`: Short description.
- `Flags`: List of parameter definitions for flags.
- `Arguments`: List of parameter definitions for command arguments.

### 1.4 ParameterDefinition

This structure describes a command-line parameter, whether it's a flag or an argument.

## 2. **Steps to Create a New Command**

### 2.1 Define the Command Structure

Create a new structure for your command that embeds `CommandDescription`. This structure will contain all the necessary configurations for your command.

Example:

```go
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
		cmds.WithLayers(
			glazedParameterLayer,
		),
	),
}, nil
```


---

## 2.5 **Different Types of `Run` Methods Based on Command Types**

The way you implement the `Run` method in your command depends on the type of command you're building. The `glazed`
framework offers several command types, each with a specific purpose and signature for the `Run` method.

### **BareCommand**
The basic command type that requires a simpler `Run` method signature.
The focus here is on providing the parsed layers and parsed flags for your command execution.
Output is entirely left to the BareCommand.

```go
type BareCommand interface {
	Command
	Run(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
	) error
}
```

### **WriterCommand**

This command type introduces the capability to write outputs to a provided writer. It's particularly useful when your
command needs to print or send data to an external output stream.

```go
type WriterCommand interface {
	Command
	RunIntoWriter(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
		w io.Writer,
	) error
}
```

### **GlazeCommand**

The GlazeCommand is a specialized command type that deals with the `glazed` framework's structured data capabilities.
Its `Run` method signature not only encompasses parsed layers and parsed flags but also a `GlazeProcessor` which can be
utilized to emit rows of data.

```go
type GlazeCommand interface {
	Command
	Run(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
		gp middlewares.Processor,
	) error
}
```

It's vital to understand the specifics of each command type to ensure your `Run` method aligns with the intended command behavior.

## 2.6 **Implementing the `Run` Method for a `GlazeCommand`**

Given the specialized nature of a `GlazeCommand`, here's how you can effectively implement the `Run` method for this
command type:

```go
func (c *ExampleCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	count := ps["count"].(int)
	test := ps["test"].(bool)

	for i := 0; i < count; i++ {
		row := types.NewRow(
			types.MRP("id", i),
			types.MRP("name", "foobar-"+strconv.Itoa(i)),
		)

		if test {
			row.Set("test", rand.Intn(100)+1)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
```

Certainly! Here's a concise description that incorporates the `NewRow*` methods for creating rows:

---

## 2.7 **Creating and Managing Rows in `GlazeCommand` with `types.Row`**

Within the `glazed` framework, `types.Row` represents an ordered map, ensuring consistent field order in rows.

### **Creating Rows with `types.Row` and `MRP`**

- **Direct Initialization**: Use `NewRow` with the `MRP` (MapRowPair) function to directly initialize rows while maintaining the field order:

  ```go
  row := types.NewRow(
      types.MRP("id", i),
      types.MRP("name", "foobar-"+strconv.Itoa(i)),
  )
  ```

  `MRP` is a utility function to quickly create a key-value pair for a row.

### **Other Methods for Creating Rows**

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

## 3. **Integrating `glazed` Commands with Cobra**

Now, the core part: adding your `glazed` command to the Cobra root command.
This requires converting the `glazed` command into a Cobra command using the provided utility functions.

```go
func registerCommands(rootCmd *cobra.Command) {
	// For a GlazeCommand
	glazeCmdInstance, err := NewMyGlazeCommand() // Assuming you've created this function
	cobra.CheckErr(err)

	command, err := BuildCobraCommandFromGlazeCommand(glazeCmdInstance)
	cobra.CheckErr(err)

	rootCmd.AddCommand(command)
	
	// Similarly, for WriterCommand or BareCommand
	// writerCmdInstance, err := NewMyWriterCommand()
	// command, err = BuildCobraCommandFromWriterCommand(writerCmdInstance)
	// rootCmd.AddCommand(command)
}
```

