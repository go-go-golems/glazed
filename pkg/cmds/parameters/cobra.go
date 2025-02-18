package parameters

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

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
func GenerateUseString(cmdName string, pds *ParameterDefinitions) string {
	fields := strings.Fields(cmdName)
	if len(fields) == 0 {
		return ""
	}
	verb := fields[0]
	useStr := verb
	var defaultValueStr string

	pds.ForEach(func(arg *ParameterDefinition) {
		if !arg.IsArgument {
			return
		}
		defaultValueStr = ""
		if arg.Default != nil {
			defaultValueStr = fmt.Sprintf(" (default: %v)", *arg.Default)
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
func (pds *ParameterDefinitions) addArgumentsToCobraCommand(cmd *cobra.Command) error {
	minArgs := 0
	// -1 signifies unbounded
	maxArgs := 0
	hadOptional := false

	if pds.Len() == 0 {
		return nil
	}

	err := pds.ForEachE(func(argument *ParameterDefinition) error {
		if !argument.IsArgument {
			return nil
		}
		if maxArgs == -1 {
			// already handling unbounded arguments
			return errors.Errorf("Cannot handle more than one unbounded argument, but found %s", argument.Name)
		}
		_, err := argument.CheckParameterDefaultValueValidity()
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
		if argument.Type.IsList() {
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

	cmd.Use = GenerateUseString(cmd.Use, pds)

	return nil
}

// AddParametersToCobraCommand takes the parameters from a CommandDescription and converts them
// to cobra flags, before adding them to the Parameters() of a the passed cobra command.
func (pds *ParameterDefinitions) AddParametersToCobraCommand(
	cmd *cobra.Command,
	prefix string,
) error {
	flagSet := cmd.Flags()

	err := pds.GetArguments().addArgumentsToCobraCommand(cmd)
	if err != nil {
		return err
	}

	err = pds.GetFlags().ForEachE(func(parameter *ParameterDefinition) error {
		_, err := parameter.CheckParameterDefaultValueValidity()
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

		helpText := parameter.Help
		helpText = fmt.Sprintf("%s - <%s>", helpText, parameter.Type)

		switch parameter.Type {
		case ParameterTypeStringListFromFile,
			ParameterTypeStringFromFile,
			ParameterTypeObjectFromFile,
			ParameterTypeObjectListFromFile,
			ParameterTypeFile:
			defaultValue := ""

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}

		case ParameterTypeStringListFromFiles,
			ParameterTypeStringFromFiles,
			ParameterTypeObjectListFromFiles,
			ParameterTypeFileList:
			defaultValue := []string{}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case ParameterTypeString:
			defaultValue := ""

			if parameter.Default != nil {
				defaultValue, err = cast.ToString(*parameter.Default)
				if err != nil {
					return errors.Errorf("Default value for parameter %s is not a string: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}
		case ParameterTypeInteger:
			defaultValue := 0

			if parameter.Default != nil {
				defaultValue, ok = cast.CastNumberInterfaceToInt[int](*parameter.Default)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not an integer: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.IntP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Int(flagName, defaultValue, helpText)
			}

		case ParameterTypeFloat:
			defaultValue := 0.0

			if parameter.Default != nil {
				defaultValue, ok = cast.CastFloatInterfaceToFloat[float64](*parameter.Default)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a float: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.Float64P(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Float64(flagName, defaultValue, helpText)
			}

		case ParameterTypeBool:
			defaultValue := false

			if parameter.Default != nil {
				defaultValue, ok = (*parameter.Default).(bool)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a bool: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.BoolP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Bool(flagName, defaultValue, helpText)
			}

		case ParameterTypeDate:
			defaultValue := ""

			if parameter.Default != nil {
				switch v_ := (*parameter.Default).(type) {
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
					return errors.Errorf("Default value for parameter %s is not a valid date: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}

		case ParameterTypeStringList, ParameterTypeChoiceList:
			var defaultValue []string

			if parameter.Default != nil {
				stringList, ok := (*parameter.Default).([]string)
				if !ok {
					defaultValue, ok := (*parameter.Default).([]interface{})
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, *parameter.Default)
					}

					// convert to string list
					stringList, ok = cast.CastList[string, interface{}](defaultValue)
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, *parameter.Default)
					}
				}

				defaultValue = stringList
			}
			if err != nil {
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", parameter.Name, *parameter.Default)
			}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case ParameterTypeKeyValue:
			var defaultValue []string

			if parameter.Default != nil {
				stringMap, ok := (*parameter.Default).(map[string]string)
				if !ok {
					defaultValue, ok := (*parameter.Default).(map[string]interface{})
					if !ok {
						return errors.Errorf("Default value for parameter %s is not a string list: %v", parameter.Name, *parameter.Default)
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
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", parameter.Name, *parameter.Default)
			}

			if parameter.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case ParameterTypeIntegerList:
			var defaultValue []int
			if parameter.Default != nil {
				defaultValue, ok = cast.CastInterfaceToIntList[int](*parameter.Default)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not an integer list: %v", parameter.Name, *parameter.Default)
				}
			}

			if parameter.ShortFlag != "" {
				flagSet.IntSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.IntSlice(flagName, defaultValue, helpText)
			}

		case ParameterTypeFloatList:
			var defaultValue []float64
			if parameter.Default != nil {
				defaultValue, ok = cast.CastInterfaceToFloatList[float64](*parameter.Default)
				if !ok {
					return errors.Errorf("Default value for parameter %s is not a float list: %v", parameter.Name, *parameter.Default)
				}
			}
			if parameter.ShortFlag != "" {
				flagSet.Float64SliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Float64Slice(flagName, defaultValue, helpText)
			}

		case ParameterTypeChoice:
			defaultValue := ""

			if parameter.Default != nil {
				defaultValue, err = cast.ToString(*parameter.Default)
				if err != nil {
					return errors.Errorf("Default value for parameter %s is not a string: %v", parameter.Name, *parameter.Default)
				}
			}

			choiceString := strings.Join(parameter.Choices, ",")

			if parameter.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, fmt.Sprintf("%s (%s)", helpText, choiceString))
			} else {
				flagSet.String(flagName, defaultValue, fmt.Sprintf("%s (%s)", helpText, choiceString))
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
func (pds *ParameterDefinitions) GatherFlagsFromCobraCommand(
	cmd *cobra.Command,
	onlyProvided bool,
	ignoreRequired bool,
	prefix string,
	options ...ParseStepOption,
) (*ParsedParameters, error) {
	ps := NewParsedParameters()

	err := pds.ForEachE(func(pd *ParameterDefinition) error {
		p := &ParsedParameter{
			ParameterDefinition: pd,
		}

		if pd.IsArgument {
			return nil
		}

		// check if the flag is set
		flagName := prefix + pd.Name
		flagName = strings.ReplaceAll(flagName, "_", "-")

		if !cmd.Flags().Changed(flagName) {
			if pd.Required {
				if ignoreRequired {
					return nil
				}

				return errors.Errorf("Parameter %s is required", pd.Name)
			}

			if pd.Default == nil {
				return nil
			}

			if onlyProvided {
				return nil
			}
		}

		options := append([]ParseStepOption{
			WithParseStepMetadata(map[string]interface{}{
				"flag": flagName,
			}),
			WithParseStepSource("cobra"),
		}, options...)

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
				return err
			}
			v2, err := pd.ParseParameter([]string{v}, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, v2)

		case ParameterTypeFloat:
			v, err := cmd.Flags().GetFloat64(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeInteger:
			v, err := cmd.Flags().GetInt(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeBool:
			v, err := cmd.Flags().GetBool(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeObjectListFromFiles,
			ParameterTypeStringFromFiles,
			ParameterTypeStringListFromFiles,
			ParameterTypeFileList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}
			v2, err := pd.ParseParameter(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, v2)

		case ParameterTypeStringList,
			ParameterTypeChoiceList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeKeyValue:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}

			// if it was changed and is empty, then skip setting from default
			if cmd.Flags().Changed(flagName) && len(v) == 0 {
				options_ := append(options,
					WithParseStepMetadata(map[string]interface{}{
						"flag":      flagName,
						"emptyFlag": true,
					}))
				p.Update(map[string]string{}, options_...)
				ps.Set(pd.Name, p)
			} else {
				v2, err := pd.ParseParameter(v, options...)
				if err != nil {
					return err
				}
				ps.Set(pd.Name, v2)
			}

		case ParameterTypeIntegerList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetIntSlice(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)

		case ParameterTypeFloatList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetFloat64Slice(flagName)
			if err != nil {
				return err
			}
			p.Update(v, options...)
			ps.Set(pd.Name, p)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ps, nil
}
