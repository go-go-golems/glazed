package table

import (
	"context"
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)

	require.Nil(t, ret)

	row := newTable.Rows[0]
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.EqualMapRowValue(t, 1, row, "foobar")
	assert2.EqualMapRowValue(t, 2, row, "baz")

}

func TestRenameRegexpSimpleMatch(t *testing.T) {
	rrs := RegexpReplacements{}

	rrs = append(rrs, &RegexpReplacement{
		regexp.MustCompile("^foo$"), "bar",
	})

	mw := NewRegexpRenameColumnMiddleware(rrs)
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
	// it's going to be hard to test that these will happen in the right
	// order as it really depends on the map
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "bar2")
	assert2.NilMapRowValue(t, row, "bar3")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "foobar")
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
	table.Rows = append(table.Rows,
		types.NewMapRow(
			types.MRP("foo", 1),
			types.MRP("baz", 2),
			types.MRP("foobar", 3),
		),
	)

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

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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

	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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

	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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

	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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
	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
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

	table := createTestTable()

	newTable, ret := mw.Process(context.Background(), table)
	require.Nil(t, ret)

	row := newTable.Rows[0]
	assert2.NilMapRowValue(t, row, "foo")
	assert2.NilMapRowValue(t, row, "foobar")
	assert2.EqualMapRowValue(t, 1, row, "bar")
	assert2.EqualMapRowValue(t, 2, row, "baz")
	assert2.EqualMapRowValue(t, 3, row, "barbar")
}
