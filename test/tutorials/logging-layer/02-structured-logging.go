package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
)

func processFile(fileName string) error {
	start := time.Now()

	log.Debug().
		Str("file", fileName).
		Msg("Starting file processing")

	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Error().
			Str("file", fileName).
			Err(err).
			Msg("Failed to read file")
		return fmt.Errorf("reading file %s: %w", fileName, err)
	}

	log.Info().
		Str("file", fileName).
		Int("bytes_processed", len(data)).
		Dur("duration", time.Since(start)).
		Msg("File processed successfully")

	return nil
}

func main() {
	// Initialize logger with debug level
	settings := &logging.LoggingSettings{
		LogLevel:  "debug",
		LogFormat: "json",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Create a test file
	testFile := "/tmp/test-logging.txt"
	testContent := "Hello, world! This is a test file for logging demonstration."

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create test file")
	}
	defer os.Remove(testFile)

	// Test file processing
	log.Info().Msg("Starting structured logging demonstration")

	if err := processFile(testFile); err != nil {
		log.Error().Err(err).Msg("File processing failed")
	}

	// Test with non-existent file
	if err := processFile("/tmp/non-existent-file.txt"); err != nil {
		log.Warn().Err(err).Msg("Expected file not found error")
	}

	log.Info().Msg("Structured logging demonstration completed")
}
