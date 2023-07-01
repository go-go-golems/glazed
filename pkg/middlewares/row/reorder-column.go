package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"strings"
)

type ReorderColumnOrderMiddleware struct {
	columns []types.FieldName
}

func (scm *ReorderColumnOrderMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewReorderColumnOrderMiddleware(columns []types.FieldName) *ReorderColumnOrderMiddleware {
	return &ReorderColumnOrderMiddleware{
		columns: columns,
	}
}

func (scm *ReorderColumnOrderMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := types.NewRow()

	existingColumns := types.GetFields(row)
	seenColumns := map[types.FieldName]interface{}{}

	for _, column := range scm.columns {
		if strings.HasSuffix(column, ".") {
			for _, existingColumn := range existingColumns {
				if strings.HasPrefix(existingColumn, column) {
					if _, ok := seenColumns[existingColumn]; !ok {
						v, _ := row.Get(existingColumn)
						newRow.Set(existingColumn, v)
						seenColumns[existingColumn] = nil
					}
				}
			}
		} else {
			if value, ok := row.Get(column); ok {
				if _, ok := seenColumns[column]; !ok {
					newRow.Set(column, value)
					seenColumns[column] = nil
				}
			}
		}
	}

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		if _, ok := seenColumns[key]; !ok {
			newRow.Set(key, value)
			seenColumns[key] = nil
		}
	}

	return []types.Row{newRow}, nil
}
