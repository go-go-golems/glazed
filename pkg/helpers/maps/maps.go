package maps

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
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

// GlazedStructToMap converts a struct pointer to a map of parameter names to values.
// It iterates through the struct fields looking for the "glazed.parameter" tag.
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
		parameterName, ok := field.Tag.Lookup("glazed.parameter")
		if !ok {
			continue
		}
		value := reflect.ValueOf(s).Elem().FieldByName(field.Name)

		ret[parameterName] = value
	}

	return ret, nil
}
