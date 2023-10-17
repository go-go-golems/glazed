package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type SkipLimitMiddleware struct {
	Skip  int
	Limit int
	count int
}

var _ middlewares.RowMiddleware = (*SkipLimitMiddleware)(nil)

func (h *SkipLimitMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	defer func() { h.count++ }()
	if h.count < h.Skip {

		return nil, nil
	}

	if h.Limit > 0 {
		if h.count >= h.Skip+h.Limit {
			return nil, nil
		}
	}

	return []types.Row{row}, nil
}

func (h *SkipLimitMiddleware) Close(ctx context.Context) error {
	return nil
}
