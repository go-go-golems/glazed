package values

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// DefaultSlug mirrors schema.DefaultSlug to avoid a dependency cycle.
const DefaultSlug = "default"

// Section defines the minimal interface needed by values.
// It intentionally avoids schema.Section to prevent import cycles.
type Section interface {
	GetDefinitions() *fields.Definitions
	GetName() string
	GetDescription() string
	GetPrefix() string
	GetSlug() string
}

type defaultSection struct {
	name        string
	description string
	prefix      string
	slug        string
	definitions *fields.Definitions
}

func newDefaultSection(slug, name string) *defaultSection {
	return &defaultSection{
		name:        name,
		description: "",
		prefix:      "",
		slug:        slug,
		definitions: fields.NewDefinitions(),
	}
}

func (d *defaultSection) GetDefinitions() *fields.Definitions {
	return d.definitions
}

func (d *defaultSection) GetName() string {
	return d.name
}

func (d *defaultSection) GetDescription() string {
	return d.description
}

func (d *defaultSection) GetPrefix() string {
	return d.prefix
}

func (d *defaultSection) GetSlug() string {
	return d.slug
}

// SectionValues is the result of "parsing" input data using a schema.Section
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query fields.
type SectionValues struct {
	Section Section
	Fields  *fields.FieldValues
}

type SectionValuesOption func(*SectionValues) error

func WithFieldValue(
	key string, value interface{},
	options ...fields.ParseOption,
) SectionValuesOption {
	return func(pl *SectionValues) error {
		pd, ok := pl.Section.GetDefinitions().Get(key)
		if !ok {
			return errors.Errorf("field definition %s not found in section %s", key, pl.Section.GetName())
		}
		p := &fields.FieldValue{
			Definition: pd,
		}
		err := p.Update(value, options...)
		if err != nil {
			return err
		}
		pl.Fields.Set(key, p)

		return nil
	}
}

func WithFields(pds *fields.FieldValues) SectionValuesOption {
	return func(pl *SectionValues) error {
		pds.ForEach(func(k string, v *fields.FieldValue) {
			pl.Fields.Set(k, v)
		})
		return nil
	}
}

func NewSectionValues(layer Section, options ...SectionValuesOption) (*SectionValues, error) {
	ret := &SectionValues{
		Section: layer,
		Fields:  fields.NewFieldValues(),
	}

	for _, o := range options {
		err := o(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// Clone returns a copy of the SectionValues with a fresh Fields map.
// However, neither the Section nor the Fields are deep copied.
func (ppl *SectionValues) Clone() *SectionValues {
	fields_, err := fields.NewFieldValues().Merge(ppl.Fields)
	if err != nil {
		panic(err)
	}
	ret := &SectionValues{
		Section: ppl.Section,
		Fields:  fields_,
	}
	ppl.Fields.ForEach(func(k string, v *fields.FieldValue) {
		ret.Fields.Set(k, v)
	})
	return ret
}

// MergeFields merges the other SectionValues into this one, overwriting any
// existing values. This doesn't replace the actual Section pointer.
func (ppl *SectionValues) MergeFields(other *SectionValues) error {
	_, err := ppl.Fields.Merge(other.Fields)
	return err
}

func (ppl *SectionValues) GetField(k string) (interface{}, bool) {
	v, ok := ppl.Fields.Get(k)
	if !ok {
		return nil, false
	}
	return v.Value, true
}

func (ppl *SectionValues) DecodeInto(s interface{}) error {
	return ppl.Fields.DecodeInto(s)
}

type Values struct {
	*orderedmap.OrderedMap[string, *SectionValues]
}

type ValuesOption func(*Values)

func WithSectionValues(slug string, v *SectionValues) ValuesOption {
	return func(pl *Values) {
		pl.Set(slug, v)
	}
}

func New(options ...ValuesOption) *Values {
	ret := &Values{
		OrderedMap: orderedmap.New[string, *SectionValues](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (p *Values) Clone() *Values {
	ret := New()
	p.ForEach(func(k string, v *SectionValues) {
		ret.Set(k, v.Clone())
	})
	return ret
}

func (p *Values) Merge(other *Values) error {
	err := p.ForEachE(func(k string, v *SectionValues) error {
		o, ok := other.Get(k)
		if ok {
			err := v.MergeFields(o)
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = other.ForEachE(func(k string, v *SectionValues) error {
		o_, ok := p.Get(k)
		if !ok {
			p.Set(k, v)
		} else {
			err := o_.MergeFields(v)
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

func (p *Values) GetOrCreate(section Section) *SectionValues {
	if section == nil {
		panic("section must not be nil")
	}
	slug := section.GetSlug()
	v, ok := p.Get(slug)
	if !ok {
		v = &SectionValues{
			Section: section,
			Fields:  fields.NewFieldValues(),
		}
		p.Set(slug, v)
	}
	return v
}

// GetDataMap is useful when rendering out templates using all passed in sections.
func (p *Values) GetDataMap() map[string]interface{} {
	ps := map[string]interface{}{}
	p.ForEach(func(k string, v *SectionValues) {
		v.Fields.ForEach(func(k string, v *fields.FieldValue) {
			ps[v.Definition.Name] = v.Value
		})
	})
	return ps
}

// DecodeSectionInto decodes a struct with values from a SectionValues specified by the key.
// If the key is "default", it creates a fresh empty default section for defaults and decodes into the struct with it.
// If the section specified by the key is not found, it returns an error.
// The struct must be passed by reference as the s parameter.
func (p *Values) DecodeSectionInto(sectionKey string, dst interface{}) error {
	// We special case Default because we will create a fresh empty default section for defaults.
	// Not sure how necessary that is, honestly
	if sectionKey == DefaultSlug {
		return p.DefaultSectionValues().DecodeInto(dst)
	}
	v, ok := p.Get(sectionKey)
	if !ok {
		return errors.Errorf("section %s not found", sectionKey)
	}
	return v.DecodeInto(dst)
}

// AllFieldValues returns a new instance of fields.FieldValues
// that merges the fields from all Values.
// The returned fields are a deep clone of the values.
func (p *Values) AllFieldValues() *fields.FieldValues {
	ret := fields.NewFieldValues()
	p.ForEach(
		func(_ string, v *SectionValues) {
			_, err := ret.Merge(v.Fields.Clone())
			if err != nil {
				// this should never happen, we don't try to do any interesting type coercion here
				panic(err)
			}
		})

	return ret
}

func (p *Values) GetField(slug string, key string) (*fields.FieldValue, bool) {
	section, ok := p.Get(slug)
	if !ok {
		return nil, false
	}
	return section.Fields.Get(key)
}

func (p *Values) DefaultSectionValues() *SectionValues {
	v, ok := p.Get(DefaultSlug)
	if ok {
		return v
	}
	defaultSection := newDefaultSection(DefaultSlug, "Default")
	defaultValues := &SectionValues{
		Section: defaultSection,
		Fields:  fields.NewFieldValues(),
	}
	p.Set(DefaultSlug, defaultValues)

	return defaultValues
}

func (p *Values) ForEach(fn func(k string, v *SectionValues)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		fn(v.Key, v.Value)
	}
}

func (p *Values) ForEachE(fn func(k string, v *SectionValues) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := fn(v.Key, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalYAML implements yaml.Marshaler for SectionValues
func (pl *SectionValues) MarshalYAML() (interface{}, error) {
	return ToSerializableSectionValues(pl), nil
}

// MarshalJSON implements json.Marshaler for SectionValues
func (pl *SectionValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToSerializableSectionValues(pl))
}

// MarshalYAML implements yaml.Marshaler for Values
func (pl *Values) MarshalYAML() (interface{}, error) {
	return ToSerializableValues(pl), nil
}

// MarshalJSON implements json.Marshaler for Values
func (pl *Values) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToSerializableValues(pl))
}
