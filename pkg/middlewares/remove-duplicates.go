package middlewares

import "github.com/go-go-golems/glazed/pkg/types"

type RemoveDuplicatesMiddleware struct {
	columns []string
}

func NewRemoveDuplicatesMiddleware(columns ...string) *RemoveDuplicatesMiddleware {
	return &RemoveDuplicatesMiddleware{
		columns: columns,
	}
}

func (r *RemoveDuplicatesMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: table.Columns,
		Rows:    make([]types.Row, 0),
	}

	var previousRowValues map[types.FieldName]types.GenericCellValue

	for _, row := range table.Rows {
		values := row.GetValues()
		if previousRowValues != nil {
			// check if the values are the same
			same := true
			for _, column := range r.columns {
				if values[column] != previousRowValues[column] {
					same = false
					break
				}
			}
			if same {
				continue
			}
		}
		ret.Rows = append(ret.Rows, row)
		previousRowValues = values
	}

	return ret, nil
}
