package cli

import (
	_ "embed"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"text/template"
	"unicode/utf8"
)

type TemplateFormatterSettings struct {
	TemplateFuncMaps []template.FuncMap
	AdditionalData   map[string]interface{} `glazed.parameter:"template-data"`
}

type OutputFormatterSettings struct {
	Output                    string `glazed.parameter:"output"`
	TableFormat               string `glazed.parameter:"table-format"`
	OutputAsObjects           bool   `glazed.parameter:"output-as-objects"`
	FlattenObjects            bool   `glazed.parameter:"flatten"`
	WithHeaders               bool   `glazed.parameter:"with-headers"`
	CsvSeparator              string `glazed.parameter:"csv-separator"`
	Template                  string
	TemplateFormatterSettings *TemplateFormatterSettings
}

type OutputFlagsDefaults struct {
	Output          string `glazed.parameter:"output"`
	OutputFile      string `glazed.parameter:"output-file"`
	TableFormat     string `glazed.parameter:"table-format"`
	WithHeaders     bool   `glazed.parameter:"with-headers"`
	CsvSeparator    string `glazed.parameter:"csv-separator"`
	OutputAsObjects bool   `glazed.parameter:"output-as-objects"`
	Flatten         bool   `glazed.parameter:"flatten"`
	TemplateFile    string `glazed.parameter:"template-file"`
}

//go:embed "flags/output.yaml"
var outputFlagsYaml []byte

type OutputParameterLayer struct {
	layers.ParameterLayerImpl
	Defaults *OutputFlagsDefaults
}

func NewOutputParameterLayer() (*OutputParameterLayer, error) {
	ret := &OutputParameterLayer{}
	err := ret.LoadFromYAML(outputFlagsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize output parameter layer")
	}
	// TODO(manuel, 2023-02-22) I'm really not sure what these defaults are about here
	//
	// The base idea is that you can update a layer with your own defaults before passing it downstream
	// and that might be done with a struct that automatically gets loaded. That's useful
	// because you can quickly overload stuff with things parsed from a yaml file
	// (for example, the factory section in geppetto command yaml).
	// But when configuring things specifically for certain verb, or by allowing overloads
	// in command and layer definition, something like a ParameterDefinition.SetDefault() might work better
	//
	// In fact we might just be doing the opposite of what we should be doing here,
	// which is actually initializing the parameter defaults from an (optional) default
	// struct.
	//
	// See https://github.com/go-go-golems/glazed/issues/161
	ret.Defaults = &OutputFlagsDefaults{}
	err = ret.InitializeStructFromParameterDefaults(ret.Defaults)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize output flags defaults")
	}

	return ret, nil
}

func (opl *OutputParameterLayer) AddFlagsToCobraCommand(cmd *cobra.Command, s interface{}) error {
	if s == nil {
		s = opl.Defaults
	}
	return opl.ParameterLayerImpl.AddFlagsToCobraCommand(cmd, s)
}

func NewOutputFormatterSettings(ps map[string]interface{}) (*OutputFormatterSettings, error) {
	s := &OutputFormatterSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize output formatter settings")
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

	var of formatters.OutputFormatter
	if ofs.Output == "json" {
		of = formatters.NewJSONOutputFormatter(ofs.OutputAsObjects)
	} else if ofs.Output == "yaml" {
		of = formatters.NewYAMLOutputFormatter()
	} else if ofs.Output == "table" {
		if ofs.TableFormat == "csv" {
			csvOf := formatters.NewCSVOutputFormatter()
			csvOf.WithHeaders = ofs.WithHeaders
			r, _ := utf8.DecodeRuneInString(ofs.CsvSeparator)
			csvOf.Separator = r
			of = csvOf
		} else if ofs.TableFormat == "tsv" {
			tsvOf := formatters.NewTSVOutputFormatter()
			tsvOf.WithHeaders = ofs.WithHeaders
			of = tsvOf
		} else {
			of = formatters.NewTableOutputFormatter(ofs.TableFormat)
		}
		of.AddTableMiddleware(middlewares.NewFlattenObjectMiddleware())
	} else if ofs.Output == "template" {
		if ofs.TemplateFormatterSettings == nil {
			ofs.TemplateFormatterSettings = &TemplateFormatterSettings{
				TemplateFuncMaps: []template.FuncMap{
					sprig.TxtFuncMap(),
					helpers.TemplateFuncs,
				},
				AdditionalData: make(map[string]interface{}),
			}
		}
		of = formatters.NewTemplateOutputFormatter(
			ofs.Template,
			ofs.TemplateFormatterSettings.TemplateFuncMaps,
			ofs.TemplateFormatterSettings.AdditionalData,
		)
	} else {
		return nil, errors.Errorf("Unknown output format: " + ofs.Output)
	}

	return of, nil
}
