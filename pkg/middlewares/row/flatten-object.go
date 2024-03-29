package row

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type FlattenObjectMiddleware struct {
}

var _ middlewares.RowMiddleware = (*FlattenObjectMiddleware)(nil)

func (fom *FlattenObjectMiddleware) Close(ctx context.Context) error {
	return nil
}

func NewFlattenObjectMiddleware() *FlattenObjectMiddleware {
	return &FlattenObjectMiddleware{}
}

func (fom *FlattenObjectMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := FlattenRow(row)
	return []types.Row{newRow}, nil
}

func FlattenRow(row types.Row) types.Row {
	ret := types.NewRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		key, value := pair.Key, pair.Value
		switch v := value.(type) {
		case map[string]interface{}:
			childRow := types.NewRowFromMap(v)
			newColumns := FlattenRow(childRow)
			for pair_ := newColumns.Oldest(); pair_ != nil; pair_ = pair_.Next() {
				k, v := pair_.Key, pair_.Value
				ret.Set(fmt.Sprintf("%s.%s", key, k), v)
			}
		case *orderedmap.OrderedMap[string, string]:
			childRow := types.NewRow()
			for pair_ := v.Oldest(); pair_ != nil; pair_ = pair_.Next() {
				k, v := pair_.Key, pair_.Value
				childRow.Set(k, v)
			}
			newColumns := FlattenRow(childRow)
			for pair_ := newColumns.Oldest(); pair_ != nil; pair_ = pair_.Next() {
				k, v := pair_.Key, pair_.Value
				ret.Set(fmt.Sprintf("%s.%s", key, k), v)
			}

		case types.Row:
			newColumns := FlattenRow(v)
			for pair_ := newColumns.Oldest(); pair_ != nil; pair_ = pair_.Next() {
				k, v := pair_.Key, pair_.Value
				ret.Set(fmt.Sprintf("%s.%s", key, k), v)
			}
		default:
			ret.Set(key, v)
		}
	}

	return ret
}
