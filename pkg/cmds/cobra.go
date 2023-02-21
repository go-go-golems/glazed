package cmds

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// CobraCommand is a subset of Command than can be registered as a
// cobra.Command by using BuildCobraCommand, and can then be executed
// when cobra runs it through RunFromCobra, passing in the full cobra object
// in case the user wants to overload anything.
type CobraCommand interface {
	Command
	RunFromCobra(cmd *cobra.Command, args []string) error
	BuildCobraCommand() (*cobra.Command, error)
}

// GatherParametersFromCobraCommand takes a cobra command, an argument list as well as a description
// of the sqleton command arguments, and returns a list of parsed parameters as a
// hashmap. It does so by parsing both the flags and the positional arguments.
func GatherParametersFromCobraCommand(
	cmd *cobra.Command,
	description *CommandDescription,
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

	createAlias, err := cmd.Flags().GetString("create-alias")
	if err != nil {
		return nil, err
	}
	if createAlias != "" {
		alias := &CommandAlias{
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

	// merge parameters and arguments
	// arguments take precedence over parameters
	for k, v := range arguments {
		ps[k] = v
	}

	return ps, nil
}

// TODO(manuel, 2023-02-21) This is probably possible to do with just a plain command
// If we can gather all the necessary parameter definitions for the command in one go
// we can make registering to cobra just as generic as registering for other interfaces
//
// See #150
func NewCobraCommand(s CobraCommand) (*cobra.Command, error) {
	description := s.Description()
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}

	err := parameters.AddFlagsToCobraCommand(cmd.Flags(), description.Flags)
	if err != nil {
		return nil, err
	}

	err = parameters.AddArgumentsToCobraCommand(cmd, description.Arguments)
	if err != nil {
		return nil, err
	}

	cmd.Flags().String("create-alias", "", "Create an alias for the query")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := s.RunFromCobra(cmd, args)
		cobra.CheckErr(err)
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

func AddCommandsToRootCommand(rootCmd *cobra.Command, commands []CobraCommand, aliases []*CommandAlias) error {
	commandsByName := map[string]Command{}

	for _, command := range commands {
		// find the proper subcommand, or create if it doesn't exist
		description := command.Description()
		parentCmd := findOrCreateParentCommand(rootCmd, description.Parents)
		cobraCommand, err := command.BuildCobraCommand()
		if err != nil {
			return err
		}

		command2 := command
		cobraCommand.Run = func(cmd *cobra.Command, args []string) {
			err := command2.RunFromCobra(cmd, args)
			cobra.CheckErr(err)
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
		cobraCommand, err := alias.BuildCobraCommand()
		if err != nil {
			return err
		}
		alias2 := alias
		cobraCommand.Run = func(cmd *cobra.Command, args []string) {
			for flagName, flagValue := range alias2.Flags {
				if !cmd.Flags().Changed(flagName) {
					err = cmd.Flags().Set(flagName, flagValue)
					cobra.CheckErr(err)
				}
			}

			// TODO(2022-12-22, manuel) This is not right because the args count is checked earlier, but when,
			// and how can i override it
			//
			// NOTE(2023-02-07, manuel) I think the above refers to the fact that an alias
			// should be able to override arguments.
			if len(args) == 0 {
				args = alias2.Arguments
			}
			err = alias2.RunFromCobra(cmd, args)
			cobra.CheckErr(err)
		}
		parentCmd.AddCommand(cobraCommand)
	}

	return nil
}
