package csv

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
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
	of.AddTableMiddleware(&table.RenameColumnMiddleware{Renames: renames})
	of.AddRow(&types.SimpleRow{Hash: types.NewMapRow(types.MRP("a", 1))})
	ctx := context.Background()
	buf := &bytes.Buffer{}
	err := of.Output(ctx, buf)
	require.NoError(t, err)

	_, data, err := csv.ParseCSV(strings.NewReader(buf.String()))
	require.NoError(t, err)
	require.Len(t, data, 1)
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
