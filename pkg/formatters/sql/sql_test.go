package sql

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOutputFormatter_EmptyTableName(t *testing.T) {
	f := NewOutputFormatter()
	if f.TableName != "output" {
		t.Errorf("Expected default table name to be 'output', got '%s'", f.TableName)
	}

}
func TestOutputFormatter_WithTableName(t *testing.T) {
	f := NewOutputFormatter(WithTableName("foo"))
	assert.Equal(t, "foo", f.TableName)
}

func TestOutputFormatter_NoRows(t *testing.T) {
	f := NewOutputFormatter()

	var b bytes.Buffer
	ctx := context.Background()
	err := f.Close(ctx, &b)
	assert.NoError(t, err)

	assert.Equal(t, "", b.String())
}

func runFormatter(f *OutputFormatter, rows []types.Row) (string, error) {
	var b bytes.Buffer
	ctx := context.Background()
	for _, row := range rows {
		err := f.OutputRow(ctx, row, &b)
		if err != nil {
			return "", err
		}
	}
	err := f.Close(ctx, &b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func TestOutputFormatter_SingleRow(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", "bar"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, "INSERT INTO output (foo) VALUES\n('bar')\n;\n", s)
}

func TestOutputFormatter_WithUseUpsert(t *testing.T) {
	f := NewOutputFormatter(WithUseUpsert(true))
	assert.Equal(t, true, f.UseUpsert)

	row := types.NewRow(
		types.MRP("foo", "bar"),
	)

	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, "INSERT INTO output (foo) VALUES\n('bar')\nON DUPLICATE KEY UPDATE\nfoo = VALUES(foo);\n", s)
}

func TestOutputFormatter_WithUseUpsertNoRows(t *testing.T) {
	f := NewOutputFormatter(WithUseUpsert(true))
	assert.Equal(t, true, f.UseUpsert)

	s, err := runFormatter(f, []types.Row{})
	assert.NoError(t, err)

	assert.Equal(t, "", s)
}

func TestOutputFormatter_SingleRowMultipleColumns(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, "INSERT INTO output (foo, baz) VALUES\n('bar', 'qux')\n;\n", s)
}

func TestOutputFormatter_SingleRowMultipleColumnsWithUpsert(t *testing.T) {
	f := NewOutputFormatter(WithUseUpsert(true))

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, "INSERT INTO output (foo, baz) VALUES\n('bar', 'qux')\nON DUPLICATE KEY UPDATE\nfoo = VALUES(foo),\nbaz = VALUES(baz);\n", s)
}

func TestOutputFormatter_WithSplitByRows_NoSplit(t *testing.T) {
	f := NewOutputFormatter(WithSplitByRows(2))

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row, row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', 'qux')
;
`, s)
}

func TestOutputFormatter_WithSplitByRows_SplitInTheMiddle(t *testing.T) {
	f := NewOutputFormatter(WithSplitByRows(2))

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row, row, row, row, row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', 'qux')
;
INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', 'qux')
;
INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
;
`, s)
}

func TestOutputFormatter_WithSplitByRows_SplitAtTheEnd(t *testing.T) {
	f := NewOutputFormatter(WithSplitByRows(2))

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row, row, row, row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', 'qux')
;
INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', 'qux')
;
`, s)
}

func TestOutputFormatter_EmptyRow(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow()
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, "", s)
}

// Test that when rows with different columns are output, we only output the columns of the first row,
// filling missing columns with NULL.
func TestOutputFormatter_RowsWithDifferentShapes(t *testing.T) {
	f := NewOutputFormatter()

	row1 := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux"),
	)
	row2 := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("quux", "corge"),
	)
	s, err := runFormatter(f, []types.Row{row1, row2})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar', 'qux')
, ('bar', NULL)
;
`, s)
}

func TestOutputFormatter_SerializeRowWithNulls(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", nil),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
(NULL, 'qux')
;
`, s)
}

func TestOutputFormatter_SerializeRowWithDifferentDatatypes(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", 1),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
(1, 'qux')
;
`, s)
}

func TestOutputFormatter_SerializeRowWithSpecialChars(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", "bar"),
		types.MRP("baz", "qux\n"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar', 'qux
')
;
`, s)
}

func TestOutputFormatter_SerializeSqlEscape(t *testing.T) {
	f := NewOutputFormatter()

	row := types.NewRow(
		types.MRP("foo", "bar'"),
		types.MRP("baz", "qux"),
	)
	s, err := runFormatter(f, []types.Row{row})
	assert.NoError(t, err)

	assert.Equal(t, `INSERT INTO output (foo, baz) VALUES
('bar''', 'qux')
;
`, s)
}
