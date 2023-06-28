package row

import (
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRemoveDuplicatesRows(rows [][]int) []types.Row {
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

	return ret.Rows
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 0, len(newRows))
}

func TestRemoveDuplicatesSingle(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
}

func TestRemoveDuplicatesTwoDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{4, 5, 6},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
	row = newRows[1]
	assert2.EqualRowValue(t, 4, row, "a")
	assert2.EqualRowValue(t, 5, row, "b")
	assert2.EqualRowValue(t, 6, row, "c")
}

func TestRemoveDuplicatesTwoSame(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{1, 2, 3},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
}

func TestRemoveDuplicatesTwoSameOneDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
	row = newRows[1]
	assert2.EqualRowValue(t, 4, row, "a")
	assert2.EqualRowValue(t, 5, row, "b")
	assert2.EqualRowValue(t, 6, row, "c")
}

func TestRemoveDuplicatesTwoTimesTwoSame(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
		{4, 5, 6},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
	row = newRows[1]
	assert2.EqualRowValue(t, 4, row, "a")
	assert2.EqualRowValue(t, 5, row, "b")
	assert2.EqualRowValue(t, 6, row, "c")
}

func TestRemoveDuplicatesTwoTimesTwoSameOneDifferent(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b", "c")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{1, 2, 3},
		{4, 5, 6},
		{4, 5, 6},
		{7, 8, 9},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 3, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
	row = newRows[1]
	assert2.EqualRowValue(t, 4, row, "a")
	assert2.EqualRowValue(t, 5, row, "b")
	assert2.EqualRowValue(t, 6, row, "c")
	row = newRows[2]
	assert2.EqualRowValue(t, 7, row, "a")
	assert2.EqualRowValue(t, 8, row, "b")
	assert2.EqualRowValue(t, 9, row, "c")
}

func TestRemoveDuplicatesTwoSameWithTwoColumns(t *testing.T) {
	removeDuplicatesMiddleware := NewRemoveDuplicatesMiddleware("a", "b")

	rows := createRemoveDuplicatesRows([][]int{
		{1, 2, 3},
		{1, 2, 4},
	})
	newRows, err := processRows(removeDuplicatesMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))
	row := newRows[0]
	assert2.EqualRowValue(t, 1, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
}
