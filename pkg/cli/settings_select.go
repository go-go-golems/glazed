package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed "flags/select.yaml"
var selectFlagsYaml []byte

type SelectSettings struct {
	SelectField    string `glazed.parameter:"select"`
	SelectTemplate string `glazed.parameter:"select-template"`
}

func NewSelectSettingsFromParameters(ps map[string]interface{}) (*SelectSettings, error) {
	s := &SelectSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
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

type SelectParameterLayer struct {
	layers.ParameterLayer
	Settings *SelectSettings
	Defaults *SelectFlagsDefaults
}

func NewSelectParameterLayer() (*SelectParameterLayer, error) {
	ret := &SelectParameterLayer{}
	err := ret.LoadFromYAML(selectFlagsYaml)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize select parameter layer")
	}
	ret.Defaults = &SelectFlagsDefaults{}
	err = ret.InitializeStructFromDefaults(ret.Defaults)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize select flags defaults"))
	}

	return ret, nil
}

func (s *SelectParameterLayer) AddFlags(cmd *cobra.Command) error {
	return s.AddFlagsToCobraCommand(cmd, s.Defaults)
}

func (s *SelectParameterLayer) ParseFlags(cmd *cobra.Command) error {
	parameters, err := s.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return err
	}

	res, err := NewSelectSettingsFromParameters(parameters)
	if err != nil {
		return err
	}

	s.Settings = res

	return nil
}
