package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// Middleware is a type alias for middlewares.Middleware.
// A Middleware resolves values from a source (cobra, env, config, defaults).
type Middleware = cmd_middlewares.Middleware

// FromCobra creates a middleware that parses values from Cobra command flags.
// It wraps middlewares.ParseFromCobraCommand.
func FromCobra(cmd *cobra.Command, opts ...parameters.ParseStepOption) Middleware {
	return cmd_middlewares.ParseFromCobraCommand(cmd, opts...)
}

// FromArgs creates a middleware that parses values from positional arguments.
// It wraps middlewares.GatherArguments.
func FromArgs(args []string, opts ...parameters.ParseStepOption) Middleware {
	return cmd_middlewares.GatherArguments(args, opts...)
}

// FromEnv creates a middleware that parses values from environment variables.
// It wraps middlewares.UpdateFromEnv.
// The prefix is typically the application name (e.g., "APP" becomes "APP_" prefix for env vars).
func FromEnv(prefix string, opts ...parameters.ParseStepOption) Middleware {
	return cmd_middlewares.UpdateFromEnv(prefix, opts...)
}

// FromDefaults creates a middleware that sets default values from field definitions.
// It wraps middlewares.SetFromDefaults.
func FromDefaults(opts ...parameters.ParseStepOption) Middleware {
	return cmd_middlewares.SetFromDefaults(opts...)
}

// Execute executes a chain of middlewares to resolve values from multiple sources.
// It wraps middlewares.ExecuteMiddlewares.
// The middlewares are executed in reverse order (first middleware's handler is called first),
// which means values are resolved with the following precedence (lowest to highest):
// 1. Defaults (if FromDefaults is first)
// 2. Config files (if added)
// 3. Environment variables (if FromEnv is added)
// 4. Cobra flags/args (if FromCobra/FromArgs are last, highest precedence)
func Execute(schema *schema.Schema, vals *values.Values, ms ...Middleware) error {
	return cmd_middlewares.ExecuteMiddlewares((*layers.ParameterLayers)(schema), (*layers.ParsedLayers)(vals), ms...)
}
