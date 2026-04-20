package cli

import (
	"context"
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

// CobraMiddlewaresFunc returns the full middleware chain for a Cobra command.
// If you set it on CobraParserConfig, it replaces the built-in parser chain entirely.
// That means you must re-add any sources you still want (for example env, config,
// args, flags, or defaults) inside the custom function.
type CobraMiddlewaresFunc func(
	parsedCommandSections *values.Values,
	cmd *cobra.Command,
	args []string,
) ([]cmd_sources.Middleware, error)

// CobraCommandDefaultMiddlewares builds a minimal middleware chain for Cobra commands.
// It reads Cobra flags and positional arguments, then applies defaults.
// Use it when you want an explicit chain; if you also need env loading or config
// file loading, add those middlewares yourself.
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
	// middlewaresFunc builds the parser chain after GlazedCommandSettings has been parsed.
	// At this point, Cobra has already parsed CLI flags and arguments, but the values
	// have not yet been turned into Glazed Values.
	//
	// In the built-in parser path, Glazed wires in env/config loading when configured.
	// If a custom function is supplied through CobraParserConfig, it replaces that
	// built-in chain and must re-add any sources it still needs.
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

type ConfigPlanBuilder func(parsedCommandSections *values.Values, cmd *cobra.Command, args []string) (*glazedConfig.Plan, error)

type CobraParserConfig struct {
	// MiddlewaresFunc replaces the default Cobra parser chain.
	// Leave it nil to keep the built-in parser path, which can add env/config loading when configured.
	MiddlewaresFunc                    CobraMiddlewaresFunc
	ShortHelpSections                  []string
	SkipCommandSettingsSection         bool
	EnableProfileSettingsSection       bool
	EnableCreateCommandSettingsSection bool
	// AppName controls the default env prefix (strings.ToUpper(AppName)) used by the
	// built-in parser path. It enables env loading, but it does not imply config discovery.
	AppName string
	// ConfigPlanBuilder resolves explicit layered config policy for the command.
	// It is only used by the built-in parser path; if nil, no config files are loaded automatically.
	ConfigPlanBuilder ConfigPlanBuilder
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
// fields specified in the Sections CommandDescription to the cobra command.
func NewCobraParserFromSections(
	sections *schema.Schema,
	cfg *CobraParserConfig,
) (*CobraParser, error) {
	// Initialize parser with defaults
	ret := &CobraParser{
		Sections:                           sections,
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
		// Build the built-in env/config-aware chain only when the caller did not supply a custom middleware function.
		if cfg.MiddlewaresFunc == nil {
			cfgCopy := *cfg
			ret.middlewaresFunc = func(parsedCommandSections *values.Values, cmd *cobra.Command, args []string) ([]cmd_sources.Middleware, error) {
				middlewares_ := []cmd_sources.Middleware{}

				// Append in reverse precedence so the last applied has highest precedence (flags)
				middlewares_ = append(middlewares_,
					cmd_sources.FromCobra(cmd,
						fields.WithSource("cobra"),
					),
				)

				middlewares_ = append(middlewares_,
					cmd_sources.FromArgs(args,
						fields.WithSource("arguments"),
					),
				)

				if cfgCopy.AppName != "" {
					envPrefix := strings.ToUpper(cfgCopy.AppName)
					middlewares_ = append(middlewares_,
						cmd_sources.FromEnv(envPrefix,
							fields.WithSource("env"),
						),
					)
				}

				if cfgCopy.ConfigPlanBuilder != nil {
					middlewares_ = append(middlewares_,
						cmd_sources.FromConfigPlanBuilder(
							func(_ context.Context, _ *values.Values) (*glazedConfig.Plan, error) {
								return cfgCopy.ConfigPlanBuilder(parsedCommandSections, cmd, args)
							},
							cmd_sources.WithParseOptions(fields.WithSource("config")),
						),
					)
				}

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
	// NOTE(manuel, 2024-01-03) Maybe add some middleware functionality to whitelist/blacklist the Sections/fields that get added to the CLI
	// If we want to remove some fields from the CLI args (for example some output settings or so)
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
