package row

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"strings"
	"text/template"
)

type TemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
	// this field is used to replace "." in keys before passing them to the template,
	// in order to avoid having to use the `index` template function to access fields
	// that contain a ".", which is frequent due to flattening.
	RenameSeparator string
	funcMaps        []template.FuncMap

	renamedColumns map[types.FieldName]types.FieldName
}

var _ middlewares.RowMiddleware = (*TemplateMiddleware)(nil)

type TemplateMiddlewareOption func(*TemplateMiddleware)

func WithRenameSeparator(separator string) TemplateMiddlewareOption {
	return func(t *TemplateMiddleware) {
		t.RenameSeparator = separator
	}
}

func WithFuncMaps(funcMaps ...template.FuncMap) TemplateMiddlewareOption {
	return func(t *TemplateMiddleware) {
		t.funcMaps = append(t.funcMaps, funcMaps...)
	}
}

// NewTemplateMiddleware creates a new TemplateMiddleware
// which is the simplest go template middleware.
//
// It will render the template for each row and return the result as a new column called with
// the given title.
//
// Because nested objects will be flattened to individual columns using the . separator,
// this will make fields inaccessible to the template. One way around this is to use
// {{ index . "field.subfield" }} in the template. Another is to pass a separator rename
// option.
//
// TODO(manuel, 2023-02-02) Add support for passing in custom funcmaps
// See #110 https://github.com/go-go-golems/glazed/issues/110
func NewTemplateMiddleware(
	templateStrings map[types.FieldName]string,
	renameSeparator string) (*TemplateMiddleware, error) {

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
		templates:       templates,
		RenameSeparator: renameSeparator,
		renamedColumns:  map[types.FieldName]types.FieldName{},
	}, nil
}

func (rgtm *TemplateMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	templateValues := map[string]interface{}{}

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value

		if rgtm.RenameSeparator != "" {
			if _, ok := rgtm.renamedColumns[key]; !ok {
				rgtm.renamedColumns[key] = strings.ReplaceAll(key, ".", rgtm.RenameSeparator)
			}

			key = rgtm.renamedColumns[key]
		}
		templateValues[key] = value
	}
	templateValues["_row"] = templateValues

	for columnName, tmpl := range rgtm.templates {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, templateValues)
		if err != nil {
			return nil, err
		}
		s := buf.String()

		row.Set(columnName, s)
	}

	return []types.Row{row}, nil
}

func (rgtm *TemplateMiddleware) Close(ctx context.Context) error {
	return nil
}
