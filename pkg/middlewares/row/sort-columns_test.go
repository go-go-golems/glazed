package row

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleSortColumns(t *testing.T) {
	row := types.NewRow(
		types.MRP("c", "value3"),
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
	)

	mw := NewSortColumnsMiddleware()
	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)

	require.Len(t, newRows, 1)
	assert2.EqualRow(t, types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}
