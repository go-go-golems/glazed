package formatters

import (
	"fmt"
	"github.com/scylladb/termtables"
	"github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
)

// This part of the library contains helper functionality to do output formatting
// for data.
//
// We want to do the following:
//    - print a Table with a header
//    - print the Table as csv
//    - render raw data as json
//    - render data as sqlite (potentially into multiple tables)
//    - support multiple tables
//        - transform tree like structures into flattened tables
//    - make it easy for the user to add data
//    - make it easy for the user to specify filters and fields
//    - provide a middleware like structure to chain filters and transformers
//    - provide a way to add documentation to the output / data schema
//    - support go templating
//    - load formatting values from a config file
//    - streaming functionality (i.e., output as values come in)
//
// Advanced functionality:
//    - excel output
//    - pager and search
//    - highlight certain values
//    - filter the input structure / output structure using a jq like query language

type OutputFormatter interface {
	// TODO(manuel, 2022-11-12) We need to be able to output to a directory / to a stream / to multiple files
	AddRow(row types.Row)
	AddTableMiddleware(m middlewares.TableMiddleware)
	Output() (string, error)
}

// The following is all geared towards tabulated output

type TableOutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
	TableFormat string
}

func NewTableOutputFormatter(tableFormat string) *TableOutputFormatter {
	return &TableOutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
		TableFormat: tableFormat,
	}
}

func (tof *TableOutputFormatter) Output() (string, error) {
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
				s = fmt.Sprintf("%v", v)
			}
			row_ = append(row_, s)
		}
		table.AddRow(row_...)
	}

	return table.Render(), nil
}

func (tof *TableOutputFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares, m)
}

func (tof *TableOutputFormatter) AddRow(row types.Row) {
	tof.Table.Rows = append(tof.Table.Rows, row)
}
