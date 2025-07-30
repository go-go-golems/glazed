package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// BareCommand example - simple command that outputs directly
type ExampleBareCommand struct {
	*cmds.CommandDescription
}

func NewExampleBareCommand() (*ExampleBareCommand, error) {
	return &ExampleBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"bare",
			cmds.WithShort("Example bare command"),
			cmds.WithLong("A simple bare command that outputs text directly"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"message",
					parameters.ParameterTypeString,
					parameters.WithDefault("Hello from bare command!"),
					parameters.WithHelp("Message to display"),
				),
			),
		),
	}, nil
}

func (c *ExampleBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := struct {
		Message string `glazed.parameter:"message"`
	}{}

	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s)
	if err != nil {
		return err
	}

	fmt.Println("BARE COMMAND:", s.Message)
	return nil
}

// WriterCommand example - outputs to a writer
type ExampleWriterCommand struct {
	*cmds.CommandDescription
}

func NewExampleWriterCommand() (*ExampleWriterCommand, error) {
	return &ExampleWriterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"writer",
			cmds.WithShort("Example writer command"),
			cmds.WithLong("A writer command that outputs to a specified writer"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(3),
					parameters.WithHelp("Number of lines to output"),
				),
			),
		),
	}, nil
}

func (c *ExampleWriterCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := struct {
		Count int `glazed.parameter:"count"`
	}{}

	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s)
	if err != nil {
		return err
	}

	for i := 0; i < s.Count; i++ {
		fmt.Fprintf(w, "Writer command output line %d\n", i+1)
	}
	return nil
}

// GlazeCommand example - outputs structured data
type ExampleGlazeCommand struct {
	*cmds.CommandDescription
}

func NewExampleGlazeCommand() (*ExampleGlazeCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	glazedLayers := layers.NewParameterLayers()
	glazedLayers.Set(settings.GlazedSlug, glazedParameterLayer)

	return &ExampleGlazeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"glaze",
			cmds.WithShort("Example glaze command"),
			cmds.WithLong("A glaze command that outputs structured data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"rows",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(2),
					parameters.WithHelp("Number of data rows to output"),
				),
			),
			cmds.WithLayers(glazedLayers),
		),
	}, nil
}

func (c *ExampleGlazeCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := struct {
		Rows int `glazed.parameter:"rows"`
	}{}

	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s)
	if err != nil {
		return err
	}

	for i := 0; i < s.Rows; i++ {
		row := types.NewRow(
			types.MRP("id", i+1),
			types.MRP("name", fmt.Sprintf("Item %d", i+1)),
			types.MRP("value", (i+1)*10),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// DualCommand example - implements both bare/writer and glaze modes
type ExampleDualCommand struct {
	*cmds.CommandDescription
}

func NewExampleDualCommand() (*ExampleDualCommand, error) {
	return &ExampleDualCommand{
		CommandDescription: cmds.NewCommandDescription(
			"dual",
			cmds.WithShort("Example dual command"),
			cmds.WithLong("A dual command that can run in both classic and glaze modes"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithDefault("World"),
					parameters.WithHelp("Name to greet"),
				),
				parameters.NewParameterDefinition(
					"times",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(1),
					parameters.WithHelp("Number of greetings"),
				),
			),
		),
	}, nil
}

// Implement BareCommand interface for classic mode
func (c *ExampleDualCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := struct {
		Name  string `glazed.parameter:"name"`
		Times int    `glazed.parameter:"times"`
	}{}

	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s)
	if err != nil {
		return err
	}

	for i := 0; i < s.Times; i++ {
		fmt.Printf("Hello, %s! (greeting %d)\n", s.Name, i+1)
	}
	return nil
}

// Implement GlazeCommand interface for glaze mode
func (c *ExampleDualCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := struct {
		Name  string `glazed.parameter:"name"`
		Times int    `glazed.parameter:"times"`
	}{}

	err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s)
	if err != nil {
		return err
	}

	for i := 0; i < s.Times; i++ {
		row := types.NewRow(
			types.MRP("greeting_num", i+1),
			types.MRP("name", s.Name),
			types.MRP("message", fmt.Sprintf("Hello, %s!", s.Name)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "register-cobra",
		Short: "Example showing different command registration approaches",
	}

	// Example 1: Bare command using traditional builder
	bareCmd, err := NewExampleBareCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bare command")
	}
	cobraBareCmd, err := cli.BuildCobraCommandFromBareCommand(bareCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra bare command")
	}
	rootCmd.AddCommand(cobraBareCmd)

	// Example 2: Writer command using traditional builder
	writerCmd, err := NewExampleWriterCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create writer command")
	}
	cobraWriterCmd, err := cli.BuildCobraCommandFromWriterCommand(writerCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra writer command")
	}
	rootCmd.AddCommand(cobraWriterCmd)

	// Example 3: Glaze command using traditional builder
	glazeCmd, err := NewExampleGlazeCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glaze command")
	}
	cobraGlazeCmd, err := cli.BuildCobraCommandFromGlazeCommand(glazeCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra glaze command")
	}
	rootCmd.AddCommand(cobraGlazeCmd)

	// Example 4: Dual command using new dual mode builder
	dualCmd, err := NewExampleDualCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dual command")
	}
	cobraDualCmd, err := cli.BuildCobraCommandFromCommand(
		dualCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra dual command")
	}
	rootCmd.AddCommand(cobraDualCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
