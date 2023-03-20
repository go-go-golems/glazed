package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRemoveDuplicatesTables(rows [][]int) *types.Table {
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

func TestRemoveDuplicatesEmpty(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 0, len(newtable.Rows))
}

func TestRemoveDuplicatesSingle(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 1, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
}

func TestRemoveDuplicatesTwoDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{4, 5, 6},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 2, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	require.Equal(t, 5, row["b"])
	require.Equal(t, 6, row["c"])
}

func TestRemoveDuplicatesTwoSame(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{1, 2, 3},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 1, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
}

func TestRemoveDuplicatesTwoSameOneDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 2, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	require.Equal(t, 5, row["b"])
	require.Equal(t, 6, row["c"])
}

func TestRemoveDuplicatesTwoTimesTwoSame(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
		{4, 5, 6},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 2, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	require.Equal(t, 5, row["b"])
	require.Equal(t, 6, row["c"])
}

func TestRemoveDuplicatesTwoTimesTwoSameOneDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
		{4, 5, 6},
		{7, 8, 9},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 3, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
	row = newtable.Rows[1].GetValues()
	require.Equal(t, 4, row["a"])
	require.Equal(t, 5, row["b"])
	require.Equal(t, 6, row["c"])
	row = newtable.Rows[2].GetValues()
	require.Equal(t, 7, row["a"])
	require.Equal(t, 8, row["b"])
	require.Equal(t, 9, row["c"])
}

func TestRemoveDuplicatesTwoSameWithTwoColumns(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b")

	table := createRemoveDuplicatesTables([][]int{
		{1, 2, 3},
		{1, 2, 4},
	})
	newtable, err := removeDuplicatesMiddleware.Process(table)
	require.NoError(t, err)

	require.Equal(t, 1, len(newtable.Rows))
	row := newtable.Rows[0].GetValues()
	require.Equal(t, 1, row["a"])
	require.Equal(t, 2, row["b"])
	require.Equal(t, 3, row["c"])
}