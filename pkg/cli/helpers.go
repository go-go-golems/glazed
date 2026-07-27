package cli

import (
	"encoding/json"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
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

// CreateStructuredOutputProcessorFromCobra creates a processor for Cobra-centric
// applications that mounted NewStructuredOutputSection on the command.
func CreateStructuredOutputProcessorFromCobra(cmd *cobra.Command) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	outputSection, err := settings.NewStructuredOutputSection()
	if err != nil {
		return nil, nil, err
	}

	schema_ := schema.NewSchema(schema.WithSections(outputSection))
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

	parsedSectionValues, ok := parsedValues.Get(settings.StructuredOutputSlug)
	if !ok {
		return nil, nil, errors.Errorf("section %s not found", settings.StructuredOutputSlug)
	}

	gp, outputFormatter, err := settings.SetupStructuredOutput(parsedSectionValues, os.Stdout)
	cobra.CheckErr(err)
	return gp, outputFormatter, nil
}

// AddStructuredOutputFlagsToCobraCommand adds --format, --output-fields, and
// --max-output-rows to a Cobra command.
func AddStructuredOutputFlagsToCobraCommand(cmd *cobra.Command, options ...schema.SectionOption) error {
	outputSection, err := settings.NewStructuredOutputSection(options...)
	if err != nil {
		return err
	}

	return outputSection.AddSectionToCobraCommand(cmd)
}

// HandleCommandSettings handles the framework-level --print-* command settings
// after parsing and before command execution. It returns handled=true when the
// caller should stop without running the command implementation.
func HandleCommandSettings(command cmds.Command, parsedValues *values.Values, w io.Writer) (bool, error) {
	commandSettingsValues, ok := parsedValues.Get(CommandSettingsSlug)
	if !ok {
		return false, nil
	}
	commandSettings := &CommandSettings{}
	if err := commandSettingsValues.DecodeInto(commandSettings); err != nil {
		return true, err
	}
	switch {
	case commandSettings.PrintParsedFields:
		return true, PrintParsedFields(w, parsedValues)
	case commandSettings.PrintYAML:
		return true, command.ToYAML(w)
	case commandSettings.PrintSchema:
		schema, err := command.Description().ToJsonSchema()
		if err != nil {
			return true, err
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return true, encoder.Encode(schema)
	default:
		return false, nil
	}
}

// PrintParsedFields writes the parsed fields and their provenance in YAML.
func PrintParsedFields(w io.Writer, parsedValues *values.Values) error {
	sectionsMap := map[string]map[string]interface{}{}
	parsedValues.ForEach(func(sectionName string, sectionValues *values.SectionValues) {
		fieldValues := map[string]interface{}{}
		sectionValues.Fields.ForEach(func(name string, fieldValue *fields.FieldValue) {
			serializable := fields.ToSerializableFieldValue(fieldValue)
			fieldMap := map[string]interface{}{
				"value": serializable.Value,
			}
			if len(serializable.Log) > 0 {
				logs := make([]map[string]interface{}, 0, len(serializable.Log))
				for _, l := range serializable.Log {
					logEntry := map[string]interface{}{
						"source": l.Source,
						"value":  l.Value,
					}
					if len(l.Metadata) > 0 {
						logEntry["metadata"] = l.Metadata
					}
					logs = append(logs, logEntry)
				}
				fieldMap["log"] = logs
			}
			fieldValues[name] = fieldMap
		})
		sectionsMap[sectionName] = fieldValues
	})

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(sectionsMap)
}

func printParsedFields(parsedValues *values.Values) {
	err := PrintParsedFields(os.Stdout, parsedValues)
	cobra.CheckErr(err)
}
