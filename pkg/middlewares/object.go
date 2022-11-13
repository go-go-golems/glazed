package middlewares

import (
	"bytes"
	"dd-cli/pkg/types"
	"text/template"
)

type ObjectGoTemplateMiddleware struct {
	template   *template.Template
	columnName types.FieldName
}

// NewObjectGoTemplateMiddleware creates a new template firmware used to process
// individual objects.
//
// It will render the template for each object and return a single field.
func NewObjectGoTemplateMiddleware(
	columName types.FieldName,
	templateString string) (*ObjectGoTemplateMiddleware, error) {
	tmpl, err := template.New("row").Parse(templateString)
	if err != nil {
		return nil, err
	}

	return &ObjectGoTemplateMiddleware{
		columnName: columName,
		template:   tmpl,
	}, nil
}

func (rgtm *ObjectGoTemplateMiddleware) Process(object interface{}) (interface{}, error) {
	ret := map[string]string{}

	var buf bytes.Buffer
	err := rgtm.template.Execute(&buf, object)
	if err != nil {
		return nil, err
	}
	ret[rgtm.columnName] = buf.String()

	return ret, nil
}
