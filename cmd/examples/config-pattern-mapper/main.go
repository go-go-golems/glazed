package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// This example demonstrates the new pattern-based config mapping system.
// It shows how to use declarative mapping rules instead of writing custom Go functions.

func main() {
	// Create parameter layers
	demoLayer, err := layers.NewParameterLayer(
		"demo",
		"Demo Layer",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString,
				parameters.WithHelp("API key for authentication")),
			parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger,
				parameters.WithHelp("Threshold value")),
			parameters.NewParameterDefinition("timeout", parameters.ParameterTypeInteger,
				parameters.WithHelp("Timeout in seconds"),
				parameters.WithDefault(30)),
			parameters.NewParameterDefinition("dev-api-key", parameters.ParameterTypeString,
				parameters.WithHelp("Development API key")),
			parameters.NewParameterDefinition("prod-api-key", parameters.ParameterTypeString,
				parameters.WithHelp("Production API key")),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	paramLayers := layers.NewParameterLayers(
		layers.WithLayers(demoLayer),
	)

	// Example 1: Simple exact match mapping
	fmt.Println("=== Example 1: Simple Exact Match ===")
	{
		mapper, err := middlewares.NewConfigMapper(paramLayers,
			middlewares.MappingRule{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
			middlewares.MappingRule{
				Source:          "app.settings.threshold",
				TargetLayer:     "demo",
				TargetParameter: "threshold",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key":   "secret123",
					"threshold": 42,
				},
			},
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 2: Named capture - environment-specific mappings
	fmt.Println("=== Example 2: Named Captures ===")
	{
		mapper, err := middlewares.NewConfigMapper(paramLayers,
			middlewares.MappingRule{
				Source:          "app.{env}.api_key",
				TargetLayer:     "demo",
				TargetParameter: "{env}-api-key",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"dev": map[string]interface{}{
					"api_key": "dev-secret",
				},
				"prod": map[string]interface{}{
					"api_key": "prod-secret",
				},
			},
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 3: Nested rules - cleaner syntax for grouped mappings
	fmt.Println("=== Example 3: Nested Rules ===")
	{
		mapper, err := middlewares.NewConfigMapper(paramLayers,
			middlewares.MappingRule{
				Source:      "app.settings",
				TargetLayer: "demo",
				Rules: []middlewares.MappingRule{
					{Source: "api_key", TargetParameter: "api-key"},
					{Source: "threshold", TargetParameter: "threshold"},
					{Source: "timeout", TargetParameter: "timeout"},
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key":   "secret123",
					"threshold": 42,
					"timeout":   60,
				},
			},
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 4: Nested rules with capture inheritance
	fmt.Println("=== Example 4: Nested Rules with Capture Inheritance ===")
	{
		mapper, err := middlewares.NewConfigMapper(paramLayers,
			middlewares.MappingRule{
				Source:      "environments.{env}.settings",
				TargetLayer: "demo",
				Rules: []middlewares.MappingRule{
					// Child rules can use {env} from parent pattern
					{Source: "api_key", TargetParameter: "{env}-api-key"},
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		config := map[string]interface{}{
			"environments": map[string]interface{}{
				"dev": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "dev-secret",
					},
				},
				"prod": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "prod-secret",
					},
				},
			},
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 5: Using with LoadParametersFromFile middleware
	fmt.Println("=== Example 5: Integration with LoadParametersFromFile ===")
	{
		mapper, err := middlewares.NewConfigMapper(paramLayers,
			middlewares.MappingRule{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		// Use the pattern mapper with LoadParametersFromFile
		_ = middlewares.LoadParametersFromFile(
			"config.yaml",
			middlewares.WithConfigMapper(mapper),
			middlewares.WithParseOptions(
				parameters.WithParseStepSource("config"),
			),
		)

		fmt.Println("Pattern mapper can be used with LoadParametersFromFile middleware")
		fmt.Println("This allows pattern-based mapping without writing custom Go functions")
	}

	fmt.Println("\n=== All examples completed successfully! ===")
}

