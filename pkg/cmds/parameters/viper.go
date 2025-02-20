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

				err := parsed.Update(*p.Default, options_...)
				if err != nil {
					return nil, err
				}
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
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeInteger:
			err := parsed.Update(viper.GetInt(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeFloat:
			err := parsed.Update(viper.GetFloat64(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeBool:
			err := parsed.Update(viper.GetBool(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeStringList:
			err := parsed.Update(viper.GetStringSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeIntegerList:
			err := parsed.Update(viper.GetIntSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeKeyValue:
			err := parsed.Update(viper.GetStringMapString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeStringListFromFile:
			err := parsed.Update(viper.GetStringSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeStringFromFile:
			// not sure if this is the best here, maybe it should be the filename?
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeChoice:
			// probably should do some checking here
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case ParameterTypeObjectFromFile:
			err := parsed.Update(viper.GetStringMap(flagName), options...)
			if err != nil {
				return nil, err
			}
			// TODO(manuel, 2023-09-19) Add more of the newer types here too
		default:
			return nil, errors.Errorf("Unknown parameter type %s for flag %s", p.Type, p.Name)
		}

		ret.Set(p.Name, parsed)
	}

	return ret, nil
}
