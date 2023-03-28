package formatters

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/termtables"
	"os"
	"strings"
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

	SetColumnOrder(columnOrder []types.FieldName)

	// AddTableMiddleware adds a middleware at the end of the processing list
	AddTableMiddleware(m middlewares.TableMiddleware)
	AddTableMiddlewareInFront(m middlewares.TableMiddleware)
	AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware)

	GetTable() (*types.Table, error)

	Output() (string, error)
}

// The following is all geared towards tabulated output

type TableOutputFormatter struct {
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
	TableFormat string
	OutputFile  string
}

func NewTableOutputFormatter(tableFormat string, outputFile string) *TableOutputFormatter {
	return &TableOutputFormatter{
		Table:       types.NewTable(),
		OutputFile:  outputFile,
		middlewares: []middlewares.TableMiddleware{},
		TableFormat: tableFormat,
	}
}

func (tof *TableOutputFormatter) GetTable() (*types.Table, error) {
	return tof.Table, nil
}

func (tof *TableOutputFormatter) Output() (string, error) {
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

func (tof *TableOutputFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares, m)
}

func (tof *TableOutputFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	tof.middlewares = append([]middlewares.TableMiddleware{m}, tof.middlewares...)
}

func (tof *TableOutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares[:i], append([]middlewares.TableMiddleware{m}, tof.middlewares[i:]...)...)
}

func (tof *TableOutputFormatter) AddRow(row types.Row) {
	tof.Table.Rows = append(tof.Table.Rows, row)
}

func (tof *TableOutputFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	tof.Table.Columns = columnOrder
}
