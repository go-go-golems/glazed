package fields

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GenerateUseString creates a string representation of the 'Use' field for a given cobra command and a list of field definitions. The first word of the existing 'Use' field is treated as the verb for the command.
// The resulting string briefly describes how to use the command respecting the following conventions:
//   - Required fields are enclosed in '<>'.
//   - Optional fields are enclosed in '[]'.
//   - Optional fields that accept multiple input (TypeStringList or TypeIntegerList) are followed by '...'.
//   - If a field has a default value, it is specified after field name like 'field (default: value)'.
//
// For example:
//   - If there is a required field 'name', and an optional field 'age' with a default value of '30', the resulting string will be: 'verb <name> [age (default: 30)]'.
//   - If there is a required field 'name', and an optional field 'colors' of type TypeStringList, the resulting Use string will be: 'verb <name> [colors...]'
func GenerateUseString(cmdName string, pds *Definitions) string {
	fields := strings.Fields(cmdName)
	if len(fields) == 0 {
		return ""
	}
	verb := fields[0]
	useStr := verb
	var defaultValueStr string

	pds.ForEach(func(arg *Definition) {
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
		if arg.Type == TypeStringList || arg.Type == TypeIntegerList {
			useStr += " " + left + arg.Name + "..." + defaultValueStr + right
		} else {
			useStr += " " + left + arg.Name + defaultValueStr + right
		}
	})

	return useStr
}

// addArgumentsToCobraCommand adds each Definition from `arguments` as positional arguments to the provided `cmd` cobra command.
// Argument ordering, optionality, and multiplicity constraints are respected:
//   - Required arguments (argument.Required == true) should come before the optional.
//   - Only one list argument (either TypeStringList or TypeIntegerList) is allowed and it should be the last one.
//
// Any violation of these conditions yields an error.
// This function processes each argument, checks their default values for validity, which if invalid,
// triggers an error return.
//
// It computes the minimum and maximum number of arguments the command can take based on the required, optional,
// and list arguments.
// If everything is successful, it assigns an argument validator (either MinimumNArgs or RangeArgs)
// to the cobra command's Args attribute.
func (pds *Definitions) addArgumentsToCobraCommand(cmd *cobra.Command) error {
	minArgs := 0
	// -1 signifies unbounded
	maxArgs := 0
	hadOptional := false

	if pds.Len() == 0 {
		return nil
	}

	err := pds.ForEachE(func(argument *Definition) error {
		if !argument.IsArgument {
			return nil
		}
		if maxArgs == -1 {
			// already handling unbounded arguments
			return errors.Errorf("Cannot handle more than one unbounded argument, but found %s", argument.Name)
		}
		_, err := argument.CheckDefaultValueValidity()
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

// AddFieldsToCobraCommand takes the fields from a CommandDescription and converts them
// to cobra flags, before adding them to the Fields() of a the passed cobra command.
func (pds *Definitions) AddFieldsToCobraCommand(
	cmd *cobra.Command,
	prefix string,
) error {
	flagSet := cmd.Flags()

	err := pds.GetArguments().addArgumentsToCobraCommand(cmd)
	if err != nil {
		return err
	}

	err = pds.GetFlags().ForEachE(func(field *Definition) error {
		_, err := field.CheckDefaultValueValidity()
		if err != nil {
			return errors.Wrapf(err, "Invalid default value for argument %s", field.Name)
		}

		flagName := prefix + field.Name
		// replace _ with -
		flagName = strings.ReplaceAll(flagName, "_", "-")
		shortFlag := field.ShortFlag
		if prefix != "" {
			// we don't allow shortflags if a prefix was given
			shortFlag = ""
		}
		ok := false

		f := flagSet.Lookup(flagName)
		if f != nil {
			return errors.Errorf("Flag '%s' (usage: %s) already exists", flagName, f.Usage)
		}

		helpText := field.Help
		helpText = fmt.Sprintf("%s - <%s>", helpText, field.Type)

		switch field.Type {
		case TypeStringListFromFile,
			TypeStringFromFile,
			TypeObjectFromFile,
			TypeObjectListFromFile,
			TypeFile:
			defaultValue := ""

			if field.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}

		case TypeStringListFromFiles,
			TypeStringFromFiles,
			TypeObjectListFromFiles,
			TypeFileList:
			defaultValue := []string{}

			if field.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case TypeString, TypeSecret:
			defaultValue := ""

			if field.Default != nil {
				defaultValue, err = cast.ToString(*field.Default)
				if err != nil {
					return errors.Errorf("Default value for field %s is not a string: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}
		case TypeInteger:
			defaultValue := 0

			if field.Default != nil {
				defaultValue, ok = cast.CastNumberInterfaceToInt[int](*field.Default)
				if !ok {
					return errors.Errorf("Default value for field %s is not an integer: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.IntP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Int(flagName, defaultValue, helpText)
			}

		case TypeFloat:
			defaultValue := 0.0

			if field.Default != nil {
				defaultValue, ok = cast.CastFloatInterfaceToFloat[float64](*field.Default)
				if !ok {
					return errors.Errorf("Default value for field %s is not a float: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.Float64P(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Float64(flagName, defaultValue, helpText)
			}

		case TypeBool:
			defaultValue := false

			if field.Default != nil {
				defaultValue, ok = (*field.Default).(bool)
				if !ok {
					return errors.Errorf("Default value for field %s is not a bool: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.BoolP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Bool(flagName, defaultValue, helpText)
			}

		case TypeDate:
			defaultValue := ""

			if field.Default != nil {
				switch v_ := (*field.Default).(type) {
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
					return errors.Errorf("Default value for field %s is not a valid date: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.String(flagName, defaultValue, helpText)
			}

		case TypeStringList, TypeChoiceList:
			var defaultValue []string

			if field.Default != nil {
				stringList, ok := (*field.Default).([]string)
				if !ok {
					defaultValue, ok := (*field.Default).([]interface{})
					if !ok {
						return errors.Errorf("Default value for field %s is not a string list: %v", field.Name, *field.Default)
					}

					// convert to string list
					stringList, ok = cast.CastList[string, interface{}](defaultValue)
					if !ok {
						return errors.Errorf("Default value for field %s is not a string list: %v", field.Name, *field.Default)
					}
				}

				defaultValue = stringList
			}

			if field.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case TypeKeyValue:
			var defaultValue []string

			if field.Default != nil {
				stringMap, ok := (*field.Default).(map[string]string)
				if !ok {
					defaultValue, ok := (*field.Default).(map[string]interface{})
					if !ok {
						return errors.Errorf("Default value for field %s is not a string list: %v", field.Name, *field.Default)
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
				return errors.Wrapf(err, "Could not convert default value for field %s to string list: %v", field.Name, *field.Default)
			}

			if field.ShortFlag != "" {
				flagSet.StringSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.StringSlice(flagName, defaultValue, helpText)
			}

		case TypeIntegerList:
			var defaultValue []int
			if field.Default != nil {
				defaultValue, ok = cast.CastInterfaceToIntList[int](*field.Default)
				if !ok {
					return errors.Errorf("Default value for field %s is not an integer list: %v", field.Name, *field.Default)
				}
			}

			if field.ShortFlag != "" {
				flagSet.IntSliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.IntSlice(flagName, defaultValue, helpText)
			}

		case TypeFloatList:
			var defaultValue []float64
			if field.Default != nil {
				defaultValue, ok = cast.CastInterfaceToFloatList[float64](*field.Default)
				if !ok {
					return errors.Errorf("Default value for field %s is not a float list: %v", field.Name, *field.Default)
				}
			}
			if field.ShortFlag != "" {
				flagSet.Float64SliceP(flagName, shortFlag, defaultValue, helpText)
			} else {
				flagSet.Float64Slice(flagName, defaultValue, helpText)
			}

		case TypeChoice:
			defaultValue := ""

			if field.Default != nil {
				defaultValue, err = cast.ToString(*field.Default)
				if err != nil {
					return errors.Errorf("Default value for field %s is not a string: %v", field.Name, *field.Default)
				}
			}

			choiceString := strings.Join(field.Choices, ",")

			if field.ShortFlag != "" {
				flagSet.StringP(flagName, shortFlag, defaultValue, fmt.Sprintf("%s (%s)", helpText, choiceString))
			} else {
				flagSet.String(flagName, defaultValue, fmt.Sprintf("%s (%s)", helpText, choiceString))
			}

		default:
			return errors.Errorf("Unknown field type for field %s: %s", field.Name, field.Type)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GatherFlagsFromCobraCommand gathers the flags from the cobra command, and parses them according
// to the field description passed in params. The result is a map of field
// names to parsed values.
//
// If onlyProvided is true, only fields that are provided
// by the user are returned (i.e. not the default values).
//
// If a field cannot be parsed correctly, or is missing even though it is not optional,
// an error is returned.
//
// The required argument checks that all the required field definitions are present.
// The provided argument only checks that the provided flags are passed.
// Prefix is prepended to all flag names.
func (pds *Definitions) GatherFlagsFromCobraCommand(
	cmd *cobra.Command,
	onlyProvided bool,
	ignoreRequired bool,
	prefix string,
	options ...ParseOption,
) (*FieldValues, error) {
	ps := NewFieldValues()

	err := pds.ForEachE(func(pd *Definition) error {
		p := &FieldValue{
			Definition: pd,
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

				return errors.Errorf("Field %s is required", pd.Name)
			}

			if pd.Default == nil {
				return nil
			}

			if onlyProvided {
				return nil
			}
		}

		options := append([]ParseOption{
			WithMetadata(map[string]interface{}{
				"flag": flagName,
			}),
			WithSource("cobra"),
		}, options...)

		switch pd.Type {
		case TypeObjectFromFile,
			TypeObjectListFromFile,
			TypeStringFromFile,
			TypeStringListFromFile,
			TypeString,
			TypeSecret,
			TypeFile,
			TypeDate,
			TypeChoice:
			v, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}
			v2, err := pd.ParseField([]string{v}, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, v2)

		case TypeFloat:
			v, err := cmd.Flags().GetFloat64(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)

		case TypeInteger:
			v, err := cmd.Flags().GetInt(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)

		case TypeBool:
			v, err := cmd.Flags().GetBool(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)

		case TypeObjectListFromFiles,
			TypeStringFromFiles,
			TypeStringListFromFiles,
			TypeFileList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}
			v2, err := pd.ParseField(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, v2)

		case TypeStringList,
			TypeChoiceList:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)

		case TypeKeyValue:
			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return err
			}

			// if it was changed and is empty, then skip setting from default
			if cmd.Flags().Changed(flagName) && len(v) == 0 {
				options_ := append(options,
					WithMetadata(map[string]interface{}{
						"flag":      flagName,
						"emptyFlag": true,
					}))
				err = p.Update(map[string]string{}, options_...)
				if err != nil {
					return err
				}
				ps.Set(pd.Name, p)
			} else {
				v2, err := pd.ParseField(v, options...)
				if err != nil {
					return err
				}
				ps.Set(pd.Name, v2)
			}

		case TypeIntegerList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetIntSlice(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)

		case TypeFloatList:
			// NOTE(manuel, 2023-04-01) Do we not check for default here?
			v, err := cmd.Flags().GetFloat64Slice(flagName)
			if err != nil {
				return err
			}
			err = p.Update(v, options...)
			if err != nil {
				return err
			}
			ps.Set(pd.Name, p)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ps, nil
}
