package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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
