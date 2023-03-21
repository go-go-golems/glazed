package table

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createSortByTables(rows [][]interface{}) *types.Table {
	ret := types.NewTable()
	ret.Columns = []types.FieldName{"a", "b", "c"}
	ret.Rows = []types.Row{}
	for _, row := range rows {
		ret.Rows = append(ret.Rows, &types.SimpleRow{
			Hash: map[string]interface{}{
				"a": row[0],
				"b": row[1],
				"c": row[2],
			},
		})
	}

	return ret
}

func TestSortByMiddlewareSingleIntColumn(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{7, 8, 9},
		{4, 5, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 7, row["a"])
}

func TestSortByMiddlewareSingleIntColumnDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{7, 8, 9},
		{4, 5, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 7, row["a"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 1, row["a"])
}

func TestSortByMiddlewareSingleStringColumn(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a")

	table := createSortByTables([][]interface{}{
		{"a", 2, 3},
		{"c", 8, 9},
		{"b", 5, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, "a", row["a"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, "b", row["a"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, "c", row["a"])
}

func TestSortByMiddlewareSingleStringColumnDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a")

	table := createSortByTables([][]interface{}{
		{"a", 2, 3},
		{"c", 8, 9},
		{"b", 5, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, "c", row["a"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, "b", row["a"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, "a", row["a"])

}

func TestSortByMiddlewareTwoColumns(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a", "b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 1, row["b"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 3, row["b"])
	row = newtable.Rows[3].GetValues()
	require.Equal(t, 2, row["a"])
	require.Equal(t, 1, row["b"])
}

func TestSortByMiddlewareTwoColumnsDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a", "-b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 2, row["a"])
	require.Equal(t, 1, row["b"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 3, row["b"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	row = newtable.Rows[3].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 1, row["b"])
}

func TestSortByMiddlewareTwoColumnsDescFirst(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a", "b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 2, row["a"])
	require.Equal(t, 1, row["b"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 1, row["b"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	row = newtable.Rows[3].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 3, row["b"])
}

func TestSortByMiddlewareTwoColumnsDescSecond(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a", "-b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	// after that, the order is undefined
}
