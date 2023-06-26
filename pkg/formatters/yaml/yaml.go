package yaml

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type OutputFormatter struct {
	Table                *types.Table
	OutputFile           string
	OutputFileTemplate   string
	OutputMultipleFiles  bool
	OutputIndividualRows bool
	middlewares          []middlewares.TableMiddleware
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

func (f *OutputFormatter) Output(ctx context.Context, w io.Writer) error {
	f.Table.Finalize()

	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return err
		}
		f.Table = newTable
	}

	if f.OutputMultipleFiles {
		if f.OutputFileTemplate == "" && f.OutputFile == "" {
			return fmt.Errorf("neither output file or output file template is set")
		}

		for i, row := range f.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			f_, err := os.Create(outputFileName)
			if err != nil {
				return err
			}

			encoder := yaml.NewEncoder(f_)
			err = encoder.Encode(row.GetValues())
			if err != nil {
				f_.Close()
				return err
			}

			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
			f_.Close()
		}

		return nil
	}

	if f.OutputIndividualRows {
		if len(f.Table.Rows) > 1 {
			return fmt.Errorf("output individual rows is set but there are multiple rows in the table")
		}

		if f.OutputFile != "" {
			f_, err := os.Create(f.OutputFile)
			if err != nil {
				return err
			}
			w = f_
			defer f_.Close()

			if len(f.Table.Rows) == 0 {
				_, _ = fmt.Fprintln(w, "Empty table, an empty file was created")
				return nil
			}
		}

		encoder := yaml.NewEncoder(w)
		err := encoder.Encode(f.Table.Rows[0].GetValues())
		if err != nil {
			return err
		}

		return nil
	} else {
		var rows []types.MapRow
		for _, row := range f.Table.Rows {
			rows = append(rows, row.GetValues())
		}

		encoder := yaml.NewEncoder(w)
		err := encoder.Encode(rows)
		if err != nil {
			return err
		}

		return nil
	}
}

func (f *OutputFormatter) ContentType() string {
	return "application/yaml"
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

func WithOutputIndividualRows(outputIndividualRows bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputIndividualRows = outputIndividualRows
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
