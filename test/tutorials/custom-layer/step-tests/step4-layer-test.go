package main

import (
	"fmt"
	"os"

	"glazed-logging-layer/logging"
)

func main() {
	fmt.Println("Testing logging layer creation...")

	// Test basic layer creation
	layer, err := logging.NewLoggingLayer()
	if err != nil {
		fmt.Printf("ERROR: Failed to create logging layer: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Basic logging layer created successfully")

	// Check layer properties
	if layer.GetSlug() != logging.LoggingSlug {
		fmt.Printf("ERROR: Expected slug '%s', got '%s'\n", logging.LoggingSlug, layer.GetSlug())
		os.Exit(1)
	}
	fmt.Println("✓ Layer slug is correct")

	if layer.GetName() != "Logging Configuration" {
		fmt.Printf("ERROR: Expected name 'Logging Configuration', got '%s'\n", layer.GetName())
		os.Exit(1)
	}
	fmt.Println("✓ Layer name is correct")

	// Check parameter definitions
	params := layer.GetParameterDefinitions()
	expectedParams := []string{
		"log-level", "log-format", "log-file",
		"with-caller", "verbose",
		"logstash-host", "logstash-port",
	}

	for _, paramName := range expectedParams {
		if param, ok := params.Get(paramName); !ok || param == nil {
			fmt.Printf("ERROR: Expected parameter '%s' not found\n", paramName)
			os.Exit(1)
		}
	}
	fmt.Println("✓ All expected parameters are present")

	// Test layer with options
	layerWithOpts, err := logging.NewLoggingLayerWithOptions(
		logging.WithDefaultLevel("debug"),
		logging.WithDefaultFormat("json"),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create logging layer with options: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Logging layer with options created successfully")

	// Check that defaults were applied
	paramsWithOpts := layerWithOpts.GetParameterDefinitions()
	if levelParam, ok := paramsWithOpts.Get("log-level"); ok && levelParam != nil {
		if levelParam.Default == nil || *levelParam.Default != "debug" {
			fmt.Printf("ERROR: Expected default level 'debug', got %v\n", levelParam.Default)
			os.Exit(1)
		}
	} else {
		fmt.Println("ERROR: log-level parameter not found in layer with options")
		os.Exit(1)
	}
	fmt.Println("✓ Custom default level applied correctly")

	if formatParam, ok := paramsWithOpts.Get("log-format"); ok && formatParam != nil {
		if formatParam.Default == nil || *formatParam.Default != "json" {
			fmt.Printf("ERROR: Expected default format 'json', got %v\n", formatParam.Default)
			os.Exit(1)
		}
	} else {
		fmt.Println("ERROR: log-format parameter not found in layer with options")
		os.Exit(1)
	}
	fmt.Println("✓ Custom default format applied correctly")

	fmt.Println("\nAll layer tests passed! ✅")
}
