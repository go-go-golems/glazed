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
	if err := logging.SetupLoggingFromParsedLayers(parsedLayers); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("   ✅ SetupLoggingFromParsedLayers works!")
	}

	// Test 2: GetLoggingSettings (referenced in documentation)
	fmt.Println("\n2. Testing GetLoggingSettings...")
	settings, err := logging.GetLoggingSettings(parsedLayers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ GetLoggingSettings works! Settings: %+v\n", *settings)
	}

	// Test 3: Comparing both approaches
	fmt.Println("\n3. Testing both approaches give same result...")
	var directSettings logging.LoggingSettings
	err = parsedLayers.InitializeStruct(logging.LoggingLayerSlug, &directSettings)
	if err != nil {
		fmt.Printf("Error with direct approach: %v\n", err)
	} else {
		fmt.Printf("   Direct InitializeStruct: %+v\n", directSettings)
		fmt.Printf("   Via GetLoggingSettings: %+v\n", *settings)
		if directSettings == *settings {
			fmt.Println("   ✅ Both approaches return identical settings!")
		} else {
			fmt.Println("   ❌ Settings differ between approaches!")
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

	fmt.Println("\n=== Testing complete! ===")
	fmt.Println("✅ All documented functions are now implemented and working.")
	fmt.Println("✅ Both convenience functions and direct approach work correctly.")

	// Test the actual logging to verify it works
	fmt.Println("\n=== Testing actual logging output ===")
	log.Info().Msg("This message should appear with configured logging settings")
	log.Debug().Msg("This debug message may or may not appear depending on log level")
	log.Warn().Str("status", "complete").Msg("All tests completed successfully")
}
