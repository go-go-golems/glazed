package compare

import (
	"testing"
)

func TestIsOfNumberType(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{1, true},
		{1.0, true},
		// try all int types
		{int(1), true},
		{int8(1), true},
		{int16(1), true},
		{int32(1), true},
		{int64(1), true},
		// try all uint types
		{uint(1), true},
		{uint8(1), true},
		{uint16(1), true},
		{uint32(1), true},
		{uint64(1), true},
		// try all float types
		{float32(1), true},
		{float64(1), true},

		{"1", false},

		// try list and map
		{[]int{1, 2, 3}, false},
		{map[string]int{"1": 1, "2": 2}, false},
	}

	for _, test := range tests {
		output := IsOfNumberType(test.input)
		if output != test.expected {
			t.Errorf("IsOfNumberType(%v) = %v, expected %v", test.input, output, test.expected)
		}
	}
}

func TestIsLowerThan(t *testing.T) {
	tests := []struct {
		a        interface{}
		b        interface{}
		expected bool
	}{
		{1, 2, true},
		{2, 1, false},
		{1, 1, false},
		{1.0, 2.0, true},
		{2.0, 1.0, false},
		{1.0, 1.0, false},
		{1.0, 2, true},
		{2.0, 1, false},
		{1.0, 1, false},
		{1, 2.0, true},
		{2, 1.0, false},
		{1, 1.0, false},
		{"1", "2", true},
		{"2", "1", false},
		{"1", "1", false},
		// alphabetical strings
		{"a", "b", true},
		{"b", "a", false},
		{"a", "a", false},
		{"aa", "ab", true},
		{"ab", "aa", false},
		{[]int{1, 2, 3}, []int{1, 2, 3}, false},
		{map[string]int{"1": 1, "2": 2}, map[string]int{"1": 1, "2": 2}, false},
	}

	for _, test := range tests {
		output := IsLowerThan(test.a, test.b)
		if output != test.expected {
			t.Errorf("IsLowerThan(%v, %v) = %v, expected %v", test.a, test.b, output, test.expected)
		}
	}

}
