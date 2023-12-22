package layers

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type ErrInvalidParameterLayer struct {
	Name     string
	Expected string
}

func (e ErrInvalidParameterLayer) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("invalid parameter layer: %s", e.Name)
	}
	return fmt.Sprintf("invalid parameter layer: %s (expected %s)", e.Name, e.Expected)
}

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
//
// TODO(manuel, 2023-12-20) This is a pretty messy interface, I think it used to be a struct?
type ParameterLayer interface {
	AddFlags(flag ...*parameters.ParameterDefinition)
	GetParameterDefinitions() parameters.ParameterDefinitions

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string

	Clone() ParameterLayer
}

const DefaultSlug = "default"

// ParsedParameterLayer is the result of "parsing" input data using a ParameterLayer
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query parameters.
type ParsedParameterLayer struct {
	Layer      ParameterLayer
	Parameters *parameters.ParsedParameters
}

func NewParsedParameterLayer(layer ParameterLayer) *ParsedParameterLayer {
	return &ParsedParameterLayer{
		Layer:      layer,
		Parameters: parameters.NewParsedParameters(),
	}
}

type JSONParameterLayer interface {
	ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (*parameters.ParsedParameters, error)
}

// Clone returns a copy of the parsedParameterLayer with a fresh Parameters map.
// However, neither the Layer nor the Parameters are deep copied.
func (ppl *ParsedParameterLayer) Clone() *ParsedParameterLayer {
	ret := &ParsedParameterLayer{
		Layer:      ppl.Layer,
		Parameters: parameters.NewParsedParameters().Merge(ppl.Parameters),
	}
	ppl.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
		ret.Parameters.Set(k, v)
	})
	return ret
}

// MergeParameters merges the other ParsedParameterLayer into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *ParsedParameterLayer) MergeParameters(other *ParsedParameterLayer) {
	ppl.Parameters.Merge(other.Parameters)
}

func GetAllParsedParameters(layers *ParsedParameterLayers) *parameters.ParsedParameters {
	ret := parameters.NewParsedParameters()
	layers.ForEach(
		func(_ string, v *ParsedParameterLayer) {
			ret.Merge(v.Parameters)
		})

	return ret
}

type ParsedParameterLayers struct {
	*orderedmap.OrderedMap[string, *ParsedParameterLayer]
}

func NewParsedParameterLayers() *ParsedParameterLayers {
	return &ParsedParameterLayers{
		OrderedMap: orderedmap.New[string, *ParsedParameterLayer](),
	}
}

// GetDataMap is useful when rendering out templates using all passed in layers.
// TODO(manuel, 2023-12-22) Allow passing middlewares so that we can blacklist layers
func (p *ParsedParameterLayers) GetDataMap() map[string]interface{} {
	ps := map[string]interface{}{}
	p.ForEach(func(k string, v *ParsedParameterLayer) {
		v.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
			ps[v.ParameterDefinition.Name] = v.Value
		})
	})
	return ps
}

func (p *ParsedParameterLayers) GetParameter(slug string, key string) (*parameters.ParsedParameter, bool) {
	layer, ok := p.Get(slug)
	if !ok {
		return nil, false
	}
	return layer.Parameters.Get(key)
}

func (p *ParsedParameterLayers) GetParameterValue(slug string, key string) interface{} {
	layer, ok := p.OrderedMap.Get(slug)
	if !ok {
		return nil
	}
	return layer.Parameters.GetValue(key)
}

func (p *ParsedParameterLayers) GetDefaultParameterLayer() *ParsedParameterLayer {
	v, ok := p.Get(DefaultSlug)
	if ok {
		return v
	}
	defaultParameterLayer, err := NewParameterLayer(DefaultSlug, "Default")
	if err != nil {
		panic(err)
	}
	defaultLayer := &ParsedParameterLayer{
		Layer:      defaultParameterLayer,
		Parameters: parameters.NewParsedParameters(),
	}
	p.Set(DefaultSlug, defaultLayer)

	return defaultLayer
}

func (p *ParsedParameterLayers) ForEach(fn func(k string, v *ParsedParameterLayer)) {
	for v := p.Oldest(); v != nil; v = v.Next() {
		fn(v.Key, v.Value)
	}
}

func (p *ParsedParameterLayers) ForEachE(fn func(k string, v *ParsedParameterLayer) error) error {
	for v := p.Oldest(); v != nil; v = v.Next() {
		err := fn(v.Key, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}
