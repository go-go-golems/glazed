package settings

import (
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/csv"
	"github.com/go-go-golems/glazed/pkg/formatters/excel"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	tableformatter "github.com/go-go-golems/glazed/pkg/formatters/table"
	templateformatter "github.com/go-go-golems/glazed/pkg/formatters/template"
	"github.com/go-go-golems/glazed/pkg/formatters/yaml"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/pkg/errors"
	"text/template"
	"unicode/utf8"
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
}

//go:embed "flags/output.yaml"
var outputFlagsYaml []byte

type OutputParameterLayer struct {
	*layers.ParameterLayerImpl
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

func NewOutputFormatterSettings(ps map[string]interface{}) (*OutputFormatterSettings, error) {
	s := &OutputFormatterSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize output formatter settings")
	}

	return s, nil
}

func (ofs *OutputFormatterSettings) computeCanonicalFormat() error {
	if ofs.Output == "csv" {
		ofs.Output = "table"
		ofs.TableFormat = "csv"
	} else if ofs.Output == "tsv" {
		ofs.Output = "table"
		ofs.TableFormat = "tsv"
	} else if ofs.Output == "markdown" {
		ofs.Output = "table"
		ofs.TableFormat = "markdown"
	} else if ofs.Output == "html" {
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
	if ofs.Output == "json" {
		// JSON can always be output as individual rows, since we don't need to know the column names up front
		of = json.NewOutputFormatter(
			json.WithOutputIndividualRows(ofs.OutputAsObjects),
			json.WithOutputFile(ofs.OutputFile),
			json.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			json.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	} else if ofs.Output == "table" {
		if ofs.Stream {
			if ofs.TableFormat == "csv" {
				csvOf := csv.NewCSVOutputFormatter(
					csv.WithOutputFile(ofs.OutputFile),
					csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
					csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
				)
				csvOf.WithHeaders = ofs.WithHeaders
				r, _ := utf8.DecodeRuneInString(ofs.CsvSeparator)
				csvOf.Separator = r
				of = csvOf
			} else if ofs.TableFormat == "tsv" {
				tsvOf := csv.NewTSVOutputFormatter(
					csv.WithOutputFile(ofs.OutputFile),
					csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
					csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
				)
				tsvOf.WithHeaders = ofs.WithHeaders
				of = tsvOf
			} else if ofs.TableFormat == "html" || ofs.TableFormat == "markdown" {
				of = tableformatter.NewOutputFormatter(ofs.TableFormat)
			} else {
				return nil, &ErrorRowFormatUnsupported{ofs.Output + ":" + ofs.TableFormat}
			}
		} else {
			// table and csv also support table output
			return nil, &ErrorRowFormatUnsupported{ofs.Output + ":" + ofs.TableFormat}

		}
	} else if ofs.Output == "yaml" {
		return nil, &ErrorRowFormatUnsupported{"yaml"}
	} else if ofs.Output == "excel" {
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
	} else if ofs.Output == "template" {
		return nil, &ErrorRowFormatUnsupported{"template"}
	} else {
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
	if ofs.Output == "json" {
		of = json.NewOutputFormatter(
			json.WithOutputIndividualRows(ofs.OutputAsObjects),
			json.WithOutputFile(ofs.OutputFile),
			json.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			json.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	} else if ofs.Output == "yaml" {
		of = yaml.NewOutputFormatter(
			yaml.WithYAMLOutputFile(ofs.OutputFile),
			yaml.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			yaml.WithOutputFileTemplate(ofs.OutputFileTemplate),
			yaml.WithOutputIndividualRows(ofs.OutputAsObjects),
		)
	} else if ofs.Output == "excel" {
		return nil, &ErrorTableFormatUnsupported{"excel"}
	} else if ofs.Output == "table" {
		if ofs.TableFormat == "csv" {
			csvOf := csv.NewCSVOutputFormatter(
				csv.WithOutputFile(ofs.OutputFile),
				csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
			)
			csvOf.WithHeaders = ofs.WithHeaders
			r, _ := utf8.DecodeRuneInString(ofs.CsvSeparator)
			csvOf.Separator = r
			of = csvOf
		} else if ofs.TableFormat == "tsv" {
			tsvOf := csv.NewTSVOutputFormatter(
				csv.WithOutputFile(ofs.OutputFile),
				csv.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				csv.WithOutputFileTemplate(ofs.OutputFileTemplate),
			)
			tsvOf.WithHeaders = ofs.WithHeaders
			of = tsvOf
		} else {
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
	} else if ofs.Output == "template" {
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
	} else {
		return nil, &ErrorUnknownFormat{ofs.Output}
	}

	return of, nil
}
