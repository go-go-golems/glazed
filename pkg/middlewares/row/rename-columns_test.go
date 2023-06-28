package row

import (
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
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
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
}

func TestRenameTwoFieldColumns(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "bar",
		"baz": "qux",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "qux")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
}

func TestRenameOverrideColumn(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foo": "foobar",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	// we get 3 because our keys are ordered now, and foobar comes last, thus overwriting the renamed foo
	assert2.EqualMapRowValue(t, 3, row, "foobar")
	assert2.EqualMapRowValue(t, 2, row, "baz")

}

func TestRenameOverrideColumnLast(t *testing.T) {
	renameTable := map[types.FieldName]types.FieldName{
		"foobar": "foo",
	}
	mw := NewFieldRenameColumnMiddleware(renameTable)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foobar")
	// the renamed foobar -> foo now overwrites the original foo value
	assert2.EqualMapRowValue(t, 3, row, "foo")
	assert2.EqualMapRowValue(t, 2, row, "baz")

}

func TestRenameRegexpSimpleMatch(t *testing.T) {
	rrs := RegexpReplacements{}

	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo$"), "bar",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
}

func TestRenameRegexpDoubleMatch(t *testing.T) {
	rrs := RegexpReplacements{}
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("f.."), "bar",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	// here, f.. should match both fields
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
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
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	row := newRows[0]
	// it's going to be hard to test that these will happen in the right
	// order as it really depends on the map
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "bar2")
	assert2.NilMapRowValue(t, row, "bar3")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
}

func createTestRows() []types.Row {
	table := types.Table{
		Columns: []types.FieldName{
			"foo",
			"baz",
			"foobar",
		},
		Rows: []types.Row{},
	}
	table.Rows = append(table.Rows,
		types.NewRow(
			types.MRP("foo", 1),
			types.MRP("baz", 2),
			types.MRP("foobar", 3),
		),
	)

	return table.Rows
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
	rows := createTestRows()

	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "baz")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "qux")
	assert2.EqualMapRowValue(t, 3, row, "bar2bar")
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

	rows := createTestRows()

	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "baz")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "qux")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
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

	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "baz")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar2")
	assert2.EqualMapRowValue(t, 2, row, "qux")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
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

	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "baz")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "qux")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
}

func TestRegexpCaptureGroupRename(t *testing.T) {
	rrs := RegexpReplacements{}
	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo(.*)$"), "bar$1",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
}

func TestRegexpCaptureGroupRenameFromYAML(t *testing.T) {
	yamlString := `
regexpRenames:
  "^foo(.*)$": bar$1
`

	decoder := yaml.NewDecoder(strings.NewReader(yamlString))
	mw, err := NewRenameColumnMiddlewareFromYAML(decoder)
	require.Nil(t, err)

	rows := createTestRows()
	newRows, err := processRows(mw, rows)
	require.Nil(t, err)

	row := newRows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
}
