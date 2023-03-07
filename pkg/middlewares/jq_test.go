package middlewares

import (
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
			&types.SimpleRow{
				Hash: map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": map[string]interface{}{
						"d": 3,
					},
					"e": "hello",
					"f": []interface{}{1, 2, 3},
				},
			},
			&types.SimpleRow{
				Hash: map[string]interface{}{
					"a": 11,
					"c": map[string]interface{}{
						"d": 13,
						"e": 12,
					},
					"e": "foobar",
					"f": []interface{}{1, 4, 2, 3},
				},
			},
		},
	}
}

func TestEmptyObjectMiddleware(t *testing.T) {
	m := createJqObjectMiddleware(t, "")
	require.Nil(t, m.query)

	o := map[string]interface{}{"a": 1}
	o2, err := m.Process(o)
	assert.NoError(t, err)
	assert.Len(t, o2, 1)
	assert.Equal(t, o, o2[0])
}

func TestSimpleJqConstant(t *testing.T) {
	m := createJqObjectMiddleware(t, "{a: 2}")
	require.NotNil(t, m.query)

	o := map[string]interface{}{"a": 1}
	o2, err := m.Process(o)
	assert.NoError(t, err)
	assert.Len(t, o2, 1)
	expected := map[string]interface{}{"a": 2}
	assert.Equal(t, expected, o2[0])
}

func TestSimpleJqConstantArray(t *testing.T) {
	m := createJqObjectMiddleware(t, "{a: [2]}")
	require.NotNil(t, m.query)

	o := map[string]interface{}{"a": 1}
	o2, err := m.Process(o)
	assert.NoError(t, err)
	expected := map[string]interface{}{"a": []interface{}{2}}
	assert.Len(t, o2, 1)
	assert.Equal(t, expected, o2[0])
}

func TestSimpleJqExtractNestedArray(t *testing.T) {
	m := createJqObjectMiddleware(t, ".a[0]")
	require.NotNil(t, m.query)

	o := map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": 2}}}
	o2, err := m.Process(o)
	assert.NoError(t, err)
	expected := map[string]interface{}{"b": 2}
	assert.Len(t, o2, 1)
	assert.Equal(t, expected, o2[0])
}

func TestSimpleJqExtract(t *testing.T) {
	m := createJqObjectMiddleware(t, ".f | map({field: .})")
	require.NotNil(t, m.query)

	table := createJqTestTable()
	v2 := table.Rows[0].GetValues()
	o2, err := m.Process(v2)
	assert.NoError(t, err)
	assert.Len(t, o2, 3)
	assert.Equal(t, map[string]interface{}{"field": 1}, o2[0])
	assert.Equal(t, map[string]interface{}{"field": 2}, o2[1])
	assert.Equal(t, map[string]interface{}{"field": 3}, o2[2])
}

func TestSimpleJqTableConstant(t *testing.T) {
	m := createJqTableMiddleware(t, map[types.FieldName]string{"a": "2"})
	require.NotNil(t, m)
	require.NotEmpty(t, m.fieldQueries)

	table := createJqTestTable()
	t2, err := m.Process(table)
	assert.NoError(t, err)

	row := t2.Rows[0].GetValues()
	assert.Equal(t, 2, row["a"])
	assert.Equal(t, 2, row["b"])
	assert.Equal(t, map[string]interface{}{"d": 3}, row["c"])

	assert.Equal(t, "hello", row["e"])
	assert.Equal(t, []interface{}{1, 2, 3}, row["f"])
}

func TestSimpleJqTableTwoFields(t *testing.T) {
	m := createJqTableMiddleware(t, map[types.FieldName]string{"a": "2", "c": ".d"})
	require.NotNil(t, m)
	require.NotEmpty(t, m.fieldQueries)

	table := createJqTestTable()
	t2, err := m.Process(table)
	assert.NoError(t, err)

	row := t2.Rows[0].GetValues()
	assert.Equal(t, 2, row["a"])
	assert.Equal(t, 2, row["b"])
	assert.Equal(t, 3, row["c"])

	assert.Equal(t, "hello", row["e"])
	assert.Equal(t, []interface{}{1, 2, 3}, row["f"])
}
