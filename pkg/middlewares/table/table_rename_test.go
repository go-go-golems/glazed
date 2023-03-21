package table

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
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
	rrs := RegexpReplacements{}

	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo$"), "bar",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
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
	rrs := RegexpReplacements{}
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("f.."), "bar",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
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

func TestRenameRegexpOrderedMatch(t *testing.T) {
	rrs := RegexpReplacements{}

	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("f..$"), "bar",
	})
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo$"), "bar2",
	})
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo$"), "bar3",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	// it's going to be hard to test that these will happen in the right
	// order as it really depends on the map
	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["bar2"])
	assert.Nil(t, rowMap["bar3"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["foobar"])
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

func TestBothFieldAndRegexpRenames(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "bar",
		"baz": "qux",
	}
	rrs := RegexpReplacements{}
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("f.."), "bar2",
	})

	mw := NewRenameColumnMiddleware(renameTable, rrs)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["baz"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["qux"])
	assert.Equal(t, 3, rowMap["bar2bar"])
}

func TestParseFromYAML(t *testing.T) {
	yamlString := `
renames:
  foo: bar
  baz: qux
`
	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	mw, err := NewRenameColumnMiddlewareFromYAML(decoder)
	require.Nil(t, err)

	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["baz"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["qux"])
	assert.Equal(t, 3, rowMap["foobar"])
}

func TestParseRegexpFromYAML(t *testing.T) {
	yamlString := `
regexpRenames:
  "^foo$": bar2
  f..: bar
  "^baz$": qux
`

	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	mw, err := NewRenameColumnMiddlewareFromYAML(decoder)
	require.Nil(t, err)

	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["baz"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar2"])
	assert.Equal(t, 2, rowMap["qux"])
	assert.Equal(t, 3, rowMap["barbar"])
}

func TestParseBothFromYAML(t *testing.T) {
	yamlString := `
renames:
  foo: bar
  baz: qux
regexpRenames:
  "^foo$": bar2
  f..: bar
  "^baz$": qux2
`

	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	mw, err := NewRenameColumnMiddlewareFromYAML(decoder)
	require.Nil(t, err)

	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["baz"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["qux"])
	assert.Equal(t, 3, rowMap["barbar"])
}

func TestRegexpCaptureGroupRename(t *testing.T) {
	rrs := RegexpReplacements{}
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo(.*)$"), "bar$1",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["barbar"])
}

func TestRegexpCaptureGroupRenameFromYAML(t *testing.T) {
	yamlString := `
regexpRenames:
  "^foo(.*)$": bar$1
`

	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	mw, err := NewRenameColumnMiddlewareFromYAML(decoder)
	require.Nil(t, err)

	table := createTestTable()

	newTable, ret := mw.Process(table)
	require.Nil(t, ret)

	row := newTable.Rows[0].(*types.SimpleRow)
	rowMap := row.GetValues()

	assert.Nil(t, rowMap["foo"])
	assert.Nil(t, rowMap["foobar"])
	assert.Equal(t, 1, rowMap["bar"])
	assert.Equal(t, 2, rowMap["baz"])
	assert.Equal(t, 3, rowMap["barbar"])
}
