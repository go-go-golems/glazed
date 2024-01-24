package cast

import (
	"errors"
	"reflect"
)

// ToInterfaceValue converts a given value of any basic Go type (scalars, slices, arrays, maps) into a corresponding
// interface{}. Scalar types are returned as is. Slices and arrays are converted to []interface{} with each element
// being recursively processed. Maps are converted to map[string]interface{} with each value being recursively
// processed. The function uses reflection to handle various types dynamically. It returns an error if it encounters
// a type that it cannot process, or if map keys are not strings.
func ToInterfaceValue(value interface{}) (interface{}, error) {
	if nil == value {
		return nil, nil
	}
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Bool, reflect.String:
		return val.Interface(), nil // Scalars are returned as-is.

	case reflect.Slice, reflect.Array:
		return processSlice(val)

	case reflect.Map:
		return processMap(val)

	default:
		return nil, errors.New("unsupported type")
	}
}

// processSlice takes a reflect.Value representing a slice or an array and converts it into []interface{}.
// Each element of the input slice/array is recursively converted to an interface{} using
// toInterfaceValue. It returns a converted slice and any error encountered during the conversion.
func processSlice(slice reflect.Value) ([]interface{}, error) {
	result := make([]interface{}, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		convertedElem, err := ToInterfaceValue(elem.Interface())
		if err != nil {
			return nil, err
		}
		result[i] = convertedElem
	}
	return result, nil
}

// processMap takes a reflect.Value representing a map and converts it into map[string]interface{} recursively.
// The function returns an error if it encounters non-string keys or if any value is of an
// unsupported type.
func processMap(mapValue reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range mapValue.MapKeys() {
		val := mapValue.MapIndex(key)
		convertedVal, err := ToInterfaceValue(val.Interface())
		if err != nil {
			return nil, err
		}

		// Ensure the key is a string
		if key.Kind() != reflect.String {
			return nil, errors.New("map key is not a string")
		}
		result[key.String()] = convertedVal
	}
	return result, nil
}
