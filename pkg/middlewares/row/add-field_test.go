package row

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createAddFieldTestRows() []types.Row {
	return []types.Row{
		types.NewRow(
			types.MRP("field1", "skip"),
			types.MRP("field2", "value2"),
		),
		types.NewRow(
			types.MRP("field1", "value1"),
			types.MRP("field2", "value3 blabla"),
		),
	}
}

func TestSingleAddField(t *testing.T) {
	addFieldMiddleware := NewAddFieldMiddleware(map[string]string{
		"field3": "value3",
	})

	rows := createAddFieldTestRows()
	finalRows, err := processRows(addFieldMiddleware, rows)
	require.NoError(t, err)
	require.Equal(t, 2, len(finalRows))

	row := finalRows[0]
	assert2.EqualRowValue(t, "skip", row, "field1")
	assert2.EqualRowValue(t, "value2", row, "field2")
	assert2.EqualRowValue(t, "value3", row, "field3")

	row = finalRows[1]
	assert2.EqualRowValue(t, "value1", row, "field1")
	assert2.EqualRowValue(t, "value3 blabla", row, "field2")
	assert2.EqualRowValue(t, "value3", row, "field3")
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

func TestMultipleAddField(t *testing.T) {
	addFieldMiddleware := NewAddFieldMiddleware(map[string]string{
		"field3": "value3",
		"field4": "value4",
	})

	rows := createAddFieldTestRows()
	finalRows, err := processRows(addFieldMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(finalRows))

	row := finalRows[0]
	assert2.EqualRowValue(t, "skip", row, "field1")
	assert2.EqualRowValue(t, "value2", row, "field2")
	assert2.EqualRowValue(t, "value3", row, "field3")
	assert2.EqualRowValue(t, "value4", row, "field4")

	row = finalRows[1]
	assert2.EqualRowValue(t, "value1", row, "field1")
	assert2.EqualRowValue(t, "value3 blabla", row, "field2")
	assert2.EqualRowValue(t, "value3", row, "field3")
	assert2.EqualRowValue(t, "value4", row, "field4")
}
