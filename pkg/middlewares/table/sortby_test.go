package table

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createSortByTables(rows [][]interface{}) *types.Table {
	ret := types.NewTable()
	ret.Columns = []types.FieldName{"a", "b", "c"}
	ret.Rows = []types.Row{}
	for _, row := range rows {
		ret.Rows = append(ret.Rows,
			types.NewRow(
				types.MRP("a", row[0]),
				types.MRP("b", row[1]),
				types.MRP("c", row[2]),
			),
		)
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

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, 1, row, "a")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, 4, row, "a")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, 7, row, "a")
}

func TestSortByMiddlewareSingleIntColumnDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{7, 8, 9},
		{4, 5, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, 7, row, "a")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, 4, row, "a")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, 1, row, "a")
}

func TestSortByMiddlewareSingleStringColumn(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a")

	table := createSortByTables([][]interface{}{
		{"a", 2, 3},
		{"c", 8, 9},
		{"b", 5, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, "a", row, "a")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, "b", row, "a")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, "c", row, "a")
}

func TestSortByMiddlewareSingleStringColumnDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a")

	table := createSortByTables([][]interface{}{
		{"a", 2, 3},
		{"c", 8, 9},
		{"b", 5, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, "c", row, "a")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, "b", row, "a")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, "a", row, "a")

}

func TestSortByMiddlewareTwoColumns(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a", "b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 2, row, "b")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 3, row, "b")
	row = newtable.Rows[3]
	assert2.EqualMapRowValue(t, 2, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
}

func TestSortByMiddlewareTwoColumnsDesc(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a", "-b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, 2, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 3, row, "b")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 2, row, "b")
	row = newtable.Rows[3]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
}

func TestSortByMiddlewareTwoColumnsDescFirst(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("-a", "b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	row := newtable.Rows[0]
	assert2.EqualMapRowValue(t, 2, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
	row = newtable.Rows[1]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 1, row, "b")
	row = newtable.Rows[2]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 2, row, "b")
	row = newtable.Rows[3]
	assert2.EqualMapRowValue(t, 1, row, "a")
	assert2.EqualMapRowValue(t, 3, row, "b")
}

func TestSortByMiddlewareTwoColumnsDescSecond(t *testing.T) {
	sortByMiddleware := NewSortByMiddlewareFromColumns("a", "-b")

	table := createSortByTables([][]interface{}{
		{1, 2, 3},
		{1, 1, 9},
		{1, 3, 6},
		{2, 1, 6},
	})

	newtable, err := sortByMiddleware.Process(context.Background(), table)
	require.NoError(t, err)

	require.Equal(t, 4, len(newtable.Rows))
	// after that, the order is undefined
}
