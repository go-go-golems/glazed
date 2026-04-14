package logging

type LogConfig struct {
	WithCaller bool
	Level      string
	LogFormat  string
	LogFile    string
}

func InitLoggerWithConfig(config *LogConfig) error {
	settings := &LoggingSettings{
		WithCaller: config.WithCaller,
		LogLevel:   config.Level,
		LogFormat:  config.LogFormat,
		LogFile:    config.LogFile,
	}
	return InitLoggerFromSettings(settings)
}
