package cli

import (
	"os"

	fields "github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CreateGlazedProcessorFromCobra is a helper for cobra centric apps that quickly want to add
// the glazed processing layer.
//
// If you are more serious about using glazed, consider using the `cmds.GlazeCommand` and `fields.Definition`
// abstraction to define your CLI applications, which allows you to use Layers and other nice features
// of the glazed ecosystem.
//
// If so, use SetupTableProcessor instead, and create a proper glazed.GlazeCommand for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	gpl, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	layers_ := schema.NewSchema(schema.WithSections(gpl))
	parser, err := NewCobraParserFromLayers(layers_, &CobraParserConfig{
		MiddlewaresFunc: CobraCommandDefaultMiddlewares,
	})
	if err != nil {
		return nil, nil, err
	}
	parsedLayers, err := parser.Parse(cmd, nil)
	if err != nil {
		return nil, nil, err
	}

	parsedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
	if !ok {
		return nil, nil, errors.Errorf("layer %s not found", settings.GlazedSlug)
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

	return gpl.AddSectionToCobraCommand(cmd)
}

func printParsedParameters(parsedLayers *values.Values) {
	layersMap := map[string]map[string]interface{}{}
	parsedLayers.ForEach(func(layerName string, layer *values.SectionValues) {
		params := map[string]interface{}{}
		layer.Parameters.ForEach(func(name string, parameter *fields.ParsedParameter) {
			paramMap := map[string]interface{}{
				"value": parameter.Value,
			}
			logs := make([]map[string]interface{}, 0, len(parameter.Log))
			for _, l := range parameter.Log {
				logEntry := map[string]interface{}{
					"source": l.Source,
					"value":  l.Value,
				}
				if len(l.Metadata) > 0 {
					logEntry["metadata"] = l.Metadata
				}
				logs = append(logs, logEntry)
			}
			if len(logs) > 0 {
				paramMap["log"] = logs
			}
			params[name] = paramMap
		})
		layersMap[layerName] = params
	})

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	err := encoder.Encode(layersMap)
	cobra.CheckErr(err)
}
