package sources

import (
	"sync"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// FromCobra creates a middleware that parses field values from a Cobra command.
// This middleware is typically used as the highest priority in the middleware chain for CLI applications.
//
// It iterates through each section, and if the section implements the CobraSection interface,
// it parses the section's fields from the Cobra command.
//
// Usage:
//
//	middleware := middlewares.FromCobra(cmd, fields.WithSource("flags"))
func FromCobra(cmd *cobra.Command, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			err = schema_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithMetadata(map[string]interface{}{
						"section":        l.GetName(),
						"section_slug":   l.GetSlug(),
						"section_prefix": l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				sectionValues := parsedValues.GetOrCreate(l)

				if cobraSection, ok := l.(schema.CobraSection); ok {
					sectionValuesFromCobra, err := cobraSection.ParseSectionFromCobraCommand(cmd, options_...)
					if err != nil {
						return err
					}

					_, err = sectionValues.Fields.Merge(sectionValuesFromCobra.Fields)
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

// FromArgs creates a middleware that parses positional arguments for the default section.
// This middleware is typically used in conjunction with FromCobra for CLI applications
// that accept positional arguments.
//
// Usage:
//
//	middleware := middlewares.FromArgs(args, fields.WithSource("args"))
func FromArgs(args []string, options ...fields.ParseOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if defaultSection, ok := schema_.Get(schema.DefaultSlug); ok {
				pds := defaultSection.GetDefinitions()
				ps_, err := pds.GatherArguments(args, false, false, append(options, fields.WithSource("arguments"))...)
				if err != nil {
					return err
				}

				sectionValues := parsedValues.GetOrCreate(defaultSection)
				_, err = sectionValues.Fields.Merge(ps_)
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
type ConfigFilesResolver func(parsedValues *values.Values, cmd *cobra.Command, args []string) ([]string, error)

// LoadFieldsFromResolvedFilesForCobra loads fields from a resolver-provided list of files
// (low -> high precedence). Each file is tracked as a separate parse step with metadata.
func LoadFieldsFromResolvedFilesForCobra(
	cmd *cobra.Command,
	args []string,
	resolver ConfigFilesResolver,
	options ...fields.ParseOption,
) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			if err := next(schema_, parsedValues); err != nil {
				return err
			}
			files, err := resolver(parsedValues, cmd, args)
			if err != nil {
				return err
			}
			// Apply as a single multi-file step using helper
			// Wrap ParseOptions into ConfigFileOptions
			configOpts := []ConfigFileOption{}
			if len(options) > 0 {
				configOpts = append(configOpts, WithParseOptions(options...))
			}
			return FromFiles(files, configOpts...)(func(_ *schema.Schema, _ *values.Values) error { return nil })(schema_, parsedValues)
		}
	}
}

// GatherFlagsFromViper creates a middleware that loads field values from Viper configuration.
// This middleware is useful for integrating Viper-based configuration management with Glazed commands.
//
// It iterates through each section, gathering flags from Viper for all fields in that section.
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
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {

			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}
			err = schema_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithSource("viper"),
					fields.WithMetadata(map[string]interface{}{
						"section":        l.GetName(),
						"section_slug":   l.GetSlug(),
						"section_prefix": l.GetPrefix(),
						"registered_key": key,
					}),
				}, options...)

				sectionValues := parsedValues.GetOrCreate(l)
				fieldDefinitions := l.GetDefinitions()
				prefix := l.GetPrefix()

				ps, err := fieldDefinitions.GatherFlagsFromViper(true, prefix, options_...)
				if err != nil {
					return err
				}

				_, err = sectionValues.Fields.Merge(ps)
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

// GatherSpecificFlagsFromViper creates a middleware that loads specific field values from Viper configuration.
// This middleware is similar to GatherFlagsFromViper, but it only loads values for the specified flags.
//
// It's useful when you want to selectively load certain fields from Viper while leaving others untouched.
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
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}
			err = schema_.ForEachE(func(key string, l schema.Section) error {
				options_ := append([]fields.ParseOption{
					fields.WithSource("viper"),
					fields.WithMetadata(map[string]interface{}{
						"section": l.GetName(),
					}),
				}, options...)

				sectionValues := parsedValues.GetOrCreate(l)
				fieldDefinitions := l.GetDefinitions()
				prefix := l.GetPrefix()

				// Filter the field definitions based on the specified flags
				filteredDefinitions := fields.NewDefinitions()
				for _, flag := range flags {
					if pd, ok := fieldDefinitions.Get(flag); ok {
						filteredDefinitions.Set(pd.Name, pd)
					}
				}

				ps, err := filteredDefinitions.GatherFlagsFromViper(true, prefix, options_...)
				if err != nil {
					return err
				}

				_, err = sectionValues.Fields.Merge(ps)
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
