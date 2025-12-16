package appconfig

import (
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/pkg/errors"
)

type parserOptions struct {
	useEnv                bool
	envPrefix             string
	configFiles           []string
	valuesForLayers       map[string]map[string]interface{}
	additionalMiddlewares []cmd_middlewares.Middleware

	// Escape hatch for advanced callers (merged into Parse() options).
	runnerParseOptions []runner.ParseOption
}

// ParserOption configures a Parser.
type ParserOption func(*parserOptions) error

// WithEnv enables parsing from environment variables using the given prefix.
func WithEnv(prefix string) ParserOption {
	return func(o *parserOptions) error {
		if prefix == "" {
			return errors.New("env prefix must not be empty")
		}
		o.useEnv = true
		o.envPrefix = prefix
		return nil
	}
}

// WithConfigFiles configures config files to load (low -> high precedence).
func WithConfigFiles(files ...string) ParserOption {
	return func(o *parserOptions) error {
		o.configFiles = append(o.configFiles, files...)
		return nil
	}
}

// WithValuesForLayers configures programmatic values for layers (optional).
func WithValuesForLayers(values map[string]map[string]interface{}) ParserOption {
	return func(o *parserOptions) error {
		o.valuesForLayers = values
		return nil
	}
}

// WithMiddlewares injects additional middlewares into the parse chain.
//
// NOTE: Middleware ordering is subtle; this is an escape hatch for advanced usage.
func WithMiddlewares(middlewares ...cmd_middlewares.Middleware) ParserOption {
	return func(o *parserOptions) error {
		o.additionalMiddlewares = append(o.additionalMiddlewares, middlewares...)
		return nil
	}
}

// WithRunnerParseOptions appends raw runner.ParseOption values (advanced escape hatch).
func WithRunnerParseOptions(options ...runner.ParseOption) ParserOption {
	return func(o *parserOptions) error {
		o.runnerParseOptions = append(o.runnerParseOptions, options...)
		return nil
	}
}
