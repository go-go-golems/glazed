package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
)

//go:embed "flags/sort.yaml"
var sortFlagsYaml []byte

type SortFlagsSettings struct {
	SortBy []string `glazed.parameter:"sort-by"`
}

func NewSortSettingsFromParameters(ps map[string]interface{}) (*SortFlagsSettings, error) {
	s := &SortFlagsSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize sort settings from parameters")
	}

	return s, nil
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

func (s *SortFlagsSettings) AddMiddlewares(of formatters.OutputFormatter) {
	if len(s.SortBy) == 0 {
		return
	}
	of.AddTableMiddleware(table.NewSortByMiddlewareFromColumns(s.SortBy...))
}
