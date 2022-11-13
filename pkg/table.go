package pkg

import (
	"fmt"
	"github.com/scylladb/termtables"
	"sort"
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
	AddRow(row Row)
	AddMiddleware(m TableMiddleware)
	Output() (string, error)
}

// The following is all geared towards tabulated output

type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type MapRow = map[FieldName]GenericCellValue

type Row interface {
	GetFields() []FieldName
	GetValues() MapRow
}

type TableMiddleware interface {
	// Process transform a single row into potential multiple rows split across multiple tables
	Process(table *Table) (*Table, error)
}

type Table struct {
	Columns []FieldName
	Rows    []Row
}

func NewTable() *Table {
	return &Table{
		Columns: []FieldName{},
		Rows:    []Row{},
	}
}

type SimpleRow struct {
	Hash MapRow
}

func (sr *SimpleRow) GetFields() []FieldName {
	ret := []FieldName{}
	for key := range sr.Hash {
		ret = append(ret, key)
	}
	return ret
}

func (sr *SimpleRow) GetValues() MapRow {
	return sr.Hash
}

type TableOutputFormatter struct {
	Table       *Table
	middlewares []TableMiddleware
	TableFormat string
}

func NewTableOutputFormatter(tableFormat string) *TableOutputFormatter {
	return &TableOutputFormatter{
		Table:       NewTable(),
		middlewares: []TableMiddleware{},
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

func (tof *TableOutputFormatter) AddMiddleware(m TableMiddleware) {
	tof.middlewares = append(tof.middlewares, m)
}

func (tof *TableOutputFormatter) AddRow(row Row) {
	tof.Table.Rows = append(tof.Table.Rows, row)
}

// Let's go with different middlewares

// FieldsFilterMiddleware keeps columns that are in the fields list and removes
// columns that are in the filters list.
//
// empty lists means that all columns are accepted.
//
// The returned rows are SimpleRows
type FieldsFilterMiddleware struct {
	fields        map[string]interface{}
	filters       map[string]interface{}
	prefixFields  []string
	prefixFilters []string
}

func NewFieldsFilterMiddleware(fields []string, filters []string) *FieldsFilterMiddleware {
	fieldHash := map[string]interface{}{}
	prefixFields := []string{}
	prefixFilters := []string{}

	for _, field := range fields {
		if strings.HasSuffix(field, ".") {
			prefixFields = append(prefixFields, field)
		} else {
			fieldHash[field] = nil
		}
	}
	filterHash := map[string]interface{}{}
	for _, filter := range filters {
		if strings.HasSuffix(filter, ".") {
			prefixFilters = append(prefixFilters, filter)
		} else {
			filterHash[filter] = nil
		}
	}
	return &FieldsFilterMiddleware{
		fields:        fieldHash,
		filters:       filterHash,
		prefixFields:  prefixFields,
		prefixFilters: prefixFilters,
	}
}

func (ffm *FieldsFilterMiddleware) Process(table *Table) (*Table, error) {
	ret := &Table{
		Columns: []FieldName{},
		Rows:    []Row{},
	}

	// how do we keep order here
	newColumns := map[FieldName]interface{}{}

	if len(ffm.fields) == 0 && len(ffm.filters) == 0 {
		return table, nil
	}

	for _, row := range table.Rows {
		values := row.GetValues()
		newRow := SimpleRow{
			Hash: map[FieldName]GenericCellValue{},
		}

	NextRow:
		for rowField, value := range values {
			// skip all of this if we already filtered that field
			if _, ok := newColumns[rowField]; !ok {
				exactMatchFound := false
				prefixMatchFound := false

				exactFilterMatchFound := false
				prefixFilterMatchFound := false

				if len(ffm.fields) > 0 || len(ffm.prefixFields) > 0 {
					// first go through exact matches
					if _, ok := ffm.fields[rowField]; ok {
						exactMatchFound = true
					} else {
						// else, test against all prefixes
						for _, prefix := range ffm.prefixFields {
							if strings.HasPrefix(rowField, prefix) {
								prefixMatchFound = true
								break
							}
						}
					}

					if !exactMatchFound && !prefixMatchFound {
						continue NextRow
					}
				}

				if len(ffm.filters) > 0 || len(ffm.prefixFilters) > 0 {
					// if an exact filter matches, move on
					if _, ok := ffm.filters[rowField]; ok {
						exactFilterMatchFound = true
						continue NextRow
					} else {
						// else, test against all prefixes
						for _, prefix := range ffm.prefixFilters {
							if strings.HasPrefix(rowField, prefix) {
								prefixFilterMatchFound = true
								break
							}
						}
					}
				}

				if exactMatchFound {
					newColumns[rowField] = nil
				} else if prefixMatchFound {
					if prefixFilterMatchFound {
						// should we do by prefix length, nah...
						// choose to include by default
						newColumns[rowField] = nil
					} else if exactFilterMatchFound {
						continue NextRow
					} else {
						newColumns[rowField] = nil
					}
				} else if prefixMatchFound || exactFilterMatchFound {
					continue NextRow
				}
			}

			newRow.Hash[rowField] = value
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	ret.Columns = PreserveColumnOrder(ret, newColumns)

	return ret, nil
}

func PreserveColumnOrder(table *Table, newColumns map[FieldName]interface{}) []FieldName {
	seenRetColumns := map[FieldName]interface{}{}
	retColumns := []FieldName{}

	// preserve previous columns order as best as possible
	for _, column := range table.Columns {
		if _, ok := newColumns[column]; ok {
			retColumns = append(retColumns, column)
			seenRetColumns[column] = nil
		}
	}
	for key := range newColumns {
		if _, ok := seenRetColumns[key]; !ok {
			retColumns = append(retColumns, key)
			seenRetColumns[key] = nil
		}
	}
	return retColumns
}

type FlattenObjectMiddleware struct {
}

func NewFlattenObjectMiddleware() *FlattenObjectMiddleware {
	return &FlattenObjectMiddleware{}
}

func (fom *FlattenObjectMiddleware) Process(table *Table) (*Table, error) {
	ret := &Table{
		Columns: []FieldName{},
		Rows:    []Row{},
	}

	newColumns := map[FieldName]interface{}{}

	for _, row := range table.Rows {
		values := row.GetValues()
		newValues := FlattenMapIntoColumns(values)
		newRow := SimpleRow{
			Hash: newValues,
		}

		for key := range newValues {
			newColumns[key] = nil
		}
		ret.Rows = append(ret.Rows, &newRow)
	}

	ret.Columns = PreserveColumnOrder(table, newColumns)

	return ret, nil
}

func FlattenMapIntoColumns(rows MapRow) MapRow {
	ret := MapRow{}

	for key, value := range rows {
		switch v := value.(type) {
		case MapRow:
			for k, v := range FlattenMapIntoColumns(v) {
				ret[fmt.Sprintf("%s.%s", key, k)] = v
			}
		default:
			ret[key] = v
		}
	}

	return ret
}

type PreserveColumnOrderMiddleware struct {
	columns []FieldName
}

func NewPreserveColumnOrderMiddleware(columns []FieldName) *PreserveColumnOrderMiddleware {
	return &PreserveColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *PreserveColumnOrderMiddleware) Process(table *Table) (*Table, error) {
	columnHash := map[FieldName]interface{}{}
	for _, column := range scm.columns {
		columnHash[column] = nil
	}

	table.Columns = PreserveColumnOrder(table, columnHash)
	return table, nil
}

type ReorderColumnOrderMiddleware struct {
	columns []FieldName
}

func NewReorderColumnOrderMiddleware(columns []FieldName) *ReorderColumnOrderMiddleware {
	return &ReorderColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *ReorderColumnOrderMiddleware) Process(table *Table) (*Table, error) {
	existingColumns := map[FieldName]interface{}{}
	for _, column := range table.Columns {
		existingColumns[column] = nil
	}

	seenColumns := map[FieldName]interface{}{}
	newColumns := []FieldName{}

	for _, column := range scm.columns {
		if strings.HasSuffix(column, ".") {
			for _, existingColumn := range table.Columns {
				if strings.HasPrefix(existingColumn, column) {
					if _, ok := seenColumns[existingColumn]; !ok {
						newColumns = append(newColumns, existingColumn)
						seenColumns[existingColumn] = nil
					}
				}
			}
		} else {
			if _, ok := seenColumns[column]; !ok {
				if _, ok := existingColumns[column]; ok {
					newColumns = append(newColumns, column)
					seenColumns[column] = nil
				}
			}

		}
	}

	table.Columns = newColumns

	return table, nil
}

type SortColumnsMiddleware struct {
}

func NewSortColumnsMiddleware() *SortColumnsMiddleware {
	return &SortColumnsMiddleware{}
}

func (scm *SortColumnsMiddleware) Process(table *Table) (*Table, error) {
	sort.Strings(table.Columns)
	return table, nil
}

type SQLiteOutputFormatter struct {
	table       *Table
	middlewares []TableMiddleware
}
