package appconfig

import (
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type parserOptions struct {
	// middlewares are collected in the order options are applied.
	//
	// IMPORTANT: The recommended interpretation is that earlier options are lower
	// precedence and later options are higher precedence (i.e. last wins).
	middlewares []cmd_middlewares.Middleware
}

// ParserOption configures a Parser.
type ParserOption func(*parserOptions) error

// WithDefaults appends the defaults middleware (lowest precedence in the typical chain).
func WithDefaults() ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares,
			cmd_middlewares.SetFromDefaults(
				parameters.WithParseStepSource(parameters.SourceDefaults),
			),
		)
		return nil
	}
}

// WithEnv enables parsing from environment variables using the given prefix.
func WithEnv(prefix string) ParserOption {
	return func(o *parserOptions) error {
		if prefix == "" {
			return errors.New("env prefix must not be empty")
		}
		o.middlewares = append(o.middlewares,
			cmd_middlewares.UpdateFromEnv(
				prefix,
				parameters.WithParseStepSource("env"),
			),
		)
		return nil
	}
}

// WithConfigFiles configures config files to load (low -> high precedence).
func WithConfigFiles(files ...string) ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares,
			cmd_middlewares.LoadParametersFromFiles(
				files,
				cmd_middlewares.WithParseOptions(parameters.WithParseStepSource("config")),
			),
		)
		return nil
	}
}

// WithValuesForLayers configures programmatic values for layers (optional).
func WithValuesForLayers(values map[string]map[string]interface{}) ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares,
			cmd_middlewares.UpdateFromMap(
				values,
				parameters.WithParseStepSource("provided-values"),
			),
		)
		return nil
	}
}

// WithMiddlewares injects additional middlewares into the parse chain.
//
// NOTE: Middleware ordering is subtle; this is an escape hatch for advanced usage.
func WithMiddlewares(middlewares ...cmd_middlewares.Middleware) ParserOption {
	return func(o *parserOptions) error {
		o.middlewares = append(o.middlewares, middlewares...)
		return nil
	}
}

// WithCobra configures the Parser to read flags and positional arguments from a Cobra command.
//
// The caller is responsible for ensuring Cobra has parsed the args (i.e. this is
// used from within a cobra Run/RunE/PreRun hook, or after Execute has parsed).
func WithCobra(cmd *cobra.Command, args []string) ParserOption {
	return func(o *parserOptions) error {
		if cmd == nil {
			return errors.New("cobra command must not be nil")
		}
		// GatherArguments is lower precedence than ParseFromCobraCommand (flags).
		o.middlewares = append(o.middlewares,
			cmd_middlewares.GatherArguments(
				append([]string(nil), args...),
				parameters.WithParseStepSource("arguments"),
			),
			cmd_middlewares.ParseFromCobraCommand(
				cmd,
				parameters.WithParseStepSource("cobra"),
			),
		)
		return nil
	}
}
