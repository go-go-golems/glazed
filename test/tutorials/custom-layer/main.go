package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"glazed-logging-layer/logging"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ProcessDataCommand demonstrates using the logging layer
type ProcessDataCommand struct {
	*cmds.CommandDescription
}

type ProcessDataSettings struct {
	InputFile  string `glazed.parameter:"input-file"`
	OutputPath string `glazed.parameter:"output-path"`
	Workers    int    `glazed.parameter:"workers"`
	DryRun     bool   `glazed.parameter:"dry-run"`
}

func (c *ProcessDataCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging first
	if err := logging.InitializeLogging(parsedLayers); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	log.Info().Msg("Starting data processing command")

	// Get command settings
	settings := &ProcessDataSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	log.Debug().
		Str("input_file", settings.InputFile).
		Str("output_path", settings.OutputPath).
		Int("workers", settings.Workers).
		Bool("dry_run", settings.DryRun).
		Msg("Command settings parsed")

	// Simulate processing
	if settings.DryRun {
		log.Info().Msg("Dry run mode - no actual processing")
	} else {
		log.Info().Msg("Starting actual data processing")
	}

	// Simulate some work with progress logging
	for i := 0; i < settings.Workers; i++ {
		log.Info().Int("worker_id", i).Msg("Starting worker")

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		// Create result row
		row := types.NewRow(
			types.MRP("worker_id", i),
			types.MRP("status", "completed"),
			types.MRP("processed_items", (i+1)*10),
			types.MRP("duration_ms", 100),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			log.Error().Err(err).Int("worker_id", i).Msg("Failed to add result row")
			return err
		}

		log.Debug().Int("worker_id", i).Msg("Worker completed")
	}

	log.Info().Msg("Data processing completed successfully")
	return nil
}

func NewProcessDataCommand() (*ProcessDataCommand, error) {
	// Create logging layer with custom options
	loggingLayer, err := logging.NewLoggingLayerWithOptions(
		logging.WithDefaultLevel("info"),
		logging.WithDefaultFormat("text"),
		// logging.WithLogstash(), // Uncomment to include Logstash options
	)
	if err != nil {
		return nil, err
	}

	// Create glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"process-data",
		cmds.WithShort("Process data with configurable logging"),
		cmds.WithLong(`
Process data files with comprehensive logging support.
This command demonstrates how to use the custom logging layer.

Examples:
  process-data --input-file data.csv --workers 4
  process-data --input-file data.csv --log-level debug
  process-data --input-file data.csv --log-format json --log-file process.log
  process-data --input-file data.csv --verbose --with-caller
        `),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Input file to process"),
				parameters.WithRequired(true),
				parameters.WithShortFlag("i"),
			),
			parameters.NewParameterDefinition(
				"output-path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Output path for processed data"),
				parameters.WithDefault("output.processed"),
				parameters.WithShortFlag("o"),
			),
			parameters.NewParameterDefinition(
				"workers",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of worker processes"),
				parameters.WithDefault(2),
				parameters.WithShortFlag("w"),
			),
			parameters.NewParameterDefinition(
				"dry-run",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Perform a dry run without actual processing"),
				parameters.WithDefault(false),
			),
		),

		// Add both logging and glazed layers
		cmds.WithLayersList(loggingLayer, glazedLayer),
	)

	return &ProcessDataCommand{
		CommandDescription: cmdDesc,
	}, nil
}

var _ cmds.GlazeCommand = &ProcessDataCommand{}

// Second command to demonstrate layer reuse
type AnalyzeDataCommand struct {
	*cmds.CommandDescription
}

type AnalyzeDataSettings struct {
	DataFile   string `glazed.parameter:"data-file"`
	Algorithm  string `glazed.parameter:"algorithm"`
	Iterations int    `glazed.parameter:"iterations"`
}

func (c *AnalyzeDataCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging (same layer, reused!)
	if err := logging.InitializeLogging(parsedLayers); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	log.Info().Msg("Starting data analysis command")

	settings := &AnalyzeDataSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	log.Info().
		Str("data_file", settings.DataFile).
		Str("algorithm", settings.Algorithm).
		Int("iterations", settings.Iterations).
		Msg("Analysis configuration")

	// Simulate analysis
	for i := 0; i < settings.Iterations; i++ {
		log.Debug().Int("iteration", i+1).Msg("Running analysis iteration")

		// Simulate some analysis work
		time.Sleep(50 * time.Millisecond)

		row := types.NewRow(
			types.MRP("iteration", i+1),
			types.MRP("algorithm", settings.Algorithm),
			types.MRP("accuracy", 0.85+float64(i)*0.01),
			types.MRP("processing_time_ms", 50),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	log.Info().Msg("Analysis completed")
	return nil
}

func NewAnalyzeDataCommand() (*AnalyzeDataCommand, error) {
	// Reuse the same logging layer - this is the power of layers!
	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"analyze-data",
		cmds.WithShort("Analyze data with configurable logging"),
		cmds.WithLong("Analyze data files using various algorithms with the same logging configuration."),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"data-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Data file to analyze"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"algorithm",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("linear", "logistic", "random-forest", "neural-net"),
				parameters.WithDefault("linear"),
				parameters.WithHelp("Analysis algorithm to use"),
			),
			parameters.NewParameterDefinition(
				"iterations",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(3),
				parameters.WithHelp("Number of analysis iterations"),
			),
		),

		cmds.WithLayersList(loggingLayer, glazedLayer),
	)

	return &AnalyzeDataCommand{
		CommandDescription: cmdDesc,
	}, nil
}

var _ cmds.GlazeCommand = &AnalyzeDataCommand{}

func main() {
	rootCmd := &cobra.Command{
		Use:   "data-processor",
		Short: "Data processing application with custom logging layer",
		Long:  "Demonstrates how to create and reuse custom parameter layers in Glazed",
	}

	// Create and register process command
	processCmd, err := NewProcessDataCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating process command: %v\n", err)
		os.Exit(1)
	}

	cobraProcessCmd, err := cli.BuildCobraCommandFromGlazeCommand(processCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building process command: %v\n", err)
		os.Exit(1)
	}

	// Create and register analyze command
	analyzeCmd, err := NewAnalyzeDataCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating analyze command: %v\n", err)
		os.Exit(1)
	}

	cobraAnalyzeCmd, err := cli.BuildCobraCommandFromGlazeCommand(analyzeCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building analyze command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraProcessCmd, cobraAnalyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
