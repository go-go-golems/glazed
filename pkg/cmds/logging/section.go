package logging

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// LoggingSettings holds the logging configuration fields
type LoggingSettings struct {
	WithCaller     bool              `glazed:"with-caller" yaml:"with-caller" json:"with-caller"`
	LogLevel       string            `glazed:"log-level" yaml:"log-level" json:"log-level"`
	LogFormat      string            `glazed:"log-format" yaml:"log-format" json:"log-format"`
	LogFile        string            `glazed:"log-file" yaml:"log-file" json:"log-file"`
	LogToStdout    bool              `glazed:"log-to-stdout" yaml:"log-to-stdout" json:"log-to-stdout"`
	LogConfigFiles []string          `glazed:"log-config" yaml:"log-config" json:"log-config"`
	LogAreas       map[string]string `glazed:"log-area" yaml:"log-area" json:"log-area"`
	Areas          map[string]string `glazed:"areas" yaml:"areas" json:"areas"`
	StrictAreas    bool              `glazed:"strict-log-areas" yaml:"strict-log-areas" json:"strict-log-areas"`
}

const LoggingSectionSlug = "logging"

// NewLoggingSection creates a new section for logging configuration
func NewLoggingSection() (schema.Section, error) {
	return schema.NewSection(
		LoggingSectionSlug,
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
				"log-config",
				fields.TypeStringList,
				fields.WithHelp("Additional logcopter profile/config file; repeatable"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"log-area",
				fields.TypeKeyValue,
				fields.WithHelp("Per-area log level override, for example app.view:debug or app.db=warn"),
				fields.WithDefault(map[string]string{}),
			),
			fields.New(
				"areas",
				fields.TypeKeyValue,
				fields.WithHelp("Per-area log level map for configuration files"),
				fields.WithDefault(map[string]string{}),
			),
			fields.New(
				"strict-log-areas",
				fields.TypeBool,
				fields.WithHelp("Fail when configured log areas do not match known generated logcopter areas"),
				fields.WithDefault(false),
			),
		),
	)
}

// AddLoggingSectionToCommand adds the logging section to a Glazed command
func AddLoggingSectionToCommand(cmd cmds.Command) (cmds.Command, error) {
	loggingSection, err := NewLoggingSection()
	if err != nil {
		return nil, err
	}

	cmd.Description().Schema.Set(LoggingSectionSlug, loggingSection)

	return cmd, nil
}

func AddLoggingSectionToRootCommand(rootCmd *cobra.Command, appName string) error {
	loggingSection, err := NewLoggingSection()
	if err != nil {
		return err
	}
	_ = loggingSection

	// XXX this would be the proper way to do it if we could easily add field definitions as persistent flags. For now, do it manually.
	// Don't delete this code
	// ---
	// loggingSection.GetDefinitions().ForEachE(func(definition *fields.Definition) error {
	// 	rootCmd.PersistentFlags().String(definition.Name, definition.Default, definition.Help)
	// 	return nil
	// })

	rootCmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file (default: stderr)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (json, text)")
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller information")
	rootCmd.PersistentFlags().Bool("log-to-stdout", false, "Log to stdout even when log-file is set")
	rootCmd.PersistentFlags().StringSlice("log-config", []string{}, "Additional logcopter profile/config file; repeatable")
	rootCmd.PersistentFlags().StringSlice("log-area", []string{}, "Per-area log level override, for example app.view:debug or app.db=warn")
	rootCmd.PersistentFlags().Bool("strict-log-areas", false, "Fail when configured log areas do not match known generated logcopter areas")

	return nil
}

// SetupLoggingFromValues configures global logger from command-line fields
func SetupLoggingFromValues(parsedValues *values.Values) error {
	settings, err := GetLoggingSettings(parsedValues)
	if err != nil {
		return fmt.Errorf("failed to get logging settings: %w", err)
	}
	return InitLoggerFromSettings(settings)
}

// GetLoggingSettings extracts logging configuration for custom validation or setup
func GetLoggingSettings(parsedValues *values.Values) (*LoggingSettings, error) {
	var settings LoggingSettings
	err := parsedValues.DecodeSectionInto(LoggingSectionSlug, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
	}
	return &settings, nil
}
