package main

import (
	"context"
	"fmt"
	"os"

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

// AppSettings maps to the app section parameters
type AppSettings struct {
	Verbose bool   `glazed:"verbose"`
	Port    int    `glazed:"port"`
	Host    string `glazed:"host"`
}

// OutputSettings maps to the output section parameters
type OutputSettings struct {
	Format string `glazed:"format"`
	Pretty bool   `glazed:"pretty"`
}

// DefaultSettings maps to the default section (positional args)
type DefaultSettings struct {
	InputFile string `glazed:"input-file"`
}

type RefactorDemoCommand struct {
	*cmds.CommandDescription
}

func NewRefactorDemoCommand() (*RefactorDemoCommand, error) {
	// Create glazed schema section for output formatting
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	// Create app section with prefix "app-"
	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithPrefix("app-"),
		schema.WithDescription("Application configuration settings"),
		schema.WithFields(
			fields.New("verbose", fields.TypeBool,
				fields.WithHelp("Enable verbose logging"),
				fields.WithDefault(false),
			),
			fields.New("port", fields.TypeInteger,
				fields.WithHelp("Server port number"),
				fields.WithDefault(8080),
			),
			fields.New("host", fields.TypeString,
				fields.WithHelp("Server host address"),
				fields.WithDefault("localhost"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create output section with prefix "output-"
	outputSection, err := schema.NewSection(
		"output",
		"Output",
		schema.WithPrefix("output-"),
		schema.WithDescription("Output formatting settings"),
		schema.WithFields(
			fields.New("format", fields.TypeChoice,
				fields.WithHelp("Output format"),
				fields.WithChoices("json", "yaml", "table"),
				fields.WithDefault("table"),
			),
			fields.New("pretty", fields.TypeBool,
				fields.WithHelp("Pretty print output"),
				fields.WithDefault(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create default section for positional arguments
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("input-file", fields.TypeString,
				fields.WithHelp("Input file to process"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create schema collection
	schema := schema.NewSchema(
		schema.WithSections(glazedSection, appSection, outputSection, defaultSection),
	)

	desc := cmds.NewCommandDefinition(
		"refactor-demo",
		cmds.WithShort("Demonstrate new wrapper packages (schema/fields/values/sources)"),
		cmds.WithLong(`This example demonstrates the new wrapper packages:
- schema: for defining schema sections
- fields: for defining field definitions
- values: for resolved values and decoding
- sources: for value resolution from env/cobra/defaults

Precedence order (lowest to highest):
1. Defaults (from field definitions)
2. Environment variables (DEMO_APP_VERBOSE=true)
3. Cobra flags (--app-verbose=true)

Example usage:
  DEMO_APP_VERBOSE=true go run ./cmd/examples/refactor-new-packages refactor-demo input.txt
  DEMO_APP_VERBOSE=true go run ./cmd/examples/refactor-new-packages refactor-demo --app-verbose=false input.txt
`),
		cmds.WithSchema(schema),
	)

	return &RefactorDemoCommand{CommandDescription: desc}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &RefactorDemoCommand{}

func (c *RefactorDemoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {

	// Decode each section into its struct
	appSettings := &AppSettings{}
	if err := vals.DecodeSectionInto("app", appSettings); err != nil {
		return fmt.Errorf("failed to decode app settings: %w", err)
	}

	outputSettings := &OutputSettings{}
	if err := vals.DecodeSectionInto("output", outputSettings); err != nil {
		return fmt.Errorf("failed to decode output settings: %w", err)
	}

	defaultSettings := &DefaultSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, defaultSettings); err != nil {
		return fmt.Errorf("failed to decode default settings: %w", err)
	}

	// Create a row showing all resolved values
	row := types.NewRow(
		types.MRP("app_verbose", appSettings.Verbose),
		types.MRP("app_port", appSettings.Port),
		types.MRP("app_host", appSettings.Host),
		types.MRP("output_format", outputSettings.Format),
		types.MRP("output_pretty", outputSettings.Pretty),
		types.MRP("input_file", defaultSettings.InputFile),
	)

	return gp.AddRow(ctx, row)
}

func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "refactor-demo",
		Short: "Demo of new wrapper packages",
	}
	return root
}

func main() {
	root := buildRoot()

	demoCmd, err := NewRefactorDemoCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraDemoCmd, err := cli.BuildCobraCommandFromCommand(
		demoCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			// AppName enables env prefix DEMO_<LAYER_PREFIX+FLAG>
			// Example: DEMO_APP_VERBOSE=true sets app.verbose
			AppName: "demo",
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	root.AddCommand(cobraDemoCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
