package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

type HandlerFunc func(layers *schema.Schema, parsedLayers *values.Values) error

type Middleware func(HandlerFunc) HandlerFunc

// layer middlewares:
// - [x] whitelist (layers, parameters)
// - [x] blacklist (layers, parameters)
// - [x] override (updateFromMap)
// - [ ] set defaults explicitly
// - [x] fill from json (updateFromMap)
// - [x] from parameter definition defaults
// - [x] fill from cobra (flags, arguments)
// - [x] fill from viper

func Identity(layers_ *schema.Schema, parsedLayers *values.Values) error {
	return nil
}

// Chain chains together a list of middlewares into a single middleware.
// It does this by iteratively wrapping each middleware around the next handler.
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
// Middlewares basically get executed in the reverse order they are provided,
// which means the first given middleware's handler will be called first.
//
// [f1, f2, f3] will be executed as f1(f2(f3(handler)))(layers_, parsedLayers).
//
// How they call the next handler is up to them, but they should always call it.
//
// Usually, the following rules of thumbs work well
//   - if all you do is modify the parsedLayers, call `next` first.
//     This means that parsedLayers will be modified in the order of the middlewares.
//     For example, executeMiddlewares(SetFromArgs(), SetFromEnv(), SetFromDefaults())
//     will first set the defaults, then the environment value, and finally the command line arguments.
//   - if you want to modify the layers before parsing, use the
//     call `next` last. This means that the middlewares further down the list will
//     get the newly updated ParameterLayers and thus potentially restrict which parameters they parse.
func ExecuteMiddlewares(layers_ *schema.Schema, parsedLayers *values.Values, middlewares ...Middleware) error {
	handler := Identity
	reversedMiddlewares := make([]Middleware, len(middlewares))
	for i, m_ := range middlewares {
		reversedMiddlewares[len(middlewares)-1-i] = m_
	}
	for _, m_ := range reversedMiddlewares {
		handler = m_(handler)
	}

	clonedLayers := layers_.Clone()
	return handler(clonedLayers, parsedLayers)
}
