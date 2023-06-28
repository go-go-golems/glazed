package row

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleReorder(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	)

	mw := NewReorderColumnOrderMiddleware(
		[]types.FieldName{"c", "a", "b"},
	)
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	require.Len(t, newRows, 1)
	assert2.EqualRow(t, types.NewRow(
		types.MRP("c", "value3"),
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
	), newRows[0])
}

func TestSimpleReorderNotAllColumns(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	)

	mw := NewReorderColumnOrderMiddleware(
		[]types.FieldName{"c", "a"},
	)
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	require.Len(t, newRows, 1)
	assert2.EqualRow(t, types.NewRow(
		types.MRP("c", "value3"),
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
	), newRows[0])
}

func TestReorderDot(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c.d", "value3"),
		types.MRP("c.e", "value4"),
	)

	mw := NewReorderColumnOrderMiddleware(
		[]types.FieldName{"c.", "a", "b"},
	)
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	require.Len(t, newRows, 1)
	assert2.EqualRow(t, types.NewRow(
		types.MRP("c.d", "value3"),
		types.MRP("c.e", "value4"),
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
	), newRows[0])
}

func TestReorderDotOverride(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c.d", "value3"),
		types.MRP("c.e", "value4"),
		types.MRP("c.f", "value5"),
	)

	mw := NewReorderColumnOrderMiddleware(
		[]types.FieldName{"c.e", "a", "c.", "b"},
	)
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	require.Len(t, newRows, 1)
	assert2.EqualRow(t, types.NewRow(
		types.MRP("c.e", "value4"),
		types.MRP("a", "value1"),
		types.MRP("c.d", "value3"),
		types.MRP("c.f", "value5"),
		types.MRP("b", "value2"),
	), newRows[0])
}
