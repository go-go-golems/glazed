package middlewares

import "github.com/go-go-golems/glazed/pkg/cmds/layers"

type HandlerFunc func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers)

type Middleware func(HandlerFunc) HandlerFunc
