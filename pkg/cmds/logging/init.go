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

	settings := &LoggingSettings{
		LogLevel:    logLevel,
		LogFile:     logFile,
		LogFormat:   logFormat,
		WithCaller:  withCaller,
		LogToStdout: logToStdout,
	}

	return InitLoggerFromSettings(settings)
}
