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

// CobraOption is the unified option type for configuring both command builder and parser
type CobraOption func(*commandBuildConfig)

// commandBuildConfig is the internal configuration struct that holds all possible 
// configurations for building a command, including parser settings and dual-mode behavior
type commandBuildConfig struct {
	DualMode         bool
	GlazeToggleFlag  string
	DefaultToGlaze   bool
	HiddenGlazeFlags []string
	ParserCfg        CobraParserConfig
}

type CobraRunFunc func(ctx context.Context, parsedLayers *layers.ParsedLayers) error

// Helper options that return CobraOption

// WithParserConfig sets the entire parser configuration
func WithParserConfig(cfg CobraParserConfig) CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg = cfg
	}
}

// WithDualMode enables or disables dual mode functionality
func WithDualMode(enabled bool) CobraOption {
	return func(config *commandBuildConfig) {
		config.DualMode = enabled
	}
}

// WithGlazeToggleFlag sets the name of the flag used to toggle glaze mode
func WithGlazeToggleFlag(name string) CobraOption {
	return func(config *commandBuildConfig) {
		config.GlazeToggleFlag = name
	}
}

// WithHiddenGlazeFlags specifies which glaze flags to keep hidden
func WithHiddenGlazeFlags(names ...string) CobraOption {
	return func(config *commandBuildConfig) {
		config.HiddenGlazeFlags = names
	}
}

// WithDefaultToGlaze makes glaze mode the default
func WithDefaultToGlaze() CobraOption {
	return func(config *commandBuildConfig) {
		config.DefaultToGlaze = true
	}
}

// Convenience functions for parser configuration that populate ParserCfg

// WithCobraMiddlewaresFunc sets the middlewares function for the parser
func WithCobraMiddlewaresFunc(fn CobraMiddlewaresFunc) CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg.MiddlewaresFunc = fn
	}
}

// WithCobraShortHelpLayers sets the short help layers for the parser
func WithCobraShortHelpLayers(layers ...string) CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg.ShortHelpLayers = append(config.ParserCfg.ShortHelpLayers, layers...)
	}
}

// WithSkipCommandSettingsLayer configures the parser to skip the command settings layer
func WithSkipCommandSettingsLayer() CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg.SkipCommandSettingsLayer = true
	}
}

// WithProfileSettingsLayer enables the profile settings layer in the parser
func WithProfileSettingsLayer() CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg.EnableProfileSettingsLayer = true
	}
}

// WithCreateCommandSettingsLayer enables the create command settings layer in the parser
func WithCreateCommandSettingsLayer() CobraOption {
	return func(config *commandBuildConfig) {
		config.ParserCfg.EnableCreateCommandSettingsLayer = true
	}
}

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
	cobraParser, err := NewCobraParserFromLayersWithOptions(description.Layers, options...)
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
	cmd, err := BuildCobraCommand(alias.AliasedCommand, options...)
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

// BuildCobraCommand is an alias to help with LLM hallucinations
// Deprecated: Use BuildCobraCommandFromCommand with CobraOption instead
func BuildCobraCommand(
	command cmds.Command,
	options ...CobraParserOption,
) (*cobra.Command, error) {
	// Convert legacy options to new format
	cfg := []CobraOption{}
	
	// Create a temporary parser config to extract values
	tempParserCfg := &CobraParserConfig{
		MiddlewaresFunc: CobraCommandDefaultMiddlewares,
	}
	
	// Apply legacy options to temporary parser to extract config values
	tempParser := &CobraParser{
		middlewaresFunc: CobraCommandDefaultMiddlewares,
	}
	
	for _, option := range options {
		err := option(tempParser)
		if err != nil {
			return nil, err
		}
	}

	// Transfer values to new config format
	tempParserCfg.MiddlewaresFunc = tempParser.middlewaresFunc
	tempParserCfg.ShortHelpLayers = tempParser.shortHelpLayers
	tempParserCfg.SkipCommandSettingsLayer = tempParser.skipCommandSettingsLayer
	tempParserCfg.EnableProfileSettingsLayer = tempParser.enableProfileSettingsLayer
	tempParserCfg.EnableCreateCommandSettingsLayer = tempParser.enableCreateCommandSettingsLayer

	cfg = append(cfg, WithParserConfig(*tempParserCfg))
	
	return BuildCobraCommandFromCommand(command, cfg...)
}

func BuildCobraCommandFromCommand(
	command cmds.Command,
	options ...CobraOption,
) (*cobra.Command, error) {
	// 1. Create default config
	cfg := &commandBuildConfig{
		GlazeToggleFlag: "with-glaze-output",
		ParserCfg: CobraParserConfig{
			MiddlewaresFunc: CobraCommandDefaultMiddlewares,
		},
	}

	// 2. Apply all options
	for _, opt := range options {
		opt(cfg)
	}

	// 3. Build the command based on config
	if cfg.DualMode {
		return buildDualModeCommand(command, cfg)
	} else {
		return buildSingleModeCommand(command, cfg)
	}
}

// buildSingleModeCommand builds a command in single mode using prioritized interface checking
func buildSingleModeCommand(command cmds.Command, cfg *commandBuildConfig) (*cobra.Command, error) {
	// Priority order: GlazeCommand, WriterCommand, BareCommand
	var runFunc CobraRunFunc
	var paramLayers *layers.ParameterLayers

	if glazeCmd, ok := command.(cmds.GlazeCommand); ok {
		// Use GlazeCommand if available
		description := glazeCmd.Description()
		paramLayers = description.Layers.Clone()
		
		// Add glazed layer if not present
		if _, hasGlazedLayer := paramLayers.Get(settings.GlazedSlug); !hasGlazedLayer {
			glazedLayer, err := settings.NewGlazedParameterLayers()
			if err != nil {
				return nil, err
			}
			paramLayers.Set(settings.GlazedSlug, glazedLayer)
		}

		runFunc = func(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
			glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
			if !ok {
				return errors.New("glazed layer not found")
			}
			gp, err := settings.SetupTableProcessor(glazedLayer)
			if err != nil {
				return err
			}

			_, err = settings.SetupProcessorOutput(gp, glazedLayer, os.Stdout)
			if err != nil {
				return err
			}

			err = glazeCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
			var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
			if errors.As(err, &exitWithoutGlazeError) {
				return nil
			}
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}

			return gp.Close(ctx)
		}
	} else if writerCmd, ok := command.(cmds.WriterCommand); ok {
		// Use WriterCommand as fallback
		description := writerCmd.Description()
		paramLayers = description.Layers.Clone()
		
		runFunc = func(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
			return writerCmd.RunIntoWriter(ctx, parsedLayers, os.Stdout)
		}
	} else if bareCmd, ok := command.(cmds.BareCommand); ok {
		// Use BareCommand as final fallback
		description := bareCmd.Description()
		paramLayers = description.Layers.Clone()
		
		runFunc = func(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
			return bareCmd.Run(ctx, parsedLayers)
		}
	} else {
		return nil, errors.Errorf("Command does not implement any supported interface: %T", command)
	}

	return buildCobraCommandFromLayers(command.Description(), paramLayers, runFunc, &cfg.ParserCfg)
}

// buildCobraCommandFromLayers creates a cobra command from layers and a run function
func buildCobraCommandFromLayers(
	description *cmds.CommandDescription,
	layers *layers.ParameterLayers,
	runFunc CobraRunFunc,
	parserCfg *CobraParserConfig,
) (*cobra.Command, error) {
	cmd := NewCobraCommandFromCommandDescription(description)
	
	cobraParser, err := NewCobraParserFromLayers(layers, parserCfg)
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
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			err := cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Handle command settings debugging
		if handleCommandSettingsDebug(cmd, parsedLayers) {
			return
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		err = runFunc(ctx, parsedLayers)
		var exitWithoutGlazeError *cmds.ExitWithoutGlazeError
		if errors.As(err, &exitWithoutGlazeError) {
			return
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			cobra.CheckErr(err)
		}
	}

	return cmd, nil
}

// handleCommandSettingsDebug handles debugging commands like --print-yaml, --print-parsed-parameters, etc.
// Returns true if a debug command was handled and execution should stop
func handleCommandSettingsDebug(cmd *cobra.Command, parsedLayers *layers.ParsedLayers) bool {
	// Handle command settings debugging first
	commandSettings := &CommandSettings{}
	if minimalLayer, ok := parsedLayers.Get(CommandSettingsSlug); ok {
		var printYAML, printParsedParameters_, printSchema bool
		err := minimalLayer.InitializeStruct(commandSettings)
		cobra.CheckErr(err)
		printYAML = commandSettings.PrintYAML
		printParsedParameters_ = commandSettings.PrintParsedParameters
		printSchema = commandSettings.PrintSchema

		if printParsedParameters_ {
			printParsedParameters(parsedLayers)
			return true
		}

		// Get command from context - we need to access the original command for YAML/schema
		// This is a limitation - we don't have access to the original command here
		// For now, we'll skip these debug features in the unified API
		if printYAML || printSchema {
			_, _ = fmt.Fprintln(os.Stderr, "YAML and schema debug features not yet supported in unified API")
			return true
		}
	}

	// Handle create command settings
	if createCommandLayer, ok := parsedLayers.Get(CreateCommandSettingsSlug); ok {
		createCommandSettings := &CreateCommandSettings{}
		err := createCommandLayer.InitializeStruct(createCommandSettings)
		cobra.CheckErr(err)

		if createCommandSettings.CreateCliopatra != "" {
			verbs := GetVerbsFromCobraCommand(cmd)
			if len(verbs) == 0 {
				cobra.CheckErr(errors.New("could not get verbs from cobra command"))
			}
			// We need the original command description here too
			_, _ = fmt.Fprintln(os.Stderr, "Create cliopatra feature not yet supported in unified API")
			return true
		}

		if createCommandSettings.CreateAlias != "" {
			// This one we can handle since it only needs the current cmd
			alias := &alias.CommandAlias{
				Name:      createCommandSettings.CreateAlias,
				Arguments: []string{}, // args not available here
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
			return true
		}

		if createCommandSettings.CreateCommand != "" {
			_, _ = fmt.Fprintln(os.Stderr, "Create command feature not yet supported in unified API")
			return true
		}
	}

	return false
}

// buildDualModeCommand builds a command that can operate in both classic and glaze modes
func buildDualModeCommand(command cmds.Command, cfg *commandBuildConfig) (*cobra.Command, error) {
	description := command.Description()
	glazedLayers := description.Layers.Clone()
	
	// Check if command already has glazed layer
	_, hasGlazedLayer := glazedLayers.Get(settings.GlazedSlug)
	
	// Add glazed layer if not present
	if !hasGlazedLayer {
		glazedLayer, err := settings.NewGlazedParameterLayers()
		if err != nil {
			return nil, err
		}
		glazedLayers.Set(settings.GlazedSlug, glazedLayer)
	}

	// Create modified description with glazed layers
	modifiedDescription := description.Clone(false)
	modifiedDescription.Layers = glazedLayers

	cmd := NewCobraCommandFromCommandDescription(modifiedDescription)

	// Add the glaze toggle flag
	if cfg.DefaultToGlaze {
		cmd.Flags().Bool("no-glaze-output", false, "Disable glaze output mode")
	} else {
		cmd.Flags().Bool(cfg.GlazeToggleFlag, false, "Switch this run to Glaze structured output")
	}

	cobraParser, err := NewCobraParserFromLayers(modifiedDescription.Layers, &cfg.ParserCfg)
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
	for _, flagName := range cfg.HiddenGlazeFlags {
		if flag := cmd.Flags().Lookup(flagName); flag != nil {
			flag.Hidden = true
		}
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		parsedLayers, err := cobraParser.Parse(cmd, args)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			err := cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// Handle command settings debugging
		if handleCommandSettingsDebug(cmd, parsedLayers) {
			return
		}

		// Determine which mode to use
		useGlazeMode := false
		if cfg.DefaultToGlaze {
			noGlazeOutput, _ := cmd.Flags().GetBool("no-glaze-output")
			useGlazeMode = !noGlazeOutput
		} else {
			useGlazeMode, _ = cmd.Flags().GetBool(cfg.GlazeToggleFlag)
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
			glazeCmd, ok := command.(cmds.GlazeCommand)
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
			if writerCmd, ok := command.(cmds.WriterCommand); ok {
				runErr = writerCmd.RunIntoWriter(ctx, parsedLayers, os.Stdout)
			} else if bareCmd, ok := command.(cmds.BareCommand); ok {
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

		cobraCommand, err := BuildCobraCommand(command, options...)
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



