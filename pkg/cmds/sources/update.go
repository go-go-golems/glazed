package sources

import (
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// FromDefaults is a middleware that sets default values from field definitions.
// It calls the next handler, and then iterates through each section and field definition.
// If a default is defined, it sets that as the field value in the parsed values.
func FromDefaults(options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}
			err = schema_.UpdateWithDefaults(parsedValues, options...)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// FromMap takes a map where the keys are section slugs and the values are
// maps of field name -> value. It calls next, and then merges the provided
// values into the parsed values, skipping any sections not present in the schema.
func FromMap(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return updateFromMap(schema_, parsedValues, m, options...)
		}
	}
}

// FromMapFirst takes a map where the keys are section slugs and the values are
// maps of field name -> value. It calls next, and then merges the provided
// values into the parsed values, skipping any sections not present in the schema.
func FromMapFirst(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := updateFromMap(schema_, parsedValues, m, options...)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// FromMapAsDefault takes a map where the keys are section slugs and the values are
// maps of field name -> value. It calls next, and then merges the provided
// values into the parsed values if the field hasn't already been set, skipping any sections not present in the schema.
func FromMapAsDefault(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return updateFromMapAsDefault(schema_, parsedValues, m, options...)
		}
	}
}

// FromMapAsDefaultFirst takes a map where the keys are section slugs and the values are
// maps of field name -> value. It calls next, and then merges the provided
// values into the parsed values if the field hasn't already been set, skipping any sections not present in the schema.
func FromMapAsDefaultFirst(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := updateFromMapAsDefault(schema_, parsedValues, m, options...)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

func updateFromMap(
	schema_ *schema.Schema,
	parsedValues *values.Values,
	m map[string]map[string]interface{},
	options ...fields.ParseOption) error {
	for k, v := range m {
		section, ok := schema_.Get(k)
		if !ok {
			continue
		}

		sectionValues := parsedValues.GetOrCreate(section)
		ps, err := section.GetDefinitions().GatherFieldsFromMap(v, true, options...)
		if err != nil {
			return err
		}
		_, err = sectionValues.Fields.Merge(ps)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateFromMapAsDefault(
	schema_ *schema.Schema,
	parsedValues *values.Values,
	m map[string]map[string]interface{},
	options ...fields.ParseOption) error {
	for k, v := range m {
		section, ok := schema_.Get(k)
		if !ok {
			continue
		}

		sectionValues := parsedValues.GetOrCreate(section)
		ps, err := section.GetDefinitions().GatherFieldsFromMap(v, true, options...)
		if err != nil {
			return err
		}
		_, err = sectionValues.Fields.MergeAsDefault(ps)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateFromEnv(
	schema_ *schema.Schema,
	parsedValues *values.Values,
	prefix string,
	options ...fields.ParseOption,
) error {
	err := schema_.ForEachE(func(key string, l schema.Section) error {
		sectionValues := parsedValues.GetOrCreate(l)
		pds := l.GetDefinitions()
		sectionPrefix := l.GetPrefix()
		err := pds.ForEachE(func(p *fields.Definition) error {
			// Compute env key based on section prefix + field name, hyphen->underscore, uppercase,
			// and optional global prefix (app name) separated by underscore.
			base := sectionPrefix + p.Name
			envKey := strings.ToUpper(strings.ReplaceAll(base, "-", "_"))
			if prefix != "" {
				envKey = strings.ToUpper(prefix) + "_" + envKey
			}

			if v, ok := os.LookupEnv(envKey); ok {
				opts := append([]fields.ParseOption{
					fields.WithMetadata(map[string]interface{}{
						"env_key": envKey,
					}),
					fields.WithSource("env"),
				}, options...)

				// Parse env string into the appropriate typed value using the field's parser.
				// For list-like types, split on commas (trim brackets and whitespace).
				var inputs []string
				if p.Type.IsList() {
					s := strings.TrimSpace(v)
					// Trim optional surrounding brackets: [a,b,c]
					if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
						s = strings.TrimPrefix(s, "[")
						s = strings.TrimSuffix(s, "]")
						s = strings.TrimSpace(s)
					}
					if s == "" {
						inputs = []string{}
					} else {
						parts := strings.Split(s, ",")
						for _, part := range parts {
							inputs = append(inputs, strings.TrimSpace(part))
						}
					}
				} else {
					inputs = []string{v}
				}

				pp, err := p.ParseField(inputs, opts...)
				if err != nil {
					return err
				}
				// Preserve parse log/metadata when updating parsed fields.
				if err := sectionValues.Fields.UpdateWithLog(p.Name, p, pp.Value, pp.Log...); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func FromEnv(prefix string, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return updateFromEnv(schema_, parsedValues, prefix, options...)
		}
	}
}

func updateFromStringList(schema_ *schema.Schema, parsedValues *values.Values, prefix string, args []string, options ...fields.ParseOption) error {
	err := schema_.ForEachE(func(key string, l schema.Section) error {
		sectionValues := parsedValues.GetOrCreate(l)
		pds := l.GetDefinitions()
		ps, remainingArgs, err := pds.GatherFlagsFromStringList(args, true, true, prefix, options...)
		if err != nil {
			return err
		}

		_, err = sectionValues.Fields.Merge(ps)
		if err != nil {
			return err
		}

		ps, err = pds.GatherArguments(remainingArgs, true, true, options...)
		if err != nil {
			return err
		}

		_, err = sectionValues.Fields.Merge(ps)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func UpdateFromStringList(prefix string, args []string, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return updateFromStringList(schema_, parsedValues, prefix, args, options...)
		}
	}
}
