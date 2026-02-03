package cli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	cmd_sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedConfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
)

// CobraMiddlewaresFunc is a function that returns a list of middlewares for a cobra command.
// It can be used to overload the default middlewares for cobra commands.
// It is mostly used to add a "load from json" section set in the GlazedCommandSettings.
type CobraMiddlewaresFunc func(
	parsedCommandSections *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]cmd_sources.Middleware, error)

// CobraCommandDefaultMiddlewares is the default implementation for creating
// the middlewares used in a Cobra command. It handles parsing parameters
// from Cobra flags, command line arguments, environment variables, and
// default values. The middlewares gather all these parameters into a
// FieldValues object.
//
// If the commandSettings specify parameters to be loaded from a file, this gets added as a middleware.
func CobraCommandDefaultMiddlewares(
	parsedCommandSections *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]cmd_sources.Middleware, error) {
	commandSettings := &CommandSettings{}
	err := parsedCommandSections.DecodeSectionInto(CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	// Default chain without legacy per-command file injection
	middlewares_ := []cmd_sources.Middleware{
		cmd_sources.FromCobra(cmd,
			fields.WithSource("cobra"),
		),
		cmd_sources.FromArgs(args,
			fields.WithSource("arguments"),
		),
	}

	middlewares_ = append(middlewares_,
		cmd_sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)

	return middlewares_, nil
}

// CobraParser takes a CommandDescription, and hooks it up to a cobra command.
// It can then be used to parse the cobra flags and arguments back into a
// set of SectionValues and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run* method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered Sections
// from the description.
type CobraParser struct {
	Sections *schema.Schema
	// middlewaresFunc is called after the command has been executed, once the
	// GlazedCommandSettings struct has been filled. At this point, cobra has done the parsing
	// of CLI flags and arguments, but these haven't yet been parsed into Values
	// by the glazed framework.
	//
	// This hooks allows the implementor to specify additional ways of loading parameters
	// (for example, sqleton loads the dbt and sql connection parameters from env as well).
	middlewaresFunc CobraMiddlewaresFunc
	// List of sections to be shown in short help, empty: always show all
	shortHelpSections []string
	// skipCommandSettingsSection controls whether the CommandSettingsSection should be automatically added
	skipCommandSettingsSection bool
	// enableProfileSettingsSection controls whether the ProfileSettingsSection should be added
	enableProfileSettingsSection bool
	// enableCreateCommandSettingsSection controls whether the CreateCommandSettingsSection should be added
	enableCreateCommandSettingsSection bool
}

// Inserted: new config struct for parser customization
type CobraParserConfig struct {
	MiddlewaresFunc                    CobraMiddlewaresFunc
	ShortHelpSections                  []string
	SkipCommandSettingsSection         bool
	EnableProfileSettingsSection       bool
	EnableCreateCommandSettingsSection bool
	// New: application name and optional explicit config path
	AppName    string
	ConfigPath string
	// New: optional callback returning an ordered list of config files (low -> high precedence)
	ConfigFilesFunc func(parsedCommandSections *values.Values, cmd *cobra.Command, args []string) ([]string, error)
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

// NewCobraParserFromSections creates a new CobraParser instance from a
// CommandDescription, initializes the underlying cobra.Command, and adds all the
// parameters specified in the Sections CommandDescription to the cobra command.
func NewCobraParserFromSections(
	paramSections *schema.Schema,
	cfg *CobraParserConfig,
) (*CobraParser, error) {
	// Initialize parser with defaults
	ret := &CobraParser{
		Sections:                           paramSections,
		middlewaresFunc:                    CobraCommandDefaultMiddlewares,
		shortHelpSections:                  []string{},
		skipCommandSettingsSection:         false,
		enableProfileSettingsSection:       false,
		enableCreateCommandSettingsSection: false,
	}
	// Apply provided config if any
	if cfg != nil {
		if cfg.MiddlewaresFunc != nil {
			ret.middlewaresFunc = cfg.MiddlewaresFunc
		}
		ret.shortHelpSections = cfg.ShortHelpSections
		ret.skipCommandSettingsSection = cfg.SkipCommandSettingsSection
		ret.enableProfileSettingsSection = cfg.EnableProfileSettingsSection
		ret.enableCreateCommandSettingsSection = cfg.EnableCreateCommandSettingsSection
		// If no custom middlewares func provided, construct one that leverages AppName/ConfigPath
		if cfg.MiddlewaresFunc == nil {
			cfgCopy := *cfg
			ret.middlewaresFunc = func(parsedCommandSections *values.Values, cmd *cobra.Command, args []string) ([]cmd_sources.Middleware, error) {
				middlewares_ := []cmd_sources.Middleware{}

				// Append in reverse precedence so the last applied has highest precedence (flags)
				// Flags (highest precedence)
				middlewares_ = append(middlewares_,
					cmd_sources.FromCobra(cmd,
						fields.WithSource("cobra"),
					),
				)

				// Positional arguments
				middlewares_ = append(middlewares_,
					cmd_sources.FromArgs(args,
						fields.WithSource("arguments"),
					),
				)

				// Environment overrides
				if cfgCopy.AppName != "" {
					envPrefix := strings.ToUpper(cfgCopy.AppName)
					middlewares_ = append(middlewares_,
						cmd_sources.FromEnv(envPrefix,
							fields.WithSource("env"),
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
							if err := parsed.DecodeSectionInto(CommandSettingsSlug, cs); err == nil {
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
				// Wrap resolver to bind parsedCommandSections captured earlier
				wrapped := func(_ *values.Values, cmd_ *cobra.Command, args_ []string) ([]string, error) {
					return resolver(parsedCommandSections, cmd_, args_)
				}
				middlewares_ = append(middlewares_,
					cmd_sources.LoadParametersFromResolvedFilesForCobra(
						cmd,
						args,
						wrapped,
						fields.WithSource("config"),
					),
				)

				// Defaults (lowest precedence)
				middlewares_ = append(middlewares_,
					cmd_sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
				)

				return middlewares_, nil
			}
		}
	}

	// Only add the glazed command section if not explicitly skipped
	if !ret.skipCommandSettingsSection {
		commandSettingsSection, err := NewCommandSettingsSection()
		if err != nil {
			return nil, err
		}
		ret.Sections.Set(commandSettingsSection.GetSlug(), commandSettingsSection)
	}

	// Only add the profile settings section if explicitly enabled
	if ret.enableProfileSettingsSection {
		profileSettingsSection, err := NewProfileSettingsSection()
		if err != nil {
			return nil, err
		}
		ret.Sections.Set(profileSettingsSection.GetSlug(), profileSettingsSection)
	}

	// Only add the create command settings section if explicitly enabled
	if ret.enableCreateCommandSettingsSection {
		createCommandSettingsSection, err := NewCreateCommandSettingsSection()
		if err != nil {
			return nil, err
		}
		ret.Sections.Set(createCommandSettingsSection.GetSlug(), createCommandSettingsSection)
	}

	return ret, nil
}

func (c *CobraParser) AddToCobraCommand(cmd *cobra.Command) error {
	// NOTE(manuel, 2024-01-03) Maybe add some middleware functionality to whitelist/blacklist the Sections/parameters that get added to the CLI
	// If we want to remove some parameters from the CLI args (for example some output settings or so)
	err := c.Sections.ForEachE(func(_ string, section schema.Section) error {
		// check that section is a CobraSection
		// if not, return an error
		cobraSection, ok := section.(schema.CobraSection)
		if !ok {
			log.Error().Str("section", section.GetName()).Msg("Section is not a CobraSection")
			return errors.Errorf("section %s is not a CobraSection", section.GetName())
		}

		err := cobraSection.AddSectionToCobraCommand(cmd)
		if err != nil {
			log.Error().Err(err).Str("section", section.GetName()).Msg("Could not add section to cobra command")
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(c.shortHelpSections) > 0 {
		shortHelperSection := strings.Join(c.shortHelpSections, ",")
		cmd.Annotations["shortHelpSections"] = shortHelperSection
	}

	return nil
}

func (c *CobraParser) Parse(
	cmd *cobra.Command,
	args []string,
) (*values.Values, error) {
	// We parse the glazed command settings first, since they will influence the following parsing
	// steps.
	parsedCommandSections, err := ParseCommandSettingsSection(cmd)
	if err != nil {
		return nil, err
	}

	// Create the middlewares by invoking the passed in middlewares constructor.
	// This is where applications can specify their own middlewares.
	middlewares_, err := c.middlewaresFunc(parsedCommandSections, cmd, args)
	if err != nil {
		return nil, err
	}

	parsedSections := values.New()
	err = cmd_sources.Execute(c.Sections, parsedSections, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedSections, nil
}

// ParseGlazedCommandSection parses the global glazed settings from the given cobra.Command, if not nil,
// and from the command itself if present.
func ParseCommandSettingsSection(cmd *cobra.Command) (*values.Values, error) {
	parsedSections := values.New()
	commandSettingsSection, err := NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	profileSettingsSection, err := NewProfileSettingsSection()
	if err != nil {
		return nil, err
	}

	createCommandSettingsSection, err := NewCreateCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	commandSettingsSections := schema.NewSchema(schema.WithSections(
		commandSettingsSection,
		profileSettingsSection,
		createCommandSettingsSection,
	))

	// Parse the glazed command settings from the cobra command
	middlewares_ := []cmd_sources.Middleware{}

	if cmd != nil {
		middlewares_ = append(middlewares_, cmd_sources.FromCobra(cmd, fields.WithSource("cobra")))
	}

	err = cmd_sources.Execute(commandSettingsSections, parsedSections, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedSections, nil

}
