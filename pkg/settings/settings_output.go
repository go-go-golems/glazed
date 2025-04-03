package settings

import (
	_ "embed"
	"fmt"
	"text/template"
	"unicode/utf8"

	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/csv"
	"github.com/go-go-golems/glazed/pkg/formatters/excel"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	"github.com/go-go-golems/glazed/pkg/formatters/sql"
	tableformatter "github.com/go-go-golems/glazed/pkg/formatters/table"
	templateformatter "github.com/go-go-golems/glazed/pkg/formatters/template"
	"github.com/go-go-golems/glazed/pkg/formatters/yaml"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/pkg/errors"
)

// TemplateFormatterSettings is probably obsolete...
type TemplateFormatterSettings struct {
	TemplateFuncMaps []template.FuncMap
}

type OutputFormatterSettings struct {
	Output                    string                 `glazed.parameter:"output"`
	OutputFile                string                 `glazed.parameter:"output-file"`
	OutputFileTemplate        string                 `glazed.parameter:"output-file-template"`
	OutputMultipleFiles       bool                   `glazed.parameter:"output-multiple-files"`
	Stream                    bool                   `glazed.parameter:"stream"`
	SheetName                 string                 `glazed.parameter:"sheet-name"`
	TableFormat               string                 `glazed.parameter:"table-format"`
	TableStyle                string                 `glazed.parameter:"table-style"`
	TableStyleFile            string                 `glazed.parameter:"table-style-file"`
	PrintTableStyle           bool                   `glazed.parameter:"print-table-style"`
	OutputAsObjects           bool                   `glazed.parameter:"output-as-objects"`
	FlattenObjects            bool                   `glazed.parameter:"flatten"`
	WithHeaders               bool                   `glazed.parameter:"with-headers"`
	CsvSeparator              string                 `glazed.parameter:"csv-separator"`
	Template                  string                 `glazed.parameter:"template-file"`
	TemplateData              map[string]interface{} `glazed.parameter:"template-data"`
	TemplateFormatterSettings *TemplateFormatterSettings
	SqlTableName              string `glazed.parameter:"sql-table-name"`
	WithUpsert                bool   `glazed.parameter:"sql-upsert"`
	SqlSplitByRows            int    `glazed.parameter:"sql-split-by-rows"`
}

//go:embed "flags/output.yaml"
var outputFlagsYaml []byte

type OutputParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

func NewOutputParameterLayer(options ...layers.ParameterLayerOptions) (*OutputParameterLayer, error) {
	ret := &OutputParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(outputFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func (f *OutputParameterLayer) Clone() layers.ParameterLayer {
	return &OutputParameterLayer{
		ParameterLayerImpl: f.ParameterLayerImpl.Clone().(*layers.ParameterLayerImpl),
	}
}

func NewOutputFormatterSettings(glazedLayer *layers.ParsedLayer) (*OutputFormatterSettings, error) {
	s := &OutputFormatterSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize output formatter settings")
	}

	return s, nil
}

func (ofs *OutputFormatterSettings) computeCanonicalFormat() error {
	switch ofs.Output {
	case "csv":
		ofs.Output = "table"
		ofs.TableFormat = "csv"
	case "tsv":
		ofs.Output = "table"
		ofs.TableFormat = "tsv"
	case "markdown":
		ofs.Output = "table"
		ofs.TableFormat = "markdown"
	case "html":
		ofs.Output = "table"
		ofs.TableFormat = "html"
	}

	if ofs.OutputMultipleFiles {
		if ofs.OutputFileTemplate == "" && ofs.OutputFile == "" {
			return errors.New("output-file or output-file-template is required for output-multiple-files")
		}
	}

	return nil
}

type ErrorUnknownFormat struct {
	format string
}

type ErrorRowFormatUnsupported struct {
	format string
}

type ErrorTableFormatUnsupported struct {
	format string
}

func (e *ErrorUnknownFormat) Error() string {
	return fmt.Sprintf("output format %s is not supported", e.format)
}

func (e *ErrorRowFormatUnsupported) Error() string {
	return fmt.Sprintf("row output format %s is not supported", e.format)
}

func (e *ErrorTableFormatUnsupported) Error() string {
	return fmt.Sprintf("table output format %s is not supported", e.format)
}

func (ofs *OutputFormatterSettings) CreateRowOutputFormatter() (formatters.RowOutputFormatter, error) {
	err := ofs.computeCanonicalFormat()
	if err != nil {
		return nil, err
	}

	var of formatters.RowOutputFormatter
	switch ofs.Output {
	case "json":
		// JSON can always be output as individual rows, since we don't need to know the column names up front
		of = json.NewOutputFormatter(
			json.WithOutputIndividualRows(ofs.OutputAsObjects),
			json.WithOutputFile(ofs.OutputFile),
			json.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			json.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	case "table":
		if ofs.Stream {
			switch ofs.TableFormat {
			case "csv":
				csvOf := csv.NewCSVOutputFormatter(
					csv.WithOutputFile(ofs.OutputFile),
					csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
					csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
				)
				csvOf.WithHeaders = ofs.WithHeaders
				r, _ := utf8.DecodeRuneInString(ofs.CsvSeparator)
				csvOf.Separator = r
				of = csvOf
			case "tsv":
				tsvOf := csv.NewTSVOutputFormatter(
					csv.WithOutputFile(ofs.OutputFile),
					csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
					csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
				)
				tsvOf.WithHeaders = ofs.WithHeaders
				of = tsvOf
			case "html", "markdown":
				of = tableformatter.NewOutputFormatter(ofs.TableFormat)
			default:
				return nil, &ErrorRowFormatUnsupported{ofs.Output + ":" + ofs.TableFormat}
			}
		} else {
			// table and csv also support table output
			return nil, &ErrorRowFormatUnsupported{ofs.Output + ":" + ofs.TableFormat}
		}
	case "yaml":
		of = yaml.NewOutputFormatter(
			yaml.WithOutputIndividualRows(ofs.OutputAsObjects),
			yaml.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			yaml.WithOutputFileTemplate(ofs.OutputFileTemplate),
			yaml.WithYAMLOutputFile(ofs.OutputFile),
		)
	case "excel":
		if ofs.OutputFile == "" {
			return nil, errors.New("output-file is required for excel output")
		}
		if ofs.OutputMultipleFiles {
			return nil, errors.New("output-multiple-files is not supported for excel output")
		}
		of = excel.NewOutputFormatter(
			excel.WithSheetName(ofs.SheetName),
			excel.WithOutputFile(ofs.OutputFile),
		)
	case "sql":
		of = sql.NewOutputFormatter(
			sql.WithTableName(ofs.SqlTableName),
			sql.WithUseUpsert(ofs.WithUpsert),
			sql.WithSplitByRows(ofs.SqlSplitByRows),
		)
	case "template":
		return nil, &ErrorRowFormatUnsupported{"template"}
	default:
		return nil, &ErrorUnknownFormat{ofs.Output}
	}

	return of, nil
}

func (ofs *OutputFormatterSettings) CreateTableOutputFormatter() (formatters.TableOutputFormatter, error) {
	err := ofs.computeCanonicalFormat()
	if err != nil {
		return nil, err
	}

	var of formatters.TableOutputFormatter
	switch ofs.Output {
	case "json":
		of = json.NewOutputFormatter(
			json.WithOutputIndividualRows(ofs.OutputAsObjects),
			json.WithOutputFile(ofs.OutputFile),
			json.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			json.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	case "yaml":
		of = yaml.NewOutputFormatter(
			yaml.WithYAMLOutputFile(ofs.OutputFile),
			yaml.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			yaml.WithOutputFileTemplate(ofs.OutputFileTemplate),
			yaml.WithOutputIndividualRows(ofs.OutputAsObjects),
		)
	case "excel":
		return nil, &ErrorTableFormatUnsupported{"excel"}
	case "table":
		switch ofs.TableFormat {
		case "csv":
			csvOf := csv.NewCSVOutputFormatter(
				csv.WithOutputFile(ofs.OutputFile),
				csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
			)
			csvOf.WithHeaders = ofs.WithHeaders
			r, _ := utf8.DecodeRuneInString(ofs.CsvSeparator)
			csvOf.Separator = r
			of = csvOf
		case "tsv":
			tsvOf := csv.NewTSVOutputFormatter(
				csv.WithOutputFile(ofs.OutputFile),
				csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
			)
			tsvOf.WithHeaders = ofs.WithHeaders
			of = tsvOf
		default:
			of = tableformatter.NewOutputFormatter(
				ofs.TableFormat,
				tableformatter.WithOutputFile(ofs.OutputFile),
				tableformatter.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				tableformatter.WithOutputFileTemplate(ofs.OutputFileTemplate),
				tableformatter.WithTableStyle(ofs.TableStyle),
				tableformatter.WithTableStyleFile(ofs.TableStyleFile),
				tableformatter.WithPrintTableStyle(ofs.PrintTableStyle),
			)
		}
	case "template":
		if ofs.TemplateFormatterSettings == nil {
			ofs.TemplateFormatterSettings = &TemplateFormatterSettings{
				TemplateFuncMaps: []template.FuncMap{
					sprig.TxtFuncMap(),
					templating.TemplateFuncs,
				},
			}
		}
		of = templateformatter.NewOutputFormatter(
			ofs.Template,
			templateformatter.WithTemplateFuncMaps(ofs.TemplateFormatterSettings.TemplateFuncMaps),
			templateformatter.WithAdditionalData(ofs.TemplateData),
			templateformatter.WithOutputFile(ofs.OutputFile),
			templateformatter.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			templateformatter.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	default:
		return nil, &ErrorUnknownFormat{ofs.Output}
	}

	return of, nil
}
