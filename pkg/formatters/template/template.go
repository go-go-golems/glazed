package template

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"os"
	"text/template"
)

type OutputFormatter struct {
	Template            string
	Table               *types.Table
	TemplateFuncMaps    []template.FuncMap
	OutputFileTemplate  string
	OutputMultipleFiles bool
	middlewares         []middlewares.TableMiddleware
	OutputFile          string
	AdditionalData      interface{}
}

func (t *OutputFormatter) GetTable() (*types.Table, error) {
	return t.Table, nil
}

func (t *OutputFormatter) AddRow(row types.Row) {
	t.Table.Rows = append(t.Table.Rows, row)
}

func (t *OutputFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	t.Table.Columns = columnOrder
}

func (t *OutputFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	t.middlewares = append(t.middlewares, m)
}

func (t *OutputFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	t.middlewares = append([]middlewares.TableMiddleware{m}, t.middlewares...)
}

func (t *OutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	t.middlewares = append(t.middlewares[:i], append([]middlewares.TableMiddleware{m}, t.middlewares[i:]...)...)
}

func (t *OutputFormatter) Output(context.Context) (string, error) {
	t.Table.Finalize()

	for _, middleware := range t.middlewares {
		newTable, err := middleware.Process(t.Table)
		if err != nil {
			return "", err
		}
		t.Table = newTable
	}

	t2 := template.New("template")
	for _, templateFuncMap := range t.TemplateFuncMaps {
		t2 = t2.Funcs(templateFuncMap)
	}
	tmpl, err := t2.Parse(t.Template)
	if err != nil {
		return "", err
	}

	if t.OutputMultipleFiles {
		if t.OutputFileTemplate == "" && t.OutputFile == "" {
			return "", fmt.Errorf("neither output file or output file template is set")
		}

		s := ""

		for i, row := range t.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(t.OutputFile, t.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			var buf bytes.Buffer
			tableData := []map[types.FieldName]interface{}{row.GetValues()}

			data := map[string]interface{}{
				"rows": tableData,
				"data": t.AdditionalData,
			}

			err = tmpl.Execute(&buf, data)

			if err != nil {
				return "", err
			}

			err = os.WriteFile(outputFileName, buf.Bytes(), 0644)
			if err != nil {
				return "", err
			}

			s += fmt.Sprintf("Wrote output to %s\n", outputFileName)
		}

		return s, nil
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

	if err != nil {
		return "", err
	}

	if t.OutputFile != "" {
		err = os.WriteFile(t.OutputFile, buf.Bytes(), 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return buf.String(), nil
}

type OutputFormatterOption func(*OutputFormatter)

func WithTemplateFuncMaps(templateFuncMaps []template.FuncMap) OutputFormatterOption {
	return func(t *OutputFormatter) {
		t.TemplateFuncMaps = templateFuncMaps
	}
}

func WithAdditionalData(additionalData interface{}) OutputFormatterOption {
	return func(t *OutputFormatter) {
		t.AdditionalData = additionalData
	}
}

func WithOutputFile(outputFile string) OutputFormatterOption {
	return func(t *OutputFormatter) {
		t.OutputFile = outputFile
	}
}

func WithOutputFileTemplate(outputFileTemplate string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFileTemplate = outputFileTemplate
	}
}

func WithOutputMultipleFiles(outputMultipleFiles bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputMultipleFiles = outputMultipleFiles
	}
}

// NewOutputFormatter creates a new OutputFormatter.
//
// TODO(manuel, 2023-02-19) This is quite an ugly constructor signature.
// See: https://github.com/go-go-golems/glazed/issues/147
func NewOutputFormatter(template string, opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		Template:       template,
		AdditionalData: map[string]interface{}{},
		Table:          types.NewTable(),
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}
