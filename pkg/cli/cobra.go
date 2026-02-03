package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cli/cliopatra"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"

	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type CobraRunFunc func(ctx context.Context, parsedValues *values.Values) error

func GetVerbsFromCobraCommand(cmd *cobra.Command) []string {
	var verbs []string
	for cmd != nil {
		verbs = append(verbs, cmd.Name())
		cmd = cmd.Parent()
	}

	list.Reverse(verbs)

	return verbs
}

// runCobraCommand executes the common run flow for all Cobra commands.
func runCobraCommand(
	cmd *cobra.Command,
	s cmds.Command,
	runFunc CobraRunFunc,
	parser *CobraParser,
	cfg *commandBuildConfig,
) {
	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Parse sections into values
		parsedValues, err := parser.Parse(cmd, args)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			_ = cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Minimal command settings: debug flags
		commandSettings := &CommandSettings{}
		if commandSettingsValues, ok := parsedValues.Get(CommandSettingsSlug); ok {
			var printYAML, shouldPrintParsedFields, printSchema bool
			err = commandSettingsValues.DecodeInto(commandSettings)
			cobra.CheckErr(err)
			printYAML = commandSettings.PrintYAML
			shouldPrintParsedFields = commandSettings.PrintParsedFields
			printSchema = commandSettings.PrintSchema

			if shouldPrintParsedFields {
				printParsedFields(parsedValues)
				return
			}
			if printYAML {
				err = s.ToYAML(os.Stdout)
				cobra.CheckErr(err)
				return
			}
			if printSchema {
				schema, err := s.Description().ToJsonSchema()
				cobra.CheckErr(err)
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				err = encoder.Encode(schema)
				cobra.CheckErr(err)
				return
			}
		}

		// Create command settings: cliopatra, alias, create
		if createSectionValues, ok := parsedValues.Get(CreateCommandSettingsSlug); ok {
			createSettings := &CreateCommandSettings{}
			err = createSectionValues.DecodeInto(createSettings)
			cobra.CheckErr(err)

			if createSettings.CreateCliopatra != "" {
				verbs := GetVerbsFromCobraCommand(cmd)
				if len(verbs) == 0 {
					cobra.CheckErr(errors.New("could not get verbs from cobra command"))
				}
				p := cliopatra.NewProgramFromCapture(
					s.Description(),
					parsedValues,
					cliopatra.WithVerbs(verbs[1:]...),
					cliopatra.WithName(createSettings.CreateCliopatra),
					cliopatra.WithPath(verbs[0]),
				)
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(p)
				cobra.CheckErr(err)
				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createSettings.CreateAlias != "" {
				alias := &alias.CommandAlias{
					Name:      createSettings.CreateAlias,
					AliasFor:  s.Description().Name,
					Arguments: args,
					Flags:     map[string]string{},
				}
				cmd.Flags().Visit(func(flag *pflag.Flag) {
					if flag.Name != "create-alias" {
						switch flag.Value.Type() {
						case "stringSlice":
							slice, _ := cmd.Flags().GetStringSlice(flag.Name)
							alias.Flags[flag.Name] = strings.Join(slice, ",")
						case "intSlice":
							slice, _ := cmd.Flags().GetIntSlice(flag.Name)
							alias.Flags[flag.Name] = strings.Join(strings2.IntSliceToStringSlice(slice), ",")
						case "floatSlice":
							slice, _ := cmd.Flags().GetFloat64Slice(flag.Name)
							alias.Flags[flag.Name] = strings.Join(strings2.Float64SliceToStringSlice(slice), ",")
						default:
							alias.Flags[flag.Name] = flag.Value.String()
						}
					}
				})
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(alias)
				cobra.CheckErr(err)
				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createSettings.CreateCommand != "" {
				schema_ := s.Description().Schema.Clone()
				cmdDesc := &cmds.CommandDescription{
					Name:   createSettings.CreateCommand,
					Short:  s.Description().Short,
					Long:   s.Description().Long,
					Schema: schema_,
				}
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(cmdDesc)
				cobra.CheckErr(err)
				fmt.Println(sb.String())
				os.Exit(0)
			}
		}

		// Determine whether to run in Glaze mode or classic mode
		useGlazeMode := false
		if cfg.DualMode {
			if cfg.DefaultToGlaze {
				noGlaze, _ := cmd.Flags().GetBool("no-glaze-output")
				useGlazeMode = !noGlaze
			} else {
				useGlazeMode, _ = cmd.Flags().GetBool(cfg.GlazeToggleFlag)
			}
		} else {
			// default: if implements GlazeCommand, use glaze mode
			if _, ok := s.(cmds.GlazeCommand); ok {
				useGlazeMode = true
			}
		}
		if useGlazeMode {
			// Run in glaze mode
			glazeCmd, ok := s.(cmds.GlazeCommand)
			if !ok {
				cobra.CheckErr(errors.New("Glaze mode requested but command does not implement GlazeCommand"))
				return
			}
			glazedSectionValues, ok := parsedValues.Get(settings.GlazedSlug)
			if !ok {
				cobra.CheckErr(errors.New("glazed section not found"))
				return
			}
			gp, err := settings.SetupTableProcessor(glazedSectionValues)
			cobra.CheckErr(err)
			_, err = settings.SetupProcessorOutput(gp, glazedSectionValues, os.Stdout)
			cobra.CheckErr(err)

			// Add signal handling for all command types
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			err = glazeCmd.RunIntoGlazeProcessor(ctx, parsedValues, gp)
			var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
			if errors.As(err, &exitWithoutGlazeError) {
				return
			}
			if !errors.Is(err, context.Canceled) {
				cobra.CheckErr(err)
			}
			// Close will run the TableMiddlewares
			err = gp.Close(ctx)
			cobra.CheckErr(err)
			return
		}

		// Classic mode: run the provided runFunc
		// Add signal handling for all command types
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		err = runFunc(ctx, parsedValues)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			os.Exit(0)
		}
		cobra.CheckErr(err)
	}
}

func BuildCobraCommandFromCommandAndFunc(
	s cmds.Command,
	run CobraRunFunc,
	opts ...CobraOption,
) (*cobra.Command, error) {
	// Initialize builder config with defaults
	cfg := &commandBuildConfig{
		DualMode:         false,
		GlazeToggleFlag:  "with-glaze-output",
		DefaultToGlaze:   false,
		HiddenGlazeFlags: nil,
		ParserCfg:        CobraParserConfig{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Start with the original description
	description := s.Description()
	// If the command implements GlazeCommand, ensure a glazed section is present
	if _, isGlazeCmd := s.(cmds.GlazeCommand); isGlazeCmd {
		originalSchema := description.Schema
		glazedSchema := originalSchema.Clone()
		if _, ok := glazedSchema.Get(settings.GlazedSlug); !ok {
			glazedSection, err := settings.NewGlazedSection()
			if err != nil {
				return nil, err
			}
			glazedSchema.Set(settings.GlazedSlug, glazedSection)
		}
		// clone the description so we don't mutate the original
		newDesc := description.Clone(false)
		newDesc.Schema = glazedSchema
		description = newDesc
	}
	cmd := NewCobraCommandFromCommandDescription(description)
	// Add glaze toggle flag if dual mode is enabled
	if cfg.DualMode {
		if cfg.DefaultToGlaze {
			cmd.Flags().Bool("no-glaze-output", false, "Disable glaze output mode")
		} else {
			cmd.Flags().Bool(cfg.GlazeToggleFlag, false, "Switch this run to Glaze structured output")
		}
	}
	// Create parser with configured parser settings
	cobraParser, err := NewCobraParserFromSections(description.Schema, &cfg.ParserCfg)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not create cobra parser")
		return nil, err
	}
	err = cobraParser.AddToCobraCommand(cmd)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not add to cobra command")
		return nil, err
	}
	// Hide specified glaze flags if requested
	if cfg.DualMode {
		for _, name := range cfg.HiddenGlazeFlags {
			if flag := cmd.Flags().Lookup(name); flag != nil {
				flag.Hidden = true
			}
		}
	}
	// Use the refactored run helper
	runCobraCommand(cmd, s, run, cobraParser, cfg)
	return cmd, nil
}

func BuildCobraCommandAlias(
	alias *alias.CommandAlias,
	opts ...CobraOption,
) (*cobra.Command, error) {
	cmd, err := BuildCobraCommand(alias.AliasedCommand, opts...)
	if err != nil {
		return nil, err
	}

	origRun := cmd.Run

	cmd.Use = alias.Name
	description := alias.AliasedCommand.Description()
	cmd.Short = fmt.Sprintf("Alias for %s", description.Name)

	minArgs := 0
	argumentDefinitions := description.GetDefaultArguments()
	provided, err := argumentDefinitions.GatherArguments(
		alias.Arguments, true, true,
		fields.WithSource("cobra-alias"),
	)
	if err != nil {
		return nil, err
	}

	argumentDefinitions.ForEach(func(argDef *fields.Definition) {
		_, ok := provided.Get(argDef.Name)
		if argDef.Required && !ok {
			minArgs++
		}
	})

	cmd.Args = cobra.MinimumNArgs(minArgs)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		for k, v := range alias.Flags {
			if !cmd.Flags().Changed(k) {
				err = cmd.Flags().Set(k, v)
				cobra.CheckErr(err)
			}
		}
		if len(args) == 0 {
			args = alias.Arguments
		}
		origRun(cmd, args)
	}

	return cmd, nil
}

// findOrCreateParentCommand will create empty commands to anchor the passed in parents.
func findOrCreateParentCommand(rootCmd *cobra.Command, parents []string) *cobra.Command {
	parentCmd := rootCmd
	for _, parent := range parents {
		subCmd, _, _ := parentCmd.Find([]string{parent})
		if subCmd == nil || subCmd == parentCmd {
			newParentCmd := &cobra.Command{
				Use:   parent,
				Short: fmt.Sprintf("All commands for %s", parent),
			}
			parentCmd.AddCommand(newParentCmd)
			parentCmd = newParentCmd
		} else {
			parentCmd = subCmd
		}
	}
	return parentCmd
}

// BuildCobraCommand is an alias to help with LLM hallucinations
func BuildCobraCommand(
	command cmds.Command,
	opts ...CobraOption,
) (*cobra.Command, error) {
	return BuildCobraCommandFromCommand(command, opts...)
}

// Unified builder: determines runFunc based on implemented interfaces and
// delegates to BuildCobraCommandFromCommandAndFunc
func BuildCobraCommandFromCommand(
	s cmds.Command,
	opts ...CobraOption,
) (*cobra.Command, error) {
	// Generic run function for classic mode (WriterCommand or BareCommand)
	runFunc := func(ctx context.Context, parsedValues *values.Values) error {
		if writerCmd, ok := s.(cmds.WriterCommand); ok {
			err := writerCmd.RunIntoWriter(ctx, parsedValues, os.Stdout)
			if _, exitWithoutGlaze := err.(*cmds.ExitWithoutGlazeError); exitWithoutGlaze {
				return err
			}
			if err != context.Canceled {
				return err
			}
			return nil
		}
		if bareCmd, ok := s.(cmds.BareCommand); ok {
			err := bareCmd.Run(ctx, parsedValues)
			if _, exitWithoutGlaze := err.(*cmds.ExitWithoutGlazeError); exitWithoutGlaze {
				return err
			}
			if err != context.Canceled {
				return err
			}
			return nil
		}
		return errors.Errorf("no non-Glaze run method implemented for %T", s)
	}
	return BuildCobraCommandFromCommandAndFunc(s, runFunc, opts...)
}

func AddCommandsToRootCommand(
	rootCmd *cobra.Command,
	commands []cmds.Command,
	aliases []*alias.CommandAlias,
	opts ...CobraOption,
) error {
	commandsByName := map[string]cmds.Command{}

	for _, command := range commands {
		// find the proper subcommand, or create if it doesn't exist
		description := command.Description()
		parentCmd := findOrCreateParentCommand(rootCmd, description.Parents)

		cobraCommand, err := BuildCobraCommandFromCommand(command, opts...)
		if err != nil {
			log.Warn().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not build cobra command")
			return nil
		}
		parentCmd.AddCommand(cobraCommand)
		commandsByName[description.Name] = command

		path := strings.Join(append(description.Parents, description.Name), " ")
		commandsByName[path] = command
	}

	for _, alias := range aliases {
		path := strings.Join(alias.Parents, " ")
		aliasedCommand, ok := commandsByName[path]
		if !ok {
			return errors.Errorf("Command %s not found for alias %s", path, alias.Name)
		}
		alias.AliasedCommand = aliasedCommand

		parentCmd := findOrCreateParentCommand(rootCmd, alias.Parents)
		cobraCommand, err := BuildCobraCommandAlias(alias, opts...)
		if err != nil {
			return err
		}
		parentCmd.AddCommand(cobraCommand)
	}

	return nil
}

// Insert foundation types for unified builder options
// CobraOption configures command and parser builder settings
type CobraOption func(cfg *commandBuildConfig)

// commandBuildConfig aggregates all builder options internally
type commandBuildConfig struct {
	DualMode         bool
	GlazeToggleFlag  string
	DefaultToGlaze   bool
	HiddenGlazeFlags []string
	ParserCfg        CobraParserConfig
}

// WithParserConfig sets parser customization on the builder
func WithParserConfig(cfg CobraParserConfig) CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg = cfg
	}
}

// WithDualMode enables or disables dual-mode behavior
func WithDualMode(enabled bool) CobraOption {
	return func(c *commandBuildConfig) {
		c.DualMode = enabled
	}
}

// WithGlazeToggleFlag customizes the glaze toggle flag name
func WithGlazeToggleFlag(name string) CobraOption {
	return func(c *commandBuildConfig) {
		c.GlazeToggleFlag = name
	}
}

// WithHiddenGlazeFlags marks glaze flags to remain hidden
func WithHiddenGlazeFlags(names ...string) CobraOption {
	return func(c *commandBuildConfig) {
		c.HiddenGlazeFlags = names
	}
}

// WithDefaultToGlaze makes glaze mode the default unless negated
func WithDefaultToGlaze() CobraOption {
	return func(c *commandBuildConfig) {
		c.DefaultToGlaze = true
	}
}

// WithCobraShortHelpSections sets the sections shown in short help.
func WithCobraShortHelpSections(sections ...string) CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.ShortHelpSections = sections
	}
}

// WithCobraMiddlewaresFunc sets a custom middleware function for parsing (deprecated)
func WithCobraMiddlewaresFunc(fn CobraMiddlewaresFunc) CobraOption {
	return func(c *commandBuildConfig) {
		if fn != nil {
			c.ParserCfg.MiddlewaresFunc = fn
		}
	}
}

// WithSkipCommandSettingsSection hides the command settings section flags.
func WithSkipCommandSettingsSection() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.SkipCommandSettingsSection = true
	}
}

// WithProfileSettingsSection enables the profile settings section.
func WithProfileSettingsSection() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.EnableProfileSettingsSection = true
	}
}

// WithCreateCommandSettingsSection enables the create-command settings section.
func WithCreateCommandSettingsSection() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.EnableCreateCommandSettingsSection = true
	}
}
