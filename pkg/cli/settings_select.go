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
	SelectField     string `glazed.parameter:"select"`
	SelectSeparator string `glazed.parameter:"select-separator"`
	SelectTemplate  string `glazed.parameter:"select-template"`
}

func NewSelectSettingsFromParameters(ps map[string]interface{}) (*SelectSettings, error) {
	s := &SelectSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize select settings from parameters")
	}

	return s, nil
}

func (tf *TemplateSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectTemplate != "" {
		tf.Templates = map[types.FieldName]string{
			"_0": ss.SelectTemplate,
		}
	}
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
