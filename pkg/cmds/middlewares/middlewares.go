package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type HandlerFunc func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error

type Middleware func(HandlerFunc) HandlerFunc

// layer middlewares:
// - whitelist
// - blacklist
// - override
// - set default
// - fill from json
// - from parameter definition defaults

// SetFromDefaults fills the parsedLayers with all default values from the layers parameter definitions
// for values that were not set, and then calls next.
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

				pds.ForEach(func(pd *parameters.ParameterDefinition) {
					if pd.Default != nil {
						parsedLayer.Parameters.SetAsDefault(pd.Name, pd, *pd.Default, options...)
					}
				})

				return nil
			})
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func UpdateFromMap(m map[string]map[string]interface{}, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
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
	}
}

func RestrictLayers(slugs []string, m Middleware) HandlerFunc {
	slugsToDelete := map[string]interface{}{}
	for _, s := range slugs {
		slugsToDelete[s] = nil
	}
	return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
		layers_.ForEach(func(key string, l layers.ParameterLayer) {
			if slugsToDelete[key] != nil {
				layers_.Delete(key)
			}
		})
		return nil
	}
}

func Identity(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
	return nil
}

func WrapWithLayerModifyingHandler(m HandlerFunc, nextMiddlewares ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			var chain Middleware
			chain = Chain(nextMiddlewares...)

			clonedLayers := layers_.Clone()
			err := m(clonedLayers, parsedLayers)
			if err != nil {
				return err
			}

			err = chain(Identity)(clonedLayers, parsedLayers)
			if err != nil {
				return err
			}

			err = next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func Chain(ms ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for _, m_ := range ms {
			next = m_(next)
		}
		return next
	}
}

// ExecuteMiddlewares executes a chain of middlewares with the given parameters.
// It starts with an initial empty handler, then iteratively wraps it with each middleware.
// Finally, it calls the resulting handler with the provided layers and parsedLayers.
//
// Middlewares basically get executed in the order they are provided.
// [f1, f2, f3] will be executed as f3(f2(f1(handler)))
//
// How they call the next handler is up to them, but they should always call it.
//
// Usually, the following rules of thumbs work well
//   - if all you do is modify the parsedLayers, call `next` first.
//     This means that parsedLayers will be modified in the order of the middlewares.
//     For example, executeMiddlewares(SetFromArgs(), SetFromEnv(), SetFromDefaults())
//     will first set the defaults, then the environment value, and finally the command line arguments.
//   - if you want to modify the layers before parsing, use the
//     call `next` last. This means that the lower level firmwares
func ExecuteMiddlewares(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers, middlewares ...Middleware) error {
	handler := Identity
	for _, m_ := range middlewares {
		handler = m_(handler)
	}

	return handler(layers_, parsedLayers)
}
