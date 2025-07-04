package search

import (
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
)

func TestConverter_SimpleFilters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "type:example",
			expected: "type filter",
		},
		{
			input:    "topic:api",
			expected: "topic filter",
		},
		{
			input:    "flag:verbose",
			expected: "flag filter",
		},
		{
			input:    "command:deploy",
			expected: "command filter",
		},
		{
			input:    "toplevel:true",
			expected: "toplevel filter",
		},
		{
			input:    "default:false",
			expected: "default filter",
		},
		{
			input:    "slug:example-1",
			expected: "slug filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			if predicate == nil {
				t.Errorf("Expected non-nil predicate")
			}

			// Test that the predicate generates valid SQL
			sql, args := query.Compile(predicate)
			if sql == "" {
				t.Errorf("Expected non-empty SQL")
			}
			if args == nil {
				t.Errorf("Expected non-nil args")
			}
		})
	}
}

func TestConverter_TypeFilters(t *testing.T) {
	tests := []struct {
		input    string
		expected model.SectionType
	}{
		{
			input:    "type:example",
			expected: model.SectionExample,
		},
		{
			input:    "type:tutorial",
			expected: model.SectionTutorial,
		},
		{
			input:    "type:application",
			expected: model.SectionApplication,
		},
		{
			input:    "type:topic",
			expected: model.SectionGeneralTopic,
		},
		{
			input:    "type:app",
			expected: model.SectionApplication,
		},
		{
			input:    "type:tut",
			expected: model.SectionTutorial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			// Test that the predicate generates correct SQL
			sql, args := query.Compile(predicate)
			if !containsString(sql, "sectionType") {
				t.Errorf("Expected SQL to contain sectionType filter")
			}
			if len(args) == 0 {
				t.Errorf("Expected arguments for type filter")
			}
			if args[0] != tt.expected.String() {
				t.Errorf("Expected type %s, got %s", tt.expected.String(), args[0])
			}
		})
	}
}

func TestConverter_BooleanFilters(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "toplevel:true",
			expected: true,
		},
		{
			input:    "toplevel:false",
			expected: false,
		},
		{
			input:    "toplevel:yes",
			expected: true,
		},
		{
			input:    "toplevel:no",
			expected: false,
		},
		{
			input:    "toplevel:1",
			expected: true,
		},
		{
			input:    "toplevel:0",
			expected: false,
		},
		{
			input:    "default:on",
			expected: true,
		},
		{
			input:    "default:off",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			// Test that the predicate generates correct SQL
			sql, args := query.Compile(predicate)
			if tt.expected {
				if !containsString(sql, "= 1") {
					t.Errorf("Expected SQL to contain '= 1' for true value")
				}
			} else {
				if !containsString(sql, "NOT") {
					t.Errorf("Expected SQL to contain 'NOT' for false value")
				}
			}
		})
	}
}

func TestConverter_TextSearch(t *testing.T) {
	tests := []struct {
		input string
	}{
		{
			input: "\"docker deployment\"",
		},
		{
			input: "kubernetes",
		},
		{
			input: "getting started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			// Test that the predicate generates FTS SQL
			sql, args := query.Compile(predicate)
			if !containsString(sql, "section_fts") {
				t.Errorf("Expected SQL to contain FTS table join")
			}
			if !containsString(sql, "MATCH") {
				t.Errorf("Expected SQL to contain MATCH clause")
			}
			if len(args) == 0 {
				t.Errorf("Expected arguments for text search")
			}
		})
	}
}

func TestConverter_ComplexQueries(t *testing.T) {
	tests := []struct {
		input string
	}{
		{
			input: "type:example AND topic:api",
		},
		{
			input: "type:example OR type:tutorial",
		},
		{
			input: "NOT type:tutorial",
		},
		{
			input: "-type:tutorial",
		},
		{
			input: "(type:example OR type:tutorial) AND topic:getting-started",
		},
		{
			input: "docker AND (type:example OR type:tutorial)",
		},
		{
			input: "type:example AND NOT flag:debug",
		},
		{
			input: "\"docker deployment\" type:tutorial -flag:verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			// Test that the predicate generates valid SQL
			sql, args := query.Compile(predicate)
			if sql == "" {
				t.Errorf("Expected non-empty SQL")
			}
			if args == nil {
				t.Errorf("Expected non-nil args")
			}

			// Basic SQL validation
			if !containsString(sql, "SELECT") {
				t.Errorf("Expected SQL to contain SELECT")
			}
			if !containsString(sql, "FROM sections") {
				t.Errorf("Expected SQL to contain FROM sections")
			}
		})
	}
}

func TestConverter_NegatedFilters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "-type:tutorial",
			expected: "NOT",
		},
		{
			input:    "NOT flag:debug",
			expected: "NOT",
		},
		{
			input:    "-\"docker deployment\"",
			expected: "NOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			converter := NewConverter()
			predicate, err := converter.Convert(query)
			if err != nil {
				t.Fatalf("Convert error: %v", err)
			}

			// Test that the predicate generates SQL with NOT
			sql, args := query.Compile(predicate)
			if !containsString(sql, tt.expected) {
				t.Errorf("Expected SQL to contain %s", tt.expected)
			}
		})
	}
}

func TestConverter_ErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			input:       "invalid_type:unknown",
			shouldError: true,
		},
		{
			input:       "type:invalid_value",
			shouldError: true,
		},
		{
			input:       "toplevel:invalid_bool",
			shouldError: true,
		},
		{
			input:       "type:example",
			shouldError: false,
		},
		{
			input:       "toplevel:true",
			shouldError: false,
		},
		{
			input:       "",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("Unexpected parse error: %v", err)
				}
				return
			}

			converter := NewConverter()
			_, err = converter.Convert(query)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected conversion error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected conversion error: %v", err)
				}
			}
		})
	}
}

func TestConvertQuery(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			input:       "type:example",
			shouldError: false,
		},
		{
			input:       "type:example AND topic:api",
			shouldError: false,
		},
		{
			input:       "invalid_syntax :",
			shouldError: true,
		},
		{
			input:       "invalid_field:value",
			shouldError: true,
		},
		{
			input:       "type:invalid_type",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			predicate, err := ConvertQuery(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if predicate == nil {
					t.Errorf("Expected non-nil predicate")
				}
			}
		})
	}
}

func TestQueryBuilder(t *testing.T) {
	// Test fluent interface
	qb := NewQueryBuilder()
	predicate := qb.
		WithType(model.SectionExample).
		WithTopic("api").
		WithFlag("verbose").
		WithCommand("deploy").
		WithTopLevel(true).
		WithDefault(false).
		WithSlug("example-1").
		WithTextSearch("docker deployment").
		Build()

	if predicate == nil {
		t.Errorf("Expected non-nil predicate")
	}

	// Test that the predicate generates valid SQL
	sql, args := query.Compile(predicate)
	if sql == "" {
		t.Errorf("Expected non-empty SQL")
	}
	if len(args) == 0 {
		t.Errorf("Expected arguments")
	}

	// Test OR operation
	qb2 := NewQueryBuilder()
	predicate2 := qb2.
		WithType(model.SectionExample).
		WithType(model.SectionTutorial).
		UseOr().
		Build()

	if predicate2 == nil {
		t.Errorf("Expected non-nil predicate")
	}

	sql2, args2 := query.Compile(predicate2)
	if !containsString(sql2, "OR") {
		t.Errorf("Expected SQL to contain OR")
	}
	if len(args2) == 0 {
		t.Errorf("Expected arguments")
	}
}

func TestQueryOptimizer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "(type:example)",
			expected: "type:example",
		},
		{
			input:    "type:example AND topic:api",
			expected: "type:example AND topic:api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			optimizer := NewQueryOptimizer()
			optimized := optimizer.Optimize(query)

			if optimized.Root == nil {
				t.Errorf("Expected non-nil optimized root")
			}

			// Basic check - optimized query should still be valid
			converter := NewConverter()
			_, err = converter.Convert(optimized)
			if err != nil {
				t.Errorf("Optimized query conversion failed: %v", err)
			}
		})
	}
}

func TestGetSupportedFields(t *testing.T) {
	fields := GetSupportedFields()
	if len(fields) == 0 {
		t.Errorf("Expected non-empty fields list")
	}

	// Check that all expected fields are present
	expected := []string{"type", "topic", "flag", "command", "toplevel", "default", "slug", "title", "content"}
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field] = true
	}

	for _, expectedField := range expected {
		if !fieldMap[expectedField] {
			t.Errorf("Expected field %s not found", expectedField)
		}
	}
}

func TestGetSupportedTypes(t *testing.T) {
	types := GetSupportedTypes()
	if len(types) == 0 {
		t.Errorf("Expected non-empty types list")
	}

	// Check that all expected types are present
	expected := []string{"topic", "generaltopic", "example", "application", "app", "tutorial", "tut"}
	typeMap := make(map[string]bool)
	for _, typ := range types {
		typeMap[typ] = true
	}

	for _, expectedType := range expected {
		if !typeMap[expectedType] {
			t.Errorf("Expected type %s not found", expectedType)
		}
	}
}

func TestGetFieldDescription(t *testing.T) {
	tests := []struct {
		field    string
		expected string
	}{
		{
			field:    "type",
			expected: "Section type (topic, example, application, tutorial)",
		},
		{
			field:    "topic",
			expected: "Topic tags associated with the section",
		},
		{
			field:    "unknown",
			expected: "Unknown field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			desc := GetFieldDescription(tt.field)
			if desc != tt.expected {
				t.Errorf("Expected description %s, got %s", tt.expected, desc)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
				s[len(s)-len(substr):] == substr ||
				strings.Contains(s, substr))))
}

// Mock for testing without actual database
type mockCompiler struct {
	sql  string
	args []interface{}
}

func (m *mockCompiler) addWhere(cond string, args ...interface{}) {
	m.sql += " WHERE " + cond
	m.args = append(m.args, args...)
}

func (m *mockCompiler) addJoin(join string) {
	m.sql += " " + join
}

func (m *mockCompiler) getUniqueAlias(base string) string {
	return base
}

func TestConverterWithMockCompiler(t *testing.T) {
	// Test that converter generates appropriate SQL structure
	query, err := ParseQuery("type:example AND topic:api")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	predicate, err := converter.Convert(query)
	if err != nil {
		t.Fatalf("Convert error: %v", err)
	}

	// Test with mock compiler
	mock := &mockCompiler{}
	// We can't directly test the predicate function without the actual query compiler
	// But we can test that the conversion succeeds
	if predicate == nil {
		t.Errorf("Expected non-nil predicate")
	}
}
