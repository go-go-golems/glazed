package list

import (
	"fmt"
	"strings"
)

func Reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// SliceToCSV converts a generic slice to a comma separated string.
func SliceToCSV[T any](items []T) string {
	var sb strings.Builder
	for i, item := range items {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprint(item))
	}
	return sb.String()
}
