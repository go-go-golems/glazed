package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	switch settings.LogLevel {
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
	}

	return InitLoggerFromSettings(settings)
}
