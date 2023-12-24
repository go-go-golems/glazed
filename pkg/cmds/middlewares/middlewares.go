package middlewares

import "github.com/go-go-golems/glazed/pkg/cmds/layers"

type Handler func(layer *layers.ParameterLayer)
