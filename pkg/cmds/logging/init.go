package logging

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLoggerFromSettings initializes the logger based on the provided settings
func InitLoggerFromSettings(settings *LoggingSettings) error {
	if settings.WithCaller {
		log.Logger = log.With().Caller().Logger()
	}

	// Set timestamp format to include milliseconds
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// default is json
	var logWriter io.Writer
	if settings.LogFormat == "text" {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	if settings.LogFile != "" {
		fileLogger := &lumberjack.Logger{
			Filename:   settings.LogFile,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28,    //days
			Compress:   false, // disabled by default
		}
		var writer io.Writer
		writer = fileLogger
		if settings.LogFormat == "text" {
			log.Info().Str("file", settings.LogFile).Msg("Logging to file")
			writer = zerolog.ConsoleWriter{
				NoColor:    true,
				Out:        fileLogger,
				TimeFormat: time.RFC3339Nano,
			}
		}
		// TODO(manuel, 2024-07-05) We used to support logging to file *and* stderr, but disabling that for now
		// because it makes logging in UI apps tricky.
		// logWriter = io.MultiWriter(logWriter, writer)
		logWriter = writer
	}

	log.Logger = log.Output(logWriter)

	logLevel := strings.ToLower(settings.LogLevel)
	if settings.Verbose && logLevel != "trace" {
		logLevel = "debug"
	}

	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}

	log.Logger.Debug().Str("format", settings.LogFormat).
		Str("level", settings.LogLevel).
		Str("file", settings.LogFile).
		Msg("Logger initialized")

	return nil
}

// InitLoggerFromViper initializes the logger using settings from Viper
func InitLoggerFromViper() error {
	settings := &LoggingSettings{
		LogLevel:   viper.GetString("log-level"),
		LogFile:    viper.GetString("log-file"),
		LogFormat:  viper.GetString("log-format"),
		WithCaller: viper.GetBool("with-caller"),
		Verbose:    viper.GetBool("verbose"),
	}

	return InitLoggerFromSettings(settings)
}

// AddLoggingLayerToCobra adds the logging layer to a Cobra command
func AddLoggingLayerToCobra(rootCmd *cobra.Command) error {
	// Create a logging layer to get the parameter definitions
	loggingLayer, err := NewLoggingLayer()
	if err != nil {
		return err
	}

	// Get parameter definitions from the layer
	if cobraLayer, ok := loggingLayer.(layers.CobraParameterLayer); ok {
		err := cobraLayer.AddLayerToCobraCommand(rootCmd)
		if err != nil {
			return err
		}
	} else {
		return errors.New("logging layer is not a CobraParameterLayer")
	}

	return nil
}

// InitLoggerFromParsedLayers initializes the logger from parsed Glazed layers
func InitLoggerFromParsedLayers(parsedLayers *layers.ParsedLayers) error {
	settings, err := GetLoggingSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	return InitLoggerFromSettings(settings)
}
