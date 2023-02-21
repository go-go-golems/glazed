package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed "flags/select.yaml"
var selectFlagsYaml []byte

var selectParameterLayer *cmds.ParameterLayer

func init() {
	var err error
	selectParameterLayer, err = cmds.NewParameterLayerFromYAML(selectFlagsYaml)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize select parameter layer"))
	}
}

type SelectSettings struct {
	SelectField    string `glazed.parameter:"select"`
	SelectTemplate string `glazed.parameter:"select-template"`
}

func NewSelectSettingsFromParameters(parameters map[string]interface{}) (*SelectSettings, error) {
	s := &SelectSettings{}
	err := cmds.InitializeStructFromParameters(s, parameters)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize select settings from parameters")
	}

	return s, nil
}

func (ofs *OutputFormatterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" || ss.SelectTemplate != "" {
		ofs.Output = "table"
		ofs.TableFormat = "tsv"
		ofs.FlattenObjects = true
		ofs.WithHeaders = false
	}
}

func (ffs *FieldsFilterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" {
		ffs.Fields = []string{ss.SelectField}
	}
}

func (tf *TemplateSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectTemplate != "" {
		tf.Templates = map[types.FieldName]string{
			"_0": ss.SelectTemplate,
		}
	}
}

type SelectFlagsDefaults struct {
	Select         string `glazed.parameter:"select"`
	SelectTemplate string `glazed.parameter:"select-template"`
}

func NewSelectFlagsDefaults() *SelectFlagsDefaults {
	s := &SelectFlagsDefaults{}
	err := selectParameterLayer.InitializeStructFromDefaults(s)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize select flags defaults"))
	}

	return s
}

func AddSelectFlags(cmd *cobra.Command, defaults *SelectFlagsDefaults) error {
	return selectParameterLayer.AddFlagsToCobraCommand(cmd, defaults)
}

func ParseSelectFlags(cmd *cobra.Command) (*SelectSettings, error) {
	parameters, err := selectParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}

	return NewSelectSettingsFromParameters(parameters)
}
