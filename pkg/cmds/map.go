package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

func ParseCommandFromMap(layers_ *layers.ParameterLayers, m map[string]interface{}) (
	*layers.ParsedLayers,
	error,
) {
	parsedLayers := layers.NewParsedLayers()

	// we now need to map the individual values in the JSON to the parsed layers as well
	err := layers_.ForEachE(func(key string, layer layers.ParameterLayer) error {
		pds := layer.GetParameterDefinitions()
		ps_, err := pds.GatherParametersFromMap(m, false)
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
