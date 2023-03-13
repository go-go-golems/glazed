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

type FieldsFilterFlagsDefaults struct {
	Fields      []string `glazed.parameter:"fields"`
	Filter      []string `glazed.parameter:"filter"`
	SortColumns bool     `glazed.parameter:"sort-columns"`
	RemoveNulls bool     `glazed.parameter:"remove-nulls"`
}

type FieldsFiltersParameterLayer struct {
	*layers.ParameterLayerImpl
}

type FieldsFilterSettings struct {
	Filters        []string `glazed.parameter:"filter"`
	Fields         []string `glazed.parameter:"fields"`
	SortColumns    bool     `glazed.parameter:"sort-columns"`
	RemoveNulls    bool     `glazed.parameter:"remove-nulls"`
	ReorderColumns []string
}

func NewFieldsFiltersParameterLayer(options ...layers.ParameterLayerOptions) (*FieldsFiltersParameterLayer, error) {
	ret := &FieldsFiltersParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(fieldsFiltersFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create fields and filters parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func (f *FieldsFiltersParameterLayer) AddFlagsToCobraCommand(cmd *cobra.Command) error {
	defaults := &FieldsFilterFlagsDefaults{}
	err := f.ParameterLayerImpl.InitializeStructFromParameterDefaults(defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize fields and filters flags defaults")
	}
	// this is not very elegant, as a new way of doing defaults
	defaultFieldHelp := defaults.Fields
	if len(defaultFieldHelp) == 0 || (len(defaultFieldHelp) == 1 && defaultFieldHelp[0] == "") {
		defaults.Fields = []string{"all"}
	}
	// this would be more elegant with a middleware for handling defaults, I think
	err = f.ParameterLayerImpl.InitializeParameterDefaultsFromStruct(defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize fields and filters flags defaults")
	}

	return f.ParameterLayerImpl.AddFlagsToCobraCommand(cmd)
}

func (f *FieldsFiltersParameterLayer) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	ps, err := f.ParameterLayerImpl.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to gather fields and filters flags from cobra command")
	}

	// if fields were manually specified, clear whatever default filters we might have set
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		ps["filter"] = []string{}
	}

	return ps, nil
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
	s.ReorderColumns = s.Fields
	return s, nil
}

func (ffs *FieldsFilterSettings) AddMiddlewares(of formatters.OutputFormatter) {
	of.AddTableMiddleware(middlewares.NewFieldsFilterMiddleware(ffs.Fields, ffs.Filters))
	if ffs.RemoveNulls {
		of.AddTableMiddleware(middlewares.NewRemoveNullsMiddleware())
	}
	if ffs.SortColumns {
		of.AddTableMiddleware(middlewares.NewSortColumnsMiddleware())
	}
	if len(ffs.ReorderColumns) > 0 {
		of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware(ffs.ReorderColumns))
	}
}
