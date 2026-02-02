package fields

import (
	"encoding/json"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// SerializableParsedParameter represents a parsed parameter in a format suitable for
// YAML/JSON serialization, excluding the Definition
type SerializableParsedParameter struct {
	Value interface{} `yaml:"value" json:"value"`
	Log   []ParseStep `yaml:"log" json:"log"`
}

// ToSerializableParsedParameter converts a ParsedParameter to its serializable representation
func ToSerializableParsedParameter(pp *ParsedParameter) *SerializableParsedParameter {
	return &SerializableParsedParameter{
		Value: pp.Value,
		Log:   pp.Log,
	}
}

// SerializableParsedParameters represents a collection of parsed parameters in a format suitable
// for YAML/JSON serialization, maintaining the order of parameters
type SerializableParsedParameters struct {
	// Using orderedmap to maintain parameter order while having name-based access
	Parameters *orderedmap.OrderedMap[string, *SerializableParsedParameter] `yaml:"parameters" json:"parameters"`
}

// ToSerializableParsedParameters converts a ParsedParameters collection to its serializable representation
func ToSerializableParsedParameters(pp *ParsedParameters) *SerializableParsedParameters {
	ret := &SerializableParsedParameters{
		Parameters: orderedmap.New[string, *SerializableParsedParameter](),
	}

	pp.ForEach(func(key string, value *ParsedParameter) {
		serialized := ToSerializableParsedParameter(value)
		ret.Parameters.Set(key, serialized)
	})

	return ret
}

// MarshalYAML implements yaml.Marshaler for SerializableParsedParameters
func (spp *SerializableParsedParameters) MarshalYAML() (interface{}, error) {
	// Convert to a map for YAML serialization
	m := make(map[string]*SerializableParsedParameter)
	for pair := spp.Parameters.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m, nil
}

// MarshalJSON implements json.Marshaler for SerializableParsedParameters
func (spp *SerializableParsedParameters) MarshalJSON() ([]byte, error) {
	// Convert to a map for JSON serialization
	m := make(map[string]*SerializableParsedParameter)
	for pair := spp.Parameters.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return json.Marshal(m)
}
