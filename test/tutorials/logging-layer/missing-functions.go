package main

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
)

// These functions are now implemented in the logging package!
// This test verifies they work correctly.

// Test the missing functions
func main() {
	fmt.Println("=== Testing missing functions implementation ===")

	// Create a logging layer
	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		fmt.Printf("Failed to create logging layer: %v\n", err)
		return
	}

	// Create test parsed layers
	parsedLayers := layers.NewParsedLayers()
	parsedLayer, err := layers.NewParsedLayer(loggingLayer)
	if err != nil {
		fmt.Printf("Failed to create parsed layer: %v\n", err)
		return
	}
	parsedLayers.Set(logging.LoggingLayerSlug, parsedLayer)

	// Test GetLoggingSettings
	fmt.Println("\n1. Testing GetLoggingSettings...")
	settings, err := logging.GetLoggingSettings(parsedLayers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! Settings: %+v\n", *settings)
	}

	// Test SetupLoggingFromParsedLayers
	fmt.Println("\n2. Testing SetupLoggingFromParsedLayers...")
	err = logging.SetupLoggingFromParsedLayers(parsedLayers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✅ Success! Logging setup completed.")
	}

	fmt.Println("\n✅ These functions are now implemented in the logging package!")
}
