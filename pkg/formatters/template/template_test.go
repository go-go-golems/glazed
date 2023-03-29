package template

import (
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
	of := NewTemplateOutputFormatter(tmpl, []template.FuncMap{
		sprig.TxtFuncMap(),
		templating.TemplateFuncs,
	}, make(map[string]interface{}), "")
	renames := map[string]string{
		"a": "b",
	}
	of.AddTableMiddleware(&table.RenameColumnMiddleware{Renames: renames})
	of.AddRow(&types.SimpleRow{Hash: map[string]interface{}{"a": 1}})
	s, err := of.Output()
	require.NoError(t, err)

	assert.Equal(t, `1`, s)
}
