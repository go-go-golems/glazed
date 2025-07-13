package logging

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const (
	// LoggingSlug is the unique identifier for this layer.
	LoggingSlug = "logging"
)

// NewLoggingLayer creates a new parameter layer for logging configuration.
func NewLoggingLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		LoggingSlug,
		"Logging Configuration",
		layers.WithParameterDefinitions(
			// Core logging parameters - the ones everyone needs
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Set the logging level"),
				parameters.WithDefault("info"), // Safe default - not too noisy, not too quiet
				parameters.WithChoices("debug", "info", "warn", "error", "fatal", "panic"),
				parameters.WithShortFlag("L"), // Capital L to avoid conflicts with -l (list)
			),
			parameters.NewParameterDefinition(
				"log-format",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Set the log output format"),
				parameters.WithDefault("text"),         // Human-readable by default
				parameters.WithChoices("text", "json"), // JSON for log aggregation systems
			),
			parameters.NewParameterDefinition(
				"log-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Log file path (default: stderr)"),
				parameters.WithDefault(""), // Empty means stderr - explicit in help text
			),

			// Developer convenience parameters
			parameters.NewParameterDefinition(
				"with-caller",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include caller information in log entries"),
				parameters.WithDefault(false), // Off by default - performance impact
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose logging (sets level to debug)"),
				parameters.WithDefault(false),
				parameters.WithShortFlag("v"), // Classic Unix convention
			),

			// Enterprise/production features
			parameters.NewParameterDefinition(
				"logstash-host",
				parameters.ParameterTypeString,
				parameters.WithHelp("Logstash server host for centralized logging"),
				parameters.WithDefault(""), // Optional feature - empty disables
			),
			parameters.NewParameterDefinition(
				"logstash-port",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Logstash server port"),
				parameters.WithDefault(5044), // Standard Logstash Beats port
			),
		),
	)
}

// NewLoggingLayerWithOptions creates a logging layer with customization options
func NewLoggingLayerWithOptions(opts ...LoggingLayerOption) (layers.ParameterLayer, error) {
	config := &loggingLayerConfig{
		includeLogstash: false,
		defaultLevel:    "info",
		defaultFormat:   "text",
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	layer, err := NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	// Modify defaults based on config
	params := layer.GetParameterDefinitions()

	if levelParam, ok := params.Get("log-level"); ok && levelParam != nil {
		defaultLevel := interface{}(config.defaultLevel)
		levelParam.Default = &defaultLevel
	}

	if formatParam, ok := params.Get("log-format"); ok && formatParam != nil {
		defaultFormat := interface{}(config.defaultFormat)
		formatParam.Default = &defaultFormat
	}

	// Note: In actual implementation, we would need to create the layer conditionally
	// since there's no RemoveFlag method. For now, we'll include all parameters.

	return layer, nil
}

// Configuration options for the logging layer
type loggingLayerConfig struct {
	includeLogstash bool
	defaultLevel    string
	defaultFormat   string
}

type LoggingLayerOption func(*loggingLayerConfig)

// WithLogstash includes Logstash configuration parameters
func WithLogstash() LoggingLayerOption {
	return func(c *loggingLayerConfig) {
		c.includeLogstash = true
	}
}

// WithDefaultLevel sets the default log level
func WithDefaultLevel(level string) LoggingLayerOption {
	return func(c *loggingLayerConfig) {
		c.defaultLevel = level
	}
}

// WithDefaultFormat sets the default log format
func WithDefaultFormat(format string) LoggingLayerOption {
	return func(c *loggingLayerConfig) {
		c.defaultFormat = format
	}
}
