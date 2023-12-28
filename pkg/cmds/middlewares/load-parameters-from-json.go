package middlewares

import (
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
	"os"
)

// LoadParametersFromJSON loads parameter definitions from a JSON file and applies them to the parameter layers.
// It returns a middleware function that can be used to process the parameter layers.
//
// The JSON file path is specified by the `jsonFile` parameter.
// The `options` parameter allows for additional customization of the parsing process.
//
// The middleware function takes two parameters:
//   - `next`: a handler function that will be called after the parameter layers are processed.
//   - `layers_`: the parameter layers to be processed.
//   - `parsedLayers`: the parsed layers containing the parameter definitions.
//
// The middleware function reads the JSON file and unmarshals it into a map.
// It then iterates over each parameter layer in the `layers_` parameter, gathering the corresponding parameter definitions from the map and applying them to the parsed layers.
// The parsed layers are then passed to the `next` handler function.
//
// If any error occurs during the loading or parsing process, it is returned by the middleware function.
//
// Example usage:
//
//	jsonFile := "parameters.json"
//	options := // define options if needed
//	middleware := LoadParametersFromJSON(jsonFile, options)
//	handler := middleware(nextHandler)
//	err := handler(parameterLayers, parsedLayers)
//	if err != nil {
//	    // handle error
//	}
func LoadParametersFromJSON(jsonFile string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			m := map[string]interface{}{}
			bytes, err := os.ReadFile(jsonFile)
			if err != nil {
				return err
			}
			err = json.Unmarshal(bytes, &m)
			if err != nil {
				cobra.CheckErr(err)
			}

			err = layers_.ForEachE(func(key string, layer layers.ParameterLayer) error {
				pds := layer.GetParameterDefinitions()
				ps_, err := pds.GatherParametersFromMap(m, false, options...)
				if err != nil {
					return err
				}
				parsedLayer := parsedLayers.GetOrCreate(layer)
				parsedLayer.Parameters.Merge(ps_)

				return nil
			})

			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

type Profile map[string]map[string]interface{}

func SetParametersFromProfile(profile Profile, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			for k, v := range profile {
				layer, ok := layers_.Get(k)
				if !ok {
					continue
				}

				pds := layer.GetParameterDefinitions()
				ps_, err := pds.GatherParametersFromMap(v, false, options...)
				if err != nil {
					return err
				}

				parsedLayer := parsedLayers.GetOrCreate(layer)
				parsedLayer.Parameters.Merge(ps_)
			}

			return next(layers_, parsedLayers)
		}
	}
}
