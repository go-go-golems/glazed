package codegen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

// Helper function to convert jen.Code to string for easy comparison
// It's a bit oblique because you can't just render arbitrary jen.Code
func codeToString(code jen.Code) string {
	j := jen.NewFile("test")
	j.NoFormat = true
	j.Add(code)
	ret := j.GoString()
	return strings.TrimPrefix(ret, "package test\n\n\n")
}

func TestTypeToJen(t *testing.T) {
	tests := []struct {
		name     string
		input    reflect.Type
		expected string
		wantErr  bool
	}{
		{
			name:     "string type",
			input:    reflect.TypeOf(""),
			expected: "string",
			wantErr:  false,
		},
		{
			name:     "int type",
			input:    reflect.TypeOf(0),
			expected: "int",
			wantErr:  false,
		},
		{
			name:     "uint type",
			input:    reflect.TypeOf(uint(0)),
			expected: "uint",
			wantErr:  false,
		},
		{
			name:     "float32 type",
			input:    reflect.TypeOf(float32(0)),
			expected: "float64", // jen package might not differentiate between float32 and float64
			wantErr:  false,
		},
		{
			name:     "bool type",
			input:    reflect.TypeOf(true),
			expected: "bool",
			wantErr:  false,
		},
		{
			name:     "pointer to int",
			input:    reflect.TypeOf(new(int)),
			expected: "* int",
			wantErr:  false,
		},
		{
			name:     "slice of strings",
			input:    reflect.TypeOf([]string{}),
			expected: "[] string",
			wantErr:  false,
		},
		{
			name: "slice of slice of strings",
			input: func() reflect.Type {
				s := []string{}
				return reflect.TypeOf([][]string{s})
			}(),
			expected: "[] [] string",
			wantErr:  false,
		},
		{
			name:     "map from string to int",
			input:    reflect.TypeOf(map[string]int{}),
			expected: "map[string] int",
			wantErr:  false,
		},
		{
			name: "interface{}",
			input: func() reflect.Type {
				s := map[string]interface{}{}
				t := reflect.TypeOf(s)
				return t.Elem()
			}(),
			expected: "interface{}",
			wantErr:  false,
		},
		{
			name:     "map from string to interface",
			input:    reflect.TypeOf(map[string]interface{}{}),
			expected: "map[string] interface{}",
			wantErr:  false,
		},
		{
			name: "map from string to slice of int",
			input: func() reflect.Type {
				s := []int{}
				return reflect.TypeOf(map[string][]int{"": s})
			}(),
			expected: "map[string] [] int",
			wantErr:  false,
		},
		{
			name:     "struct type",
			input:    reflect.TypeOf(struct{ Name string }{}),
			expected: "struct{\nName string\n}",
			wantErr:  false,
		},
		{
			name: "struct type with pointer to struct",
			input: func() reflect.Type {
				type T struct {
					Name string
				}
				return reflect.TypeOf(struct{ Name *T }{})
			}(),
			expected: "struct{\nName * struct{\nName string\n}\n}",
			wantErr:  false,
		},
		// Error case example
		{
			name:     "unsupported type",
			input:    reflect.TypeOf(make(chan int)),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TypeToJen(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypeToJen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got == nil {
				panic("got nil jen.Code")
			}
			assert.Equal(t, tt.expected, codeToString(got))
		})
	}
}

type StructTest struct {
	Name string
}

func TestLiteralToJen(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			"int",
			1,
			"1",
			false,
		},
		{
			"int32",
			int32(1),
			"int32(1)",
			false,
		},
		{
			"uint64",
			uint64(1),
			"uint64(0x1)",
			false,
		},
		{
			"float",
			1.0,
			"1.0",
			false,
		},
		{
			"float32",
			float32(1.2),
			"float32(1.2)",
			false,
		},
		{
			"string",
			"1",
			"\"1\"",
			false,
		},
		{
			name:     "bool type",
			input:    true,
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "slice of strings",
			input:    []string{},
			expected: "[] string {}",
			wantErr:  false,
		},
		{name: "slice of strings (1 value)",
			input:    []string{"a"},
			expected: "[] string {\"a\"}",
			wantErr:  false,
		},
		{
			name:     "slice of strings (3 values)",
			input:    []string{"a", "b", "c"},
			expected: "[] string {\"a\",\"b\",\"c\"}",
			wantErr:  false,
		},
		{
			name:     "slice of slice of strings",
			input:    [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
			expected: "[] [] string {[] string {\"a\",\"b\",\"c\"},[] string {\"d\",\"e\",\"f\"}}",
			wantErr:  false,
		},
		{
			name:     "map from string to int (empty)",
			input:    map[string]int{},
			expected: "map[string] int {}",
			wantErr:  false,
		},
		{
			name:     "map from string to int (1 value)",
			input:    map[string]int{"a": 1},
			expected: "map[string] int {\"a\":1}",
			wantErr:  false,
		},
		{
			name:     "map from string to int (3 values)",
			input:    map[string]int{"a": 1, "b": 2, "c": 3},
			expected: "map[string] int {\n\"a\":1,\n\"b\":2,\n\"c\":3,\n}",
			wantErr:  false,
		},
		{
			name:     "interface{}",
			input:    map[string]interface{}{},
			expected: "map[string] interface{} {}",
			wantErr:  false,
		},
		{
			name:     "map from string to interface",
			input:    map[string]interface{}{"a": 1, "b": "2", "c": true},
			expected: "map[string] interface{} {\n\"a\":1,\n\"b\":\"2\",\n\"c\":true,\n}",
			wantErr:  false,
		},
		{
			name:     "map from string to slice of int",
			input:    map[string][]int{"a": {1, 2, 3}, "b": {4, 5, 6}},
			expected: "map[string] [] int {\n\"a\":[] int {1,2,3},\n\"b\":[] int {4,5,6},\n}",
			wantErr:  false,
		},
		{
			name: "struct type with name",
			input: func() StructTest {
				return StructTest{Name: "test"}
			}(),
			expected: "StructTest {Name : \"test\"}",
		},
		{
			name:     "struct type",
			input:    struct{ Name string }{Name: "test"},
			expected: "struct{\nName string\n} {Name : \"test\"}",
			wantErr:  false,
		},
		{
			name:     "struct type with pointer to struct",
			input:    struct{ Name *struct{ Name string } }{Name: &struct{ Name string }{Name: "test"}},
			expected: "struct{\nName * struct{\nName string\n}\n} {Name : & struct{\nName string\n} {Name : \"test\"}}",
			wantErr:  false,
		},
		// Error case example
		{
			name:    "unsupported type",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LiteralToJen(reflect.ValueOf(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("LiteralToJen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got == nil {
				panic("got nil jen.Code")
			}
			assert.Equal(t, tt.expected, codeToString(got))
		})
	}

}
