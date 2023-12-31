package cast

import (
	"fmt"
	"reflect"
)

func ConvertMapToInterfaceMap(input interface{}) (map[string]interface{}, error) {
	value := reflect.ValueOf(input)

	// Check if the input is a map
	if value.Kind() != reflect.Map {
		return nil, fmt.Errorf("input is not a map")
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
