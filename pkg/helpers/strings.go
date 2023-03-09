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

func ToAlphaString(n int) string {
	if n <= 0 {
		return ""
	}
	// we need to subtract 1 because the first column index is A, not 0
	n--
	// divide n by 26 to get the quotient and remainder
	quotient := n / 26
	remainder := n % 26
	// if the quotient is 0, the column index is just the corresponding letter
	if quotient == 0 {
		return string(rune('A' + remainder))
	}
	// otherwise, we need to recursively call ToAlphaString with the quotient and add the corresponding letter
	return ToAlphaString(quotient) + string(rune('A'+remainder))
}
