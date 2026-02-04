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
	SelectField     string `glazed:"select"`
	SelectSeparator string `glazed:"select-separator"`
	SelectTemplate  string `glazed:"select-template"`
}

func NewSelectSettingsFromValues(glazedValues *values.SectionValues) (*SelectSettings, error) {
	s := &SelectSettings{}
	err := glazedValues.Fields.DecodeInto(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize select settings from fields")
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

type SelectSection struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewSelectSection(options ...schema.SectionOption) (*SelectSection, error) {
	ret := &SelectSection{}
	section, err := schema.NewSectionFromYAML(selectFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create select field section")
	}
	ret.SectionImpl = section

	return ret, nil
}

func (f *SelectSection) Clone() schema.Section {
	return &SelectSection{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}
