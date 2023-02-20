package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed "flags/fields-filters.yaml"
var fieldsFiltersFlagsYaml []byte

var fieldsFiltersFlagsParameters map[string]*cmds.ParameterDefinition
var fieldsFiltersFlagsParametersList []*cmds.ParameterDefinition

func init() {
	fieldsFiltersFlagsParameters, fieldsFiltersFlagsParametersList = cmds.InitFlagsFromYaml(fieldsFiltersFlagsYaml)
}

type FieldsFilterSettings struct {
	Filters        []string `glazed.parameter:"filter"`
	Fields         []string `glazed.parameter:"fields"`
	SortColumns    bool     `glazed.parameter:"sort-columns"`
	ReorderColumns []string
}

func NewFieldsFilterSettings(parameters map[string]interface{}) (*FieldsFilterSettings, error) {
	s := &FieldsFilterSettings{}
	err := cmds.InitializeStructFromParameters(s, parameters)
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

// TODO(manuel, 2022-11-20) Make it easy for the developer to configure which flag they want
// and which they don't
//
// See https://github.com/go-go-golems/glazed/issues/130

type FieldsFilterFlagsDefaults struct {
	Fields      []string `glazed.parameter:"fields"`
	Filter      []string `glazed.parameter:"filter"`
	SortColumns bool     `glazed.parameter:"sort-columns"`
}

func NewFieldsFilterFlagsDefaults() *FieldsFilterFlagsDefaults {
	s := &FieldsFilterFlagsDefaults{}
	err := cmds.InitializeStructFromParameterDefinitions(s, fieldsFiltersFlagsParameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize fields and filters flags defaults"))
	}

	return s
}

// AddFieldsFilterFlags adds the flags for the following middlewares to the cmd:
// - FieldsFilterMiddleware
// - SortColumnsMiddleware
// - ReorderColumnOrderMiddleware
func AddFieldsFilterFlags(cmd *cobra.Command, defaults *FieldsFilterFlagsDefaults) error {
	defaultFieldHelp := defaults.Fields
	if len(defaultFieldHelp) == 0 || (len(defaultFieldHelp) == 1 && defaultFieldHelp[0] == "") {
		defaults.Fields = []string{"all"}
	}
	parameters, err := cmds.CloneParameterDefinitionsWithDefaultsStruct(fieldsFiltersFlagsParametersList, defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to clone fields and filters flags parameters")
	}
	err = cmds.AddFlagsToCobraCommand(cmd.PersistentFlags(), parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to add fields and filters flags to cobra command")
	}

	cmds.AddFlagGroupToCobraCommand(cmd, "fields-filters", "Glazed fields filtering", parameters)

	return nil
}

func ParseFieldsFilterFlags(cmd *cobra.Command) (*FieldsFilterSettings, error) {
	parameters, err := cmds.GatherFlagsFromCobraCommand(cmd, fieldsFiltersFlagsParametersList, false)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to gather fields and filters flags from cobra command")
	}

	res, err := NewFieldsFilterSettings(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create fields and filters settings from parameters")
	}

	// if fields were manually specified, clear whatever default filters we might have set
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		res.Filters = []string{}
	}

	return res, nil
}
