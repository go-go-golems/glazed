package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/logcopter/pkg/logcopter"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

// InitLoggerFromSettings initializes the logger based on the provided settings.
func InitLoggerFromSettings(settings *LoggingSettings) error {
	if settings == nil {
		settings = DefaultLoggingSettings()
	}
	merged, err := MergeLoggingSettings(DefaultLoggingSettings(), settings)
	if err != nil {
		return err
	}
	return initLoggerFromMergedSettings(merged)
}

func initLoggerFromMergedSettings(settings *LoggingSettings) error {
	// Set timestamp format to include milliseconds/nanos.
	zerolog.TimeFieldFormat = time.RFC3339Nano

	baseWriter, err := loggingWriter(settings)
	if err != nil {
		return err
	}

	base := zerolog.New(baseWriter).With().Timestamp().Logger()
	if settings.WithCaller {
		base = base.With().Caller().Logger()
	}

	defaultLevel, err := logcopter.ParseLevel(settings.LogLevel)
	if err != nil {
		return errors.Wrap(err, "parse log level")
	}

	// Keep zerolog's process-wide gate open; normal filtering is done on the
	// global logger and on logcopter's per-area child loggers. This lets one area
	// emit trace while another remains at warn.
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = base.Level(defaultLevel)
	zerolog.DefaultContextLogger = &log.Logger

	areas := mergeStringMaps(settings.Areas, settings.LogAreas)
	if err := logcopter.Configure(base, logcopter.Config{
		Level:       settings.LogLevel,
		Format:      settings.LogFormat,
		Output:      outputName(settings),
		Caller:      settings.WithCaller,
		Timestamp:   true,
		Areas:       areas,
		StrictAreas: settings.StrictAreas,
	}); err != nil {
		return err
	}

	log.Logger.Debug().Str("format", settings.LogFormat).
		Str("level", settings.LogLevel).
		Str("file", settings.LogFile).
		Bool("logToStdout", settings.LogToStdout).
		Int("areas", len(areas)).
		Msg("Logger initialized")

	return nil
}

func loggingWriter(settings *LoggingSettings) (io.Writer, error) {
	stream := io.Writer(os.Stderr)
	if settings.LogToStdout && settings.LogFile == "" {
		stream = os.Stdout
	}

	var logWriter io.Writer
	if strings.EqualFold(settings.LogFormat, "json") {
		logWriter = stream
	} else {
		logWriter = zerolog.ConsoleWriter{
			Out:        stream,
			NoColor:    !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()),
			TimeFormat: time.RFC3339Nano,
		}
	}

	if settings.LogFile == "" {
		return logWriter, nil
	}

	fileLogger := &lumberjack.Logger{
		Filename:   settings.LogFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    // days
		Compress:   false, // disabled by default
	}
	var fileWriter io.Writer = fileLogger
	if strings.EqualFold(settings.LogFormat, "text") || settings.LogFormat == "" {
		fileWriter = zerolog.ConsoleWriter{
			NoColor:    true,
			Out:        fileLogger,
			TimeFormat: time.RFC3339Nano,
		}
	}

	if settings.LogToStdout {
		return io.MultiWriter(logWriter, fileWriter), nil
	}
	return fileWriter, nil
}

func outputName(settings *LoggingSettings) string {
	if settings.LogToStdout && settings.LogFile == "" {
		return logcopter.OutputStdout
	}
	return logcopter.OutputStderr
}

// InitLoggerFromCobra initializes the logger using flags parsed by Cobra on the given command.
// Call this in PersistentPreRun(E) to initialize logging after Cobra parsed flags but before command execution.
// Flags are added by AddLoggingSectionToRootCommand.
func InitLoggerFromCobra(cmd *cobra.Command) error {
	if cmd == nil {
		return errors.Errorf("nil cobra command passed to InitLoggerFromCobra")
	}

	settings := DefaultLoggingSettings()
	if v, err := cmd.Flags().GetStringSlice("log-config"); err == nil {
		settings.LogConfigFiles = v
	}
	merged, err := MergeLoggingSettings(DefaultLoggingSettings(), settings)
	if err != nil {
		return err
	}

	flags := cmd.Flags()
	if flags.Changed("log-level") {
		v, err := flags.GetString("log-level")
		if err != nil {
			return errors.Wrap(err, "reading --log-level")
		}
		merged.LogLevel = v
	}
	if flags.Changed("log-file") {
		v, err := flags.GetString("log-file")
		if err != nil {
			return errors.Wrap(err, "reading --log-file")
		}
		merged.LogFile = v
	}
	if flags.Changed("log-format") {
		v, err := flags.GetString("log-format")
		if err != nil {
			return errors.Wrap(err, "reading --log-format")
		}
		merged.LogFormat = v
	}
	if flags.Changed("with-caller") {
		v, err := flags.GetBool("with-caller")
		if err != nil {
			return errors.Wrap(err, "reading --with-caller")
		}
		merged.WithCaller = v
	}
	if flags.Changed("log-to-stdout") {
		v, err := flags.GetBool("log-to-stdout")
		if err != nil {
			return errors.Wrap(err, "reading --log-to-stdout")
		}
		merged.LogToStdout = v
	}
	if flags.Changed("strict-log-areas") {
		v, err := flags.GetBool("strict-log-areas")
		if err != nil {
			return errors.Wrap(err, "reading --strict-log-areas")
		}
		merged.StrictAreas = v
	}
	if flags.Changed("log-area") {
		v, err := flags.GetStringSlice("log-area")
		if err != nil {
			return errors.Wrap(err, "reading --log-area")
		}
		areas, err := ParseAreaOverrides(v)
		if err != nil {
			return err
		}
		merged.LogAreas = mergeStringMaps(merged.LogAreas, areas)
	}

	return initLoggerFromMergedSettings(merged)
}

func DefaultLoggingSettings() *LoggingSettings {
	return &LoggingSettings{
		LogLevel:       "info",
		LogFormat:      "text",
		LogFile:        "",
		LogConfigFiles: []string{},
		LogAreas:       map[string]string{},
		Areas:          map[string]string{},
	}
}

// MergeLoggingSettings applies deterministic precedence: base, explicit
// log-config files in order, then direct settings.
func MergeLoggingSettings(base *LoggingSettings, direct *LoggingSettings) (*LoggingSettings, error) {
	merged := cloneLoggingSettings(base)
	if merged == nil {
		merged = DefaultLoggingSettings()
	}
	if direct == nil {
		return merged, nil
	}

	for _, path := range direct.LogConfigFiles {
		profile, err := LoadLoggingSettingsFile(path)
		if err != nil {
			return nil, err
		}
		applyLoggingSettings(merged, profile, true)
	}
	applyLoggingSettings(merged, direct, false)
	return merged, nil
}

func LoadLoggingSettingsFile(path string) (*LoggingSettings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read log config %s", path)
	}

	var root map[string]interface{}
	if err := yaml.Unmarshal(b, &root); err != nil {
		return nil, errors.Wrapf(err, "parse log config %s", path)
	}
	if loggingNode, ok := root[LoggingSectionSlug]; ok {
		b, err = yaml.Marshal(loggingNode)
		if err != nil {
			return nil, err
		}
	}

	var settings LoggingSettings
	if err := yaml.Unmarshal(b, &settings); err != nil {
		return nil, errors.Wrapf(err, "decode log config %s", path)
	}
	var direct struct {
		Level       string            `yaml:"level"`
		Format      string            `yaml:"format"`
		Output      string            `yaml:"output"`
		Caller      bool              `yaml:"caller"`
		Areas       map[string]string `yaml:"areas"`
		StrictAreas bool              `yaml:"strict_areas"`
	}
	if err := yaml.Unmarshal(b, &direct); err != nil {
		return nil, errors.Wrapf(err, "decode direct log config %s", path)
	}
	if settings.LogLevel == "" {
		settings.LogLevel = direct.Level
	}
	if settings.LogFormat == "" {
		settings.LogFormat = direct.Format
	}
	if direct.Output == logcopter.OutputStdout {
		settings.LogToStdout = true
	}
	if direct.Caller {
		settings.WithCaller = true
	}
	if direct.StrictAreas {
		settings.StrictAreas = true
	}
	settings.Areas = mergeStringMaps(settings.Areas, direct.Areas, normalizeMap(root["areas"]))
	settings.LogAreas = mergeStringMaps(settings.LogAreas, normalizeMap(root["log-area"]))
	return &settings, nil
}

func ParseAreaOverrides(values []string) (map[string]string, error) {
	areas := map[string]string{}
	for _, value := range values {
		for _, entry := range strings.Split(value, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			key, level, ok := strings.Cut(entry, ":")
			if !ok {
				key, level, ok = strings.Cut(entry, "=")
			}
			key = strings.TrimSpace(key)
			level = strings.TrimSpace(level)
			if !ok || key == "" || level == "" {
				return nil, errors.Errorf("invalid log-area override %q; expected area:level or area=level", entry)
			}
			areas[key] = level
		}
	}
	return areas, nil
}

func cloneLoggingSettings(in *LoggingSettings) *LoggingSettings {
	if in == nil {
		return nil
	}
	out := *in
	out.LogConfigFiles = append([]string{}, in.LogConfigFiles...)
	out.LogAreas = mergeStringMaps(in.LogAreas)
	out.Areas = mergeStringMaps(in.Areas)
	return &out
}

func applyLoggingSettings(dst *LoggingSettings, src *LoggingSettings, includeConfigFiles bool) {
	if src == nil {
		return
	}
	if src.WithCaller {
		dst.WithCaller = true
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	if src.LogFormat != "" {
		dst.LogFormat = src.LogFormat
	}
	if src.LogFile != "" {
		dst.LogFile = src.LogFile
	}
	if src.LogToStdout {
		dst.LogToStdout = true
	}
	if includeConfigFiles && len(src.LogConfigFiles) > 0 {
		dst.LogConfigFiles = append(dst.LogConfigFiles, src.LogConfigFiles...)
	}
	if len(src.Areas) > 0 {
		dst.Areas = mergeStringMaps(dst.Areas, src.Areas)
	}
	if len(src.LogAreas) > 0 {
		dst.LogAreas = mergeStringMaps(dst.LogAreas, src.LogAreas)
	}
	if src.StrictAreas {
		dst.StrictAreas = true
	}
}

func mergeStringMaps(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			if strings.TrimSpace(k) == "" {
				continue
			}
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out
}

func normalizeMap(v interface{}) map[string]string {
	out := map[string]string{}
	m, ok := v.(map[string]interface{})
	if !ok {
		return out
	}
	for k, value := range m {
		out[k] = strings.TrimSpace(toString(value))
	}
	return out
}

func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}
