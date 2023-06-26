package json

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJSONRenameEndToEnd(t *testing.T) {
	of := NewOutputFormatter()
	renames := map[string]string{
		"a": "b",
	}
	of.AddTableMiddleware(&table.RenameColumnMiddleware{Renames: renames})
	obj := types.NewMapRow(types.MRP("a", 1))
	of.AddRow(&types.SimpleRow{Hash: obj})
	ctx := context.Background()
	buf := &bytes.Buffer{}
	err := of.Output(ctx, buf)
	require.NoError(t, err)

	s := buf.String()
	// parse s
	data := []types.MapRow{}
	err = json.Unmarshal([]byte(s), &data)
	require.NoError(t, err)
	require.Len(t, data, 1)

	// check if the rename worked
	v, ok := data[0].Get("b")
	assert.True(t, ok)
	assert.Equal(t, 1.0, v)
}
