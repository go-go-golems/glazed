package table

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type RemoveDuplicatesMiddleware struct {
	columns []string
}

func NewRemoveDuplicatesMiddleware(columns ...string) *RemoveDuplicatesMiddleware {
	return &RemoveDuplicatesMiddleware{
		columns: columns,
	}
}

func (r *RemoveDuplicatesMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: table.Columns,
		Rows:    make([]types.Row, 0),
	}

	var previousRowValues types.MapRow

	for _, row := range table.Rows {
		values := row.GetValues()
		if previousRowValues != nil {
			// check if the values are the same
			same := true
			for _, column := range r.columns {
				v, ok := values.Get(column)
				v2, ok2 := previousRowValues.Get(column)
				if ok != ok2 || v != v2 {
					same = false
					break
				}
			}
			if same {
				continue
			}
		}
		ret.Rows = append(ret.Rows, row)
		previousRowValues = values
	}

	return ret, nil
}
