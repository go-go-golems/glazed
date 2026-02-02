package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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

func NewSelectSettingsFromParameters(glazedLayer *values.SectionValues) (*SelectSettings, error) {
	s := &SelectSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
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
	*schema.SectionImpl `yaml:",inline"`
}

func NewSelectParameterLayer(options ...schema.SectionOption) (*SelectParameterLayer, error) {
	ret := &SelectParameterLayer{}
	layer, err := schema.NewSectionFromYAML(selectFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create select parameter layer")
	}
	ret.SectionImpl = layer

	return ret, nil
}

func (f *SelectParameterLayer) Clone() schema.Section {
	return &SelectParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}
