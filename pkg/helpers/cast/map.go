package cast

import (
	"github.com/pkg/errors"
	"reflect"
)

func ConvertMapToInterfaceMap(input interface{}) (map[string]interface{}, error) {
	value := reflect.ValueOf(input)

	// Check if the input is a map
	if value.Kind() != reflect.Map {
		return nil, errors.New("input is not a map")
	}

	// Initialize the result map
	result := make(map[string]interface{})

	// Iterate over the map keys
	for _, key := range value.MapKeys() {
		// Get the string key
		strKey := key.Interface().(string)

		// Get the value associated with the key
		val := value.MapIndex(key)

		// Set the key/value pair in the result map
		result[strKey] = val.Interface()
	}

	return result, nil
}

func CastStringMap[To any, From any](m map[string]From) (map[string]To, bool) {
	ret := map[string]To{}

	for k, v := range m {
		var item interface{} = v
		casted, ok := item.(To)
		if !ok {
			return ret, false
		}

		ret[k] = casted
	}

	return ret, true
}

func CastInterfaceToStringMap[To any, From any](m interface{}) (map[string]To, bool) {
	ret := map[string]To{}

	switch m := m.(type) {
	case map[string]To:
		return m, true
	case map[string]interface{}:
		return CastStringMap[To, interface{}](m)
	case map[string]From:
		return CastStringMap[To, From](m)
	default:
		return ret, false
	}
}

func CastMapToInterfaceMap[From any](m map[string]From) map[string]interface{} {
	ret := map[string]interface{}{}

	for k, v := range m {
		ret[k] = v
	}

	return ret
}
