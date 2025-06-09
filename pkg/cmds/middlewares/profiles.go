package middlewares

import (
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func GatherFlagsFromProfiles(
	defaultProfileFile string,
	profileFile string,
	profile string,
	options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			// if the file does not exist and is not the defaultProfileFile, fail
			_, err = os.Stat(profileFile)
			if os.IsNotExist(err) {
				if profileFile != defaultProfileFile {
					return errors.Errorf("profile file %s does not exist", profileFile)
				}
				return nil
			}

			// parse profileFile as yaml
			f, err := os.Open(profileFile)
			if err != nil {
				return err
			}
			defer func(f *os.File) {
				_ = f.Close()
			}(f)

			// profile1:
			//   layer1:
			//     parameterName: parameterValue
			//   layer2:
			//     parameterName: parameterValue
			// etc...
			v := map[string]map[string]map[string]interface{}{}
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(&v)
			if err != nil {
				return err
			}

			if profileMap, ok := v[profile]; ok {
				return updateFromMap(layers_, parsedLayers, profileMap, options...)
			} else {
				if profile != "default" {
					return errors.Errorf("profile %s not found in %s", profile, profileFile)
				}

				return nil
			}
		}
	}
}

// GatherFlagsFromCustomProfiles creates a middleware that loads profile configuration from custom sources.
// This middleware allows loading profiles from:
// 1. A specific profile file path
// 2. Another app's default profile locations
//
// This is useful when you want to load profile configuration from a different source than the default
// application profiles.
//
// Usage:
//
//	// Load from specific profile file
//	middleware := middlewares.GatherFlagsFromCustomProfiles(
//	    "production",
//	    middlewares.WithProfileFile("/path/to/custom-profiles.yaml"),
//	    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
//	)
//
//	// Load from another app's profiles
//	middleware := middlewares.GatherFlagsFromCustomProfiles(
//	    "shared-profile",
//	    middlewares.WithProfileAppName("other-app"),
//	    middlewares.WithProfileParseOptions(parameters.WithParseStepSource("other-app-profiles")),
//	)
func GatherFlagsFromCustomProfiles(profileName string, options ...ProfileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			config := &ProfileConfig{
				ProfileName: profileName,
			}
			for _, opt := range options {
				opt(config)
			}

			// Determine profile file path
			profileFile, err := resolveProfileFilePath(config)
			if err != nil {
				return err
			}

			// Check if file exists
			_, err = os.Stat(profileFile)
			if os.IsNotExist(err) {
				if config.Required {
					return errors.Errorf("profile file %s does not exist", profileFile)
				}
				return nil
			}

			// Load and parse profile file
			profileMap, err := loadProfileFromFile(profileFile, profileName)
			if err != nil {
				return err
			}

			if profileMap == nil {
				if profileName != "default" && config.Required {
					return errors.Errorf("profile %s not found in %s", profileName, profileFile)
				}
				return nil
			}

			return updateFromMap(layers_, parsedLayers, profileMap, config.ParseOptions...)
		}
	}
}

// ProfileConfig holds configuration for the custom profile middleware
type ProfileConfig struct {
	ProfileName   string
	ProfileFile   string
	AppName       string
	Required      bool
	ParseOptions  []parameters.ParseStepOption
}

// ProfileOption is a function that configures ProfileConfig
type ProfileOption func(*ProfileConfig)

// WithProfileFile sets a specific profile file path to load
func WithProfileFile(profileFile string) ProfileOption {
	return func(c *ProfileConfig) {
		c.ProfileFile = profileFile
	}
}

// WithProfileAppName sets the app name to use for loading profiles from standard locations
func WithProfileAppName(appName string) ProfileOption {
	return func(c *ProfileConfig) {
		c.AppName = appName
	}
}

// WithProfileRequired sets whether the profile file and profile must exist
func WithProfileRequired(required bool) ProfileOption {
	return func(c *ProfileConfig) {
		c.Required = required
	}
}

// WithProfileParseOptions adds parse step options to the middleware
func WithProfileParseOptions(options ...parameters.ParseStepOption) ProfileOption {
	return func(c *ProfileConfig) {
		c.ParseOptions = append(c.ParseOptions, options...)
	}
}

// resolveProfileFilePath determines the profile file path based on the config
func resolveProfileFilePath(config *ProfileConfig) (string, error) {
	if config.ProfileFile != "" {
		// Use explicit file path
		return config.ProfileFile, nil
	} else if config.AppName != "" {
		// Use app's default profile locations
		xdgConfigPath, err := os.UserConfigDir()
		if err != nil {
			return "", errors.Wrapf(err, "failed to get user config directory")
		}
		return filepath.Join(xdgConfigPath, config.AppName, "profiles.yaml"), nil
	} else {
		// No profile source specified
		return "", errors.New("either ProfileFile or AppName must be specified")
	}
}

// loadProfileFromFile loads a specific profile from a YAML file
func loadProfileFromFile(profileFile, profileName string) (map[string]map[string]interface{}, error) {
	f, err := os.Open(profileFile)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Profile file structure:
	// profile1:
	//   layer1:
	//     parameterName: parameterValue
	//   layer2:
	//     parameterName: parameterValue
	// profile2:
	//   ...
	v := map[string]map[string]map[string]interface{}{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&v)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse profile file %s", profileFile)
	}

	if profileMap, ok := v[profileName]; ok {
		return profileMap, nil
	}

	return nil, nil // Profile not found
}
