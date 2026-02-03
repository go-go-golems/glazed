package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type StatusSettings struct {
	Verbose bool `glazed:"verbose"`
}

type StatusCommand struct {
	*cmds.CommandDescription
}

func NewStatusCommand() (*StatusCommand, error) {
	desc := cmds.NewCommandDescription(
		"status",
		cmds.WithShort("Show status (dual-mode)"),
		cmds.WithLong(`Demonstrates dual-mode commands:
- classic mode (BareCommand)
- structured mode (GlazeCommand via --with-glaze-output)`),
		cmds.WithFlags(
			fields.New(
				"verbose",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Show additional details"),
				fields.WithShortFlag("v"),
			),
		),
	)

	return &StatusCommand{CommandDescription: desc}, nil
}

var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}

func (c *StatusCommand) Run(ctx context.Context, vals *values.Values) error {
	settings := &StatusSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to decode settings")
	}

	fmt.Println("System Status:")
	fmt.Println("  Status: Healthy")
	if settings.Verbose {
		fmt.Println("  Updated:", time.Now().Format(time.RFC3339))
		fmt.Println("  Version: 1.0.0")
	}

	return nil
}

func (c *StatusCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &StatusSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to decode settings")
	}

	row := types.NewRow(
		types.MRP("status", "healthy"),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)
	if settings.Verbose {
		row.Set("version", "1.0.0")
	}

	return gp.AddRow(ctx, row)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "new-api-dual-mode",
		Short: "New API example: dual-mode command",
	}

	statusCmd, err := NewStatusCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraStatusCmd, err := cli.BuildCobraCommand(statusCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraStatusCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
