package cli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	cmd_sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedConfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
)

// CobraMiddlewaresFunc is a function that returns a list of middlewares for a cobra command.
// It can be used to overload the default middlewares for cobra commands.
// It is mostly used to add a "load from json" layer set in the GlazedCommandSettings.
type CobraMiddlewaresFunc func(
	parsedCommandLayers *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]cmd_middlewares.Middleware, error)

// CobraCommandDefaultMiddlewares is the default implementation for creating
// the middlewares used in a Cobra command. It handles parsing parameters
// from Cobra flags, command line arguments, environment variables, and
// default values. The middlewares gather all these parameters into a
// ParsedParameters object.
//
// If the commandSettings specify parameters to be loaded from a file, this gets added as a middleware.
func CobraCommandDefaultMiddlewares(
	parsedCommandLayers *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]cmd_middlewares.Middleware, error) {
	commandSettings := &CommandSettings{}
	err := values.DecodeSectionInto(parsedCommandLayers, CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	// Default chain without legacy per-command file injection
	middlewares_ := []cmd_middlewares.Middleware{
		cmd_sources.FromCobra(cmd,
			cmd_sources.WithSource("cobra"),
		),
		cmd_sources.FromArgs(args,
			cmd_sources.WithSource("arguments"),
		),
	}

	middlewares_ = append(middlewares_,
		cmd_sources.FromDefaults(cmd_sources.WithSource(cmd_sources.SourceDefaults)),
	)

	return middlewares_, nil
}

// CobraParser takes a CommandDescription, and hooks it up to a cobra command.
// It can then be used to parse the cobra flags and arguments back into a
// set of ParsedLayer and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run* method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered Layers
// from the description.
type CobraParser struct {
	Layers *schema.Schema
	// middlewaresFunc is called after the command has been executed, once the
	// GlazedCommandSettings struct has been filled. At this point, cobra has done the parsing
	// of CLI flags and arguments, but these haven't yet been parsed into ParsedLayers
	// by the glazed framework.
	//
	// This hooks allows the implementor to specify additional ways of loading parameters
	// (for example, sqleton loads the dbt and sql connection parameters from env as well).
	middlewaresFunc CobraMiddlewaresFunc
	// List of layers to be shown in short help, empty: always show all
	shortHelpLayers []string
	// skipCommandSettingsLayer controls whether the CommandSettingsLayer should be automatically added
	skipCommandSettingsLayer bool
	// enableProfileSettingsLayer controls whether the ProfileSettingsLayer should be added
	enableProfileSettingsLayer bool
	// enableCreateCommandSettingsLayer controls whether the CreateCommandSettingsLayer should be added
	enableCreateCommandSettingsLayer bool
}

// Inserted: new config struct for parser customization
type CobraParserConfig struct {
	MiddlewaresFunc                  CobraMiddlewaresFunc
	ShortHelpLayers                  []string
	SkipCommandSettingsLayer         bool
	EnableProfileSettingsLayer       bool
	EnableCreateCommandSettingsLayer bool
	// New: application name and optional explicit config path
	AppName    string
	ConfigPath string
	// New: optional callback returning an ordered list of config files (low -> high precedence)
	ConfigFilesFunc func(parsedCommandLayers *values.Values, cmd *cobra.Command, args []string) ([]string, error)
}

func NewCobraCommandFromCommandDescription(
	description *cmds.CommandDescription,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}
	return cmd
}

// NewCobraParserFromLayers creates a new CobraParser instance from a
// CommandDescription, initializes the underlying cobra.Command, and adds all the
// parameters specified in the Layers CommandDescription to the cobra command.
func NewCobraParserFromLayers(
	paramLayers *schema.Schema,
	cfg *CobraParserConfig,
) (*CobraParser, error) {
	// Initialize parser with defaults
	ret := &CobraParser{
		Layers:                           paramLayers,
		middlewaresFunc:                  CobraCommandDefaultMiddlewares,
		shortHelpLayers:                  []string{},
		skipCommandSettingsLayer:         false,
		enableProfileSettingsLayer:       false,
		enableCreateCommandSettingsLayer: false,
	}
	// Apply provided config if any
	if cfg != nil {
		if cfg.MiddlewaresFunc != nil {
			ret.middlewaresFunc = cfg.MiddlewaresFunc
		}
		ret.shortHelpLayers = cfg.ShortHelpLayers
		ret.skipCommandSettingsLayer = cfg.SkipCommandSettingsLayer
		ret.enableProfileSettingsLayer = cfg.EnableProfileSettingsLayer
		ret.enableCreateCommandSettingsLayer = cfg.EnableCreateCommandSettingsLayer
		// If no custom middlewares func provided, construct one that leverages AppName/ConfigPath
		if cfg.MiddlewaresFunc == nil {
			cfgCopy := *cfg
			ret.middlewaresFunc = func(parsedCommandLayers *values.Values, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error) {
				middlewares_ := []cmd_middlewares.Middleware{}

				// Append in reverse precedence so the last applied has highest precedence (flags)
				// Flags (highest precedence)
				middlewares_ = append(middlewares_,
					cmd_sources.FromCobra(cmd,
						cmd_sources.WithSource("cobra"),
					),
				)

				// Positional arguments
				middlewares_ = append(middlewares_,
					cmd_sources.FromArgs(args,
						cmd_sources.WithSource("arguments"),
					),
				)

				// Environment overrides
				if cfgCopy.AppName != "" {
					envPrefix := strings.ToUpper(cfgCopy.AppName)
					middlewares_ = append(middlewares_,
						cmd_sources.FromEnv(envPrefix,
							cmd_sources.WithSource("env"),
						),
					)
				}

				// Config files (low -> high precedence) via a single middleware
				resolver := cfgCopy.ConfigFilesFunc
				if resolver == nil {
					resolver = func(parsed *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
						var files []string
						if cfgCopy.ConfigPath != "" || cfgCopy.AppName != "" {
							explicit := cfgCopy.ConfigPath
							cs := &CommandSettings{}
							if err := values.DecodeSectionInto(parsed, CommandSettingsSlug, cs); err == nil {
								if cs.ConfigFile != "" {
									explicit = cs.ConfigFile
								}
							}
							p, _ := glazedConfig.ResolveAppConfigPath(cfgCopy.AppName, explicit)
							if p != "" {
								files = []string{p}
							}
						}
						return files, nil
					}
				}
				// Wrap resolver to bind parsedCommandLayers captured earlier
				wrapped := func(_ *values.Values, cmd_ *cobra.Command, args_ []string) ([]string, error) {
					return resolver(parsedCommandLayers, cmd_, args_)
				}
				middlewares_ = append(middlewares_,
					cmd_middlewares.LoadParametersFromResolvedFilesForCobra(
						cmd,
						args,
						wrapped,
						cmd_sources.WithSource("config"),
					),
				)

				// Defaults (lowest precedence)
				middlewares_ = append(middlewares_,
					cmd_sources.FromDefaults(cmd_sources.WithSource(cmd_sources.SourceDefaults)),
				)

				return middlewares_, nil
			}
		}
	}

	// Only add the glazed command layer if not explicitly skipped
	if !ret.skipCommandSettingsLayer {
		commandSettingsLayer, err := NewCommandSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(commandSettingsLayer.GetSlug(), commandSettingsLayer)
	}

	// Only add the profile settings layer if explicitly enabled
	if ret.enableProfileSettingsLayer {
		profileSettingsLayer, err := NewProfileSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(profileSettingsLayer.GetSlug(), profileSettingsLayer)
	}

	// Only add the create command settings layer if explicitly enabled
	if ret.enableCreateCommandSettingsLayer {
		createCommandSettingsLayer, err := NewCreateCommandSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(createCommandSettingsLayer.GetSlug(), createCommandSettingsLayer)
	}

	return ret, nil
}

func (c *CobraParser) AddToCobraCommand(cmd *cobra.Command) error {
	// NOTE(manuel, 2024-01-03) Maybe add some middleware functionality to whitelist/blacklist the Layers/parameters that get added to the CLI
	// If we want to remove some parameters from the CLI args (for example some output settings or so)
	err := c.Layers.ForEachE(func(_ string, layer schema.Section) error {
		// check that layer is a CobraParameterLayer
		// if not, return an error
		cobraLayer, ok := layer.(layers.CobraParameterLayer)
		if !ok {
			log.Error().Str("layer", layer.GetName()).Msg("Layer is not a CobraParameterLayer")
			return errors.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		err := cobraLayer.AddLayerToCobraCommand(cmd)
		if err != nil {
			log.Error().Err(err).Str("layer", layer.GetName()).Msg("Could not add layer to cobra command")
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(c.shortHelpLayers) > 0 {
		shortHelperLayer := strings.Join(c.shortHelpLayers, ",")
		cmd.Annotations["shortHelpLayers"] = shortHelperLayer
	}

	return nil
}

func (c *CobraParser) Parse(
	cmd *cobra.Command,
	args []string,
) (*values.Values, error) {
	// We parse the glazed command settings first, since they will influence the following parsing
	// steps.
	parsedCommandLayers, err := ParseCommandSettingsLayer(cmd)
	if err != nil {
		return nil, err
	}

	// Create the middlewares by invoking the passed in middlewares constructor.
	// This is where applications can specify their own middlewares.
	middlewares_, err := c.middlewaresFunc(parsedCommandLayers, cmd, args)
	if err != nil {
		return nil, err
	}

	parsedLayers := values.New()
	err = cmd_sources.Execute(c.Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}

// ParseGlazedCommandLayer parses the global glazed settings from the given cobra.Command, if not nil,
// and from the command itself if present.
func ParseCommandSettingsLayer(cmd *cobra.Command) (*values.Values, error) {
	parsedLayers := values.New()
	commandSettingsLayer, err := NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	profileSettingsLayer, err := NewProfileSettingsLayer()
	if err != nil {
		return nil, err
	}

	createCommandSettingsLayer, err := NewCreateCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	commandSettingsLayers := schema.NewSchema(schema.WithSections(
		commandSettingsLayer,
		profileSettingsLayer,
		createCommandSettingsLayer,
	))

	// Parse the glazed command settings from the cobra command
	middlewares_ := []cmd_middlewares.Middleware{}

	if cmd != nil {
		middlewares_ = append(middlewares_, cmd_sources.FromCobra(cmd, cmd_sources.WithSource("cobra")))
	}

	err = cmd_sources.Execute(commandSettingsLayers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil

}
