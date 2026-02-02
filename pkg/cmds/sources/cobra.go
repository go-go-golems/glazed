package sources

import (
	"sync"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// FromCobra creates a middleware that parses parameter values from a Cobra command.
// This middleware is typically used as the highest priority in the middleware chain for CLI applications.
//
// It iterates through each layer, and if the layer implements the CobraSection interface,
// it parses the layer's parameters from the Cobra command.
//
// Usage:
//
//	middleware := middlewares.FromCobra(cmd, fields.WithSource("flags"))
func FromCobra(cmd *cobra.Command, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			err = layers_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithMetadata(map[string]interface{}{
						"layer":          l.GetName(),
						"layer_slug":     l.GetSlug(),
						"layer_prefix":   l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)

				if cobraLayer, ok := l.(schema.CobraSection); ok {
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

// FromArgs creates a middleware that parses positional arguments for the default layer.
// This middleware is typically used in conjunction with FromCobra for CLI applications
// that accept positional arguments.
//
// Usage:
//
//	middleware := middlewares.FromArgs(args, fields.WithSource("args"))
func FromArgs(args []string, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if defaultLayer, ok := layers_.Get(schema.DefaultSlug); ok {
				pds := defaultLayer.GetDefinitions()
				ps_, err := pds.GatherArguments(args, false, false, append(options, fields.WithSource("arguments"))...)
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

// ConfigFilesResolver is a callback used by Cobra-specific middleware to resolve the list
// of config files to load in low -> high precedence order.
type ConfigFilesResolver func(parsedCommandLayers *values.Values, cmd *cobra.Command, args []string) ([]string, error)

// LoadParametersFromResolvedFilesForCobra loads parameters from a resolver-provided list of files
// (low -> high precedence). Each file is tracked as a separate parse step with metadata.
func LoadParametersFromResolvedFilesForCobra(
	cmd *cobra.Command,
	args []string,
	resolver ConfigFilesResolver,
	options ...fields.ParseOption,
) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			if err := next(layers_, parsedLayers); err != nil {
				return err
			}
			files, err := resolver(parsedLayers, cmd, args)
			if err != nil {
				return err
			}
			// Apply as a single multi-file step using helper
			// Wrap ParseOptions into ConfigFileOptions
			configOpts := []ConfigFileOption{}
			if len(options) > 0 {
				configOpts = append(configOpts, WithParseOptions(options...))
			}
			return FromFiles(files, configOpts...)(func(_ *schema.Schema, _ *values.Values) error { return nil })(layers_, parsedLayers)
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
//	middleware := middlewares.GatherFlagsFromViper(fields.WithSource("viper"))
//
// Deprecated: Use FromFiles and FromEnv instead.
func GatherFlagsFromViper(options ...fields.ParseOption) Middleware {
	warnGatherViperOnce.Do(func() {
		log.Warn().Msg("middlewares.GatherFlagsFromViper is deprecated; use FromFiles + FromEnv")
	})
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {

			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithSource("viper"),
					fields.WithMetadata(map[string]interface{}{
						"layer":          l.GetName(),
						"layer_slug":     l.GetSlug(),
						"layer_prefix":   l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetDefinitions()
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
//	    fields.WithSource("viper"),
//	)
//
// Deprecated: Use FromFiles and FromEnv instead.
func GatherSpecificFlagsFromViper(flags []string, options ...fields.ParseOption) Middleware {
	warnGatherViperOnce.Do(func() {
		log.Warn().Msg("middlewares.GatherSpecificFlagsFromViper is deprecated; use FromFiles + FromEnv")
	})
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}
			err = layers_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithSource("viper"),
					fields.WithMetadata(map[string]interface{}{
						"layer": l.GetName(),
					}),
				}, options...)

				parsedLayer := parsedLayers.GetOrCreate(l)
				parameterDefinitions := l.GetDefinitions()
				prefix := l.GetPrefix()

				// Filter the parameter definitions based on the specified flags
				filteredDefinitions := fields.NewDefinitions()
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

var warnGatherViperOnce sync.Once
