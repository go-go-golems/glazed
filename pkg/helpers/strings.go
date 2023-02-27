package helpers

import "fmt"

func StringInSlice(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func InterfaceListToStringList(list []interface{}) []string {
	var result []string
	for _, item := range list {
		result = append(result, item.(string))
	}
	return result
}

func InterfaceToStringList(list interface{}) []string {
	return InterfaceListToStringList(list.([]interface{}))
}

func IntSliceToStringSlice(list []int) []string {
	var result []string
	for _, item := range list {
		result = append(result, fmt.Sprintf("%d", item))
	}
	return result
}

func Float64SliceToStringSlice(list []float64) []string {
	var result []string
	for _, item := range list {
		result = append(result, fmt.Sprintf("%f", item))
	}
	return result
}
