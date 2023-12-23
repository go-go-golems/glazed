package parameters

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addArgumentsToCobraCommand adds each ParameterDefinition from `arguments` as positional arguments to the provided `cmd` cobra command.
// Argument ordering, optionality, and multiplicity constraints are respected:
//   - Required arguments (argument.Required == true) should come before the optional.
//   - Only one list argument (either ParameterTypeStringList or ParameterTypeIntegerList) is allowed and it should be the last one.
//
// Any violation of these conditions yields an error.
// This function processes each argument, checks their default values for validity, which if invalid,
// triggers an error return.
//
// It computes the minimum and maximum number of arguments the command can take based on the required, optional,
// and list arguments.
// If everything is successful, it assigns an argument validator (either MinimumNArgs or RangeArgs)
// to the cobra command's Args attribute.
func addArgumentsToCobraCommand(cmd *cobra.Command, arguments *ParameterDefinitions) error {
	minArgs := 0
	// -1 signifies unbounded
	maxArgs := 0
	hadOptional := false

	if arguments.Len() == 0 {
		return nil
	}

	err := arguments.ForEachE(func(argument *ParameterDefinition) error {
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
		if IsListParameter(argument.Type) {
			maxArgs = -1
		}

		return nil
	})

	if err != nil {
		return err
	}

	cmd.Args = cobra.MinimumNArgs(minArgs)
	if maxArgs != -1 {
		cmd.Args = cobra.RangeArgs(minArgs, maxArgs)
	}

	cmd.Use = GenerateUseString(cmd, arguments)

	return nil
}

// GenerateUseString creates a string representation of the 'Use' field for a given cobra command and a list of parameter definitions. The first word of the existing 'Use' field is treated as the verb for the command.
// The resulting string briefly describes how to use the command respecting the following conventions:
//   - Required parameters are enclosed in '<>'.
//   - Optional parameters are enclosed in '[]'.
//   - Optional parameters that accept multiple input (ParameterTypeStringList or ParameterTypeIntegerList) are followed by '...'.
//   - If a parameter has a default value, it is specified after parameter name like 'parameter (default: value)'.
//
// For example:
//   - If there is a required parameter 'name', and an optional parameter 'age' with a default value of '30', the resulting string will be: 'verb <name> [age (default: 30)]'.
//   - If there is a required parameter 'name', and an optional parameter 'colors' of type ParameterTypeStringList, the resulting Use string will be: 'verb <name> [colors...]'
func GenerateUseString(cmd *cobra.Command, arguments *ParameterDefinitions) string {
	fields := strings.Fields(cmd.Use)
	if len(fields) == 0 {
		return ""
	}
	verb := fields[0]
	useStr := verb
	var defaultValueStr string

	arguments.ForEach(func(arg *ParameterDefinition) {
		defaultValueStr = ""
		if arg.Default != nil {
			defaultValueStr = fmt.Sprintf(" (default: %v)", arg.Default)
		}
		left, right := "[", "]"
		if arg.Required {
			left, right = "<", ">"
		}
		if arg.Type == ParameterTypeStringList || arg.Type == ParameterTypeIntegerList {
			useStr += " " + left + arg.Name + "..." + defaultValueStr + right
		} else {
			useStr += " " + left + arg.Name + defaultValueStr + right
		}
	})

	return useStr
}

// GatherArguments parses positional string arguments into a map of values.
//
// It takes the command-line arguments, the expected argument definitions,
// and a boolean indicating whether to only include explicitly provided arguments,
// not the default values of missing arguments.
//
// Only the last parameter definitions can be a list parameter type.
//
// Required arguments missing from the input will result in an error.
// Arguments with default values can be included based on the onlyProvided flag.
//
// The result is a map with argument names as keys and parsed values.
// Argument order is maintained.
//
// Any extra arguments not defined will result in an error.
// Parsing errors for individual arguments will also return errors.
func GatherArguments(
	args []string,
	arguments *ParameterDefinitions,
	onlyProvided bool,
	ignoreRequired bool,
) (*ParsedParameters, error) {
	_ = args
	result := NewParsedParameters()
	argsIdx := 0
	for v := arguments.Oldest(); v != nil; v = v.Next() {
		argument := v.Value
		p := &ParsedParameter{
			ParameterDefinition: argument,
		}

		if argsIdx >= len(args) {
			if argument.Required {
				if ignoreRequired {
					continue
				}
				return nil, fmt.Errorf("Argument %s not found", argument.Name)
			} else {
				if argument.Default != nil && !onlyProvided {
					p.SetWithSource("default", argument.Default)
					result.Set(argument.Name, p)
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

		result.Set(argument.Name, i2)
	}
	if argsIdx < len(args) {
		return nil, fmt.Errorf("Too many arguments")
	}
	return result, nil
}

// AddParametersToCobraCommand takes the parameters from a CommandDescription and converts them
// to cobra flags, before adding them to the Parameters() of a the passed cobra command.
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
func AddParametersToCobraCommand(
	cmd *cobra.Command,
	pds *ParameterDefinitions,
	prefix string,
) error {
	flagSet := cmd.Flags()

	err := addArgumentsToCobraCommand(cmd, pds.GetArguments())
	if err != nil {
		return err
	}

	err = pds.GetFlags().ForEachE(func(parameter *ParameterDefinition) error {
		err := parameter.CheckParameterDefaultValueValidity()
		if err != nil {
			return errors.Wrapf(err, "Invalid default value for argument %s", parameter.Name)
		}

		flagName := prefix + parameter.Name
		// replace _ with -
		flagName = strings.ReplaceAll(flagName, "_", "-")
		shortFlag := parameter.ShortFlag
		if prefix != "" {
			// we don't allow shortflags if a prefix was given
			shortFlag = ""
		}
		ok := false

		f := flagSet.Lookup(flagName)
		if f != nil {
			return errors.Errorf("Flag '%s' already exists", flagName)
		}

		switch parameter.Type {
		case ParameterTypeStringListFromFile,
			ParameterTypeStringFromFile,
			ParameterTypeObjectFromFile,
			ParameterTypeObjectListFromFile,
			ParameterTypeFile:
			defaultValue := ""

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.String(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeStringListFromFiles,
			ParameterTypeStringFromFiles,
			ParameterTypeObjectListFromFiles,
			ParameterTypeFileList:
			defaultValue := []string{}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.StringSlice(flagName, defaultValue, parameter.Help)
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
				defaultValue, ok = cast.CastNumberInterfaceToInt[int](parameter.Default)
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
				defaultValue, ok = cast.CastFloatInterfaceToFloat[float64](parameter.Default)
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
				switch v_ := parameter.Default.(type) {
				case string:
					_, err2 := ParseDate(v_)
					if err2 != nil {
						return err2
					}
					defaultValue = v_
				case time.Time:
					// nothing to do
					defaultValue = v_.Format("2006-01-02")
				default:
					return errors.Errorf("Default value for parameter %s is not a valid date: %v", parameter.Name, parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, parameter.Help)
			} else {
				flagSet.String(flagName, defaultValue, parameter.Help)
			}

		case ParameterTypeStringList, ParameterTypeChoiceList:
			var defaultValue []string

			if parameter.Default != nil {
				stringList, ok := parameter.Default.([]string)
				if !ok {
					defaultValue, ok := parameter.Default.([]interface{})
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, parameter.Default)
					}

					// convert to string list
					stringList, ok = cast.CastList[string, interface{}](defaultValue)
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
				defaultValue, ok = cast.CastInterfaceToIntList[int](parameter.Default)
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
				defaultValue, ok = cast.CastInterfaceToFloatList[float64](parameter.Default)
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
			return errors.Errorf("Unknown parameter type for parameter %s: %s", parameter.Name, parameter.Type)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func GatherFlagsFromViper(
	params *ParameterDefinitions,
	onlyProvided bool,
	prefix string,
) (*ParsedParameters, error) {
	ret := NewParsedParameters()

	for v := params.Oldest(); v != nil; v = v.Next() {
		p := v.Value

		parsed := &ParsedParameter{
			ParameterDefinition: p,
		}

		flagName := prefix + p.Name
		if onlyProvided && !viper.IsSet(flagName) {
			continue
		}
		if !onlyProvided && !viper.IsSet(flagName) {
			if p.Default != nil {
				parsed.SetWithSource("default", p.Default)
				ret.Set(p.Name, parsed)
			}
			continue
		}

		// TODO(manuel, 2023-12-22) Would be cool if viper were to tell us where the flag came from...
		options := []ParseStepOption{
			WithParseStepMetadata(map[string]interface{}{
				"flag": flagName,
			}),
			WithParseStepSource("viper"),
		}
		//exhaustive:ignore
		switch p.Type {
		case ParameterTypeString:
			parsed.Set(viper.GetString(flagName), options...)
		case ParameterTypeInteger:
			parsed.Set(viper.GetInt(flagName), options...)
		case ParameterTypeFloat:
			parsed.Set(viper.GetFloat64(flagName), options...)
		case ParameterTypeBool:
			parsed.Set(viper.GetBool(flagName), options...)
		case ParameterTypeStringList:
			parsed.Set(viper.GetStringSlice(flagName), options...)
		case ParameterTypeIntegerList:
			parsed.Set(viper.GetIntSlice(flagName), options...)
		case ParameterTypeKeyValue:
			parsed.Set(viper.GetStringMapString(flagName), options...)
		case ParameterTypeStringListFromFile:
			parsed.Set(viper.GetStringSlice(flagName), options...)
		case ParameterTypeStringFromFile:
			// not sure if this is the best here, maybe it should be the filename?
			parsed.Set(viper.GetString(flagName), options...)
		case ParameterTypeChoice:
			// probably should do some checking here
			parsed.Set(viper.GetString(flagName), options...)
		case ParameterTypeObjectFromFile:
			parsed.Set(viper.GetStringMap(flagName), options...)
			// TODO(manuel, 2023-09-19) Add more of the newer types here too
		default:
			return nil, errors.Errorf("Unknown parameter type %s for flag %s", p.Type, p.Name)
		}

		ret.Set(p.Name, parsed)
	}

	return ret, nil
}

// GatherFlagsFromCobraCommand gathers the flags from the cobra command, and parses them according
// to the parameter description passed in params. The result is a map of parameter
// names to parsed values.
//
// If onlyProvided is true, only parameters that are provided
// by the user are returned (i.e. not the default values).
//
// If a parameter cannot be parsed correctly, or is missing even though it is not optional,
// an error is returned.
//
// The required argument checks that all the required parameter definitions are present.
// The provided argument only checks that the provided flags are passed.
// Prefix is prepended to all flag names.
func GatherFlagsFromCobraCommand(
	cmd *cobra.Command,
	params *ParameterDefinitions,
	onlyProvided bool,
	ignoreRequired bool,
	prefix string,
) (*ParsedParameters, error) {
	ps := NewParsedParameters()

	for v := params.Oldest(); v != nil; v = v.Next() {
		pd := v.Value
		p := &ParsedParameter{
			ParameterDefinition: pd,
		}

		if pd.IsArgument {
			continue
		}

		// check if the flag is set
		flagName := prefix + pd.Name
		flagName = strings.ReplaceAll(flagName, "_", "-")

		if !cmd.Flags().Changed(flagName) {
			if pd.Required {
				if ignoreRequired {
					continue
				}

				return nil, errors.Errorf("Parameter %s is required", pd.Name)
			}

			if pd.Default == nil {
				continue
			}

			if onlyProvided {
				continue
			}
		}

		options := []ParseStepOption{
			WithParseStepMetadata(map[string]interface{}{
				"flag": flagName,
			}),
			WithParseStepSource("cobra"),
		}

		switch pd.Type {
		case ParameterTypeObjectFromFile,
			ParameterTypeObjectListFromFile,
			ParameterTypeStringFromFile,
			ParameterTypeStringListFromFile,
			ParameterTypeString,
			ParameterTypeFile,
			ParameterTypeDate,
			ParameterTypeChoice:
			v, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return nil, err
			}
			v2, err := pd.ParseParameter([]string{v})
			if err != nil {
				return nil, err
			}
			ps.Set(pd.Name, v2)

		case ParameterTypeFloat:
			v, err := cmd.Flags().GetFloat64(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeInteger:
			v, err := cmd.Flags().GetInt(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeBool:
			v, err := cmd.Flags().GetBool(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeObjectListFromFiles,
			ParameterTypeStringFromFiles,
			ParameterTypeStringListFromFiles,
			ParameterTypeFileList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return nil, err
			}
			v2, err := pd.ParseParameter(v)
			if err != nil {
				return nil, err
			}
			ps.Set(pd.Name, v2)

		case ParameterTypeStringList,
			ParameterTypeChoiceList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeKeyValue:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return nil, err
			}

			// if it was changed and is empty, then skip setting from default
			if cmd.Flags().Changed(flagName) && len(v) == 0 {
				options := append(options,
					WithParseStepMetadata(map[string]interface{}{
						"flag":      flagName,
						"emptyFlag": true,
					}))
				p.Set(map[string]string{}, options...)
				ps.Set(pd.Name, p)
			} else {
				v2, err := pd.ParseParameter(v)
				if err != nil {
					return nil, err
				}
				ps.Set(pd.Name, v2)
			}

		case ParameterTypeIntegerList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetIntSlice(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeFloatList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetFloat64Slice(flagName)
			if err != nil {
				return nil, err
			}
			p.Set(v, options...)
			ps.Set(pd.Name, p)
		}
	}
	return ps, nil
}
