package table

import (
	"github.com/go-go-golems/glazed/pkg/types"
)

type AddFieldMiddleware struct {
	Fields map[string]string
}

func NewAddFieldMiddleware(fields map[string]string) *AddFieldMiddleware {
	return &AddFieldMiddleware{Fields: fields}
}

func (a *AddFieldMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: table.Columns,
		Rows:    make([]types.Row, 0),
	}

	for _, row := range table.Rows {
		values := row.GetValues()
		newValues := make(map[types.FieldName]types.GenericCellValue)
		for key, value := range values {
			newValues[key] = value
		}
		for key, value := range a.Fields {
			newValues[key] = value
		}
		ret.Rows = append(ret.Rows, &types.SimpleRow{Hash: newValues})
	}

	for key := range a.Fields {
		ret.Columns = append(ret.Columns, key)
	}

	return ret, nil
}
