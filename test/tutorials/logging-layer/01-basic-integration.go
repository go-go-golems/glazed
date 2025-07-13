package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/rs/zerolog/log"
)

type MyCommand struct {
	*cmds.CommandDescription
}

func NewMyCommand() (*MyCommand, error) {
	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create logging layer: %w", err)
	}

	cmdDesc := cmds.NewCommandDescription(
		"my-command",
		cmds.WithShort("Command with logging support"),
		cmds.WithLayersList(loggingLayer),
	)

	return &MyCommand{CommandDescription: cmdDesc}, nil
}

func (c *MyCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Extract logging settings from parsed layers
	var settings logging.LoggingSettings
	err := parsedLayers.InitializeStruct(logging.LoggingLayerSlug, &settings)
	if err != nil {
		return fmt.Errorf("failed to get logging settings: %w", err)
	}

	// Setup logging
	if err := logging.InitLoggerFromSettings(&settings); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	log.Info().Msg("Processing started")
	log.Debug().Str("command", "my-command").Msg("Debug information")
	log.Info().Msg("Processing completed")
	return nil
}

func main() {
	cmd, err := NewMyCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create command")
	}

	// Create test parsed layers
	parsedLayers := layers.NewParsedLayers()
	loggingLayer, _ := logging.NewLoggingLayer()

	// Add logging layer with default settings
	parsedLayer, err := layers.NewParsedLayer(loggingLayer)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create parsed layer")
	}
	parsedLayers.Set(logging.LoggingLayerSlug, parsedLayer)

	// Run command
	err = cmd.RunIntoGlazeProcessor(context.Background(), parsedLayers, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Command failed")
	}
}
