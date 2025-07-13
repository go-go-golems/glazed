package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
)

func validateUser(userID string) error {
	if userID == "invalid" {
		return fmt.Errorf("user %s is not valid", userID)
	}
	return nil
}

func processUser(userID string) error {
	userLogger := log.With().
		Str("user_id", userID).
		Str("operation", "user_processing").
		Logger()

	userLogger.Info().Msg("Starting user processing")

	if err := validateUser(userID); err != nil {
		userLogger.Error().
			Err(err).
			Msg("User validation failed")
		return err
	}

	userLogger.Info().Msg("User processing completed")
	return nil
}

func main() {
	// Initialize logger with text format for better readability
	settings := &logging.LoggingSettings{
		LogLevel:  "info",
		LogFormat: "text",
	}

	if err := logging.InitLoggerFromSettings(settings); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	log.Info().Msg("Starting contextual logger demonstration")

	// Test valid user
	if err := processUser("user123"); err != nil {
		log.Error().Err(err).Msg("User processing failed")
	}

	// Test invalid user
	if err := processUser("invalid"); err != nil {
		log.Warn().Err(err).Msg("Expected user validation error")
	}

	// Test another valid user
	if err := processUser("admin456"); err != nil {
		log.Error().Err(err).Msg("User processing failed")
	}

	log.Info().Msg("Contextual logger demonstration completed")
}
