package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
)

func main() {
	logFile := "/tmp/test-app.log"

	// Remove log file if it exists
	os.Remove(logFile)

	fmt.Println("=== Testing file logging ===")

	// Test logging to file only
	fmt.Println("\n--- Logging to file only ---")
	settings := &logging.LoggingSettings{
		LogLevel:  "info",
		LogFormat: "json",
		LogFile:   logFile,
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	log.Info().Msg("This message goes to file only")
	log.Warn().Str("component", "file-test").Msg("Warning message to file")

	// Check if file was created and read its contents
	if data, err := os.ReadFile(logFile); err == nil {
		fmt.Printf("Log file contents:\n%s\n", string(data))
	} else {
		fmt.Printf("Failed to read log file: %v\n", err)
	}

	// Test logging to both file and stdout
	fmt.Println("\n--- Logging to both file and stdout ---")
	settings = &logging.LoggingSettings{
		LogLevel:    "debug",
		LogFormat:   "text",
		LogFile:     logFile,
		LogToStdout: true,
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	log.Info().Msg("This message goes to both file and stdout")
	log.Debug().Str("mode", "dual-output").Msg("Debug message to both outputs")

	fmt.Println("\nFinal log file contents:")
	if data, err := os.ReadFile(logFile); err == nil {
		fmt.Printf("%s\n", string(data))
	} else {
		fmt.Printf("Failed to read log file: %v\n", err)
	}

	// Clean up
	os.Remove(logFile)
}
