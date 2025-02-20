package cast

import (
	"fmt"
	"reflect"
)

// NormalizeValue converts various types into a normalized format for comparison
func NormalizeValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case []interface{}:
		return normalizeSlice(val)
	case map[interface{}]interface{}:
		return normalizeMap(val)
	case map[string]interface{}:
		return normalizeStringMap(val)
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return val, nil
	default:
		// Handle slices that aren't []interface{}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			interfaceSlice := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				interfaceSlice[i] = rv.Index(i).Interface()
			}
			return normalizeSlice(interfaceSlice)
		}

		// Try to convert to string if possible
		str, err := ToString(v)
		if err == nil {
			return str, nil
		}
		// Return as-is if we can't normalize
		return v, nil
	}
}

func normalizeSlice(slice []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		normalized, err := NormalizeValue(v)
		if err != nil {
			return nil, fmt.Errorf("error normalizing slice element %d: %w", i, err)
		}
		result[i] = normalized
	}
	return result, nil
}

func normalizeMap(m map[interface{}]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		key, err := ToString(k)
		if err != nil {
			return nil, fmt.Errorf("error converting map key to string: %w", err)
		}
		normalized, err := NormalizeValue(v)
		if err != nil {
			return nil, fmt.Errorf("error normalizing map value for key %s: %w", key, err)
		}
		result[key] = normalized
	}
	return result, nil
}

func normalizeStringMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		normalized, err := NormalizeValue(v)
		if err != nil {
			return nil, fmt.Errorf("error normalizing map value for key %s: %w", k, err)
		}
		result[k] = normalized
	}
	return result, nil
}
