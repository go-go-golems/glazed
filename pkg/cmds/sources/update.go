package sources

import (
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// FromDefaults is a middleware that sets default values from parameter definitions.
// It calls the next handler, and then iterates through each layer and parameter definition.
// If a default is defined, it sets that as the parameter value in the parsed layer.
func FromDefaults(options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.UpdateWithDefaults(parsedLayers, options...)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// FromMap takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers, skipping any layers not present in layers_.
func FromMap(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromMap(layers_, parsedLayers, m, options...)
		}
	}
}

// FromMapFirst takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers, skipping any layers not present in layers_.
func FromMapFirst(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := updateFromMap(layers_, parsedLayers, m, options...)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

// FromMapAsDefault takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers if the parameter hasn't already been set, skipping any layers not present in layers_.
func FromMapAsDefault(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromMapAsDefault(layers_, parsedLayers, m, options...)
		}
	}
}

// FromMapAsDefaultFirst takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers if the parameter hasn't already been set, skipping any layers not present in layers_.
func FromMapAsDefaultFirst(m map[string]map[string]interface{}, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := updateFromMapAsDefault(layers_, parsedLayers, m, options...)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

func updateFromMap(
	layers_ *schema.Schema,
	parsedLayers *values.Values,
	m map[string]map[string]interface{},
	options ...fields.ParseOption) error {
	for k, v := range m {
		layer, ok := layers_.Get(k)
		if !ok {
			continue
		}

		parsedLayer := parsedLayers.GetOrCreate(layer)
		ps, err := layer.GetDefinitions().GatherParametersFromMap(v, true, options...)
		if err != nil {
			return err
		}
		_, err = parsedLayer.Parameters.Merge(ps)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateFromMapAsDefault(
	layers_ *schema.Schema,
	parsedLayers *values.Values,
	m map[string]map[string]interface{},
	options ...fields.ParseOption) error {
	for k, v := range m {
		layer, ok := layers_.Get(k)
		if !ok {
			continue
		}

		parsedLayer := parsedLayers.GetOrCreate(layer)
		ps, err := layer.GetDefinitions().GatherParametersFromMap(v, true, options...)
		if err != nil {
			return err
		}
		_, err = parsedLayer.Parameters.MergeAsDefault(ps)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateFromEnv(
	layers_ *schema.Schema,
	parsedLayers *values.Values,
	prefix string,
	options ...fields.ParseOption,
) error {
	err := layers_.ForEachE(func(key string, l schema.Section) error {
		parsedLayer := parsedLayers.GetOrCreate(l)
		pds := l.GetDefinitions()
		layerPrefix := l.GetPrefix()
		err := pds.ForEachE(func(p *fields.Definition) error {
			// Compute env key based on layer prefix + param name, hyphen->underscore, uppercase,
			// and optional global prefix (app name) separated by underscore.
			base := layerPrefix + p.Name
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

				// Parse env string into the appropriate typed value using the parameter's parser.
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

				pp, err := p.ParseParameter(inputs, opts...)
				if err != nil {
					return err
				}
				// Preserve parse log/metadata when updating parsed fields.
				if err := parsedLayer.Parameters.UpdateWithLog(p.Name, p, pp.Value, pp.Log...); err != nil {
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
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromEnv(layers_, parsedLayers, prefix, options...)
		}
	}
}

func updateFromStringList(layers_ *schema.Schema, parsedLayers *values.Values, prefix string, args []string, options ...fields.ParseOption) error {
	err := layers_.ForEachE(func(key string, l schema.Section) error {
		parsedLayer := parsedLayers.GetOrCreate(l)
		pds := l.GetDefinitions()
		ps, remainingArgs, err := pds.GatherFlagsFromStringList(args, true, true, prefix, options...)
		if err != nil {
			return err
		}

		_, err = parsedLayer.Parameters.Merge(ps)
		if err != nil {
			return err
		}

		ps, err = pds.GatherArguments(remainingArgs, true, true, options...)
		if err != nil {
			return err
		}

		_, err = parsedLayer.Parameters.Merge(ps)
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
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromStringList(layers_, parsedLayers, prefix, args, options...)
		}
	}
}
