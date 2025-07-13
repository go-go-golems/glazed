package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ExampleCommand demonstrates programmatic execution from the documentation
type ExampleCommand struct {
	*cmds.CommandDescription
}

// ExampleSettings mirrors the command parameters
type ExampleSettings struct {
	Count  int    `glazed.parameter:"count"`
	Format string `glazed.parameter:"format"`
}

func (c *ExampleCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ExampleSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	fmt.Printf("Running example command with count=%d, format=%s\n", s.Count, s.Format)

	// Generate sample data
	for i := 0; i < s.Count; i++ {
		row := types.NewRow(
			types.MRP("id", i+1),
			types.MRP("name", fmt.Sprintf("item-%d", i+1)),
			types.MRP("format", s.Format),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewExampleCommand creates a new example command
func NewExampleCommand() (*ExampleCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"example",
		cmds.WithShort("Example command for programmatic execution"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"count",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(3),
				parameters.WithHelp("Number of items to generate"),
			),
			parameters.NewParameterDefinition(
				"format",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("json", "yaml", "table"),
				parameters.WithDefault("table"),
				parameters.WithHelp("Output format"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &ExampleCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ExampleCommand{}

func main() {
	// Demonstrate the programmatic execution example from the documentation
	fmt.Println("=== Glazed Programmatic Execution Demo ===\n")

	// Create command instance
	cmd, err := NewExampleCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}
	fmt.Println("✓ Created command instance")

	// Set up execution context
	ctx := context.Background()
	fmt.Println("✓ Set up execution context")

	// Demo 1: Basic execution with explicit values
	fmt.Println("\n--- Demo 1: Basic execution with explicit values ---")
	parseOptions := []runner.ParseOption{
		runner.WithValuesForLayers(map[string]map[string]interface{}{
			"default": {
				"count":  5,
				"format": "json",
			},
		}),
	}

	runOptions := []runner.RunOption{
		runner.WithWriter(os.Stdout),
	}

	fmt.Println("Running with count=5, format=json:")
	err = runner.ParseAndRun(ctx, cmd, parseOptions, runOptions)
	if err != nil {
		log.Fatalf("Error running command: %v", err)
	}

	// Demo 2: With environment variables
	fmt.Println("\n--- Demo 2: With environment variables ---")
	os.Setenv("EXAMPLE_COUNT", "3")
	os.Setenv("EXAMPLE_FORMAT", "yaml")

	parseOptions2 := []runner.ParseOption{
		runner.WithEnvMiddleware("EXAMPLE_"),
	}

	runOptions2 := []runner.RunOption{
		runner.WithWriter(os.Stdout),
	}

	fmt.Println("Running with environment variables EXAMPLE_COUNT=3, EXAMPLE_FORMAT=yaml:")
	err = runner.ParseAndRun(ctx, cmd, parseOptions2, runOptions2)
	if err != nil {
		log.Fatalf("Error running command: %v", err)
	}

	// Demo 3: Multiple parameter sources (priority order demonstration)
	fmt.Println("\n--- Demo 3: Multiple parameter sources (priority order) ---")

	// Set environment variables (lower priority)
	os.Setenv("EXAMPLE_COUNT", "10")
	os.Setenv("EXAMPLE_FORMAT", "table")

	parseOptions3 := []runner.ParseOption{
		// Load from environment with prefix (lower priority)
		runner.WithEnvMiddleware("EXAMPLE_"),

		// Set explicit values (higher priority - will override env vars)
		runner.WithValuesForLayers(map[string]map[string]interface{}{
			"default": {
				"count": 2, // This will override EXAMPLE_COUNT=10
				// format will come from environment (table)
			},
		}),
	}

	runOptions3 := []runner.RunOption{
		runner.WithWriter(os.Stdout),
	}

	fmt.Println("Environment: COUNT=10, FORMAT=table")
	fmt.Println("Explicit values: count=2 (will override env)")
	fmt.Println("Result should show count=2, format=table:")
	err = runner.ParseAndRun(ctx, cmd, parseOptions3, runOptions3)
	if err != nil {
		log.Fatalf("Error running command: %v", err)
	}

	// Demo 4: Different output formats
	fmt.Println("\n--- Demo 4: Different output formats ---")

	formats := []string{"table", "json", "yaml", "csv"}
	for _, format := range formats {
		fmt.Printf("\nOutput format: %s\n", format)
		fmt.Println("---")

		parseOptions := []runner.ParseOption{
			runner.WithValuesForLayers(map[string]map[string]interface{}{
				"default": {
					"count":  2,
					"format": format,
				},
				"glazed": {
					"output": format,
				},
			}),
		}

		runOptions := []runner.RunOption{
			runner.WithWriter(os.Stdout),
		}

		err = runner.ParseAndRun(ctx, cmd, parseOptions, runOptions)
		if err != nil {
			log.Printf("Error with format %s: %v", format, err)
		}
		fmt.Println()
	}

	fmt.Println("\n=== Demo completed successfully! ===")
	fmt.Println("\nKey points demonstrated:")
	fmt.Println("1. Commands can be run programmatically without Cobra")
	fmt.Println("2. Parameters can be loaded from multiple sources")
	fmt.Println("3. Environment variables work with prefixes")
	fmt.Println("4. Explicit values override environment variables")
	fmt.Println("5. Different output formats are handled automatically")

	// Clean up environment
	os.Unsetenv("EXAMPLE_COUNT")
	os.Unsetenv("EXAMPLE_FORMAT")
}
