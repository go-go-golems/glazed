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

// CastListToStringList casts a list of items to a list of strings.
// It handles both string aliases and string declarations.
func CastListToStringList(list interface{}) ([]string, error) {
	val := reflect.ValueOf(list)

	// Check if the value is a slice or array
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("the provided value is not a list")
	}

	elemKind := val.Type().Elem().Kind()
	if elemKind != reflect.String && elemKind != reflect.Interface {
		return nil, fmt.Errorf("the list does not contain strings")
	}

	result := make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		str, err := ToString(val.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		result[i] = str
	}

	return result, nil
}

// CastStringListToList casts a list of strings to a list of a specified type T.
// The target type T must be a string alias or string declaration.
func CastStringListToList[T ~string](list []string) ([]T, error) {
	result := make([]T, len(list))
	for i, str := range list {
		result[i] = T(str)
	}
	return result, nil
}
