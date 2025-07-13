package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"glazed-logging-layer/logging"
)

func main() {
	fmt.Println("Testing logging initialization functions...")

	// Test SetupDefaultLogging
	err := logging.SetupDefaultLogging()
	if err != nil {
		fmt.Printf("ERROR: Failed to setup default logging: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ SetupDefaultLogging works")

	// Test that logger is configured
	log.Info().Msg("Test log message after default setup")
	fmt.Println("✓ Logger is working after default setup")

	// Test settings validation and setup
	settings := &logging.LoggingSettings{
		Level:      "debug",
		Format:     "json",
		File:       "",
		WithCaller: true,
		Verbose:    false,
	}

	err = settings.SetupLogger()
	if err != nil {
		fmt.Printf("ERROR: Failed to setup logger with custom settings: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Custom logger setup works")

	// Test that logger configuration changed
	log.Debug().Msg("Debug message after custom setup")
	fmt.Println("✓ Debug logging is enabled after custom setup")

	// Test logger with file output
	fileSettings := &logging.LoggingSettings{
		Level:      "info",
		Format:     "json",
		File:       "test-init.log",
		WithCaller: false,
		Verbose:    false,
	}

	err = fileSettings.SetupLogger()
	if err != nil {
		fmt.Printf("ERROR: Failed to setup file logging: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ File logging setup works")

	log.Info().Str("test", "value").Msg("Test message to file")
	fmt.Println("✓ Logging to file works")

	// Check if log file was created
	if _, err := os.Stat("test-init.log"); os.IsNotExist(err) {
		fmt.Println("ERROR: Log file was not created")
		os.Exit(1)
	}
	fmt.Println("✓ Log file was created")

	// Cleanup
	os.Remove("test-init.log")

	fmt.Println("\nAll initialization tests passed! ✅")
}
