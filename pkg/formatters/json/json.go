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
}

func (f *OutputFormatter) RegisterMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (f *OutputFormatter) ContentType() string {
	return "application/json"
}

func (f *OutputFormatter) Output(ctx context.Context, table_ *types.Table, w io.Writer) error {
	if f.OutputMultipleFiles {
		if f.OutputFileTemplate == "" && f.OutputFile == "" {
			return fmt.Errorf("neither output file or output file template is set")
		}

		for i, row := range table_.Rows {
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
			err = encoder.Encode(row)
			if err != nil {
				_ = f_.Close()
				return err
			}
			_ = f_.Close()
			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
		}

		return nil
	}

	if f.OutputFile != "" {
		f_, err := os.Create(f.OutputFile)
		if err != nil {
			return err
		}
		defer func(f_ *os.File) {
			_ = f_.Close()
		}(f_)
		w = f_
	}

	if f.OutputIndividualRows {
		for _, row := range table_.Rows {
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(row)
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

		rowCount := len(table_.Rows)
		for i, row := range table_.Rows {
			// Reset the encoder to avoid memory leaks
			enc.Reset(w)

			// Encode each element in the array
			err = enc.Encode(row)
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

func NewOutputFormatter(options ...OutputFormatterOption) *OutputFormatter {
	ret := &OutputFormatter{
		OutputIndividualRows: false,
		OutputFile:           "",
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

// RowOutputFormatter is a streaming formatter that can only output individual rows as dictionaries.
type RowOutputFormatter struct {
	indent string
}

type RowOutputFormatterOption func(*RowOutputFormatter)

func WithIndent(indent string) RowOutputFormatterOption {
	return func(formatter *RowOutputFormatter) {
		formatter.indent = indent
	}
}

func NewRowOutputFormatter(options ...RowOutputFormatterOption) *RowOutputFormatter {
	ret := &RowOutputFormatter{
		indent: "  ",
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

func (r *RowOutputFormatter) RegisterMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (r *RowOutputFormatter) Output(ctx context.Context, row types.Row, w io.Writer) error {
	m := types.RowToMap(row)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", r.indent)
	err := encoder.Encode(m)
	if err != nil {
		return err
	}

	return nil
}

func (r *RowOutputFormatter) ContentType() string {
	return "application/json"
}
