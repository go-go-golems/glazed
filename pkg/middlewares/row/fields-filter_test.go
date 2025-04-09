package row

import (
	"context"
	"testing"

	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
)

func createFieldsFilterTestRows() []types.Row {
	return []types.Row{
		types.NewRow(
			types.MRP("a", 1),
			types.MRP("b", 2),
			types.MRP("c", 3),
			types.MRP("test1", "value1"),
			types.MRP("test2", "value2"),
			types.MRP("other1", "other1"),
		),
		types.NewRow(
			types.MRP("a", 4),
			types.MRP("b", 5),
			types.MRP("c", 6),
			types.MRP("test3", "value3"),
			types.MRP("test4", "value4"),
			types.MRP("other2", "other2"),
		),
	}
}

func TestNoFieldsFilter(t *testing.T) {
	mw := NewFieldsFilterMiddleware()

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"a": 1, "b": 2, "c": 3,
		"test1": "value1", "test2": "value2",
		"other1": "other1",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"a": 4, "b": 5, "c": 6,
		"test3": "value3", "test4": "value4",
		"other2": "other2",
	})
}

func TestFieldsA(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithFields([]string{"a"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{"a": 1})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{"a": 4})
}

func TestFieldsBD(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithFields([]string{"b", "d"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{"b": 2})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{"b": 5})
}

func TestFilterA(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithFilters([]string{"a"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"b": 2, "c": 3,
		"test1": "value1", "test2": "value2",
		"other1": "other1",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"b": 5, "c": 6,
		"test3": "value3", "test4": "value4",
		"other2": "other2",
	})
}

func TestFilterBD(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithFilters([]string{"b", "d"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"a": 1, "c": 3,
		"test1": "value1", "test2": "value2",
		"other1": "other1",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"a": 4, "c": 6,
		"test3": "value3", "test4": "value4",
		"other2": "other2",
	})
}

func TestRegexFields(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithRegexFields([]string{"^test[0-9]+$"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"test1": "value1",
		"test2": "value2",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"test3": "value3",
		"test4": "value4",
	})
}

func TestRegexFilters(t *testing.T) {
	mw := NewFieldsFilterMiddleware(WithRegexFilters([]string{"^test[0-9]+$"}))

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"a": 1, "b": 2, "c": 3,
		"other1": "other1",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"a": 4, "b": 5, "c": 6,
		"other2": "other2",
	})
}

func TestCombinedFilters(t *testing.T) {
	mw := NewFieldsFilterMiddleware(
		WithFields([]string{"a"}),
		WithRegexFields([]string{"^test[0-9]+$"}),
		WithFilters([]string{"test1"}),
		WithRegexFilters([]string{"^other[0-9]+$"}),
	)

	rows := createFieldsFilterTestRows()
	finalRows, err := processRows(mw, rows)
	require.NoError(t, err)
	require.Len(t, finalRows, 2)
	assert2.EqualRowValues(t, rows[0], map[string]interface{}{
		"a":     1,
		"test2": "value2",
	})
	assert2.EqualRowValues(t, rows[1], map[string]interface{}{
		"a":     4,
		"test3": "value3",
		"test4": "value4",
	})
}

func processRows(rm middlewares.RowMiddleware, rows []types.Row) ([]types.Row, error) {
	finalRows := make([]types.Row, 0)
	for _, row := range rows {
		rows_, err := rm.Process(context.Background(), row)
		if err != nil {
			return nil, err
		}
		finalRows = append(finalRows, rows_...)
	}

	return finalRows, nil
}
