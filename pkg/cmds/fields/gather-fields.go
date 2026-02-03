package fields

import (
	"reflect"

	"github.com/pkg/errors"
)

// GatherFieldsFromMap gathers field values from a map.
//
// For each Definition, it checks if a matching value is present in the map:
//
// - If the field is missing and required, an error is returned.
// - If the field is missing and optional, the default value is used.
// - If the value is provided, it is validated against the definition.
//
// Values are looked up by field name, as well as short flag if provided.
//
// The returned map contains the gathered field values, with defaults filled in
// for any missing optional fields.
func (pds *Definitions) GatherFieldsFromMap(
	m map[string]interface{},
	onlyProvided bool,
	options ...ParseOption,
) (*FieldValues, error) {
	ret := NewFieldValues()

	err := pds.ForEachE(func(p *Definition) error {
		parsed := &FieldValue{
			Definition: p,
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
					return errors.Errorf("Missing required field %s", p.Name)
				}
			}
		}
		options_ := append(options,
			WithMetadata(map[string]interface{}{
				"map-value": v_,
			}))

		if s, ok := v_.(string); ok {
			v__, err := p.ParseField([]string{s})
			if err != nil {
				return errors.Wrapf(err, "Invalid value for field %s", p.Name)
			}
			v_ = v__.Value
		}

		// TODO(manuel, 2023-12-28) We need to check if nil means remove the value or use the default or whatever that means
		v2, err := p.CheckValueValidity(v_)
		if err != nil {
			return errors.Wrapf(err, "Invalid value for field %s", p.Name)
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

func (p *Definition) GatherValueFromInterface(value reflect.Value) error {
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
