package layers

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// ParsedLayer is the result of "parsing" input data using a ParameterLayer
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query parameters.
type ParsedLayer struct {
	Layer      ParameterLayer
	Parameters *parameters.ParsedParameters
}

type ParsedLayerOption func(*ParsedLayer) error

func WithParsedParameterValue(
	key string, value interface{},
	options ...parameters.ParseStepOption,
) ParsedLayerOption {
	return func(pl *ParsedLayer) error {
		pd, ok := pl.Layer.GetParameterDefinitions().Get(key)
		if !ok {
			return errors.Errorf("parameter definition %s not found in layer %s", key, pl.Layer.GetName())
		}
		p := &parameters.ParsedParameter{
			ParameterDefinition: pd,
		}
		err := p.Update(value, options...)
		if err != nil {
			return err
		}
		pl.Parameters.Set(key, p)

		return nil
	}
}

func WithParsedParameters(pds *parameters.ParsedParameters) ParsedLayerOption {
	return func(pl *ParsedLayer) error {
		pds.ForEach(func(k string, v *parameters.ParsedParameter) {
			pl.Parameters.Set(k, v)
		})
		return nil
	}
}

func NewParsedLayer(layer ParameterLayer, options ...ParsedLayerOption) (*ParsedLayer, error) {
	ret := &ParsedLayer{
		Layer:      layer,
		Parameters: parameters.NewParsedParameters(),
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
func (ppl *ParsedLayer) Clone() *ParsedLayer {
	parameters_, err := parameters.NewParsedParameters().Merge(ppl.Parameters)
	if err != nil {
		panic(err)
	}
	ret := &ParsedLayer{
		Layer:      ppl.Layer,
		Parameters: parameters_,
	}
	ppl.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
		ret.Parameters.Set(k, v)
	})
	return ret
}

// MergeParameters merges the other ParsedLayer into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *ParsedLayer) MergeParameters(other *ParsedLayer) error {
	_, err := ppl.Parameters.Merge(other.Parameters)
	return err
}

func (ppl *ParsedLayer) GetParameter(k string) (interface{}, bool) {
	v, ok := ppl.Parameters.Get(k)
	if !ok {
		return nil, false
	}
	return v.Value, true
}

func (ppl *ParsedLayer) InitializeStruct(s interface{}) error {
	return ppl.Parameters.InitializeStruct(s)
}

type ParsedLayers struct {
	*orderedmap.OrderedMap[string, *ParsedLayer]
}

type ParsedLayersOption func(*ParsedLayers)

func WithParsedLayer(slug string, v *ParsedLayer) ParsedLayersOption {
	return func(pl *ParsedLayers) {
		pl.Set(slug, v)
	}
}

func NewParsedLayers(options ...ParsedLayersOption) *ParsedLayers {
	ret := &ParsedLayers{
		OrderedMap: orderedmap.New[string, *ParsedLayer](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (p *ParsedLayers) Clone() *ParsedLayers {
	ret := NewParsedLayers()
	p.ForEach(func(k string, v *ParsedLayer) {
		ret.Set(k, v.Clone())
	})
	return ret
}

func (p *ParsedLayers) Merge(other *ParsedLayers) error {
	err := p.ForEachE(func(k string, v *ParsedLayer) error {
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
	err = other.ForEachE(func(k string, v *ParsedLayer) error {
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

func (p *ParsedLayers) GetOrCreate(layer ParameterLayer) *ParsedLayer {
	if layer == nil {
		panic("layer must not be nil")
	}
	slug := layer.GetSlug()
	v, ok := p.Get(slug)
	if !ok {
		v = &ParsedLayer{
			Layer:      layer,
			Parameters: parameters.NewParsedParameters(),
		}
		p.Set(slug, v)
	}
	return v
}

// GetDataMap is useful when rendering out templates using all passed in layers.
func (p *ParsedLayers) GetDataMap() map[string]interface{} {
	ps := map[string]interface{}{}
	p.ForEach(func(k string, v *ParsedLayer) {
		v.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
			ps[v.ParameterDefinition.Name] = v.Value
		})
	})
	return ps
}

// InitializeStruct initializes a struct with values from a ParsedLayer specified by the key.
// If the key is "default", it creates a fresh empty default layer for defaults and initializes the struct with it.
// If the layer specified by the key is not found, it returns an error.
// The struct must be passed by reference as the s parameter.
func (p *ParsedLayers) InitializeStruct(layerKey string, dst interface{}) error {
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

// GetAllParsedParameters returns a new instance of parameters.ParsedParameters
// that merges the parameters from all ParsedLayers.
// The returned parameters are a deep clone of the parameters.
func (p *ParsedLayers) GetAllParsedParameters() *parameters.ParsedParameters {
	ret := parameters.NewParsedParameters()
	p.ForEach(
		func(_ string, v *ParsedLayer) {
			_, err := ret.Merge(v.Parameters.Clone())
			if err != nil {
				// this should never happen, we don't try to do any interesting type coercion here
				panic(err)
			}
		})

	return ret
}

func (p *ParsedLayers) GetParameter(slug string, key string) (*parameters.ParsedParameter, bool) {
	layer, ok := p.Get(slug)
	if !ok {
		return nil, false
	}
	return layer.Parameters.Get(key)
}

func (p *ParsedLayers) GetDefaultParameterLayer() *ParsedLayer {
	v, ok := p.Get(DefaultSlug)
	if ok {
		return v
	}
	defaultParameterLayer, err := NewParameterLayer(DefaultSlug, "Default")
	if err != nil {
		panic(err)
	}
	defaultLayer := &ParsedLayer{
		Layer:      defaultParameterLayer,
		Parameters: parameters.NewParsedParameters(),
	}
	p.Set(DefaultSlug, defaultLayer)

	return defaultLayer
}

func (p *ParsedLayers) ForEach(fn func(k string, v *ParsedLayer)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		fn(v.Key, v.Value)
	}
}

func (p *ParsedLayers) ForEachE(fn func(k string, v *ParsedLayer) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := fn(v.Key, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalYAML implements yaml.Marshaler for ParsedLayer
func (pl *ParsedLayer) MarshalYAML() (interface{}, error) {
	return ToSerializableParsedLayer(pl), nil
}

// MarshalJSON implements json.Marshaler for ParsedLayer
func (pl *ParsedLayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToSerializableParsedLayer(pl))
}

// MarshalYAML implements yaml.Marshaler for ParsedLayers
func (pl *ParsedLayers) MarshalYAML() (interface{}, error) {
	return ToSerializableParsedLayers(pl), nil
}

// MarshalJSON implements json.Marshaler for ParsedLayers
func (pl *ParsedLayers) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToSerializableParsedLayers(pl))
}
