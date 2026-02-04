package fields

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type FieldValue struct {
	Value      interface{}
	Definition *Definition
	// Log contains a history of the parsing steps that were taken to arrive at the value.
	// Last step is the final value.
	Log []ParseStep
}

type ParseOption func(*ParseStep)

func WithMetadata(metadata map[string]interface{}) ParseOption {
	return func(p *ParseStep) {
		if p.Metadata == nil {
			p.Metadata = metadata
			return
		}

		for k, v := range metadata {
			p.Metadata[k] = v
		}
	}
}

func WithSource(source string) ParseOption {
	return func(p *ParseStep) {
		p.Source = source
	}
}

func WithParseStepValue(value interface{}) ParseOption {
	return func(p *ParseStep) {
		p.Value = value
	}
}

func NewParseStep(options ...ParseOption) ParseStep {
	ret := ParseStep{}
	for _, o := range options {
		o(&ret)
	}
	return ret
}

// Update sets the value of the fieldValue, and appends a new parseStep.
func (p *FieldValue) Update(value interface{}, options ...ParseOption) error {
	v, err := p.Definition.CheckValueValidity(value)
	if err != nil {
		return err
	}
	p.Value = v
	step := ParseStep{
		Value:  value,
		Source: "none",
	}
	for _, o := range options {
		o(&step)
	}

	p.Log = append(p.Log, step)
	return nil
}

func (p *FieldValue) RenderValue() (string, error) {
	return RenderValue(p.Definition.Type, p.Value)
}

// UpdateWithLog sets the value of the fieldValue, and appends the given log.
func (p *FieldValue) UpdateWithLog(value interface{}, log ...ParseStep) error {
	v, err := p.Definition.CheckValueValidity(value)
	if err != nil {
		return err
	}
	p.Value = v

	p.Log = append(p.Log, log...)
	return nil
}

// Set sets the value of the fieldValue, and manually updates the log
func (p *FieldValue) Set(value interface{}, log ...ParseStep) {
	v, err := p.Definition.CheckValueValidity(value)
	if err != nil {
		// XXX add proper error return here
		panic(err)
	}
	p.Value = v
	p.Log = log
}

func (p *FieldValue) Merge(v *FieldValue, options ...ParseOption) {
	options = append(options, WithSource("merge"), WithParseStepValue(v.Value))
	p.Log = append(p.Log, NewParseStep(options...))
	p.Log = append(p.Log, v.Log...)
	p.Value = v.Value
}

func (p *FieldValue) Clone() *FieldValue {
	ret := &FieldValue{
		Value:      p.Value,
		Definition: p.Definition,
		Log:        make([]ParseStep, len(p.Log)),
	}
	copy(ret.Log, p.Log)
	return ret
}

// GetInterfaceValue returns the value as an interface{}. If the type of the field is a list,
// it will return a []interface{}. If the type is an object, it will return a map[string]interface{}.
// If the type is a list of objects, it will return a []interface{} of map[string]interface{}.
func (p *FieldValue) GetInterfaceValue() (interface{}, error) {
	fieldType := p.Definition.Type
	switch {
	case fieldType.IsList():
		ret, err := cast.CastListToInterfaceList(p.Value)
		if err != nil {
			return nil, err
		}
		return ret, nil

	case fieldType.IsObject(),
		fieldType.IsKeyValue():
		return cast.ConvertMapToInterfaceMap(p.Value)

	case fieldType.IsObjectList():
		r_, err := cast.CastListToInterfaceList(p.Value)
		if err != nil {
			return nil, err
		}

		ret := []interface{}{}
		for _, m := range r_ {
			m_, err := cast.ConvertMapToInterfaceMap(m)
			if err != nil {
				return nil, err
			}
			ret = append(ret, m_)
		}
		return ret, nil

	default:
		return p.Value, nil
	}
}

type FieldValues struct {
	*orderedmap.OrderedMap[string, *FieldValue]
}

type FieldValuesOption func(*FieldValues)

func WithFieldValue(pd *Definition, key string, value interface{}) FieldValuesOption {
	return func(p *FieldValues) {
		p.Set(key, &FieldValue{
			Definition: pd,
			Value:      value,
		})
	}
}

func NewFieldValues(options ...FieldValuesOption) *FieldValues {
	ret := &FieldValues{
		OrderedMap: orderedmap.New[string, *FieldValue](),
	}
	for _, o := range options {
		o(ret)
	}
	return ret
}

func (p *FieldValues) GetValue(key string) interface{} {
	v, ok := p.Get(key)
	if !ok {
		return nil
	}
	return v.Value
}

func (p *FieldValues) Clone() *FieldValues {
	ret := NewFieldValues()
	p.ForEach(func(k string, v *FieldValue) {
		ret.Set(k, v.Clone())
	})
	return ret
}

// UpdateExistingValue updates the value of an existing field, and returns true if the field existed.
// If the field did not exist, it returns false.
func (p *FieldValues) UpdateExistingValue(
	key string, v interface{},
	options ...ParseOption,
) (bool, error) {
	v_, ok := p.Get(key)
	if !ok {
		return false, nil
	}
	err := v_.Update(v, options...)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *FieldValues) Update(
	key string, pp *FieldValue,
) {
	v_, ok := p.Get(key)
	if !ok {
		p.Set(key, pp)
	} else {
		v_.Merge(pp)
	}
}

// XXX Add proper error return handling here
func (p *FieldValues) UpdateValue(
	key string,
	pd *Definition,
	v interface{},
	options ...ParseOption,
) error {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &FieldValue{
			Definition: pd,
		}
		p.Set(key, v_)
	}
	err := v_.Update(v, options...)
	if err != nil {
		return err
	}
	return nil
}

func (p *FieldValues) MustUpdateValue(
	key string,
	v interface{},
	options ...ParseOption,
) error {
	v_, ok := p.Get(key)
	if !ok {
		return errors.Errorf("field %s not found", key)
	}
	err := v_.Update(v, options...)
	if err != nil {
		return err
	}
	return nil
}

func (p *FieldValues) UpdateWithLog(
	key string, pd *Definition,
	v interface{},
	log ...ParseStep,
) error {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &FieldValue{
			Definition: pd,
		}
		p.Set(key, v_)
	}
	err := v_.UpdateWithLog(v, log...)
	if err != nil {
		return err
	}
	return nil
}

// SetAsDefault sets the current value of the field if no value has yet been set.
func (p *FieldValues) SetAsDefault(
	key string, pd *Definition, v interface{},
	options ...ParseOption,
) error {
	if _, ok := p.Get(key); !ok {
		err := p.UpdateValue(key, pd, v, options...)
		if err != nil {
			return err
		}
	}
	return nil
}

// ForEach applies the passed function to each key-value pair from oldest to newest
// in FieldValues.
func (p *FieldValues) ForEach(f func(key string, value *FieldValue)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		f(v.Key, v.Value)
	}
}

// ForEachE applies the passed function (that returns an error) to each pair in
// FieldValues. It stops at, and returns, the first error encountered.
func (p *FieldValues) ForEachE(f func(key string, value *FieldValue) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := f(v.Key, v.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Merge is actually more complex than it seems, other takes precedence. If the key already exists in the map,
// we actually merge the FieldValue themselves, by appending the entire history of the other field to the
// current one.
func (p *FieldValues) Merge(other *FieldValues) (*FieldValues, error) {
	err := other.ForEachE(func(k string, v *FieldValue) error {
		err := p.UpdateWithLog(k, v.Definition, v.Value, v.Log...)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// MergeAsDefault only sets the value if the key does not already exist in the map.
func (p *FieldValues) MergeAsDefault(other *FieldValues, options ...ParseOption) (*FieldValues, error) {
	err := other.ForEachE(func(k string, v *FieldValue) error {
		err := p.SetAsDefault(k, v.Definition, v.Value, options...)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ToMap converts FieldValues to map[string]interface{} by assigning each FieldValue's value to its key.
func (p *FieldValues) ToMap() map[string]interface{} {
	ret := map[string]interface{}{}
	p.ForEach(func(k string, v *FieldValue) {
		ret[k] = v.Value
	})
	return ret
}

// ToInterfaceMap converts FieldValues to map[string]interface{} by converting each FieldValue's value to interface{}.
// It returns an error if it fails to convert any FieldValue's value.
func (p *FieldValues) ToInterfaceMap() (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	err := p.ForEachE(func(k string, v *FieldValue) error {
		var err error
		ret[k], err = v.GetInterfaceValue()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MarshalYAML implements yaml.Marshaler for FieldValues
func (p *FieldValues) MarshalYAML() (interface{}, error) {
	return ToSerializableFieldValues(p), nil
}

// MarshalJSON implements json.Marshaler for FieldValues
func (p *FieldValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToSerializableFieldValues(p))
}
