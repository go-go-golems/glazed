package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
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
	*layers.ParameterLayerImpl
}

func NewSelectParameterLayer(options ...layers.ParameterLayerOptions) (*SelectParameterLayer, error) {
	ret := &SelectParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(selectFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create select parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}
