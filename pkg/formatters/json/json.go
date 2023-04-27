package json

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/ugorji/go/codec"
	"io"
	"os"
)

type OutputFormatter struct {
	OutputIndividualRows bool
	OutputFile           string
	OutputFileTemplate   string
	OutputMultipleFiles  bool
	Table                *types.Table
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

// TODO(manuel: 2023-04-25) This could actually all be cleaned up with OutputFormatterOption

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

			encoder := json.NewEncoder(f_)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(row.GetValues())
			if err != nil {
				f_.Close()
				return err
			}
			f_.Close()
			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
		}

		return nil
	}

	if f.OutputFile != "" {
		f_, err := os.Create(f.OutputFile)
		if err != nil {
			return err
		}
		defer f_.Close()
		w = f_
	}

	if f.OutputIndividualRows {
		for _, row := range f.Table.Rows {
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(row.GetValues())
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		jh := &codec.JsonHandle{
			Indent: 2,
		}
		enc := codec.NewEncoder(w, jh)

		// Write the opening bracket for the array
		_, err := w.Write([]byte("[\n"))
		if err != nil {
			return err
		}

		rowCount := len(f.Table.Rows)
		for i, row := range f.Table.Rows {
			// Reset the encoder to avoid memory leaks
			enc.Reset(w)

			// Encode each element in the array
			err = enc.Encode(row.GetValues())
			if err != nil {
				return err
			}

			// Write a comma between elements, except for the last element
			if i < rowCount-1 {
				_, err = w.Write([]byte(",\n"))
				if err != nil {
					return err
				}
			}
		}

		// Write the closing bracket for the array
		_, err = w.Write([]byte("]\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

type OutputFormatterOption func(*OutputFormatter)

func WithOutputIndividualRows(outputIndividualRows bool) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.OutputIndividualRows = outputIndividualRows
	}
}

func WithOutputFile(file string) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.OutputFile = file
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

func WithTable(table *types.Table) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.Table = table
	}
}

func NewOutputFormatter(options ...OutputFormatterOption) *OutputFormatter {
	ret := &OutputFormatter{
		OutputIndividualRows: false,
		Table:                types.NewTable(),
		OutputFile:           "",
		middlewares:          []middlewares.TableMiddleware{},
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}
