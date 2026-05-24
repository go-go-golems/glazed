package a

import (
	"context"
	"flag"
	"io"
	"os"
	operating "os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func badEnv() string {
	return os.Getenv("PAGER") // want `use Glazed config/env middleware`
}

func badAliasedEnv() string {
	return operating.Getenv("HOME") // want `use Glazed config/env middleware`
}

func badGoFlagPackage() {
	_ = flag.String("config", "", "config file") // want `define CLI flags with cmds.WithFlags`
	flag.Parse()                                 // want `define CLI flags with cmds.WithFlags`
}

func badPFlagPackage() {
	_ = pflag.String("profile", "", "profile") // want `define CLI flags with cmds.WithFlags`
}

func badCobraFlagMethod() *cobra.Command {
	var address string
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&address, "address", ":8080", "listen address") // want `define CLI flags with cmds.WithFlags`
	cmd.PersistentFlags().Bool("verbose", false, "verbose")               // want `define CLI flags with cmds.WithFlags`
	return cmd
}

type TextCommand struct {
	*cmds.CommandDescription
}

func NewTextCommand() (*TextCommand, error) {
	glazedSection, _ := settings.NewGlazedSection()
	return &TextCommand{
		CommandDescription: cmds.NewCommandDescription(
			"text",
			cmds.WithSections(glazedSection), // want `exposes Glazed output flags but does not implement RunIntoGlazeProcessor`
		),
	}, nil
}

func (c *TextCommand) RunIntoWriter(ctx context.Context, parsed *values.Values, w io.Writer) error {
	return nil
}

type SchemaTextCommand struct {
	*cmds.CommandDescription
}

func NewSchemaTextCommand() (*SchemaTextCommand, error) {
	glazedSection, _ := settings.NewGlazedSchema()
	return &SchemaTextCommand{
		CommandDescription: cmds.NewCommandDescription(
			"schema-text",
			cmds.WithSections(glazedSection), // want `exposes Glazed output flags but does not implement RunIntoGlazeProcessor`
		),
	}, nil
}

type RowsCommand struct {
	*cmds.CommandDescription
}

func NewRowsCommand() (*RowsCommand, error) {
	glazedSection, _ := settings.NewGlazedSection()
	return &RowsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"rows",
			cmds.WithSections(glazedSection),
		),
	}, nil
}

func (c *RowsCommand) RunIntoGlazeProcessor(ctx context.Context, parsed *values.Values, gp middlewares.Processor) error {
	return nil
}

type PlainBareCommand struct {
	*cmds.CommandDescription
}

func NewPlainBareCommand() (*PlainBareCommand, error) {
	return &PlainBareCommand{
		CommandDescription: cmds.NewCommandDescription("plain"),
	}, nil
}
