package settings

import (
	"io"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/csv"
	jsonformatter "github.com/go-go-golems/glazed/pkg/formatters/json"
	tableformatter "github.com/go-go-golems/glazed/pkg/formatters/table"
	yamlformatter "github.com/go-go-golems/glazed/pkg/formatters/yaml"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
)

type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputJSONL OutputFormat = "jsonl"
	OutputCSV   OutputFormat = "csv"
	OutputTSV   OutputFormat = "tsv"
	OutputYAML  OutputFormat = "yaml"
)

const (
	StructuredOutputSlug = "structured-output"
	StructuredOutputFlag = "format"
)

var structuredOutputFormats = []string{
	string(OutputTable),
	string(OutputJSON),
	string(OutputJSONL),
	string(OutputCSV),
	string(OutputTSV),
	string(OutputYAML),
}

type StructuredOutputSettings struct {
	Format        OutputFormat `glazed:"format"`
	OutputFields  []string     `glazed:"output-fields"`
	MaxOutputRows int          `glazed:"max-output-rows"`
}

func NewStructuredOutputSection(options ...schema.SectionOption) (*schema.SectionImpl, error) {
	sectionOptions := []schema.SectionOption{
		schema.WithDescription("Structured stdout serialization settings"),
		schema.WithFields(
			fields.New(
				StructuredOutputFlag,
				fields.TypeChoice,
				fields.WithHelp("Structured output format"),
				fields.WithChoices(structuredOutputFormats...),
				fields.WithDefault(string(OutputTable)),
			),
			fields.New(
				"output-fields",
				fields.TypeStringList,
				fields.WithHelp("Fields to include in output (requested order is preserved by tabular formats)"),
				fields.WithDefault([]string{}),
			),
			fields.New(
				"max-output-rows",
				fields.TypeInteger,
				fields.WithHelp("Maximum number of rows to serialize (0 means unlimited)"),
				fields.WithDefault(0),
			),
		),
	}
	sectionOptions = append(sectionOptions, options...)
	return schema.NewSection(StructuredOutputSlug, "Structured output", sectionOptions...)
}

func DecodeStructuredOutputSettings(sectionValues *values.SectionValues) (*StructuredOutputSettings, error) {
	settings := &StructuredOutputSettings{}
	if err := sectionValues.DecodeInto(settings); err != nil {
		return nil, errors.Wrap(err, "failed to decode structured output settings")
	}
	if settings.MaxOutputRows < 0 {
		return nil, errors.New("max-output-rows must be greater than or equal to zero")
	}

	seen := map[string]struct{}{}
	outputFields := make([]string, 0, len(settings.OutputFields))
	for _, field := range settings.OutputFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		outputFields = append(outputFields, field)
	}
	settings.OutputFields = outputFields
	return settings, nil
}

// SetupStructuredProcessor creates the row-processing portion of structured
// output without attaching a formatter. This is useful for programmatic callers
// that want projected/capped rows as a Table instead of serialized bytes.
func SetupStructuredProcessor(
	sectionValues *values.SectionValues,
	options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, *StructuredOutputSettings, error) {
	settings, err := DecodeStructuredOutputSettings(sectionValues)
	if err != nil {
		return nil, nil, err
	}

	processor := middlewares.NewTableProcessor(options...)
	if len(settings.OutputFields) > 0 {
		processor.AddRowMiddleware(row.NewOutputFieldsMiddleware(settings.OutputFields...))
	}
	if settings.MaxOutputRows > 0 {
		processor.AddRowMiddleware(&row.SkipLimitMiddleware{Limit: settings.MaxOutputRows})
	}
	return processor, settings, nil
}

func SetupStructuredOutput(
	sectionValues *values.SectionValues,
	writer io.Writer,
	options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, formatters.OutputFormatter, error) {
	processor, settings, err := SetupStructuredProcessor(sectionValues, options...)
	if err != nil {
		return nil, nil, err
	}

	formatter, rowOutput, err := newStructuredOutputFormatter(settings.Format)
	if err != nil {
		return nil, nil, err
	}
	if rowOutput {
		rowFormatter := formatter.(formatters.RowOutputFormatter)
		if err := rowFormatter.RegisterRowMiddlewares(processor); err != nil {
			return nil, nil, err
		}
		processor.AddRowMiddleware(row.NewOutputMiddleware(rowFormatter, writer))
	} else {
		tableFormatter := formatter.(formatters.TableOutputFormatter)
		if err := tableFormatter.RegisterTableMiddlewares(processor); err != nil {
			return nil, nil, err
		}
		processor.AddTableMiddleware(table.NewOutputMiddleware(tableFormatter, writer))
	}

	return processor, formatter, nil
}

func newStructuredOutputFormatter(format OutputFormat) (formatters.OutputFormatter, bool, error) {
	switch format {
	case OutputTable:
		return tableformatter.NewOutputFormatter("ascii"), false, nil
	case OutputJSON:
		return jsonformatter.NewArrayOutputFormatter(), true, nil
	case OutputJSONL:
		return jsonformatter.NewLinesOutputFormatter(), true, nil
	case OutputCSV:
		return csv.NewCSVOutputFormatter(), false, nil
	case OutputTSV:
		return csv.NewTSVOutputFormatter(), false, nil
	case OutputYAML:
		return yamlformatter.NewOutputFormatter(), false, nil
	default:
		return nil, false, errors.Errorf("unsupported structured output format %q", format)
	}
}
