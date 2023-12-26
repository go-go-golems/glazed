package parameters

import (
	"github.com/wk8/go-ordered-map/v2"
)

type ParsedParameter struct {
	Value               interface{}
	ParameterDefinition *ParameterDefinition
	// Log contains a history of the parsing steps that were taken to arrive at the value.
	// Last step is the final value.
	Log []ParseStep
}

type ParseStepOption func(*ParseStep)

func WithParseStepMetadata(metadata map[string]interface{}) ParseStepOption {
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

func WithParseStepSource(source string) ParseStepOption {
	return func(p *ParseStep) {
		p.Source = source
	}
}

func WithParseStepValue(value interface{}) ParseStepOption {
	return func(p *ParseStep) {
		p.Value = value
	}
}

func NewParseStep(options ...ParseStepOption) ParseStep {
	ret := ParseStep{}
	for _, o := range options {
		o(&ret)
	}
	return ret
}

func (p *ParsedParameter) Set(value interface{}, options ...ParseStepOption) {
	p.Value = value
	step := ParseStep{
		Value:  value,
		Source: "none",
	}
	for _, o := range options {
		o(&step)
	}

	p.Log = append(p.Log, step)
}

func (p *ParsedParameter) Merge(v *ParsedParameter, options ...ParseStepOption) {
	options = append(options, WithParseStepSource("merge"), WithParseStepValue(v.Value))
	p.Log = append(p.Log, NewParseStep(options...))
	p.Log = append(p.Log, v.Log...)
	p.Value = v.Value
}

func (p *ParsedParameter) Clone() *ParsedParameter {
	ret := &ParsedParameter{
		Value:               p.Value,
		ParameterDefinition: p.ParameterDefinition,
		Log:                 make([]ParseStep, len(p.Log)),
	}
	copy(ret.Log, p.Log)
	return ret
}

type ParsedParameters struct {
	*orderedmap.OrderedMap[string, *ParsedParameter]
}

type ParsedParametersOption func(*ParsedParameters)

func WithParsedParameter(pd *ParameterDefinition, key string, value interface{}) ParsedParametersOption {
	return func(p *ParsedParameters) {
		p.Set(key, &ParsedParameter{
			ParameterDefinition: pd,
			Value:               value,
		})
	}
}

func NewParsedParameters(options ...ParsedParametersOption) *ParsedParameters {
	ret := &ParsedParameters{
		OrderedMap: orderedmap.New[string, *ParsedParameter](),
	}
	for _, o := range options {
		o(ret)
	}
	return ret
}

func (p *ParsedParameters) GetValue(key string) interface{} {
	v, ok := p.Get(key)
	if !ok {
		return nil
	}
	return v.Value
}

func (p *ParsedParameters) Clone() *ParsedParameters {
	ret := NewParsedParameters()
	p.ForEach(func(k string, v *ParsedParameter) {
		ret.Set(k, v.Clone())
	})
	return ret
}

// UpdateExistingValue updates the value of an existing parameter, and returns true if the parameter existed.
// If the parameter did not exist, it returns false.
func (p *ParsedParameters) UpdateExistingValue(
	key string, v interface{},
	options ...ParseStepOption,
) bool {
	v_, ok := p.Get(key)
	if !ok {
		return false
	}
	v_.Set(v, options...)
	return true
}

func (p *ParsedParameters) UpdateValue(
	key string, pd *ParameterDefinition,
	v interface{},
	options ...ParseStepOption,
) {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
	}
	v_.Set(v, options...)
}

// SetAsDefault sets the current value of the parameter if no value has yet been set.
func (p *ParsedParameters) SetAsDefault(
	key string, pd *ParameterDefinition, v interface{},
	options ...ParseStepOption) {
	if _, ok := p.Get(key); !ok {
		v_ := &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
		v_.Set(v, options...)
	}
}

func (p *ParsedParameters) ForEach(f func(key string, value *ParsedParameter)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		f(v.Key, v.Value)
	}
}

func (p *ParsedParameters) ForEachE(f func(key string, value *ParsedParameter) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := f(v.Key, v.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Merge is actually more complex than it seems, other takes precedence. If the key already exists in the map,
// we actually merge the ParsedParameter themselves, by appending the entire history of the other parameter to the
// current one.
func (p *ParsedParameters) Merge(other *ParsedParameters, options ...ParseStepOption) *ParsedParameters {
	other.ForEach(func(k string, v *ParsedParameter) {
		v_, ok := p.Get(k)
		if ok {
			v_.Merge(v, options...)
		} else {
			p.Set(k, v)
		}
	})
	return p
}

func (p *ParsedParameters) ToMap() map[string]interface{} {
	ret := map[string]interface{}{}
	p.ForEach(func(k string, v *ParsedParameter) {
		ret[k] = v.Value
	})
	return ret
}
