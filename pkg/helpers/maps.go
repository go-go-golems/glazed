package helpers

import (
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
