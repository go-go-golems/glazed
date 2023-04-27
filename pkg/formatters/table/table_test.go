package table

import (
	"bytes"
	"context"
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
	of.AddTableMiddleware(&table.RenameColumnMiddleware{Renames: renames})
	of.AddRow(&types.SimpleRow{Hash: map[string]interface{}{"a": 1}})
	ctx := context.Background()
	buf := &bytes.Buffer{}
	err := of.Output(ctx, buf)
	require.NoError(t, err)

	// parse s
	assert.Equal(t, "| b |\n| --- |\n| 1 |", buf.String())
}
