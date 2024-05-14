package template

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"io"
	"os"
	"text/template"
)

type OutputFormatter struct {
	Template            string
	TemplateFuncMaps    []template.FuncMap
	OutputFileTemplate  string
	OutputMultipleFiles bool
	OutputFile          string
	AdditionalData      interface{}
}

var _ formatters.TableOutputFormatter = (*OutputFormatter)(nil)

func (t *OutputFormatter) Close(ctx context.Context, w io.Writer) error {
	return nil
}

func (t *OutputFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (t *OutputFormatter) RegisterRowMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (t *OutputFormatter) OutputTable(ctx context.Context, table_ *types.Table, w io.Writer) error {
	t2 := template.New("template")
	for _, templateFuncMap := range t.TemplateFuncMaps {
		t2 = t2.Funcs(templateFuncMap)
	}
	tmpl, err := t2.Parse(t.Template)
	if err != nil {
		return err
	}

	if t.OutputMultipleFiles {
		if t.OutputFileTemplate == "" && t.OutputFile == "" {
			return errors.New("neither output file or output file template is set")
		}

		for i, row := range table_.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(t.OutputFile, t.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			f_, err := os.Create(outputFileName)
			if err != nil {
				return err
			}
			defer func(f_ *os.File) {
				_ = f_.Close()
			}(f_)

			tableData := []types.Row{row}

			data := map[string]interface{}{
				// TODO(manuel, 2023-06-25) Convert to normal maps for templating
				"rows": tableData,
				"data": t.AdditionalData,
			}

			err = tmpl.Execute(f_, data)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
		}

		return nil
	}

	var tableData []map[string]interface{}

	for _, row := range table_.Rows {
		m := make(map[string]interface{})

		for pair := row.Oldest(); pair != nil; pair = pair.Next() {
			m[pair.Key] = pair.Value
		}

		tableData = append(tableData, m)
	}
	data := map[string]interface{}{
		// TODO(manuel, 2023-06-25) Convert to normal maps for templating
		"rows": tableData,
		"data": t.AdditionalData,
	}

	if t.OutputFile != "" {
		f_, err := os.Create(t.OutputFile)
		if err != nil {
			return err
		}
		defer func(f_ *os.File) {
			_ = f_.Close()
		}(f_)

		w = f_
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) ContentType() string {
	return "text/plain"
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
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}
