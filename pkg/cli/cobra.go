package cli

import (
	"context"
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

// runCobraCommand executes the common run flow for all Cobra commands. It uses
// RunE so command and setup errors propagate to the application's Execute call;
// libraries must not choose process exit codes for their callers.
func runCobraCommand(
	cmd *cobra.Command,
	s cmds.Command,
	runFunc CobraRunFunc,
	parser *CobraParser,
	cfg *commandBuildConfig,
) {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Parse sections into values
		parsedValues, err := parser.Parse(cmd, args)
		if err != nil {
			_ = cmd.Help()
			return err
		}

		// Minimal command settings: debug flags
		if handled, err := HandleCommandSettings(s, parsedValues, os.Stdout); handled || err != nil {
			return err
		}

		// Create command settings: cliopatra, alias, create
		if createSectionValues, ok := parsedValues.Get(CreateCommandSettingsSlug); ok {
			createSettings := &CreateCommandSettings{}
			if err = createSectionValues.DecodeInto(createSettings); err != nil {
				return err
			}

			if createSettings.CreateCliopatra != "" {
				verbs := GetVerbsFromCobraCommand(cmd)
				if len(verbs) == 0 {
					return errors.New("could not get verbs from cobra command")
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
				if err = encoder.Encode(p); err != nil {
					return err
				}
				fmt.Println(sb.String())
				return nil
			}

			if createSettings.CreateAlias != "" {
				alias := &alias.CommandAlias{
					Name:      createSettings.CreateAlias,
					AliasFor:  alias.NewAliasTargetFromString(s.Description().Name),
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
				if err = encoder.Encode(alias); err != nil {
					return err
				}
				fmt.Println(sb.String())
				return nil
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
				if err = encoder.Encode(cmdDesc); err != nil {
					return err
				}
				fmt.Println(sb.String())
				return nil
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
				return errors.New("Glaze mode requested but command does not implement GlazeCommand")
			}
			structuredOutputValues, ok := parsedValues.Get(settings.StructuredOutputSlug)
			if !ok {
				return errors.New("structured output section not found")
			}
			gp, _, err := settings.SetupStructuredOutput(structuredOutputValues, os.Stdout)
			if err != nil {
				return err
			}

			// Add signal handling for all command types
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			err = glazeCmd.RunIntoGlazeProcessor(ctx, parsedValues, gp)
			var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
			if errors.As(err, &exitWithoutGlazeError) {
				return nil
			}
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			// Close will run the TableMiddlewares.
			return gp.Close(ctx)
		}

		// Classic mode: run the provided runFunc
		// Add signal handling for all command types
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		err = runFunc(ctx, parsedValues)
		var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
		if errors.As(err, &exitWithoutGlazeError) {
			return nil
		}
		return err
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
	// If the command implements GlazeCommand, ensure structured output settings are present.
	if _, isGlazeCmd := s.(cmds.GlazeCommand); isGlazeCmd {
		originalSchema := description.Schema
		structuredSchema := originalSchema.Clone()
		if _, ok := structuredSchema.Get(settings.StructuredOutputSlug); !ok {
			structuredOutputSection, err := settings.NewStructuredOutputSection()
			if err != nil {
				return nil, err
			}
			structuredSchema.Set(settings.StructuredOutputSlug, structuredOutputSection)
		}
		// clone the description so we don't mutate the original
		newDesc := description.Clone(false)
		newDesc.Schema = structuredSchema
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

	origRunE := cmd.RunE

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

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		for k, v := range alias.Flags {
			if !cmd.Flags().Changed(k) {
				if err := cmd.Flags().Set(k, v); err != nil {
					return err
				}
			}
		}
		if len(args) == 0 {
			args = alias.Arguments
		}
		return origRunE(cmd, args)
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
	type pendingCommand struct {
		parents []string
		command *cobra.Command
	}

	commandsByName := map[string]cmds.Command{}
	pending := make([]pendingCommand, 0, len(commands)+len(aliases))

	// Build every command before mutating the root. A flag collision or invalid
	// schema must not leave a partially mounted command tree.
	for _, command := range commands {
		description := command.Description()
		cobraCommand, err := BuildCobraCommandFromCommand(command, opts...)
		if err != nil {
			return errors.Wrapf(err, "could not build command %s from %s", description.FullPath(), description.Source)
		}
		pending = append(pending, pendingCommand{
			parents: append([]string(nil), description.Parents...),
			command: cobraCommand,
		})
		commandsByName[description.Name] = command

		pathParts := append(append([]string(nil), description.Parents...), description.Name)
		path := strings.Join(pathParts, " ")
		commandsByName[path] = command
	}

	for _, alias := range aliases {
		resolvedPath := alias.ResolveAliasedCommandPath()
		path := strings.Join(resolvedPath, " ")
		aliasedCommand, ok := commandsByName[path]
		if !ok {
			return errors.Errorf("command %s not found for alias %s", path, alias.Name)
		}
		alias.AliasedCommand = aliasedCommand

		cobraCommand, err := BuildCobraCommandAlias(alias, opts...)
		if err != nil {
			return errors.Wrapf(err, "could not build alias %s", alias.Name)
		}
		pending = append(pending, pendingCommand{
			parents: append([]string(nil), alias.Parents...),
			command: cobraCommand,
		})
	}

	for _, command := range pending {
		parentCmd := findOrCreateParentCommand(rootCmd, command.parents)
		parentCmd.AddCommand(command.command)
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
