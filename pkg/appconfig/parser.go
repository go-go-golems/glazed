package appconfig

import (
	"io"
	"reflect"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
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

	const stubCommandName = "appconfig-parser"
	desc := cmds.NewCommandDescription(stubCommandName)
	for _, r := range p.regs {
		desc.Layers.Set(string(r.slug), r.layer)
	}

	var parsedLayers *layers.ParsedLayers
	if p.opts.useCobra {
		// Cobra mode: build explicit middleware chain with flags/args at highest precedence.
		middlewares_ := []cmd_middlewares.Middleware{}

		// Additional middlewares (advanced escape hatch) — kept consistent with runner behavior:
		// additional middlewares are highest precedence (added first; executed last due to reverse execution).
		if len(p.opts.additionalMiddlewares) > 0 {
			middlewares_ = append(middlewares_, p.opts.additionalMiddlewares...)
		}

		// Highest precedence first (because middlewares execute in reverse).
		middlewares_ = append(middlewares_,
			cmd_middlewares.ParseFromCobraCommand(
				p.opts.cobraCmd,
				parameters.WithParseStepSource("cobra"),
			),
		)
		middlewares_ = append(middlewares_,
			cmd_middlewares.GatherArguments(
				p.opts.cobraArgs,
				parameters.WithParseStepSource("arguments"),
			),
		)

		// Env
		if p.opts.useEnv {
			middlewares_ = append(middlewares_,
				cmd_middlewares.UpdateFromEnv(
					p.opts.envPrefix,
					parameters.WithParseStepSource("env"),
				),
			)
		}
		// Config files (low -> high precedence)
		if len(p.opts.configFiles) > 0 {
			middlewares_ = append(middlewares_,
				cmd_middlewares.LoadParametersFromFiles(
					p.opts.configFiles,
					cmd_middlewares.WithParseOptions(parameters.WithParseStepSource("config")),
				),
			)
		}
		// Provided values
		if p.opts.valuesForLayers != nil {
			middlewares_ = append(middlewares_,
				cmd_middlewares.UpdateFromMap(
					p.opts.valuesForLayers,
					parameters.WithParseStepSource("provided-values"),
				),
			)
		}
		// Defaults (lowest)
		middlewares_ = append(middlewares_,
			cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
		)

		parsedLayers = layers.NewParsedLayers()
		if err := cmd_middlewares.ExecuteMiddlewares(desc.Layers, parsedLayers, middlewares_...); err != nil {
			return nil, errors.Wrap(err, "failed to parse parameters (cobra mode)")
		}
	} else {
		// Runner mode: reuse runner.ParseCommandParameters (defaults/config/env/provided-values).
		cmd := &stubCommand{desc: desc}
		parseOpts := []runner.ParseOption{}
		parseOpts = append(parseOpts, p.opts.runnerParseOptions...)
		if len(p.opts.additionalMiddlewares) > 0 {
			parseOpts = append(parseOpts, runner.WithAdditionalMiddlewares(p.opts.additionalMiddlewares...))
		}
		if p.opts.useEnv {
			parseOpts = append(parseOpts, runner.WithEnvMiddleware(p.opts.envPrefix))
		}
		if len(p.opts.configFiles) > 0 {
			parseOpts = append(parseOpts, runner.WithConfigFiles(p.opts.configFiles...))
		}
		if p.opts.valuesForLayers != nil {
			parseOpts = append(parseOpts, runner.WithValuesForLayers(p.opts.valuesForLayers))
		}

		var err error
		parsedLayers, err = runner.ParseCommandParameters(cmd, parseOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse parameters")
		}
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

// stubCommand exists so we can reuse runner.ParseCommandParameters without
// depending on any specific command implementation.
type stubCommand struct {
	desc *cmds.CommandDescription
}

var _ cmds.Command = (*stubCommand)(nil)

func (s *stubCommand) Description() *cmds.CommandDescription {
	return s.desc
}

func (s *stubCommand) ToYAML(w io.Writer) error {
	return s.desc.ToYAML(w)
}
