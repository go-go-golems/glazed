package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 1: Simple Layer Creation
// This demonstrates the basic layer creation pattern from the documentation

func NewDatabaseLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"database",               // Layer identifier
		"Database Configuration", // Human-readable name
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"db-host",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Database host to connect to"),
			),
			parameters.NewParameterDefinition(
				"db-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5432),
				parameters.WithHelp("Database port (PostgreSQL default: 5432)"),
			),
			parameters.NewParameterDefinition(
				"db-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Database name (required for connection)"),
				parameters.WithRequired(true),
			),
		),
	)
}

func main() {
	fmt.Println("=== Testing Simple Layer Creation ===")

	// Create the database layer
	layer, err := NewDatabaseLayer()
	if err != nil {
		log.Fatalf("Failed to create database layer: %v", err)
	}

	fmt.Printf("Layer created successfully!\n")
	fmt.Printf("Slug: %s\n", layer.GetSlug())
	fmt.Printf("Name: %s\n", layer.GetName())

	// Get parameter definitions
	params := layer.GetParameterDefinitions()
	fmt.Printf("\nParameter definitions (%d total):\n", params.Len())

	params.ForEach(func(param *parameters.ParameterDefinition) {
		fmt.Printf("  %s:\n", param.Name)
		fmt.Printf("    Type: %s\n", param.Type)
		if param.Default != nil {
			fmt.Printf("    Default: %v\n", *param.Default)
		} else {
			fmt.Printf("    Default: <none>\n")
		}
		fmt.Printf("    Required: %v\n", param.Required)
		fmt.Printf("    Help: %s\n", param.Help)
		fmt.Println()
	})
}
