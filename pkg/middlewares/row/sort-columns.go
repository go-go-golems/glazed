package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"sort"
)

type SortColumnsMiddleware struct {
}

func NewSortColumnsMiddleware() *SortColumnsMiddleware {
	return &SortColumnsMiddleware{}
}

func (scm *SortColumnsMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	fields := types.GetFields(row)
	sort.Strings(fields)

	newRow := types.NewRow()
	for _, field := range fields {
		value, _ := row.Get(field)
		newRow.Set(field, value)
	}
	return []types.Row{newRow}, nil
}
