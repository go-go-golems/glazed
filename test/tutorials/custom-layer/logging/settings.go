package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggingSettings represents all logging configuration options.
// This struct serves as both the parameter binding target and the
// configuration container for logger initialization.
type LoggingSettings struct {
	// Core logging settings - the 80% use case
	Level  string `glazed.parameter:"log-level"`
	Format string `glazed.parameter:"log-format"`
	File   string `glazed.parameter:"log-file"`

	// Developer convenience settings
	WithCaller bool `glazed.parameter:"with-caller"`
	Verbose    bool `glazed.parameter:"verbose"`

	// Production/enterprise features
	LogstashHost string `glazed.parameter:"logstash-host"`
	LogstashPort int    `glazed.parameter:"logstash-port"`
}

// Validate checks if the logging settings are valid.
// Input validation prevents runtime failures from invalid configuration.
func (s *LoggingSettings) Validate() error {
	// Validate log level - catch typos early
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, s.Level) {
		return fmt.Errorf("invalid log level '%s', must be one of: %s",
			s.Level, strings.Join(validLevels, ", "))
	}

	// Validate log format - prevent silent failures in log parsing
	validFormats := []string{"text", "json"}
	if !contains(validFormats, s.Format) {
		return fmt.Errorf("invalid log format '%s', must be one of: %s",
			s.Format, strings.Join(validFormats, ", "))
	}

	// Validate logstash configuration - partial config leads to confusion
	if s.LogstashHost != "" && (s.LogstashPort < 1 || s.LogstashPort > 65535) {
		return fmt.Errorf("logstash port must be between 1 and 65535, got %d", s.LogstashPort)
	}

	return nil
}

// GetLogLevel converts string level to zerolog.Level.
// The verbose flag overrides the configured level for debugging convenience.
func (s *LoggingSettings) GetLogLevel() zerolog.Level {
	// Verbose flag takes precedence over configured level
	if s.Verbose {
		return zerolog.DebugLevel
	}

	switch strings.ToLower(s.Level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		// Default to info level for balanced logging
		return zerolog.InfoLevel
	}
}

// GetWriter returns the appropriate writer for log output.
// Defaults to stderr to separate log output from program output.
func (s *LoggingSettings) GetWriter() (io.Writer, error) {
	if s.File == "" {
		// Use stderr for log output to avoid mixing with program output
		return os.Stderr, nil
	}

	// Append to log files to preserve history across restarts
	file, err := os.OpenFile(s.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file '%s': %w", s.File, err)
	}

	return file, nil
}

// SetupLogger configures the global logger with these settings
func (s *LoggingSettings) SetupLogger() error {
	// Validate settings first
	if err := s.Validate(); err != nil {
		return err
	}

	// Set log level
	zerolog.SetGlobalLevel(s.GetLogLevel())

	// Get writer
	writer, err := s.GetWriter()
	if err != nil {
		return err
	}

	// Configure output format
	var output io.Writer = writer
	if s.Format == "text" {
		// Pretty console output for text format
		if s.File == "" { // Only if writing to stderr
			output = zerolog.ConsoleWriter{
				Out:        writer,
				TimeFormat: time.RFC3339,
				NoColor:    false,
			}
		}
	}

	// Create logger
	logger := zerolog.New(output).With().Timestamp()

	// Add caller information if requested
	if s.WithCaller {
		logger = logger.Caller()
	}

	// Set as global logger
	log.Logger = logger.Logger()

	return nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
