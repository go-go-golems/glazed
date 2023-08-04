---
Title: Creating a simple CLI application with glaze
Slug: a-simple-table-cli
Topics:
- glaze
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

The goal is to create a command called `ExampleCommand` that generates a specified number of rows of data. Each row will
have an `id` and `name` field, and optionally a `test` field if a flag is set. The rows are then added to a `Processor`
for output.

## 2. APIs Used

- `CommandDescription`: This struct is used to describe the command. It includes the name of the command, a short
  description, and any flags or arguments the command takes.
- `ParameterDefinition`: This struct is used to define the parameters (flags or arguments) that the command takes. It
  includes the name of the parameter, the type, and any default value.
- `Command` and `GlazeCommand`: These interfaces define the methods that a command must implement. The `Run` method is
  where the main functionality of the command is implemented.
- `Row`: This struct represents a row of data. It can have any number of fields, each with a name and value.

## 3. Creating the Command

First, we create a new struct that embeds the `Command` interface. We also add a `description` field of
type `*CommandDescription`.

```go
type ExampleCommand struct {
	description *cmds.CommandDescription
}
```

Next, we create a constructor function for `ExampleCommand`. In this function, we initialize the `description` field
using the `NewCommandDescription` function, passing in the name of the command, a short description, and any flags or
arguments the command takes.

```go
func NewExampleCommand() (*ExampleCommand, error) {
	return &ExampleCommand{
		description: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Example command"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of rows to output"),
					parameters.WithDefault(10),
				),
				parameters.NewParameterDefinition(
					"test",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Whether to add a test column"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}
```

## 4. Implementing the Run Method
The `Run` method is where the main functionality of the command is implemented. In this method, we first retrieve the values of the `count` and `test` flags. Then, we generate `count` number of rows, each with an `id` and `name` field. If `test` is true, we also add a `test` field with a random integer between 1 and 100. Finally, we add the rows to the `Processor` for output.

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

## 5. Extending the Command
There are many ways you could extend this command. Here are a few ideas:
- Add more flags to customize the output, such as a flag to set the prefix of the `name` field.
- Add error handling to ensure the `count` flag is a positive integer.
- Add a flag to specify the range of the random integer for the `test` field.
- Instead of generating random data, read data from a file (see the types StringFromFile and StringListFromFile)
