package cli

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/spf13/cobra"
)

// CobraParser takes a CommandDescription, and hooks it up to a cobra command.
// It can then be used to parse the cobra flags and arguments back into a
// set of ParsedLayer and a map[string]interface{} for the lose stuff.
//
// That command however doesn't have a Run* method, which is left to the caller to implement.
//
// This returns a CobraParser that can be used to parse the registered layers
// from the description.
//
// NOTE(manuel, 2023-09-18) Now that I've removed the parserFunc, this feels a bit unnecessary too
// Or it could be something that is actually an interface on top of Command, like a CobraCommand.
type CobraParser struct {
	Cmd             *cobra.Command
	description     *cmds.CommandDescription
	middlewaresFunc CobraMiddlewaresFunc
}

type CobraParserOption func(*CobraParser) error

func WithCobraMiddlewaresFunc(middlewaresFunc CobraMiddlewaresFunc) CobraParserOption {
	return func(c *CobraParser) error {
		c.middlewaresFunc = middlewaresFunc
		return nil
	}
}

func WithAddGlazedCommandLayer() CobraParserOption {
	return func(c *CobraParser) error {
		glazedCommandLayer, err := NewGlazedCommandLayer()
		if err != nil {
			return err
		}
		c.description.Layers.Set(glazedCommandLayer.GetSlug(), glazedCommandLayer)
		return nil
	}
}

func NewCobraParserFromCommandDescription(
	description *cmds.CommandDescription,
	options ...CobraParserOption,
) (*CobraParser, error) {
	cmd := &cobra.Command{
		Use:   description.Name,
		Short: description.Short,
		Long:  description.Long,
	}

	ret := &CobraParser{
		Cmd:             cmd,
		description:     description,
		middlewaresFunc: CobraCommandDefaultMiddlewares,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	err := description.Layers.ForEachE(func(_ string, layer layers.ParameterLayer) error {
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
		return nil, err
	}

	return ret, nil
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

	err = cmd_middlewares.ExecuteMiddlewares(c.description.Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, err
	}

	return parsedLayers, nil
}
