package middlewares

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TableMiddleware interface {
	// Process transforms a full table
	Process(ctx context.Context, table *types.Table) (*types.Table, error)
}

type ObjectMiddleware interface {
	// Process transforms each individual object. Each object can return multiple
	// objects which will get processed individually downstream.
	Process(ctx context.Context, object types.Row) ([]types.Row, error)
}

type RowMiddleware interface {
	Process(ctx context.Context, row types.Row) ([]types.Row, error)
}
