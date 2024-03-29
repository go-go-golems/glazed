package table

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/compare"
	"github.com/go-go-golems/glazed/pkg/types"
	"sort"
)

type columnOrder struct {
	name string
	asc  bool
}

// SortByMiddleware sorts the table by the given columns, in ascending or descending order.
// If the column contains rows with the same type, the order is undefined.
type SortByMiddleware struct {
	columns []columnOrder
}

func (s *SortByMiddleware) Close(ctx context.Context) error {
	return nil
}

// NewSortByMiddlewareFromColumns creates a new SortByMiddleware from the given columns.
// To sort in descending order, prefix the column name with a minus sign.
//
// Example:
//
//	NewSortByMiddlewareFromColumns("name", "-age")
//
// This will sort by name in ascending order and by age in descending order.
func NewSortByMiddlewareFromColumns(columns ...string) *SortByMiddleware {
	ret := &SortByMiddleware{
		columns: make([]columnOrder, 0),
	}

	for _, column := range columns {
		if len(column) == 0 {
			continue
		}

		isAsc := true

		if column[0] == '-' {
			column = column[1:]
			isAsc = false
		}

		ret.columns = append(ret.columns, columnOrder{
			name: column,
			asc:  isAsc,
		})
	}

	return ret
}

func (s *SortByMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: table.Columns,
		Rows:    make([]types.Row, 0),
	}

	ret.Rows = append(ret.Rows, table.Rows...)

	if len(s.columns) == 0 {
		return ret, nil
	}

	sort.Slice(ret.Rows, func(i, j int) bool {
		rowA := ret.Rows[i]
		rowB := ret.Rows[j]

		for _, column := range s.columns {
			v, ok := rowA.Get(column.name)
			v2, ok2 := rowB.Get(column.name)
			if ok == ok2 && v == v2 {
				continue
			}

			if compare.IsLowerThan(v, v2) {
				return column.asc
			} else {
				return !column.asc
			}
		}

		return false
	})

	return ret, nil
}
