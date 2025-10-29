package middlewares

import (
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
			err = layers_.UpdateWithDefaults(parsedLayers, options...)
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
		_, err = parsedLayer.Parameters.Merge(ps)
		if err != nil {
			return err
		}
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
		_, err = parsedLayer.Parameters.MergeAsDefault(ps)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateFromEnv(
	layers_ *layers.ParameterLayers,
	parsedLayers *layers.ParsedLayers,
	prefix string,
	options ...parameters.ParseStepOption,
) error {
	err := layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
		parsedLayer := parsedLayers.GetOrCreate(l)
		pds := l.GetParameterDefinitions()
		layerPrefix := l.GetPrefix()
		err := pds.ForEachE(func(p *parameters.ParameterDefinition) error {
			// Compute env key based on layer prefix + param name, hyphen->underscore, uppercase,
			// and optional global prefix (app name) separated by underscore.
			base := layerPrefix + p.Name
			envKey := strings.ToUpper(strings.ReplaceAll(base, "-", "_"))
			if prefix != "" {
				envKey = strings.ToUpper(prefix) + "_" + envKey
			}

			if v, ok := os.LookupEnv(envKey); ok {
				opts := append([]parameters.ParseStepOption{
					parameters.WithParseStepMetadata(map[string]interface{}{
						"env_key": envKey,
					}),
				}, options...)
				// Store under the logical parameter name within the layer
				if err := parsedLayer.Parameters.UpdateValue(p.Name, p, v, opts...); err != nil {
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
