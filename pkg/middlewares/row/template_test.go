package row

import (
	"context"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleTemplate(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	)

	mw, err := NewTemplateMiddleware(
		map[types.FieldName]string{
			"a": "{{.a}}-{{.b}}",
		}, "")
	require.NoError(t, err)

	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)
	require.Len(t, newRows, 1)

	assert2.EqualRow(t, types.NewRow(
		types.MRP("a", "value1-value2"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}

func TestSimpleDoubleTemplate(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	)

	mw, err := NewTemplateMiddleware(
		map[types.FieldName]string{
			"a": "{{.a}}-{{.b}}",
			"b": "{{.a}}-{{.b}}",
		}, "")
	require.NoError(t, err)

	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)
	require.Len(t, newRows, 1)

	assert2.EqualRow(t, types.NewRow(
		types.MRP("a", "value1-value2"),
		types.MRP("b", "value1-value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}

func TestSimpleDoubleTemplateDifferentOrder(t *testing.T) {
	row := types.NewRow(
		types.MRP("b", "value2"),
		types.MRP("a", "value1"),
		types.MRP("c", "value3"),
	)

	mw, err := NewTemplateMiddleware(
		map[types.FieldName]string{
			"b": "{{.a}}-{{.b}}",
			"a": "{{.a}}-{{.b}}",
		}, "")
	require.NoError(t, err)

	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)
	require.Len(t, newRows, 1)

	assert2.EqualRow(t, types.NewRow(
		types.MRP("b", "value1-value2"),
		types.MRP("a", "value1-value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}

func TestSimpleDoubleTemplateWithDot(t *testing.T) {
	row := types.NewRow(
		types.MRP("b.d", "value2"),
		types.MRP("a", "value1"),
		types.MRP("c", "value3"),
	)

	mw, err := NewTemplateMiddleware(
		map[types.FieldName]string{
			"b.d": "{{.a}}-{{.b_d}}",
			"a":   "{{.a}}-{{.b_d}}",
		}, "_")
	require.NoError(t, err)

	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)
	require.Len(t, newRows, 1)

	assert2.EqualRow(t, types.NewRow(
		types.MRP("b.d", "value1-value2"),
		types.MRP("a", "value1-value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}

func TestSimpleTemplateRow(t *testing.T) {
	row := types.NewRow(
		types.MRP("a", "value1"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	)

	mw, err := NewTemplateMiddleware(
		map[types.FieldName]string{
			"a": "{{.a}}-{{._row.b}}",
		}, "")
	require.NoError(t, err)

	newRows, err := mw.Process(context.Background(), row)
	require.NoError(t, err)
	require.Len(t, newRows, 1)

	assert2.EqualRow(t, types.NewRow(
		types.MRP("a", "value1-value2"),
		types.MRP("b", "value2"),
		types.MRP("c", "value3"),
	), newRows[0])
}
