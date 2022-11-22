package middlewares

import (
	"bytes"
	"glazed/pkg/types"
	"strings"
	"text/template"
)

type ObjectGoTemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
}

// NewObjectGoTemplateMiddleware creates a new template firmware used to process
// individual objects.
//
// It will render the template for each object and return a single field.
func NewObjectGoTemplateMiddleware(templateStrings map[types.FieldName]string) (*ObjectGoTemplateMiddleware, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	templates := map[types.FieldName]*template.Template{}
	for columnName, templateString := range templateStrings {
		tmpl, err := template.New("row").Funcs(funcMap).Parse(templateString)
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
func (rgtm *ObjectGoTemplateMiddleware) Process(object map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}

	for key, tmpl := range rgtm.templates {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, object)
		if err != nil {
			return nil, err
		}
		ret[key] = buf.String()
	}

	return ret, nil
}
