package cli

import (
	_ "embed"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/csv"
	"github.com/go-go-golems/glazed/pkg/formatters/excel"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	table2 "github.com/go-go-golems/glazed/pkg/formatters/table"
	template2 "github.com/go-go-golems/glazed/pkg/formatters/template"
	"github.com/go-go-golems/glazed/pkg/formatters/yaml"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
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
	SheetName                 string                 `glazed.parameter:"sheet-name"`
	TableFormat               string                 `glazed.parameter:"table-format"`
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

	// if template-file is set, use it for Template
	_, ok := ps["template-file"]
	if ok {
		s.Template = ps["template-file"].(string)
	}
	return s, nil
}

func (ofs *OutputFormatterSettings) CreateOutputFormatter() (formatters.OutputFormatter, error) {
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
			return nil, errors.New("output-file or output-file-template is required for output-multiple-files")
		}
	}

	var of formatters.OutputFormatter
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
		)
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
		of.AddTableMiddleware(table.NewFlattenObjectMiddleware())
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
			of = table2.NewOutputFormatter(
				ofs.TableFormat,
				table2.WithOutputFile(ofs.OutputFile),
				table2.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
				table2.WithOutputFileTemplate(ofs.OutputFileTemplate),
			)
		}
		of.AddTableMiddleware(table.NewFlattenObjectMiddleware())
	} else if ofs.Output == "template" {
		if ofs.TemplateFormatterSettings == nil {
			ofs.TemplateFormatterSettings = &TemplateFormatterSettings{
				TemplateFuncMaps: []template.FuncMap{
					sprig.TxtFuncMap(),
					templating.TemplateFuncs,
				},
			}
		}
		of = template2.NewOutputFormatter(
			ofs.Template,
			template2.WithTemplateFuncMaps(ofs.TemplateFormatterSettings.TemplateFuncMaps),
			template2.WithAdditionalData(ofs.TemplateData),
			template2.WithOutputFile(ofs.OutputFile),
			template2.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
			template2.WithOutputFileTemplate(ofs.OutputFileTemplate),
		)
	} else {
		return nil, errors.Errorf("Unknown output format: " + ofs.Output)
	}

	return of, nil
}
