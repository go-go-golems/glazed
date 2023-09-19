package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

func ParseCommandFromMap(description *CommandDescription, m map[string]interface{}) (
	map[string]*layers.ParsedParameterLayer,
	map[string]interface{},
	error,
) {
	parsedLayers := map[string]*layers.ParsedParameterLayer{}
	ps := map[string]interface{}{}

	// we now need to map the individual values in the JSON to the parsed layers as well
	for _, layer := range description.Layers {
		jsonParameterLayer, ok := layer.(layers.JSONParameterLayer)
		if !ok {
			err := errors.Errorf("layer %s is not a JSONParameterLayer", layer.GetName())
			return nil, nil, err
		}

		ps_, err := jsonParameterLayer.ParseFlagsFromJSON(m, false)
		if err != nil {
			return nil, nil, err
		}
		parsedLayers[layer.GetSlug()] = &layers.ParsedParameterLayer{
			Layer:      layer,
			Parameters: ps_,
		}

		for k, v := range ps_ {
			ps[k] = v
		}
	}

	ps_, err := parameters.GatherParametersFromMap(m, description.GetFlagMap(), false)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}

	// this should check for required arguments and fill them, but at this points it's already too late...
	ps_, err = parameters.GatherParametersFromMap(m, description.GetArgumentMap(), false)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}

	return parsedLayers, ps, nil
}
