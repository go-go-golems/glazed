package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// DemoSettings maps to the demo layer parameters
type DemoSettings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type DemoCommand struct {
	*cmds.CommandDescription
}

func NewDemoCommand() (*DemoCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	demoLayer, err := layers.NewParameterLayer(
		"demo",
		"Demo settings",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"api-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("API key loaded from config/env/flags"),
			),
			parameters.NewParameterDefinition(
				"threshold",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(10),
				parameters.WithHelp("Numeric threshold"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"demo",
		cmds.WithShort("Demonstrate config/env/flags middlewares"),
		cmds.WithLayersList(glazedLayer, demoLayer),
	)

	return &DemoCommand{CommandDescription: desc}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &DemoCommand{}

func (c *DemoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DemoSettings{}
	if err := parsedLayers.InitializeStruct("demo", settings); err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("api_key", settings.ApiKey),
		types.MRP("threshold", settings.Threshold),
	)
	return gp.AddRow(ctx, row)
}

func buildRoot() *cobra.Command {
	root := &cobra.Command{Use: "glazed-mw-demo"}
	return root
}

func main() {
	root := buildRoot()

	demoCmd, err := NewDemoCommand()
	if err != nil {
		panic(err)
	}

	cobraDemoCmd, err := cli.BuildCobraCommandFromCommand(
		demoCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			// AppName enables env prefix APP_<LAYER_PREFIX+FLAG>
			AppName: "glazed-mw-demo",
			// Explicit config file for demo
			ConfigPath: "cmd/examples/middlewares-config-env/config.yaml",
		}),
	)
	if err != nil {
		panic(err)
	}
	root.AddCommand(cobraDemoCmd)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
