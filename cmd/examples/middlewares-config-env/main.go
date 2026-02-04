package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// DemoSettings maps to the demo section fields
type DemoSettings struct {
	ApiKey    string `glazed:"api-key"`
	Threshold int    `glazed:"threshold"`
}

type DemoCommand struct {
	*cmds.CommandDescription
}

func NewDemoCommand() (*DemoCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	demoSection, err := schema.NewSection(
		"demo",
		"Demo settings",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New(
				"api-key",
				fields.TypeString,
				fields.WithHelp("API key loaded from config/env/flags"),
			),
			fields.New(
				"threshold",
				fields.TypeInteger,
				fields.WithDefault(10),
				fields.WithHelp("Numeric threshold"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"demo",
		cmds.WithShort("Demonstrate config/env/flags middlewares"),
		cmds.WithSections(glazedSection, demoSection),
	)

	return &DemoCommand{CommandDescription: desc}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &DemoCommand{}

func (c *DemoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &DemoSettings{}
	if err := vals.DecodeSectionInto("demo", settings); err != nil {
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
			// AppName enables env prefix APP_<SECTION_PREFIX+FLAG>
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
