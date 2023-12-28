package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

func ParseFromCobraCommand(cmd *cobra.Command, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				options_ := append([]parameters.ParseStepOption{
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer": l.GetName(),
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)

				if cobraLayer, ok := l.(layers.CobraParameterLayer); ok {
					cobraLayer, err := cobraLayer.ParseLayerFromCobraCommand(cmd, options_...)
					if err != nil {
						return err
					}

					parsedLayer.Parameters.Merge(cobraLayer.Parameters)
				}

				return nil
			})
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func GatherArguments(args []string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if defaultLayer, ok := layers_.Get(layers.DefaultSlug); ok {
				pds := defaultLayer.GetParameterDefinitions()
				ps_, err := pds.GatherArguments(args, false, false, options...)
				if err != nil {
					return err
				}

				parsedLayer := parsedLayers.GetOrCreate(defaultLayer)
				parsedLayer.Parameters.Merge(ps_)
			}

			return nil
		}
	}
}

func GatherFlagsFromViper(options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				options_ := append([]parameters.ParseStepOption{
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer": l.GetName(),
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetParameterDefinitions()
				prefix := l.GetPrefix()

				ps, err := parameterDefinitions.GatherFlagsFromViper(false, prefix, options_...)
				if err != nil {
					return err
				}

				parsedLayer.Parameters.Merge(ps)

				return nil
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}
