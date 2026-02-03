package schema

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	"github.com/spf13/cobra"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Section is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
//
// TODO(manuel, 2023-12-20) This is a pretty messy interface, I think it used to be a struct?
type Section interface {
	AddFields(flag ...*fields.Definition)
	GetDefinitions() *fields.Definitions

	InitializeDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string

	Clone() Section
}

const DefaultSlug = "default"

type Schema struct {
	*orderedmap.OrderedMap[string, Section]
}

type SchemaOption func(*Schema)

func WithSections(layers ...Section) SchemaOption {
	return func(pl *Schema) {
		for _, l := range layers {
			pl.Set(l.GetSlug(), l)

		}
	}
}

func NewSchema(options ...SchemaOption) *Schema {
	ret := &Schema{
		OrderedMap: orderedmap.New[string, Section](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (pl *Schema) Subset(slugs ...string) *Schema {
	ret := NewSchema()

	for _, slug := range slugs {
		if l, ok := pl.Get(slug); ok {
			ret.Set(slug, l)
		}
	}

	return ret
}

// ForEach iterates over each element in the Schema map and applies the given function to each key-value pair.
func (pl *Schema) ForEach(f func(key string, p Section)) {
	for v := pl.Oldest(); v != nil; v = v.Next() {
		f(v.Key, v.Value)
	}
}

// ForEachE applies a function to each key-value pair in the Schema, in oldest-to-newest order.
// It stops iteration and returns the first error encountered, if any.
func (pl *Schema) ForEachE(f func(key string, p Section) error) error {
	for v := pl.Oldest(); v != nil; v = v.Next() {
		if err := f(v.Key, v.Value); err != nil {
			return err
		}
	}
	return nil
}

func (pl *Schema) AppendLayers(layers ...Section) {
	for _, l := range layers {
		pl.Set(l.GetSlug(), l)
	}
}

func (pl *Schema) PrependLayers(layers ...Section) {
	list.Reverse[Section](layers)

	for _, l := range layers {
		slug := l.GetSlug()
		pl.Set(slug, l)
		_ = pl.MoveToFront(slug)
	}
}

func (pl *Schema) Merge(p *Schema) *Schema {
	p.ForEach(func(k string, v Section) {
		pl.Set(k, v.Clone())
	})

	return pl
}

func (pl *Schema) AsList() []Section {
	ret := make([]Section, 0, pl.Len())
	pl.ForEach(func(_ string, v Section) {
		ret = append(ret, v)
	})
	return ret
}

func (pl *Schema) Clone() *Schema {
	ret := NewSchema()
	return ret.Merge(pl)
}

func (pl *Schema) GetAllDefinitions() *fields.Definitions {
	ret := fields.NewDefinitions()
	pl.ForEach(func(_ string, v Section) {
		v.GetDefinitions().ForEach(func(p *fields.Definition) {
			prefix := v.GetPrefix()
			ret.Set(prefix+p.Name, p)
		})
	})
	return ret
}

func (pl *Schema) AddToCobraCommand(cmd *cobra.Command) error {
	return pl.ForEachE(func(_ string, v Section) error {
		if v.(CobraSection) != nil {
			err := v.(CobraSection).AddSectionToCobraCommand(cmd)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (pl *Schema) InitializeFromDefaults(options ...fields.ParseOption) (*values.Values, error) {
	ret := values.New()
	err := pl.UpdateWithDefaults(ret, options...)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func InitializeSectionWithDefaults(
	v Section,
	parsedLayer *values.SectionValues,
	options ...fields.ParseOption,
) error {
	pds := v.GetDefinitions()

	err := pds.ForEachE(func(pd *fields.Definition) error {
		v, err := pd.CheckParameterDefaultValueValidity()
		if err != nil {
			return err
		}
		if v != nil {
			err := parsedLayer.Parameters.SetAsDefault(pd.Name, pd, v, options...)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (pl *Schema) UpdateWithDefaults(parsedLayers *values.Values, options ...fields.ParseOption) error {
	err := pl.ForEachE(func(_ string, v Section) error {
		parsedLayer := parsedLayers.GetOrCreate(v)
		return InitializeSectionWithDefaults(v, parsedLayer, options...)
	})

	return err
}

// MarshalYAML implements yaml.Marshaler interface
func (pl *Schema) MarshalYAML() (interface{}, error) {
	return LayersToSerializable(pl), nil
}

// MarshalJSON implements json.Marshaler interface
func (pl *Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(LayersToSerializable(pl))
}
