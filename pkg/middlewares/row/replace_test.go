package row

import (
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func createReplaceTestRows() []types.Row {
	return []types.Row{
		types.NewRow(
			types.MRP("field1", "skip"),
			types.MRP("field2", "value2"),
		),
		types.NewRow(
			types.MRP("field1", "value1"),
			types.MRP("field2", "value3 blabla"),
		),
	}
}

func TestSingleSkip(t *testing.T) {
	replaceMiddleware := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{
			"field1": {
				&Skip{
					Pattern: "skip",
				},
			},
		},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(replaceMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "value1", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestSingleReplacement(t *testing.T) {
	replacementMiddleware := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{
			"field1": {
				&Replacement{
					Pattern:     "value1",
					Replacement: "replaced",
				},
			},
		},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(replacementMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")
	row = newRows[1]
	assert2.EqualMapRowValue(t, "replaced", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestTwoSkips(t *testing.T) {
	replaceMiddleware := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{
			"field1": {
				&Skip{
					Pattern: "skip",
				},
				&Skip{
					Pattern: "value1",
				},
			},
		},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(replaceMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 0, len(newRows))
}

func TestTwoColumnSkips(t *testing.T) {
	mw := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{
			"field1": {
				&Skip{
					Pattern: "skip",
				},
			},
			"field2": {
				&Skip{
					Pattern: "value3",
				},
			},
		},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(mw, rows)
	require.NoError(t, err)

	require.Equal(t, 0, len(newRows))

}

func TestSingleRegexpSkip(t *testing.T) {
	replaceMiddleware := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{
			"field1": {
				&RegexpSkip{
					Regexp: regexp.MustCompile("^s..p$"),
				},
			},
		},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(replaceMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "value1", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")

	replaceMiddleware.RegexSkips = map[types.FieldName][]*RegexpSkip{
		"field1": {
			&RegexpSkip{
				Regexp: regexp.MustCompile("kip$"),
			},
		},
	}

	newRows, err = processRows(replaceMiddleware, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))

	row = newRows[0]
	assert2.EqualMapRowValue(t, "value1", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestTwoReplacements(t *testing.T) {
	rep := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{
			"field1": {
				&Replacement{
					Pattern:     "val",
					Replacement: "replaced ",
				},
				&Replacement{
					Pattern:     "ue1",
					Replacement: "replaced2",
				},
			},
		},
		map[types.FieldName][]*RegexpReplacement{},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(rep, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")
	row = newRows[1]
	assert2.EqualMapRowValue(t, "replaced replaced2", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestSingleRegexpReplacement(t *testing.T) {
	rep := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{
			"field1": {
				&RegexpReplacement{
					Regexp:      regexp.MustCompile("^v.*1$"),
					Replacement: "replaced",
				},
			},
		},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(rep, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")

	row = newRows[1]
	assert2.EqualMapRowValue(t, "replaced", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestSingleRegexpCaptureReplacement(t *testing.T) {
	rep := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{
			"field1": {
				&RegexpReplacement{
					Regexp:      regexp.MustCompile("^v(.*)1$"),
					Replacement: "replaced$1",
				},
			},
		},
		map[types.FieldName][]*RegexpSkip{},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(rep, rows)
	require.NoError(t, err)

	require.Equal(t, 2, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "skip", row, "field1")
	assert2.EqualMapRowValue(t, "value2", row, "field2")

	row = newRows[1]
	assert2.EqualMapRowValue(t, "replacedalue", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestRegexpAndSkip(t *testing.T) {
	rep := NewReplaceMiddleware(
		map[types.FieldName][]*Replacement{},
		map[types.FieldName][]*RegexpReplacement{
			"field1": {
				&RegexpReplacement{
					Regexp:      regexp.MustCompile("^v.*1$"),
					Replacement: "replaced",
				},
			},
		},
		map[types.FieldName][]*RegexpSkip{
			"field1": {
				&RegexpSkip{
					Regexp: regexp.MustCompile("kip$"),
				},
			},
		},
		map[types.FieldName][]*Skip{},
	)

	rows := createReplaceTestRows()
	newRows, err := processRows(rep, rows)
	require.NoError(t, err)

	require.Equal(t, 1, len(newRows))

	row := newRows[0]
	assert2.EqualMapRowValue(t, "replaced", row, "field1")
	assert2.EqualMapRowValue(t, "value3 blabla", row, "field2")
}

func TestReplaceMiddlewareFromYAML(t *testing.T) {
	yaml := `
field1:
  replace:
    - p: v
    - p2: v2
  regex_replace:
    - ^v.*1$: replaced
  regex_skip:
    - kip$
  skip:
    - skip
field2:
  replace:
    - p: v
    - p2: v2
  regex_replace:
    - ^v.*2$: replaced
  regex_skip:
    - value2$
  skip:
    - value2
    - value3	
`
	rep, err := NewReplaceMiddlewareFromYAML([]byte(yaml))
	require.NoError(t, err)

	require.Equal(t, 2, len(rep.Replacements))
	require.Equal(t, 2, len(rep.RegexReplacements))
	require.Equal(t, 2, len(rep.RegexSkips))
	require.Equal(t, 2, len(rep.Skips))

	replacements := rep.Replacements["field1"]
	require.Equal(t, 2, len(replacements))
	assert.Equal(t, "p", replacements[0].Pattern)
	assert.Equal(t, "v", replacements[0].Replacement)
	assert.Equal(t, "p2", replacements[1].Pattern)
	assert.Equal(t, "v2", replacements[1].Replacement)

	regexpReplacements := rep.RegexReplacements["field1"]
	require.Equal(t, 1, len(regexpReplacements))

	assert.Equal(t, "^v.*1$", regexpReplacements[0].Regexp.String())
	assert.Equal(t, "replaced", regexpReplacements[0].Replacement)

	regexpSkips := rep.RegexSkips["field1"]
	require.Equal(t, 1, len(regexpSkips))
	assert.Equal(t, "kip$", regexpSkips[0].Regexp.String())

	skips := rep.Skips["field1"]
	require.Equal(t, 1, len(skips))
	assert.Equal(t, "skip", skips[0].Pattern)

	replacements = rep.Replacements["field2"]
	require.Equal(t, 2, len(replacements))
	assert.Equal(t, "p", replacements[0].Pattern)
	assert.Equal(t, "v", replacements[0].Replacement)
	assert.Equal(t, "p2", replacements[1].Pattern)
	assert.Equal(t, "v2", replacements[1].Replacement)

	regexpReplacements = rep.RegexReplacements["field2"]
	require.Equal(t, 1, len(regexpReplacements))
	assert.Equal(t, "^v.*2$", regexpReplacements[0].Regexp.String())
	assert.Equal(t, "replaced", regexpReplacements[0].Replacement)

	regexpSkips = rep.RegexSkips["field2"]
	require.Equal(t, 1, len(regexpSkips))
	assert.Equal(t, "value2$", regexpSkips[0].Regexp.String())

	skips = rep.Skips["field2"]
	require.Equal(t, 2, len(skips))
	assert.Equal(t, "value2", skips[0].Pattern)
	assert.Equal(t, "value3", skips[1].Pattern)
}
