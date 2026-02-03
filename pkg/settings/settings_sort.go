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
	SortBy []string `glazed:"sort-by"`
}

func NewSortSettingsFromValues(glazedValues *values.SectionValues) (*SortFlagsSettings, error) {
	s := &SortFlagsSettings{}
	err := glazedValues.Fields.DecodeInto(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize sort settings from fields")
	}

	return s, nil
}

type SortSection struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewSortSection(options ...schema.SectionOption) (*SortSection, error) {
	ret := &SortSection{}
	section, err := schema.NewSectionFromYAML(sortFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create sort field section")
	}
	ret.SectionImpl = section

	return ret, nil
}

func (f *SortSection) Clone() schema.Section {
	return &SortSection{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func (s *SortFlagsSettings) AddMiddlewares(p_ *middlewares.TableProcessor) {
	if len(s.SortBy) == 0 {
		return
	}
	p_.AddTableMiddleware(table.NewSortByMiddlewareFromColumns(s.SortBy...))
}
