package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
)

func demonstrateLogLevels() {
	log.Trace().Msg("This is a trace message")
	log.Debug().Msg("This is a debug message")
	log.Info().Msg("This is an info message")
	log.Warn().Msg("This is a warning message")
	log.Error().Msg("This is an error message")
}

func main() {
	fmt.Println("=== Testing different log levels and formats ===")

	// Test with text format and debug level
	fmt.Println("\n--- Text format, debug level ---")
	settings := &logging.LoggingSettings{
		LogLevel:  "debug",
		LogFormat: "text",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	demonstrateLogLevels()

	// Test with JSON format and info level
	fmt.Println("\n--- JSON format, info level ---")
	settings = &logging.LoggingSettings{
		LogLevel:  "info",
		LogFormat: "json",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	demonstrateLogLevels()

	// Test with text format and warn level
	fmt.Println("\n--- Text format, warn level ---")
	settings = &logging.LoggingSettings{
		LogLevel:  "warn",
		LogFormat: "text",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	demonstrateLogLevels()

	// Test with caller information
	fmt.Println("\n--- Text format with caller information ---")
	settings = &logging.LoggingSettings{
		LogLevel:   "info",
		LogFormat:  "text",
		WithCaller: true,
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	log.Info().Msg("Message with caller information")
}
