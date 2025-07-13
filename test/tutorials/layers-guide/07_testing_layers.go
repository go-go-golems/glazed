package main

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Example 7: Testing Layers
// This demonstrates unit and integration testing approaches from the documentation

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
				parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
				parameters.WithDefault("info"),
				parameters.WithHelp("Logging level"),
			),
		),
	)
}

// Test functions that simulate what would be in *_test.go files

func TestDatabaseLayer() bool {
	fmt.Println("  Running TestDatabaseLayer...")

	layer, err := NewDatabaseLayer()
	if err != nil {
		fmt.Printf("    ❌ Failed to create layer: %v\n", err)
		return false
	}

	if layer.GetSlug() != "database" {
		fmt.Printf("    ❌ Expected slug 'database', got '%s'\n", layer.GetSlug())
		return false
	}

	// Test parameter definitions
	params := layer.GetParameterDefinitions()
	expectedParams := []string{"db-host", "db-port", "db-name"}

	if params.Len() != len(expectedParams) {
		fmt.Printf("    ❌ Expected %d parameters, got %d\n", len(expectedParams), params.Len())
		return false
	}

	for _, expected := range expectedParams {
		if _, ok := params.Get(expected); !ok {
			fmt.Printf("    ❌ Parameter '%s' not found\n", expected)
			return false
		}
	}

	// Test default values
	if hostParam, ok := params.Get("db-host"); ok {
		if hostParam.Default == nil || *hostParam.Default != "localhost" {
			fmt.Printf("    ❌ Expected db-host default 'localhost', got %v\n", hostParam.Default)
			return false
		}
	}

	if portParam, ok := params.Get("db-port"); ok {
		if portParam.Default == nil || *portParam.Default != 5432 {
			fmt.Printf("    ❌ Expected db-port default 5432, got %v\n", portParam.Default)
			return false
		}
	}

	fmt.Println("    ✅ TestDatabaseLayer passed")
	return true
}

func TestParameterValidation() bool {
	fmt.Println("  Running TestParameterValidation...")

	layer, err := NewLoggingLayer()
	if err != nil {
		fmt.Printf("    ❌ Failed to create layer: %v\n", err)
		return false
	}

	// Test valid choices
	params := layer.GetParameterDefinitions()
	if logLevelParam, ok := params.Get("log-level"); ok {
		validChoices := []string{"debug", "info", "warn", "error", "fatal"}
		if len(logLevelParam.Choices) != len(validChoices) {
			fmt.Printf("    ❌ Expected %d choices, got %d\n", len(validChoices), len(logLevelParam.Choices))
			return false
		}

		for i, expected := range validChoices {
			if logLevelParam.Choices[i] != expected {
				fmt.Printf("    ❌ Expected choice '%s', got '%s'\n", expected, logLevelParam.Choices[i])
				return false
			}
		}
	}

	// Test required parameters
	dbLayer, _ := NewDatabaseLayer()
	dbParams := dbLayer.GetParameterDefinitions()
	if dbNameParam, ok := dbParams.Get("db-name"); ok {
		if !dbNameParam.Required {
			fmt.Printf("    ❌ Expected db-name to be required\n")
			return false
		}
	}

	fmt.Println("    ✅ TestParameterValidation passed")
	return true
}

func TestCommandComposition() bool {
	fmt.Println("  Running TestCommandComposition...")

	serverLayer, err := NewDatabaseLayer()
	if err != nil {
		fmt.Printf("    ❌ Failed to create database layer: %v\n", err)
		return false
	}

	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		fmt.Printf("    ❌ Failed to create logging layer: %v\n", err)
		return false
	}

	command := cmds.NewCommandDescription("test-command",
		cmds.WithLayersList(serverLayer, loggingLayer))

	if command == nil {
		fmt.Printf("    ❌ Command is nil\n")
		return false
	}

	// Verify all layers are present
	commandLayers := command.Layers
	if commandLayers.Len() != 2 {
		fmt.Printf("    ❌ Expected 2 layers, got %d\n", commandLayers.Len())
		return false
	}

	// Verify no parameter conflicts
	allParams := make(map[string]bool)
	success := true
	commandLayers.ForEach(func(slug string, layer layers.ParameterLayer) {
		layer.GetParameterDefinitions().ForEach(func(param *parameters.ParameterDefinition) {
			if allParams[param.Name] {
				fmt.Printf("    ❌ Parameter %s defined in multiple layers\n", param.Name)
				success = false
			}
			allParams[param.Name] = true
		})
	})

	if !success {
		return false
	}

	fmt.Println("    ✅ TestCommandComposition passed")
	return true
}

func TestParameterResolution() bool {
	fmt.Println("  Running TestParameterResolution...")

	// Create test command with layers
	databaseLayer, _ := NewDatabaseLayer()
	loggingLayer, _ := NewLoggingLayer()

	command := cmds.NewCommandDescription("test-command",
		cmds.WithLayersList(databaseLayer, loggingLayer))

	// Test default value resolution
	parsedLayers, err := command.Layers.InitializeFromDefaults()
	if err != nil {
		fmt.Printf("    ❌ Failed to initialize parsed layers: %v\n", err)
		return false
	}

	// Verify default values are used
	if dbLayer, ok := parsedLayers.Get("database"); ok {
		if dbHost, ok := dbLayer.GetParameter("db-host"); ok {
			if dbHost != "localhost" {
				fmt.Printf("    ❌ Expected db-host 'localhost', got %v\n", dbHost)
				return false
			}
		} else {
			fmt.Printf("    ❌ db-host parameter not found\n")
			return false
		}

		if dbPort, ok := dbLayer.GetParameter("db-port"); ok {
			if dbPort != 5432 {
				fmt.Printf("    ❌ Expected db-port 5432, got %v\n", dbPort)
				return false
			}
		} else {
			fmt.Printf("    ❌ db-port parameter not found\n")
			return false
		}
	} else {
		fmt.Printf("    ❌ Database layer not found in parsed layers\n")
		return false
	}

	if loggingLayer, ok := parsedLayers.Get("logging"); ok {
		if logLevel, ok := loggingLayer.GetParameter("log-level"); ok {
			if logLevel != "info" {
				fmt.Printf("    ❌ Expected log-level 'info', got %v\n", logLevel)
				return false
			}
		} else {
			fmt.Printf("    ❌ log-level parameter not found\n")
			return false
		}
	} else {
		fmt.Printf("    ❌ Logging layer not found in parsed layers\n")
		return false
	}

	fmt.Println("    ✅ TestParameterResolution passed")
	return true
}

// Database settings struct for testing struct initialization
type DatabaseSettings struct {
	Host string `glazed.parameter:"db-host"`
	Port int    `glazed.parameter:"db-port"`
	Name string `glazed.parameter:"db-name"`
}

func TestStructInitialization() bool {
	fmt.Println("  Running TestStructInitialization...")

	databaseLayer, _ := NewDatabaseLayer()
	paramLayers := layers.NewParameterLayers(layers.WithLayers(databaseLayer))

	parsedLayers, err := paramLayers.InitializeFromDefaults()
	if err != nil {
		fmt.Printf("    ❌ Failed to initialize parsed layers: %v\n", err)
		return false
	}

	// Test struct initialization
	dbSettings := &DatabaseSettings{}
	err = parsedLayers.InitializeStruct("database", dbSettings)
	if err != nil {
		fmt.Printf("    ❌ Failed to initialize struct: %v\n", err)
		return false
	}

	if dbSettings.Host != "localhost" {
		fmt.Printf("    ❌ Expected Host 'localhost', got '%s'\n", dbSettings.Host)
		return false
	}

	if dbSettings.Port != 5432 {
		fmt.Printf("    ❌ Expected Port 5432, got %d\n", dbSettings.Port)
		return false
	}

	fmt.Println("    ✅ TestStructInitialization passed")
	return true
}

func main() {
	fmt.Println("=== Testing Layers: Unit and Integration Tests ===")

	tests := []struct {
		name string
		test func() bool
	}{
		{"Unit Testing Layer Definitions", TestDatabaseLayer},
		{"Unit Testing Parameter Validation", TestParameterValidation},
		{"Integration Testing Command Composition", TestCommandComposition},
		{"Integration Testing Parameter Resolution", TestParameterResolution},
		{"Integration Testing Struct Initialization", TestStructInitialization},
	}

	passed := 0
	total := len(tests)

	for i, test := range tests {
		fmt.Printf("\n%d. %s:\n", i+1, test.name)
		if test.test() {
			passed++
		}
	}

	fmt.Printf("\n=== Test Results ===\n")
	fmt.Printf("Passed: %d/%d\n", passed, total)
	if passed == total {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Printf("❌ %d tests failed\n", total-passed)
	}
}
