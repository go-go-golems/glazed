package cli

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers"
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
	ps, err := parameters.GatherFlagsFromCobraCommand(cmd, description.Flags, false)
	if err != nil {
		return nil, err
	}

	arguments, err := parameters.GatherArguments(args, description.Arguments, false)
	if err != nil {
		return nil, err
	}

	layers := description.Layers
	for _, layer := range layers {
		layerFlags, err := layer.ParseFlagsFromCobraCommand(cmd)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse flags for layer")
		}

		for k, v := range layerFlags {
			ps[k] = v
		}
	}

	// merge parameters and arguments
	// arguments take precedence over parameters
	for k, v := range arguments {
		ps[k] = v
	}

	createAlias, err := cmd.Flags().GetString("create-alias")
	if err != nil {
		return nil, err
	}
	if createAlias != "" {
		alias := &cmds.CommandAlias{
			Name:      createAlias,
			AliasFor:  description.Name,
			Arguments: args,
			Flags:     map[string]string{},
		}

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			if flag.Name != "create-alias" {
				alias.Flags[flag.Name] = flag.Value.String()
			}
		})

		// marshal alias to yaml
		sb := strings.Builder{}
		encoder := yaml.NewEncoder(&sb)
		err = encoder.Encode(alias)
		if err != nil {
			return nil, err
		}

		fmt.Println(sb.String())
		os.Exit(0)
	}

	return ps, nil
}

func BuildCobraCommand(s cmds.Command) (*cobra.Command, error) {
	description := s.Description()
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}

	err := parameters.AddFlagsToCobraCommand(cmd.Flags(), description.Flags)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add flags for command '%s'", description.Name)
	}

	for _, layer := range description.Layers {
		err = layer.AddFlagsToCobraCommand(cmd, nil)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to add flags for command '%s' layer '%s'",
				description.Name,
				layer.GetSlug(),
			)
		}
	}

	err = parameters.AddArgumentsToCobraCommand(cmd, description.Arguments)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add arguments for command '%s'", description.Name)
	}

	cmd.Flags().String("create-command", "", "Create a new command for the query, with the defaults updated")
	cmd.Flags().String("create-alias", "", "Create a CLI alias for the query")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		ps, err := GatherParametersFromCobraCommand(cmd, description, args)
		cobra.CheckErr(err)

		createCliAlias, err := cmd.Flags().GetString("create-alias")
		cobra.CheckErr(err)
		if createCliAlias != "" {
			alias := &cmds.CommandAlias{
				Name:      createCliAlias,
				AliasFor:  description.Name,
				Arguments: args,
				Flags:     map[string]string{},
			}

			cmd.Flags().Visit(func(flag *pflag.Flag) {
				if flag.Name != "create-alias" {
					fmt.Printf("flag: %s, value: %s, type: %s\n", flag.Name, flag.Value.String(), flag.Value.Type())
					alias.Flags[flag.Name] = flag.Value.String()
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

		gp, of, err := SetupProcessor(ps)
		cobra.CheckErr(err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			err := helpers.CancelOnSignal(ctx, os.Interrupt, cancel)
			if err != nil {
				fmt.Println(err)
			}
		}()

		err = s.Run(ctx, ps, gp)
		cobra.CheckErr(err)

		s, err := of.Output()
		cobra.CheckErr(err)

		fmt.Println(s)
	}

	return cmd, nil
}

func BuildCobraCommandAlias(alias *cmds.CommandAlias) (*cobra.Command, error) {
	s := alias.AliasedCommand

	cmd, err := BuildCobraCommand(s)
	if err != nil {
		return nil, err
	}

	origRun := cmd.Run

	cmd.Use = alias.Name
	cmd.Short = fmt.Sprintf("Alias for %s", s.Description().Name)

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

func AddCommandsToRootCommand(rootCmd *cobra.Command, commands []cmds.Command, aliases []*cmds.CommandAlias) error {
	commandsByName := map[string]cmds.Command{}

	for _, command := range commands {
		// find the proper subcommand, or create if it doesn't exist
		description := command.Description()
		parentCmd := findOrCreateParentCommand(rootCmd, description.Parents)
		cobraCommand, err := BuildCobraCommand(command)
		if err != nil {
			return err
		}

		parentCmd.AddCommand(cobraCommand)

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

// CreateGlazedProcessorFromCobra is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
//
// If you are more serious about using glazed, consider using the `cmds.Command` and `parameters.ParameterDefinition`
// abstraction to define your CLI applications, which allows you to use layers and other nice features
// of the glazed ecosystem.
//
// If so, use SetupProcessor instead, and create a proper glazed.Command for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (
	*cmds.GlazeProcessor,
	formatters.OutputFormatter,
	error,
) {
	gpl, err := NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	ps, err := gpl.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	return SetupProcessor(ps)
}

func AddGlazedProcessorFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error {
	gpl, err := NewGlazedParameterLayers()
	if err != nil {
		return err
	}

	return gpl.AddFlagsToCobraCommand(cmd, defaults)
}
