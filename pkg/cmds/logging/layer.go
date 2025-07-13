package logging

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// LoggingSettings holds the logging configuration parameters
type LoggingSettings struct {
	WithCaller          bool   `glazed.parameter:"with-caller"`
	LogLevel            string `glazed.parameter:"log-level"`
	LogFormat           string `glazed.parameter:"log-format"`
	LogFile             string `glazed.parameter:"log-file"`
	LogToStdout         bool   `glazed.parameter:"log-to-stdout"`
	LogstashEnabled     bool   `glazed.parameter:"logstash-enabled"`
	LogstashHost        string `glazed.parameter:"logstash-host"`
	LogstashPort        int    `glazed.parameter:"logstash-port"`
	LogstashProtocol    string `glazed.parameter:"logstash-protocol"`
	LogstashAppName     string `glazed.parameter:"logstash-app-name"`
	LogstashEnvironment string `glazed.parameter:"logstash-environment"`
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
				parameters.WithHelp("Log level (trace, debug, info, warn, error, fatal)"),
				parameters.WithDefault("info"),
				parameters.WithChoices("trace", "debug", "info", "warn", "error", "fatal", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"),
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
				"log-to-stdout",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Log to stdout even when log-file is set"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"logstash-enabled",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable logging to Logstash"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"logstash-host",
				parameters.ParameterTypeString,
				parameters.WithHelp("Logstash host"),
				parameters.WithDefault("logstash"),
			),
			parameters.NewParameterDefinition(
				"logstash-port",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Logstash port"),
				parameters.WithDefault(5044),
			),
			parameters.NewParameterDefinition(
				"logstash-protocol",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Logstash protocol (tcp, udp)"),
				parameters.WithDefault("tcp"),
				parameters.WithChoices("tcp", "udp"),
			),
			parameters.NewParameterDefinition(
				"logstash-app-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Application name for Logstash logs"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"logstash-environment",
				parameters.ParameterTypeString,
				parameters.WithHelp("Environment name for Logstash logs (development, staging, production)"),
				parameters.WithDefault("development"),
				parameters.WithChoices("development", "staging", "production"),
			),
		),
	)
}

// AddLoggingLayerToCommand adds the logging layer to a Glazed command
func AddLoggingLayerToCommand(cmd cmds.Command) (cmds.Command, error) {
	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	cmd.Description().Layers.Set(LoggingLayerSlug, loggingLayer)

	return cmd, nil
}

func AddLoggingLayerToRootCommand(rootCmd *cobra.Command, appName string) error {
	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return err
	}
	_ = loggingLayer

	// XXX this would be the proper way to do it if we could easily add parameter definitions as persistent flags. For now, do it manually.
	// Don't delete this code
	// ---
	// loggingLayer.GetParameterDefinitions().ForEachE(func(definition *parameters.ParameterDefinition) error {
	// 	rootCmd.PersistentFlags().String(definition.Name, definition.Default, definition.Help)
	// 	return nil
	// })

	rootCmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file (default: stderr)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (json, text)")
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller information")
	rootCmd.PersistentFlags().Bool("log-to-stdout", false, "Log to stdout even when log-file is set")

	// Add logstash flags
	rootCmd.PersistentFlags().Bool("logstash-enabled", false, "Enable logging to Logstash")
	rootCmd.PersistentFlags().String("logstash-host", "logstash", "Logstash host")
	rootCmd.PersistentFlags().Int("logstash-port", 5044, "Logstash port")
	rootCmd.PersistentFlags().String("logstash-protocol", "tcp", "Logstash protocol (tcp, udp)")
	rootCmd.PersistentFlags().String("logstash-app-name", appName, "Application name for Logstash logs")
	rootCmd.PersistentFlags().String("logstash-environment", "development", "Environment name for Logstash logs (development, staging, production)")

	return nil
}

// SetupLoggingFromParsedLayers configures global logger from command-line parameters
func SetupLoggingFromParsedLayers(parsedLayers *layers.ParsedLayers) error {
	settings, err := GetLoggingSettings(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to get logging settings: %w", err)
	}
	return InitLoggerFromSettings(settings)
}

// GetLoggingSettings extracts logging configuration for custom validation or setup
func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
	var settings LoggingSettings
	err := parsedLayers.InitializeStruct(LoggingLayerSlug, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
	}
	return &settings, nil
}
