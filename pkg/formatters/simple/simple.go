package simple

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"os"
	"strings"
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

func (s *SingleColumnFormatter) Output(context.Context) (string, error) {
	s.Table.Finalize()

	for _, middleware := range s.middlewares {
		newTable, err := middleware.Process(s.Table)
		if err != nil {
			return "", err
		}
		s.Table = newTable
	}

	if s.OutputMultipleFiles {
		if s.OutputFileTemplate == "" && s.OutputFile == "" {
			return "", fmt.Errorf("neither output file or output file template is set")
		}

		ret := ""

		for i, row := range s.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(s.OutputFile, s.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			if s_, ok := row.GetValues()[s.Column]; ok {
				v := fmt.Sprintf("%v", s_)
				err = os.WriteFile(outputFileName, []byte(v), 0644)
				if err != nil {
					return "", err
				}
				ret += fmt.Sprintf("Wrote output to %s\n", outputFileName)
			}
		}

		return ret, nil

	}

	strs := []string{}

	for _, row := range s.Table.Rows {
		if value, ok := row.GetValues()[s.Column]; ok {
			strs = append(strs, fmt.Sprintf("%v", value))
		}
	}

	v := strings.Join(strs, s.Separator)

	if s.OutputFile != "" {
		err := os.WriteFile(s.OutputFile, []byte(v), 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return v, nil
}
