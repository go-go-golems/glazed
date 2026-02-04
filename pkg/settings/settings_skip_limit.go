package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

//go:embed "flags/skip-limit.yaml"
var skipLimitFlagsYaml []byte

type SkipLimitSettings struct {
	Skip  int `glazed:"glazed-skip"`
	Limit int `glazed:"glazed-limit"`
}

func NewSkipLimitSettingsFromValues(glazedValues *values.SectionValues) (*SkipLimitSettings, error) {
	s := &SkipLimitSettings{}
	err := glazedValues.Fields.DecodeInto(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize skipLimit settings from fields")
	}

	return s, nil
}

type SkipLimitSection struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewSkipLimitSection(options ...schema.SectionOption) (*SkipLimitSection, error) {
	ret := &SkipLimitSection{}
	section, err := schema.NewSectionFromYAML(skipLimitFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create skipLimit field section")
	}
	ret.SectionImpl = section

	return ret, nil
}
func (f *SkipLimitSection) Clone() schema.Section {
	return &SkipLimitSection{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}
