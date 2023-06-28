package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type RemoveNullsMiddleware struct {
}

func NewRemoveNullsMiddleware() *RemoveNullsMiddleware {
	return &RemoveNullsMiddleware{}
}

func (rnm *RemoveNullsMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := types.NewMapRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		if value != nil {
			newRow.Set(key, value)
		}
	}

	return []types.Row{newRow}, nil
}
