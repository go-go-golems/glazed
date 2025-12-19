package logging

import (
	"io"
	"strings"

	"github.com/spf13/pflag"
)

// filterEarlyLoggingArgs filters args to only include logging-related flags.
// This avoids pflag stopping early when it encounters unknown flags (which is
// expected before we register all cobra subcommands).
func filterEarlyLoggingArgs(args []string) []string {
	allowedKV := map[string]struct{}{
		"--log-level":            {},
		"--log-file":             {},
		"--log-format":           {},
		"--logstash-host":        {},
		"--logstash-port":        {},
		"--logstash-protocol":    {},
		"--logstash-app-name":    {},
		"--logstash-environment": {},
	}
	allowedBool := map[string]struct{}{
		"--with-caller":      {},
		"--log-to-stdout":    {},
		"--logstash-enabled": {},
	}

	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]

		// Handle --flag=value form
		if strings.HasPrefix(a, "--") && strings.Contains(a, "=") {
			name := a[:strings.Index(a, "=")]
			if _, ok := allowedKV[name]; ok {
				out = append(out, a)
				continue
			}
			if _, ok := allowedBool[name]; ok {
				out = append(out, a)
				continue
			}
			continue
		}

		// Handle bare bool flags
		if _, ok := allowedBool[a]; ok {
			out = append(out, a)
			continue
		}

		// Handle --flag value form
		if _, ok := allowedKV[a]; ok {
			out = append(out, a)
			if i+1 < len(args) {
				out = append(out, args[i+1])
				i++
			}
			continue
		}
	}

	return out
}

// InitEarlyLoggingFromArgs initializes logging from command-line arguments before
// cobra commands are registered. This allows logging configuration (like --log-level)
// to be respected during command discovery/loading.
//
// This function:
// - Filters args to only logging flags (ignores unknown flags)
// - Parses flags using a standalone pflag.FlagSet
// - Initializes the global logger with parsed settings
//
// Defaults match AddLoggingLayerToRootCommand in layer.go.
func InitEarlyLoggingFromArgs(args []string, appName string) error {
	// We want to initialize logging before we load/register commands, so that any
	// logging during command discovery respects --log-level etc.
	//
	// We cannot use rootCmd.ParseFlags() here because:
	// - it errors on --help ("pflag: help requested") and
	// - it would fail on unknown flags (all command-specific flags) before we
	//   have registered those commands.
	//
	// So: pre-parse ONLY logging flags from args, ignoring everything else.
	fs := pflag.NewFlagSet("early-logging", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.SetInterspersed(true)

	// Defaults must match glazed/pkg/cmds/logging/layer.go:AddLoggingLayerToRootCommand
	logLevel := fs.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal)")
	logFile := fs.String("log-file", "", "Log file (default: stderr)")
	logFormat := fs.String("log-format", "text", "Log format (json, text)")
	withCaller := fs.Bool("with-caller", false, "Log caller information")
	logToStdout := fs.Bool("log-to-stdout", false, "Log to stdout even when log-file is set")

	logstashEnabled := fs.Bool("logstash-enabled", false, "Enable logging to Logstash")
	logstashHost := fs.String("logstash-host", "logstash", "Logstash host")
	logstashPort := fs.Int("logstash-port", 5044, "Logstash port")
	logstashProtocol := fs.String("logstash-protocol", "tcp", "Logstash protocol (tcp, udp)")
	logstashAppName := fs.String("logstash-app-name", appName, "Application name for Logstash logs")
	logstashEnvironment := fs.String("logstash-environment", "development", "Environment name for Logstash logs (development, staging, production)")

	fs.ParseErrorsAllowlist.UnknownFlags = true
	// Always attempt parsing, but never fail early logging init on parsing errors.
	// The critical behavior is: default to info-level logging (quiet), and if we
	// successfully parse --log-level etc, honor them.
	filteredArgs := filterEarlyLoggingArgs(args)
	_ = fs.Parse(filteredArgs)

	return InitLoggerFromSettings(&LoggingSettings{
		LogLevel:            *logLevel,
		LogFile:             *logFile,
		LogFormat:           *logFormat,
		WithCaller:          *withCaller,
		LogToStdout:         *logToStdout,
		LogstashEnabled:     *logstashEnabled,
		LogstashHost:        *logstashHost,
		LogstashPort:        *logstashPort,
		LogstashProtocol:    *logstashProtocol,
		LogstashAppName:     *logstashAppName,
		LogstashEnvironment: *logstashEnvironment,
	})
}
