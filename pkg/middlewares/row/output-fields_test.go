package row

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestOutputFieldsMiddlewareProjectsInRequestedOrder(t *testing.T) {
	middleware := NewOutputFieldsMiddleware("name", "id", "missing")
	input := types.NewRow(
		types.MRP("id", 42),
		types.MRP("name", "Ada"),
		types.MRP("extra", true),
	)

	rows, err := middleware.Process(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, []types.FieldName{"name", "id"}, types.GetFields(rows[0]))
	name, ok := rows[0].Get("name")
	require.True(t, ok)
	require.Equal(t, "Ada", name)
	id, ok := rows[0].Get("id")
	require.True(t, ok)
	require.Equal(t, 42, id)
}

func TestOutputFieldsMiddlewareEmptyListPassesThrough(t *testing.T) {
	middleware := NewOutputFieldsMiddleware()
	input := types.NewRow(types.MRP("id", 42))

	rows, err := middleware.Process(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Same(t, input, rows[0])
}
