package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed "flags/fields-filters.yaml"
var fieldsFiltersFlagsYaml []byte

// TODO(manuel, 2022-11-20) Make it easy for the developer to configure which flag they want
// and which they don't
//
// See https://github.com/go-go-golems/glazed/issues/130

type FieldsFilterFlagsDefaults struct {
	Fields      []string `glazed.parameter:"fields"`
	Filter      []string `glazed.parameter:"filter"`
	SortColumns bool     `glazed.parameter:"sort-columns"`
}

type FieldsFiltersParameterLayer struct {
	layers.ParameterLayer
	Settings *FieldsFilterSettings
	Defaults *FieldsFilterFlagsDefaults
}

type FieldsFilterSettings struct {
	Filters        []string `glazed.parameter:"filter"`
	Fields         []string `glazed.parameter:"fields"`
	SortColumns    bool     `glazed.parameter:"sort-columns"`
	ReorderColumns []string
}

func NewFieldsFiltersParameterLayer() (*FieldsFiltersParameterLayer, error) {
	ret := &FieldsFiltersParameterLayer{}
	err := ret.LoadFromYAML(fieldsFiltersFlagsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize fields and filters parameter layer")
	}
	ret.Defaults = &FieldsFilterFlagsDefaults{}
	err = ret.InitializeStructFromDefaults(ret.Defaults)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize fields and filters flags defaults")
	}
	return ret, nil
}

func (f *FieldsFiltersParameterLayer) AddFlags(cmd *cobra.Command) error {
	defaultFieldHelp := f.Defaults.Fields
	if len(defaultFieldHelp) == 0 || (len(defaultFieldHelp) == 1 && defaultFieldHelp[0] == "") {
		f.Defaults.Fields = []string{"all"}
	}
	return f.AddFlagsToCobraCommand(cmd, f.Defaults)
}

func (f *FieldsFiltersParameterLayer) ParseFlags(cmd *cobra.Command) error {
	parameters, err := f.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return errors.Wrap(err, "Failed to gather fields and filters flags from cobra command")
	}

	res, err := NewFieldsFilterSettings(parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to create fields and filters settings from parameters")
	}

	// if fields were manually specified, clear whatever default filters we might have set
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		res.Filters = []string{}
	}

	f.Settings = res

	return nil
}

func NewFieldsFilterSettings(ps map[string]interface{}) (*FieldsFilterSettings, error) {
	s := &FieldsFilterSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, err
	}

	if len(s.Fields) == 1 && s.Fields[0] == "all" {
		s.Fields = []string{}
	}
	return s, nil
}

func (ffs *FieldsFilterSettings) AddMiddlewares(of formatters.OutputFormatter) {
	of.AddTableMiddleware(middlewares.NewFieldsFilterMiddleware(ffs.Fields, ffs.Filters))
	if ffs.SortColumns {
		of.AddTableMiddleware(middlewares.NewSortColumnsMiddleware())
	}
	if len(ffs.ReorderColumns) > 0 {
		of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware(ffs.ReorderColumns))
	}
}
