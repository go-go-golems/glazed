package help

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
)

func TestQuerySectionsWithBooleanLogic(t *testing.T) {
	hs := NewHelpSystem()

	sections := []*model.Section{
		{
			Slug:        "example-1",
			SectionType: model.SectionExample,
			Title:       "Example 1",
			Content:     "This is an example",
			Topics:      []string{"templates", "basic"},
		},
		{
			Slug:        "example-2",
			SectionType: model.SectionExample,
			Title:       "Example 2",
			Content:     "Another example",
			Topics:      []string{"advanced"},
		},
		{
			Slug:        "tutorial-1",
			SectionType: model.SectionTutorial,
			Title:       "Tutorial 1",
			Content:     "This is a tutorial",
			Topics:      []string{"templates", "basic"},
		},
		{
			Slug:        "topic-1",
			SectionType: model.SectionGeneralTopic,
			Title:       "Topic 1",
			Content:     "This is a topic",
			Topics:      []string{"advanced"},
		},
	}

	for _, section := range sections {
		hs.AddSection(section)
	}

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Simple AND query",
			query:    "type:example AND topic:templates",
			expected: 1,
		},
		{
			name:     "Simple OR query",
			query:    "type:example OR type:tutorial",
			expected: 3,
		},
		{
			name:     "NOT query",
			query:    "NOT type:topic",
			expected: 3,
		},
		{
			name:     "Complex query with parentheses",
			query:    "(type:example OR type:tutorial) AND topic:templates",
			expected: 2,
		},
		{
			name:     "Text search with boolean",
			query:    "\"example\" AND topic:advanced",
			expected: 1,
		},
		{
			name:     "Text search fallback",
			query:    "type:example",
			expected: 2,
		},
		{
			name:     "Implicit text search",
			query:    "another example",
			expected: 1,
		},
		{
			name:     "Single quote text search",
			query:    "'This is an example'",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := hs.QuerySections(tt.query)
			if err != nil {
				t.Fatalf("QuerySections() error = %v", err)
			}
			if len(results) != tt.expected {
				t.Errorf("QuerySections() got %d results, expected %d for query '%s'",
					len(results), tt.expected, tt.query)
				for _, result := range results {
					t.Logf("  - %s", result.Slug)
				}
			}
		})
	}
}

func TestQuerySectionsErrorHandling(t *testing.T) {
	hs := NewHelpSystem()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Invalid AND query",
			query:   "type:example AND",
			wantErr: true,
		},
		{
			name:    "Invalid parentheses",
			query:   "(type:example AND topic:test",
			wantErr: true,
		},
		{
			name:    "Valid legacy query",
			query:   "type:unknown",
			wantErr: true,
		},
		{
			name:    "Valid boolean query",
			query:   "type:example AND topic:test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hs.QuerySections(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("QuerySections() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
