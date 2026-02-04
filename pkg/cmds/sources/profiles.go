package sources

import (
	fields "github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func GatherFlagsFromProfiles(
	defaultProfileFile string,
	profileFile string,
	profile string,
	defaultProfileName string,
	options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if defaultProfileName == "" {
				defaultProfileName = "default"
			}

			// If the file does not exist:
			// - If the user explicitly set a non-default file path, fail.
			// - If this is the default profile file and a non-default profile is requested, fail
			//   (otherwise we silently skip profile loading, which is confusing and breaks
			//   expectations like PINOCCHIO_PROFILE=foobar should error).
			// - If this is the default profile file and the default profile is requested, skip.
			_, err = os.Stat(profileFile)
			if os.IsNotExist(err) {
				if profileFile != defaultProfileFile {
					return errors.Errorf("profile file %s does not exist", profileFile)
				}
				if profile != defaultProfileName {
					return errors.Errorf("profile file %s does not exist (requested profile %s)", profileFile, profile)
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
			//   section1:
			//     fieldName: fieldValue
			//   section2:
			//     fieldName: fieldValue
			// etc...
			v := map[string]map[string]map[string]interface{}{}
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(&v)
			if err != nil {
				return err
			}

			if profileMap, ok := v[profile]; ok {
				return updateFromMap(schema_, parsedValues, profileMap, options...)
			} else {
				if profile != defaultProfileName {
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
//	    middlewares.WithProfileParseOptions(fields.WithSource("custom-profiles")),
//	)
//
//	// Load from another app's profiles
//	middleware := middlewares.GatherFlagsFromCustomProfiles(
//	    "shared-profile",
//	    middlewares.WithProfileAppName("other-app"),
//	    middlewares.WithProfileParseOptions(fields.WithSource("other-app-profiles")),
//	)
func GatherFlagsFromCustomProfiles(profileName string, options ...ProfileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
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

			return updateFromMap(schema_, parsedValues, profileMap, config.ParseOptions...)
		}
	}
}

// ProfileConfig holds configuration for the custom profile middleware
type ProfileConfig struct {
	ProfileName  string
	ProfileFile  string
	AppName      string
	Required     bool
	ParseOptions []fields.ParseOption
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
func WithProfileParseOptions(options ...fields.ParseOption) ProfileOption {
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
	//   section1:
	//     fieldName: fieldValue
	//   section2:
	//     fieldName: fieldValue
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
