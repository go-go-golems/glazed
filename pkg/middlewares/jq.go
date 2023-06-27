package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
)

type JqObjectMiddleware struct {
	expression string
	query      *gojq.Query
}

func NewJqObjectMiddleware(
	expression string,
) (*JqObjectMiddleware, error) {
	ret := &JqObjectMiddleware{
		expression: expression,
	}

	if expression != "" {
		query, err := gojq.Parse(expression)
		if err != nil {
			return nil, err
		}

		ret.query = query
	}

	return ret, nil
}

func (jqm *JqObjectMiddleware) Process(
	object types.MapRow,
) ([]types.MapRow, error) {
	ret := []types.MapRow{}

	if jqm.query != nil {
		// TODO(manuel, 2023-06-25) Transform to map before passing to jq
		m := map[string]interface{}{}
		for pair := object.Oldest(); pair != nil; pair = pair.Next() {
			m[pair.Key] = pair.Value
		}
		iter := jqm.query.Run(m)

		for {
			v, ok := iter.Next()
			if !ok {
				break
			}

			switch v_ := v.(type) {
			case error:
				return nil, v_

			case []interface{}:
				for _, v := range v_ {
					switch v_ := v.(type) {
					case error:
						return nil, v_
					case map[string]interface{}:
						ret = append(ret, types.NewMapRowFromMap(v_))
					case types.MapRow:
						ret = append(ret, v_)
					default:
						return nil, errors.Errorf("Expected object, got %T", v)
					}
				}

				continue
			case types.MapRow:
				ret = append(ret, v_)

			case map[string]interface{}:
				ret = append(ret, types.NewMapRowFromMap(v_))

			}
		}
	} else {
		ret = append(ret, object)
	}

	return ret, nil
}

type JqTableMiddleware struct {
	fieldExpressions map[types.FieldName]string
	fieldQueries     map[types.FieldName]*gojq.Query
}

func NewJqTableMiddleware(
	fieldExpressions map[types.FieldName]string,
) (*JqTableMiddleware, error) {
	ret := &JqTableMiddleware{
		fieldExpressions: fieldExpressions,
		fieldQueries:     map[types.FieldName]*gojq.Query{},
	}

	for columnName, fieldExpression := range fieldExpressions {
		query, err := gojq.Parse(fieldExpression)
		if err != nil {
			return nil, err
		}
		ret.fieldQueries[columnName] = query
	}

	return ret, nil
}

func (jqm *JqTableMiddleware) Process(table *types.Table) (*types.Table, error) {
	ret := &types.Table{
		Columns: []types.FieldName{},
		Rows:    []types.Row{},
	}

	ret.Columns = append(ret.Columns, table.Columns...)

	for _, row := range table.Rows {
		values := row.GetValues()
		newRow := types.SimpleRow{
			Hash: types.NewMapRow(),
		}

		for pair := values.Oldest(); pair != nil; pair = pair.Next() {
			rowField, value := pair.Key, pair.Value
			query, ok := jqm.fieldQueries[rowField]
			if !ok {
				newRow.Hash.Set(rowField, value)
				continue
			}

			// TODO(manuel, 2023-03-06) Support generating multiple rows out of jq field queries
			//
			// See https://github.com/go-go-golems/glazed/issues/203
			//
			// currently, we only support single value returning queries.
			// in the future, we could image individual rows being "flattened"
			// out into multiple rows, but that will come later

			iter := query.Run(value)
			v, ok := iter.Next()
			if ok {
				if err, ok := v.(error); ok {
					return nil, err
				}

				newRow.Hash.Set(rowField, v)
			}
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	return ret, nil
}
