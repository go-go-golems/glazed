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
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// GatherParametersFromCobraCommand takes a cobra command, an argument list as well as a description
// of the command, and returns a list of parsed parameters as a
// hashmap. It does so by parsing both the flags and the positional arguments.
//
// TODO(manuel, 2021-02-04) This function should return how a parameter was set
// This would match warg's behaviour. It would also make it possible for us to
// record if it was set from ENV, from the query, from default, frmo which config file,
// etc...
//
// See https://github.com/go-go-golems/glazed/issues/172
func GatherParametersFromCobraCommand(
	cmd *cobra.Command,
	description *cmds.CommandDescription,
	args []string,
) (map[string]interface{}, error) {
	ps, err := parameters.GatherFlagsFromCobraCommand(cmd, description.Flags, false, "")
	if err != nil {
		return nil, err
	}

	arguments, err := parameters.GatherArguments(args, description.Arguments, false)
	if err != nil {
		return nil, err
	}

	// merge parameters and arguments
	// arguments take precedence over parameters
	for p := arguments.Oldest(); p != nil; p = p.Next() {
		k, v := p.Key, p.Value
		ps[k] = v
	}

	return ps, nil
}

type CobraRunFunc func(ctx context.Context, parsedLayers map[string]*layers.ParsedParameterLayer, ps map[string]interface{}) error

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

	cobraParser, err := NewCobraParserFromCommandDescription(description)
	if err != nil {
		return nil, err
	}

	cmd := cobraParser.Cmd

	cmd.Flags().String("create-command", "", "Create a new command for the query, with the defaults updated")
	cmd.Flags().String("create-alias", "", "Create a CLI alias for the query")
	cmd.Flags().String("create-cliopatra", "", "Print the CLIopatra YAML for the command")
	cmd.Flags().Bool("print-yaml", false, "Print the command's YAML")
	cmd.Flags().String("load-parameters-from-json", "", "Load the command's flags from JSON")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		loadParametersFromJSON, err := cmd.Flags().GetString("load-parameters-from-json")
		if err != nil {
			cobra.CheckErr(err)
			os.Exit(1)
		}

		parsedLayers := map[string]*layers.ParsedParameterLayer{}
		ps := map[string]interface{}{}

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

			// we now need to map the individual values in the JSON to the parsed layers as well
			for _, layer := range description.Layers {
				layerPs := map[string]interface{}{}

				for _, parameter := range layer.GetParameterDefinitions() {
					value, ok := result[parameter.Name]
					if ok {
						err := parameter.CheckValueValidity(value)
						if err != nil {
							err = errors.Wrapf(err, "invalid value for parameter %s: %s", parameter.Name, err)
							cobra.CheckErr(err)
						}
						layerPs[parameter.Name] = value
					} else {
						if parameter.Default != nil {
							layerPs[parameter.Name] = parameter.Default
						}

						if parameter.Required {
							err := errors.Errorf("missing required parameter %s", parameter.Name)
							cobra.CheckErr(err)
						}
					}

					ps[parameter.Name] = layerPs[parameter.Name]
				}

				parsedLayer := &layers.ParsedParameterLayer{
					Layer:      layer,
					Parameters: layerPs,
				}
				parsedLayers[layer.GetName()] = parsedLayer
			}
		} else {
			parsedLayers, ps, err = cobraParser.Parse(args)
			// show help if there is an error
			if err != nil {
				fmt.Println(err)
				err := cmd.Help()
				cobra.CheckErr(err)
				os.Exit(1)
			}
		}

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
				ps,
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

		createNewCommand, _ := cmd.Flags().GetString("create-command")
		// TODO(manuel, 2023-02-26) This only outputs the command description, not the actual command
		// This is already helpful, but is really just half the story. To make this work
		// generically, CreateNewCommand() should be part of the interface.
		//
		// See https://github.com/go-go-golems/glazed/issues/170
		if createNewCommand != "" {
			clonedArguments := []*parameters.ParameterDefinition{}
			for _, arg := range description.Arguments {
				newArg := arg.Copy()
				v, ok := ps[arg.Name]
				if ok {
					newArg.Default = v
				}
				clonedArguments = append(clonedArguments, newArg)
			}
			clonedFlags := []*parameters.ParameterDefinition{}
			for _, flag := range description.Flags {
				newFlag := flag.Copy()
				v, ok := ps[flag.Name]
				if ok {
					newFlag.Default = v
				}
				clonedFlags = append(clonedFlags, newFlag)
			}

			cmd := &cmds.CommandDescription{
				Name:      createNewCommand,
				Short:     description.Short,
				Long:      description.Long,
				Arguments: clonedArguments,
				Flags:     clonedFlags,
			}

			// encode as yaml
			sb := strings.Builder{}
			encoder := yaml.NewEncoder(&sb)
			err = encoder.Encode(cmd)
			cobra.CheckErr(err)

			fmt.Println(sb.String())
			os.Exit(0)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			err := helpers.CancelOnSignal(ctx, os.Interrupt, cancel)
			if err != nil && err != context.Canceled {
				fmt.Println(err)
			}
		}()

		err = run(ctx, parsedLayers, ps)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			os.Exit(0)
		}

	}

	return cmd, nil
}

func BuildCobraCommandFromBareCommand(c cmds.BareCommand) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommandAndFunc(c, func(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
	) error {
		err := c.Run(ctx, parsedLayers, ps)
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
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
	) error {
		err := s.RunIntoWriter(ctx, parsedLayers, ps, os.Stdout)
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
	cmd.Short = fmt.Sprintf("Alias for %s", alias.AliasedCommand.Description().Name)

	argumentDefinitions := alias.AliasedCommand.Description().Arguments
	minArgs := 0
	provided, err := parameters.GatherArguments(alias.Arguments, argumentDefinitions, true)

	for _, argDef := range argumentDefinitions {
		_, ok := provided.Get(argDef.Name)
		if argDef.Required && !ok {
			minArgs++
		}
	}

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
// set of ParsedParameterLayer and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered layers
// from the description.
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

	err := parameters.AddFlagsToCobraCommand(cmd.Flags(), description.Flags, "")
	if err != nil {
		return nil, err
	}

	for _, layer := range description.Layers {
		// check that layer is a CobraParameterLayer
		// if not, return an error
		cobraLayer, ok := layer.(CobraParameterLayer)
		if !ok {
			return nil, fmt.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		err := cobraLayer.AddFlagsToCobraCommand(cmd)
		if err != nil {
			return nil, err
		}
	}

	err = parameters.AddArgumentsToCobraCommand(cmd, description.Arguments)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

type CobraParameterLayer interface {
	// AddFlagsToCobraCommand adds all the flags defined in this layer to the given cobra command.
	//
	// NOTE(manuel, 2023-02-27) This can be moved to use that ParameterLayerParser API
	// As I'm working out what it means to parse layers and use it to fill structs,
	// and how defaults should be registered, it makes sense to move this out.
	// Further more, defaults should probably be managed in the layer entirely, and
	// thus not be shown in the interface here.
	//
	// Do we want to keep the parsers in the layer itself, so that when a command is registered,
	// it gets registered here? Or should the parsers and registerers be outside,
	// and generic enough to be able to process all the layers of a command without
	// the command framework knowing about it. This seems to make more sense.
	AddFlagsToCobraCommand(cmd *cobra.Command) error
	ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error)
}

func (c *CobraParser) Parse(args []string) (map[string]*layers.ParsedParameterLayer, map[string]interface{}, error) {
	parsedLayers := map[string]*layers.ParsedParameterLayer{}
	ps := map[string]interface{}{}

	for _, layer := range c.description.Layers {
		cobraLayer, ok := layer.(CobraParameterLayer)
		if !ok {
			return nil, nil, fmt.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		// parse the flags from commands
		ps, err := cobraLayer.ParseFlagsFromCobraCommand(c.Cmd)
		if err != nil {
			return nil, nil, err
		}

		parsedLayer := &layers.ParsedParameterLayer{Parameters: ps, Layer: layer}
		parsedLayers[layer.GetSlug()] = parsedLayer

		// TODO(manuel, 2021-02-04) This is a legacy conserving hack since all commands use a map for now
		//
		// the question is, is this even necessary? It is for a generic
		// map[string]interface{} based approach, but in the future we might
		// want to just pass in the parsed layers downstream
		//
		// See https://github.com/go-go-golems/glazed/issues/173
		for k, v := range parsedLayer.Parameters {
			ps[k] = v
		}
	}

	// This can be used to override layer arguments, not sure how useful
	// that is or if it's something we want to actually forbid.
	//
	// This might not even be possible in the first place, because it would mean that
	// we used cobra to register the same flag twice.
	ps_, err := GatherParametersFromCobraCommand(c.Cmd, c.description, args)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range ps_ {
		ps[k] = v
	}

	return parsedLayers, ps, nil
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

	ps, err := gpl.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	gp, err := settings.SetupTableProcessor(ps)
	cobra.CheckErr(err)

	of, err := settings.SetupProcessorOutput(gp, ps, os.Stdout)
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

	return gpl.AddFlagsToCobraCommand(cmd)
}

func BuildCobraCommandFromGlazeCommand(cmd_ cmds.GlazeCommand) (*cobra.Command, error) {
	cmd, err := BuildCobraCommandFromCommandAndFunc(cmd_, func(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
	) error {
		gp, err := settings.SetupTableProcessor(ps)
		cobra.CheckErr(err)

		_, err = settings.SetupProcessorOutput(gp, ps, os.Stdout)
		cobra.CheckErr(err)

		err = cmd_.Run(ctx, parsedLayers, ps, gp)
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
