package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type LambdaMiddleware struct {
	Function func(ctx context.Context, row types.Row) ([]types.Row, error)
}

var _ middlewares.RowMiddleware = (*LambdaMiddleware)(nil)

func (l *LambdaMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewLambdaMiddleware(function func(ctx context.Context, row types.Row) ([]types.Row, error)) *LambdaMiddleware {
	return &LambdaMiddleware{Function: function}
}

func (l *LambdaMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	return l.Function(ctx, row)
}
