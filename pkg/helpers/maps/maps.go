package maps

import (
	"reflect"
	"strings"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func StructToMap(i interface{}, lowerCaseKeys bool) map[string]interface{} {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	m := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		if lowerCaseKeys {
			name = strings.ToLower(name)
		}
		m[name] = val.Field(i).Interface()
	}

	return m
}

func StructToMapThroughYAML(s interface{}) (map[string]interface{}, error) {
	// Marshal the struct to YAML
	yamlData, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML into a map
	var m map[string]interface{}
	err = yaml.Unmarshal(yamlData, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func GetValues[Key comparable, Value any](m map[Key]Value) []Value {
	values := make([]Value, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}

func IsStructPointer(s interface{}) bool {
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return false
	}
	// check if nil
	if reflect.ValueOf(s).IsNil() {
		return false
	}
	if reflect.TypeOf(s).Elem().Kind() != reflect.Struct {
		return false
	}

	return true
}

// GlazedStructToMap converts a struct pointer to a map of field names to values.
// It iterates through the struct fields looking for the "glazed" tag.
// For each field with the tag, it will add an entry to the returned map with the
// tag value as the key and the field's value as the map value.
// Returns an error if s is not a pointer to a struct.
func GlazedStructToMap(s interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}

	// check that s is indeed a pointer to a struct
	if !IsStructPointer(s) {
		return nil, errors.Errorf("s is not a pointer to a struct")
	}

	st := reflect.TypeOf(s).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		fieldName, ok := field.Tag.Lookup("glazed")
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		ret[fieldName] = value
	}

	return ret, nil
}

// Get retrieves a nested value from a map using a sequence of keys.
// It safely handles missing keys or incorrect types at any level.
// Returns the value cast to type T and true if found and castable, otherwise the zero value of T and false.
func Get[T any](m map[string]interface{}, keys ...string) (T, bool) {
	var current interface{} = m

	for i, key := range keys {
		mapCurrent, ok := current.(map[string]interface{})
		if !ok {
			var zero T
			return zero, false
		}

		val, exists := mapCurrent[key]
		if !exists {
			var zero T
			return zero, false
		}

		if i == len(keys)-1 {
			// Last key, try to cast to T
			finalVal, ok := val.(T)
			if !ok {
				var zero T
				return zero, false
			}
			return finalVal, true
		} else {
			// Navigate deeper
			current = val
		}
	}

	// Should not happen if keys is not empty, but handle defensively
	var zero T
	return zero, false
}

// GetString retrieves a nested string value from a map.
// It's a convenience wrapper around Get[string].
func GetString(m map[string]interface{}, keys ...string) (string, bool) {
	return Get[string](m, keys...)
}

// GetInteger retrieves a nested integer value from a map.
// It handles potential numeric type mismatches (e.g., float64 from JSON) using cast helpers.
func GetInteger[T cast.SignedInt | cast.UnsignedInt](m map[string]interface{}, keys ...string) (T, bool) {
	rawValue, ok := Get[interface{}](m, keys...)
	if !ok {
		var zero T
		return zero, false
	}

	// Use cast helper to handle potential float64 etc.
	return cast.CastNumberInterfaceToInt[T](rawValue)
}
