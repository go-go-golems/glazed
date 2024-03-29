package middlewares

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func createJqObjectMiddleware(t *testing.T, e string) *JqObjectMiddleware {
	ret, err := NewJqObjectMiddleware(e)
	require.NoError(t, err)

	return ret
}

func createJqTableMiddleware(t *testing.T, m map[types.FieldName]string) *JqTableMiddleware {
	ret, err := NewJqTableMiddleware(m)
	require.NoError(t, err)

	return ret
}

func createJqTestTable() *types.Table {
	return &types.Table{
		Columns: []types.FieldName{},
		Rows: []types.Row{
			types.NewRow(
				types.MRP("a", 1),
				types.MRP("b", 2),
				types.MRP("c", map[string]interface{}{
					"d": 3,
				}),
				types.MRP("e", "hello"),
				types.MRP("f", []interface{}{1, 2, 3}),
			),
			types.NewRow(
				types.MRP("a", 11),
				types.MRP("c", map[string]interface{}{
					"d": 13,
					"e": 12,
				}),
				types.MRP("e", "foobar"),
				types.MRP("f", []interface{}{1, 4, 2, 3}),
			),
		},
	}
}

func TestEmptyObjectMiddleware(t *testing.T) {
	m := createJqObjectMiddleware(t, "")
	require.Nil(t, m.query)

	ctx := context.Background()
	obj := types.NewRow(types.MRP("a", 1))
	o2, err := m.Process(ctx, obj)
	require.NoError(t, err)
	assert.Len(t, o2, 1)
	assert2.EqualRow(t, obj, o2[0])
}

func TestSimpleJqConstant(t *testing.T) {
	m := createJqObjectMiddleware(t, "{a: 2}")
	require.NotNil(t, m.query)

	ctx := context.Background()
	o := types.NewRow(types.MRP("a", 1))
	o2, err := m.Process(ctx, o)
	require.NoError(t, err)
	assert.Len(t, o2, 1)
	expected := types.NewRow(types.MRP("a", 2))
	assert2.EqualRow(t, expected, o2[0])
}

func TestSimpleJqConstantArray(t *testing.T) {
	m := createJqObjectMiddleware(t, "{a: [2]}")
	require.NotNil(t, m.query)

	ctx := context.Background()
	o := types.NewRow(types.MRP("a", 1))
	o2, err := m.Process(ctx, o)
	require.NoError(t, err)
	expected := types.NewRow(types.MRP("a", []interface{}{2}))
	assert.Len(t, o2, 1)
	assert2.EqualRow(t, expected, o2[0])
}

func TestSimpleJqExtractNestedArray(t *testing.T) {
	m := createJqObjectMiddleware(t, ".a[0]")
	require.NotNil(t, m.query)

	ctx := context.Background()
	o := types.NewRow(types.MRP("a", []interface{}{map[string]interface{}{"b": 2}}))
	o2, err := m.Process(ctx, o)
	require.NoError(t, err)
	expected := types.NewRow(types.MRP("b", 2))
	assert.Len(t, o2, 1)
	assert2.EqualRow(t, expected, o2[0])
}

func TestSimpleJqExtract(t *testing.T) {
	m := createJqObjectMiddleware(t, ".f | map({field: .})")
	require.NotNil(t, m.query)

	ctx := context.Background()
	table := createJqTestTable()
	v2 := table.Rows[0]
	o2, err := m.Process(ctx, v2)
	require.NoError(t, err)
	assert.Len(t, o2, 3)
	assert2.EqualRowMap(t, map[string]interface{}{"field": 1}, o2[0])
	assert2.EqualRowMap(t, map[string]interface{}{"field": 2}, o2[1])
	assert2.EqualRowMap(t, map[string]interface{}{"field": 3}, o2[2])
}

func TestSimpleJqTableConstant(t *testing.T) {
	m := createJqTableMiddleware(t, map[types.FieldName]string{"a": "2"})
	require.NotNil(t, m)
	require.NotEmpty(t, m.fieldQueries)

	ctx := context.Background()
	table := createJqTestTable()
	t2, err := m.Process(ctx, table)
	require.NoError(t, err)

	row := t2.Rows[0]
	v, ok := row.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 2, v)
	assert2.EqualRowValue(t, 2, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, map[string]interface{}{"d": 3}, row, "c")
	assert2.EqualRowValue(t, "hello", row, "e")
	assert2.EqualRowValue(t, []interface{}{1, 2, 3}, row, "f")
}

func TestSimpleJqTableTwoFields(t *testing.T) {
	m := createJqTableMiddleware(t, map[types.FieldName]string{"a": "2", "c": ".d"})
	require.NotNil(t, m)
	require.NotEmpty(t, m.fieldQueries)

	ctx := context.Background()
	table := createJqTestTable()
	t2, err := m.Process(ctx, table)
	require.NoError(t, err)

	row := t2.Rows[0]
	assert2.EqualRowValue(t, 2, row, "a")
	assert2.EqualRowValue(t, 2, row, "b")
	assert2.EqualRowValue(t, 3, row, "c")
	assert2.EqualRowValue(t, "hello", row, "e")
	assert2.EqualRowValue(t, []interface{}{1, 2, 3}, row, "f")
}
