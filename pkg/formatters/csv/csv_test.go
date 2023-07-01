package csv

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestCSVRenameEndToEnd(t *testing.T) {
	of := NewCSVOutputFormatter()
	renames := map[string]string{
		"a": "b",
	}

	buf := &bytes.Buffer{}
	p_ := middlewares.NewTableProcessor(
		middlewares.WithRowMiddleware(row.NewFieldRenameColumnMiddleware(renames)),
		middlewares.WithTableMiddleware(table.NewOutputMiddleware(of, buf)),
	)
	ctx := context.Background()
	err := p_.AddRow(ctx, types.NewRow(types.MRP("a", 1)))
	require.NoError(t, err)

	err = p_.Close(ctx)
	require.NoError(t, err)

	_, data, err := csv.ParseCSV(strings.NewReader(buf.String()))
	require.NoError(t, err)

	require.Len(t, data, 1)
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
