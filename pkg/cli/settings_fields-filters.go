package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

//go:embed "flags/fields-filters.yaml"
var fieldsFiltersFlagsYaml []byte

var fieldsFiltersFlagsParameters map[string]*cmds.ParameterDefinition
var fieldsFiltersFlagsParametersList []*cmds.ParameterDefinition

func init() {
	fieldsFiltersFlagsParameters, fieldsFiltersFlagsParametersList = initFlagsFromYaml(fieldsFiltersFlagsYaml)
}

type FieldsFilterSettings struct {
	Filters        []string `glazed.parameter:"filter"`
	Fields         []string `glazed.parameter:"fields"`
	SortColumns    bool     `glazed.parameter:"sort-columns"`
	ReorderColumns []string
}

func (fff *FieldsFilterSettings) AddMiddlewares(of formatters.OutputFormatter) {
	of.AddTableMiddleware(middlewares.NewFieldsFilterMiddleware(fff.Fields, fff.Filters))
	if fff.SortColumns {
		of.AddTableMiddleware(middlewares.NewSortColumnsMiddleware())
	}
	if len(fff.ReorderColumns) > 0 {
		of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware(fff.ReorderColumns))
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
func AddFieldsFilterFlags(cmd *cobra.Command, defaults *FieldsFilterFlagsDefaults) {
	defaultFieldHelp := defaults.Fields
	if len(defaultFieldHelp) == 0 || (len(defaultFieldHelp) == 1 && defaultFieldHelp[0] == "") {
		defaults.Fields = []string{"all"}
	}
	parameters, err := cmds.CloneParameterDefinitionsWithDefaultsStruct(fieldsFiltersFlagsParametersList, defaults)
	if err != nil {
		// TODO(manuel, 2023-02-12) This needs proper error handling
		panic(errors.Wrap(err, "Failed to add fields and filters flags to cobra command"))
	}
	err = cmds.AddFlagsToCobraCommand(cmd, parameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to add fields and filters flags to cobra command"))
	}
}

func ParseFieldsFilterFlags(cmd *cobra.Command) (*FieldsFilterSettings, error) {
	fieldStr := cmd.Flag("fields").Value.String()
	filters := []string{}
	fields := []string{}
	if fieldStr != "" {
		fields = strings.Split(fieldStr, ",")
	}
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		filters = []string{}
	} else {
		filterStr := cmd.Flag("filter").Value.String()
		if filterStr != "" {
			filters = strings.Split(filterStr, ",")
		}
	}

	sortColumns, err := cmd.Flags().GetBool("sort-columns")
	if err != nil {
		return nil, err
	}

	return &FieldsFilterSettings{
		Fields:         fields,
		Filters:        filters,
		SortColumns:    sortColumns,
		ReorderColumns: fields,
	}, nil
}
