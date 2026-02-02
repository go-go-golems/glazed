package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	pm "github.com/go-go-golems/glazed/pkg/cmds/sources/patternmapper"
	"gopkg.in/yaml.v3"
)

//go:embed mappings.yaml
var mappingsYAML []byte

//go:embed config-example.yaml
var configExampleYAML []byte

//go:embed config-ex1.yaml
var configEx1 []byte

//go:embed config-ex2.yaml
var configEx2 []byte

//go:embed config-ex3.yaml
var configEx3 []byte

//go:embed config-ex4.yaml
var configEx4 []byte

//go:embed config-ex6.yaml
var configEx6 []byte

//go:embed config-ex7.yaml
var configEx7 []byte

//go:embed config-ex8.yaml
var configEx8 []byte

// This example demonstrates the new pattern-based config mapping system.
// It shows how to use declarative mapping rules instead of writing custom Go functions.

func main() {
	// Create schema section
	demoLayer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("api-key", fields.TypeString, fields.WithHelp("API key for authentication")),
			fields.New("threshold", fields.TypeInteger, fields.WithHelp("Threshold value")),
			fields.New("timeout", fields.TypeInteger, fields.WithHelp("Timeout in seconds"), fields.WithDefault(30)),
			fields.New("dev-api-key", fields.TypeString, fields.WithHelp("Development API key")),
			fields.New("prod-api-key", fields.TypeString, fields.WithHelp("Production API key")),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	paramLayers := schema.NewSchema(schema.WithSections(demoLayer))

	// Simple CLI switch: `validate [config.yaml]` validates the config against mappings.yaml
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		mappingPath := "cmd/examples/config-pattern-mapper/mappings.yaml"
		configPath := "cmd/examples/config-pattern-mapper/config-example.yaml"
		if len(os.Args) > 2 {
			configPath = os.Args[2]
		}

		rules, err := pm.LoadRulesFromFile(mappingPath)
		if err != nil {
			log.Fatal(err)
		}
		mapper, err := pm.NewConfigMapper(paramLayers, rules...)
		if err != nil {
			log.Fatal(err)
		}
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
		var cfg map[string]interface{}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatal(err)
		}
		if _, err := mapper.Map(cfg); err != nil {
			log.Fatal(err)
		}
		fmt.Println("OK")
		return
	}

	// Example 1: Simple exact match mapping
	fmt.Println("=== Example 1: Simple Exact Match ===")
	{
		mapper, err := pm.NewConfigMapper(paramLayers,
			pm.MappingRule{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
			pm.MappingRule{
				Source:          "app.settings.threshold",
				TargetLayer:     "demo",
				TargetParameter: "threshold",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx1, &config); err != nil {
			log.Fatal(err)
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
		mapper, err := pm.NewConfigMapper(paramLayers,
			pm.MappingRule{
				Source:          "app.{env}.api_key",
				TargetLayer:     "demo",
				TargetParameter: "{env}-api-key",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx2, &config); err != nil {
			log.Fatal(err)
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
		mapper, err := pm.NewConfigMapper(paramLayers,
			pm.MappingRule{
				Source:      "app.settings",
				TargetLayer: "demo",
				Rules: []pm.MappingRule{
					{Source: "api_key", TargetParameter: "api-key"},
					{Source: "threshold", TargetParameter: "threshold"},
					{Source: "timeout", TargetParameter: "timeout"},
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx3, &config); err != nil {
			log.Fatal(err)
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
		mapper, err := pm.NewConfigMapper(paramLayers,
			pm.MappingRule{
				Source:      "environments.{env}.settings",
				TargetLayer: "demo",
				Rules: []pm.MappingRule{
					// Child rules can use {env} from parent pattern
					{Source: "api_key", TargetParameter: "{env}-api-key"},
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx4, &config); err != nil {
			log.Fatal(err)
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
		mapper, err := pm.NewConfigMapper(paramLayers,
			pm.MappingRule{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		// Use the pattern mapper with sources.FromFile (wraps LoadParametersFromFile)
		_ = sources.FromFile(
			"config.yaml",
			sources.WithConfigMapper(mapper),
			sources.WithParseOptions(fields.WithSource("config")),
		)

		fmt.Println("Pattern mapper can be used with LoadParametersFromFile middleware")
		fmt.Println("This allows pattern-based mapping without writing custom Go functions")
	}

	// Example 6: Builder API - Simple Exact Match
	fmt.Println("=== Example 6: Builder API - Simple Exact Match ===")
	{
		b := pm.NewConfigMapperBuilder(paramLayers).
			Map("app.settings.api_key", "demo", "api-key")

		mapper, err := b.Build()
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx6, &config); err != nil {
			log.Fatal(err)
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 7: Builder API - Nested Rules with Capture Inheritance
	fmt.Println("=== Example 7: Builder API - Nested Rules with Capture Inheritance ===")
	{
		b := pm.NewConfigMapperBuilder(paramLayers).
			MapObject("environments.{env}.settings", "demo", []pm.MappingRule{
				pm.Child("api_key", "{env}-api-key"),
			})

		mapper, err := b.Build()
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx7, &config); err != nil {
			log.Fatal(err)
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 8: Builder API - Required Flag
	fmt.Println("=== Example 8: Builder API - Required Flag ===")
	{
		b := pm.NewConfigMapperBuilder(paramLayers).
			Map("app.settings.api_key", "demo", "api-key", true)

		mapper, err := b.Build()
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configEx8, &config); err != nil {
			log.Fatal(err)
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	// Example 9: YAML/JSON Loader (go:embed) - Load rules and config from embedded files
	fmt.Println("=== Example 9: YAML/JSON Loader (go:embed) ===")
	{
		rules, err := pm.LoadRulesFromReader(bytes.NewReader(mappingsYAML))
		if err != nil {
			log.Fatal(err)
		}
		mapper, err := pm.NewConfigMapper(paramLayers, rules...)
		if err != nil {
			log.Fatal(err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configExampleYAML, &config); err != nil {
			log.Fatal(err)
		}

		result, err := mapper.Map(config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("mappings.yaml (embedded) length: %d bytes\n", len(mappingsYAML))
		fmt.Printf("config-example.yaml (embedded) length: %d bytes\n", len(configExampleYAML))
		fmt.Printf("Config: %v\n", config)
		fmt.Printf("Mapped: %v\n\n", result)
	}

	fmt.Println("\n=== All examples completed successfully! ===")
}
