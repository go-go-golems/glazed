package table

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTableRenameEndToEnd(t *testing.T) {
	of := NewOutputFormatter("markdown")
	renames := map[string]string{
		"a": "b",
	}

	obj := types.NewMapRow(types.MRP("a", 1))
	ctx := context.Background()

	p_ := middlewares.NewProcessor(middlewares.WithTableMiddleware(&table.RenameColumnMiddleware{Renames: renames}))
	err := p_.AddRow(ctx, &types.SimpleRow{Hash: obj})
	require.NoError(t, err)
	err = p_.FinalizeTable(ctx)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	err = of.Output(ctx, p_.GetTable(), buf)
	require.NoError(t, err)

	// parse s
	assert.Equal(t, "| b |\n| --- |\n| 1 |", buf.String())
}
