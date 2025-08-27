package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
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
		logWriter = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()),
			TimeFormat: time.RFC3339Nano,
		}
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

		// If LogToStdout is enabled, log to both file and stdout
		if settings.LogToStdout {
			logWriter = io.MultiWriter(logWriter, writer)
		} else {
			logWriter = writer
		}
	}

	// Configure Logstash logging if enabled
	if settings.LogstashEnabled {
		// Use default app name if not specified
		appName := settings.LogstashAppName
		if appName == "" {
			appName = viper.GetEnvPrefix()
			if appName == "" {
				appName = "app"
			}
		}

		logstashWriter := SetupLogstashLogger(
			settings.LogstashHost,
			settings.LogstashPort,
			settings.LogstashProtocol,
			appName,
			settings.LogstashEnvironment,
		)

		// Create a multi-writer that logs to both the existing writer and Logstash
		logWriter = zerolog.MultiLevelWriter(logWriter, logstashWriter)

		log.Info().
			Str("host", settings.LogstashHost).
			Int("port", settings.LogstashPort).
			Str("protocol", settings.LogstashProtocol).
			Str("app", appName).
			Str("environment", settings.LogstashEnvironment).
			Msg("Logging to Logstash")
	}

	log.Logger = log.Output(logWriter)

	// Set the default context logger
	zerolog.DefaultContextLogger = &log.Logger

	switch strings.ToLower(settings.LogLevel) {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
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
		Bool("logToStdout", settings.LogToStdout).
		Bool("logstash", settings.LogstashEnabled).
		Msg("Logger initialized")

	return nil
}

// InitLoggerFromViper initializes the logger using settings from Viper
func InitLoggerFromViper() error {
	settings := &LoggingSettings{
		LogLevel:            viper.GetString("log-level"),
		LogFile:             viper.GetString("log-file"),
		LogFormat:           viper.GetString("log-format"),
		WithCaller:          viper.GetBool("with-caller"),
		LogToStdout:         viper.GetBool("log-to-stdout"),
		LogstashEnabled:     viper.GetBool("logstash-enabled"),
		LogstashHost:        viper.GetString("logstash-host"),
		LogstashPort:        viper.GetInt("logstash-port"),
		LogstashProtocol:    viper.GetString("logstash-protocol"),
		LogstashAppName:     viper.GetString("logstash-app-name"),
		LogstashEnvironment: viper.GetString("logstash-environment"),
	}

	return InitLoggerFromSettings(settings)
}
