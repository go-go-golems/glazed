package table

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/termtables"
	"os"
	"strings"
)

type OutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
	TableFormat string
	OutputFile  string
}

type OutputFormatterOption func(*OutputFormatter)

func WithOutputFile(outputFile string) OutputFormatterOption {
	return func(tof *OutputFormatter) {
		tof.OutputFile = outputFile
	}
}

func NewOutputFormatter(tableFormat string, opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
		TableFormat: tableFormat,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (tof *OutputFormatter) GetTable() (*types.Table, error) {
	return tof.Table, nil
}

func (tof *OutputFormatter) Output() (string, error) {
	tof.Table.Finalize()

	for _, middleware := range tof.middlewares {
		newTable, err := middleware.Process(tof.Table)
		if err != nil {
			return "", err
		}
		tof.Table = newTable
	}

	table := termtables.CreateTable()

	if tof.TableFormat == "markdown" {
		table.SetModeMarkdown()
	} else if tof.TableFormat == "html" {
		table.SetModeHTML()
	} else {
		table.SetModeTerminal()
	}

	for _, column := range tof.Table.Columns {
		table.AddHeaders(column)
	}

	for _, row := range tof.Table.Rows {
		var row_ []interface{}
		values := row.GetValues()
		for _, column := range tof.Table.Columns {
			s := ""
			if v, ok := values[column]; ok {
				if v_, ok := v.([]interface{}); ok {
					var elms []string
					for _, elm := range v_ {
						elms = append(elms, fmt.Sprintf("%v", elm))
					}
					s = strings.Join(elms, ", ")
				} else {
					s = fmt.Sprintf("%v", v)
				}
			}
			row_ = append(row_, s)
		}
		table.AddRow(row_...)
	}

	s := table.Render()

	if tof.OutputFile != "" {
		log.Debug().Str("file", tof.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(tof.OutputFile, []byte(s), 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return s, nil
}

func (tof *OutputFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares, m)
}

func (tof *OutputFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	tof.middlewares = append([]middlewares.TableMiddleware{m}, tof.middlewares...)
}

func (tof *OutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares[:i], append([]middlewares.TableMiddleware{m}, tof.middlewares[i:]...)...)
}

func (tof *OutputFormatter) AddRow(row types.Row) {
	tof.Table.Rows = append(tof.Table.Rows, row)
}

func (tof *OutputFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	tof.Table.Columns = columnOrder
}
