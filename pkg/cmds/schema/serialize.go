package schema

import (
	"encoding/json"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableSection represents a schema section in a format suitable for
// YAML/JSON serialization.
type SerializableSection struct {
	Name        string              `yaml:"name" json:"name"`
	Slug        string              `yaml:"slug" json:"slug"`
	Description string              `yaml:"description" json:"description"`
	Prefix      string              `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Fields      *fields.Definitions `yaml:"fields" json:"fields"`
}

// ToSerializable converts a Section to its serializable representation.
func ToSerializable(section Section) *SerializableSection {
	return &SerializableSection{
		Name:        section.GetName(),
		Slug:        section.GetSlug(),
		Description: section.GetDescription(),
		Prefix:      section.GetPrefix(),
		Fields:      section.GetDefinitions(),
	}
}

// SerializableSchema represents a collection of sections in a format suitable
// for YAML/JSON serialization, maintaining the order of sections.
type SerializableSchema struct {
	// Using orderedmap to maintain section order while having slug-based access.
	Sections *orderedmap.OrderedMap[string, *SerializableSection] `yaml:"sections" json:"sections"`
}

// SchemaToSerializable converts a Schema collection to its serializable representation.
func SchemaToSerializable(schema *Schema) *SerializableSchema {
	ret := &SerializableSchema{
		Sections: orderedmap.New[string, *SerializableSection](),
	}

	schema.ForEach(func(_ string, section Section) {
		serialized := ToSerializable(section)
		ret.Sections.Set(section.GetSlug(), serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableSchema.
func (sl *SerializableSchema) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableSection)
	for pair := sl.Sections.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableSchema.
func (sl *SerializableSchema) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableSection)
	for pair := sl.Sections.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
