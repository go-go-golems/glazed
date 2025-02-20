package layers

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	"github.com/spf13/cobra"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
//
// TODO(manuel, 2023-12-20) This is a pretty messy interface, I think it used to be a struct?
type ParameterLayer interface {
	AddFlags(flag ...*parameters.ParameterDefinition)
	GetParameterDefinitions() *parameters.ParameterDefinitions

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string

	Clone() ParameterLayer
}

const DefaultSlug = "default"

type ParameterLayers struct {
	*orderedmap.OrderedMap[string, ParameterLayer]
}

type ParameterLayersOption func(*ParameterLayers)

func WithLayers(layers ...ParameterLayer) ParameterLayersOption {
	return func(pl *ParameterLayers) {
		for _, l := range layers {
			pl.Set(l.GetSlug(), l)

		}
	}
}

func NewParameterLayers(options ...ParameterLayersOption) *ParameterLayers {
	ret := &ParameterLayers{
		OrderedMap: orderedmap.New[string, ParameterLayer](),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (pl *ParameterLayers) Subset(slugs ...string) *ParameterLayers {
	ret := NewParameterLayers()

	for _, slug := range slugs {
		if l, ok := pl.Get(slug); ok {
			ret.Set(slug, l)
		}
	}

	return ret
}

// ForEach iterates over each element in the ParameterLayers map and applies the given function to each key-value pair.
func (pl *ParameterLayers) ForEach(f func(key string, p ParameterLayer)) {
	for v := pl.Oldest(); v != nil; v = v.Next() {
		f(v.Key, v.Value)
	}
}

// ForEachE applies a function to each key-value pair in the ParameterLayers, in oldest-to-newest order.
// It stops iteration and returns the first error encountered, if any.
func (pl *ParameterLayers) ForEachE(f func(key string, p ParameterLayer) error) error {
	for v := pl.Oldest(); v != nil; v = v.Next() {
		if err := f(v.Key, v.Value); err != nil {
			return err
		}
	}
	return nil
}

func (pl *ParameterLayers) AppendLayers(layers ...ParameterLayer) {
	for _, l := range layers {
		pl.Set(l.GetSlug(), l)
	}
}

func (pl *ParameterLayers) PrependLayers(layers ...ParameterLayer) {
	list.Reverse[ParameterLayer](layers)

	for _, l := range layers {
		slug := l.GetSlug()
		pl.Set(slug, l)
		_ = pl.MoveToFront(slug)
	}
}

func (pl *ParameterLayers) Merge(p *ParameterLayers) *ParameterLayers {
	p.ForEach(func(k string, v ParameterLayer) {
		pl.Set(k, v.Clone())
	})

	return pl
}

func (pl *ParameterLayers) AsList() []ParameterLayer {
	ret := make([]ParameterLayer, 0, pl.Len())
	pl.ForEach(func(_ string, v ParameterLayer) {
		ret = append(ret, v)
	})
	return ret
}

func (pl *ParameterLayers) Clone() *ParameterLayers {
	ret := NewParameterLayers()
	return ret.Merge(pl)
}

func (pl *ParameterLayers) GetAllParameterDefinitions() *parameters.ParameterDefinitions {
	ret := parameters.NewParameterDefinitions()
	pl.ForEach(func(_ string, v ParameterLayer) {
		v.GetParameterDefinitions().ForEach(func(p *parameters.ParameterDefinition) {
			prefix := v.GetPrefix()
			ret.Set(prefix+p.Name, p)
		})
	})
	return ret
}

func (pl *ParameterLayers) AddToCobraCommand(cmd *cobra.Command) error {
	return pl.ForEachE(func(_ string, v ParameterLayer) error {
		if v.(CobraParameterLayer) != nil {
			err := v.(CobraParameterLayer).AddLayerToCobraCommand(cmd)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (pl *ParameterLayers) InitializeFromDefaults(options ...parameters.ParseStepOption) (*ParsedLayers, error) {
	ret := NewParsedLayers()
	err := pl.UpdateWithDefaults(ret, options...)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func InitializeParameterLayerWithDefaults(
	v ParameterLayer,
	parsedLayer *ParsedLayer,
	options ...parameters.ParseStepOption,
) error {
	pds := v.GetParameterDefinitions()

	err := pds.ForEachE(func(pd *parameters.ParameterDefinition) error {
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

func (pl *ParameterLayers) UpdateWithDefaults(parsedLayers *ParsedLayers, options ...parameters.ParseStepOption) error {
	err := pl.ForEachE(func(_ string, v ParameterLayer) error {
		parsedLayer := parsedLayers.GetOrCreate(v)
		return InitializeParameterLayerWithDefaults(v, parsedLayer, options...)
	})

	return err
}

// MarshalYAML implements yaml.Marshaler interface
func (pl *ParameterLayers) MarshalYAML() (interface{}, error) {
	return LayersToSerializable(pl), nil
}

// MarshalJSON implements json.Marshaler interface
func (pl *ParameterLayers) MarshalJSON() ([]byte, error) {
	return json.Marshal(LayersToSerializable(pl))
}
