package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/viper"
)

// GatherFlagsFromCustomViper creates a middleware that loads parameter values from a custom Viper instance.
// This middleware allows loading configuration from:
// 1. A specific config file path
// 2. Another app's default config locations
//
// This is useful when you want to load configuration from a different source than the default
// application configuration.
//
// Usage:
//
//	// Load from specific config file
//	middleware := middlewares.GatherFlagsFromCustomViper(
//	    middlewares.WithConfigFile("/path/to/config.yaml"),
//	    parameters.WithParseStepSource("custom-config"),
//	)
//
//	// Load from another app's config
//	middleware := middlewares.GatherFlagsFromCustomViper(
//	    middlewares.WithAppName("other-app"),
//	    parameters.WithParseStepSource("other-app-config"),
//	)
func GatherFlagsFromCustomViper(options ...CustomViperOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			config := &CustomViperConfig{}
			for _, opt := range options {
				opt(config)
			}

			// Create custom viper instance
			customViper, err := createCustomViperInstance(config)
			if err != nil {
				return err
			}

			// Store the original viper instance
			originalViper := viper.GetViper()

			// Temporarily replace global viper
			viper.Reset()
			for key, value := range customViper.AllSettings() {
				viper.Set(key, value)
			}

			// Process layers with custom viper
			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				parseOptions := append([]parameters.ParseStepOption{
					parameters.WithParseStepSource("custom-viper"),
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer":      l.GetName(),
						"configFile": config.ConfigFile,
						"appName":    config.AppName,
					}),
				}, config.ParseOptions...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetParameterDefinitions()
				prefix := l.GetPrefix()

				ps, err := parameterDefinitions.GatherFlagsFromViper(true, prefix, parseOptions...)
				if err != nil {
					return err
				}

				_, err = parsedLayer.Parameters.Merge(ps)
				if err != nil {
					return err
				}

				return nil
			})

			// Restore original viper
			viper.Reset()
			for key, value := range originalViper.AllSettings() {
				viper.Set(key, value)
			}

			return err
		}
	}
}

// CustomViperConfig holds configuration for the custom viper middleware
type CustomViperConfig struct {
	ConfigFile   string
	AppName      string
	ParseOptions []parameters.ParseStepOption
}

// CustomViperOption is a function that configures CustomViperConfig
type CustomViperOption func(*CustomViperConfig)

// WithConfigFile sets a specific config file path to load
func WithConfigFile(configFile string) CustomViperOption {
	return func(c *CustomViperConfig) {
		c.ConfigFile = configFile
	}
}

// WithAppName sets the app name to use for loading config from standard locations
func WithAppName(appName string) CustomViperOption {
	return func(c *CustomViperConfig) {
		c.AppName = appName
	}
}

// WithParseOptions adds parse step options to the middleware
func WithParseOptions(options ...parameters.ParseStepOption) CustomViperOption {
	return func(c *CustomViperConfig) {
		c.ParseOptions = append(c.ParseOptions, options...)
	}
}

// createCustomViperInstance creates a new viper instance based on the config
func createCustomViperInstance(config *CustomViperConfig) (*viper.Viper, error) {
	if config.ConfigFile != "" {
		// Load from specific config file
		return logging.InitViperInstanceWithAppName("", config.ConfigFile)
	} else if config.AppName != "" {
		// Load from app's default config locations
		return logging.InitViperInstanceWithAppName(config.AppName, "")
	} else {
		// Return empty viper instance if no config specified
		v := viper.New()
		return v, nil
	}
}
