package formatters

import (
	middlewares2 "dd-cli/pkg/middlewares"
	"dd-cli/pkg/types"
	"encoding/csv"
	"fmt"
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

func (f *CSVOutputFormatter) AddMiddleware(m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares, m)
}

func (f *CSVOutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
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

	// TODO(manuel, 2022-11-13) add flag to make header output optional
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
