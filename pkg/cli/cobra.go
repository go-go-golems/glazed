package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli/cliopatra"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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

func BuildCobraCommandFromCommandAndFunc(s cmds.Command, run CobraRunFunc) (*cobra.Command, error) {
	description := s.Description()

	// check if we need to add the glazedCommandLayer
	addGlazedCommandLayer := true
	description.Layers.ForEach(func(_ string, layer layers.ParameterLayer) {
		if layer.GetSlug() == "glazed-command" {
			addGlazedCommandLayer = false
		}
	})

	layers_ := description.Layers.Clone()

	// TODO(manuel, 2023-12-21) Not sure if this cobra specific location is the best place to add the glazed-command layer
	if addGlazedCommandLayer {
		glazedCommandLayer, err := layers.NewParameterLayer(
			"glazed-command",
			"General purpose Command options",
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition(
					"create-command",
					parameters.ParameterTypeString,
					parameters.WithHelp("Create a new command for the query, with the defaults updated"),
				),
				parameters.NewParameterDefinition(
					"create-alias",
					parameters.ParameterTypeString,
					parameters.WithHelp("Create a CLI alias for the query"),
				),
				parameters.NewParameterDefinition(
					"create-cliopatra",
					parameters.ParameterTypeString,
					parameters.WithHelp("Print the CLIopatra YAML for the command"),
				),
				parameters.NewParameterDefinition(
					"print-yaml",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print the command's YAML"),
				),
				parameters.NewParameterDefinition(
					"load-parameters-from-json",
					parameters.ParameterTypeString,
					parameters.WithHelp("Load the command's flags from JSON"),
				),
			),
		)
		if err != nil {
			return nil, err
		}

		// NOTE(manuel, 2023-12-20) Should we clone the layer list here?
		layers_.Set(glazedCommandLayer.GetSlug(), glazedCommandLayer)
	}

	description.Layers = layers_

	cobraParser, err := NewCobraParserFromCommandDescription(description)
	if err != nil {
		return nil, err
	}

	cmd := cobraParser.Cmd

	cmd.Run = func(cmd *cobra.Command, args []string) {
		loadParametersFromJSON, err := cmd.Flags().GetString("load-parameters-from-json")
		if err != nil {
			cobra.CheckErr(err)
			os.Exit(1)
		}

		var parsedLayers *layers.ParsedLayers

		if loadParametersFromJSON != "" {
			result := map[string]interface{}{}
			bytes, err := os.ReadFile(loadParametersFromJSON)
			if err != nil {
				cobra.CheckErr(err)
			}
			err = json.Unmarshal(bytes, &result)
			if err != nil {
				cobra.CheckErr(err)
			}

			parsedLayers, err = cmds.ParseCommandFromMap(description, result)
			if err != nil {
				cobra.CheckErr(err)
			}

			// Need to update the parsedLayers from command line flags too...

			err = parsedLayers.ForEachE(func(_ string, layer *layers.ParsedLayer) error {
				ps_, err := parameters.GatherFlagsFromCobraCommand(
					cmd,
					layer.Layer.GetParameterDefinitions(),
					true, true,
					layer.Layer.GetPrefix())
				if err != nil {
					return err
				}

				layer.Parameters.Merge(ps_)
				return nil
			})
			cobra.CheckErr(err)
		} else {
			parsedLayers, err = cobraParser.Parse()
			// show help if there is an error
			if err != nil {
				fmt.Println(err)
				err := cmd.Help()
				cobra.CheckErr(err)
				os.Exit(1)
			}
		}

		layer := parsedLayers.GetDefaultParameterLayer()

		arguments, err := parameters.GatherArguments(args, description.GetDefaultArguments(), true, true)
		if err != nil {
			cobra.CheckErr(err)
		}
		layer.Parameters.Merge(arguments)

		printYAML, err := cmd.Flags().GetBool("print-yaml")
		cobra.CheckErr(err)

		if printYAML {
			err = s.ToYAML(os.Stdout)
			cobra.CheckErr(err)
			return
		}

		createCliopatra, err := cmd.Flags().GetString("create-cliopatra")
		cobra.CheckErr(err)

		if createCliopatra != "" {
			verbs := GetVerbsFromCobraCommand(cmd)
			if len(verbs) == 0 {
				cobra.CheckErr(errors.New("could not get verbs from cobra command"))
			}
			p := cliopatra.NewProgramFromCapture(
				s.Description(),
				parsedLayers,
				cliopatra.WithVerbs(verbs[1:]...),
				cliopatra.WithName(createCliopatra),
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

		createCliAlias, err := cmd.Flags().GetString("create-alias")
		cobra.CheckErr(err)
		if createCliAlias != "" {
			alias := &alias.CommandAlias{
				Name:      createCliAlias,
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

		//createNewCommand, _ := cmd.Parameters().GetString("create-command")
		// TODO(manuel, 2023-02-26) This only outputs the command description, not the actual command
		// This is already helpful, but is really just half the story. To make this work
		// generically, CreateNewCommand() should be part of the interface.
		//
		// See https://github.com/go-go-golems/glazed/issues/170
		//
		// NOTE(manuel, 2023-12-21) Disabling this for now while I move flagand argument handling
		// over to be done by the default layer'
		//
		//if createNewCommand != "" {
		//	clonedArguments := []*parameters.ParameterDefinition{}
		//	for _, arg := range description.Arguments {
		//		newArg := arg.Copy()
		//		v, ok := ps[arg.Name]
		//		if ok {
		//			newArg.Default = v
		//		}
		//		clonedArguments = append(clonedArguments, newArg)
		//	}
		//	clonedFlags := []*parameters.ParameterDefinition{}
		//	for _, flag := range description.Parameters {
		//		newFlag := flag.Copy()
		//		v, ok := ps[flag.Name]
		//		if ok {
		//			newFlag.Default = v
		//		}
		//		clonedFlags = append(clonedFlags, newFlag)
		//	}
		//
		//	cmd := &cmds.CommandDescription{
		//		Name:      createNewCommand,
		//		Short:     description.Short,
		//		Long:      description.Long,
		//		Arguments: clonedArguments,
		//		Parameters:     clonedFlags,
		//	}
		//
		//	// encode as yaml
		//	sb := strings.Builder{}
		//	encoder := yaml.NewEncoder(&sb)
		//	err = encoder.Encode(cmd)
		//	cobra.CheckErr(err)
		//
		//	fmt.Println(sb.String())
		//	os.Exit(0)
		//}

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

func BuildCobraCommandFromBareCommand(c cmds.BareCommand) (*cobra.Command, error) {
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
	})

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func BuildCobraCommandFromWriterCommand(s cmds.WriterCommand) (*cobra.Command, error) {
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
	})

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func BuildCobraCommandAlias(alias *alias.CommandAlias) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommand(alias.AliasedCommand)
	if err != nil {
		return nil, err
	}

	origRun := cmd.Run

	cmd.Use = alias.Name
	description := alias.AliasedCommand.Description()
	cmd.Short = fmt.Sprintf("Alias for %s", description.Name)

	minArgs := 0
	argumentDefinitions := description.GetDefaultArguments()
	provided, err := parameters.GatherArguments(alias.Arguments, argumentDefinitions, true, true)
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

func BuildCobraCommandFromCommand(command cmds.Command) (*cobra.Command, error) {
	var cobraCommand *cobra.Command
	var err error
	switch c := command.(type) {
	case cmds.BareCommand:
		cobraCommand, err = BuildCobraCommandFromBareCommand(c)

	case cmds.WriterCommand:
		cobraCommand, err = BuildCobraCommandFromWriterCommand(c)

	case cmds.GlazeCommand:
		cobraCommand, err = BuildCobraCommandFromGlazeCommand(c)

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
) error {
	commandsByName := map[string]cmds.Command{}

	for _, command := range commands {
		// find the proper subcommand, or create if it doesn't exist
		description := command.Description()
		parentCmd := findOrCreateParentCommand(rootCmd, description.Parents)

		cobraCommand, err := BuildCobraCommandFromCommand(command)
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
		cobraCommand, err := BuildCobraCommandAlias(alias)
		if err != nil {
			return err
		}
		parentCmd.AddCommand(cobraCommand)
	}

	return nil
}

// CobraParser takes a CommandDescription, and hooks it up to a cobra command.
// It can then be used to parse the cobra flags and arguments back into a
// set of ParsedLayer and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run* method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered layers
// from the description.
//
// NOTE(manuel, 2023-09-18) Now that I've removed the parserFunc, this feels a bit unnecessary too
// Or it could be something that is actually an interface on top of Command, like a CobraCommand.
type CobraParser struct {
	Cmd         *cobra.Command
	description *cmds.CommandDescription
}

func NewCobraParserFromCommandDescription(description *cmds.CommandDescription) (*CobraParser, error) {
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}

	ret := &CobraParser{
		Cmd:         cmd,
		description: description,
	}

	err := description.Layers.ForEachE(func(_ string, layer layers.ParameterLayer) error {
		// check that layer is a CobraParameterLayer
		// if not, return an error
		cobraLayer, ok := layer.(layers.CobraParameterLayer)
		if !ok {
			return fmt.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		err := cobraLayer.AddLayerToCobraCommand(cmd)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func ParseFlagsFromViperAndCobraCommand(cmd *cobra.Command, d layers.ParameterLayer) (*parameters.ParsedParameters, error) {
	// actually hijack and load everything from viper instead of cobra...
	parameterDefinitions := d.GetParameterDefinitions()
	prefix := d.GetPrefix()

	ps, err := parameters.GatherFlagsFromViper(parameterDefinitions, false, prefix)
	if err != nil {
		return nil, err
	}

	// now load from flag overrides
	ps2, err := parameters.GatherFlagsFromCobraCommand(cmd, parameterDefinitions, true, false, prefix)
	if err != nil {
		return nil, err
	}
	ps.Merge(ps2)

	return ps, nil
}

func (c *CobraParser) Parse() (*layers.ParsedLayers, error) {
	parsedLayers := layers.NewParsedLayers()

	err := c.description.Layers.ForEachE(func(_ string, layer layers.ParameterLayer) error {
		cobraLayer, ok := layer.(layers.CobraParameterLayer)
		if !ok {
			return fmt.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		// parse the flags from commands
		parsedLayer, err := cobraLayer.ParseLayerFromCobraCommand(c.Cmd)
		if err != nil {
			return err
		}

		parsedLayers.Set(layer.GetSlug(), parsedLayer)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}

// CreateGlazedProcessorFromCobra is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
//
// If you are more serious about using glazed, consider using the `cmds.GlazeCommand` and `parameters.ParameterDefinition`
// abstraction to define your CLI applications, which allows you to use layers and other nice features
// of the glazed ecosystem.
//
// If so, use SetupTableProcessor instead, and create a proper glazed.GlazeCommand for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	gpl, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	parsedLayer, err := gpl.ParseLayerFromCobraCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	var glazedLayer *layers.ParsedLayer

	//glazedLayer, ok := parsedLayers["glazed"]
	//if !ok {
	//	return nil, errors.New("glazed layer not found")
	//}

	gp, err := settings.SetupTableProcessor(glazedLayer)
	cobra.CheckErr(err)

	of, err := settings.SetupProcessorOutput(gp, parsedLayer, os.Stdout)
	cobra.CheckErr(err)

	return gp, of, nil
}

// AddGlazedProcessorFlagsToCobraCommand is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
func AddGlazedProcessorFlagsToCobraCommand(cmd *cobra.Command, options ...settings.GlazeParameterLayerOption) error {
	gpl, err := settings.NewGlazedParameterLayers(options...)
	if err != nil {
		return err
	}

	return gpl.AddLayerToCobraCommand(cmd)
}

func BuildCobraCommandFromGlazeCommand(cmd_ cmds.GlazeCommand) (*cobra.Command, error) {
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
	})

	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func ParseLayersFromCobraCommand(cmd *cobra.Command, layers_ []layers.CobraParameterLayer) (
	*layers.ParsedLayers,
	error,
) {
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
