package parameters

import "fmt"

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
func (pds *ParameterDefinitions) GatherArguments(
	args []string,
	onlyProvided bool,
	ignoreRequired bool,
	parseOptions ...ParseStepOption,
) (*ParsedParameters, error) {
	_ = args
	result := NewParsedParameters()
	argsIdx := 0

	err := pds.ForEachE(func(argument *ParameterDefinition) error {
		if !argument.IsArgument {
			return nil
		}

		p := &ParsedParameter{
			ParameterDefinition: argument,
		}

		if argsIdx >= len(args) {
			if argument.Required {
				if ignoreRequired {
					return nil
				}
				return fmt.Errorf("Argument %s not found", argument.Name)
			} else {
				if argument.Default != nil && !onlyProvided {
					parseOptions_ := append(parseOptions, WithParseStepSource("default"))

					p.Update(*argument.Default, parseOptions_...)
					result.Set(argument.Name, p)
				}
				return nil
			}
		}

		v := []string{args[argsIdx]}

		if argument.Type.IsList() {
			v = args[argsIdx:]
			argsIdx = len(args)
		} else {
			argsIdx++
		}
		i2, err := argument.ParseParameter(v, parseOptions...)
		if err != nil {
			return err
		}

		result.Set(argument.Name, i2)

		return nil
	})
	if err != nil {
		return nil, err
	}

	if argsIdx < len(args) {
		return nil, fmt.Errorf("Too many arguments")
	}
	return result, nil
}
