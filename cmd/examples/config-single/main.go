package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

type DemoSettings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type DemoBareCommand struct {
	*cmds.CommandDescription
}

func NewDemoBareCommand() (*DemoBareCommand, error) {
	demoLayer, err := layers.NewParameterLayer(
		"demo",
		"Demo settings",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"api-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("API key from config/env/flags"),
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
		cmds.WithShort("Minimal custom layer with single config file"),
		cmds.WithLayersList(demoLayer),
	)

	return &DemoBareCommand{CommandDescription: desc}, nil
}

var _ cmds.BareCommand = &DemoBareCommand{}

func (c *DemoBareCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &DemoSettings{}
	if err := pl.InitializeStruct("demo", s); err != nil {
		return err
	}
	fmt.Printf("api_key=%s threshold=%d\n", s.ApiKey, s.Threshold)
	return nil
}

func main() {
	root := &cobra.Command{Use: "config-single"}

	demo, err := NewDemoBareCommand()
	if err != nil {
		panic(err)
	}

	// Use a single explicit config path for simplicity
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		demo,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsLayer: true,
			// Adjust path to your environment if needed
			ConfigPath: "cmd/examples/config-single/config.yaml",
		}),
	)
	if err != nil {
		panic(err)
	}
	root.AddCommand(cobraCmd)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
