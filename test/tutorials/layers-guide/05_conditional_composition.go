package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 5: CLI Tool with Optional Features
// Layer composition for applications with conditional functionality

// Feature layers for optional inclusion
func NewCacheLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"cache",
		"Caching Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"cache-enabled",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true),
				parameters.WithHelp("Enable caching"),
			),
			parameters.NewParameterDefinition(
				"cache-ttl",
				parameters.ParameterTypeString,
				parameters.WithDefault("1h"),
				parameters.WithHelp("Cache time-to-live"),
			),
			parameters.NewParameterDefinition(
				"cache-size",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(1000),
				parameters.WithHelp("Maximum cache entries"),
			),
		),
	)
}

func NewMetricsLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"metrics",
		"Metrics and Monitoring",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"metrics-enabled",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Enable metrics collection"),
			),
			parameters.NewParameterDefinition(
				"metrics-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(9090),
				parameters.WithHelp("Metrics server port"),
			),
			parameters.NewParameterDefinition(
				"metrics-path",
				parameters.ParameterTypeString,
				parameters.WithDefault("/metrics"),
				parameters.WithHelp("Metrics endpoint path"),
			),
		),
	)
}

func NewAuthLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"auth",
		"Authentication Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"auth-enabled",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Enable authentication"),
			),
			parameters.NewParameterDefinition(
				"auth-secret",
				parameters.ParameterTypeSecret,
				parameters.WithHelp("Authentication secret key"),
			),
			parameters.NewParameterDefinition(
				"auth-timeout",
				parameters.ParameterTypeString,
				parameters.WithDefault("24h"),
				parameters.WithHelp("Authentication token timeout"),
			),
		),
	)
}

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

// Command builder with optional features
type AppCommandBuilder struct {
	baseLayers    []layers.ParameterLayer
	enableCache   bool
	enableMetrics bool
	enableAuth    bool
}

func NewAppCommandBuilder() *AppCommandBuilder {
	loggingLayer, _ := NewLoggingLayer()

	return &AppCommandBuilder{
		baseLayers: []layers.ParameterLayer{loggingLayer},
	}
}

func (b *AppCommandBuilder) WithCache() *AppCommandBuilder {
	b.enableCache = true
	return b
}

func (b *AppCommandBuilder) WithMetrics() *AppCommandBuilder {
	b.enableMetrics = true
	return b
}

func (b *AppCommandBuilder) WithAuth() *AppCommandBuilder {
	b.enableAuth = true
	return b
}

func (b *AppCommandBuilder) BuildProcessCommand() (*cmds.CommandDescription, error) {
	commandLayers := append([]layers.ParameterLayer{}, b.baseLayers...)

	// Add optional layers based on enabled features
	if b.enableCache {
		cacheLayer, err := NewCacheLayer()
		if err != nil {
			return nil, err
		}
		commandLayers = append(commandLayers, cacheLayer)
	}

	if b.enableMetrics {
		metricsLayer, err := NewMetricsLayer()
		if err != nil {
			return nil, err
		}
		commandLayers = append(commandLayers, metricsLayer)
	}

	if b.enableAuth {
		authLayer, err := NewAuthLayer()
		if err != nil {
			return nil, err
		}
		commandLayers = append(commandLayers, authLayer)
	}

	return cmds.NewCommandDescription(
		"process",
		cmds.WithShort("Process data with optional features"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-file",
				parameters.ParameterTypeFile,
				parameters.WithHelp("Input file to process"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"output-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Output file path"),
			),
		),
		cmds.WithLayersList(commandLayers...),
	), nil
}

func main() {
	fmt.Println("=== Testing Conditional Layer Composition ===")

	// Test basic command (only logging)
	fmt.Println("\n1. Basic Command (only logging):")
	basicCmd, err := NewAppCommandBuilder().BuildProcessCommand()
	if err != nil {
		log.Fatalf("Failed to create basic command: %v", err)
	}
	printCommandInfo(basicCmd)

	// Test command with cache
	fmt.Println("\n2. Command with Cache:")
	cacheCmd, err := NewAppCommandBuilder().WithCache().BuildProcessCommand()
	if err != nil {
		log.Fatalf("Failed to create cache command: %v", err)
	}
	printCommandInfo(cacheCmd)

	// Test command with metrics
	fmt.Println("\n3. Command with Metrics:")
	metricsCmd, err := NewAppCommandBuilder().WithMetrics().BuildProcessCommand()
	if err != nil {
		log.Fatalf("Failed to create metrics command: %v", err)
	}
	printCommandInfo(metricsCmd)

	// Test feature-rich command (cache + metrics + auth)
	fmt.Println("\n4. Feature-Rich Command (cache + metrics + auth):")
	advancedCmd, err := NewAppCommandBuilder().
		WithCache().
		WithMetrics().
		WithAuth().
		BuildProcessCommand()
	if err != nil {
		log.Fatalf("Failed to create advanced command: %v", err)
	}
	printCommandInfo(advancedCmd)

	// Test parameter extraction from advanced command
	fmt.Println("\n5. Testing Parameter Extraction from Advanced Command:")
	parsedLayers, err := advancedCmd.Layers.InitializeFromDefaults()
	if err != nil {
		log.Fatalf("Failed to initialize parsed layers: %v", err)
	}

	// Check which features are enabled by default
	fmt.Println("Default feature settings:")

	if cacheLayer, ok := parsedLayers.Get("cache"); ok {
		if cacheEnabled, ok := cacheLayer.GetParameter("cache-enabled"); ok {
			fmt.Printf("  Cache enabled: %v\n", cacheEnabled)
		}
	}

	if metricsLayer, ok := parsedLayers.Get("metrics"); ok {
		if metricsEnabled, ok := metricsLayer.GetParameter("metrics-enabled"); ok {
			fmt.Printf("  Metrics enabled: %v\n", metricsEnabled)
		}
	}

	if authLayer, ok := parsedLayers.Get("auth"); ok {
		if authEnabled, ok := authLayer.GetParameter("auth-enabled"); ok {
			fmt.Printf("  Auth enabled: %v\n", authEnabled)
		}
	}

	if loggingLayer, ok := parsedLayers.Get("logging"); ok {
		if logLevel, ok := loggingLayer.GetParameter("log-level"); ok {
			fmt.Printf("  Log level: %v\n", logLevel)
		}
	}
}

func printCommandInfo(cmd *cmds.CommandDescription) {
	fmt.Printf("  Command: %s\n", cmd.Name)
	fmt.Printf("  Short: %s\n", cmd.Short)
	fmt.Printf("  Layers (%d total):\n", cmd.Layers.Len())

	cmd.Layers.ForEach(func(slug string, layer layers.ParameterLayer) {
		fmt.Printf("    - %s: %s\n", slug, layer.GetName())
		params := layer.GetParameterDefinitions()
		fmt.Printf("      Parameters (%d):", params.Len())
		params.ForEach(func(param *parameters.ParameterDefinition) {
			fmt.Printf(" %s", param.Name)
		})
		fmt.Println()
	})
}
