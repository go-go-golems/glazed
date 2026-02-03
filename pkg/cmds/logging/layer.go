package logging

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// LoggingSettings holds the logging configuration parameters
type LoggingSettings struct {
	WithCaller          bool   `glazed:"with-caller"`
	LogLevel            string `glazed:"log-level"`
	LogFormat           string `glazed:"log-format"`
	LogFile             string `glazed:"log-file"`
	LogToStdout         bool   `glazed:"log-to-stdout"`
	LogstashEnabled     bool   `glazed:"logstash-enabled"`
	LogstashHost        string `glazed:"logstash-host"`
	LogstashPort        int    `glazed:"logstash-port"`
	LogstashProtocol    string `glazed:"logstash-protocol"`
	LogstashAppName     string `glazed:"logstash-app-name"`
	LogstashEnvironment string `glazed:"logstash-environment"`
}

const LoggingLayerSlug = "logging"

// NewLoggingLayer creates a new parameter layer for logging configuration
func NewLoggingLayer() (schema.Section, error) {
	return schema.NewSection(
		LoggingLayerSlug,
		"Logging configuration options",
		schema.WithFields(
			fields.New(
				"with-caller",
				fields.TypeBool,
				fields.WithHelp("Log caller information"),
				fields.WithDefault(false),
			),
			fields.New(
				"log-level",
				fields.TypeChoice,
				fields.WithHelp("Log level (trace, debug, info, warn, error, fatal)"),
				fields.WithDefault("info"),
				fields.WithChoices("trace", "debug", "info", "warn", "error", "fatal", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"),
			),
			fields.New(
				"log-format",
				fields.TypeChoice,
				fields.WithHelp("Log format (json, text)"),
				fields.WithDefault("text"),
				fields.WithChoices("json", "text"),
			),
			fields.New(
				"log-file",
				fields.TypeString,
				fields.WithHelp("Log file (default: stderr)"),
				fields.WithDefault(""),
			),
			fields.New(
				"log-to-stdout",
				fields.TypeBool,
				fields.WithHelp("Log to stdout even when log-file is set"),
				fields.WithDefault(false),
			),
			fields.New(
				"logstash-enabled",
				fields.TypeBool,
				fields.WithHelp("Enable logging to Logstash"),
				fields.WithDefault(false),
			),
			fields.New(
				"logstash-host",
				fields.TypeString,
				fields.WithHelp("Logstash host"),
				fields.WithDefault("logstash"),
			),
			fields.New(
				"logstash-port",
				fields.TypeInteger,
				fields.WithHelp("Logstash port"),
				fields.WithDefault(5044),
			),
			fields.New(
				"logstash-protocol",
				fields.TypeChoice,
				fields.WithHelp("Logstash protocol (tcp, udp)"),
				fields.WithDefault("tcp"),
				fields.WithChoices("tcp", "udp"),
			),
			fields.New(
				"logstash-app-name",
				fields.TypeString,
				fields.WithHelp("Application name for Logstash logs"),
				fields.WithDefault(""),
			),
			fields.New(
				"logstash-environment",
				fields.TypeString,
				fields.WithHelp("Environment name for Logstash logs (development, staging, production)"),
				fields.WithDefault("development"),
				fields.WithChoices("development", "staging", "production"),
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
	// loggingLayer.GetDefinitions().ForEachE(func(definition *fields.Definition) error {
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

// SetupLoggingFromValues configures global logger from command-line parameters
func SetupLoggingFromValues(parsedLayers *values.Values) error {
	settings, err := GetLoggingSettings(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to get logging settings: %w", err)
	}
	return InitLoggerFromSettings(settings)
}

// GetLoggingSettings extracts logging configuration for custom validation or setup
func GetLoggingSettings(parsedLayers *values.Values) (*LoggingSettings, error) {
	var settings LoggingSettings
	err := parsedLayers.InitializeStruct(LoggingLayerSlug, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
	}
	return &settings, nil
}
