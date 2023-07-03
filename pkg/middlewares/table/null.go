package table

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

// The NullTableMiddleware is only used to keep rows in the TableProcessor.Table.
type NullTableMiddleware struct{}

func (n *NullTableMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	return table, nil
}

func (n *NullTableMiddleware) Close(ctx context.Context) error {
	return nil
}
