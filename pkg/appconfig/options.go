package appconfig

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	cmd_sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type parserOptions struct {
	// middlewares are collected in the order options are applied.
	//
	// IMPORTANT: The recommended interpretation is that earlier options are lower
	// precedence and later options are higher precedence (i.e. last wins).
	middlewares []cmd_sources.Middleware

	// Bookkeeping for advanced options that need to derive additional behavior
	// (for example, profiles require a bootstrap parse of profile selection).
	//
	// These fields are internal API and may be extended over time.
	envPrefixes []string
	configFiles []string
	cobraCmd    *cobra.Command
	cobraArgs   []string
}

// ParserOption configures a Parser.
type ParserOption func(*parserOptions) error

// WithDefaults appends the defaults middleware (lowest precedence in the typical chain).
func WithDefaults() ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares,
			cmd_sources.FromDefaults(
				fields.WithSource(fields.SourceDefaults),
			),
		)
		return nil
	}
}

// WithEnv enables parsing from environment variables using the given prefix.
func WithEnv(prefix string) ParserOption {
	return func(o *parserOptions) error {
		if prefix == "" {
			return errors.New("env prefix must not be empty")
		}
		o.envPrefixes = append(o.envPrefixes, prefix)
		o.middlewares = append(o.middlewares,
			cmd_sources.FromEnv(
				prefix,
				fields.WithSource("env"),
			),
		)
		return nil
	}
}

// WithConfigFiles configures config files to load (low -> high precedence).
func WithConfigFiles(files ...string) ParserOption {
	return func(o *parserOptions) error {
		o.configFiles = append(o.configFiles, files...)
		o.middlewares = append(o.middlewares,
			cmd_sources.FromFiles(
				files,
				cmd_sources.WithParseOptions(fields.WithSource("config")),
			),
		)
		return nil
	}
}

// WithValuesForLayers configures programmatic values for layers (optional).
func WithValuesForLayers(values map[string]map[string]interface{}) ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares,
			cmd_sources.FromMap(
				values,
				fields.WithSource("provided-values"),
			),
		)
		return nil
	}
}

// WithMiddlewares injects additional middlewares into the parse chain.
//
// NOTE: Middleware ordering is subtle; this is an escape hatch for advanced usage.
func WithMiddlewares(middlewares ...cmd_sources.Middleware) ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares, middlewares...)
		return nil
	}
}

// WithCobra configures the Parser to read flags and positional arguments from a Cobra command.
//
// The caller is responsible for ensuring Cobra has parsed the args (i.e. this is
// used from within a cobra Run/RunE/PreRun hook, or after Execute has parsed).
func WithCobra(cmd *cobra.Command, args []string) ParserOption {
	return func(o *parserOptions) error {
		if cmd == nil {
			return errors.New("cobra command must not be nil")
		}
		o.cobraCmd = cmd
		o.cobraArgs = append([]string(nil), args...)
		// GatherArguments is lower precedence than ParseFromCobraCommand (flags).
		o.middlewares = append(o.middlewares,
			cmd_sources.FromArgs(
				append([]string(nil), args...),
				fields.WithSource("arguments"),
			),
			cmd_sources.FromCobra(
				cmd,
				fields.WithSource("cobra"),
			),
		)
		return nil
	}
}

type profileOption struct {
	envPrefix          string
	defaultProfile     string
	defaultProfileFile string
	profileFile        string
}

// ProfileOption configures WithProfile.
type ProfileOption func(*profileOption)

// WithProfileEnvPrefix overrides the env prefix used when bootstrap-parsing `profile-settings`.
//
// If not provided, WithProfile defaults to strings.ToUpper(appName).
func WithProfileEnvPrefix(prefix string) ProfileOption {
	return func(o *profileOption) {
		o.envPrefix = prefix
	}
}

// WithProfileDefaultName overrides the default profile name ("default").
func WithProfileDefaultName(profile string) ProfileOption {
	return func(o *profileOption) {
		o.defaultProfile = profile
	}
}

// WithProfileDefaultFile overrides the computed "well-known default" profile file location
// used for error semantics (default: os.UserConfigDir() + "/<appName>/profiles.yaml").
func WithProfileDefaultFile(path string) ProfileOption {
	return func(o *profileOption) {
		o.defaultProfileFile = path
	}
}

// WithProfileFile sets a default profile file path to use when `profile-settings.profile-file`
// is not provided by config/env/flags.
//
// Note: This does not change the "well-known default" path used by GatherFlagsFromProfiles'
// error semantics unless also paired with WithProfileDefaultFile.
func WithProfileFile(path string) ProfileOption {
	return func(o *profileOption) {
		o.profileFile = path
	}
}

// WithProfile enables profiles.yaml loading with circularity-safe bootstrap parsing of profile selection.
//
// It does a mini "bootstrap parse" for the `profile-settings` layer to resolve:
// - profile-settings.profile
// - profile-settings.profile-file
//
// using the sources configured on the parser (cobra/env/config) + defaults, then applies the
// selected profile at the correct precedence level (profiles override defaults, but are
// overridden by config/env/flags).
//
// IMPORTANT: Ordering matters. Recommended option order:
//
//	appconfig.WithDefaults(),
//	appconfig.WithProfile(appName),
//	appconfig.WithConfigFiles(...),
//	appconfig.WithEnv("APP"),
//	appconfig.WithCobra(cmd, args),
//
// Note: For Cobra flags like `--profile` / `--profile-file` to be accepted by Cobra, callers
// must ensure those flags exist on the command (typically by adding the ProfileSettings layer
// to the Cobra command elsewhere). WithProfile can still resolve selection from env/config
// without Cobra flags.
func WithProfile(appName string, opts ...ProfileOption) ParserOption {
	return func(o *parserOptions) error {
		if strings.TrimSpace(appName) == "" {
			return errors.New("profile appName must not be empty")
		}

		pcfg := &profileOption{
			defaultProfile: "default",
		}
		for _, opt := range opts {
			if opt != nil {
				opt(pcfg)
			}
		}

		defaultProfileFile := pcfg.defaultProfileFile
		if defaultProfileFile == "" {
			xdgConfigPath, err := os.UserConfigDir()
			if err != nil {
				return errors.Wrap(err, "failed to get user config directory")
			}
			defaultProfileFile = filepath.Join(xdgConfigPath, appName, "profiles.yaml")
		}

		o.middlewares = append(o.middlewares,
			func(next cmd_sources.HandlerFunc) cmd_sources.HandlerFunc {
				return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
					// 1) Bootstrap-parse profile selection.
					psLayer, err := cli.NewProfileSettingsLayer()
					if err != nil {
						return err
					}

					bootstrapLayers := schema.NewSchema(schema.WithSections(psLayer))
					bootstrapParsed := values.New()

					// IMPORTANT: Profile selection env vars should default to the profile "appName",
					// not to any other WithEnv(...) prefix used for the rest of the application's settings.
					// Example: you might parse general settings from PRESCRIBE_*, but still want
					// PINOCCHIO_PROFILE / PINOCCHIO_PROFILE_FILE to select profiles from pinocchio/profiles.yaml.
					envPrefix := pcfg.envPrefix
					if envPrefix == "" {
						envPrefix = strings.ToUpper(appName)
					}

					bootstrapMiddlewares := []cmd_sources.Middleware{}
					if o.cobraCmd != nil {
						bootstrapMiddlewares = append(bootstrapMiddlewares,
							cmd_sources.FromCobra(o.cobraCmd,
								fields.WithSource("cobra"),
							),
						)
					}
					if envPrefix != "" {
						bootstrapMiddlewares = append(bootstrapMiddlewares,
							cmd_sources.FromEnv(envPrefix,
								fields.WithSource("env"),
							),
						)
					}
					if len(o.configFiles) > 0 {
						bootstrapMiddlewares = append(bootstrapMiddlewares,
							cmd_sources.FromFiles(
								o.configFiles,
								cmd_sources.WithParseOptions(fields.WithSource("config")),
							),
						)
					}
					bootstrapMiddlewares = append(bootstrapMiddlewares,
						cmd_sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
					)

					if err := cmd_sources.Execute(bootstrapLayers, bootstrapParsed, bootstrapMiddlewares...); err != nil {
						return errors.Wrap(err, "failed to bootstrap-parse profile-settings")
					}

					ps := &cli.ProfileSettings{}
					if err := bootstrapParsed.DecodeSectionInto(cli.ProfileSettingsSlug, ps); err != nil {
						return errors.Wrap(err, "failed to initialize bootstrap profile settings")
					}

					profileName := ps.Profile
					if profileName == "" {
						profileName = pcfg.defaultProfile
					}
					profileFile := ps.ProfileFile
					if profileFile == "" {
						if pcfg.profileFile != "" {
							profileFile = pcfg.profileFile
						} else {
							profileFile = defaultProfileFile
						}
					}

					// 2) Run lower-precedence chain (typically defaults + provided-values).
					if err := next(layers_, parsedLayers); err != nil {
						return err
					}

					// 3) Apply profiles.yaml at the intended precedence layer.
					mw := cmd_sources.GatherFlagsFromProfiles(
						defaultProfileFile,
						profileFile,
						profileName,
						pcfg.defaultProfile,
						fields.WithSource("profiles"),
						fields.WithMetadata(map[string]interface{}{
							"profileFile": profileFile,
							"profile":     profileName,
						}),
					)
					handler := mw(func(_ *schema.Schema, _ *values.Values) error { return nil })
					return handler(layers_, parsedLayers)
				}
			},
		)

		return nil
	}
}
