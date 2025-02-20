package parameters

import (
	"reflect"

	"github.com/pkg/errors"
)

// GatherParametersFromMap gathers parameter values from a map.
//
// For each ParameterDefinition, it checks if a matching value is present in the map:
//
// - If the parameter is missing and required, an error is returned.
// - If the parameter is missing and optional, the default value is used.
// - If the value is provided, it is validated against the definition.
//
// Values are looked up by parameter name, as well as short flag if provided.
//
// The returned map contains the gathered parameter values, with defaults filled in
// for any missing optional parameters.
func (pds *ParameterDefinitions) GatherParametersFromMap(
	m map[string]interface{},
	onlyProvided bool,
	options ...ParseStepOption,
) (*ParsedParameters, error) {
	ret := NewParsedParameters()

	err := pds.ForEachE(func(p *ParameterDefinition) error {
		parsed := &ParsedParameter{
			ParameterDefinition: p,
		}

		v_, ok := m[p.Name]
		if !ok {
			if p.ShortFlag != "" {
				v_, ok = m[p.ShortFlag]
			}
			if onlyProvided {
				return nil
			}
			if !ok {
				if p.Default != nil {
					err := parsed.Update(*p.Default)
					if err != nil {
						return err
					}
					ret.Set(p.Name, parsed)
					return nil
				}

				if p.Required {
					return errors.Errorf("Missing required parameter %s", p.Name)
				}
			}
		}
		options_ := append(options,
			WithParseStepMetadata(map[string]interface{}{
				"map-value": v_,
			}))

		if s, ok := v_.(string); ok {
			v__, err := p.ParseParameter([]string{s})
			if err != nil {
				return errors.Wrapf(err, "Invalid value for parameter %s", p.Name)
			}
			v_ = v__.Value
		}

		// TODO(manuel, 2023-12-28) We need to check if nil means remove the value or use the default or whatever that means
		v2, err := p.CheckValueValidity(v_)
		if err != nil {
			return errors.Wrapf(err, "Invalid value for parameter %s", p.Name)
		}

		// NOTE(manuel, 2023-12-22) We might want to pass in that name instead of just saying from-map
		err = parsed.Update(v2, options_...)
		if err != nil {
			return err
		}
		ret.Set(p.Name, parsed)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *ParameterDefinition) GatherValueFromInterface(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}

	// Check if the value is valid
	castValue, err := p.CheckValueValidity(value.Interface())
	if err != nil {
		return err
	}

	// Set the value
	return p.setReflectValue(value, castValue)
}
