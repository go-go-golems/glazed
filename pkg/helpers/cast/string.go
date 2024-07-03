package cast

import (
	"fmt"
	"reflect"
)

// ToString converts a value to a string using reflection.
// It handles both type aliases and type declarations based on string.
//
// This function is useful for dynamically converting values to strings
// at runtime, especially when dealing with types that may not be known
// at compile time.
func ToString(value interface{}) (string, error) {
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	// Check if the kind is string
	if v.Kind() == reflect.String {
		return v.String(), nil
	}

	// Check if the underlying type is string
	if t.Kind() == reflect.String {
		return v.Convert(reflect.TypeOf("")).String(), nil
	}

	return "", fmt.Errorf("cannot convert %v to string", t)
}
