package logging

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// LoggingSettings holds the logging configuration parameters
type LoggingSettings struct {
	WithCaller bool   `glazed.parameter:"with-caller"`
	LogLevel   string `glazed.parameter:"log-level"`
	LogFormat  string `glazed.parameter:"log-format"`
	LogFile    string `glazed.parameter:"log-file"`
	Verbose    bool   `glazed.parameter:"verbose"`
}

const LoggingLayerSlug = "logging"

// NewLoggingLayer creates a new parameter layer for logging configuration
func NewLoggingLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		LoggingLayerSlug,
		"Logging configuration options",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"with-caller",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Log caller information"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log level (debug, info, warn, error, fatal)"),
				parameters.WithDefault("info"),
				parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
			),
			parameters.NewParameterDefinition(
				"log-format",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log format (json, text)"),
				parameters.WithDefault("text"),
				parameters.WithChoices("json", "text"),
			),
			parameters.NewParameterDefinition(
				"log-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Log file (default: stderr)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Verbose output"),
				parameters.WithDefault(false),
			),
		),
	)
}

// GetLoggingSettingsFromParsedLayers extracts logging settings from parsed layers
func GetLoggingSettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
	s := &LoggingSettings{}
	if err := parsedLayers.InitializeStruct("logging", s); err != nil {
		return nil, err
	}
	return s, nil
}
