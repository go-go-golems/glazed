package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// Test extended parameter types from the tutorial
type ExtendedParamsCommand struct {
	*cmds.CommandDescription
}

type ExtendedParamsSettings struct {
	ConfigFile string    `glazed.parameter:"config-file"`
	Format     string    `glazed.parameter:"format"`
	Date       time.Time `glazed.parameter:"date"`
}

func (c *ExtendedParamsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ExtendedParamsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	row := types.NewRow(
		types.MRP("config_file", settings.ConfigFile),
		types.MRP("format", settings.Format),
		types.MRP("date", settings.Date.Format("2006-01-02")),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	return gp.AddRow(ctx, row)
}

func NewExtendedParamsCommand() (*ExtendedParamsCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"test-params",
		cmds.WithShort("Test extended parameter types"),
		cmds.WithLong(`Test file, choice, and duration parameter types from the tutorial.`),
		cmds.WithFlags(
			// File parameter - validates file exists
			parameters.NewParameterDefinition(
				"config-file",
				parameters.ParameterTypeFile,
				parameters.WithHelp("Configuration file path"),
			),

			// Choice parameter - limits valid options
			parameters.NewParameterDefinition(
				"format",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("json", "yaml", "xml"),
				parameters.WithDefault("json"),
				parameters.WithHelp("Output format"),
			),

			// Date parameter - parses date strings like "2023-01-15"
			parameters.NewParameterDefinition(
				"date",
				parameters.ParameterTypeDate,
				parameters.WithDefault("2023-01-15"),
				parameters.WithHelp("Date value"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &ExtendedParamsCommand{
		CommandDescription: cmdDesc,
	}, nil
}

var _ cmds.GlazeCommand = &ExtendedParamsCommand{}

func main() {
	rootCmd := &cobra.Command{
		Use:   "extended-params-test",
		Short: "Test extended parameter types",
	}

	testParamsCmd, err := NewExtendedParamsCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraTestParamsCmd, err := cli.BuildCobraCommandFromGlazeCommand(testParamsCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraTestParamsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
