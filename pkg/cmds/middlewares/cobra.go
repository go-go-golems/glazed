package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ParseFromCobraCommand creates a middleware that parses parameter values from a Cobra command.
// This middleware is typically used as the highest priority in the middleware chain for CLI applications.
//
// It iterates through each layer, and if the layer implements the CobraParameterLayer interface,
// it parses the layer's parameters from the Cobra command.
//
// Usage:
//
//	middleware := middlewares.ParseFromCobraCommand(cmd, parameters.WithParseStepSource("flags"))
func ParseFromCobraCommand(cmd *cobra.Command, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				log.Debug().Str("registeredKey", key).Str("layerSlug", l.GetSlug()).Str("layerName", l.GetName()).Msg("ParseFromCobraCommand: iterating layer")
				options_ := append([]parameters.ParseStepOption{
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer":          l.GetName(),
						"layer_slug":     l.GetSlug(),
						"layer_prefix":   l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)

				if cobraLayer, ok := l.(layers.CobraParameterLayer); ok {
					log.Debug().Str("layerSlug", l.GetSlug()).Msg("ParseFromCobraCommand: parsing layer from cobra")
					cobraLayer, err := cobraLayer.ParseLayerFromCobraCommand(cmd, options_...)
					if err != nil {
						return err
					}

					_, err = parsedLayer.Parameters.Merge(cobraLayer.Parameters)
					if err != nil {
						return err
					}
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

// GatherArguments creates a middleware that parses positional arguments for the default layer.
// This middleware is typically used in conjunction with ParseFromCobraCommand for CLI applications
// that accept positional arguments.
//
// Usage:
//
//	middleware := middlewares.GatherArguments(args, parameters.WithParseStepSource("args"))
func GatherArguments(args []string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if defaultLayer, ok := layers_.Get(layers.DefaultSlug); ok {
				pds := defaultLayer.GetParameterDefinitions()
				ps_, err := pds.GatherArguments(args, false, false, append(options, parameters.WithParseStepSource("arguments"))...)
				if err != nil {
					return err
				}

				parsedLayer := parsedLayers.GetOrCreate(defaultLayer)
				_, err = parsedLayer.Parameters.Merge(ps_)
				if err != nil {
					return err
				}
			}

			return nil
		}
	}
}

// GatherFlagsFromViper creates a middleware that loads parameter values from Viper configuration.
// This middleware is useful for integrating Viper-based configuration management with Glazed commands.
//
// It iterates through each layer, gathering flags from Viper for all parameters in that layer.
//
// Usage:
//
//	middleware := middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper"))
func GatherFlagsFromViper(options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {

			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				options_ := append([]parameters.ParseStepOption{
					parameters.WithParseStepSource("viper"),
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer":          l.GetName(),
						"layer_slug":     l.GetSlug(),
						"layer_prefix":   l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetParameterDefinitions()
				prefix := l.GetPrefix()

				ps, err := parameterDefinitions.GatherFlagsFromViper(true, prefix, options_...)
				if err != nil {
					return err
				}

				_, err = parsedLayer.Parameters.Merge(ps)
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
	}
}

// GatherSpecificFlagsFromViper creates a middleware that loads specific parameter values from Viper configuration.
// This middleware is similar to GatherFlagsFromViper, but it only loads values for the specified flags.
//
// It's useful when you want to selectively load certain parameters from Viper while leaving others untouched.
//
// Usage:
//
//	middleware := middlewares.GatherSpecificFlagsFromViper(
//	    []string{"flag1", "flag2"},
//	    parameters.WithParseStepSource("viper"),
//	)
func GatherSpecificFlagsFromViper(flags []string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l layers.ParameterLayer) error {
				options_ := append([]parameters.ParseStepOption{
					parameters.WithParseStepSource("viper"),
					parameters.WithParseStepMetadata(map[string]interface{}{
						"layer": l.GetName(),
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetParameterDefinitions()
				prefix := l.GetPrefix()

				// Filter the parameter definitions based on the specified flags
				filteredDefinitions := parameters.NewParameterDefinitions()
				for _, flag := range flags {
					if pd, ok := parameterDefinitions.Get(flag); ok {
						filteredDefinitions.Set(pd.Name, pd)
					}
				}

				ps, err := filteredDefinitions.GatherFlagsFromViper(true, prefix, options_...)
				if err != nil {
					return err
				}

				_, err = parsedLayer.Parameters.Merge(ps)
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
	}
}
