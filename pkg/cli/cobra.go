package cli

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli/cliopatra"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	glazed_middlewares "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"os"
	"os/signal"
	"strings"
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

// CobraMiddlewaresFunc is a function that returns a list of middlewares for a cobra command.
// It can be used to overload the default middlewares for cobra commands
type CobraMiddlewaresFunc func(commandSettings *GlazedCommandSettings, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error)

func CobraCommandDefaultMiddlewares(commandSettings *GlazedCommandSettings, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error) {
	middlewares_ := []cmd_middlewares.Middleware{
		cmd_middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		cmd_middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	if commandSettings.LoadParametersFromFile != "" {
		middlewares_ = append(middlewares_,
			cmd_middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	middlewares_ = append(middlewares_,
		cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewares_, nil
}

func BuildCobraCommandFromCommandAndFunc(
	s cmds.Command,
	run CobraRunFunc,
	options ...CobraParserOption,
) (*cobra.Command, error) {
	description := s.Description()

	cobraParser, err := NewCobraParserFromCommandDescription(description, options...)
	if err != nil {
		return nil, err
	}

	cmd := cobraParser.Cmd

	cmd.Run = func(cmd *cobra.Command, args []string) {
		parsedLayers, err := cobraParser.Parse(cmd, args)
		// show help if there is an error
		if err != nil {
			fmt.Println(err)
			err := cmd.Help()
			cobra.CheckErr(err)
			os.Exit(1)
		}

		// get the command settings
		commandSettings := &GlazedCommandSettings{}
		if glazeCommandLayer, ok := parsedLayers.Get(GlazedCommandSlug); ok {
			err = glazeCommandLayer.InitializeStruct(commandSettings)
			cobra.CheckErr(err)
		}

		// TODO(manuel, 2023-12-28) After loading all the parameters, we potentially need to post process some layers
		// This is what ParseLayerFromCobraCommand is gdoing currently, and it seems the only place
		// it is actually used seems to be by the FieldsFilter galzed layer.
		// See Muji (1) sketchbook p.21

		// TODO(manuel, 2023-12-28) Handle GlazeCommandLayer options here
		if commandSettings.PrintYAML {
			err = s.ToYAML(os.Stdout)
			cobra.CheckErr(err)
			return
		}

		if commandSettings.CreateCliopatra != "" {
			verbs := GetVerbsFromCobraCommand(cmd)
			if len(verbs) == 0 {
				cobra.CheckErr(errors.New("could not get verbs from cobra command"))
			}
			p := cliopatra.NewProgramFromCapture(
				s.Description(),
				parsedLayers,
				cliopatra.WithVerbs(verbs[1:]...),
				cliopatra.WithName(commandSettings.CreateCliopatra),
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

		if commandSettings.CreateAlias != "" {
			alias := &alias.CommandAlias{
				Name:      commandSettings.CreateAlias,
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

		// TODO(manuel, 2023-02-26) This only outputs the command description, not the actual command
		// This is already helpful, but is really just half the story. To make this work
		// generically, CreateNewCommand() should be part of the interface.
		//
		// See https://github.com/go-go-golems/glazed/issues/170
		//
		// NOTE(manuel, 2023-12-21) Disabling this for now while I move flagand argument handling
		// over to be done by the default layer'
		//
		if commandSettings.CreateCommand != "" {
			layers_ := description.Layers.Clone()

			cmd := &cmds.CommandDescription{
				Name:   commandSettings.CreateCommand,
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

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		err = run(ctx, parsedLayers)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			os.Exit(0)
		}

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

func ParseLayersFromCobraCommand(cmd *cobra.Command, layers_ []layers.CobraParameterLayer) (
	*layers.ParsedLayers,
	error,
) {
	// TODO(manuel, 2023-12-28) Use middlewares to do the parsing here
	options := []layers.ParsedLayersOption{}

	for _, layer := range layers_ {
		ps, err := layer.ParseLayerFromCobraCommand(cmd)
		if err != nil {
			return nil, err
		}
		options = append(options, layers.WithParsedLayer(layer.GetSlug(), ps))
	}
	ret := layers.NewParsedLayers(options...)

	return ret, nil
}

// CreateGlazedProcessorFromCobra is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
//
// If you are more serious about using glazed, consider using the `cmds.GlazeCommand` and `parameters.ParameterDefinition`
// abstraction to define your CLI applications, which allows you to use layers and other nice features
// of the glazed ecosystem.
//
// If so, use SetupTableProcessor instead, and create a proper glazed.GlazeCommand for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (*glazed_middlewares.TableProcessor, formatters.OutputFormatter, error) {
	gpl, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	parsedLayer, err := gpl.ParseLayerFromCobraCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	gp, err := settings.SetupTableProcessor(parsedLayer)
	cobra.CheckErr(err)

	of, err := settings.SetupProcessorOutput(gp, parsedLayer, os.Stdout)
	cobra.CheckErr(err)

	return gp, of, nil
}

// AddGlazedProcessorFlagsToCobraCommand is a helper for cobra centric apps that quickly want to add

func AddGlazedProcessorFlagsToCobraCommand(cmd *cobra.Command, options ...settings.GlazeParameterLayerOption) error {
	gpl, err := settings.NewGlazedParameterLayers(options...)
	if err != nil {
		return err
	}

	return gpl.AddLayerToCobraCommand(cmd)
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
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			return nil
		}
		if err != context.Canceled {
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
