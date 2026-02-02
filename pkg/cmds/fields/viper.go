package fields

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Deprecated: Use config files + env middlewares instead.
func (pds *Definitions) GatherFlagsFromViper(
	onlyProvided bool,
	prefix string,
	options ...ParseOption,
) (*ParsedParameters, error) {
	warnGatherViperParamsOnce.Do(func() {
		log.Warn().Msg("fields.GatherFlagsFromViper is deprecated; use LoadParametersFromFiles + UpdateFromEnv")
	})
	ret := NewParsedParameters()

	for v := pds.Oldest(); v != nil; v = v.Next() {
		p := v.Value

		parsed := &ParsedParameter{
			Definition: p,
		}

		flagName := prefix + p.Name
		if onlyProvided && !viper.IsSet(flagName) {
			continue
		}
		if !onlyProvided && !viper.IsSet(flagName) {
			if p.Default != nil {
				options_ := append(options, WithSource("default"))

				err := parsed.Update(*p.Default, options_...)
				if err != nil {
					return nil, err
				}
				ret.Set(p.Name, parsed)
			}
			continue
		}

		// Add metadata about the viper key and derived env key shape
		upperKey := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
		meta := map[string]interface{}{
			"flag":    flagName,
			"env_key": upperKey,
		}
		options := append([]ParseOption{
			WithMetadata(meta),
			WithSource("viper"),
		}, options...)
		//exhaustive:ignore
		switch p.Type {
		case TypeString, TypeSecret:
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeInteger:
			err := parsed.Update(viper.GetInt(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeFloat:
			err := parsed.Update(viper.GetFloat64(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeBool:
			err := parsed.Update(viper.GetBool(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeStringList:
			err := parsed.Update(viper.GetStringSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeIntegerList:
			err := parsed.Update(viper.GetIntSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeKeyValue:
			err := parsed.Update(viper.GetStringMapString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeStringListFromFile:
			err := parsed.Update(viper.GetStringSlice(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeStringFromFile:
			// not sure if this is the best here, maybe it should be the filename?
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeChoice:
			// probably should do some checking here
			err := parsed.Update(viper.GetString(flagName), options...)
			if err != nil {
				return nil, err
			}
		case TypeObjectFromFile:
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

var warnGatherViperParamsOnce sync.Once
