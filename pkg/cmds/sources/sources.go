package sources

import (
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// Middleware is a type alias for middlewares.Middleware.
// A Middleware resolves values from a source (cobra, env, config, defaults).
type Middleware = cmd_middlewares.Middleware

// ParseOption is a type alias for parameters.ParseStepOption.
// Parse options annotate parsed values with source/metadata (useful for debugging and precedence tracking).
type ParseOption = parameters.ParseStepOption

// ConfigFileOption is a type alias for middlewares.ConfigFileOption.
// It configures how config files are parsed and mapped into layer values.
type ConfigFileOption = cmd_middlewares.ConfigFileOption

// ConfigFileMapper is a type alias for middlewares.ConfigFileMapper.
// It maps arbitrary config file structures into map[layerSlug]map[paramName]value.
type ConfigFileMapper = cmd_middlewares.ConfigFileMapper

// ConfigMapper is a type alias for middlewares.ConfigMapper (function-based or pattern-based mappers).
type ConfigMapper = cmd_middlewares.ConfigMapper

// SourceDefaults is the canonical source label for default values.
// It wraps parameters.SourceDefaults.
const SourceDefaults = parameters.SourceDefaults

// WithSource sets the parse-step source label (e.g. "env", "flags", "config", "defaults").
// It wraps parameters.WithParseStepSource.
func WithSource(source string) ParseOption {
	return parameters.WithParseStepSource(source)
}

// WithMetadata attaches metadata to the parse step (e.g. {"env_key":"APP_FOO"}).
// It wraps parameters.WithParseStepMetadata.
func WithMetadata(metadata map[string]interface{}) ParseOption {
	return parameters.WithParseStepMetadata(metadata)
}

// WithParseOptions configures parse-step options for config-file loading middlewares.
// It wraps middlewares.WithParseOptions.
func WithParseOptions(opts ...ParseOption) ConfigFileOption {
	return cmd_middlewares.WithParseOptions(opts...)
}

// WithConfigFileMapper provides a function-based mapper for config file structures.
// It wraps middlewares.WithConfigFileMapper.
func WithConfigFileMapper(mapper ConfigFileMapper) ConfigFileOption {
	return cmd_middlewares.WithConfigFileMapper(mapper)
}

// WithConfigMapper provides a mapper (function-based or pattern-based) for config file structures.
// It wraps middlewares.WithConfigMapper.
func WithConfigMapper(mapper ConfigMapper) ConfigFileOption {
	return cmd_middlewares.WithConfigMapper(mapper)
}

// FromCobra creates a middleware that parses values from Cobra command flags.
// It wraps middlewares.ParseFromCobraCommand.
func FromCobra(cmd *cobra.Command, opts ...ParseOption) Middleware {
	return cmd_middlewares.ParseFromCobraCommand(cmd, opts...)
}

// FromArgs creates a middleware that parses values from positional arguments.
// It wraps middlewares.GatherArguments.
func FromArgs(args []string, opts ...ParseOption) Middleware {
	return cmd_middlewares.GatherArguments(args, opts...)
}

// FromEnv creates a middleware that parses values from environment variables.
// It wraps middlewares.UpdateFromEnv.
// The prefix is typically the application name (e.g., "APP" becomes "APP_" prefix for env vars).
func FromEnv(prefix string, opts ...ParseOption) Middleware {
	return cmd_middlewares.UpdateFromEnv(prefix, opts...)
}

// FromDefaults creates a middleware that sets default values from field definitions.
// It wraps middlewares.SetFromDefaults.
func FromDefaults(opts ...ParseOption) Middleware {
	return cmd_middlewares.SetFromDefaults(opts...)
}

// FromFile creates a middleware that loads values from a config file (JSON or YAML).
// It wraps middlewares.LoadParametersFromFile.
// The config file should have the structure:
//
//	layer-slug:
//	  parameter-name: value
func FromFile(filename string, opts ...ConfigFileOption) Middleware {
	return cmd_middlewares.LoadParametersFromFile(filename, opts...)
}

// FromFiles creates a middleware that loads values from multiple config files (low -> high precedence).
// It wraps middlewares.LoadParametersFromFiles.
func FromFiles(files []string, opts ...ConfigFileOption) Middleware {
	return cmd_middlewares.LoadParametersFromFiles(files, opts...)
}

// FromMap creates a middleware that sets values from a custom map.
// It wraps middlewares.UpdateFromMap.
// The map structure is: map[layerSlug]map[parameterName]value
func FromMap(m map[string]map[string]interface{}, opts ...ParseOption) Middleware {
	return cmd_middlewares.UpdateFromMap(m, opts...)
}

// FromMapFirst creates a middleware that applies UpdateFromMap with first-apply semantics.
// It wraps middlewares.UpdateFromMapFirst.
func FromMapFirst(m map[string]map[string]interface{}, opts ...ParseOption) Middleware {
	return cmd_middlewares.UpdateFromMapFirst(m, opts...)
}

// FromMapAsDefault creates a middleware that updates values only when they are unset.
// It wraps middlewares.UpdateFromMapAsDefault.
func FromMapAsDefault(m map[string]map[string]interface{}, opts ...ParseOption) Middleware {
	return cmd_middlewares.UpdateFromMapAsDefault(m, opts...)
}

// FromMapAsDefaultFirst creates a middleware that applies UpdateFromMapAsDefault with first-apply semantics.
// It wraps middlewares.UpdateFromMapAsDefaultFirst.
func FromMapAsDefaultFirst(m map[string]map[string]interface{}, opts ...ParseOption) Middleware {
	return cmd_middlewares.UpdateFromMapAsDefaultFirst(m, opts...)
}

// Execute executes a chain of middlewares to resolve values from multiple sources.
// It wraps middlewares.ExecuteMiddlewares.
//
// Ordering note:
// ExecuteMiddlewares reverses the list internally; for the common "call next first" style middlewares,
// this means the *first* middleware you pass has the *highest* precedence (it runs last and can override),
// while the *last* middleware you pass has the *lowest* precedence (it runs first).
//
// A typical precedence chain (lowest -> highest) is:
//
//	defaults < config files < programmatic map < environment < cobra flags/args
//
// To achieve that, pass middlewares in the *reverse* order:
//
//	FromCobra, FromEnv, FromMap, FromFile/FromFiles, FromDefaults
func Execute(schema_ *schema.Schema, vals *values.Values, ms ...Middleware) error {
	return cmd_middlewares.ExecuteMiddlewares(schema_, vals, ms...)
}
