package cli

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// CobraMiddlewaresFunc is a function that returns a list of middlewares for a cobra command.
// It can be used to overload the default middlewares for cobra commands.
// It is mostly used to add a "load from json" layer set in the GlazedCommandSettings.
type CobraMiddlewaresFunc func(commandSettings *GlazedCommandSettings, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error)

// CobraCommandDefaultMiddlewares is the default implementation for creating
// the middlewares used in a Cobra command. It handles parsing parameters
// from Cobra flags, command line arguments, environment variables, and
// default values. The middlewares gather all these parameters into a
// ParsedParameters object.
//
// If the commandSettings specify parameters to be loaded from a file, this gets added as a middleware.
func CobraCommandDefaultMiddlewares(commandSettings *GlazedCommandSettings, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error) {
	middlewares_ := []cmd_middlewares.Middleware{
		cmd_middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		cmd_middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	if commandSettings.LoadParametersFromFile != "" {
		middlewares_ = append(middlewares_,
			cmd_middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	middlewares_ = append(middlewares_,
		cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewares_, nil
}

// CobraParser takes a CommandDescription, and hooks it up to a cobra command.
// It can then be used to parse the cobra flags and arguments back into a
// set of ParsedLayer and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run* method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered Layers
// from the description.
type CobraParser struct {
	Layers *layers.ParameterLayers
	// middlewaresFunc is called after the command has been executed, once the
	// GlazedCommandSettings struct has been filled. At this point, cobra has done the parsing
	// of CLI flags and arguments, but these haven't yet been parsed into ParsedLayers
	// by the glazed framework.
	//
	// This hooks allows the implementor to specify additional ways of loading parameters
	// (for example, sqleton loads the dbt and sql connection parameters from env and viper as well).
	middlewaresFunc CobraMiddlewaresFunc
}

type CobraParserOption func(*CobraParser) error

func WithCobraMiddlewaresFunc(middlewaresFunc CobraMiddlewaresFunc) CobraParserOption {
	return func(c *CobraParser) error {
		c.middlewaresFunc = middlewaresFunc
		return nil
	}
}

func NewCobraCommandFromCommandDescription(
	description *cmds.CommandDescription,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}
	return cmd
}

// NewCobraParserFromLayers creates a new CobraParser instance from a
// CommandDescription, initializes the underlying cobra.Command, and adds all the
// parameters specified in the Layers CommandDescription to the cobra command.
func NewCobraParserFromLayers(
	layers *layers.ParameterLayers,
	options ...CobraParserOption,
) (*CobraParser, error) {
	ret := &CobraParser{
		Layers:          layers,
		middlewaresFunc: CobraCommandDefaultMiddlewares,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	// NOTE(manuel, 2023-12-30) I actually think we always want to have the glazed-command layer
	glazedCommandLayer, err := NewGlazedCommandLayer()
	if err != nil {
		return nil, err
	}
	ret.Layers.Set(glazedCommandLayer.GetSlug(), glazedCommandLayer)

	return ret, nil
}

func (c *CobraParser) AddToCobraCommand(cmd *cobra.Command) error {

	// NOTE(manuel, 2024-01-03) Maybe add some middleware functionality to whitelist/blacklist the Layers/parameters that get added to the CLI
	// If we want to remove some parameters from the CLI args (for example some output settings or so)
	err := c.Layers.ForEachE(func(_ string, layer layers.ParameterLayer) error {
		// check that layer is a CobraParameterLayer
		// if not, return an error
		cobraLayer, ok := layer.(layers.CobraParameterLayer)
		if !ok {
			return fmt.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		err := cobraLayer.AddLayerToCobraCommand(cmd)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *CobraParser) Parse(
	cmd *cobra.Command,
	args []string,
) (*layers.ParsedLayers, error) {
	parsedLayers := layers.NewParsedLayers()

	glazedCommandLayer, err := NewGlazedCommandLayer()
	if err != nil {
		return nil, err
	}

	pds := glazedCommandLayer.GetParameterDefinitions()
	parsedParameters, err := pds.GatherFlagsFromCobraCommand(cmd, true, true, glazedCommandLayer.GetPrefix())
	if err != nil {
		return nil, err
	}

	commandSettings := &GlazedCommandSettings{}
	err = parsedParameters.InitializeStruct(commandSettings)
	if err != nil {
		return nil, err
	}

	middlewares_, err := c.middlewaresFunc(commandSettings, cmd, args)
	if err != nil {
		return nil, err
	}

	err = cmd_middlewares.ExecuteMiddlewares(c.Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}

func ParseLayersFromCobraCommand(cmd *cobra.Command, layers_ *layers.ParameterLayers) (
	*layers.ParsedLayers,
	error,
) {
	middlewares := []cmd_middlewares.Middleware{
		cmd_middlewares.ParseFromCobraCommand(cmd, parameters.WithParseStepSource("cobra")),
		cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	}
	parsedLayers := layers.NewParsedLayers()
	err := cmd_middlewares.ExecuteMiddlewares(layers_, parsedLayers, middlewares...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}
