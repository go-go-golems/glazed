package formatters

import (
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
	of.AddTableMiddleware(&middlewares.RenameColumnMiddleware{Renames: renames})
	of.AddRow(&types.SimpleRow{Hash: map[string]interface{}{"a": 1}})
	s, err := of.Output()
	require.NoError(t, err)

	data, err := helpers.ParseCSV(strings.NewReader(s))
	require.NoError(t, err)
	require.Len(t, data, 1)
	v, ok := data[0]["b"]
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}
