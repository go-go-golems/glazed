package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

type HandlerFunc func(schema_ *schema.Schema, parsedValues *values.Values) error

type Middleware func(HandlerFunc) HandlerFunc

// section middlewares:
// - [x] whitelist (sections, fields)
// - [x] blacklist (sections, fields)
// - [x] override (updateFromMap)
// - [ ] set defaults explicitly
// - [x] fill from json (updateFromMap)
// - [x] from field definition defaults
// - [x] fill from cobra (flags, arguments)
// - [x] fill from viper

func Identity(schema_ *schema.Schema, parsedValues *values.Values) error {
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

// Execute executes a chain of middlewares with the given schema.
// It starts with an initial empty handler, then iteratively wraps it with each middleware.
// Finally, it calls the resulting handler with the provided schema and parsed values.
//
// Middlewares basically get executed in the reverse order they are provided,
// which means the first given middleware's handler will be called first.
//
// [f1, f2, f3] will be executed as f1(f2(f3(handler)))(schema_, parsedValues).
//
// How they call the next handler is up to them, but they should always call it.
//
// Usually, the following rules of thumbs work well
//   - if all you do is modify parsed values, call `next` first.
//     This means parsed values will be modified in the order of the middlewares.
//     For example, executeMiddlewares(SetFromArgs(), SetFromEnv(), FromDefaults())
//     will first set the defaults, then the environment value, and finally the command line arguments.
//   - if you want to modify the schema before parsing,
//     call `next` last. This means that the middlewares further down the list will
//     get the newly updated schema and thus potentially restrict which fields they parse.
func Execute(schema_ *schema.Schema, parsedValues *values.Values, middlewares ...Middleware) error {
	handler := Identity
	reversedMiddlewares := make([]Middleware, len(middlewares))
	for i, m_ := range middlewares {
		reversedMiddlewares[len(middlewares)-1-i] = m_
	}
	for _, m_ := range reversedMiddlewares {
		handler = m_(handler)
	}

	clonedSchema := schema_.Clone()
	return handler(clonedSchema, parsedValues)
}
