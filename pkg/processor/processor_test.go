package processor

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters/json"
	"testing"
)

// BenchmarkSimpleGlazeProcessor benchmarks the simple glaze processor with no middlewares, and CSV output.
func BenchmarkSimpleGlazeProcessor(b *testing.B) {
	ctx := context.Background()
	gp := NewSimpleGlazeProcessor()
	data := map[string]interface{}{
		"name":    "Manuel Manuel",
		"age":     30,
		"job":     "Software Engineer",
		"city":    "Lisbon",
		"country": "Portugal",
		"email":   "manuel@gogogogogogogogo.gogo",
		"phone":   "+351 123 456 789",
	}
	for i := 0; i < b.N; i++ {
		_ = gp.ProcessInputObject(ctx, data)
	}
	s, _ := gp.OutputFormatter().Output(ctx)
	_ = s
}

func BenchmarkGlazeProcessor_JSONOutputFormatter(b *testing.B) {
	ctx := context.Background()

	gp := NewGlazeProcessor(json.NewOutputFormatter())
	data := map[string]interface{}{
		"name":    "Manuel Manuel",
		"age":     30,
		"job":     "Software Engineer",
		"city":    "Lisbon",
		"country": "Portugal",
		"email":   "manuel@gogogogogogogogo.gogo",
		"phone":   "+351 123 456 789",
	}
	for i := 0; i < b.N; i++ {
		_ = gp.ProcessInputObject(ctx, data)
	}
	s, _ := gp.OutputFormatter().Output(ctx)
	_ = s
}
