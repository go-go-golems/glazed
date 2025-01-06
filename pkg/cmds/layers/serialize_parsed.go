package layers

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableParsedLayer represents a parsed layer in a format suitable for
// YAML/JSON serialization
type SerializableParsedLayer struct {
	Layer      *SerializableParameterLayer              `yaml:"layer,omitempty" json:"layer,omitempty"`
	Parameters *parameters.SerializableParsedParameters `yaml:"parameters" json:"parameters"`
}

// ToSerializableParsedLayer converts a ParsedLayer to its serializable representation
func ToSerializableParsedLayer(pl *ParsedLayer) *SerializableParsedLayer {
	return &SerializableParsedLayer{
		// Layer:      ToSerializable(pl.Layer),
		Parameters: parameters.ToSerializableParsedParameters(pl.Parameters),
	}
}

// SerializableParsedLayers represents a collection of parsed layers in a format suitable
// for YAML/JSON serialization, maintaining the order of layers
type SerializableParsedLayers struct {
	// Using orderedmap to maintain layer order while having slug-based access
	Layers *orderedmap.OrderedMap[string, *SerializableParsedLayer] `yaml:"layers" json:"layers"`
}

// ToSerializableParsedLayers converts a ParsedLayers collection to its serializable representation
func ToSerializableParsedLayers(pl *ParsedLayers) *SerializableParsedLayers {
	ret := &SerializableParsedLayers{
		Layers: orderedmap.New[string, *SerializableParsedLayer](),
	}

	pl.ForEach(func(key string, value *ParsedLayer) {
		serialized := ToSerializableParsedLayer(value)
		ret.Layers.Set(key, serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableParsedLayers
func (spl *SerializableParsedLayers) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableParsedLayer)
	for pair := spl.Layers.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableParsedLayers
func (spl *SerializableParsedLayers) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableParsedLayer)
	for pair := spl.Layers.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
