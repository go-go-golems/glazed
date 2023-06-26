package template

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestTemplateRenameEndToEnd(t *testing.T) {
	// template that gets rows[0].b
	tmpl := `{{ (index .rows 0).b }}`
	of := NewOutputFormatter(tmpl,
		WithTemplateFuncMaps([]template.FuncMap{
			sprig.TxtFuncMap(),
			templating.TemplateFuncs,
		}))
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

	assert.Equal(t, `1`, buf.String())
}
