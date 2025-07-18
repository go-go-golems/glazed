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

func BuildCobraCommandFromCommandAndFunc(
	s cmds.Command,
	run CobraRunFunc,
	options ...CobraParserOption,
) (*cobra.Command, error) {
	description := s.Description()

	cmd := NewCobraCommandFromCommandDescription(description)
	cobraParser, err := NewCobraParserFromLayers(description.Layers, options...)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not create cobra parser")
		return nil, err
	}
	err = cobraParser.AddToCobraCommand(cmd)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not add to cobra command")
		return nil, err
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		parsedLayers, err := cobraParser.Parse(cmd, args)
		// show help if there is an error
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			err := cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Try minimal command settings
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

		// Handle the rest of the full command settings if available
		if createCommandLayer, ok := parsedLayers.Get(CreateCommandSettingsSlug); ok {
			createCommandSettings := &CreateCommandSettings{}
			err = createCommandLayer.InitializeStruct(createCommandSettings)
			cobra.CheckErr(err)

			if createCommandSettings.CreateCliopatra != "" {
				verbs := GetVerbsFromCobraCommand(cmd)
				if len(verbs) == 0 {
					cobra.CheckErr(errors.New("could not get verbs from cobra command"))
				}
				p := cliopatra.NewProgramFromCapture(
					s.Description(),
					parsedLayers,
					cliopatra.WithVerbs(verbs[1:]...),
					cliopatra.WithName(createCommandSettings.CreateCliopatra),
					cliopatra.WithPath(verbs[0]),
				)

				// print as yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(p)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createCommandSettings.CreateAlias != "" {
				alias := &alias.CommandAlias{
					Name:      createCommandSettings.CreateAlias,
					AliasFor:  description.Name,
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

				// marshal alias to yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(alias)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createCommandSettings.CreateCommand != "" {
				// XXX this is broken now I think anyway
				layers_ := description.Layers.Clone()

				cmd := &cmds.CommandDescription{
					Name:   createCommandSettings.CreateCommand,
					Short:  description.Short,
					Long:   description.Long,
					Layers: layers_,
				}

				// encode as yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(cmd)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		err = run(ctx, parsedLayers)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			os.Exit(0)
		}

		cobra.CheckErr(err)
	}

	return cmd, nil
}

func BuildCobraCommandFromBareCommand(c cmds.BareCommand, options ...CobraParserOption) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommandAndFunc(c, func(
		ctx context.Context,
		parsedLayers *layers.ParsedLayers,
	) error {
		err := c.Run(ctx, parsedLayers)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			return nil
		}
		if err != context.Canceled {
			cobra.CheckErr(err)
		}
		return nil
	}, options...)

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func BuildCobraCommandFromWriterCommand(s cmds.WriterCommand, options ...CobraParserOption) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommandAndFunc(s, func(
		ctx context.Context,
		parsedLayers *layers.ParsedLayers,
	) error {
		err := s.RunIntoWriter(ctx, parsedLayers, os.Stdout)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			return nil
		}
		if err != context.Canceled {
			cobra.CheckErr(err)
		}
		return nil
	}, options...)

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func BuildCobraCommandAlias(
	alias *alias.CommandAlias,
	options ...CobraParserOption,
) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommand(alias.AliasedCommand, options...)
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

func BuildCobraCommandFromGlazeCommand(cmd_ cmds.GlazeCommand, options ...CobraParserOption) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommandAndFunc(cmd_, func(
		ctx context.Context,
		parsedLayers *layers.ParsedLayers,
	) error {
		glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
		if !ok {
			return errors.New("glazed layer not found")
		}
		gp, err := settings.SetupTableProcessor(glazedLayer)
		cobra.CheckErr(err)

		_, err = settings.SetupProcessorOutput(gp, glazedLayer, os.Stdout)
		cobra.CheckErr(err)

		err = cmd_.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
		var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
		if errors.As(err, &exitWithoutGlazeError) {
			return nil
		}
		if !errors.Is(err, context.Canceled) {
			cobra.CheckErr(err)
		}

		// Close will run the TableMiddlewares
		err = gp.Close(ctx)
		cobra.CheckErr(err)

		return nil
	},
		options...)

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func BuildCobraCommandFromCommand(
	command cmds.Command,
	options ...CobraParserOption,
) (*cobra.Command, error) {
	var cobraCommand *cobra.Command
	var err error
	switch c := command.(type) {
	case cmds.BareCommand:
		cobraCommand, err = BuildCobraCommandFromBareCommand(c, options...)

	case cmds.WriterCommand:
		cobraCommand, err = BuildCobraCommandFromWriterCommand(c, options...)

	case cmds.GlazeCommand:
		cobraCommand, err = BuildCobraCommandFromGlazeCommand(c, options...)

	default:
		return nil, errors.Errorf("Unknown command type %T", c)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Error building command %s", command.Description().Name)
	}

	return cobraCommand, nil
}

func AddCommandsToRootCommand(
	rootCmd *cobra.Command,
	commands []cmds.Command,
	aliases []*alias.CommandAlias,
	options ...CobraParserOption,
) error {
	commandsByName := map[string]cmds.Command{}

	for _, command := range commands {
		// find the proper subcommand, or create if it doesn't exist
		description := command.Description()
		parentCmd := findOrCreateParentCommand(rootCmd, description.Parents)

		cobraCommand, err := BuildCobraCommandFromCommand(command, options...)
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
		cobraCommand, err := BuildCobraCommandAlias(alias, options...)
		if err != nil {
			return err
		}
		parentCmd.AddCommand(cobraCommand)
	}

	return nil
}

// DualModeOption provides customization options for BuildCobraCommandDualMode
type DualModeOption interface {
	apply(*dualModeConfig)
}

type dualModeConfig struct {
	glazeToggleFlag  string
	hiddenGlazeFlags []string
	defaultToGlaze   bool
}

type glazeToggleFlagOption struct {
	name string
}

func (g glazeToggleFlagOption) apply(cfg *dualModeConfig) {
	cfg.glazeToggleFlag = g.name
}

// WithGlazeToggleFlag lets you rename or shorten the toggle flag
func WithGlazeToggleFlag(name string) DualModeOption {
	return glazeToggleFlagOption{name: name}
}

type hiddenGlazeFlagsOption struct {
	flagNames []string
}

func (h hiddenGlazeFlagsOption) apply(cfg *dualModeConfig) {
	cfg.hiddenGlazeFlags = h.flagNames
}

// WithHiddenGlazeFlags marks specific Glaze‑layer flags to stay hidden even
// after the toggle; use when you expose only JSON rendering, for instance.
func WithHiddenGlazeFlags(flagNames ...string) DualModeOption {
	return hiddenGlazeFlagsOption{flagNames: flagNames}
}

type defaultToGlazeOption struct{}

func (d defaultToGlazeOption) apply(cfg *dualModeConfig) {
	cfg.defaultToGlaze = true
}

// WithDefaultToGlaze makes Glaze mode the default unless the user disables it
// with --no-glaze-output (builder auto‑creates the negated flag).
func WithDefaultToGlaze() DualModeOption {
	return defaultToGlazeOption{}
}

// BuildCobraCommandDualMode creates a cobra command that can run in both classic and glaze modes
func BuildCobraCommandDualMode(
	c cmds.Command,
	opts ...DualModeOption,
) (*cobra.Command, error) {
	config := &dualModeConfig{
		glazeToggleFlag: "with-glaze-output",
	}

	for _, opt := range opts {
		opt.apply(config)
	}

	description := c.Description()

	// Check if we need to inject a glazed layer for glaze mode
	glazedLayers := description.Layers.Clone()
	hasGlazedLayer := false
	if _, ok := glazedLayers.Get(settings.GlazedSlug); ok {
		hasGlazedLayer = true
	}

	// If command implements GlazeCommand but doesn't have glazed layer, inject one
	if _, isGlazeCommand := c.(cmds.GlazeCommand); isGlazeCommand && !hasGlazedLayer {
		glazedLayer, err := settings.NewGlazedParameterLayers()
		if err != nil {
			return nil, err
		}
		glazedLayers.Set(settings.GlazedSlug, glazedLayer)
	}

	// Create a modified command description with the potential glazed layer
	modifiedDescription := description.Clone(false)
	modifiedDescription.Layers = glazedLayers

	cmd := NewCobraCommandFromCommandDescription(modifiedDescription)

	// Add the glaze toggle flag
	if config.defaultToGlaze {
		cmd.Flags().Bool("no-glaze-output", false, "Disable glaze output mode")
	} else {
		cmd.Flags().Bool(config.glazeToggleFlag, false, "Switch this run to Glaze structured output")
	}

	cobraParser, err := NewCobraParserFromLayers(modifiedDescription.Layers)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not create cobra parser")
		return nil, err
	}

	err = cobraParser.AddToCobraCommand(cmd)
	if err != nil {
		log.Error().Err(err).Str("command", description.Name).Str("source", description.Source).Msg("Could not add to cobra command")
		return nil, err
	}

	// Hide glaze flags by default if glaze layer was injected
	if !hasGlazedLayer {
		glazedLayer, ok := glazedLayers.Get(settings.GlazedSlug)
		if ok {
			glazedLayer.GetParameterDefinitions().ForEach(func(pd *parameters.ParameterDefinition) {
				if flag := cmd.Flags().Lookup(pd.Name); flag != nil {
					flag.Hidden = true
				}
			})
		}
	}

	// Hide specific flags if requested
	for _, flagName := range config.hiddenGlazeFlags {
		if flag := cmd.Flags().Lookup(flagName); flag != nil {
			flag.Hidden = true
		}
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		parsedLayers, err := cobraParser.Parse(cmd, args)
		// show help if there is an error
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			err := cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Handle command settings debugging first (same as existing builders)
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
				err = c.ToYAML(os.Stdout)
				cobra.CheckErr(err)
				return
			}

			if printSchema {
				schema, err := c.Description().ToJsonSchema()
				cobra.CheckErr(err)
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				err = encoder.Encode(schema)
				cobra.CheckErr(err)
				return
			}
		}

		// Handle create command settings (same as existing builders)
		if createCommandLayer, ok := parsedLayers.Get(CreateCommandSettingsSlug); ok {
			createCommandSettings := &CreateCommandSettings{}
			err = createCommandLayer.InitializeStruct(createCommandSettings)
			cobra.CheckErr(err)

			if createCommandSettings.CreateCliopatra != "" {
				verbs := GetVerbsFromCobraCommand(cmd)
				if len(verbs) == 0 {
					cobra.CheckErr(errors.New("could not get verbs from cobra command"))
				}
				p := cliopatra.NewProgramFromCapture(
					c.Description(),
					parsedLayers,
					cliopatra.WithVerbs(verbs[1:]...),
					cliopatra.WithName(createCommandSettings.CreateCliopatra),
					cliopatra.WithPath(verbs[0]),
				)

				// print as yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(p)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createCommandSettings.CreateAlias != "" {
				alias := &alias.CommandAlias{
					Name:      createCommandSettings.CreateAlias,
					AliasFor:  description.Name,
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

				// marshal alias to yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(alias)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}

			if createCommandSettings.CreateCommand != "" {
				layers_ := description.Layers.Clone()

				cmd := &cmds.CommandDescription{
					Name:   createCommandSettings.CreateCommand,
					Short:  description.Short,
					Long:   description.Long,
					Layers: layers_,
				}

				// encode as yaml
				sb := strings.Builder{}
				encoder := yaml.NewEncoder(&sb)
				err = encoder.Encode(cmd)
				cobra.CheckErr(err)

				fmt.Println(sb.String())
				os.Exit(0)
			}
		}

		// Determine which mode to use
		useGlazeMode := false
		if config.defaultToGlaze {
			noGlazeOutput, _ := cmd.Flags().GetBool("no-glaze-output")
			useGlazeMode = !noGlazeOutput
		} else {
			useGlazeMode, _ = cmd.Flags().GetBool(config.glazeToggleFlag)
		}

		// Unhide glaze flags if glaze mode is requested and they were hidden
		if useGlazeMode && !hasGlazedLayer {
			glazedLayer, ok := glazedLayers.Get(settings.GlazedSlug)
			if ok {
				glazedLayer.GetParameterDefinitions().ForEach(func(pd *parameters.ParameterDefinition) {
					if flag := cmd.Flags().Lookup(pd.Name); flag != nil {
						flag.Hidden = false
					}
				})
			}
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		var runErr error
		if useGlazeMode {
			// Run in glaze mode
			glazeCmd, ok := c.(cmds.GlazeCommand)
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

			err = glazeCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
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
		} else {
			// Run in classic mode
			if writerCmd, ok := c.(cmds.WriterCommand); ok {
				runErr = writerCmd.RunIntoWriter(ctx, parsedLayers, os.Stdout)
			} else if bareCmd, ok := c.(cmds.BareCommand); ok {
				runErr = bareCmd.Run(ctx, parsedLayers)
			} else {
				cobra.CheckErr(errors.New("no non‑Glaze run method implemented"))
				return
			}

			if _, ok := runErr.(*cmds.ExitWithoutGlazeError); ok {
				return
			}
			if runErr != context.Canceled {
				cobra.CheckErr(runErr)
			}
		}
	}

	return cmd, nil
}
