package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSkipLimitMiddleware_Process(t *testing.T) {
	tests := []struct {
		name   string
		skip   int
		limit  int
		rows   []types.Row
		expect []types.Row
	}{
		{
			name:  "NoSkipNoLimit",
			skip:  0,
			limit: 0,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
			},
		},
		{
			name:  "SkipNoLimit",
			skip:  1,
			limit: 0,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
		},
		{
			name:  "NoSkipLimit",
			skip:  0,
			limit: 1,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
			},
		},
		{
			name:  "SkipLimit",
			skip:  2,
			limit: 2,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
				types.NewRow(
					types.MRP("field1", "value5"),
					types.MRP("field2", "value6"),
				),
				types.NewRow(
					types.MRP("field1", "value7"),
					types.MRP("field2", "value8"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value5"),
					types.MRP("field2", "value6"),
				),
				types.NewRow(
					types.MRP("field1", "value7"),
					types.MRP("field2", "value8"),
				),
			},
		},
		{
			name:  "NegativeSkip",
			skip:  -2,
			limit: 0,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
		},
		{
			name:  "NegativeLimit",
			skip:  0,
			limit: -2,
			rows: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
			expect: []types.Row{
				types.NewRow(
					types.MRP("field1", "value1"),
					types.MRP("field2", "value2"),
				),
				types.NewRow(
					types.MRP("field1", "value3"),
					types.MRP("field2", "value4"),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipLimitMiddleware := &SkipLimitMiddleware{Skip: tt.skip, Limit: tt.limit}
			var got []map[string]interface{}
			for _, row := range tt.rows {
				rows, err := skipLimitMiddleware.Process(context.Background(), row)
				assert.NoError(t, err)
				for _, row_ := range rows {
					got = append(got, types.RowToMap(row_))
				}
			}

			var expectMaps []map[string]interface{}
			for _, row := range tt.expect {
				expectMaps = append(expectMaps, types.RowToMap(row))
			}
			assert.Equal(t, expectMaps, got)
		})
	}
}
