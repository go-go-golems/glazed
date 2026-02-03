package values

import (
	"encoding/json"

	fields "github.com/go-go-golems/glazed/pkg/cmds/fields"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableSection represents a schema section in a format suitable for
// YAML/JSON serialization.
type SerializableSection struct {
	Name        string              `yaml:"name" json:"name"`
	Slug        string              `yaml:"slug" json:"slug"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	Prefix      string              `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Fields      *fields.Definitions `yaml:"fields" json:"fields"`
}

// SerializableSectionValues represents a parsed section in a format suitable for
// YAML/JSON serialization.
type SerializableSectionValues struct {
	Section *SerializableSection            `yaml:"section,omitempty" json:"section,omitempty"`
	Fields  *fields.SerializableFieldValues `yaml:"fields" json:"fields"`
}

// ToSerializableSectionValues converts a SectionValues to its serializable representation.
func ToSerializableSectionValues(pl *SectionValues) *SerializableSectionValues {
	var section *SerializableSection
	if pl.Section != nil {
		section = &SerializableSection{
			Name:        pl.Section.GetName(),
			Slug:        pl.Section.GetSlug(),
			Description: pl.Section.GetDescription(),
			Prefix:      pl.Section.GetPrefix(),
			Fields:      pl.Section.GetDefinitions(),
		}
	}
	return &SerializableSectionValues{
		Section: section,
		Fields:  fields.ToSerializableFieldValues(pl.Fields),
	}
}

// SerializableValues represents a collection of parsed sections in a format suitable
// for YAML/JSON serialization, maintaining the order of sections.
type SerializableValues struct {
	// Using orderedmap to maintain section order while having slug-based access.
	Sections *orderedmap.OrderedMap[string, *SerializableSectionValues] `yaml:"sections" json:"sections"`
}

// ToSerializableValues converts a Values collection to its serializable representation.
func ToSerializableValues(pl *Values) *SerializableValues {
	ret := &SerializableValues{
		Sections: orderedmap.New[string, *SerializableSectionValues](),
	}

	pl.ForEach(func(key string, value *SectionValues) {
		serialized := ToSerializableSectionValues(value)
		ret.Sections.Set(key, serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableValues
func (spl *SerializableValues) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableSectionValues)
	for pair := spl.Sections.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableValues
func (spl *SerializableValues) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableSectionValues)
	for pair := spl.Sections.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
