package logging

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/rs/zerolog/log"
)

// GetLoggingSettings extracts logging settings from parsed layers
func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
	settings := &LoggingSettings{}
	if err := parsedLayers.InitializeStruct(LoggingSlug, settings); err != nil {
		return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
	}
	return settings, nil
}

// InitializeLogging sets up logging from parsed layers
func InitializeLogging(parsedLayers *layers.ParsedLayers) error {
	settings, err := GetLoggingSettings(parsedLayers)
	if err != nil {
		return err
	}

	if err := settings.SetupLogger(); err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	log.Debug().
		Str("level", settings.Level).
		Str("format", settings.Format).
		Str("file", settings.File).
		Bool("with_caller", settings.WithCaller).
		Bool("verbose", settings.Verbose).
		Msg("Logging initialized")

	return nil
}

// MustInitializeLogging sets up logging or panics on error
func MustInitializeLogging(parsedLayers *layers.ParsedLayers) {
	if err := InitializeLogging(parsedLayers); err != nil {
		panic(fmt.Sprintf("Failed to initialize logging: %v", err))
	}
}

// SetupDefaultLogging configures logging with default settings (useful for testing)
func SetupDefaultLogging() error {
	settings := &LoggingSettings{
		Level:      "info",
		Format:     "text",
		File:       "",
		WithCaller: false,
		Verbose:    false,
	}
	return settings.SetupLogger()
}
