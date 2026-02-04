package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

// Deprecated: Use Glazed config middlewares (LoadFieldsFromFiles + UpdateFromEnv) and InitGlazed/InitLoggerFromCobra.
func InitViperWithAppName(appName string, configFile string) error {
	log.Warn().Msg("logging.InitViperWithAppName is deprecated; use Glazed config middlewares and InitLoggerFromCobra")
	viper.SetEnvPrefix(appName)

	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("yaml")
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", appName))

		xdgConfigPath, err := os.UserConfigDir()
		if err == nil {
			viper.AddConfigPath(fmt.Sprintf("%s/%s", xdgConfigPath, appName))
		}
	}

	// Read the configuration file into Viper
	err := viper.ReadInConfig()
	// if the file does not exist, continue normally
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		return err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return nil
}

// Deprecated: Use InitGlazed(appName, rootCmd) and InitLoggerFromCobra in PersistentPreRunE.
func InitViper(appName string, rootCmd *cobra.Command) error {
	log.Warn().Msg("logging.InitViper is deprecated; use InitGlazed and InitLoggerFromCobra")
	rootCmd.PersistentFlags().String("config", "",
		fmt.Sprintf("Path to config file (default ~/.%s/config.yml)", appName))

	// parse the flags one time just to catch --config
	configFile := ""
	for idx, arg := range os.Args {
		if arg == "--config" {
			if len(os.Args) > idx+1 {
				configFile = os.Args[idx+1]
			}
		}
	}

	err := InitViperWithAppName(appName, configFile)
	if err != nil {
		return err
	}

	// Bind the variables to the command-line flags
	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		return err
	}

	return nil
}

// Deprecated: Avoid Viper instances for config parsing; use config middlewares.
func InitViperInstanceWithAppName(appName string, configFile string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix(appName)

	if configFile != "" {
		v.SetConfigFile(configFile)
		v.SetConfigType("yaml")
	} else {
		v.SetConfigType("yaml")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", appName))

		xdgConfigPath, err := os.UserConfigDir()
		if err == nil {
			v.AddConfigPath(fmt.Sprintf("%s/%s", xdgConfigPath, appName))
		}
	}

	// Read the configuration file into Viper
	err := v.ReadInConfig()
	// if the file does not exist, continue normally
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		return nil, err
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	return v, nil
}
