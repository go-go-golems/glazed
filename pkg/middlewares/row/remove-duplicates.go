package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type RemoveDuplicatesMiddleware struct {
	columns           []string
	previousRowValues types.Row
}

func NewRemoveDuplicatesMiddleware(columns ...string) *RemoveDuplicatesMiddleware {
	return &RemoveDuplicatesMiddleware{
		columns: columns,
	}
}

func (r *RemoveDuplicatesMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	if r.previousRowValues != nil {
		// check if the values are the same
		same := true
		for _, column := range r.columns {
			v, ok := row.Get(column)
			v2, ok2 := r.previousRowValues.Get(column)
			if ok != ok2 || v != v2 {
				same = false
				break
			}
		}
		if same {
			return []types.Row{}, nil
		}
	}
	r.previousRowValues = row
	return []types.Row{row}, nil
}
