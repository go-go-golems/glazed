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
	Skip  int `glazed.parameter:"glazed-skip"`
	Limit int `glazed.parameter:"glazed-limit"`
}

func NewSkipLimitSettingsFromParameters(glazedLayer *values.SectionValues) (*SkipLimitSettings, error) {
	s := &SkipLimitSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize skipLimit settings from parameters")
	}

	return s, nil
}

type SkipLimitParameterLayer struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewSkipLimitParameterLayer(options ...schema.SectionOption) (*SkipLimitParameterLayer, error) {
	ret := &SkipLimitParameterLayer{}
	layer, err := schema.NewSectionFromYAML(skipLimitFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create skipLimit parameter layer")
	}
	ret.SectionImpl = layer

	return ret, nil
}
func (f *SkipLimitParameterLayer) Clone() schema.Section {
	return &SkipLimitParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}
