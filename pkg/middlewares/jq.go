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

func (jqm *JqObjectMiddleware) Process(object map[string]interface{}) (map[string]interface{}, error) {
	if jqm.query != nil {
		iter := jqm.query.Run(object)
		// See https://github.com/go-go-golems/glazed/issues/202
		// A middleware should be able to produce multiple outputs rows
		v, ok := iter.Next()
		if !ok {
			return nil, errors.New("no result")
		}

		if err, ok := v.(error); ok {
			return nil, err
		}

		object = v.(map[string]interface{})
	}

	return object, nil
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
			Hash: map[types.FieldName]types.GenericCellValue{},
		}

		for rowField, value := range values {
			query, ok := jqm.fieldQueries[rowField]
			if !ok {
				newRow.Hash[rowField] = value
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

				newRow.Hash[rowField] = v
			}
		}

		ret.Rows = append(ret.Rows, &newRow)
	}

	return ret, nil
}
