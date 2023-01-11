package formatters

import (
	"bytes"
	"github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
	"text/template"
)

type TemplateFormatter struct {
	Template       string
	Table          *types.Table
	TemplateFuncs  template.FuncMap
	middlewares    []middlewares.TableMiddleware
	AdditionalData interface{}
}

func (t *TemplateFormatter) AddRow(row types.Row) {
	t.Table.Rows = append(t.Table.Rows, row)
}

func (t *TemplateFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	t.Table.Columns = columnOrder
}

func (t *TemplateFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	t.middlewares = append(t.middlewares, m)
}

func (t *TemplateFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	t.middlewares = append([]middlewares.TableMiddleware{m}, t.middlewares...)
}

func (t *TemplateFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	t.middlewares = append(t.middlewares[:i], append([]middlewares.TableMiddleware{m}, t.middlewares[i:]...)...)
}

func (t *TemplateFormatter) Output() (string, error) {
	for _, middleware := range t.middlewares {
		newTable, err := middleware.Process(t.Table)
		if err != nil {
			return "", err
		}
		t.Table = newTable
	}

	tmpl, err := template.New("template").Funcs(t.TemplateFuncs).Parse(t.Template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	var tableData []map[types.FieldName]interface{}
	for _, row := range t.Table.Rows {
		tableData = append(tableData, row.GetValues())
	}
	data := map[string]interface{}{
		"rows": tableData,
		"data": t.AdditionalData,
	}

	err = tmpl.Execute(&buf, data)

	return buf.String(), err
}

func NewTemplateOutputFormatter(template string, templateFuncs template.FuncMap, additionalData interface{}) *TemplateFormatter {
	return &TemplateFormatter{
		Template:       template,
		Table:          types.NewTable(),
		TemplateFuncs:  templateFuncs,
		AdditionalData: additionalData,
	}
}
