package middlewares

import (
	"bytes"
	"dd-cli/pkg/types"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

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

func (ffm *FieldsFilterMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	// how do we keep order here
	newColumns := map[types.FieldName]interface{}{}

	if len(ffm.fields) == 0 && len(ffm.filters) == 0 {
		return table, nil
	}

	for _, row := range table.Rows {
		values := row.GetValues()
		newRow := types.SimpleRow{
			Hash: map[types.FieldName]types.GenericCellValue{},
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
				} else if exactFilterMatchFound {
					continue NextRow
				} else if len(ffm.fields) == 0 {
					newColumns[rowField] = nil
				}
			}

			newRow.Hash[rowField] = value
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	ret.Columns = PreserveColumnOrder(ret, newColumns)

	return ret, nil
}

func PreserveColumnOrder(table *types.Table, newColumns map[types.FieldName]interface{}) []types.FieldName {
	seenRetColumns := map[types.FieldName]interface{}{}
	retColumns := []types.FieldName{}

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

func (fom *FlattenObjectMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	newColumns := map[types.FieldName]interface{}{}

	for _, row := range table.Rows {
		values := row.GetValues()
		newValues := FlattenMapIntoColumns(values)
		newRow := types.SimpleRow{
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

func FlattenMapIntoColumns(rows types.MapRow) types.MapRow {
	ret := types.MapRow{}

	for key, value := range rows {
		switch v := value.(type) {
		case types.MapRow:
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
	columns []types.FieldName
}

func NewPreserveColumnOrderMiddleware(columns []types.FieldName) *PreserveColumnOrderMiddleware {
	return &PreserveColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *PreserveColumnOrderMiddleware) Process(table *types.Table) (*types.Table, error) {
	columnHash := map[types.FieldName]interface{}{}
	for _, column := range scm.columns {
		columnHash[column] = nil
	}

	table.Columns = PreserveColumnOrder(table, columnHash)
	return table, nil
}

type ReorderColumnOrderMiddleware struct {
	columns []types.FieldName
}

func NewReorderColumnOrderMiddleware(columns []types.FieldName) *ReorderColumnOrderMiddleware {
	return &ReorderColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *ReorderColumnOrderMiddleware) Process(table *types.Table) (*types.Table, error) {
	existingColumns := map[types.FieldName]interface{}{}
	for _, column := range table.Columns {
		existingColumns[column] = nil
	}

	seenColumns := map[types.FieldName]interface{}{}
	newColumns := []types.FieldName{}

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

func (scm *SortColumnsMiddleware) Process(table *types.Table) (*types.Table, error) {
	sort.Strings(table.Columns)
	return table, nil
}

type RowGoTemplateMiddleware struct {
	templates map[types.FieldName]*template.Template
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
func NewRowGoTemplateMiddleware(templateStrings map[types.FieldName]string) (*RowGoTemplateMiddleware, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	templates := map[types.FieldName]*template.Template{}
	for columnName, templateString := range templateStrings {
		tmpl, err := template.New("row").Funcs(funcMap).Parse(templateString)
		if err != nil {
			return nil, err
		}
		templates[columnName] = tmpl
	}

	return &RowGoTemplateMiddleware{
		templates: templates,
	}, nil
}

func (rgtm *RowGoTemplateMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	columnRenames := map[types.FieldName]types.FieldName{}
	existingColumns := map[types.FieldName]interface{}{}
	newColumns := map[types.FieldName]interface{}{}

	for _, row := range table.Rows {
		newRow := types.SimpleRow{
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

		for columnName, tmpl := range rgtm.templates {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, templateValues)
			if err != nil {
				return nil, err
			}

			// we need to handle the fact that some rows might not have all the keys, and thus
			// avoid counting columns as existing twice
			if _, ok := newColumns[columnName]; !ok {
				if _, ok := newRow.Hash[columnName]; ok {
					newColumns[columnName] = nil
				}
			} else {
				if _, ok := newRow.Hash[columnName]; !ok {
					existingColumns[columnName] = nil
				}
			}
			newRow.Hash[columnName] = buf.String()
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	// I guess another solution would just be to remove the duplicates once we are done...
	for columnName := range newColumns {
		if _, ok := existingColumns[columnName]; !ok {
			ret.Columns = append(table.Columns, columnName)
		}
	}

	return ret, nil
}
