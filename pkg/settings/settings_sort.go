package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
)

//go:embed "flags/sort.yaml"
var sortFlagsYaml []byte

type SortFlagsSettings struct {
	SortBy []string `glazed.parameter:"sort-by"`
}

func NewSortSettingsFromParameters(glazedLayer *layers.ParsedLayer) (*SortFlagsSettings, error) {
	s := &SortFlagsSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize sort settings from parameters")
	}

	return s, nil
}

type SortParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

func NewSortParameterLayer(options ...layers.ParameterLayerOptions) (*SortParameterLayer, error) {
	ret := &SortParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(sortFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create sort parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func (f *SortParameterLayer) Clone() layers.ParameterLayer {
	return &SortParameterLayer{
		ParameterLayerImpl: f.ParameterLayerImpl.Clone().(*layers.ParameterLayerImpl),
	}
}

func (s *SortFlagsSettings) AddMiddlewares(p_ *middlewares.TableProcessor) {
	if len(s.SortBy) == 0 {
		return
	}
	p_.AddTableMiddleware(table.NewSortByMiddlewareFromColumns(s.SortBy...))
}
