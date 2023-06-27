package csv

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
	p_ := middlewares.NewProcessor(middlewares.WithTableMiddleware(&table.RenameColumnMiddleware{
		Renames: renames,
	}))
	ctx := context.Background()
	err := p_.AddRow(ctx, &types.SimpleRow{Hash: types.NewMapRow(types.MRP("a", 1))})
	require.NoError(t, err)

	err = p_.FinalizeTable(ctx)
	require.NoError(t, err)
	table_ := p_.GetTable()

	buf := &bytes.Buffer{}
	err = of.Output(ctx, table_, buf)
	require.NoError(t, err)

	_, data, err := csv.ParseCSV(strings.NewReader(buf.String()))
	require.NoError(t, err)

	require.Len(t, data, 1)
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
