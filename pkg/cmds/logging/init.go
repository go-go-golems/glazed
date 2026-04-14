package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
		// Use a stable fallback app name if not specified explicitly.
		appName := settings.LogstashAppName
		if appName == "" {
			appName = "app"
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

// InitLoggerFromCobra initializes the logger using flags parsed by Cobra on the given command.
// Call this in PersistentPreRun(E) to initialize logging after Cobra parsed flags but before command execution.
// Flags are added by AddLoggingSectionToRootCommand.
func InitLoggerFromCobra(cmd *cobra.Command) error {
	if cmd == nil {
		return errors.Errorf("nil cobra command passed to InitLoggerFromCobra")
	}

	getString := func(name string) (string, error) { return cmd.Flags().GetString(name) }
	getBool := func(name string) (bool, error) { return cmd.Flags().GetBool(name) }
	getInt := func(name string) (int, error) { return cmd.Flags().GetInt(name) }

	logLevel, err := getString("log-level")
	if err != nil {
		return errors.Wrap(err, "reading --log-level")
	}
	logFile, err := getString("log-file")
	if err != nil {
		return errors.Wrap(err, "reading --log-file")
	}
	logFormat, err := getString("log-format")
	if err != nil {
		return errors.Wrap(err, "reading --log-format")
	}
	withCaller, err := getBool("with-caller")
	if err != nil {
		return errors.Wrap(err, "reading --with-caller")
	}
	logToStdout, err := getBool("log-to-stdout")
	if err != nil {
		return errors.Wrap(err, "reading --log-to-stdout")
	}
	lsEnabled, err := getBool("logstash-enabled")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-enabled")
	}
	lsHost, err := getString("logstash-host")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-host")
	}
	lsPort, err := getInt("logstash-port")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-port")
	}
	lsProto, err := getString("logstash-protocol")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-protocol")
	}
	lsAppName, err := getString("logstash-app-name")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-app-name")
	}
	lsEnv, err := getString("logstash-environment")
	if err != nil {
		return errors.Wrap(err, "reading --logstash-environment")
	}

	settings := &LoggingSettings{
		LogLevel:            logLevel,
		LogFile:             logFile,
		LogFormat:           logFormat,
		WithCaller:          withCaller,
		LogToStdout:         logToStdout,
		LogstashEnabled:     lsEnabled,
		LogstashHost:        lsHost,
		LogstashPort:        lsPort,
		LogstashProtocol:    lsProto,
		LogstashAppName:     lsAppName,
		LogstashEnvironment: lsEnv,
	}

	return InitLoggerFromSettings(settings)
}
