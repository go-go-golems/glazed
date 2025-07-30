package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli/cliopatra"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"

	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type CobraRunFunc func(ctx context.Context, parsedLayers *layers.ParsedLayers) error

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
		// Parse layers
		parsedLayers, err := parser.Parse(cmd, args)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			_ = cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Minimal command settings: debug flags
		commandSettings := &CommandSettings{}
		if minimalLayer, ok := parsedLayers.Get(CommandSettingsSlug); ok {
			var printYAML, printParsedParameters_, printSchema bool
			err = minimalLayer.InitializeStruct(commandSettings)
			cobra.CheckErr(err)
			printYAML = commandSettings.PrintYAML
			printParsedParameters_ = commandSettings.PrintParsedParameters
			printSchema = commandSettings.PrintSchema

			if printParsedParameters_ {
				printParsedParameters(parsedLayers)
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
		if createLayer, ok := parsedLayers.Get(CreateCommandSettingsSlug); ok {
			createSettings := &CreateCommandSettings{}
			err = createLayer.InitializeStruct(createSettings)
			cobra.CheckErr(err)

			if createSettings.CreateCliopatra != "" {
				verbs := GetVerbsFromCobraCommand(cmd)
				if len(verbs) == 0 {
					cobra.CheckErr(errors.New("could not get verbs from cobra command"))
				}
				p := cliopatra.NewProgramFromCapture(
					s.Description(),
					parsedLayers,
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
				layers_ := s.Description().Layers.Clone()
				cmdDesc := &cmds.CommandDescription{
					Name:   createSettings.CreateCommand,
					Short:  s.Description().Short,
					Long:   s.Description().Long,
					Layers: layers_,
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
			glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
			if !ok {
				cobra.CheckErr(errors.New("glazed layer not found"))
				return
			}
			gp, err := settings.SetupTableProcessor(glazedLayer)
			cobra.CheckErr(err)
			_, err = settings.SetupProcessorOutput(gp, glazedLayer, os.Stdout)
			cobra.CheckErr(err)
			err = glazeCmd.RunIntoGlazeProcessor(cmd.Context(), parsedLayers, gp)
			var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
			if errors.As(err, &exitWithoutGlazeError) {
				return
			}
			if !errors.Is(err, context.Canceled) {
				cobra.CheckErr(err)
			}
			// Close will run the TableMiddlewares
			err = gp.Close(cmd.Context())
			cobra.CheckErr(err)
			return
		}

		// Classic mode: run the provided runFunc
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()
		err = runFunc(ctx, parsedLayers)
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
	// If the command implements GlazeCommand, ensure a glazed parameter layer is present
	if _, isGlazeCmd := s.(cmds.GlazeCommand); isGlazeCmd {
		originalLayers := description.Layers
		glazedLayers := originalLayers.Clone()
		if _, ok := glazedLayers.Get(settings.GlazedSlug); !ok {
			glLayer, err := settings.NewGlazedParameterLayers()
			if err != nil {
				return nil, err
			}
			glazedLayers.Set(settings.GlazedSlug, glLayer)
		}
		// clone the description so we don't mutate the original
		newDesc := description.Clone(false)
		newDesc.Layers = glazedLayers
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
	cobraParser, err := NewCobraParserFromLayers(description.Layers, &cfg.ParserCfg)
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
		parameters.WithParseStepSource("cobra-alias"),
	)
	if err != nil {
		return nil, err
	}

	argumentDefinitions.ForEach(func(argDef *parameters.ParameterDefinition) {
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
	runFunc := func(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
		if writerCmd, ok := s.(cmds.WriterCommand); ok {
			err := writerCmd.RunIntoWriter(ctx, parsedLayers, os.Stdout)
			if _, exitWithoutGlaze := err.(*cmds.ExitWithoutGlazeError); exitWithoutGlaze {
				return err
			}
			if err != context.Canceled {
				return err
			}
			return nil
		}
		if bareCmd, ok := s.(cmds.BareCommand); ok {
			err := bareCmd.Run(ctx, parsedLayers)
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

// Backwards compatibility helpers for old parser options
// WithCobraShortHelpLayers sets the layers shown in short help (deprecated)
func WithCobraShortHelpLayers(layers ...string) CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.ShortHelpLayers = layers
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

// Deprecated: use BuildCobraCommandFromCommand(c, WithDualMode(true)).
func BuildCobraCommandDualMode(
	c cmds.Command,
	_ ...interface{},
) (*cobra.Command, error) {
	return BuildCobraCommandFromCommand(c, WithDualMode(true))
}

// Deprecated wrappers for backwards compatibility with earlier APIs
// Use BuildCobraCommand or BuildCobraCommand from the unified API instead.
// Deprecated: use BuildCobraCommand(c, opts...)
func BuildCobraCommandFromBareCommand(c cmds.BareCommand, opts ...CobraOption) (*cobra.Command, error) {
	return BuildCobraCommand(c, opts...)
}

// Deprecated: use BuildCobraCommand(c, opts...)
func BuildCobraCommandFromWriterCommand(s cmds.WriterCommand, opts ...CobraOption) (*cobra.Command, error) {
	return BuildCobraCommand(s, opts...)
}

// Deprecated: use BuildCobraCommand(c, opts...)
func BuildCobraCommandFromGlazeCommand(cmd_ cmds.GlazeCommand, opts ...CobraOption) (*cobra.Command, error) {
	return BuildCobraCommand(cmd_, opts...)
}

// WithSkipCommandSettingsLayer hides the command settings layer flags (deprecated)
func WithSkipCommandSettingsLayer() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.SkipCommandSettingsLayer = true
	}
}

// WithProfileSettingsLayer enables the profile settings layer (deprecated)
func WithProfileSettingsLayer() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.EnableProfileSettingsLayer = true
	}
}

// WithCreateCommandSettingsLayer enables the create-command settings layer (deprecated)
func WithCreateCommandSettingsLayer() CobraOption {
	return func(c *commandBuildConfig) {
		c.ParserCfg.EnableCreateCommandSettingsLayer = true
	}
}
