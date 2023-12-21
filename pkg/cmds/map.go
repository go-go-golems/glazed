package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

func ParseCommandFromMap(description *CommandDescription, m map[string]interface{}) (
	map[string]*layers.ParsedParameterLayer,
	error,
) {
	parsedLayers := map[string]*layers.ParsedParameterLayer{}

	// we now need to map the individual values in the JSON to the parsed layers as well
	for _, layer := range description.Layers {
		jsonParameterLayer, ok := layer.(layers.JSONParameterLayer)
		if !ok {
			err := errors.Errorf("layer %s is not a JSONParameterLayer", layer.GetName())
			return nil, err
		}

		ps_, err := jsonParameterLayer.ParseFlagsFromJSON(m, false)
		if err != nil {
			return nil, err
		}
		parsedLayers[layer.GetSlug()] = &layers.ParsedParameterLayer{
			Layer:      layer,
			Parameters: ps_,
		}
	}

	return parsedLayers, nil
}
