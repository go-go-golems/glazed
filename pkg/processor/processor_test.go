package processor

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	"github.com/go-go-golems/glazed/pkg/types"
	"testing"
)

// BenchmarkSimpleGlazeProcessor benchmarks the simple glaze processor with no middlewares, and CSV output.
func BenchmarkSimpleGlazeProcessor(b *testing.B) {
	ctx := context.Background()
	gp := NewSimpleGlazeProcessor()
	data := types.NewMapRow(
		types.MRP("name", "Manuel Manuel"),
		types.MRP("age", 30),
		types.MRP("job", "Software Engineer"),
		types.MRP("city", "Lisbon"),
		types.MRP("country", "Portugal"),

		types.MRP("email", "manuel@example.com"),
		types.MRP("phone", "+351 123 456 789"),
	)
	for i := 0; i < b.N; i++ {
		_ = gp.ProcessInputObject(ctx, data)
	}
	buf := &bytes.Buffer{}
	_ = gp.OutputFormatter().Output(ctx, buf)
}

func BenchmarkGlazeProcessor_JSONOutputFormatter(b *testing.B) {
	ctx := context.Background()

	gp := NewGlazeProcessor(json.NewOutputFormatter())
	data := types.NewMapRow(
		types.MRP("name", "Manuel Manuel"),
		types.MRP("age", 30),
		types.MRP("job", "Software Engineer"),
		types.MRP("city", "Lisbon"),
		types.MRP("country", "Portugal"),

		types.MRP("email", "manuel@example.com"),
		types.MRP("phone", "+351 123 456 789"),
	)
	for i := 0; i < b.N; i++ {
		_ = gp.ProcessInputObject(ctx, data)
	}
	buf := &bytes.Buffer{}
	_ = gp.OutputFormatter().Output(ctx, buf)
}
