package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"strings"
)

// FieldsFilterMiddleware keeps columns that are in the fields list and removes
// columns that are in the filters list.
//
// empty lists means that all columns are accepted.
//
// The returned rows are SimpleRows
type FieldsFilterMiddleware struct {
	fields        map[string]interface{}
	filters       map[string]interface{}
	prefixFields  []string
	prefixFilters []string

	newColumns map[types.FieldName]interface{}
}

func (ffm *FieldsFilterMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewFieldsFilterMiddleware(fields []string, filters []string) *FieldsFilterMiddleware {
	fieldHash := map[string]interface{}{}
	prefixFields := []string{}
	prefixFilters := []string{}

	for _, field := range fields {
		if strings.HasSuffix(field, ".") {
			prefixFields = append(prefixFields, field)
		} else {
			fieldHash[field] = nil
		}
	}
	filterHash := map[string]interface{}{}
	for _, filter := range filters {
		if strings.HasSuffix(filter, ".") {
			prefixFilters = append(prefixFilters, filter)
		} else {
			filterHash[filter] = nil
		}
	}
	return &FieldsFilterMiddleware{
		fields:        fieldHash,
		filters:       filterHash,
		prefixFields:  prefixFields,
		prefixFilters: prefixFilters,
		newColumns:    map[types.FieldName]interface{}{},
	}
}

func (ffm *FieldsFilterMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	if len(ffm.fields) == 0 && len(ffm.filters) == 0 {
		return []types.Row{row}, nil
	}

	newRow := types.NewRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		rowField, value := pair.Key, pair.Value

		// skip all of this if we already filtered that field
		if _, ok := ffm.newColumns[rowField]; !ok {
			exactMatchFound := false
			prefixMatchFound := false

			exactFilterMatchFound := false
			prefixFilterMatchFound := false

			// go through all the fields and prefix fields and check if the current field matches
			if len(ffm.fields) > 0 || len(ffm.prefixFields) > 0 {
				// first go through exact matches
				if _, ok := ffm.fields[rowField]; ok {
					exactMatchFound = true
				} else {
					// else, test against all prefixes
					for _, prefix := range ffm.prefixFields {
						if strings.HasPrefix(rowField, prefix) {
							prefixMatchFound = true
							break
						}
					}
				}

				if !exactMatchFound && !prefixMatchFound {
					continue
				}
			}

			if len(ffm.filters) > 0 || len(ffm.prefixFilters) > 0 {
				// if an exact filter matches, move on
				if _, ok := ffm.filters[rowField]; ok {
					exactFilterMatchFound = true
					continue
				} else {
					// else, test against all prefixes
					for _, prefix := range ffm.prefixFilters {
						if strings.HasPrefix(rowField, prefix) {
							prefixFilterMatchFound = true
							break
						}
					}
				}
			}

			if exactMatchFound {
				ffm.newColumns[rowField] = nil
			} else if prefixMatchFound {
				if prefixFilterMatchFound {
					// should we do by prefix length, nah...
					// choose to include by default
					ffm.newColumns[rowField] = nil
				} else if exactFilterMatchFound {
					continue
				} else {
					ffm.newColumns[rowField] = nil
				}
			} else if exactFilterMatchFound {
				continue
			} else if len(ffm.fields) == 0 {
				ffm.newColumns[rowField] = nil
			}
		}

		newRow.Set(rowField, value)
	}

	return []types.Row{newRow}, nil
}
