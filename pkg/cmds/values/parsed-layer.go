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
	Layer      Section
	Parameters *fields.ParsedParameters
}

type SectionValuesOption func(*SectionValues) error

func WithParameterValue(
	key string, value interface{},
	options ...fields.ParseOption,
) SectionValuesOption {
	return func(pl *SectionValues) error {
		pd, ok := pl.Layer.GetDefinitions().Get(key)
		if !ok {
			return errors.Errorf("parameter definition %s not found in layer %s", key, pl.Layer.GetName())
		}
		p := &fields.ParsedParameter{
			Definition: pd,
		}
		err := p.Update(value, options...)
		if err != nil {
			return err
		}
		pl.Parameters.Set(key, p)

		return nil
	}
}

func WithParameters(pds *fields.ParsedParameters) SectionValuesOption {
	return func(pl *SectionValues) error {
		pds.ForEach(func(k string, v *fields.ParsedParameter) {
			pl.Parameters.Set(k, v)
		})
		return nil
	}
}

func NewSectionValues(layer Section, options ...SectionValuesOption) (*SectionValues, error) {
	ret := &SectionValues{
		Layer:      layer,
		Parameters: fields.NewParsedParameters(),
	}

	for _, o := range options {
		err := o(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// Clone returns a copy of the parsedParameterLayer with a fresh Parameters map.
// However, neither the Layer nor the Parameters are deep copied.
func (ppl *SectionValues) Clone() *SectionValues {
	parameters_, err := fields.NewParsedParameters().Merge(ppl.Parameters)
	if err != nil {
		panic(err)
	}
	ret := &SectionValues{
		Layer:      ppl.Layer,
		Parameters: parameters_,
	}
	ppl.Parameters.ForEach(func(k string, v *fields.ParsedParameter) {
		ret.Parameters.Set(k, v)
	})
	return ret
}

// MergeParameters merges the other SectionValues into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *SectionValues) MergeParameters(other *SectionValues) error {
	_, err := ppl.Parameters.Merge(other.Parameters)
	return err
}

func (ppl *SectionValues) GetParameter(k string) (interface{}, bool) {
	v, ok := ppl.Parameters.Get(k)
	if !ok {
		return nil, false
	}
	return v.Value, true
}

func (ppl *SectionValues) InitializeStruct(s interface{}) error {
	return ppl.Parameters.InitializeStruct(s)
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
			err := v.MergeParameters(o)
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
			err := o_.MergeParameters(v)
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

func (p *Values) GetOrCreate(layer Section) *SectionValues {
	if layer == nil {
		panic("layer must not be nil")
	}
	slug := layer.GetSlug()
	v, ok := p.Get(slug)
	if !ok {
		v = &SectionValues{
			Layer:      layer,
			Parameters: fields.NewParsedParameters(),
		}
		p.Set(slug, v)
	}
	return v
}

// GetDataMap is useful when rendering out templates using all passed in layers.
func (p *Values) GetDataMap() map[string]interface{} {
	ps := map[string]interface{}{}
	p.ForEach(func(k string, v *SectionValues) {
		v.Parameters.ForEach(func(k string, v *fields.ParsedParameter) {
			ps[v.Definition.Name] = v.Value
		})
	})
	return ps
}

// InitializeStruct initializes a struct with values from a SectionValues specified by the key.
// If the key is "default", it creates a fresh empty default layer for defaults and initializes the struct with it.
// If the layer specified by the key is not found, it returns an error.
// The struct must be passed by reference as the s parameter.
func (p *Values) InitializeStruct(layerKey string, dst interface{}) error {
	// We special case Default because we will create a fresh empty default layer for defaults.
	// Not sure how necessary that is, honestly
	if layerKey == DefaultSlug {
		return p.GetDefaultParameterLayer().InitializeStruct(dst)
	}
	v, ok := p.Get(layerKey)
	if !ok {
		return errors.Errorf("layer %s not found", layerKey)
	}
	return v.InitializeStruct(dst)
}

// GetAllParsedParameters returns a new instance of fields.ParsedParameters
// that merges the parameters from all Values.
// The returned parameters are a deep clone of the fields.
func (p *Values) GetAllParsedParameters() *fields.ParsedParameters {
	ret := fields.NewParsedParameters()
	p.ForEach(
		func(_ string, v *SectionValues) {
			_, err := ret.Merge(v.Parameters.Clone())
			if err != nil {
				// this should never happen, we don't try to do any interesting type coercion here
				panic(err)
			}
		})

	return ret
}

func (p *Values) GetParameter(slug string, key string) (*fields.ParsedParameter, bool) {
	layer, ok := p.Get(slug)
	if !ok {
		return nil, false
	}
	return layer.Parameters.Get(key)
}

func (p *Values) GetDefaultParameterLayer() *SectionValues {
	v, ok := p.Get(DefaultSlug)
	if ok {
		return v
	}
	defaultParameterLayer := newDefaultSection(DefaultSlug, "Default")
	defaultLayer := &SectionValues{
		Layer:      defaultParameterLayer,
		Parameters: fields.NewParsedParameters(),
	}
	p.Set(DefaultSlug, defaultLayer)

	return defaultLayer
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
