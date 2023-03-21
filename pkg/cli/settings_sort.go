package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/pkg/errors"
)

//go:embed "flags/sort.yaml"
var sortFlagsYaml []byte

type SortFlagsSettings struct {
	SortBy []string `glazed.parameter:"sort-by"`
}

type SortParameterLayer struct {
	*layers.ParameterLayerImpl
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

func (s *SortParameterLayer) AddMiddlewares(of formatters.OutputFormatter) {
}
