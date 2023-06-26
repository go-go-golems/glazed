package simple

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
	"os"
)

type SingleColumnFormatter struct {
	Table               *types.Table
	Column              types.FieldName
	OutputFile          string
	OutputFileTemplate  string
	OutputMultipleFiles bool
	Separator           string
	middlewares         []middlewares.TableMiddleware
}

type SingleColumnFormatterOption func(*SingleColumnFormatter)

func WithSeparator(separator string) SingleColumnFormatterOption {
	return func(f *SingleColumnFormatter) {
		f.Separator = separator
	}
}

func WithOutputFileTemplate(template string) SingleColumnFormatterOption {
	return func(f *SingleColumnFormatter) {
		f.OutputFileTemplate = template
	}
}

func WithOutputMultipleFiles(outputMultipleFiles bool) SingleColumnFormatterOption {
	return func(f *SingleColumnFormatter) {
		f.OutputMultipleFiles = outputMultipleFiles
	}
}

func WithOutputFile(outputFile string) SingleColumnFormatterOption {
	return func(f *SingleColumnFormatter) {
		f.OutputFile = outputFile
	}
}

func NewSingleColumnFormatter(column types.FieldName, opts ...SingleColumnFormatterOption) *SingleColumnFormatter {
	f := &SingleColumnFormatter{
		Table:     &types.Table{},
		Column:    column,
		Separator: "\n",
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (s *SingleColumnFormatter) ContentType() string {
	return "text/plain"
}

func (s *SingleColumnFormatter) AddRow(row types.Row) {
	s.Table.Rows = append(s.Table.Rows, row)
}

func (s *SingleColumnFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	s.Table.Columns = columnOrder
}

func (s *SingleColumnFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	s.middlewares = append(s.middlewares, m)
}

func (s *SingleColumnFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	s.middlewares = append([]middlewares.TableMiddleware{m}, s.middlewares...)
}

func (s *SingleColumnFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	s.middlewares = append(s.middlewares[:i], append([]middlewares.TableMiddleware{m}, s.middlewares[i:]...)...)
}

func (s *SingleColumnFormatter) GetTable() (*types.Table, error) {
	return s.Table, nil
}

func (s *SingleColumnFormatter) Output(ctx context.Context, w io.Writer) error {
	s.Table.Finalize()

	for _, middleware := range s.middlewares {
		newTable, err := middleware.Process(s.Table)
		if err != nil {
			return err
		}
		s.Table = newTable
	}

	if s.OutputMultipleFiles {
		if s.OutputFileTemplate == "" && s.OutputFile == "" {
			return fmt.Errorf("neither output file or output file template is set")
		}

		for i, row := range s.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(s.OutputFile, s.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			values := row.GetValues()

			if s_, ok := values.Get(s.Column); ok {
				v := fmt.Sprintf("%v", s_)
				err = os.WriteFile(outputFileName, []byte(v), 0644)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
			}
		}

		return nil

	}

	if s.OutputFile != "" {
		f_, err := os.Create(s.OutputFile)
		if err != nil {
			return err
		}

		w = f_
	}

	for i, row := range s.Table.Rows {
		if value, ok := row.GetValues().Get(s.Column); ok {
			_, err := fmt.Fprintf(w, "%v", value)
			if err != nil {
				return err
			}
			if i < len(s.Table.Rows)-1 {
				_, err := fmt.Fprintf(w, "%s", s.Separator)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
