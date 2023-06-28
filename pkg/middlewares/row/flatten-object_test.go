package row

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlattenNoNestedObjects(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
	)

	mw := NewFlattenObjectMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	assert.Len(t, newRows, 1)
	assert2.EqualMapRows(t, row, newRows[0])
}

func TestFlattenSingleNestedObject(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", map[string]interface{}{
			"d": "value3",
		}),
	)

	mw := NewFlattenObjectMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	assert.Len(t, newRows, 1)
	assert2.EqualMapRows(t, types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c.d", "value3"),
	), newRows[0])
}

func TestFlattenTwoObjects(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", map[string]interface{}{
			"d": "value3",
			"a": "value4",
		}),
		types.MRP("e", map[string]interface{}{
			"f": "value5",
		}),
	)

	mw := NewFlattenObjectMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	assert.Len(t, newRows, 1)
	assert2.EqualMapRows(t, types.NewRow(
		types.MRP("a", "value1"),

		types.MRP("b", "value2"),
		types.MRP("c.a", "value4"),
		types.MRP("c.d", "value3"),
		types.MRP("e.f", "value5"),
	), newRows[0])
}

func TestNestedMapInterface(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", map[string]interface{}{
			"d": "value3",
			"a": "value4",
			"e": map[string]interface{}{
				"f": "value5",
			},
		}),
	)

	mw := NewFlattenObjectMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	assert.Len(t, newRows, 1)
	assert2.EqualMapRows(t, types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c.a", "value4"),
		types.MRP("c.d", "value3"),
		types.MRP("c.e.f", "value5"),
	), newRows[0])
}

func TestNestedMapInterfaceRow(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", types.NewRow(
			types.MRP("d", "value3"),
			types.MRP("a", "value4"),
			types.MRP("e", map[string]interface{}{
				"f": "value5",
				"a": "value6",
			}),
		)),
	)

	mw := NewFlattenObjectMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	assert.Len(t, newRows, 1)
	assert2.EqualMapRows(t, types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c.d", "value3"),
		types.MRP("c.a", "value4"),
		types.MRP("c.e.a", "value6"),
		types.MRP("c.e.f", "value5"),
	), newRows[0])
}
