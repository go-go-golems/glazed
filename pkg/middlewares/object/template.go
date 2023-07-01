package object

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/types"
	"text/template"
)

type TemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
}

func (rgtm *TemplateMiddleware) Close(ctx context.Context) error {
	return nil
}

// NewTemplateMiddleware creates a new template firmware used to process
// individual objects.
//
// It will render the template for each object and return a single field.
//
// TODO(manuel, 2023-02-02) Add support for passing in custom funcmaps
// See #110 https://github.com/go-go-golems/glazed/issues/110
func NewTemplateMiddleware(
	templateStrings map[types.FieldName]string,
) (*TemplateMiddleware, error) {
	templates := map[types.FieldName]*template.Template{}
	for columnName, templateString := range templateStrings {
		tmpl, err := template.New("row").
			Funcs(sprig.TxtFuncMap()).
			Funcs(templating.TemplateFuncs).
			Parse(templateString)
		if err != nil {
			return nil, err
		}
		templates[columnName] = tmpl
	}

	return &TemplateMiddleware{
		templates: templates,
	}, nil
}

// Process will render each template for the input object and return an object with the newly created fields.
//
// TODO(manuel, 2022-11-21) This should allow merging the new results straight back
func (rgtm *TemplateMiddleware) Process(ctx context.Context, object types.Row) ([]types.Row, error) {
	ret := types.NewRow()

	for key, tmpl := range rgtm.templates {
		var buf bytes.Buffer
		m := map[string]interface{}{}

		for pair := object.Oldest(); pair != nil; pair = pair.Next() {
			m[pair.Key] = pair.Value
		}
		err := tmpl.Execute(&buf, m)
		if err != nil {
			return nil, err
		}
		ret.Set(key, buf.String())
	}

	return []types.Row{ret}, nil
}
