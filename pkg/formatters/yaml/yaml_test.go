package yaml

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestYAMLRenameEndToEnd(t *testing.T) {
	of := NewOutputFormatter()
	renames := map[string]string{
		"a": "b",
	}
	obj := types.NewRow(types.MRP("a", 1))
	ctx := context.Background()

	buf := &bytes.Buffer{}
	p_ := middlewares.NewTableProcessor(
		middlewares.WithRowMiddleware(row.NewFieldRenameColumnMiddleware(renames)),
		middlewares.WithTableMiddleware(table.NewOutputMiddleware(of, buf)),
	)
	err := p_.AddRow(ctx, obj)
	require.NoError(t, err)
	err = p_.Close(ctx)
	require.NoError(t, err)

	// parse s
	data := []map[string]interface{}{}
	err = yaml.Unmarshal(buf.Bytes(), &data)
	require.NoError(t, err)
	require.Len(t, data, 1)

	// check if the rename worked
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
