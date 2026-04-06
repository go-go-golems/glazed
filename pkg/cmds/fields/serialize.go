package fields

import (
	"encoding/json"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableFieldValue represents a parsed field in a format suitable for
// YAML/JSON serialization, excluding the Definition
type SerializableFieldValue struct {
	Value interface{} `yaml:"value" json:"value"`
	Log   []ParseStep `yaml:"log" json:"log"`
}

func toSerializableParseLog(def *Definition, log []ParseStep) []ParseStep {
	if len(log) == 0 {
		return nil
	}

	ret := make([]ParseStep, 0, len(log))
	for _, step := range log {
		if def != nil {
			ret = append(ret, RedactParseStep(def.Type, step))
		} else {
			ret = append(ret, step)
		}
	}

	return ret
}

// ToSerializableFieldValue converts a FieldValue to its serializable representation.
func ToSerializableFieldValue(pp *FieldValue) *SerializableFieldValue {
	if pp == nil {
		return nil
	}

	value := pp.Value
	if pp.Definition != nil {
		value = RedactValue(pp.Definition.Type, value)
	}

	return &SerializableFieldValue{
		Value: value,
		Log:   toSerializableParseLog(pp.Definition, pp.Log),
	}
}

// SerializableFieldValues represents a collection of parsed fields in a format suitable
// for YAML/JSON serialization, maintaining the order of fields
type SerializableFieldValues struct {
	// Using orderedmap to maintain field order while having name-based access
	Fields *orderedmap.OrderedMap[string, *SerializableFieldValue] `yaml:"fields" json:"fields"`
}

// ToSerializableFieldValues converts a FieldValues collection to its serializable representation
func ToSerializableFieldValues(pp *FieldValues) *SerializableFieldValues {
	ret := &SerializableFieldValues{
		Fields: orderedmap.New[string, *SerializableFieldValue](),
	}

	pp.ForEach(func(key string, value *FieldValue) {
		serialized := ToSerializableFieldValue(value)
		ret.Fields.Set(key, serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableFieldValues
func (spp *SerializableFieldValues) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableFieldValue)
	for pair := spp.Fields.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableFieldValues
func (spp *SerializableFieldValues) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableFieldValue)
	for pair := spp.Fields.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
