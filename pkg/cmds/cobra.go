package cmds

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
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
	parameters, err := GatherFlagsFromCobraCommand(cmd, description.Flags, false)
	if err != nil {
		return nil, err
	}

	arguments, err := GatherArguments(args, description.Arguments, false)
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
		parameters[k] = v
	}

	return parameters, nil
}

// AddArgumentsToCobraCommand adds the arguments (not the flags) of a CommandDescription to a cobra command
// as positional arguments.
// An optional argument cannot be followed by a required argument.
// Similarly, a list of arguments cannot be followed by any argument (since we otherwise wouldn't
// know how many belong to the list and where to do the cut off).
func AddArgumentsToCobraCommand(cmd *cobra.Command, arguments []*ParameterDefinition) error {
	minArgs := 0
	// -1 signifies unbounded
	maxArgs := 0
	hadOptional := false

	for _, argument := range arguments {
		if maxArgs == -1 {
			// already handling unbounded arguments
			return errors.Errorf("Cannot handle more than one unbounded argument, but found %s", argument.Name)
		}
		err := argument.CheckParameterDefaultValueValidity()
		if err != nil {
			return errors.Wrapf(err, "Invalid default value for argument %s", argument.Name)
		}

		if argument.Required {
			if hadOptional {
				return errors.Errorf("Cannot handle required argument %s after optional argument", argument.Name)
			}
			minArgs++
		} else {
			hadOptional = true
		}
		maxArgs++
		if argument.Type == ParameterTypeStringList || argument.Type == ParameterTypeIntegerList {
			maxArgs = -1
		}
	}

	cmd.Args = cobra.MinimumNArgs(minArgs)
	if maxArgs != -1 {
		cmd.Args = cobra.RangeArgs(minArgs, maxArgs)
	}

	return nil
}

// GatherArguments parses the positional arguments passed as a list of strings into a map of
// parsed values. If onlyProvided is true, then only arguments that are provided are returned
// (i.e. the default values are not included).
func GatherArguments(args []string, arguments []*ParameterDefinition, onlyProvided bool) (map[string]interface{}, error) {
	_ = args
	result := make(map[string]interface{})
	argsIdx := 0
	for _, argument := range arguments {
		if argsIdx >= len(args) {
			if argument.Required {
				return nil, errors.Errorf("Argument %s not found", argument.Name)
			} else {
				if argument.Default != nil && !onlyProvided {
					result[argument.Name] = argument.Default
				}
				continue
			}
		}

		v := []string{args[argsIdx]}

		if IsListParameter(argument.Type) {
			v = args[argsIdx:]
			argsIdx = len(args)
		} else {
			argsIdx++
		}
		i2, err := argument.ParseParameter(v)
		if err != nil {
			return nil, err
		}

		result[argument.Name] = i2
	}
	if argsIdx < len(args) {
		return nil, errors.Errorf("Too many arguments")
	}
	return result, nil
}

// AddFlagsToCobraCommand takes the parameters from a CommandDescription and converts them
// to cobra flags, before adding them to the Flags() of a the passed cobra command.
//
// # TODO(manuel, 2023-02-12) We need to handle arbitrary defaults here
//
// See https://github.com/go-go-golems/glazed/issues/132
//
// Currently, usage of this functions just passes the defaults encoded in
// the metadata YAML files (for glazed flags at least), but really we want
// to override this on a per command basis easily without having to necessarily
// copy or mutate the parameters loaded from yaml.
//
// One option would be to remove the defaults structs, and do the overloading
// by command with ParameterList manipulating functions, so that it is easy for the
// library user to override and further tweak the defaults.
//
// Currently, that behaviour is encoded in the AddFieldsFilterFlags function itself.
//
// What also needs to be considered is the bigger context that these declarative flags
// and arguments definitions are going to be used in a lot of different contexts,
// and might need to be overloaded and initialized in different ways.
//
// For example:
// - REST API
// - CLI
// - GRPC service
// - TUI bubbletea UI
// - Web UI
// - declarative config files
//
// --- 2023-02-12 - manuel
//
// I went with the following solution:
//
// One other option would be to pass this function a map with overloaded default,
// but while that feels easier and cleaner in the short term, I think it limits the
// concept of what it means for a library user to overload the defaults handling
// mechanism. This already becomes apparent in the FieldsFilterDefaults handling, where
// an empty list or a list containing "all" should be treated the same.
func AddFlagsToCobraCommand(flagSet *flag.FlagSet, flags []*ParameterDefinition) error {
	for _, parameter := range flags {
		err := parameter.CheckParameterDefaultValueValidity()
		if err != nil {
			return errors.Wrapf(err, "Invalid default value for argument %s", parameter.Name)
		}

		flagName := parameter.Name
		// replace _ with -
		flagName = strings.ReplaceAll(flagName, "_", "-")
		shortFlag := parameter.ShortFlag
		ok := false

		switch parameter.Type {
		case ParameterTypeStringListFromFile:
			fallthrough
		case ParameterTypeStringFromFile:
			fallthrough
		case ParameterTypeObjectFromFile:
			fallthrough
		case ParameterTypeObjectListFromFile:
			defaultValue := ""

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.String(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeString:
			defaultValue := ""

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(string)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a string: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.String(flagName, defaultValue, parameter.Help)
			}
		case ParameterTypeInteger:
			defaultValue := 0

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(int)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not an integer: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.IntP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.Int(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeFloat:
			defaultValue := 0.0

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(float64)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a float: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.Float64P(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.Float64(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeBool:
			defaultValue := false

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(bool)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a bool: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.BoolP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.Bool(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeDate:
			defaultValue := ""

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(string)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a string: %v", parameter.Name, parameter.Default)
				}

				parsedDate, err2 := ParseDate(defaultValue)
				if err2 != nil {
					return err2
				}
				_ = parsedDate
			}

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.String(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeStringList:
			var defaultValue []string

			if parameter.Default != nil {
				stringList, ok := parameter.Default.([]string)
				if !ok {
					defaultValue, ok := parameter.Default.([]interface{})
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, parameter.Default)
					}

					// convert to string list
					stringList, ok = helpers.CastList[string, interface{}](defaultValue)
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, parameter.Default)
					}
				}

				defaultValue = stringList
			}
			if err != nil {
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", parameter.Name, parameter.Default)
			}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.StringSlice(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeKeyValue:
			var defaultValue []string

			if parameter.Default != nil {
				stringMap, ok := parameter.Default.(map[string]string)
				if !ok {
					defaultValue, ok := parameter.Default.(map[string]interface{})
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, parameter.Default)
					}

					stringMap = make(map[string]string)
					for k, v := range defaultValue {
						stringMap[k] = fmt.Sprintf("%v", v)
					}
				}

				stringList := make([]string, 0)
				for k, v := range stringMap {
					// TODO(manuel, 2023-02-11) This is fixed to : but should be configurable
					// See https://github.com/go-go-golems/glazed/issues/129
					stringList = append(stringList, fmt.Sprintf("%s:%s", k, v))
				}

				defaultValue = stringList
			}
			if err != nil {
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", parameter.Name, parameter.Default)
			}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.StringSlice(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeIntegerList:
			var defaultValue []int
			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.([]int)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not an integer list: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.IntSliceP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.IntSlice(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeFloatList:
			var defaultValue []float64
			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.([]float64)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a float list: %v", parameter.Name, parameter.Default)
				}
			}
			if parameter.ShortFlag != "" {
				flagSet.Float64SliceP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.Float64Slice(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeChoice:
			defaultValue := ""

			if parameter.Default != nil {
				defaultValue, ok = parameter.Default.(string)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a string: %v", parameter.Name, parameter.Default)
				}
			}

			choiceString := strings.Join(parameter.Choices, ",")

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, fmt.Sprintf("%s (%s)", parameter.Help, choiceString))
			} else {
				flagSet.String(flagName, defaultValue, fmt.Sprintf("%s (%s)", parameter.Help, choiceString))
			}

		default:
			panic(fmt.Sprintf("Unknown parameter type %s", parameter.Type))
		}
	}

	return nil
}

// GatherFlagsFromCobraCommand gathers the flags from the cobra command, and parses them according
// to the parameter description passed in params. The result is a map of parameter
// names to parsed values. If onlyProvided is true, only parameters that are provided
// by the user are returned (i.e. not the default values).
// If a parameter cannot be parsed correctly, or is missing even though it is not optional,
// an error is returned.
func GatherFlagsFromCobraCommand(
	cmd *cobra.Command,
	params []*ParameterDefinition,
	onlyProvided bool,
) (map[string]interface{}, error) {
	parameters := map[string]interface{}{}

	for _, parameter := range params {
		// check if the flag is set
		flagName := parameter.Name
		flagName = strings.ReplaceAll(flagName, "_", "-")

		if !cmd.Flags().Changed(flagName) {
			if parameter.Required {
				return nil, errors.Errorf("ParameterDefinition %s is required", parameter.Name)
			}

			if parameter.Default == nil {
				continue
			}

			if onlyProvided {
				continue
			}
		}

		switch parameter.Type {
		case ParameterTypeObjectFromFile:
			fallthrough
		case ParameterTypeObjectListFromFile:
			fallthrough
		case ParameterTypeStringFromFile:
			fallthrough
		case ParameterTypeStringListFromFile:
			fallthrough
		case ParameterTypeString:
			fallthrough
		case ParameterTypeDate:
			fallthrough
		case ParameterTypeChoice:
			v, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return nil, err
			}
			v2, err := parameter.ParseParameter([]string{v})
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v2

		case ParameterTypeFloat:
			v, err := cmd.Flags().GetFloat64(flagName)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v

		case ParameterTypeInteger:
			v, err := cmd.Flags().GetInt(flagName)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v

		case ParameterTypeBool:
			v, err := cmd.Flags().GetBool(flagName)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v

		case ParameterTypeStringList:
			fallthrough
		case ParameterTypeKeyValue:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return nil, err
			}
			v2, err := parameter.ParseParameter(v)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v2

		case ParameterTypeIntegerList:
			v, err := cmd.Flags().GetIntSlice(flagName)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v

		case ParameterTypeFloatList:
			v, err := cmd.Flags().GetFloat64Slice(flagName)
			if err != nil {
				return nil, err
			}
			parameters[parameter.Name] = v
		}
	}
	return parameters, nil
}

func NewCobraCommand(s CobraCommand) (*cobra.Command, error) {
	description := s.Description()
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}

	err := AddFlagsToCobraCommand(cmd.Flags(), description.Flags)
	if err != nil {
		return nil, err
	}

	err = AddArgumentsToCobraCommand(cmd, description.Arguments)
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
