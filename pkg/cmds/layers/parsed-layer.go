package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/wk8/go-ordered-map/v2"
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
	source string,
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
		p.SetWithSource(source, value, options...)
		pl.Parameters.Set(key, p)

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
	ret := &ParsedLayer{
		Layer:      ppl.Layer,
		Parameters: parameters.NewParsedParameters().Merge(ppl.Parameters),
	}
	ppl.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
		ret.Parameters.Set(k, v)
	})
	return ret
}

// MergeParameters merges the other ParsedLayer into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *ParsedLayer) MergeParameters(other *ParsedLayer) {
	ppl.Parameters.Merge(other.Parameters)
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

// GetDataMap is useful when rendering out templates using all passed in layers.
// TODO(manuel, 2023-12-22) Allow passing middlewares so that we can blacklist layers
func (p *ParsedLayers) GetDataMap() map[string]interface{} {
	ps := map[string]interface{}{}
	p.ForEach(func(k string, v *ParsedLayer) {
		v.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
			ps[v.ParameterDefinition.Name] = v.Value
		})
	})
	return ps
}

func GetAllParsedParameters(layers *ParsedLayers) *parameters.ParsedParameters {
	ret := parameters.NewParsedParameters()
	layers.ForEach(
		func(_ string, v *ParsedLayer) {
			ret.Merge(v.Parameters)
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
