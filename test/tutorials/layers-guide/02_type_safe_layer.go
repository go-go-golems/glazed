package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 2: Type-Safe Layer with Settings Struct
// This demonstrates the structured approach from the documentation

// 1. Define settings struct
type DatabaseSettings struct {
	Host     string `glazed.parameter:"db-host"`
	Port     int    `glazed.parameter:"db-port"`
	Name     string `glazed.parameter:"db-name"`
	Username string `glazed.parameter:"db-username"`
	Password string `glazed.parameter:"db-password"`
	SSLMode  string `glazed.parameter:"db-ssl-mode"`
}

// 2. Create layer with parameter definitions
func NewDatabaseLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"database",
		"Database Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"db-host",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Database host"),
			),
			parameters.NewParameterDefinition(
				"db-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5432),
				parameters.WithHelp("Database port"),
			),
			parameters.NewParameterDefinition(
				"db-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Database name"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"db-username",
				parameters.ParameterTypeString,
				parameters.WithHelp("Database username"),
			),
			parameters.NewParameterDefinition(
				"db-password",
				parameters.ParameterTypeSecret, // Masked in output
				parameters.WithHelp("Database password"),
			),
			parameters.NewParameterDefinition(
				"db-ssl-mode",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("disable", "require", "verify-ca", "verify-full"),
				parameters.WithDefault("require"),
				parameters.WithHelp("SSL mode for database connection"),
			),
		),
	)
}

// 3. Helper function for settings extraction
func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
	settings := &DatabaseSettings{}
	err := parsedLayers.InitializeStruct("database", settings)
	return settings, err
}

func main() {
	fmt.Println("=== Testing Type-Safe Layer with Settings Struct ===")

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
		if len(param.Choices) > 0 {
			fmt.Printf("    Choices: %v\n", param.Choices)
		}
		fmt.Printf("    Help: %s\n", param.Help)
		fmt.Println()
	})

	// Test creating parsed layers with defaults
	fmt.Println("=== Testing Parameter Resolution ===")

	// Create parameter layers collection
	paramLayers := layers.NewParameterLayers(layers.WithLayers(layer))

	// Initialize with defaults
	parsedLayers, err := paramLayers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize parsed layers: %v", err)
	}

	// Extract settings using helper function
	settings, err := GetDatabaseSettings(parsedLayers)
	if err != nil {
		log.Fatalf("Failed to extract database settings: %v", err)
	}

	fmt.Printf("Extracted database settings:\n")
	fmt.Printf("  Host: %s\n", settings.Host)
	fmt.Printf("  Port: %d\n", settings.Port)
	fmt.Printf("  Name: %s\n", settings.Name)
	fmt.Printf("  Username: %s\n", settings.Username)
	fmt.Printf("  Password: %s\n", settings.Password)
	fmt.Printf("  SSLMode: %s\n", settings.SSLMode)
}
