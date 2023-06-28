package table

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/types"
	"sort"
	"strings"
	"text/template"
)

type SortColumnsMiddleware struct {
}

func NewSortColumnsMiddleware() *SortColumnsMiddleware {
	return &SortColumnsMiddleware{}
}

func (scm *SortColumnsMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
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

func (rgtm *RowGoTemplateMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
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
		newRow := row

		templateValues := map[string]interface{}{}

		for pair := newRow.Oldest(); pair != nil; pair = pair.Next() {
			key, value := pair.Key, pair.Value

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
			newRow.Set(columnName, s)
		}

		ret.Rows = append(ret.Rows, newRow)
	}

	return ret, nil
}
