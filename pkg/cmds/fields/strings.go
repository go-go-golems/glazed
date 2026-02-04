package fields

import (
	"github.com/pkg/errors"
	"strings"
)

// GatherFlagsFromStringList parses command line arguments into a FieldValues
// map. It accepts a slice of string arguments, bools to control required/provided
// flag handling, and a prefix to prepend to flag names.
//
// It returns the parsed fields map, any non-flag arguments, and any error
// encountered during parsing.
//
// onlyProvided controls whether to only include flags that were explicitly
// provided on the command line. If false, default values will be included
// for any flags that were not provided.
//
// ignoreRequired controls whether required flags are enforced.
// If true, missing required flags will not trigger an error.
//
// prefix allows prepending a string to flag names to namespace them.
// For example, a prefix of "user" would namespace flags like "--name"
// to "--user-name". This allows reuse of a flag set across different
// sub-commands. The prefix has "-" appended automatically.
func (pds *Definitions) GatherFlagsFromStringList(
	args []string,
	onlyProvided bool,
	ignoreRequired bool,
	prefix string,
	parseOptions ...ParseOption,
) (*FieldValues, []string, error) {
	flagMap := NewDefinitions()
	flagNames := map[string]string{}
	remainingArgs := []string{}

	// build a map of flag names to field definitions, including through shortflags
	err := pds.ForEachE(func(param *Definition) error {
		if param.IsArgument {
			return nil
		}
		flagName := prefix + param.Name
		flagName = strings.ReplaceAll(flagName, "_", "-")
		if _, ok := flagMap.Get(flagName); ok {
			return errors.Errorf("duplicate flag: --%s", flagName)
		}
		flagMap.Set(flagName, param)
		flagNames[flagName] = param.Name

		if prefix == "" {
			if param.ShortFlag != "" {
				if _, ok := flagMap.Get(param.ShortFlag); ok {
					return errors.Errorf("duplicate flag: -%s", param.ShortFlag)
				}
				flagMap.Set(param.ShortFlag, param)
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// TODO(manuel, 2023-12-25) Handle -- and switch to full flags
	// See: https://github.com/go-go-golems/glazed/issues/399
	rawValues := make(map[string][]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var flagName string
		var param *Definition
		var ok bool
		if strings.HasPrefix(arg, "--") {
			splitArg := strings.SplitN(arg[2:], "=", 2)
			flagName = splitArg[0]
			param, ok = flagMap.Get(flagName)
			if !ok {
				return nil, nil, errors.Errorf("unknown flag: --%s", flagName)
			}
			if len(splitArg) == 2 {
				if param.Type.IsList() {
					value := strings.Trim(splitArg[1], "[]")
					values := strings.Split(value, ",")
					rawValues[flagName] = append(rawValues[flagName], values...)
				} else {
					rawValues[flagName] = append(rawValues[flagName], splitArg[1])
				}
				continue
			}
		} else if strings.HasPrefix(arg, "-") {
			flagName = arg[1:]
			param, ok = flagMap.Get(flagName)
			if !ok {
				return nil, nil, errors.Errorf("unknown flag: -%s", flagName)
			}
		} else {
			remainingArgs = append(remainingArgs, arg)
			continue
		}

		if param.Type == TypeBool {
			rawValues[flagName] = append(rawValues[flagName], "true")
		} else {
			if i+1 >= len(args) {
				return nil, nil, errors.Errorf("missing value for flag: -%s", flagName)
			}
			value := args[i+1]
			i++
			if param.Type.IsList() {
				value = strings.Trim(value, "[]")
				values := strings.Split(value, ",")
				rawValues[flagName] = append(rawValues[flagName], values...)
				continue
			}
			rawValues[flagName] = append(rawValues[flagName], value)
		}
	}

	result := NewFieldValues()
	for paramName, values := range rawValues {
		param, ok := flagMap.Get(paramName)
		if !ok || param == nil {
			return nil, nil, errors.Errorf("unknown flag: --%s", paramName)
		}
		parsedValue, err := param.ParseField(values, parseOptions...)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "invalid value for flag --%s", paramName)
		}
		result.Set(param.Name, parsedValue)
	}

	err = pds.ForEachE(func(param *Definition) error {
		if param.Required && !ignoreRequired {
			if _, ok := rawValues[param.Name]; !ok {
				return errors.Errorf("missing required flag: --%s", flagNames[param.Name])
			}
		}
		if !onlyProvided {
			if _, ok := result.Get(param.Name); !ok && param.Default != nil {
				p := &FieldValue{
					Definition: param,
				}
				// show that this was set out of the default
				parseOptions_ := append(parseOptions, WithMetadata(map[string]interface{}{
					"default": true,
				}), WithSource("default"))
				err := p.Update(*param.Default, parseOptions_...)
				if err != nil {
					return err
				}
				result.Set(param.Name, p)
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return result, remainingArgs, nil
}
