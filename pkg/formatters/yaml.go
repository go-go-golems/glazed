package formatters

import (
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
)

type YAMLOutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
}

func (Y *YAMLOutputFormatter) AddRow(row types.Row) {
	Y.Table.Rows = append(Y.Table.Rows, row)
}

func (f *YAMLOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (Y *YAMLOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares, mw)
}

func (Y *YAMLOutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	Y.middlewares = append([]middlewares.TableMiddleware{mw}, Y.middlewares...)
}

func (Y *YAMLOutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares[:i], append([]middlewares.TableMiddleware{mw}, Y.middlewares[i:]...)...)
}

func (Y *YAMLOutputFormatter) Output() (string, error) {
	Y.Table.Finalize()

	for _, middleware := range Y.middlewares {
		newTable, err := middleware.Process(Y.Table)
		if err != nil {
			return "", err
		}
		Y.Table = newTable
	}

	var rows []map[string]interface{}
	for _, row := range Y.Table.Rows {
		rows = append(rows, row.GetValues())

	}

	d, err := yaml.Marshal(rows)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func NewYAMLOutputFormatter() *YAMLOutputFormatter {
	return &YAMLOutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
	}
}
