package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
	"os"
)

type OutputFormatter struct {
	OutputFile          string
	OutputFileTemplate  string
	OutputMultipleFiles bool
	WithHeaders         bool
	Separator           rune
}

func (f *OutputFormatter) RegisterMiddlewares(mw *middlewares.Processor) error {
	mw.AddRowMiddlewareInFront(row.NewFlattenObjectMiddleware())
	return nil
}

func (f *OutputFormatter) ContentType() string {
	if f.Separator == '\t' {
		return "text/tab-separated-values"
	}
	return "text/csv"
}

type OutputFormatterOption func(*OutputFormatter)

func WithOutputFile(outputFile string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFile = outputFile
	}
}

func WithHeaders(withHeaders bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.WithHeaders = withHeaders
	}
}

func WithSeparator(separator rune) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.Separator = separator
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

func NewCSVOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		WithHeaders: true,
		Separator:   ',',
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func NewTSVOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		WithHeaders: true,
		Separator:   '\t',
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *OutputFormatter) Output(ctx context.Context, table_ *types.Table, w_ io.Writer) error {
	if f.OutputMultipleFiles {
		for i, row := range table_.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
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

			csvWriter, err := f.newCSVWriter(table_.Columns, f_)
			if err != nil {
				return err
			}

			err = f.writeRow(table_.Columns, row, csvWriter)
			if err != nil {
				return err
			}

			csvWriter.Flush()

			if err := csvWriter.Error(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w_, "Written output to %s\n", outputFileName)
		}

		return nil
	}

	var csvWriter *csv.Writer
	if f.OutputFile != "" {
		f_, err := os.Create(f.OutputFile)
		if err != nil {
			return err
		}
		defer func(f_ *os.File) {
			_ = f_.Close()
		}(f_)

		csvWriter, err = f.newCSVWriter(table_.Columns, w_)
		if err != nil {
			return err
		}
	} else {
		var err error
		csvWriter, err = f.newCSVWriter(table_.Columns, w_)
		if err != nil {
			return err
		}
	}

	for _, row := range table_.Rows {
		err2 := f.writeRow(table_.Columns, row, csvWriter)
		if err2 != nil {
			return err2
		}
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) newCSVWriter(columns []types.FieldName, w_ io.Writer) (*csv.Writer, error) {
	// create a buffer writer
	w := csv.NewWriter(w_)
	w.Comma = f.Separator

	var err error
	if f.WithHeaders {
		err = w.Write(columns)
	}
	return w, err
}

func (f *OutputFormatter) writeRow(columns []types.FieldName, row types.Row, w *csv.Writer) error {
	values := []string{}
	for _, column := range columns {
		if v, ok := row.Get(column); ok {
			values = append(values, fmt.Sprintf("%v", v))
		} else {
			values = append(values, "")
		}
	}
	err := w.Write(values)
	if err != nil {
		return err
	}
	return nil
}
