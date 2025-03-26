package cli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// CobraMiddlewaresFunc is a function that returns a list of middlewares for a cobra command.
// It can be used to overload the default middlewares for cobra commands.
// It is mostly used to add a "load from json" layer set in the GlazedCommandSettings.
type CobraMiddlewaresFunc func(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]cmd_middlewares.Middleware, error)

// CobraCommandDefaultMiddlewares is the default implementation for creating
// the middlewares used in a Cobra command. It handles parsing parameters
// from Cobra flags, command line arguments, environment variables, and
// default values. The middlewares gather all these parameters into a
// ParsedParameters object.
//
// If the commandSettings specify parameters to be loaded from a file, this gets added as a middleware.
func CobraCommandDefaultMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]cmd_middlewares.Middleware, error) {
	commandSettings := &CommandSettings{}
	err := parsedCommandLayers.InitializeStruct(CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

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
		cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
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
	// List of layers to be shown in short help, empty: always show all
	shortHelpLayers []string
	// skipCommandSettingsLayer controls whether the CommandSettingsLayer should be automatically added
	skipCommandSettingsLayer bool
	// enableProfileSettingsLayer controls whether the ProfileSettingsLayer should be added
	enableProfileSettingsLayer bool
	// enableCreateCommandSettingsLayer controls whether the CreateCommandSettingsLayer should be added
	enableCreateCommandSettingsLayer bool
}

type CobraParserOption func(*CobraParser) error

func WithCobraMiddlewaresFunc(middlewaresFunc CobraMiddlewaresFunc) CobraParserOption {
	return func(c *CobraParser) error {
		c.middlewaresFunc = middlewaresFunc
		return nil
	}
}

func WithCobraShortHelpLayers(layers ...string) CobraParserOption {
	return func(c *CobraParser) error {
		c.shortHelpLayers = append(c.shortHelpLayers, layers...)
		return nil
	}
}

func WithSkipCommandSettingsLayer() CobraParserOption {
	return func(c *CobraParser) error {
		c.skipCommandSettingsLayer = true
		return nil
	}
}

func WithProfileSettingsLayer() CobraParserOption {
	return func(c *CobraParser) error {
		c.enableProfileSettingsLayer = true
		return nil
	}
}

func WithCreateCommandSettingsLayer() CobraParserOption {
	return func(c *CobraParser) error {
		c.enableCreateCommandSettingsLayer = true
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

	// Only add the glazed command layer if not explicitly skipped
	if !ret.skipCommandSettingsLayer {
		commandSettingsLayer, err := NewCommandSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(commandSettingsLayer.GetSlug(), commandSettingsLayer)
	}

	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		return nil, err
	}
	ret.Layers.Set(logging.LoggingLayerSlug, loggingLayer)

	// Only add the profile settings layer if explicitly enabled
	if ret.enableProfileSettingsLayer {
		profileSettingsLayer, err := NewProfileSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(profileSettingsLayer.GetSlug(), profileSettingsLayer)
	}

	// Only add the create command settings layer if explicitly enabled
	if ret.enableCreateCommandSettingsLayer {
		createCommandSettingsLayer, err := NewCreateCommandSettingsLayer()
		if err != nil {
			return nil, err
		}
		ret.Layers.Set(createCommandSettingsLayer.GetSlug(), createCommandSettingsLayer)
	}

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
			log.Error().Str("layer", layer.GetName()).Msg("Layer is not a CobraParameterLayer")
			return errors.Errorf("layer %s is not a CobraParameterLayer", layer.GetName())
		}

		err := cobraLayer.AddLayerToCobraCommand(cmd)
		if err != nil {
			log.Error().Err(err).Str("layer", layer.GetName()).Msg("Could not add layer to cobra command")
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(c.shortHelpLayers) > 0 {
		shortHelperLayer := strings.Join(c.shortHelpLayers, ",")
		cmd.Annotations["shortHelpLayers"] = shortHelperLayer
	}

	return nil
}

func (c *CobraParser) Parse(
	cmd *cobra.Command,
	args []string,
) (*layers.ParsedLayers, error) {
	// We parse the glazed command settings first, since they will influence the following parsing
	// steps.
	parsedCommandLayers, err := ParseCommandSettingsLayer(cmd)
	if err != nil {
		return nil, err
	}

	// Create the middlewares by invoking the passed in middlewares constructor.
	// This is where applications can specify their own middlewares.
	middlewares_, err := c.middlewaresFunc(parsedCommandLayers, cmd, args)
	if err != nil {
		return nil, err
	}

	parsedLayers := layers.NewParsedLayers()
	err = cmd_middlewares.ExecuteMiddlewares(c.Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}

// ParseGlazedCommandLayer parses the global glazed settings from the given cobra.Command, if not nil,
// and from the configured viper config file.
func ParseCommandSettingsLayer(cmd *cobra.Command) (*layers.ParsedLayers, error) {
	parsedLayers := layers.NewParsedLayers()
	commandSettingsLayer, err := NewCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	profileSettingsLayer, err := NewProfileSettingsLayer()
	if err != nil {
		return nil, err
	}

	createCommandSettingsLayer, err := NewCreateCommandSettingsLayer()
	if err != nil {
		return nil, err
	}

	commandSettingsLayers := layers.NewParameterLayers(
		layers.WithLayers(
			commandSettingsLayer,
			profileSettingsLayer,
			createCommandSettingsLayer,
		),
	)

	// Parse the glazed command settings from the cobra command and config file
	middlewares_ := []cmd_middlewares.Middleware{}

	if cmd != nil {
		middlewares_ = append(middlewares_, cmd_middlewares.ParseFromCobraCommand(cmd, parameters.WithParseStepSource("cobra")))
	}

	middlewares_ = append(middlewares_, cmd_middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")))

	err = cmd_middlewares.ExecuteMiddlewares(commandSettingsLayers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil

}
