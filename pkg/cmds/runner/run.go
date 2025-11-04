package runner

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
)

// RunOptions contains configuration for running a command
type RunOptions struct {
	Writer         io.Writer
	GlazeProcessor middlewares.Processor
}

type RunOption func(*RunOptions)

// WithWriter sets the writer to use for WriterCommand output
func WithWriter(w io.Writer) RunOption {
	return func(o *RunOptions) {
		o.Writer = w
	}
}

// WithGlazeProcessor sets the processor to use for GlazeCommand output
func WithGlazeProcessor(p middlewares.Processor) RunOption {
	return func(o *RunOptions) {
		o.GlazeProcessor = p
	}
}

// RunCommand executes a Glazed command with the given parsed parameters and options
func RunCommand(
	ctx context.Context,
	cmd cmds.Command,
	parsedLayers *layers.ParsedLayers,
	options ...RunOption,
) error {
	// Setup default options
	opts := &RunOptions{
		Writer: os.Stdout,
	}

	// Apply provided options
	for _, opt := range options {
		opt(opts)
	}

	// Handle different command types
	switch c := cmd.(type) {
	case cmds.BareCommand:
		return c.Run(ctx, parsedLayers)

	case cmds.WriterCommand:
		return c.RunIntoWriter(ctx, parsedLayers, opts.Writer)

	case cmds.GlazeCommand:
		// If no processor is provided, create one from glazed settings
		if opts.GlazeProcessor == nil {
			glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
			if !ok {
				return fmt.Errorf("glazed layer not found")
			}
			gp, err := settings.SetupTableProcessor(glazedLayer)
			if err != nil {
				return fmt.Errorf("failed to setup processor: %w", err)
			}
			_, err = settings.SetupProcessorOutput(gp, glazedLayer, opts.Writer)
			if err != nil {
				return fmt.Errorf("failed to setup processor output: %w", err)
			}
			opts.GlazeProcessor = gp
		}

		err := c.RunIntoGlazeProcessor(ctx, parsedLayers, opts.GlazeProcessor)
		if err != nil {
			return err
		}

		return opts.GlazeProcessor.Close(ctx)

	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

// ParseOptions contains configuration for parameter parsing
type ParseOptions struct {
	ValuesForLayers       map[string]map[string]interface{}
	EnvPrefix             string
	AdditionalMiddlewares []cmd_middlewares.Middleware
	UseViper              bool
	UseEnv                bool
	ConfigFiles           []string
}

type ParseOption func(*ParseOptions)

// WithValuesForLayers sets values for parameters in specified layers
func WithValuesForLayers(values map[string]map[string]interface{}) ParseOption {
	return func(o *ParseOptions) {
		o.ValuesForLayers = values
	}
}

// WithEnvPrefix sets the prefix for environment variable parsing
func WithEnvPrefix(prefix string) ParseOption {
	return func(o *ParseOptions) {
		o.EnvPrefix = prefix
	}
}

// WithAdditionalMiddlewares adds custom middlewares to the parsing chain
func WithAdditionalMiddlewares(middlewares ...cmd_middlewares.Middleware) ParseOption {
	return func(o *ParseOptions) {
		o.AdditionalMiddlewares = append(o.AdditionalMiddlewares, middlewares...)
	}
}

// WithViper enables loading parameters from Viper configuration
func WithViper() ParseOption {
	return func(o *ParseOptions) {
		o.UseViper = true
	}
}

// WithConfigFiles applies a list of config files (low -> high precedence)
func WithConfigFiles(files ...string) ParseOption {
	return func(o *ParseOptions) {
		o.ConfigFiles = append(o.ConfigFiles, files...)
	}
}

// WithEnvMiddleware enables environment variable parsing with the given prefix
func WithEnvMiddleware(prefix string) ParseOption {
	return func(o *ParseOptions) {
		o.UseEnv = true
		o.EnvPrefix = prefix
	}
}

// ParseCommandParameters parses parameters for a command using a configurable middleware chain
func ParseCommandParameters(
	cmd cmds.Command,
	options ...ParseOption,
) (*layers.ParsedLayers, error) {
	opts := &ParseOptions{}

	// Apply provided options
	for _, opt := range options {
		opt(opts)
	}

	// Create middleware chain
	middlewares_ := []cmd_middlewares.Middleware{}
	middlewares_ = append(middlewares_, opts.AdditionalMiddlewares...)

	// Deprecated: Viper support is removed; Use WithConfigFiles/WithEnvMiddleware instead

	// Add environment variables middleware if enabled
	if opts.UseEnv {
		middlewares_ = append(middlewares_,
			cmd_middlewares.UpdateFromEnv(opts.EnvPrefix,
				parameters.WithParseStepSource("env"),
			),
		)
	}

	// Add config files middleware if provided
	if len(opts.ConfigFiles) > 0 {
		middlewares_ = append(middlewares_,
			cmd_middlewares.LoadParametersFromFiles(opts.ConfigFiles,
				cmd_middlewares.WithParseOptions(
					parameters.WithParseStepSource("config"),
				),
			),
		)
	}

	// Add values for layers middleware if provided
	if opts.ValuesForLayers != nil {
		middlewares_ = append(middlewares_,
			cmd_middlewares.UpdateFromMap(opts.ValuesForLayers,
				parameters.WithParseStepSource("provided-values"),
			),
		)
	}

	// Add base defaults middleware
	middlewares_ = append(middlewares_,
		cmd_middlewares.SetFromDefaults(
			parameters.WithParseStepSource(parameters.SourceDefaults),
		),
	)

	// Create parsed layers and execute middleware chain
	parsedLayers := layers.NewParsedLayers()
	err := cmd_middlewares.ExecuteMiddlewares(
		cmd.Description().Layers,
		parsedLayers,
		middlewares_...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	return parsedLayers, nil
}

// ParseAndRun combines parameter parsing and command execution into a single function
func ParseAndRun(
	ctx context.Context,
	cmd cmds.Command,
	parseOptions []ParseOption,
	runOptions []RunOption,
) error {
	parsedLayers, err := ParseCommandParameters(cmd, parseOptions...)
	if err != nil {
		return err
	}

	return RunCommand(ctx, cmd, parsedLayers, runOptions...)
}
