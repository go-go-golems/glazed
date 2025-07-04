package dsl

import (
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/store"
)

// TestLexer tests the lexer functionality
func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			input:    "type:example",
			expected: []TokenType{TokenIdent, TokenColon, TokenIdent, TokenEOF},
		},
		{
			input:    "examples AND tutorials",
			expected: []TokenType{TokenIdent, TokenAnd, TokenIdent, TokenEOF},
		},
		{
			input:    "NOT type:application",
			expected: []TokenType{TokenNot, TokenIdent, TokenColon, TokenIdent, TokenEOF},
		},
		{
			input:    "(examples OR tutorials) AND topic:database",
			expected: []TokenType{TokenLeftParen, TokenIdent, TokenOr, TokenIdent, TokenRightParen, TokenAnd, TokenIdent, TokenColon, TokenIdent, TokenEOF},
		},
		{
			input:    "\"full text search\"",
			expected: []TokenType{TokenString, TokenEOF},
		},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens := lexer.GetAllTokens()

		if len(tokens) != len(tt.expected) {
			t.Errorf("input %q: expected %d tokens, got %d", tt.input, len(tt.expected), len(tokens))
			continue
		}

		for i, expectedType := range tt.expected {
			if tokens[i].Type != expectedType {
				t.Errorf("input %q: token %d expected type %s, got %s", tt.input, i, expectedType, tokens[i].Type)
			}
		}
	}
}

// TestParser tests the parser functionality
func TestParser(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		description string
	}{
		{
			input:       "type:example",
			shouldError: false,
			description: "simple field expression",
		},
		{
			input:       "examples",
			shouldError: false,
			description: "implicit text search",
		},
		{
			input:       "\"full text search\"",
			shouldError: false,
			description: "quoted text search",
		},
		{
			input:       "examples AND tutorials",
			shouldError: false,
			description: "implicit text search with AND",
		},
		{
			input:       "type:example OR type:tutorial",
			shouldError: false,
			description: "OR expression",
		},
		{
			input:       "NOT type:application",
			shouldError: false,
			description: "NOT expression",
		},
		{
			input:       "(examples OR tutorials) AND topic:database",
			shouldError: false,
			description: "grouped expression",
		},
		{
			input:       "type:example AND (topic:database OR topic:sqlite)",
			shouldError: false,
			description: "complex nested expression",
		},
		{
			input:       "type:",
			shouldError: true,
			description: "missing value after colon",
		},
		{
			input:       "type:example AND",
			shouldError: true,
			description: "missing right operand",
		},
		{
			input:       "(examples OR tutorials",
			shouldError: true,
			description: "missing closing parenthesis",
		},
		{
			input:       "examples OR tutorials)",
			shouldError: true,
			description: "unexpected closing parenthesis",
		},
	}

	for _, tt := range tests {
		_, err := Parse(tt.input)
		if tt.shouldError && err == nil {
			t.Errorf("input %q (%s): expected error but got none", tt.input, tt.description)
		} else if !tt.shouldError && err != nil {
			t.Errorf("input %q (%s): unexpected error: %v", tt.input, tt.description, err)
		}
	}
}

// TestCompiler tests the compiler functionality
func TestCompiler(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		description string
	}{
		{
			input:       "type:example",
			shouldError: false,
			description: "valid type field",
		},
		{
			input:       "topic:database",
			shouldError: false,
			description: "valid topic field",
		},
		{
			input:       "flag:--output",
			shouldError: false,
			description: "valid flag field with --",
		},
		{
			input:       "flag:output",
			shouldError: false,
			description: "valid flag field without --",
		},
		{
			input:       "command:json",
			shouldError: false,
			description: "valid command field",
		},
		{
			input:       "slug:my-slug",
			shouldError: false,
			description: "valid slug field",
		},
		{
			input:       "toplevel:true",
			shouldError: false,
			description: "valid boolean field",
		},
		{
			input:       "default:false",
			shouldError: false,
			description: "valid boolean field false",
		},
		{
			input:       "examples",
			shouldError: false,
			description: "valid text search",
		},
		{
			input:       "\"full text search\"",
			shouldError: false,
			description: "valid text search",
		},
		{
			input:       "type:invalid",
			shouldError: true,
			description: "invalid type value",
		},
		{
			input:       "invalidfield:value",
			shouldError: true,
			description: "invalid field name",
		},
		{
			input:       "invalidshortcut",
			shouldError: false,
			description: "text search (formerly invalid shortcut)",
		},
		{
			input:       "toplevel:maybe",
			shouldError: true,
			description: "invalid boolean value",
		},
	}

	for _, tt := range tests {
		_, err := ParseQuery(tt.input)
		if tt.shouldError && err == nil {
			t.Errorf("input %q (%s): expected error but got none", tt.input, tt.description)
		} else if !tt.shouldError && err != nil {
			t.Errorf("input %q (%s): unexpected error: %v", tt.input, tt.description, err)
		}
	}
}

// TestQueryExamples tests the example queries from the specification
func TestQueryExamples(t *testing.T) {
	examples := []string{
		"type:example",
		"topic:database",
		"flag:--output",
		"command:json",
		"toplevel:true",
		"default:true",
		"template:true",
		"\"SQLite database\"",
		"type:example AND topic:database",
		"flag:--output AND command:json",
		"type:example OR type:tutorial",
		"NOT type:application",
		"type:example AND NOT topic:advanced",
		"(type:example OR type:tutorial) AND topic:database",
		"examples",
		"tutorials",
		"topics",
		"applications",
		"toplevel",
		"defaults",
		"examples AND topic:database",
		"defaults AND (examples OR tutorials)",
		"(type:example OR type:tutorial) AND (topic:database OR topic:sql) AND NOT \"advanced\"",
		"\"full text search\" AND type:tutorial",
		"\"performance\" OR \"optimization\"",
	}

	for _, example := range examples {
		_, err := ParseQuery(example)
		if err != nil {
			t.Errorf("example query %q failed: %v", example, err)
		}
	}
}

// TestCaseInsensitivity tests case insensitive parsing
func TestCaseInsensitivity(t *testing.T) {
	tests := []struct {
		input1 string
		input2 string
	}{
		{"type:example", "Type:Example"},
		{"type:example", "TYPE:EXAMPLE"},
		{"examples AND tutorials", "Examples AND Tutorials"},
		{"examples AND tutorials", "EXAMPLES AND TUTORIALS"},
		{"examples and tutorials", "Examples AND Tutorials"},
		{"NOT type:application", "not Type:Application"},
		{"topic:database", "Topic:Database"},
	}

	for _, tt := range tests {
		predicate1, err1 := ParseQuery(tt.input1)
		predicate2, err2 := ParseQuery(tt.input2)

		if err1 != nil || err2 != nil {
			t.Errorf("case insensitive test failed: %q vs %q, errors: %v, %v", tt.input1, tt.input2, err1, err2)
			continue
		}

		if predicate1 == nil || predicate2 == nil {
			t.Errorf("case insensitive test failed: %q vs %q, got nil predicates", tt.input1, tt.input2)
			continue
		}

		// Both should parse successfully (detailed equality testing would require more complex comparison)
	}
}

// TestErrorMessages tests error message quality
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		input               string
		expectedErrorSubstr string
	}{
		{
			input:               "type:invalid",
			expectedErrorSubstr: "unknown section type",
		},
		{
			input:               "invalidfield:value",
			expectedErrorSubstr: "unknown field",
		},

		{
			input:               "type:",
			expectedErrorSubstr: "expected value after",
		},
		{
			input:               "(examples OR tutorials",
			expectedErrorSubstr: "expected ')'",
		},
		{
			input:               "toplevel:maybe",
			expectedErrorSubstr: "invalid boolean value",
		},
	}

	for _, tt := range tests {
		_, err := ParseQuery(tt.input)
		if err == nil {
			t.Errorf("input %q: expected error but got none", tt.input)
			continue
		}

		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.expectedErrorSubstr)) {
			t.Errorf("input %q: expected error containing %q, got %q", tt.input, tt.expectedErrorSubstr, err.Error())
		}
	}
}

// TestEmptyQuery tests empty query handling
func TestEmptyQuery(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\t",
		"\n",
	}

	for _, input := range tests {
		predicate, err := ParseQuery(input)
		if err != nil {
			t.Errorf("empty query %q should not error: %v", input, err)
		}
		if predicate == nil {
			t.Errorf("empty query %q should return valid predicate", input)
		}
	}
}

// TestValidationFunctions tests the validation helper functions
func TestValidationFunctions(t *testing.T) {
	// Test field validation
	validFields := []string{"type", "topic", "flag", "command", "slug", "toplevel", "default", "template"}
	for _, field := range validFields {
		if !IsValidField(field) {
			t.Errorf("field %q should be valid", field)
		}
	}

	invalidFields := []string{"invalid", "unknown", "badfield"}
	for _, field := range invalidFields {
		if IsValidField(field) {
			t.Errorf("field %q should be invalid", field)
		}
	}

	// Test type validation
	validTypes := []string{"example", "tutorial", "topic", "application"}
	for _, typeValue := range validTypes {
		if !IsValidType(typeValue) {
			t.Errorf("type %q should be valid", typeValue)
		}
	}

	invalidTypes := []string{"invalid", "unknown", "badtype"}
	for _, typeValue := range invalidTypes {
		if IsValidType(typeValue) {
			t.Errorf("type %q should be invalid", typeValue)
		}
	}

	// Shortcut validation removed - shortcuts are no longer supported
}

// TestQueryInfo tests the query information function
func TestQueryInfo(t *testing.T) {
	info := GetQueryInfo()

	if len(info.ValidFields) == 0 {
		t.Error("QueryInfo should have valid fields")
	}

	if len(info.ValidTypes) == 0 {
		t.Error("QueryInfo should have valid types")
	}

	// Shortcuts are no longer supported

	if len(info.Examples) == 0 {
		t.Error("QueryInfo should have examples")
	}

	// Test that all examples are valid
	for _, example := range info.Examples {
		_, err := ParseQuery(example)
		if err != nil {
			t.Errorf("example query %q should be valid: %v", example, err)
		}
	}
}

// TestPredicateGeneration tests that predicates are generated correctly
func TestPredicateGeneration(t *testing.T) {
	tests := []struct {
		input       string
		description string
	}{
		{
			input:       "type:example",
			description: "should generate type predicate",
		},
		{
			input:       "examples",
			description: "should generate shortcut predicate",
		},
		{
			input:       "examples AND tutorials",
			description: "should generate AND predicate",
		},
		{
			input:       "examples OR tutorials",
			description: "should generate OR predicate",
		},
		{
			input:       "NOT examples",
			description: "should generate NOT predicate",
		},
		{
			input:       "\"test search\"",
			description: "should generate text search predicate",
		},
	}

	for _, tt := range tests {
		predicate, err := ParseQuery(tt.input)
		if err != nil {
			t.Errorf("input %q (%s): unexpected error: %v", tt.input, tt.description, err)
			continue
		}

		if predicate == nil {
			t.Errorf("input %q (%s): expected predicate but got nil", tt.input, tt.description)
			continue
		}

		// Test that predicate can be applied to a query compiler
		compiler := store.NewQueryCompiler()
		predicate(compiler)

		// The predicate should have modified the compiler in some way
		query, _ := compiler.BuildQuery()
		if query == "" {
			t.Errorf("input %q (%s): predicate should generate non-empty query", tt.input, tt.description)
		}

		// Basic validation that query looks like SQL
		if !strings.Contains(query, "FROM sections") {
			t.Errorf("input %q (%s): query should contain FROM sections", tt.input, tt.description)
		}
	}
}

// TestComplexQueries tests complex query combinations
func TestComplexQueries(t *testing.T) {
	complexQueries := []string{
		"(examples OR tutorials) AND topic:database AND NOT \"advanced\"",
		"type:example AND (topic:database OR topic:sqlite) AND toplevel:true",
		"(flag:--output OR flag:--format) AND (command:json OR command:yaml)",
		"defaults AND (examples OR tutorials) AND (topic:database OR topic:json)",
		"\"full text\" AND type:tutorial AND NOT template:true",
		"(toplevel:true AND default:true) OR (type:example AND topic:database)",
	}

	for _, query := range complexQueries {
		predicate, err := ParseQuery(query)
		if err != nil {
			t.Errorf("complex query %q failed: %v", query, err)
			continue
		}

		if predicate == nil {
			t.Errorf("complex query %q returned nil predicate", query)
			continue
		}

		// Test that predicate can be applied
		compiler := store.NewQueryCompiler()
		predicate(compiler)
		sqlQuery, _ := compiler.BuildQuery()

		if !strings.Contains(sqlQuery, "FROM sections") {
			t.Errorf("complex query %q generated invalid SQL: %s", query, sqlQuery)
		}
	}
}

// BenchmarkParsing benchmarks the parsing performance
func BenchmarkParsing(b *testing.B) {
	queries := []string{
		"type:example",
		"examples AND tutorials",
		"(examples OR tutorials) AND topic:database",
		"NOT type:application",
		"\"full text search\"",
		"type:example AND topic:database AND NOT \"advanced\"",
	}

	for _, query := range queries {
		b.Run(query, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseQuery(query)
				if err != nil {
					b.Errorf("benchmark query %q failed: %v", query, err)
				}
			}
		})
	}
}
