package pkg

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

type TableMiddleware interface {
	// Process transform a single row into potential multiple rows split across multiple tables
	Process(table *Table) (*Table, error)
}

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

	for column := range existingColumns {
		if _, ok := seenColumns[column]; !ok {
			newColumns = append(newColumns, column)
			seenColumns[column] = nil
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

type RowGoTemplateMiddleware struct {
	template   *template.Template
	columnName FieldName
	// this field is used to replace "." in keys before passing them to the template,
	// in order to avoid having to use the `index` template function to access fields
	// that contain a ".", which is frequent due to flattening.
	RenameSeparator string
}

// NewRowGoTemplateMiddleware creates a new RowGoTemplateMiddleware
// which is the simplest go template middleware.
//
// It will render the template for each row and return the result as a new column called with
// the given title.
//
// Because nested objects will be flattened to individual columns using the . separator,
// this will make fields inaccessible to the template. One way around this is to use
// {{ index . "field.subfield" }} in the template. Another is to pass a separator rename
// option.
func NewRowGoTemplateMiddleware(
	columName FieldName,
	templateString string) (*RowGoTemplateMiddleware, error) {
	tmpl, err := template.New("row").Parse(templateString)
	if err != nil {
		return nil, err
	}

	return &RowGoTemplateMiddleware{
		columnName: columName,
		template:   tmpl,
	}, nil
}

func (rgtm *RowGoTemplateMiddleware) Process(table *Table) (*Table, error) {
	ret := &Table{
		Columns: []FieldName{},
		Rows:    []Row{},
	}

	columnRenames := map[FieldName]FieldName{}

	isNewColumn := true

	for _, row := range table.Rows {
		newRow := SimpleRow{
			Hash: row.GetValues(),
		}

		templateValues := map[string]interface{}{}

		for key, value := range newRow.Hash {
			if rgtm.RenameSeparator != "" {
				if _, ok := columnRenames[key]; !ok {
					columnRenames[key] = strings.ReplaceAll(key, ".", rgtm.RenameSeparator)
				}
			}
			templateValues[columnRenames[key]] = value
		}
		templateValues["_row"] = templateValues

		var buf bytes.Buffer
		err := rgtm.template.Execute(&buf, templateValues)
		if err != nil {
			return nil, err
		}

		if _, ok := newRow.Hash[rgtm.columnName]; ok {
			isNewColumn = false
		}
		newRow.Hash[rgtm.columnName] = buf.String()

		ret.Rows = append(ret.Rows, &newRow)
	}

	if isNewColumn {
		ret.Columns = append(table.Columns, rgtm.columnName)
	}

	return ret, nil
}
