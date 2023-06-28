package processor

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

// BenchmarkSimpleGlazeProcessor benchmarks the simple glaze processor with no middlewares, and CSV output.
func BenchmarkSimpleGlazeProcessor(b *testing.B) {
	ctx := context.Background()
	gp, err := NewSimpleGlazeProcessor()
	require.NoError(b, err)
	data := types.NewRow(
		types.MRP("name", "Manuel Manuel"),
		types.MRP("age", 30),
		types.MRP("job", "Software Engineer"),
		types.MRP("city", "Lisbon"),
		types.MRP("country", "Portugal"),

		types.MRP("email", "manuel@example.com"),
		types.MRP("phone", "+351 123 456 789"),
	)
	for i := 0; i < b.N; i++ {
		_ = gp.AddRow(ctx, data)
	}
	buf := &bytes.Buffer{}
	err = gp.Finalize(ctx)
	if err != nil {
		b.Fatal(err)
	}

	_ = gp.OutputFormatter().Output(ctx, gp.GetTable(), buf)
}

func BenchmarkGlazeProcessor_JSONOutputFormatter(b *testing.B) {
	ctx := context.Background()

	gp, err := NewGlazeProcessor(json.NewOutputFormatter())
	require.NoError(b, err)

	data := types.NewRow(
		types.MRP("name", "Manuel Manuel"),
		types.MRP("age", 30),
		types.MRP("job", "Software Engineer"),
		types.MRP("city", "Lisbon"),
		types.MRP("country", "Portugal"),

		types.MRP("email", "manuel@example.com"),
		types.MRP("phone", "+351 123 456 789"),
	)
	for i := 0; i < b.N; i++ {
		_ = gp.AddRow(ctx, data)
	}
	buf := &bytes.Buffer{}
	err = gp.Finalize(ctx)
	require.NoError(b, err)
	_ = gp.OutputFormatter().Output(ctx, gp.GetTable(), buf)
}
