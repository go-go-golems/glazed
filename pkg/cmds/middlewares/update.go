package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"os"
	"strings"
)

// SetFromDefaults is a middleware that sets default values from parameter definitions.
// It calls the next handler, and then iterates through each layer and parameter definition.
// If a default is defined, it sets that as the parameter value in the parsed layer.
func SetFromDefaults(options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				pds := l.GetParameterDefinitions()
				parsedLayer := parsedLayers.GetOrCreate(l)

				err := pds.ForEachE(func(pd *parameters.ParameterDefinition) error {
					err := pd.CheckParameterDefaultValueValidity()
					if err != nil {
						return err
					}
					if pd.Default != nil {
						parsedLayer.Parameters.SetAsDefault(pd.Name, pd, *pd.Default, options...)
					}
					return nil
				})

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
	}
}

// UpdateFromMap takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers, skipping any layers not present in layers_.
func UpdateFromMap(m map[string]map[string]interface{}, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromMap(layers_, parsedLayers, m, options...)
		}
	}
}

// UpdateFromMapFirst takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers, skipping any layers not present in layers_.
func UpdateFromMapFirst(m map[string]map[string]interface{}, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := updateFromMap(layers_, parsedLayers, m, options...)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

// UpdateFromMapAsDefault takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers if the parameter hasn't already been set, skipping any layers not present in layers_.
func UpdateFromMapAsDefault(m map[string]map[string]interface{}, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromMapAsDefault(layers_, parsedLayers, m, options...)
		}
	}
}

// UpdateFromMapAsDefaultFirst takes a map where the keys are layer slugs and the values are
// maps of parameter name -> value. It calls next, and then merges the provided
// values into the parsed layers if the parameter hasn't already been set, skipping any layers not present in layers_.
func UpdateFromMapAsDefaultFirst(m map[string]map[string]interface{}, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := updateFromMapAsDefault(layers_, parsedLayers, m, options...)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

func updateFromMap(
	layers_ *layers.ParameterLayers,
	parsedLayers *layers.ParsedLayers,
	m map[string]map[string]interface{},
	options ...parameters.ParseStepOption) error {
	for k, v := range m {
		layer, ok := layers_.Get(k)
		if !ok {
			continue
		}

		parsedLayer := parsedLayers.GetOrCreate(layer)
		ps, err := layer.GetParameterDefinitions().GatherParametersFromMap(v, true, options...)
		if err != nil {
			return err
		}
		parsedLayer.Parameters.Merge(ps)
	}
	return nil
}

func updateFromMapAsDefault(
	layers_ *layers.ParameterLayers,
	parsedLayers *layers.ParsedLayers,
	m map[string]map[string]interface{},
	options ...parameters.ParseStepOption) error {
	for k, v := range m {
		layer, ok := layers_.Get(k)
		if !ok {
			continue
		}

		parsedLayer := parsedLayers.GetOrCreate(layer)
		ps, err := layer.GetParameterDefinitions().GatherParametersFromMap(v, true, options...)
		if err != nil {
			return err
		}
		parsedLayer.Parameters.MergeAsDefault(ps)
	}
	return nil
}

func updateFromEnv(
	layers_ *layers.ParameterLayers,
	parsedLayers *layers.ParsedLayers,
	prefix string,
	options ...parameters.ParseStepOption,
) error {
	layers_.ForEach(func(key string, l layers.ParameterLayer) {
		parsedLayer := parsedLayers.GetOrCreate(l)
		pds := l.GetParameterDefinitions()
		pds.ForEach(func(p *parameters.ParameterDefinition) {
			name := p.Name
			if prefix != "" {
				name = prefix + "_" + name
			}
			name = strings.ToUpper(name)

			if v, ok := os.LookupEnv(name); ok {
				parsedLayer.Parameters.UpdateValue(name, p, v, options...)
			}
		})
	})

	return nil
}

func UpdateFromEnv(prefix string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromEnv(layers_, parsedLayers, prefix, options...)
		}
	}
}

func updateFromStringList(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers, prefix string, args []string, options ...parameters.ParseStepOption) error {
	err := layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
		parsedLayer := parsedLayers.GetOrCreate(l)
		pds := l.GetParameterDefinitions()
		ps, remainingArgs, err := pds.GatherFlagsFromStringList(args, true, true, prefix, options...)
		if err != nil {
			return err
		}

		parsedLayer.Parameters.Merge(ps)

		ps, err = pds.GatherArguments(remainingArgs, true, true, options...)
		if err != nil {
			return err
		}

		parsedLayer.Parameters.Merge(ps)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func UpdateFromStringList(prefix string, args []string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return updateFromStringList(layers_, parsedLayers, prefix, args, options...)
		}
	}
}
