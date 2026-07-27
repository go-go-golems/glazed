package middlewares

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
)

type processorTestTableMiddleware struct{}

func (*processorTestTableMiddleware) Process(_ context.Context, table *types.Table) (*types.Table, error) {
	return table, nil
}

func (*processorTestTableMiddleware) Close(context.Context) error { return nil }

func TestTableProcessorPreservesPreferredOrderAcrossSparseRows(t *testing.T) {
	processor := NewTableProcessor(WithTableMiddleware(&processorTestTableMiddleware{}))
	processor.SetPreferredColumnOrder("a", "missing", "b")

	ctx := context.Background()
	require.NoError(t, processor.AddRow(ctx, types.NewRow(types.MRP("a", 1))))
	require.NoError(t, processor.AddRow(ctx, types.NewRow(types.MRP("b", 2))))

	require.Equal(t, []types.FieldName{"a", "b"}, processor.Table.Columns)
}
