package cli

import (
	"fmt"
	"strings"
)

// CleanupColumns removes columns that match the filters and keeps columns that match the fields
func CleanupColumns(rows []map[string]interface{}, fields []string, filters []string) []string {
	ret := map[string]interface{}{}

	for _, row := range rows {
	Keys:
		for key := range row {
			if key != "name" {
				if len(filters) > 0 {
					for _, filter := range filters {
						if strings.HasSuffix(filter, ".") {
							if strings.HasPrefix(key, filter) {
								continue Keys
							}
						} else {
							if key == filter {
								continue Keys
							}
						}
					}
				}

				if len(fields) > 0 {
					for _, field := range fields {
						if strings.HasSuffix(field, ".") {
							if strings.HasPrefix(key, field) {
								ret[key] = nil
							}
						} else {
							if key == field {
								ret[key] = nil
							}
						}
					}
				} else {
					ret[key] = nil
				}
			}
		}
	}

	var keys []string
	for k := range ret {
		keys = append(keys, k)
	}

	return keys
}

func FlattenMapIntoColumns(rows map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{}

	for key, value := range rows {
		switch v := value.(type) {
		case map[string]interface{}:
			for k, v := range FlattenMapIntoColumns(v) {
				ret[fmt.Sprintf("%s.%s", key, k)] = v
			}
		default:
			ret[key] = v
		}
	}

	return ret
}
