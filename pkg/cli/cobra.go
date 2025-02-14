package cli

import (
	"context"
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
		return nil, err
	}
	err = cobraParser.AddToCobraCommand(cmd)
	if err != nil {
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

		// Handle both full and minimal command settings

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
				// TODO(manuel, 2024-01-09) Implement schema printing
				log.Warn().Msg("Schema printing not yet implemented")
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
