package help

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
)

func TestQuerySectionsWithBooleanLogic(t *testing.T) {
	// Create a test help system with some sections
	hs := NewHelpSystem()

	// Add test sections
	sections := []*Section{
		{
			Section: &model.Section{
				Slug:        "example-1",
				SectionType: SectionExample,
				Title:       "Example 1",
				Content:     "This is an example",
				Topics:      []string{"templates", "basic"},
			},
		},
		{
			Section: &model.Section{
				Slug:        "example-2",
				SectionType: SectionExample,
				Title:       "Example 2",
				Content:     "Another example",
				Topics:      []string{"advanced"},
			},
		},
		{
			Section: &model.Section{
				Slug:        "tutorial-1",
				SectionType: SectionTutorial,
				Title:       "Tutorial 1",
				Content:     "This is a tutorial",
				Topics:      []string{"templates", "basic"},
			},
		},
		{
			Section: &model.Section{
				Slug:        "topic-1",
				SectionType: SectionGeneralTopic,
				Title:       "Topic 1",
				Content:     "This is a topic",
				Topics:      []string{"advanced"},
			},
		},
	}

	// Add sections to the help system
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
			expected: 1, // Only example-1
		},
		{
			name:     "Simple OR query",
			query:    "type:example OR type:tutorial",
			expected: 3, // example-1, example-2, tutorial-1
		},
		{
			name:     "NOT query",
			query:    "NOT type:topic",
			expected: 3, // All except topic-1
		},
		{
			name:     "Complex query with parentheses",
			query:    "(type:example OR type:tutorial) AND topic:templates",
			expected: 2, // example-1 and tutorial-1
		},
		{
			name:     "Text search with boolean",
			query:    "\"example\" AND topic:advanced",
			expected: 1, // Only example-2
		},
		{
			name:     "Text search fallback",
			query:    "type:example",
			expected: 2, // example-1 and example-2
		},
		{
			name:     "Implicit text search",
			query:    "another example",
			expected: 1, // Should find example-2 which contains "Another example"
		},
		{
			name:     "Single quote text search",
			query:    "'This is an example'",
			expected: 1, // Should find example-1
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
			wantErr: false,
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
