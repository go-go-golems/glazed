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

// FillFromDefaults fills the parsedLayers with all default values from the layers parameter definitions
// for values that were not set, and then calls next.
func FillFromDefaults(options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				pds := l.GetParameterDefinitions()
				parsedLayer, ok := parsedLayers.Get(key)
				if !ok {
					var err error
					parsedLayer, err = layers.NewParsedLayer(l)
					if err != nil {
						return err
					}
					parsedLayers.Set(key, parsedLayer)
				}

				pds.ForEach(func(pd *parameters.ParameterDefinition) {
					parsedLayer.Parameters.SetAsDefault(pd.Name, pd, pd.Default, options...)
				})

				return nil
			})
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

func ExecuteMiddlewares(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers, middlewares ...Middleware) error {
	var err error
	for _, m := range middlewares {
		err = m(func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			return nil
		})(layers_, parsedLayers)
		if err != nil {
			return err
		}
	}

	return nil
}
