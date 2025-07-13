package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
)

// Example 8: Using the Built-in Glazed Layer
// This demonstrates the Glazed Layer for output formatting capabilities

func NewLoggingLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"logging",
		"Logging Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("debug", "info", "warn", "error"),
				parameters.WithDefault("info"),
				parameters.WithHelp("Logging level"),
			),
		),
	)
}

func createDataExportCommand() (*cmds.CommandDescription, error) {
	// Create custom logging layer
	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	// Create glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return cmds.NewCommandDescription(
		"export-data",
		cmds.WithShort("Export data with flexible formatting"),
		cmds.WithLong("Export data from the database with various output formats and filtering options"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"source",
				parameters.ParameterTypeString,
				parameters.WithHelp("Data source to export from"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Query to execute"),
				parameters.WithDefault("SELECT * FROM users"),
			),
		),
		cmds.WithLayersList(loggingLayer, glazedLayer),
	), nil
}

func main() {
	fmt.Println("=== Testing Built-in Glazed Layer ===")

	// Create command with Glazed layer
	cmd, err := createDataExportCommand()
	if err != nil {
		log.Fatalf("Failed to create command: %v", err)
	}

	fmt.Printf("Command: %s\n", cmd.Name)
	fmt.Printf("Short: %s\n", cmd.Short)
	fmt.Printf("Long: %s\n", cmd.Long)

	fmt.Printf("\nLayers (%d total):\n", cmd.Layers.Len())
	cmd.Layers.ForEach(func(slug string, layer layers.ParameterLayer) {
		fmt.Printf("  - %s: %s\n", slug, layer.GetName())

		// Show some sample parameters from each layer
		params := layer.GetParameterDefinitions()
		fmt.Printf("    Parameters (%d total)\n", params.Len())

		// Show first few parameters as examples
		count := 0
		params.ForEach(func(param *parameters.ParameterDefinition) {
			if count < 5 { // Show only first 5 parameters to avoid cluttering
				fmt.Printf("      - %s (%s)", param.Name, param.Type)
				if param.Default != nil {
					fmt.Printf(" [default: %v]", *param.Default)
				}
				if len(param.Choices) > 0 {
					fmt.Printf(" [choices: %v]", param.Choices)
				}
				fmt.Printf(" - %s\n", param.Help)
				count++
			}
		})
		if params.Len() > 5 {
			fmt.Printf("      ... and %d more parameters\n", params.Len()-5)
		}
		fmt.Println()
	})

	// Test parameter parsing with glazed layer
	fmt.Println("=== Testing Glazed Layer Parameter Resolution ===")

	parsedLayers, err := cmd.Layers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize parsed layers: %v", err)
	}

	// Check some common glazed parameters
	fmt.Println("Glazed layer default settings:")
	if glazedLayer, ok := parsedLayers.Get("glazed"); ok {
		// Check common output parameters
		if output, ok := glazedLayer.GetParameter("output"); ok {
			fmt.Printf("  Output format: %v\n", output)
		}

		if fields, ok := glazedLayer.GetParameter("fields"); ok {
			fmt.Printf("  Fields filter: %v\n", fields)
		}

		if sortColumns, ok := glazedLayer.GetParameter("sort-columns"); ok {
			fmt.Printf("  Sort columns: %v\n", sortColumns)
		}

		if outputFile, ok := glazedLayer.GetParameter("output-file"); ok {
			fmt.Printf("  Output file: %v\n", outputFile)
		}
	}

	// Check logging layer
	fmt.Println("Logging layer default settings:")
	if loggingLayer, ok := parsedLayers.Get("logging"); ok {
		if logLevel, ok := loggingLayer.GetParameter("log-level"); ok {
			fmt.Printf("  Log level: %v\n", logLevel)
		}
	}

	// Show default layer (command-specific flags)
	fmt.Println("Default layer (command-specific) settings:")
	if defaultLayer, ok := parsedLayers.Get("default"); ok {
		if source, ok := defaultLayer.GetParameter("source"); ok {
			fmt.Printf("  Source: %v\n", source)
		}
		if query, ok := defaultLayer.GetParameter("query"); ok {
			fmt.Printf("  Query: %v\n", query)
		}
	}

	fmt.Println("\n=== Glazed Layer Features Demonstration ===")
	fmt.Println("The Glazed layer provides comprehensive output formatting:")
	fmt.Println("- Multiple output formats (table, csv, json, yaml, etc.)")
	fmt.Println("- Field filtering and selection")
	fmt.Println("- Column sorting and ordering")
	fmt.Println("- Row filtering with expressions")
	fmt.Println("- Template-based output transformation")
	fmt.Println("- File output capabilities")
	fmt.Println("- Data manipulation with jq-style queries")
	fmt.Println("- Pagination with skip/limit options")
	fmt.Println()

	fmt.Println("Example usage:")
	fmt.Println("  myapp export-data --source users.db --output json --fields name,email --sort-columns name")
	fmt.Println("  myapp export-data --source users.db --output csv --output-file users.csv")
	fmt.Println("  myapp export-data --source users.db --output table --filter 'age > 18'")
}
