package search

import (
	"testing"
)

func TestLexer_SimpleTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			input:    "type:example",
			expected: []TokenType{TokenFilter, TokenEOF},
		},
		{
			input:    "topic:api",
			expected: []TokenType{TokenFilter, TokenEOF},
		},
		{
			input:    "-flag:verbose",
			expected: []TokenType{TokenMinus, TokenFilter, TokenEOF},
		},
		{
			input:    "word1 word2",
			expected: []TokenType{TokenWord, TokenWord, TokenEOF},
		},
		{
			input:    "\"quoted string\"",
			expected: []TokenType{TokenString, TokenEOF},
		},
		{
			input:    "AND OR NOT",
			expected: []TokenType{TokenAnd, TokenOr, TokenNot, TokenEOF},
		},
		{
			input:    "( word )",
			expected: []TokenType{TokenLParen, TokenWord, TokenRParen, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, token := range tokens {
				if token.Type != tt.expected[i] {
					t.Errorf("Token %d: expected %v, got %v", i, tt.expected[i], token.Type)
				}
			}
		})
	}
}

func TestLexer_FilterParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "type:example",
			expected: "type:example",
		},
		{
			input:    "topic:\"api documentation\"",
			expected: "topic:api documentation",
		},
		{
			input:    "flag=verbose",
			expected: "flag=verbose",
		},
		{
			input:    "command~deploy",
			expected: "command~deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 || tokens[0].Type != TokenFilter {
				t.Fatalf("Expected filter token, got %v", tokens[0].Type)
			}

			field, operator, value, err := ParseFilter(tokens[0].Value)
			if err != nil {
				t.Fatalf("Error parsing filter: %v", err)
			}

			reconstructed := field + operator + value
			if reconstructed != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, reconstructed)
			}
		})
	}
}

func TestLexer_QuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "\"simple string\"",
			expected: "simple string",
		},
		{
			input:    "\"string with spaces\"",
			expected: "string with spaces",
		},
		{
			input:    "\"string with \\\"quotes\\\"\"",
			expected: "string with \"quotes\"",
		},
		{
			input:    "'single quoted'",
			expected: "single quoted",
		},
		{
			input:    "\"escaped\\nnewline\"",
			expected: "escaped\nnewline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) < 1 || tokens[0].Type != TokenString {
				t.Fatalf("Expected string token, got %v", tokens[0].Type)
			}

			if tokens[0].Value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tokens[0].Value)
			}
		})
	}
}

func TestLexer_ErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			input:       "\"unterminated string",
			shouldError: true,
		},
		{
			input:       "type:",
			shouldError: true,
		},
		{
			input:       ":value",
			shouldError: true,
		},
		{
			input:       "valid query",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			_, err := lexer.Tokenize()

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestLexer_ComplexQueries(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			input: "type:example AND topic:api",
			expected: []TokenType{
				TokenFilter, TokenAnd, TokenFilter, TokenEOF,
			},
		},
		{
			input: "(type:example OR type:tutorial) AND topic:getting-started",
			expected: []TokenType{
				TokenLParen, TokenFilter, TokenOr, TokenFilter, TokenRParen,
				TokenAnd, TokenFilter, TokenEOF,
			},
		},
		{
			input: "-type:tutorial \"docker deployment\"",
			expected: []TokenType{
				TokenMinus, TokenFilter, TokenString, TokenEOF,
			},
		},
		{
			input: "NOT flag:debug AND (topic:api OR topic:cli)",
			expected: []TokenType{
				TokenNot, TokenFilter, TokenAnd, TokenLParen,
				TokenFilter, TokenOr, TokenFilter, TokenRParen, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, token := range tokens {
				if token.Type != tt.expected[i] {
					t.Errorf("Token %d: expected %v, got %v", i, tt.expected[i], token.Type)
				}
			}
		})
	}
}

func TestParseFilter(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator string
		value    string
		hasError bool
	}{
		{
			input:    "type:example",
			field:    "type",
			operator: ":",
			value:    "example",
			hasError: false,
		},
		{
			input:    "topic=api",
			field:    "topic",
			operator: "=",
			value:    "api",
			hasError: false,
		},
		{
			input:    "flag~verbose",
			field:    "flag",
			operator: "~",
			value:    "verbose",
			hasError: false,
		},
		{
			input:    "invalid_filter",
			hasError: true,
		},
		{
			input:    "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			field, operator, value, err := ParseFilter(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if field != tt.field {
				t.Errorf("Expected field %s, got %s", tt.field, field)
			}
			if operator != tt.operator {
				t.Errorf("Expected operator %s, got %s", tt.operator, operator)
			}
			if value != tt.value {
				t.Errorf("Expected value %s, got %s", tt.value, value)
			}
		})
	}
}

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		field string
		valid bool
	}{
		{"type", true},
		{"topic", true},
		{"flag", true},
		{"command", true},
		{"toplevel", true},
		{"default", true},
		{"slug", true},
		{"title", true},
		{"content", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := IsValidFieldName(tt.field)
			if result != tt.valid {
				t.Errorf("Expected %v for field %s, got %v", tt.valid, tt.field, result)
			}
		})
	}
}

func TestNormalizeFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TYPE", "type"},
		{"Topic", "topic"},
		{"COMMAND", "command"},
		{"flag", "flag"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"and", true},
		{"AND", true},
		{"or", true},
		{"OR", true},
		{"not", true},
		{"NOT", true},
		{"word", false},
		{"type", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := IsKeyword(tt.word)
			if result != tt.expected {
				t.Errorf("Expected %v for word %s, got %v", tt.expected, tt.word, result)
			}
		})
	}
}
