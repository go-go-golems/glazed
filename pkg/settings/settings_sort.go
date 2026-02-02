package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
)

//go:embed "flags/sort.yaml"
var sortFlagsYaml []byte

type SortFlagsSettings struct {
	SortBy []string `glazed.parameter:"sort-by"`
}

func NewSortSettingsFromParameters(glazedLayer *values.SectionValues) (*SortFlagsSettings, error) {
	s := &SortFlagsSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize sort settings from parameters")
	}

	return s, nil
}

type SortParameterLayer struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewSortParameterLayer(options ...schema.SectionOption) (*SortParameterLayer, error) {
	ret := &SortParameterLayer{}
	layer, err := schema.NewSectionFromYAML(sortFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create sort parameter layer")
	}
	ret.SectionImpl = layer

	return ret, nil
}

func (f *SortParameterLayer) Clone() schema.Section {
	return &SortParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func (s *SortFlagsSettings) AddMiddlewares(p_ *middlewares.TableProcessor) {
	if len(s.SortBy) == 0 {
		return
	}
	p_.AddTableMiddleware(table.NewSortByMiddlewareFromColumns(s.SortBy...))
}
