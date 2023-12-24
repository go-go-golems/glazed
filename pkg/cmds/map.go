package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

func ParseCommandFromMap(description *CommandDescription, m map[string]interface{}) (
	*layers.ParsedLayers,
	error,
) {
	parsedLayers := layers.NewParsedLayers()

	// we now need to map the individual values in the JSON to the parsed layers as well
	err := description.Layers.ForEachE(func(key string, layer layers.ParameterLayer) error {
		jsonParameterLayer, ok := layer.(layers.JSONParameterLayer)
		if !ok {
			return errors.Errorf("layer %s is not a JSONParameterLayer", layer.GetName())
		}

		ps_, err := jsonParameterLayer.ParseFlagsFromJSON(m, false)
		if err != nil {
			return err
		}
		parsedLayers.Set(layer.GetSlug(), &layers.ParsedLayer{
			Layer:      layer,
			Parameters: ps_,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}
