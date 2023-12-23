package parameters

import "github.com/wk8/go-ordered-map/v2"

type ParsedParameter struct {
	Value               interface{}
	ParameterDefinition *ParameterDefinition
	// Log contains a history of the parsing steps that were taken to arrive at the value.
	// Last step is the final value.
	Log []ParseStep
}

func (p *ParsedParameter) Set(source string, value interface{}) {
	p.Value = value
	p.Log = append(p.Log, ParseStep{
		Source: source,
		Value:  value,
	})
}

func (p *ParsedParameter) SetWithMetadata(source string, value interface{}, metadata map[string]interface{}) {
	p.Value = value
	p.Log = append(p.Log, ParseStep{
		Source:   source,
		Value:    value,
		Metadata: metadata,
	})
}

func (p *ParsedParameter) Merge(v *ParsedParameter) {
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

func (p *ParsedParameters) GetCheckedValue(key string) (interface{}, bool) {
	v, ok := p.Get(key)
	if !ok {
		return nil, false
	}
	return v.Value, true
}

func (p *ParsedParameters) GetValue(key string) interface{} {
	v, ok := p.Get(key)
	if !ok {
		return nil
	}
	return v.Value
}

// UpdateExistingValue updates the value of an existing parameter, and returns true if the parameter existed.
// If the parameter did not exist, it returns false.
func (p *ParsedParameters) UpdateExistingValue(key string, source string, v interface{}) bool {
	v_, ok := p.Get(key)
	if !ok {
		return false
	}
	v_.Set(source, v)
	return true
}

func (p *ParsedParameters) UpdateExistingValueWithMetadata(
	key string,
	source string,
	v interface{},
	metadata map[string]interface{},
) bool {
	v_, ok := p.Get(key)
	if !ok {
		return false
	}
	v_.SetWithMetadata(source, v, metadata)
	return true
}

func (p *ParsedParameters) UpdateValue(key string, pd *ParameterDefinition, source string, v interface{}) {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
	}
	v_.Set(source, v)
}

func (p *ParsedParameters) UpdateValueWithMetadata(
	key string,
	pd *ParameterDefinition,
	source string,
	v interface{},
	metadata map[string]interface{},
) {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
	}
	v_.SetWithMetadata(source, v, metadata)
}

// SetAsDefault sets the current value of the parameter if no value has yet been set.
func (p *ParsedParameters) SetAsDefault(key string, pd *ParameterDefinition, source string, v interface{}) {
	if _, ok := p.Get(key); !ok {
		v_ := &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
		v_.Set(source, v)
	}
}

func (p *ParsedParameters) SetAsDefaultWithMetadata(
	key string,
	pd *ParameterDefinition,
	source string,
	v interface{},
	metadata map[string]interface{},
) {
	if _, ok := p.Get(key); !ok {
		v_ := &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
		v_.SetWithMetadata(source, v, metadata)
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
func (p *ParsedParameters) Merge(other *ParsedParameters) *ParsedParameters {
	other.ForEach(func(k string, v *ParsedParameter) {
		v_, ok := p.Get(k)
		if ok {
			v_.Merge(v)
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

func (p *ParsedParameters) InitializeStruct(s interface{}) error {
	return InitializeStructFromParameters(s, p)
}
