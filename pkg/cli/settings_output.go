package cli

import (
	_ "embed"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/cmds"
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
	cmds.ParameterLayer
	Settings *OutputFormatterSettings
	Defaults *OutputFlagsDefaults
}

func NewOutputParameterLayer() (*OutputParameterLayer, error) {
	ret := &OutputParameterLayer{}
	err := ret.LoadFromYAML(outputFlagsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize output parameter layer")
	}
	ret.Defaults = &OutputFlagsDefaults{}
	err = ret.InitializeStructFromDefaults(ret.Defaults)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize output flags defaults")
	}

	return ret, nil
}

func (opl *OutputParameterLayer) AddFlags(cmd *cobra.Command) error {
	return opl.AddFlagsToCobraCommand(cmd, opl.Defaults)
}

func (opl *OutputParameterLayer) ParseFlags(cmd *cobra.Command) error {
	// TODO(manuel, 2023-02-12): This is not enough, because the flags template-file is not handled properly by just parsing it into here
	// Really what this should be parsed into is a defaults struct, and then loading that into the settings by hand
	parameters, err := opl.ParseFlagsFromCobraCommand(cmd)

	if err != nil {
		return err
	}

	// TODO(manuel, 2023-02-21) This part should actually be outside of the cobra handling too
	// See #150
	opl.Settings, err = NewOutputFormatterSettings(parameters)
	if err != nil {
		return err
	}

	return nil
}

func NewOutputFormatterSettings(parameters map[string]interface{}) (*OutputFormatterSettings, error) {
	s := &OutputFormatterSettings{}
	err := cmds.InitializeStructFromParameters(s, parameters)
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
