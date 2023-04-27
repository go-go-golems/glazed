package json

import (
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
	of.AddRow(&types.SimpleRow{Hash: map[string]interface{}{"a": 1}})
	ctx := context.Background()
	s, err := of.Output(ctx)
	require.NoError(t, err)

	// parse s
	data := []map[string]interface{}{}
	err = json.Unmarshal([]byte(s), &data)
	require.NoError(t, err)
	require.Len(t, data, 1)

	// check if the rename worked
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1.0, v)
}
