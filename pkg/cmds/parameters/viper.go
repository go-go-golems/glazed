package parameters

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func (pds *ParameterDefinitions) GatherFlagsFromViper(
	onlyProvided bool,
	prefix string,
	options ...ParseStepOption,
) (*ParsedParameters, error) {
	ret := NewParsedParameters()

	for v := pds.Oldest(); v != nil; v = v.Next() {
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
				options_ := append(options, WithParseStepSource("default"))

				parsed.Update(*p.Default, options_...)
				ret.Set(p.Name, parsed)
			}
			continue
		}

		// TODO(manuel, 2023-12-22) Would be cool if viper were to tell us where the flag came from...
		options := append([]ParseStepOption{
			WithParseStepMetadata(map[string]interface{}{
				"flag": flagName,
			}),
			WithParseStepSource("viper"),
		}, options...)
		//exhaustive:ignore
		switch p.Type {
		case ParameterTypeString:
			parsed.Update(viper.GetString(flagName), options...)
		case ParameterTypeInteger:
			parsed.Update(viper.GetInt(flagName), options...)
		case ParameterTypeFloat:
			parsed.Update(viper.GetFloat64(flagName), options...)
		case ParameterTypeBool:
			parsed.Update(viper.GetBool(flagName), options...)
		case ParameterTypeStringList:
			parsed.Update(viper.GetStringSlice(flagName), options...)
		case ParameterTypeIntegerList:
			parsed.Update(viper.GetIntSlice(flagName), options...)
		case ParameterTypeKeyValue:
			parsed.Update(viper.GetStringMapString(flagName), options...)
		case ParameterTypeStringListFromFile:
			parsed.Update(viper.GetStringSlice(flagName), options...)
		case ParameterTypeStringFromFile:
			// not sure if this is the best here, maybe it should be the filename?
			parsed.Update(viper.GetString(flagName), options...)
		case ParameterTypeChoice:
			// probably should do some checking here
			parsed.Update(viper.GetString(flagName), options...)
		case ParameterTypeObjectFromFile:
			parsed.Update(viper.GetStringMap(flagName), options...)
			// TODO(manuel, 2023-09-19) Add more of the newer types here too
		default:
			return nil, errors.Errorf("Unknown parameter type %s for flag %s", p.Type, p.Name)
		}

		ret.Set(p.Name, parsed)
	}

	return ret, nil
}
