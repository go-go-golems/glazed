package simple

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"io"
	"os"
)

type SingleColumnFormatter struct {
	Column              types.FieldName
	OutputFile          string
	OutputFileTemplate  string
	OutputMultipleFiles bool
	Separator           string
}

var _ formatters.TableOutputFormatter = (*SingleColumnFormatter)(nil)

func (s *SingleColumnFormatter) Close(ctx context.Context, w io.Writer) error {
	return nil
}

func (s *SingleColumnFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (s *SingleColumnFormatter) RegisterRowMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
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

func (s *SingleColumnFormatter) OutputTable(ctx context.Context, table_ *types.Table, w io.Writer) error {
	if s.OutputMultipleFiles {
		if s.OutputFileTemplate == "" && s.OutputFile == "" {
			return errors.New("neither output file or output file template is set")
		}

		for i, row := range table_.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(s.OutputFile, s.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			if s_, ok := row.Get(s.Column); ok {
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

	for i, row := range table_.Rows {
		if value, ok := row.Get(s.Column); ok {
			_, err := fmt.Fprintf(w, "%v", value)
			if err != nil {
				return err
			}
			if i < len(table_.Rows)-1 {
				_, err := fmt.Fprintf(w, "%s", s.Separator)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
