package table

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createAddFieldTestTable() *types.Table {
	ret := types.NewTable()
	ret.Columns = []types.FieldName{"field1", "field2"}
	ret.Rows = []types.Row{
		&types.SimpleRow{
			Hash: map[types.FieldName]types.GenericCellValue{
				"field1": "skip",
				"field2": "value2",
			},
		},
		&types.SimpleRow{
			Hash: map[types.FieldName]types.GenericCellValue{
				"field1": "value1",
				"field2": "value3 blabla",
			},
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

	row := newtable.Rows[0]
	require.Equal(t, "skip", row.GetValues()["field1"])
	require.Equal(t, "value2", row.GetValues()["field2"])
	require.Equal(t, "value3", row.GetValues()["field3"])

	row = newtable.Rows[1]
	require.Equal(t, "value1", row.GetValues()["field1"])
	require.Equal(t, "value3 blabla", row.GetValues()["field2"])
	require.Equal(t, "value3", row.GetValues()["field3"])
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

	row := newtable.Rows[0]
	require.Equal(t, "skip", row.GetValues()["field1"])
	require.Equal(t, "value2", row.GetValues()["field2"])
	require.Equal(t, "value3", row.GetValues()["field3"])
	require.Equal(t, "value4", row.GetValues()["field4"])

	row = newtable.Rows[1]
	require.Equal(t, "value1", row.GetValues()["field1"])
	require.Equal(t, "value3 blabla", row.GetValues()["field2"])
	require.Equal(t, "value3", row.GetValues()["field3"])
	require.Equal(t, "value4", row.GetValues()["field4"])
}
