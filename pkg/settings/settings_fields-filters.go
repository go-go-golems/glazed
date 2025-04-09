package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed "flags/fields-filters.yaml"
var fieldsFiltersFlagsYaml []byte

type FieldsFilterFlagsDefaults struct {
	Fields           []string `glazed.parameter:"fields"`
	Filter           []string `glazed.parameter:"filter"`
	SortColumns      bool     `glazed.parameter:"sort-columns"`
	RemoveNulls      bool     `glazed.parameter:"remove-nulls"`
	RemoveDuplicates []string `glazed.parameter:"remove-duplicates"`
}

type FieldsFiltersParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

var _ layers.CobraParameterLayer = &FieldsFiltersParameterLayer{}
var _ layers.ParameterLayer = &FieldsFiltersParameterLayer{}

func (f *FieldsFiltersParameterLayer) Clone() layers.ParameterLayer {
	return &FieldsFiltersParameterLayer{
		ParameterLayerImpl: f.ParameterLayerImpl.Clone().(*layers.ParameterLayerImpl),
	}
}

type FieldsFilterSettings struct {
	Filters          []string `glazed.parameter:"filter"`
	Fields           []string `glazed.parameter:"fields"`
	SortColumns      bool     `glazed.parameter:"sort-columns"`
	RemoveNulls      bool     `glazed.parameter:"remove-nulls"`
	RemoveDuplicates []string `glazed.parameter:"remove-duplicates"`
	ReorderColumns   []string `glazed.parameter:"reorder-columns"`
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

func (f *FieldsFiltersParameterLayer) AddLayerToCobraCommand(cmd *cobra.Command) error {
	defaults := &FieldsFilterFlagsDefaults{}
	err := f.InitializeStructFromParameterDefaults(defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize fields and filters flags defaults")
	}
	// this is not very elegant, as a new way of doing defaults
	defaultFieldHelp := defaults.Fields
	if len(defaultFieldHelp) == 0 || (len(defaultFieldHelp) == 1 && defaultFieldHelp[0] == "") {
		defaults.Fields = []string{"all"}
	}
	// this would be more elegant with a middleware for handling defaults, I think
	err = f.InitializeParameterDefaultsFromStruct(defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize fields and filters flags defaults")
	}

	return f.ParameterLayerImpl.AddLayerToCobraCommand(cmd)
}

func (f *FieldsFiltersParameterLayer) ParseLayerFromCobraCommand(
	cmd *cobra.Command,
	options ...parameters.ParseStepOption,
) (*layers.ParsedLayer, error) {
	l, err := f.ParameterLayerImpl.ParseLayerFromCobraCommand(cmd, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to gather fields and filters flags from cobra command")
	}

	// if fields were manually specified, clear whatever default filters we might have set
	// TODO(manuel, 2023-12-28) This should be moved to somewhere outside of the cobra parsing, I think
	// This means we'd have to store if a flag was changed in the parsed layer
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		parsedFilter, ok := l.Parameters.Get("filter")
		options_ := append(options, parameters.WithParseStepSource("override-fields-filter"))
		if !ok {
			pd, ok := f.ParameterDefinitions.Get("filter")
			if !ok {
				return nil, errors.New("Failed to find default filter parameter definition")
			}
			p := &parameters.ParsedParameter{
				ParameterDefinition: pd,
			}
			err := p.Update([]string{}, options_...)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to update filter parameter")
			}
			l.Parameters.Set("filter", p)
		} else {
			err := parsedFilter.Update([]string{}, options_...)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to update filter parameter")
			}
		}
	}

	return l, nil
}

func NewFieldsFilterSettings(glazedLayer *layers.ParsedLayer) (*FieldsFilterSettings, error) {
	s := &FieldsFilterSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, err
	}

	if len(s.Fields) == 1 && s.Fields[0] == "all" {
		s.Fields = []string{}
	}
	if s.ReorderColumns == nil {
		s.ReorderColumns = s.Fields
	}
	return s, nil
}

func (ffs *FieldsFilterSettings) AddMiddlewares(p_ *middlewares.TableProcessor) {
	p_.AddRowMiddleware(row.NewFieldsFilterMiddleware(row.WithFields(ffs.Fields), row.WithFilters(ffs.Filters)))
	if ffs.RemoveNulls {
		p_.AddRowMiddleware(row.NewRemoveNullsMiddleware())
	}
	if ffs.SortColumns {
		p_.AddRowMiddleware(row.NewSortColumnsMiddleware())
	}
	if len(ffs.ReorderColumns) > 0 {
		p_.AddRowMiddleware(row.NewReorderColumnOrderMiddleware(ffs.ReorderColumns))
	}
	if len(ffs.RemoveDuplicates) > 0 {
		p_.AddRowMiddleware(row.NewRemoveDuplicatesMiddleware(ffs.RemoveDuplicates...))
	}
}
