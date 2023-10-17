package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type AddFieldMiddleware struct {
	Fields map[string]string
}

var _ middlewares.RowMiddleware = (*AddFieldMiddleware)(nil)

func (a *AddFieldMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewAddFieldMiddleware(fields map[string]string) *AddFieldMiddleware {
	return &AddFieldMiddleware{Fields: fields}
}

func (a *AddFieldMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newValues := types.NewRow()
	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		newValues.Set(key, value)
	}
	for key, value := range a.Fields {
		newValues.Set(key, value)
	}
	return []types.Row{newValues}, nil
}
