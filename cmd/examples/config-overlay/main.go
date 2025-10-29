package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmdmw "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

type OverlaySettings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type OverlayCommand struct{ *cmds.CommandDescription }

func NewOverlayCommand() (*OverlayCommand, error) {
	demo, err := layers.NewParameterLayer(
		"demo",
		"Overlay demo",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString, parameters.WithHelp("API key")),
			parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger, parameters.WithDefault(10), parameters.WithHelp("Threshold")),
		),
	)
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription("overlay", cmds.WithShort("Multiple config overlays"), cmds.WithLayersList(demo))
	return &OverlayCommand{desc}, nil
}

var _ cmds.BareCommand = &OverlayCommand{}

func (c *OverlayCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &OverlaySettings{}
	if err := pl.InitializeStruct("demo", s); err != nil {
		return err
	}
	fmt.Printf("api_key=%s threshold=%d\n", s.ApiKey, s.Threshold)
	return nil
}

func main() {
	root := &cobra.Command{Use: "config-overlay"}
	overlay, err := NewOverlayCommand()
	if err != nil {
		panic(err)
	}

	// Provide a custom files resolver returning low->high precedence files
	filesResolver := func(_ *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
		return []string{
			"cmd/examples/config-overlay/base.yaml",
			"cmd/examples/config-overlay/env.yaml",
			"cmd/examples/config-overlay/local.yaml",
		}, nil
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		overlay,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsLayer: true,
			ConfigFilesFunc:          filesResolver,
		}),
	)
	if err != nil {
		panic(err)
	}

	// For demonstration, add an extra command to show parsed steps by printing the map
	// Users can also run with env overrides: e.g. DEMO_API_KEY and demo threshold flags
	_ = cmdmw.UpdateFromEnv // ensure middleware is referenced (no-op here)

	root.AddCommand(cobraCmd)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
