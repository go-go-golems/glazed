package table

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
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

	obj := types.NewRow(types.MRP("a", 1))
	ctx := context.Background()

	p_ := middlewares.NewTableProcessor(middlewares.WithRowMiddleware(row.NewFieldRenameColumnMiddleware(renames)))
	err := p_.AddRow(ctx, obj)
	require.NoError(t, err)
	err = p_.RunTableMiddlewares(ctx)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	err = of.OutputTable(ctx, p_.GetTable(), buf)
	require.NoError(t, err)

	// parse s
	assert.Equal(t, "| b |\n| --- |\n| 1 |", buf.String())
}
