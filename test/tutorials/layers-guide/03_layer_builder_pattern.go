package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 3: Layer Builder Pattern
// This demonstrates the builder pattern for complex scenarios requiring conditional parameters

type DatabaseLayerBuilder struct {
	layer       layers.ParameterLayer
	includeSSL  bool
	includePool bool
}

func NewDatabaseLayerBuilder() *DatabaseLayerBuilder {
	layer, _ := layers.NewParameterLayer("database", "Database Configuration")
	return &DatabaseLayerBuilder{layer: layer}
}

func (b *DatabaseLayerBuilder) WithSSL() *DatabaseLayerBuilder {
	b.includeSSL = true
	return b
}

func (b *DatabaseLayerBuilder) WithConnectionPool() *DatabaseLayerBuilder {
	b.includePool = true
	return b
}

func (b *DatabaseLayerBuilder) Build() (layers.ParameterLayer, error) {
	// Add basic parameters
	b.layer.AddFlags(
		parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString,
			parameters.WithDefault("localhost")),
		parameters.NewParameterDefinition("db-port", parameters.ParameterTypeInteger,
			parameters.WithDefault(5432)),
	)

	// Conditionally add SSL parameters
	if b.includeSSL {
		b.layer.AddFlags(
			parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
				parameters.WithChoices("disable", "require", "verify-ca"),
				parameters.WithDefault("require")),
			parameters.NewParameterDefinition("db-ssl-cert", parameters.ParameterTypeFile,
				parameters.WithHelp("SSL certificate file")),
		)
	}

	// Conditionally add connection pool parameters
	if b.includePool {
		b.layer.AddFlags(
			parameters.NewParameterDefinition("db-max-connections", parameters.ParameterTypeInteger,
				parameters.WithDefault(10)),
			parameters.NewParameterDefinition("db-idle-timeout", parameters.ParameterTypeString,
				parameters.WithDefault("5m")),
		)
	}

	return b.layer, nil
}

func main() {
	fmt.Println("=== Testing Layer Builder Pattern ===")

	// Test basic layer (no optional features)
	fmt.Println("\n1. Basic Database Layer:")
	basicLayer, err := NewDatabaseLayerBuilder().Build()
	if err != nil {
		log.Fatalf("Failed to build basic layer: %v", err)
	}

	printLayerInfo(basicLayer)

	// Test layer with SSL
	fmt.Println("\n2. Database Layer with SSL:")
	sslLayer, err := NewDatabaseLayerBuilder().WithSSL().Build()
	if err != nil {
		log.Fatalf("Failed to build SSL layer: %v", err)
	}

	printLayerInfo(sslLayer)

	// Test layer with connection pool
	fmt.Println("\n3. Database Layer with Connection Pool:")
	poolLayer, err := NewDatabaseLayerBuilder().WithConnectionPool().Build()
	if err != nil {
		log.Fatalf("Failed to build pool layer: %v", err)
	}

	printLayerInfo(poolLayer)

	// Test layer with all features
	fmt.Println("\n4. Database Layer with All Features:")
	fullLayer, err := NewDatabaseLayerBuilder().
		WithSSL().
		WithConnectionPool().
		Build()
	if err != nil {
		log.Fatalf("Failed to build full layer: %v", err)
	}

	printLayerInfo(fullLayer)
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
