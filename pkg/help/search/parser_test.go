package search

import (
	"testing"
)

func TestParser_SimpleQueries(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "type:example",
			expected: "type:example",
		},
		{
			input:    "topic:api",
			expected: "topic:api",
		},
		{
			input:    "\"docker deployment\"",
			expected: "\"docker deployment\"",
		},
		{
			input:    "simple word",
			expected: "simple",
		},
		{
			input:    "-type:tutorial",
			expected: "-type:tutorial",
		},
		{
			input:    "NOT flag:debug",
			expected: "NOT flag:debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if query.Root == nil {
				t.Fatalf("Expected non-nil root")
			}

			// Check that the query can be converted to string
			result := query.Root.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParser_BooleanQueries(t *testing.T) {
	tests := []struct {
		input       string
		shouldParse bool
	}{
		{
			input:       "type:example AND topic:api",
			shouldParse: true,
		},
		{
			input:       "type:example OR type:tutorial",
			shouldParse: true,
		},
		{
			input:       "NOT type:tutorial",
			shouldParse: true,
		},
		{
			input:       "(type:example OR type:tutorial) AND topic:getting-started",
			shouldParse: true,
		},
		{
			input:       "type:example AND NOT flag:debug",
			shouldParse: true,
		},
		{
			input:       "docker AND (type:example OR type:tutorial)",
			shouldParse: true,
		},
		{
			input:       "type:example AND",
			shouldParse: false,
		},
		{
			input:       "OR type:example",
			shouldParse: false,
		},
		{
			input:       "(type:example",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)

			if tt.shouldParse {
				if err != nil {
					t.Errorf("Expected successful parse, got error: %v", err)
				}
				if query.Root == nil {
					t.Errorf("Expected non-nil root")
				}
			} else {
				if err == nil {
					t.Errorf("Expected parse error, got successful parse")
				}
			}
		})
	}
}

func TestParser_ComplexQueries(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, *Query)
	}{
		{
			input: "type:example AND topic:api",
			check: func(t *testing.T, q *Query) {
				if !IsAndNode(q.Root) {
					t.Errorf("Expected AND node at root")
				}
				and := q.Root.(*BooleanNode)
				if _, ok := and.Left.(*FilterNode); !ok {
					t.Errorf("Expected FilterNode on left")
				}
				if _, ok := and.Right.(*FilterNode); !ok {
					t.Errorf("Expected FilterNode on right")
				}
			},
		},
		{
			input: "(type:example OR type:tutorial) AND topic:getting-started",
			check: func(t *testing.T, q *Query) {
				if !IsAndNode(q.Root) {
					t.Errorf("Expected AND node at root")
				}
				and := q.Root.(*BooleanNode)
				if _, ok := and.Left.(*GroupNode); !ok {
					t.Errorf("Expected GroupNode on left")
				}
				if _, ok := and.Right.(*FilterNode); !ok {
					t.Errorf("Expected FilterNode on right")
				}
			},
		},
		{
			input: "NOT type:tutorial",
			check: func(t *testing.T, q *Query) {
				if !IsNotNode(q.Root) {
					t.Errorf("Expected NOT node at root")
				}
				not := q.Root.(*BooleanNode)
				if _, ok := not.Left.(*FilterNode); !ok {
					t.Errorf("Expected FilterNode in NOT")
				}
			},
		},
		{
			input: "-type:tutorial \"docker deployment\"",
			check: func(t *testing.T, q *Query) {
				if !IsAndNode(q.Root) {
					t.Errorf("Expected AND node at root")
				}
				and := q.Root.(*BooleanNode)
				if filter, ok := and.Left.(*FilterNode); ok {
					if !filter.Negated {
						t.Errorf("Expected negated filter")
					}
				} else {
					t.Errorf("Expected FilterNode on left")
				}
				if _, ok := and.Right.(*TextNode); !ok {
					t.Errorf("Expected TextNode on right")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			tt.check(t, query)
		})
	}
}

func TestParser_ImplicitAND(t *testing.T) {
	tests := []struct {
		input    string
		expected int // Expected number of terms
	}{
		{
			input:    "word1 word2",
			expected: 2,
		},
		{
			input:    "type:example topic:api",
			expected: 2,
		},
		{
			input:    "docker kubernetes deployment",
			expected: 3,
		},
		{
			input:    "type:example \"quoted string\" word",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			terms := CollectTerms(query.Root)
			if len(terms) != tt.expected {
				t.Errorf("Expected %d terms, got %d", tt.expected, len(terms))
			}
		})
	}
}

func TestParser_NegatedTerms(t *testing.T) {
	tests := []struct {
		input    string
		negated  bool
		termType string
	}{
		{
			input:    "-type:tutorial",
			negated:  true,
			termType: "filter",
		},
		{
			input:    "-\"quoted string\"",
			negated:  true,
			termType: "text",
		},
		{
			input:    "-word",
			negated:  true,
			termType: "text",
		},
		{
			input:    "type:example",
			negated:  false,
			termType: "filter",
		},
		{
			input:    "\"quoted string\"",
			negated:  false,
			termType: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			terms := CollectTerms(query.Root)
			if len(terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(terms))
			}

			term := terms[0]
			if term.IsNegated() != tt.negated {
				t.Errorf("Expected negated=%v, got %v", tt.negated, term.IsNegated())
			}

			switch tt.termType {
			case "filter":
				if _, ok := term.(*FilterNode); !ok {
					t.Errorf("Expected FilterNode, got %T", term)
				}
			case "text":
				if _, ok := term.(*TextNode); !ok {
					t.Errorf("Expected TextNode, got %T", term)
				}
			}
		})
	}
}

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{
			input:    "type:example",
			hasError: false,
		},
		{
			input:    "invalid_field:value",
			hasError: true,
		},
		{
			input:    "type:",
			hasError: true,
		},
		{
			input:    "topic:\"\"",
			hasError: true,
		},
		{
			input:    "\"\"",
			hasError: true,
		},
		{
			input:    "type:example AND topic:api",
			hasError: false,
		},
		{
			input:    "",
			hasError: false, // Empty query is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			err = ValidateQuery(query)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzeQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected QueryInfo
	}{
		{
			input: "type:example",
			expected: QueryInfo{
				HasFilters:    true,
				HasTextSearch: false,
				Fields:        []string{"type"},
				TextTerms:     []string{},
				IsSimple:      true,
			},
		},
		{
			input: "\"docker deployment\"",
			expected: QueryInfo{
				HasFilters:    false,
				HasTextSearch: true,
				Fields:        []string{},
				TextTerms:     []string{"docker deployment"},
				IsSimple:      true,
			},
		},
		{
			input: "type:example AND topic:api",
			expected: QueryInfo{
				HasFilters:    true,
				HasTextSearch: false,
				Fields:        []string{"type", "topic"},
				TextTerms:     []string{},
				IsSimple:      false,
			},
		},
		{
			input: "docker type:example",
			expected: QueryInfo{
				HasFilters:    true,
				HasTextSearch: true,
				Fields:        []string{"type"},
				TextTerms:     []string{"docker"},
				IsSimple:      false,
			},
		},
		{
			input: "",
			expected: QueryInfo{
				HasFilters:    false,
				HasTextSearch: false,
				Fields:        []string{},
				TextTerms:     []string{},
				IsSimple:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			info := AnalyzeQuery(query)

			if info.HasFilters != tt.expected.HasFilters {
				t.Errorf("Expected HasFilters=%v, got %v", tt.expected.HasFilters, info.HasFilters)
			}
			if info.HasTextSearch != tt.expected.HasTextSearch {
				t.Errorf("Expected HasTextSearch=%v, got %v", tt.expected.HasTextSearch, info.HasTextSearch)
			}
			if info.IsSimple != tt.expected.IsSimple {
				t.Errorf("Expected IsSimple=%v, got %v", tt.expected.IsSimple, info.IsSimple)
			}

			// Check fields (order may vary)
			if len(info.Fields) != len(tt.expected.Fields) {
				t.Errorf("Expected %d fields, got %d", len(tt.expected.Fields), len(info.Fields))
			} else {
				fieldMap := make(map[string]bool)
				for _, field := range info.Fields {
					fieldMap[field] = true
				}
				for _, expectedField := range tt.expected.Fields {
					if !fieldMap[expectedField] {
						t.Errorf("Expected field %s not found", expectedField)
					}
				}
			}

			// Check text terms
			if len(info.TextTerms) != len(tt.expected.TextTerms) {
				t.Errorf("Expected %d text terms, got %d", len(tt.expected.TextTerms), len(info.TextTerms))
			} else {
				for i, term := range info.TextTerms {
					if term != tt.expected.TextTerms[i] {
						t.Errorf("Expected text term %s, got %s", tt.expected.TextTerms[i], term)
					}
				}
			}
		})
	}
}

func TestFormatQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "type:example",
			expected: "type:example",
		},
		{
			input:    "\"docker deployment\"",
			expected: "\"docker deployment\"",
		},
		{
			input:    "type:example AND topic:api",
			expected: "AND\n  type:example\n  topic:api",
		},
		{
			input:    "NOT type:tutorial",
			expected: "NOT\n  type:tutorial",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			query, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			result := FormatQuery(query)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestCollectFunctions(t *testing.T) {
	query, err := ParseQuery("type:example AND \"docker deployment\" OR -flag:debug")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Test CollectFilters
	filters := CollectFilters(query.Root)
	if len(filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filters))
	}

	// Test CollectTextSearches
	texts := CollectTextSearches(query.Root)
	if len(texts) != 1 {
		t.Errorf("Expected 1 text search, got %d", len(texts))
	}

	// Test CollectTerms
	terms := CollectTerms(query.Root)
	if len(terms) != 3 {
		t.Errorf("Expected 3 terms, got %d", len(terms))
	}

	// Test GetAllFields
	fields := GetAllFields(query.Root)
	expectedFields := []string{"type", "flag"}
	if len(fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(fields))
	}

	// Test ContainsField
	if !ContainsField(query.Root, "type") {
		t.Errorf("Expected to find 'type' field")
	}
	if ContainsField(query.Root, "topic") {
		t.Errorf("Did not expect to find 'topic' field")
	}

	// Test GetFieldValues
	typeValues := GetFieldValues(query.Root, "type")
	if len(typeValues) != 1 || typeValues[0] != "example" {
		t.Errorf("Expected ['example'], got %v", typeValues)
	}

	// Test HasTextSearch
	if !HasTextSearch(query.Root) {
		t.Errorf("Expected to find text search")
	}

	// Test GetTextSearchTerms
	textTerms := GetTextSearchTerms(query.Root)
	if len(textTerms) != 1 || textTerms[0] != "docker deployment" {
		t.Errorf("Expected ['docker deployment'], got %v", textTerms)
	}

	// Test CombineTextSearchTerms
	combined := CombineTextSearchTerms(query.Root)
	if combined != "docker deployment" {
		t.Errorf("Expected 'docker deployment', got %s", combined)
	}
}

func TestSimplifyAST(t *testing.T) {
	query, err := ParseQuery("(type:example)")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	simplified := SimplifyAST(query.Root)
	if _, ok := simplified.(*FilterNode); !ok {
		t.Errorf("Expected FilterNode after simplification, got %T", simplified)
	}
}
