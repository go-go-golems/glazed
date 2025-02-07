# Tutorial: Creating and Using a Custom Parameters Layer in Glazed

## 1. Creating a New Parameters Layer

A parameters layer in Glazed allows you to group related parameters together. Here's how to create a new one:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func NewCustomParameterLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "custom-layer",
        "Custom Layer Description",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "custom-flag",
                parameters.ParameterTypeString,
                parameters.WithHelp("A custom flag"),
                parameters.WithDefault("default-value"),
            ),
            parameters.NewParameterDefinition(
                "custom-int",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("A custom integer flag"),
                parameters.WithDefault(42),
            ),
        ),
    )
}
```

This creates a new parameter layer with a custom string flag and a custom integer flag.

## 2. Creating a Settings Struct

To easily access the parsed values of your custom layer, create a struct that maps to these parameters:

```go
type CustomSettings struct {
    CustomFlag string `glazed.parameter:"custom-flag"`
    CustomInt  int    `glazed.parameter:"custom-int"`
}
```

Note that the struct tags must exactly match the parameter names defined in your layer.

## 3. Parsing the Layer into the Settings Struct

When implementing the `Run` method of your command, you can parse the layer into your settings struct like this:

```go
func (c *MyCustomCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &CustomSettings{}
    if err := parsedLayers.InitializeStruct("custom-layer", s); err != nil {
        return err
    }

    // Now you can use s.CustomFlag and s.CustomInt
    // ...

    return nil
}
```

## 4. Adding the Custom Layer to a Glazed Command

To add your custom layer to a Glazed command, you'll need to include it when creating the command description:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds"
)

func NewMyCustomCommand() (*MyCustomCommand, error) {
    customLayer, err := NewCustomParameterLayer()
    if err != nil {
        return nil, err
    }

    return &MyCustomCommand{
        CommandDescription: cmds.NewCommandDescription(
            "my-custom-command",
            cmds.WithShort("A custom command example"),
            cmds.WithLayersList(
                customLayer,
                // You can add other layers here, like the glazed parameter layer
            ),
        ),
    }, nil
}
```

## 5. Implementing the Command

Now, let's put it all together in a complete example of a Glazed command that uses our custom layer:

```go
import (
    "context"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

type MyCustomCommand struct {
    *cmds.CommandDescription
}

func NewMyCustomCommand() (*MyCustomCommand, error) {
    customLayer, err := NewCustomParameterLayer()
    if err != nil {
        return nil, err
    }

    return &MyCustomCommand{
        CommandDescription: cmds.NewCommandDescription(
            "my-custom-command",
            cmds.WithShort("A custom command example"),
            cmds.WithLayersList(customLayer),
        ),
    }, nil
}

func (c *MyCustomCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &CustomSettings{}
    if err := parsedLayers.InitializeStruct("custom-layer", s); err != nil {
        return err
    }

    row := types.NewRow(
        types.MRP("custom-flag", s.CustomFlag),
        types.MRP("custom-int", s.CustomInt),
    )

    return gp.AddRow(ctx, row)
}
```

## 6. Integrating with Cobra

Finally, to use this command with Cobra, you can do the following:

```go
import (
    "github.com/spf13/cobra"
    "github.com/go-go-golems/glazed/pkg/cli"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "myapp",
        Short: "My application with custom Glazed command",
    }

    customCmd, err := NewMyCustomCommand()
    cobra.CheckErr(err)

    // Build the cobra command with short help for our custom layer
    cobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(
        customCmd,
        cli.WithCobraShortHelpLayers("custom-layer"),
    )
    cobra.CheckErr(err)

    rootCmd.AddCommand(cobraCmd)

    err = rootCmd.Execute()
    cobra.CheckErr(err)
}
```

The `WithCobraShortHelpLayers` option specifies which parameter layers should be shown in the short help output (when running `myapp my-custom-command -h`). This helps keep the help output clean and focused on the most important parameters, while still allowing users to see all parameters with `--help`.

This tutorial demonstrates how to create a custom parameters layer, parse it into a settings struct, and incorporate it into a Glazed command. By following these steps, you can extend the functionality of your Glazed-based CLI applications with custom parameter layers.
