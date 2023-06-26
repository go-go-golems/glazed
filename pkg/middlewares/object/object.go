package object

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/types"
	"text/template"
)

type ObjectGoTemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
}

// NewObjectGoTemplateMiddleware creates a new template firmware used to process
// individual objects.
//
// It will render the template for each object and return a single field.
//
// TODO(manuel, 2023-02-02) Add support for passing in custom funcmaps
// See #110 https://github.com/go-go-golems/glazed/issues/110
func NewObjectGoTemplateMiddleware(
	templateStrings map[types.FieldName]string,
) (*ObjectGoTemplateMiddleware, error) {
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

	return &ObjectGoTemplateMiddleware{
		templates: templates,
	}, nil
}

// Process will render each template for the input object and return an object with the newly created fields.
//
// TODO(manuel, 2022-11-21) This should allow merging the new results straight back
func (rgtm *ObjectGoTemplateMiddleware) Process(object types.MapRow) ([]types.MapRow, error) {
	ret := types.NewMapRow()

	for key, tmpl := range rgtm.templates {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, object)
		if err != nil {
			return nil, err
		}
		ret.Set(key, buf.String())
	}

	return []types.MapRow{ret}, nil
}
