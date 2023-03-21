package table

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"regexp"
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

	ret.Columns = PreserveColumnOrder(table.Columns, newColumns)

	return ret, nil
}

func PreserveColumnOrder(oldColumns []types.FieldName, newColumns map[types.FieldName]interface{}) []types.FieldName {
	seenRetColumns := map[types.FieldName]interface{}{}
	retColumns := []types.FieldName{}

	// preserve previous columns order as best as possible
	for _, column := range oldColumns {
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

type RemoveNullsMiddleware struct {
}

func NewRemoveNullsMiddleware() *RemoveNullsMiddleware {
	return &RemoveNullsMiddleware{}
}

func (rnm *RemoveNullsMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: table.Columns,
		Rows:    []types.Row{},
	}

	for _, row := range table.Rows {
		values := row.GetValues()
		newRow := types.SimpleRow{
			Hash: map[types.FieldName]types.GenericCellValue{},
		}

		for key, value := range values {
			if value != nil {
				newRow.Hash[key] = value
			}
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	return ret, nil
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

	ret.Columns = PreserveColumnOrder(table.Columns, newColumns)

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

	table.Columns = PreserveColumnOrder(table.Columns, columnHash)
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
//
// TODO(manuel, 2023-02-02) Add support for passing in custom funcmaps
// See #110 https://github.com/go-go-golems/glazed/issues/110
func NewRowGoTemplateMiddleware(
	templateStrings map[types.FieldName]string,
	renameSeparator string) (*RowGoTemplateMiddleware, error) {

	templates := map[types.FieldName]*template.Template{}
	for columnName, templateString := range templateStrings {
		tmpl, err := template.New("row").
			Funcs(sprig.TxtFuncMap()).
			Funcs(templating.TemplateFuncs).
			Parse(templateString)
		if err != nil {
			return nil, err
		}
		templates[columnName] = tmpl
	}

	return &RowGoTemplateMiddleware{
		templates:       templates,
		RenameSeparator: renameSeparator,
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

	for _, columnName := range table.Columns {
		existingColumns[columnName] = nil
		ret.Columns = append(ret.Columns, columnName)
	}

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
			} else {
				columnRenames[key] = key
			}
			newKey := columnRenames[key]
			templateValues[newKey] = value
		}
		templateValues["_row"] = templateValues

		for columnName, tmpl := range rgtm.templates {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, templateValues)
			if err != nil {
				return nil, err
			}
			s := buf.String()

			// we need to handle the fact that some rows might not have all the keys, and thus
			// avoid counting columns as existing twice
			if _, ok := newColumns[columnName]; !ok {
				newColumns[columnName] = nil
				ret.Columns = append(ret.Columns, columnName)
			}
			newRow.Hash[columnName] = s
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	return ret, nil
}

type RenameColumnMiddleware struct {
	Renames map[types.FieldName]types.FieldName
	// orderedmap *regexp.Regexp -> string
	RegexpRenames RegexpReplacements
}

func NewFieldRenameColumnMiddleware(renames map[types.FieldName]types.FieldName) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:       renames,
		RegexpRenames: RegexpReplacements{},
	}
}

func NewRegexpRenameColumnMiddleware(renames RegexpReplacements) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:       map[types.FieldName]types.FieldName{},
		RegexpRenames: renames,
	}
}

func NewRenameColumnMiddleware(
	renames map[types.FieldName]types.FieldName,
	regexpRenames RegexpReplacements,
) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:       renames,
		RegexpRenames: regexpRenames,
	}
}

type RegexpReplacement struct {
	Regexp      *regexp.Regexp
	Replacement string
}

type RegexpReplacements []*RegexpReplacement

func (rr *RegexpReplacements) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", value.Kind)
	}
	*rr = RegexpReplacements{}
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]

		if key.Kind != yaml.ScalarNode {
			return fmt.Errorf("expected a scalar node, got %v", key.Kind)
		}
		if val.Kind != yaml.ScalarNode {
			return fmt.Errorf("expected a scalar node, got %v", val.Kind)
		}
		re, err := regexp.Compile(key.Value)
		if err != nil {
			return err
		}
		*rr = append(*rr, &RegexpReplacement{
			Regexp:      re,
			Replacement: val.Value,
		})
	}

	return nil
}

type ColumnMiddlewareConfig struct {
	FieldRenames map[types.FieldName]types.FieldName `yaml:"renames"`
	// FIXME regex renames actually need to ordered
	RegexpRenames RegexpReplacements `yaml:"regexpRenames"`
}

func NewRenameColumnMiddlewareFromYAML(decoder *yaml.Decoder) (*RenameColumnMiddleware, error) {
	var config ColumnMiddlewareConfig
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return NewRenameColumnMiddleware(config.FieldRenames, config.RegexpRenames), nil
}

func (r *RenameColumnMiddleware) RenameColumns(
	columns []types.FieldName,
) ([]types.FieldName, map[types.FieldName]types.FieldName) {
	var columnRenames = map[types.FieldName]types.FieldName{}
	var renamedColumns = map[types.FieldName]interface{}{}
	var orderedColumns = []types.FieldName{}

	// first, we create a map of all the original columns to the new columns
columnLoop:
	for _, column := range columns {
		// we run string renames first, as we consider them more exhaustive matches
		for match, rename := range r.Renames {
			if column == match {
				if _, ok := renamedColumns[rename]; !ok {
					orderedColumns = append(orderedColumns, rename)
					renamedColumns[rename] = nil
				}
				columnRenames[match] = rename
				continue columnLoop
			}
		}

		for _, rr := range r.RegexpRenames {
			rename := rr.Regexp.ReplaceAllString(column, rr.Replacement)
			if rename != column {
				if _, ok := renamedColumns[rename]; !ok {
					orderedColumns = append(orderedColumns, rename)
					renamedColumns[rename] = nil
				}
				columnRenames[column] = rename
				continue columnLoop
			}
		}

		// check if we already had a rename
		if _, ok := renamedColumns[column]; !ok {
			columnRenames[column] = column
			renamedColumns[column] = nil
			orderedColumns = append(orderedColumns, column)
		}
	}

	return orderedColumns, columnRenames
}

func (r *RenameColumnMiddleware) Process(table *types.Table) (*types.Table, error) {
	orderedColumns, renamedColumns := r.RenameColumns(table.Columns)

	ret := &types.Table{
		Columns: orderedColumns,
		Rows:    []types.Row{},
	}

	// TODO(2022-12-28, manuel): we need to formalize the copy/clone behaviour of middlewares
	// This is wrt to mutability, and also how things can be used in a streaming context
	// I wonder if immutability is really necessary, or if the whole thing by design meshes
	// well with just passing references to previous rows wrt efficiency.
	// See: https://github.com/go-go-golems/glazed/issues/74

	// we must now go through every row, and rename the hash keys.
	// this really requires us to copy most of the maps.
	// whatever, we'll address efficient renames later
	for _, row := range table.Rows {
		newRow := &types.SimpleRow{
			Hash: map[types.FieldName]interface{}{},
		}
		values := row.GetValues()
		for key, value := range values {
			newKey, ok := renamedColumns[key]
			if !ok {
				// skip, it means columns were overwritten in the rename
				continue
			}
			newRow.Hash[newKey] = value
		}
		ret.Rows = append(ret.Rows, newRow)
	}

	return ret, nil
}
