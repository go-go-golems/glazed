package simple

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"strings"
)

type SingleColumnFormatter struct {
	Table       *types.Table
	Column      types.FieldName
	Separator   string
	middlewares []middlewares.TableMiddleware
}

type SingleColumnFormatterOption func(*SingleColumnFormatter)

func WithSeparator(separator string) SingleColumnFormatterOption {
	return func(f *SingleColumnFormatter) {
		f.Separator = separator
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

func (s *SingleColumnFormatter) Output() (string, error) {
	s.Table.Finalize()

	for _, middleware := range s.middlewares {
		newTable, err := middleware.Process(s.Table)
		if err != nil {
			return "", err
		}
		s.Table = newTable
	}

	buf := strings.Builder{}

	for _, row := range s.Table.Rows {
		if value, ok := row.GetValues()[s.Column]; ok {
			buf.WriteString(fmt.Sprintf("%v%s", value, s.Separator))
		}
	}

	return buf.String(), nil
}
