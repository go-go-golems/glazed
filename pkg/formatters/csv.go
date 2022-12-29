package formatters

import (
	"encoding/csv"
	"fmt"
	middlewares2 "github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
	"strings"
)

type CSVOutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares2.TableMiddleware
	WithHeaders bool
	Separator   rune
}

func NewCSVOutputFormatter() *CSVOutputFormatter {
	return &CSVOutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares2.TableMiddleware{},
		WithHeaders: true,
		Separator:   ',',
	}
}

func NewTSVOutputFormatter() *CSVOutputFormatter {
	return &CSVOutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares2.TableMiddleware{},
		WithHeaders: true,
		Separator:   '\t',
	}
}

func (f *CSVOutputFormatter) AddTableMiddleware(m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares, m)
}

func (f *CSVOutputFormatter) AddTableMiddlewareInFront(m middlewares2.TableMiddleware) {
	f.middlewares = append([]middlewares2.TableMiddleware{m}, f.middlewares...)
}

func (f *CSVOutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares[:i], append([]middlewares2.TableMiddleware{m}, f.middlewares[i:]...)...)
}

func (f *CSVOutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
}

func (f *CSVOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (f *CSVOutputFormatter) Output() (string, error) {
	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return "", err
		}
		f.Table = newTable
	}

	// create a buffer writer
	buf := strings.Builder{}
	w := csv.NewWriter(&buf)
	w.Comma = f.Separator

	var err error
	if f.WithHeaders {
		err = w.Write(f.Table.Columns)
	}
	if err != nil {
		return "", err
	}

	for _, row := range f.Table.Rows {
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
			return "", err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
