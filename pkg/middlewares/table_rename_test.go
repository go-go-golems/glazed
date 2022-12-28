package middlewares

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wesen/glazed/pkg/types"
	"regexp"
	"testing"
)

func TestSingleRename(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "bar",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	table := createTestTable()

	newTable, ret := mw.Process(table)

	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["foobar"])
}

func TestRenameTwoFieldColumns(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "bar",
		"baz": "qux",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["qux"])
	assert.Equal(t, 3, rowMap["foobar"])
}

func TestRenameOverrideColumn(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "foobar",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Equal(t, 1, rowMap["foobar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Nil(t, rowMap["foo"])

}

func TestRenameRegexpSimpleMatch(t *testing.T) {
	regexpTable := map[*regexp.Regexp]string{
		regexp.MustCompile("^f..$"): "bar",
	}
	mw := NewRegexpRenameColumnMiddleware(regexpTable)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["foobar"])
}

func TestRenameRegexpDoubleMatch(t *testing.T) {
	regexpTable := map[*regexp.Regexp]string{
		regexp.MustCompile("f.."): "bar",
	}
	mw := NewRegexpRenameColumnMiddleware(regexpTable)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	// here, f.. should match both fields
	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["barbar"])

}

func createTestTable() *types.Table {
	table := types.Table{
		Columns: []types.FieldName{
			"foo",
			"baz",
			"foobar",
		},
		Rows: []types.Row{},
	}
	table.Rows = append(table.Rows, &types.SimpleRow{
		Hash: types.MapRow{
			"foo":    1,
			"baz":    2,
			"foobar": 3,
		},
	})

	return &table
}
