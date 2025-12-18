package appconfig

import (
	"reflect"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/pkg/errors"
)

type registration[T any] struct {
	slug  LayerSlug
	layer layers.ParameterLayer
	bind  func(*T) any
}

// LayerSlug is a distinct type to encourage declaring layer slugs as constants.
//
// Example:
//
//	const RedisSlug appconfig.LayerSlug = "redis"
type LayerSlug string

// Parser is an incremental config boundary:
// - callers register layers and bind them to sub-struct pointers inside T
// - Parse executes a configurable middleware chain and returns a populated T.
//
// V1 hydration uses ParsedLayers.InitializeStruct, which means fields are only
// populated when the destination structs have explicit `glazed.parameter` tags.
type Parser[T any] struct {
	opts parserOptions
	regs []registration[T]
}

// NewParser constructs a Parser for the grouped settings struct type T.
func NewParser[T any](options ...ParserOption) (*Parser[T], error) {
	p := &Parser[T]{
		opts: parserOptions{},
	}
	for _, opt := range options {
		if opt == nil {
			continue
		}
		if err := opt(&p.opts); err != nil {
			return nil, errors.Wrap(err, "failed to apply ParserOption")
		}
	}
	return p, nil
}

// Register associates a layer slug and ParameterLayer with a binder that returns
// a pointer to the corresponding sub-struct inside the grouped settings struct T.
//
// Invariants:
// - slug must be non-empty and unique
// - layer must be non-nil
// - bind must be non-nil
// - slug must match layer.GetSlug() (to avoid mismatches between registration keys and parsed layer keys)
func (p *Parser[T]) Register(slug LayerSlug, layer layers.ParameterLayer, bind func(*T) any) error {
	if slug == "" {
		return errors.New("slug must not be empty")
	}
	if layer == nil {
		return errors.New("layer must not be nil")
	}
	if bind == nil {
		return errors.New("bind must not be nil")
	}
	if layer.GetSlug() != string(slug) {
		return errors.Errorf("slug %q does not match layer.GetSlug() %q", string(slug), layer.GetSlug())
	}
	for _, r := range p.regs {
		if r.slug == slug {
			return errors.Errorf("layer slug %q already registered", string(slug))
		}
	}
	p.regs = append(p.regs, registration[T]{slug: slug, layer: layer, bind: bind})
	return nil
}

// Parse runs the configured middleware chain and returns a populated T.
func (p *Parser[T]) Parse() (*T, error) {
	if len(p.regs) == 0 {
		return nil, errors.New("no layers registered")
	}

	paramLayers := layers.NewParameterLayers()
	for _, r := range p.regs {
		paramLayers.Set(string(r.slug), r.layer)
	}

	if len(p.opts.middlewares) == 0 {
		return nil, errors.New("no parsing sources configured (pass WithDefaults/WithConfigFiles/WithEnv/WithCobra/...)")
	}

	// Options collect middlewares in low->high precedence order (last wins).
	// ExecuteMiddlewares expects the reverse order.
	execMiddlewares := make([]cmd_middlewares.Middleware, 0, len(p.opts.middlewares))
	for i := len(p.opts.middlewares) - 1; i >= 0; i-- {
		execMiddlewares = append(execMiddlewares, p.opts.middlewares[i])
	}

	parsedLayers := layers.NewParsedLayers()
	if err := cmd_middlewares.ExecuteMiddlewares(paramLayers, parsedLayers, execMiddlewares...); err != nil {
		return nil, errors.Wrap(err, "failed to parse parameters")
	}

	var t T
	for _, r := range p.regs {
		dst := r.bind(&t)
		if dst == nil {
			return nil, errors.Errorf("bind returned nil for layer %q", string(r.slug))
		}
		v := reflect.ValueOf(dst)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			return nil, errors.Errorf("bind for layer %q must return a non-nil pointer, got %T", string(r.slug), dst)
		}
		if err := parsedLayers.InitializeStruct(string(r.slug), dst); err != nil {
			return nil, errors.Wrapf(err, "failed to initialize settings for layer %q", string(r.slug))
		}
	}

	return &t, nil
}
