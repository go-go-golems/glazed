package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type RemoveNullsMiddleware struct {
}

var _ middlewares.RowMiddleware = (*RemoveNullsMiddleware)(nil)

func (rnm *RemoveNullsMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewRemoveNullsMiddleware() *RemoveNullsMiddleware {
	return &RemoveNullsMiddleware{}
}

func (rnm *RemoveNullsMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := types.NewRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		if value != nil {
			newRow.Set(key, value)
		}
	}

	return []types.Row{newRow}, nil
}
