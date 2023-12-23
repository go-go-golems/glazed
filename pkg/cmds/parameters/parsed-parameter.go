package parameters

import (
	reflect2 "github.com/go-go-golems/glazed/pkg/helpers/reflect"
	"github.com/pkg/errors"
	"github.com/wk8/go-ordered-map/v2"
	"reflect"
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

func (p *ParsedParameter) SetWithSource(source string, value interface{}, options ...ParseStepOption) {
	p.Value = value
	step := ParseStep{
		Source: source,
		Value:  value,
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
func (p *ParsedParameters) UpdateExistingValue(
	key string, source string, v interface{},
	options ...ParseStepOption,
) bool {
	v_, ok := p.Get(key)
	if !ok {
		return false
	}
	v_.SetWithSource(source, v, options...)
	return true
}

func (p *ParsedParameters) UpdateValue(
	key string, pd *ParameterDefinition,
	source string, v interface{},
	options ...ParseStepOption,
) {
	v_, ok := p.Get(key)
	if !ok {
		v_ = &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
	}
	v_.SetWithSource(source, v, options...)
}

// SetAsDefault sets the current value of the parameter if no value has yet been set.
func (p *ParsedParameters) SetAsDefault(
	key string, pd *ParameterDefinition, source string, v interface{},
	options ...ParseStepOption) {
	if _, ok := p.Get(key); !ok {
		v_ := &ParsedParameter{
			ParameterDefinition: pd,
		}
		p.Set(key, v_)
		v_.SetWithSource(source, v, options...)
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

// InitializeStruct initializes a struct from a ParsedParameters map.
//
// It iterates through the struct fields looking for those tagged with
// "glazed.parameter". For each tagged field, it will lookup the corresponding
// parameter value in the ParsedParameters map and set the field's value.
//
// Struct fields that are pointers to other structs are handled recursively.
//
// s should be a pointer to the struct to initialize.
//
// ps is the ParsedParameters map to lookup parameter values from.
//
// Returns an error if:
// - s is not a pointer to a struct
// - A tagged field does not have a matching parameter value in ps
// - Failed to set the value of a field
func (p *ParsedParameters) InitializeStruct(s interface{}) error {
	if s == nil {
		return errors.Errorf("Can't initialize nil struct")
	}
	// check that s is indeed a pointer to a struct
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return errors.Errorf("s is not a pointer")
	}
	if reflect.TypeOf(s).Elem().Kind() != reflect.Struct {
		return errors.Errorf("s is not a pointer to a struct")
	}
	st := reflect.TypeOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		v, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		v_, ok := p.Get(v)
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		if field.Type.Kind() == reflect.Ptr {
			elem := field.Type.Elem()
			if value.IsNil() {
				value.Set(reflect.New(elem))
			}
			//exhaustive:ignore
			switch elem.Kind() {
			case reflect.Struct:
				err := p.InitializeStruct(value.Interface())
				if err != nil {
					return errors.Wrapf(err, "failed to initialize struct for %s", v)
				}
			default:
				err := reflect2.SetReflectValue(value.Elem(), v_.Value)
				if err != nil {
					return errors.Wrapf(err, "failed to set value for %s", v)
				}
			}

		} else {
			err := reflect2.SetReflectValue(value, v_.Value)
			if err != nil {
				return errors.Wrapf(err, "failed to set value for %s", v)
			}
		}
	}

	return nil
}
