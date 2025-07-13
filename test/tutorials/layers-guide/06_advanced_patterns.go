package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 6: Advanced Patterns - Layer Inheritance and Composition
// This demonstrates extending existing layers without modification

// Base database layer
func NewBaseDatabaseLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"database",
		"Database Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString,
				parameters.WithDefault("localhost")),
			parameters.NewParameterDefinition("db-port", parameters.ParameterTypeInteger,
				parameters.WithDefault(5432)),
		),
	)
}

// Extended database layer with additional features
func NewAdvancedDatabaseLayer() (layers.ParameterLayer, error) {
	// Start with base layer
	baseLayer, err := NewBaseDatabaseLayer()
	if err != nil {
		return nil, err
	}

	// Clone to avoid modifying the original
	advancedLayer := baseLayer.Clone()

	// Add additional parameters
	advancedLayer.AddFlags(
		parameters.NewParameterDefinition("db-pool-size", parameters.ParameterTypeInteger,
			parameters.WithDefault(10)),
		parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
			parameters.WithChoices("disable", "require", "verify-full")),
		parameters.NewParameterDefinition("db-connection-timeout", parameters.ParameterTypeString,
			parameters.WithDefault("30s")),
	)

	return advancedLayer, nil
}

// Environment-Specific Layer Configuration
type EnvironmentConfig struct {
	Environment string   // "development", "staging", "production"
	Features    []string // Additional features to enable
}

func NewEnvironmentAwareDatabaseLayer(config EnvironmentConfig) (layers.ParameterLayer, error) {
	layer, err := NewBaseDatabaseLayer()
	if err != nil {
		return nil, err
	}

	// Add environment-specific parameters
	switch config.Environment {
	case "development":
		layer.AddFlags(
			parameters.NewParameterDefinition("db-debug-queries", parameters.ParameterTypeBool,
				parameters.WithDefault(true)),
			parameters.NewParameterDefinition("db-auto-migrate", parameters.ParameterTypeBool,
				parameters.WithDefault(true)),
		)
	case "production":
		layer.AddFlags(
			parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
				parameters.WithChoices("require", "verify-full"),
				parameters.WithDefault("verify-full")),
			parameters.NewParameterDefinition("db-connection-pool-size", parameters.ParameterTypeInteger,
				parameters.WithDefault(50)),
		)
	}

	// Add feature-specific parameters
	for _, feature := range config.Features {
		switch feature {
		case "monitoring":
			layer.AddFlags(
				parameters.NewParameterDefinition("db-monitor-slow-queries", parameters.ParameterTypeBool),
				parameters.NewParameterDefinition("db-slow-query-threshold", parameters.ParameterTypeString,
					parameters.WithDefault("1s")),
			)
		case "backup":
			layer.AddFlags(
				parameters.NewParameterDefinition("db-backup-enabled", parameters.ParameterTypeBool),
				parameters.NewParameterDefinition("db-backup-schedule", parameters.ParameterTypeString),
			)
		}
	}

	return layer, nil
}

func main() {
	fmt.Println("=== Testing Advanced Patterns: Layer Inheritance and Composition ===")

	// Test base database layer
	fmt.Println("\n1. Base Database Layer:")
	baseLayer, err := NewBaseDatabaseLayer()
	if err != nil {
		log.Fatalf("Failed to create base layer: %v", err)
	}
	printLayerInfo(baseLayer)

	// Test advanced database layer (extends base)
	fmt.Println("\n2. Advanced Database Layer (extends base):")
	advancedLayer, err := NewAdvancedDatabaseLayer()
	if err != nil {
		log.Fatalf("Failed to create advanced layer: %v", err)
	}
	printLayerInfo(advancedLayer)

	// Test environment-specific layers
	fmt.Println("\n3. Development Environment Layer:")
	devConfig := EnvironmentConfig{
		Environment: "development",
		Features:    []string{"monitoring"},
	}
	devLayer, err := NewEnvironmentAwareDatabaseLayer(devConfig)
	if err != nil {
		log.Fatalf("Failed to create dev layer: %v", err)
	}
	printLayerInfo(devLayer)

	fmt.Println("\n4. Production Environment Layer:")
	prodConfig := EnvironmentConfig{
		Environment: "production",
		Features:    []string{"monitoring", "backup"},
	}
	prodLayer, err := NewEnvironmentAwareDatabaseLayer(prodConfig)
	if err != nil {
		log.Fatalf("Failed to create prod layer: %v", err)
	}
	printLayerInfo(prodLayer)

	// Test that original base layer is unmodified
	fmt.Println("\n5. Verifying Base Layer Unchanged (should have only 2 parameters):")
	printLayerInfo(baseLayer)

	// Test parameter extraction differences
	fmt.Println("\n6. Testing Parameter Extraction Differences:")

	// Create parameter layer collections
	devLayers := layers.NewParameterLayers(layers.WithLayers(devLayer))
	prodLayers := layers.NewParameterLayers(layers.WithLayers(prodLayer))

	// Initialize with defaults
	devParsed, err := devLayers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize dev parsed layers: %v", err)
	}

	prodParsed, err := prodLayers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize prod parsed layers: %v", err)
	}

	// Show differences
	fmt.Println("Development defaults:")
	if devDbLayer, ok := devParsed.Get("database"); ok {
		if debugQueries, ok := devDbLayer.GetParameter("db-debug-queries"); ok {
			fmt.Printf("  Debug queries: %v\n", debugQueries)
		}
		if autoMigrate, ok := devDbLayer.GetParameter("db-auto-migrate"); ok {
			fmt.Printf("  Auto migrate: %v\n", autoMigrate)
		}
		if monitorSlow, ok := devDbLayer.GetParameter("db-monitor-slow-queries"); ok {
			fmt.Printf("  Monitor slow queries: %v\n", monitorSlow)
		}
	}

	fmt.Println("Production defaults:")
	if prodDbLayer, ok := prodParsed.Get("database"); ok {
		if sslMode, ok := prodDbLayer.GetParameter("db-ssl-mode"); ok {
			fmt.Printf("  SSL mode: %v\n", sslMode)
		}
		if poolSize, ok := prodDbLayer.GetParameter("db-connection-pool-size"); ok {
			fmt.Printf("  Connection pool size: %v\n", poolSize)
		}
		if backupEnabled, ok := prodDbLayer.GetParameter("db-backup-enabled"); ok {
			fmt.Printf("  Backup enabled: %v\n", backupEnabled)
		}
	}
}

func printLayerInfo(layer layers.ParameterLayer) {
	fmt.Printf("  Slug: %s\n", layer.GetSlug())
	fmt.Printf("  Name: %s\n", layer.GetName())

	params := layer.GetParameterDefinitions()
	fmt.Printf("  Parameters (%d total):\n", params.Len())

	params.ForEach(func(param *parameters.ParameterDefinition) {
		fmt.Printf("    - %s (%s)", param.Name, param.Type)
		if param.Default != nil {
			fmt.Printf(" [default: %v]", *param.Default)
		}
		if len(param.Choices) > 0 {
			fmt.Printf(" [choices: %v]", param.Choices)
		}
		fmt.Println()
	})
}
