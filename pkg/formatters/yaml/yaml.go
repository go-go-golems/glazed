package yaml

import (
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

type OutputFormatter struct {
	Table               *types.Table
	OutputFile          string
	OutputFileTemplate  string
	OutputMultipleFiles bool
	middlewares         []middlewares.TableMiddleware
}

func (Y *OutputFormatter) GetTable() (*types.Table, error) {
	return Y.Table, nil
}

func (Y *OutputFormatter) AddRow(row types.Row) {
	Y.Table.Rows = append(Y.Table.Rows, row)
}

func (f *OutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (Y *OutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares, mw)
}

func (Y *OutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	Y.middlewares = append([]middlewares.TableMiddleware{mw}, Y.middlewares...)
}

func (Y *OutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares[:i], append([]middlewares.TableMiddleware{mw}, Y.middlewares[i:]...)...)
}

func (Y *OutputFormatter) Output() (string, error) {
	Y.Table.Finalize()

	for _, middleware := range Y.middlewares {
		newTable, err := middleware.Process(Y.Table)
		if err != nil {
			return "", err
		}
		Y.Table = newTable
	}

	var rows []map[string]interface{}
	for _, row := range Y.Table.Rows {
		rows = append(rows, row.GetValues())

	}

	d, err := yaml.Marshal(rows)
	if err != nil {
		return "", err
	}

	if Y.OutputFile != "" {
		log.Debug().Str("file", Y.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(Y.OutputFile, d, 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return string(d), nil
}

type OutputFormatterOption func(*OutputFormatter)

func WithYAMLOutputFile(outputFile string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFile = outputFile
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

func NewOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}
