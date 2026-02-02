package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// Note: This example demonstrates manual middleware execution using the sources package.

// Settings struct for the command
type ConfigSettings struct {
	ApiKey  string `glazed.parameter:"api-key"`
	Timeout int    `glazed.parameter:"timeout"`
	Debug   bool   `glazed.parameter:"debug"`
}

// BareCommand that uses sources package for manual middleware execution
type SourcesExampleCommand struct {
	*cmds.CommandDescription
}

func NewSourcesExampleCommand() (*SourcesExampleCommand, error) {
	// Create a simple section with a few fields
	configSection, err := schema.NewSection(
		"config",
		"Config",
		schema.WithPrefix("config-"),
		schema.WithDescription("Configuration settings"),
		schema.WithFields(
			fields.New("api-key", fields.TypeString,
				fields.WithHelp("API key for authentication"),
				fields.WithDefault("default-key"),
			),
			fields.New("timeout", fields.TypeInteger,
				fields.WithHelp("Request timeout in seconds"),
				fields.WithDefault(30),
			),
			fields.New("debug", fields.TypeBool,
				fields.WithHelp("Enable debug mode"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create schema
	schema_ := schema.NewSchema(
		schema.WithSections(configSection),
	)

	// Create command definition
	desc := cmds.NewCommandDefinition(
		"sources-example",
		cmds.WithShort("Example demonstrating sources package for manual middleware execution"),
		cmds.WithLong(`This example shows how to use the sources package to manually
execute middleware chains for value resolution from multiple sources:
- Environment variables (APP_CONFIG_API_KEY=env-key)
- Custom map (programmatic overrides)
- Config file (JSON/YAML)
- Cobra flags (--config-api-key=flag-key)

Precedence order (lowest to highest):
1. Defaults
2. Config file
3. Custom map
4. Environment variables
5. Cobra flags`),
		cmds.WithSchema(schema_),
	)

	return &SourcesExampleCommand{CommandDescription: desc}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &SourcesExampleCommand{}

func (c *SourcesExampleCommand) Run(ctx context.Context, vals *values.Values) error {
	// Decode settings from resolved values using the new API
	settings := &ConfigSettings{}
	if err := values.DecodeSectionInto(vals, "config", settings); err != nil {
		return fmt.Errorf("failed to decode config settings: %w", err)
	}

	// Use the resolved settings
	fmt.Printf("Resolved Configuration:\n")
	fmt.Printf("  API Key: %s\n", settings.ApiKey)
	fmt.Printf("  Timeout: %d seconds\n", settings.Timeout)
	fmt.Printf("  Debug: %v\n", settings.Debug)

	return nil
}

func main() {
	root := &cobra.Command{
		Use:   "sources-example",
		Short: "Example of manual middleware execution with sources package",
	}

	// Create command
	cmd, err := NewSourcesExampleCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Get the schema from the command
	// cmd.Layers is *schema.Schema, but we convert to schema.Schema (type alias)
	// to use the new API vocabulary
	cmdSchema := (*schema.Schema)(cmd.Layers)

	// Register flags manually
	root.Flags().String("config-file", "", "Config file path")
	root.Flags().String("config-api-key", "", "API key")
	root.Flags().Int("config-timeout", 0, "Timeout")
	root.Flags().Bool("config-debug", false, "Debug mode")

	// Parse command line
	if err := root.ParseFlags(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Get config file path from flag (if set)
	configFile, _ := root.Flags().GetString("config-file")

	// Create empty values collection
	vals := values.New()

	// Build middleware chain with sources package
	// Ordering: first middleware has highest precedence (runs last); last middleware has lowest precedence (runs first).
	// So we pass: flags > env > map > config-file > defaults.
	middlewares := []sources.Middleware{
		sources.FromCobra(root, sources.WithSource("flags")),
		sources.FromEnv("APP", sources.WithSource("env")),
		sources.FromMap(map[string]map[string]interface{}{
			"config": {
				"api-key": "custom-map-key",
				"timeout": 60,
			},
		}, sources.WithSource("custom-map")),
	}
	if configFile != "" {
		middlewares = append(middlewares,
			sources.FromFile(configFile,
				sources.WithParseOptions(
					sources.WithSource("config-file"),
				),
			),
		)
	}
	middlewares = append(middlewares,
		sources.FromDefaults(sources.WithSource("defaults")),
	)

	// Execute middleware chain using sources.Execute with new API types
	// sources.Execute accepts schema.Schema and values.Values (the new API vocabulary)
	if err := sources.Execute(cmdSchema, vals, middlewares...); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing middlewares: %v\n", err)
		os.Exit(1)
	}

	// Run the command with resolved values
	if err := cmd.Run(context.Background(), vals); err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}
