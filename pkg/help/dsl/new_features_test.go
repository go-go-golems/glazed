package dsl

import (
	"testing"
)

// TestSingleQuoteSupport tests single quote support in the lexer and parser
func TestSingleQuoteSupport(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "'single quoted text'",
			expected: "\"single quoted text\"",
		},
		{
			input:    "'hello world'",
			expected: "\"hello world\"",
		},
		{
			input:    "'text with special chars: @#$%'",
			expected: "\"text with special chars: @#$%\"",
		},
	}

	for _, tt := range tests {
		expr, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", tt.input, err)
			continue
		}

		if expr.String() != tt.expected {
			t.Errorf("Parse(%q) = %q, expected %q", tt.input, expr.String(), tt.expected)
		}
	}
}

// TestImplicitTextSearch tests that unquoted words are treated as text search
func TestImplicitTextSearch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "hello",
			expected: "\"hello\"",
		},
		{
			input:    "hello world",
			expected: "\"hello world\"",
		},
		{
			input:    "multiple word text search",
			expected: "\"multiple word text search\"",
		},
		{
			input:    "database tutorial",
			expected: "\"database tutorial\"",
		},
	}

	for _, tt := range tests {
		expr, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", tt.input, err)
			continue
		}

		if expr.String() != tt.expected {
			t.Errorf("Parse(%q) = %q, expected %q", tt.input, expr.String(), tt.expected)
		}
	}
}

// TestImplicitTextSearchWithBooleans tests implicit text search combined with boolean operators
func TestImplicitTextSearchWithBooleans(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "hello world AND type:example",
			expected: "(\"hello world\" AND type:example)",
		},
		{
			input:    "database tutorial OR type:advanced",
			expected: "(\"database tutorial\" OR type:advanced)",
		},
		{
			input:    "performance optimization AND NOT type:application",
			expected: "(\"performance optimization\" AND (NOT type:application))",
		},
		{
			input:    "type:example AND advanced features",
			expected: "(type:example AND \"advanced features\")",
		},
	}

	for _, tt := range tests {
		expr, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", tt.input, err)
			continue
		}

		if expr.String() != tt.expected {
			t.Errorf("Parse(%q) = %q, expected %q", tt.input, expr.String(), tt.expected)
		}
	}
}

// TestFormerShortcutsAsTextSearch tests that former shortcuts are now treated as text search
func TestFormerShortcutsAsTextSearch(t *testing.T) {
	formerShortcuts := []string{
		"examples",
		"tutorials",
		"topics",
		"applications",
		"toplevel",
		"defaults",
		"templates",
	}

	for _, shortcut := range formerShortcuts {
		expr, err := Parse(shortcut)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", shortcut, err)
			continue
		}

		expected := "\"" + shortcut + "\""
		if expr.String() != expected {
			t.Errorf("Parse(%q) = %q, expected %q (text search)", shortcut, expr.String(), expected)
		}
	}
}

// TestMixedQuoteTypes tests mixing single and double quotes
func TestMixedQuoteTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "\"double quoted\" AND 'single quoted'",
			expected: "(\"double quoted\" AND \"single quoted\")",
		},
		{
			input:    "'single' OR \"double\"",
			expected: "(\"single\" OR \"double\")",
		},
		{
			input:    "unquoted text AND 'single quoted' AND \"double quoted\"",
			expected: "((\"unquoted text\" AND \"single quoted\") AND \"double quoted\")",
		},
	}

	for _, tt := range tests {
		expr, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", tt.input, err)
			continue
		}

		if expr.String() != tt.expected {
			t.Errorf("Parse(%q) = %q, expected %q", tt.input, expr.String(), tt.expected)
		}
	}
}

// TestBackwardCompatibility tests that explicit field queries still work
func TestBackwardCompatibility(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		description string
	}{
		{
			input:       "type:example",
			shouldError: false,
			description: "explicit type field",
		},
		{
			input:       "topic:database",
			shouldError: false,
			description: "topic field",
		},
		{
			input:       "flag:--output",
			shouldError: false,
			description: "flag field",
		},
		{
			input:       "command:json",
			shouldError: false,
			description: "command field",
		},
		{
			input:       "toplevel:true",
			shouldError: false,
			description: "boolean field",
		},
		{
			input:       "type:example AND topic:database",
			shouldError: false,
			description: "field queries with boolean",
		},
		{
			input:       "(type:example OR type:tutorial) AND topic:advanced",
			shouldError: false,
			description: "complex field query",
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

// TestEdgeCases tests edge cases in the new parsing logic
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		expected    string
		description string
	}{
		{
			input:       "word",
			shouldError: false,
			expected:    "\"word\"",
			description: "single word becomes text search",
		},
		{
			input:       "word AND",
			shouldError: true,
			expected:    "",
			description: "incomplete boolean expression",
		},
		{
			input:       "word:",
			shouldError: true,
			expected:    "",
			description: "incomplete field expression",
		},
		{
			input:       "'unclosed single quote",
			shouldError: false,
			expected:    "\"unclosed single quote\"",
			description: "unclosed quote treated as text",
		},
		{
			input:       "word1 word2 AND word3 word4",
			shouldError: false,
			expected:    "(\"word1 word2\" AND \"word3 word4\")",
			description: "multiple text groups with AND",
		},
	}

	for _, tt := range tests {
		expr, err := Parse(tt.input)

		if tt.shouldError {
			if err == nil {
				t.Errorf("input %q (%s): expected error but got none", tt.input, tt.description)
			}
		} else {
			if err != nil {
				t.Errorf("input %q (%s): unexpected error: %v", tt.input, tt.description, err)
				continue
			}

			if tt.expected != "" && expr.String() != tt.expected {
				t.Errorf("input %q (%s): got %q, expected %q", tt.input, tt.description, expr.String(), tt.expected)
			}
		}
	}
}

// TestLexerTokenization tests that both quote types produce the same token type
func TestLexerTokenization(t *testing.T) {
	tests := []struct {
		input1 string
		input2 string
	}{
		{`"double"`, `'single'`},
		{`"hello world"`, `'hello world'`},
		{`"special chars: @#$"`, `'special chars: @#$'`},
	}

	for _, tt := range tests {
		lexer1 := NewLexer(tt.input1)
		lexer2 := NewLexer(tt.input2)

		tok1 := lexer1.NextToken()
		tok2 := lexer2.NextToken()

		if tok1.Type != tok2.Type {
			t.Errorf("Token types differ: %q -> %s, %q -> %s", tt.input1, tok1.Type, tt.input2, tok2.Type)
		}

		if tok1.Type != TokenString {
			t.Errorf("Expected TokenString, got %s for %q", tok1.Type, tt.input1)
		}
	}
}
