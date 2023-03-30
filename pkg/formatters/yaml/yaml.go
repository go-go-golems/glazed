package yaml

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
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

func (f *OutputFormatter) GetTable() (*types.Table, error) {
	return f.Table, nil
}

func (f *OutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
}

func (f *OutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (f *OutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	f.middlewares = append(f.middlewares, mw)
}

func (f *OutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	f.middlewares = append([]middlewares.TableMiddleware{mw}, f.middlewares...)
}

func (f *OutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	f.middlewares = append(f.middlewares[:i], append([]middlewares.TableMiddleware{mw}, f.middlewares[i:]...)...)
}

func (f *OutputFormatter) Output() (string, error) {
	f.Table.Finalize()

	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return "", err
		}
		f.Table = newTable
	}

	if f.OutputMultipleFiles {
		if f.OutputFileTemplate == "" && f.OutputFile == "" {
			return "", fmt.Errorf("neither output file or output file template is set")
		}

		s := ""

		for i, row := range f.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			d, err := yaml.Marshal(row.GetValues())
			if err != nil {
				return "", err
			}

			err = os.WriteFile(outputFileName, d, 0644)
			if err != nil {
				return "", err
			}
			s += fmt.Sprintf("Wrote output to %s\n", outputFileName)
		}

		return s, nil
	}

	var rows []map[string]interface{}
	for _, row := range f.Table.Rows {
		rows = append(rows, row.GetValues())
	}

	d, err := yaml.Marshal(rows)
	if err != nil {
		return "", err
	}

	if f.OutputFile != "" {
		log.Debug().Str("file", f.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(f.OutputFile, d, 0644)
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
