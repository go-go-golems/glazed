package table

import (
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
	s, err := of.Output()
	require.NoError(t, err)

	// parse s
	assert.Equal(t, "| b |\n| - |\n| 1 |\n", s)
}
