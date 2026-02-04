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
// the glazed processing section.
//
// If you are more serious about using glazed, consider using the `cmds.GlazeCommand` and `fields.Definition`
// abstraction to define your CLI applications, which allows you to use sections and other nice features
// of the glazed ecosystem.
//
// If so, use SetupTableProcessor instead, and create a proper glazed.GlazeCommand for your command.
func CreateGlazedProcessorFromCobra(cmd *cobra.Command) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	gpl, err := settings.NewGlazedSection()
	if err != nil {
		return nil, nil, err
	}

	schema_ := schema.NewSchema(schema.WithSections(gpl))
	parser, err := NewCobraParserFromSections(schema_, &CobraParserConfig{
		MiddlewaresFunc: CobraCommandDefaultMiddlewares,
	})
	if err != nil {
		return nil, nil, err
	}
	parsedValues, err := parser.Parse(cmd, nil)
	if err != nil {
		return nil, nil, err
	}

	parsedSectionValues, ok := parsedValues.Get(settings.GlazedSlug)
	if !ok {
		return nil, nil, errors.Errorf("section %s not found", settings.GlazedSlug)
	}

	gp, err := settings.SetupTableProcessor(parsedSectionValues)
	cobra.CheckErr(err)

	of, err := settings.SetupProcessorOutput(gp, parsedSectionValues, os.Stdout)
	cobra.CheckErr(err)

	return gp, of, nil
}

// AddGlazedProcessorFlagsToCobraCommand is a helper for cobra centric apps that quickly want to add
// the glazed processing section to their CLI flags.
func AddGlazedProcessorFlagsToCobraCommand(cmd *cobra.Command, options ...settings.GlazeSectionOption) error {
	gpl, err := settings.NewGlazedSection(options...)
	if err != nil {
		return err
	}

	return gpl.AddSectionToCobraCommand(cmd)
}

func printParsedFields(parsedValues *values.Values) {
	sectionsMap := map[string]map[string]interface{}{}
	parsedValues.ForEach(func(sectionName string, sectionValues *values.SectionValues) {
		fieldValues := map[string]interface{}{}
		sectionValues.Fields.ForEach(func(name string, fieldValue *fields.FieldValue) {
			fieldMap := map[string]interface{}{
				"value": fieldValue.Value,
			}
			logs := make([]map[string]interface{}, 0, len(fieldValue.Log))
			for _, l := range fieldValue.Log {
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
				fieldMap["log"] = logs
			}
			fieldValues[name] = fieldMap
		})
		sectionsMap[sectionName] = fieldValues
	})

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	err := encoder.Encode(sectionsMap)
	cobra.CheckErr(err)
}
