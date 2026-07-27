package row

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// OutputFieldsMiddleware projects rows onto an explicit ordered list of fields.
// Missing fields are omitted. An empty field list passes rows through unchanged.
type OutputFieldsMiddleware struct {
	fields []types.FieldName
}

var _ middlewares.RowMiddleware = (*OutputFieldsMiddleware)(nil)

func NewOutputFieldsMiddleware(fields ...string) *OutputFieldsMiddleware {
	ret := &OutputFieldsMiddleware{
		fields: make([]types.FieldName, 0, len(fields)),
	}
	for _, field := range fields {
		ret.fields = append(ret.fields, types.FieldName(field))
	}
	return ret
}

func (m *OutputFieldsMiddleware) Process(_ context.Context, input types.Row) ([]types.Row, error) {
	if len(m.fields) == 0 {
		return []types.Row{input}, nil
	}

	output := types.NewRow()
	for _, field := range m.fields {
		if value, ok := input.Get(field); ok {
			output.Set(field, value)
		}
	}
	return []types.Row{output}, nil
}

func (m *OutputFieldsMiddleware) Close(context.Context) error {
	return nil
}
