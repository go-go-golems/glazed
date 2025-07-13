package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/rs/zerolog/log"
)

// The documentation references these functions that we need to test:
// 1. SetupLoggingFromParsedLayers(parsedLayers)
// 2. GetLoggingSettings(parsedLayers)

func TestMissingFunctions() {
	// Create a command with logging layer
	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create logging layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"test-command",
		cmds.WithShort("Test command for missing functions"),
		cmds.WithLayersList(loggingLayer),
	)

	// Create test parsed layers
	parsedLayers := layers.NewParsedLayers()
	parsedLayer, err := layers.NewParsedLayer(loggingLayer)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create parsed layer")
	}
	parsedLayers.Set(logging.LoggingLayerSlug, parsedLayer)

	fmt.Println("=== Testing documented but missing functions ===")

	// Test 1: SetupLoggingFromParsedLayers (referenced in documentation)
	fmt.Println("\n1. Testing SetupLoggingFromParsedLayers...")
	// This function is documented but doesn't exist
	// if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
	//     fmt.Printf("Error: %v\n", err)
	// }
	fmt.Println("   Function SetupLoggingFromParsedLayers does not exist!")

	// Test 2: GetLoggingSettings (referenced in documentation)
	fmt.Println("\n2. Testing GetLoggingSettings...")
	// This function is documented but doesn't exist
	// settings, err := logging.GetLoggingSettings(parsedLayers)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	// }
	fmt.Println("   Function GetLoggingSettings does not exist!")

	// Test 3: What actually works - InitializeStruct
	fmt.Println("\n3. Testing what actually works - InitializeStruct...")
	var settings logging.LoggingSettings
	err = parsedLayers.InitializeStruct(logging.LoggingLayerSlug, &settings)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success! Settings: %+v\n", settings)

		// This works for setting up logging
		if err := logging.InitLoggerFromSettings(&settings); err != nil {
			fmt.Printf("Error setting up logger: %v\n", err)
		} else {
			fmt.Println("Logger setup successful!")
		}
	}

	_ = cmdDesc // Avoid unused variable warning
}

type TestCommand struct {
	*cmds.CommandDescription
}

func (c *TestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// This is the proper way to use logging with parsed layers
	var settings logging.LoggingSettings
	err := parsedLayers.InitializeStruct(logging.LoggingLayerSlug, &settings)
	if err != nil {
		return fmt.Errorf("failed to get logging settings: %w", err)
	}

	if err := logging.InitLoggerFromSettings(&settings); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	log.Info().Msg("Command executed successfully with proper logging setup")
	return nil
}

func main() {
	TestMissingFunctions()

	fmt.Println("\n=== Creating missing functions ===")
	fmt.Println("The documentation references functions that should be implemented.")
	fmt.Println("See the report for details on what needs to be added.")
}
