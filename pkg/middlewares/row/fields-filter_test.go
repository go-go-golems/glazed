package row

import (
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createFieldsFilterTestRows() []types.Row {
	return []types.Row{
		types.NewMapRow(
			types.MRP("a", 1),
			types.MRP("b", 2),
			types.MRP("c", 3),
		),
		types.NewMapRow(
			types.MRP("a", 4),
			types.MRP("b", 5),
			types.MRP("c", 6),
		),
		types.NewMapRow(
			types.MRP("a", 7),
			types.MRP("b", 8),
			types.MRP("c", 9),
			types.MRP("d", 10),
		),
	}
}

func TestNoFieldsFilter(t *testing.T) {
	mw := NewFieldsFilterMiddleware([]string{}, []string{})

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 3)
	assert2.EqualMapRowValues(t, rows[0], map[string]interface{}{"a": 1, "b": 2, "c": 3})
	assert2.EqualMapRowValues(t, rows[1], map[string]interface{}{"a": 4, "b": 5, "c": 6})
	assert2.EqualMapRowValues(t, rows[2], map[string]interface{}{"a": 7, "b": 8, "c": 9, "d": 10})
}

func TestFieldsA(t *testing.T) {
	mw := NewFieldsFilterMiddleware([]string{"a"}, []string{})

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 3)
	assert2.EqualMapRowValues(t, rows[0], map[string]interface{}{"a": 1})
	assert2.EqualMapRowValues(t, rows[1], map[string]interface{}{"a": 4})
	assert2.EqualMapRowValues(t, rows[2], map[string]interface{}{"a": 7})
}

func TestFieldsBD(t *testing.T) {
	mw := NewFieldsFilterMiddleware([]string{"b", "d"}, []string{})

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 3)
	assert2.EqualMapRowValues(t, rows[0], map[string]interface{}{"b": 2})
	assert2.EqualMapRowValues(t, rows[1], map[string]interface{}{"b": 5})
	assert2.EqualMapRowValues(t, rows[2], map[string]interface{}{"b": 8, "d": 10})
}

func TestFilterA(t *testing.T) {
	mw := NewFieldsFilterMiddleware([]string{}, []string{"a"})

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 3)
	assert2.EqualMapRowValues(t, rows[0], map[string]interface{}{"b": 2, "c": 3})
	assert2.EqualMapRowValues(t, rows[1], map[string]interface{}{"b": 5, "c": 6})
	assert2.EqualMapRowValues(t, rows[2], map[string]interface{}{"b": 8, "c": 9, "d": 10})
}

func TestFilterBD(t *testing.T) {
	mw := NewFieldsFilterMiddleware([]string{}, []string{"b", "d"})

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 3)
	assert2.EqualMapRowValues(t, rows[0], map[string]interface{}{"a": 1, "c": 3})
	assert2.EqualMapRowValues(t, rows[1], map[string]interface{}{"a": 4, "c": 6})
	assert2.EqualMapRowValues(t, rows[2], map[string]interface{}{"a": 7, "c": 9})
}
