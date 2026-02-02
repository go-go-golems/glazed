package schema

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableParameterLayer represents a parameter layer in a format suitable for
// YAML/JSON serialization
type SerializableParameterLayer struct {
	Name        string              `yaml:"name" json:"name"`
	Slug        string              `yaml:"slug" json:"slug"`
	Description string              `yaml:"description" json:"description"`
	Prefix      string              `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Parameters  *fields.Definitions `yaml:"parameters" json:"parameters"`
}

// ToSerializable converts a Section to its serializable representation
func ToSerializable(layer Section) *SerializableParameterLayer {
	return &SerializableParameterLayer{
		Name:        layer.GetName(),
		Slug:        layer.GetSlug(),
		Description: layer.GetDescription(),
		Prefix:      layer.GetPrefix(),
		Parameters:  layer.GetDefinitions(),
	}
}

// SerializableLayers represents a collection of parameter layers in a format suitable
// for YAML/JSON serialization, maintaining the order of layers
type SerializableLayers struct {
	// Using orderedmap to maintain layer order while having slug-based access
	Layers *orderedmap.OrderedMap[string, *SerializableParameterLayer] `yaml:"layers" json:"layers"`
}

// LayersToSerializable converts a Schema collection to its serializable representation
func LayersToSerializable(layers *Schema) *SerializableLayers {
	ret := &SerializableLayers{
		Layers: orderedmap.New[string, *SerializableParameterLayer](),
	}

	layers.ForEach(func(_ string, layer Section) {
		serialized := ToSerializable(layer)
		ret.Layers.Set(layer.GetSlug(), serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableLayers
func (sl *SerializableLayers) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableParameterLayer)
	for pair := sl.Layers.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableLayers
func (sl *SerializableLayers) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableParameterLayer)
	for pair := sl.Layers.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
