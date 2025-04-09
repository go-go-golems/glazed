package row

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// FieldsFilterMiddleware keeps columns that are in the fields list and removes
// columns that are in the filters list.
//
// empty lists means that all columns are accepted.
//
// The returned rows are SimpleRows
type FieldsFilterMiddleware struct {
	fields        *orderedmap.OrderedMap[string, interface{}]
	filters       *orderedmap.OrderedMap[string, interface{}]
	prefixFields  []string
	prefixFilters []string
	regexFields   []*regexp.Regexp
	regexFilters  []*regexp.Regexp

	newColumns map[types.FieldName]interface{}
}

type FieldsFilterOption func(*FieldsFilterMiddleware)

// WithFields adds exact field matches to keep
func WithFields(fields []string) FieldsFilterOption {
	return func(ffm *FieldsFilterMiddleware) {
		for _, field := range fields {
			if strings.HasSuffix(field, ".") {
				ffm.prefixFields = append(ffm.prefixFields, field)
			} else {
				ffm.fields.Set(field, nil)
			}
		}
	}
}

// WithFilters adds exact field matches to remove
func WithFilters(filters []string) FieldsFilterOption {
	return func(ffm *FieldsFilterMiddleware) {
		for _, filter := range filters {
			if strings.HasSuffix(filter, ".") {
				ffm.prefixFilters = append(ffm.prefixFilters, filter)
			} else {
				ffm.filters.Set(filter, nil)
			}
		}
	}
}

// WithRegexFields adds regex patterns for fields to keep
func WithRegexFields(patterns []string) FieldsFilterOption {
	return func(ffm *FieldsFilterMiddleware) {
		for _, pattern := range patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				ffm.regexFields = append(ffm.regexFields, re)
			}
		}
	}
}

// WithRegexFilters adds regex patterns for fields to remove
func WithRegexFilters(patterns []string) FieldsFilterOption {
	return func(ffm *FieldsFilterMiddleware) {
		for _, pattern := range patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				ffm.regexFilters = append(ffm.regexFilters, re)
			}
		}
	}
}

var _ middlewares.RowMiddleware = (*FieldsFilterMiddleware)(nil)

func (ffm *FieldsFilterMiddleware) Close(ctx context.Context) error {
	return nil
}

// NewFieldsFilterMiddleware creates a new FieldsFilterMiddleware with the given options
func NewFieldsFilterMiddleware(options ...FieldsFilterOption) *FieldsFilterMiddleware {
	ffm := &FieldsFilterMiddleware{
		fields:     orderedmap.New[string, interface{}](),
		filters:    orderedmap.New[string, interface{}](),
		newColumns: map[types.FieldName]interface{}{},
	}

	for _, option := range options {
		option(ffm)
	}

	return ffm
}

func (ffm *FieldsFilterMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	if ffm.fields.Len() == 0 && ffm.filters.Len() == 0 &&
		len(ffm.prefixFields) == 0 && len(ffm.prefixFilters) == 0 &&
		len(ffm.regexFields) == 0 && len(ffm.regexFilters) == 0 {
		return []types.Row{row}, nil
	}

	newRow := types.NewRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		rowField, value := pair.Key, pair.Value

		// skip if we already filtered that field
		if _, ok := ffm.newColumns[rowField]; !ok {
			exactMatchFound := false
			prefixMatchFound := false
			regexMatchFound := false

			exactFilterMatchFound := false
			prefixFilterMatchFound := false
			regexFilterMatchFound := false

			// Check exact matches
			if ffm.fields.Len() > 0 {
				if _, ok := ffm.fields.Get(rowField); ok {
					exactMatchFound = true
				}
			}

			// Check prefix matches
			if !exactMatchFound && len(ffm.prefixFields) > 0 {
				for _, prefix := range ffm.prefixFields {
					if strings.HasPrefix(rowField, prefix) {
						prefixMatchFound = true
						break
					}
				}
			}

			// Check regex matches
			if !exactMatchFound && !prefixMatchFound && len(ffm.regexFields) > 0 {
				for _, re := range ffm.regexFields {
					if re.MatchString(rowField) {
						regexMatchFound = true
						break
					}
				}
			}

			// Check filters
			if ffm.filters.Len() > 0 {
				if _, ok := ffm.filters.Get(rowField); ok {
					exactFilterMatchFound = true
				}
			}

			if !exactFilterMatchFound && len(ffm.prefixFilters) > 0 {
				for _, prefix := range ffm.prefixFilters {
					if strings.HasPrefix(rowField, prefix) {
						prefixFilterMatchFound = true
						break
					}
				}
			}

			if !exactFilterMatchFound && !prefixFilterMatchFound && len(ffm.regexFilters) > 0 {
				for _, re := range ffm.regexFilters {
					if re.MatchString(rowField) {
						regexFilterMatchFound = true
						break
					}
				}
			}

			// Determine if field should be included
			shouldInclude := false

			if exactMatchFound {
				shouldInclude = true
			} else if prefixMatchFound {
				if prefixFilterMatchFound {
					// Include by default when both prefix match
					shouldInclude = true
				} else if exactFilterMatchFound {
					shouldInclude = false
				} else {
					shouldInclude = true
				}
			} else if regexMatchFound {
				if regexFilterMatchFound {
					// Include by default when both regex match
					shouldInclude = true
				} else if exactFilterMatchFound {
					shouldInclude = false
				} else {
					shouldInclude = true
				}
			} else if exactFilterMatchFound || prefixFilterMatchFound || regexFilterMatchFound {
				shouldInclude = false
			} else if ffm.fields.Len() == 0 && len(ffm.prefixFields) == 0 && len(ffm.regexFields) == 0 {
				// If no fields specified, include everything not filtered
				shouldInclude = true
			}

			if shouldInclude {
				ffm.newColumns[rowField] = nil
				newRow.Set(rowField, value)
			}
		} else {
			newRow.Set(rowField, value)
		}
	}

	return []types.Row{newRow}, nil
}
