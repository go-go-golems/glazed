package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

//go:embed "flags/skip-limit.yaml"
var skipLimitFlagsYaml []byte

type SkipLimitSettings struct {
	Skip  int `glazed.parameter:"glazed-skip"`
	Limit int `glazed.parameter:"glazed-limit"`
}

func NewSkipLimitSettingsFromParameters(glazedLayer *layers.ParsedLayer) (*SkipLimitSettings, error) {
	s := &SkipLimitSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize skipLimit settings from parameters")
	}

	return s, nil
}

type SkipLimitParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

func NewSkipLimitParameterLayer(options ...layers.ParameterLayerOptions) (*SkipLimitParameterLayer, error) {
	ret := &SkipLimitParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(skipLimitFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create skipLimit parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}
func (f *SkipLimitParameterLayer) Clone() layers.ParameterLayer {
	return &SkipLimitParameterLayer{
		ParameterLayerImpl: f.ParameterLayerImpl.Clone().(*layers.ParameterLayerImpl),
	}
}
