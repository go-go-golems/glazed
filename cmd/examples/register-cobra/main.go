package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
				fields.New(
					"message",
					fields.TypeString,
					fields.WithDefault("Hello from bare command!"),
					fields.WithHelp("Message to display"),
				),
			),
		),
	}, nil
}

func (c *ExampleBareCommand) Run(ctx context.Context, vals *values.Values) error {
	s := struct {
		Message string `glazed:"message"`
	}{}

	err := vals.DecodeSectionInto(schema.DefaultSlug, &s)
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
				fields.New(
					"count",
					fields.TypeInteger,
					fields.WithDefault(3),
					fields.WithHelp("Number of lines to output"),
				),
			),
		),
	}, nil
}

func (c *ExampleWriterCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := struct {
		Count int `glazed:"count"`
	}{}

	err := vals.DecodeSectionInto(schema.DefaultSlug, &s)
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
	return &ExampleGlazeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"glaze",
			cmds.WithShort("Example glaze command"),
			cmds.WithLong("A glaze command that outputs structured data"),
			cmds.WithFlags(
				fields.New(
					"rows",
					fields.TypeInteger,
					fields.WithDefault(2),
					fields.WithHelp("Number of data rows to output"),
				),
			),
		),
	}, nil
}

func (c *ExampleGlazeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := struct {
		Rows int `glazed:"rows"`
	}{}

	err := vals.DecodeSectionInto(schema.DefaultSlug, &s)
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
				fields.New(
					"name",
					fields.TypeString,
					fields.WithDefault("World"),
					fields.WithHelp("Name to greet"),
				),
				fields.New(
					"times",
					fields.TypeInteger,
					fields.WithDefault(1),
					fields.WithHelp("Number of greetings"),
				),
			),
		),
	}, nil
}

// Implement BareCommand interface for classic mode
func (c *ExampleDualCommand) Run(ctx context.Context, vals *values.Values) error {
	s := struct {
		Name  string `glazed:"name"`
		Times int    `glazed:"times"`
	}{}

	err := vals.DecodeSectionInto(schema.DefaultSlug, &s)
	if err != nil {
		return err
	}

	for i := 0; i < s.Times; i++ {
		fmt.Printf("Hello, %s! (greeting %d)\n", s.Name, i+1)
	}
	return nil
}

// Implement GlazeCommand interface for glaze mode
func (c *ExampleDualCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := struct {
		Name  string `glazed:"name"`
		Times int    `glazed:"times"`
	}{}

	err := vals.DecodeSectionInto(schema.DefaultSlug, &s)
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
	cobraBareCmd, err := cli.BuildCobraCommand(bareCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra bare command")
	}
	rootCmd.AddCommand(cobraBareCmd)

	// Example 2: Writer command using traditional builder
	writerCmd, err := NewExampleWriterCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create writer command")
	}
	cobraWriterCmd, err := cli.BuildCobraCommand(writerCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra writer command")
	}
	rootCmd.AddCommand(cobraWriterCmd)

	// Example 3: Glaze command using traditional builder
	glazeCmd, err := NewExampleGlazeCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glaze command")
	}
	cobraGlazeCmd, err := cli.BuildCobraCommand(glazeCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra glaze command")
	}
	rootCmd.AddCommand(cobraGlazeCmd)

	// Example 4: Dual command using new dual mode builder
	dualCmd, err := NewExampleDualCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dual command")
	}
	cobraDualCmd, err := cli.BuildCobraCommand(dualCmd,
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
