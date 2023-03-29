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

func NewSingleColumnFormatter(column types.FieldName, separator string) *SingleColumnFormatter {
	return &SingleColumnFormatter{
		Table:     &types.Table{},
		Column:    column,
		Separator: separator,
	}
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
