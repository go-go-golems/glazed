package table

import (
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createAddFieldTestTable() *types.Table {
	ret := types.NewTable()
	ret.Columns = []types.FieldName{"field1", "field2"}
	ret.Rows = []types.Row{
		&types.SimpleRow{
			Hash: types.NewMapRow(
				types.MRP("field1", "skip"),
				types.MRP("field2", "value2"),
			),
		},
		&types.SimpleRow{
			Hash: types.NewMapRow(
				types.MRP("field1", "value1"),
				types.MRP("field2", "value3 blabla"),
			),
		},
	}

	return ret
}

func TestSingleAddField(t *testing.T) {
	addFieldMiddleware := NewAddFieldMiddleware(map[string]string{
		"field3": "value3",
	})

	table := createAddFieldTestTable()
	newtable, err := addFieldMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 2, len(newtable.Rows))

	row := newtable.Rows[0].GetValues()
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")
	assert2.EqualMapRowValue(t, "value3", row, "field3")

	row = newtable.Rows[1].GetValues()
	assert2.EqualMapRowValue(t, "value1", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
	assert2.EqualMapRowValue(t, "value3", row, "field3")
}

func TestMultipleAddField(t *testing.T) {
	addFieldMiddleware := NewAddFieldMiddleware(map[string]string{
		"field3": "value3",
		"field4": "value4",
	})

	table := createAddFieldTestTable()
	newtable, err := addFieldMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 2, len(newtable.Rows))

	row := newtable.Rows[0].GetValues()
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")
	assert2.EqualMapRowValue(t, "value3", row, "field3")
	assert2.EqualMapRowValue(t, "value4", row, "field4")

	row = newtable.Rows[1].GetValues()
	assert2.EqualMapRowValue(t, "value1", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
	assert2.EqualMapRowValue(t, "value3", row, "field3")
	assert2.EqualMapRowValue(t, "value4", row, "field4")
}
