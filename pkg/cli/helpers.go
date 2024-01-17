package cli

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
	"os"
)

// CreateGlazedProcessorFromCobra is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
//
// If you are more serious about using glazed, consider using the `cmds.GlazeCommand` and `parameters.ParameterDefinition`
// abstraction to define your CLI applications, which allows you to use Layers and other nice features
// of the glazed ecosystem.
//
// If so, use SetupTableProcessor instead, and create a proper glazed.GlazeCommand for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	gpl, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	layers_ := layers.NewParameterLayers(layers.WithLayers(gpl))
	parser, err := NewCobraParserFromLayers(layers_,
		WithCobraMiddlewaresFunc(CobraCommandDefaultMiddlewares))
	if err != nil {
		return nil, nil, err
	}
	parsedLayers, err := parser.Parse(cmd, nil)
	if err != nil {
		return nil, nil, err
	}

	parsedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
	if !ok {
		return nil, nil, fmt.Errorf("layer %s not found", settings.GlazedSlug)
	}

	gp, err := settings.SetupTableProcessor(parsedLayer)
	cobra.CheckErr(err)

	of, err := settings.SetupProcessorOutput(gp, parsedLayer, os.Stdout)
	cobra.CheckErr(err)

	return gp, of, nil
}

// AddGlazedProcessorFlagsToCobraCommand is a helper for cobra centric apps that quickly want to add
// the glazed processing layer to their CLI flags.
func AddGlazedProcessorFlagsToCobraCommand(cmd *cobra.Command, options ...settings.GlazeParameterLayerOption) error {
	gpl, err := settings.NewGlazedParameterLayers(options...)
	if err != nil {
		return err
	}

	return gpl.AddLayerToCobraCommand(cmd)
}

func printParsedParameters(parsedLayers *layers.ParsedLayers) {
	parsedLayers.ForEach(func(layerName string, layer *layers.ParsedLayer) {
		fmt.Printf("# %s:\n", layerName)
		layer.Parameters.ForEach(func(name string, parameter *parameters.ParsedParameter) {
			fmt.Printf("%s: (%s)\n  value: '%v'\n", name, parameter.ParameterDefinition.Type, parameter.Value)
			for _, l := range parameter.Log {
				fmt.Printf("\tsource: %s: %v\n", l.Source, l.Value)
				for k, v := range l.Metadata {
					fmt.Printf("\t\t%s: %v\n", k, v)
				}
			}
		})
	})
}
