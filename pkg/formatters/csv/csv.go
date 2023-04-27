package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	middlewares2 "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

type OutputFormatter struct {
	Table               *types.Table
	middlewares         []middlewares2.TableMiddleware
	OutputFile          string
	OutputFileTemplate  string
	OutputMultipleFiles bool
	WithHeaders         bool
	Separator           rune
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
		Table:       types.NewTable(),
		middlewares: []middlewares2.TableMiddleware{},
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
		Table:       types.NewTable(),
		middlewares: []middlewares2.TableMiddleware{},
		WithHeaders: true,
		Separator:   '\t',
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *OutputFormatter) GetTable() (*types.Table, error) {
	return f.Table, nil
}

func (f *OutputFormatter) AddTableMiddleware(m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares, m)
}

func (f *OutputFormatter) AddTableMiddlewareInFront(m middlewares2.TableMiddleware) {
	f.middlewares = append([]middlewares2.TableMiddleware{m}, f.middlewares...)
}

func (f *OutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares[:i], append([]middlewares2.TableMiddleware{m}, f.middlewares[i:]...)...)
}

func (f *OutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
}

func (f *OutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (f *OutputFormatter) Output(ctx context.Context) (string, error) {
	f.Table.Finalize()

	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return "", err
		}
		f.Table = newTable
	}

	if f.OutputMultipleFiles {
		s := ""

		for i, row := range f.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			buf, w, err := f.newCSVWriter()
			if err != nil {
				return "", err
			}

			err = f.writeRow(row, w)
			if err != nil {
				return "", err
			}

			w.Flush()

			if err := w.Error(); err != nil {
				return "", err
			}

			err = os.WriteFile(outputFileName, []byte(buf.String()), 0644)
			if err != nil {
				return "", err
			}
			s += fmt.Sprintf("Written output to %s\n", outputFileName)
		}

		return s, nil
	}

	buf, w, err := f.newCSVWriter()
	if err != nil {
		return "", err
	}

	for _, row := range f.Table.Rows {
		err2 := f.writeRow(row, w)
		if err2 != nil {
			return "", err2
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return "", err
	}

	if f.OutputFile != "" {
		log.Debug().Str("file", f.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(f.OutputFile, []byte(buf.String()), 0644)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Written output to %s\n", f.OutputFile), nil
	}

	return buf.String(), nil
}

func (f *OutputFormatter) newCSVWriter() (*strings.Builder, *csv.Writer, error) {
	// create a buffer writer
	buf := strings.Builder{}
	w := csv.NewWriter(&buf)
	w.Comma = f.Separator

	var err error
	if f.WithHeaders {
		err = w.Write(f.Table.Columns)
	}
	return &buf, w, err
}

func (f *OutputFormatter) writeRow(row types.Row, w *csv.Writer) error {
	values := []string{}
	for _, column := range f.Table.Columns {
		if v, ok := row.GetValues()[column]; ok {
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
