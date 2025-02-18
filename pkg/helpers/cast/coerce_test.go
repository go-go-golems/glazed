package cast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name           string
	a              interface{}
	b              interface{}
	expectedA      interface{}
	expectedB      interface{}
	expectSuccess  bool
	skipComparison bool // For cases where we can't predict exact values
}

func TestTypeCoerce(t *testing.T) {
	tests := []testCase{
		// Nil cases
		{
			name:          "both nil",
			a:             nil,
			b:             nil,
			expectedA:     nil,
			expectedB:     nil,
			expectSuccess: true,
		},
		{
			name:          "one nil",
			a:             nil,
			b:             42,
			expectSuccess: false,
		},

		// Basic type cases
		{
			name:          "identical types",
			a:             42,
			b:             43,
			expectedA:     42,
			expectedB:     43,
			expectSuccess: true,
		},
		{
			name:          "interface to concrete",
			a:             interface{}(42),
			b:             43,
			expectedA:     42,
			expectedB:     43,
			expectSuccess: true,
		},

		// Numeric coercion cases
		{
			name:          "int to float",
			a:             42,
			b:             42.0,
			expectedA:     42.0,
			expectedB:     42.0,
			expectSuccess: true,
		},
		{
			name:          "different int types",
			a:             int32(42),
			b:             int64(42),
			expectedA:     int64(42),
			expectedB:     int64(42),
			expectSuccess: true,
		},

		// String coercion cases
		{
			name:          "string to interface",
			a:             "test",
			b:             interface{}("test"),
			expectedA:     "test",
			expectedB:     "test",
			expectSuccess: true,
		},
		{
			name:          "number to string",
			a:             42,
			b:             "42",
			expectedA:     "42",
			expectedB:     "42",
			expectSuccess: true,
		},

		// Slice coercion cases
		{
			name:          "identical slices",
			a:             []int{1, 2, 3},
			b:             []int{1, 2, 3},
			expectedA:     []int{1, 2, 3},
			expectedB:     []int{1, 2, 3},
			expectSuccess: true,
		},
		{
			name:          "interface slice to concrete",
			a:             []interface{}{1, 2, 3},
			b:             []int{1, 2, 3},
			expectedA:     []int{1, 2, 3},
			expectedB:     []int{1, 2, 3},
			expectSuccess: true,
		},
		{
			name:          "mixed type slice",
			a:             []interface{}{1, "2", 3.0},
			b:             []string{"1", "2", "3"},
			expectedA:     []string{"1", "2", "3"},
			expectedB:     []string{"1", "2", "3"},
			expectSuccess: true,
		},

		// Map coercion cases
		{
			name:          "identical maps",
			a:             map[string]int{"a": 1, "b": 2},
			b:             map[string]int{"a": 1, "b": 2},
			expectedA:     map[string]int{"a": 1, "b": 2},
			expectedB:     map[string]int{"a": 1, "b": 2},
			expectSuccess: true,
		},
		{
			name:          "interface map to concrete",
			a:             map[string]interface{}{"a": 1, "b": "2"},
			b:             map[string]string{"a": "1", "b": "2"},
			expectedA:     map[string]string{"a": "1", "b": "2"},
			expectedB:     map[string]string{"a": "1", "b": "2"},
			expectSuccess: true,
		},

		// Nested structure cases
		{
			name: "nested maps and slices",
			a: map[string]interface{}{
				"nums": []interface{}{1, 2, 3},
				"map":  map[string]interface{}{"a": 1},
			},
			b: map[string]interface{}{
				"nums": []int{1, 2, 3},
				"map":  map[string]int{"a": 1},
			},
			expectedA: map[string]interface{}{
				"nums": []int{1, 2, 3},
				"map":  map[string]int{"a": 1},
			},
			expectedB: map[string]interface{}{
				"nums": []int{1, 2, 3},
				"map":  map[string]int{"a": 1},
			},
			expectSuccess: true,
		},

		// Error cases
		{
			name:          "incompatible types",
			a:             struct{ name string }{"test"},
			b:             42,
			expectSuccess: false,
		},
		{
			name:          "incompatible slice types",
			a:             []int{1, 2, 3},
			b:             []bool{true, false, true},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coercedA, coercedB, ok := TypeCoerce(tt.a, tt.b)
			assert.Equal(t, tt.expectSuccess, ok, "TypeCoerce success mismatch")

			if tt.expectSuccess && !tt.skipComparison {
				assert.Equal(t, tt.expectedA, coercedA, "coerced value A mismatch")
				assert.Equal(t, tt.expectedB, coercedB, "coerced value B mismatch")
			}
		})
	}
}

func TestCoerceSlices(t *testing.T) {
	tests := []testCase{
		{
			name:          "empty slices",
			a:             []interface{}{},
			b:             []interface{}{},
			expectedA:     []interface{}{},
			expectedB:     []interface{}{},
			expectSuccess: true,
		},
		{
			name:          "different length slices",
			a:             []int{1, 2},
			b:             []int{1},
			expectSuccess: false,
		},
		{
			name:          "mixed numeric types",
			a:             []interface{}{1, int32(2), int64(3)},
			b:             []int64{1, 2, 3},
			expectedA:     []int64{1, 2, 3},
			expectedB:     []int64{1, 2, 3},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coercedA, coercedB, ok := coerceSlices(tt.a, tt.b)
			assert.Equal(t, tt.expectSuccess, ok)

			if tt.expectSuccess && !tt.skipComparison {
				assert.Equal(t, tt.expectedA, coercedA)
				assert.Equal(t, tt.expectedB, coercedB)
			}
		})
	}
}

func TestCoerceMaps(t *testing.T) {
	tests := []testCase{
		{
			name:          "empty maps",
			a:             map[string]interface{}{},
			b:             map[string]interface{}{},
			expectedA:     map[string]interface{}{},
			expectedB:     map[string]interface{}{},
			expectSuccess: true,
		},
		{
			name:          "different key sets",
			a:             map[string]int{"a": 1},
			b:             map[string]int{"b": 2},
			expectSuccess: false,
		},
		{
			name:          "mixed value types",
			a:             map[string]interface{}{"a": 1, "b": "2"},
			b:             map[string]string{"a": "1", "b": "2"},
			expectedA:     map[string]string{"a": "1", "b": "2"},
			expectedB:     map[string]string{"a": "1", "b": "2"},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coercedA, coercedB, ok := coerceMaps(tt.a, tt.b)
			assert.Equal(t, tt.expectSuccess, ok)

			if tt.expectSuccess && !tt.skipComparison {
				assert.Equal(t, tt.expectedA, coercedA)
				assert.Equal(t, tt.expectedB, coercedB)
			}
		})
	}
}

func TestCoerceNumbers(t *testing.T) {
	tests := []testCase{
		{
			name:          "same type integers",
			a:             42,
			b:             43,
			expectedA:     42,
			expectedB:     43,
			expectSuccess: true,
		},
		{
			name:          "int to float",
			a:             42,
			b:             42.0,
			expectedA:     42.0,
			expectedB:     42.0,
			expectSuccess: true,
		},
		{
			name:          "different integer types",
			a:             int32(42),
			b:             int64(42),
			expectedA:     int64(42),
			expectedB:     int64(42),
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coercedA, coercedB, ok := coerceNumbers(tt.a, tt.b)
			assert.Equal(t, tt.expectSuccess, ok)

			if tt.expectSuccess && !tt.skipComparison {
				assert.Equal(t, tt.expectedA, coercedA)
				assert.Equal(t, tt.expectedB, coercedB)
			}
		})
	}
}

func TestCoerceStrings(t *testing.T) {
	tests := []testCase{
		{
			name:          "identical strings",
			a:             "test",
			b:             "test",
			expectedA:     "test",
			expectedB:     "test",
			expectSuccess: true,
		},
		{
			name:          "number to string",
			a:             42,
			b:             "42",
			expectedA:     "42",
			expectedB:     "42",
			expectSuccess: true,
		},
		{
			name:          "invalid string conversion",
			a:             make(chan int),
			b:             "test",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coercedA, coercedB, ok := coerceStrings(tt.a, tt.b)
			assert.Equal(t, tt.expectSuccess, ok)

			if tt.expectSuccess && !tt.skipComparison {
				assert.Equal(t, tt.expectedA, coercedA)
				assert.Equal(t, tt.expectedB, coercedB)
			}
		})
	}
}
