package runner

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	cmd_sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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

// RunCommand executes a Glazed command with the given parsed values and options.
func RunCommand(
	ctx context.Context,
	cmd cmds.Command,
	parsedValues *values.Values,
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
		return c.Run(ctx, parsedValues)

	case cmds.WriterCommand:
		return c.RunIntoWriter(ctx, parsedValues, opts.Writer)

	case cmds.GlazeCommand:
		// If no processor is provided, create one from glazed settings
		if opts.GlazeProcessor == nil {
			glazedSection, ok := parsedValues.Get(settings.GlazedSlug)
			if !ok {
				return fmt.Errorf("glazed section not found")
			}
			gp, err := settings.SetupTableProcessor(glazedSection)
			if err != nil {
				return fmt.Errorf("failed to setup processor: %w", err)
			}
			_, err = settings.SetupProcessorOutput(gp, glazedSection, opts.Writer)
			if err != nil {
				return fmt.Errorf("failed to setup processor output: %w", err)
			}
			opts.GlazeProcessor = gp
		}

		err := c.RunIntoGlazeProcessor(ctx, parsedValues, opts.GlazeProcessor)
		if err != nil {
			return err
		}

		return opts.GlazeProcessor.Close(ctx)

	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

// ParseOptions contains configuration for value parsing.
type ParseOptions struct {
	ValuesForSections     map[string]map[string]interface{}
	EnvPrefix             string
	AdditionalMiddlewares []cmd_sources.Middleware
	UseViper              bool
	UseEnv                bool
	ConfigFiles           []string
}

type ParseOption func(*ParseOptions)

// WithValuesForSections sets values for fields in specified sections.
func WithValuesForSections(values map[string]map[string]interface{}) ParseOption {
	return func(o *ParseOptions) {
		o.ValuesForSections = values
	}
}

// WithEnvPrefix sets the prefix for environment variable parsing
func WithEnvPrefix(prefix string) ParseOption {
	return func(o *ParseOptions) {
		o.EnvPrefix = prefix
	}
}

// WithAdditionalMiddlewares adds custom middlewares to the parsing chain
func WithAdditionalMiddlewares(middlewares ...cmd_sources.Middleware) ParseOption {
	return func(o *ParseOptions) {
		o.AdditionalMiddlewares = append(o.AdditionalMiddlewares, middlewares...)
	}
}

// WithViper enables loading values from Viper configuration.
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

// ParseCommandValues parses values for a command using a configurable middleware chain.
func ParseCommandValues(
	cmd cmds.Command,
	options ...ParseOption,
) (*values.Values, error) {
	opts := &ParseOptions{}

	// Apply provided options
	for _, opt := range options {
		opt(opts)
	}

	// Create middleware chain
	middlewares_ := []cmd_sources.Middleware{}
	middlewares_ = append(middlewares_, opts.AdditionalMiddlewares...)

	// Deprecated: Viper support is removed; Use WithConfigFiles/WithEnvMiddleware instead

	// Add environment variables middleware if enabled
	if opts.UseEnv {
		middlewares_ = append(middlewares_,
			cmd_sources.FromEnv(opts.EnvPrefix,
				fields.WithSource("env"),
			),
		)
	}

	// Add config files middleware if provided
	if len(opts.ConfigFiles) > 0 {
		middlewares_ = append(middlewares_,
			cmd_sources.FromFiles(opts.ConfigFiles,
				cmd_sources.WithParseOptions(
					fields.WithSource("config"),
				),
			),
		)
	}

	// Add values for sections middleware if provided
	if opts.ValuesForSections != nil {
		middlewares_ = append(middlewares_,
			cmd_sources.FromMap(opts.ValuesForSections,
				fields.WithSource("provided-values"),
			),
		)
	}

	// Add base defaults middleware
	middlewares_ = append(middlewares_,
		cmd_sources.FromDefaults(
			fields.WithSource(fields.SourceDefaults),
		),
	)

	// Create parsed values and execute middleware chain
	parsedValues := values.New()
	err := cmd_sources.Execute(
		cmd.Description().Layers,
		parsedValues,
		middlewares_...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse values: %w", err)
	}

	return parsedValues, nil
}

// ParseAndRun combines value parsing and command execution into a single function.
func ParseAndRun(
	ctx context.Context,
	cmd cmds.Command,
	parseOptions []ParseOption,
	runOptions []RunOption,
) error {
	parsedValues, err := ParseCommandValues(cmd, parseOptions...)
	if err != nil {
		return err
	}

	return RunCommand(ctx, cmd, parsedValues, runOptions...)
}
