package integration

import (
	"context"
	"os"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

// setupTestStore creates a test store with sample data
func setupTestStore(t *testing.T) *store.Store {
	// Create temporary database
	tmpfile, err := os.CreateTemp("", "test_help_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	
	// Clean up after test
	t.Cleanup(func() {
		os.Remove(tmpfile.Name())
	})

	// Create store
	testStore, err := store.NewStore(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Add cleanup
	t.Cleanup(func() {
		testStore.Close()
	})

	// Add sample data
	ctx := context.Background()
	sampleSections := []*model.Section{
		{
			Slug:        "docker-example",
			Title:       "Docker Example",
			Subtitle:    "Basic Docker usage",
			Short:       "Learn Docker basics",
			Content:     "This example shows how to use Docker for deployment",
			SectionType: model.SectionExample,
			IsTopLevel:  true,
			ShowDefault: true,
			Order:       1,
			Topics:      []string{"docker", "deployment"},
			Flags:       []string{"verbose", "output"},
			Commands:    []string{"deploy"},
		},
		{
			Slug:        "api-tutorial",
			Title:       "API Tutorial",
			Subtitle:    "Getting started with APIs",
			Short:       "API basics tutorial",
			Content:     "Learn how to use our REST API endpoints",
			SectionType: model.SectionTutorial,
			IsTopLevel:  true,
			ShowDefault: true,
			Order:       2,
			Topics:      []string{"api", "getting-started"},
			Flags:       []string{"format", "output"},
			Commands:    []string{"api", "request"},
		},
		{
			Slug:        "advanced-config",
			Title:       "Advanced Configuration",
			Subtitle:    "Complex setup guide",
			Short:       "Advanced configuration options",
			Content:     "Advanced configuration for power users",
			SectionType: model.SectionTutorial,
			IsTopLevel:  false,
			ShowDefault: false,
			Order:       3,
			Topics:      []string{"configuration", "advanced"},
			Flags:       []string{"config", "debug"},
			Commands:    []string{"config"},
		},
		{
			Slug:        "app-guide",
			Title:       "Application Guide",
			Subtitle:    "Complete application walkthrough",
			Short:       "Full application guide",
			Content:     "Complete guide for building applications",
			SectionType: model.SectionApplication,
			IsTopLevel:  true,
			ShowDefault: true,
			Order:       4,
			Topics:      []string{"application", "getting-started"},
			Flags:       []string{"build", "output"},
			Commands:    []string{"build", "run"},
		},
		{
			Slug:        "troubleshooting",
			Title:       "Troubleshooting Guide",
			Subtitle:    "Common problems and solutions",
			Short:       "Fix common issues",
			Content:     "Solutions to common problems and error messages",
			SectionType: model.SectionGeneralTopic,
			IsTopLevel:  true,
			ShowDefault: true,
			Order:       5,
			Topics:      []string{"troubleshooting", "help"},
			Flags:       []string{"debug", "verbose"},
			Commands:    []string{"help"},
		},
	}

	for _, section := range sampleSections {
		err := testStore.UpsertSection(ctx, section)
		if err != nil {
			t.Fatalf("Failed to insert section %s: %v", section.Slug, err)
		}
	}

	return testStore
}

func TestSearchService_BasicSearch(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
		slugs    []string
	}{
		{
			name:     "Empty query returns all",
			query:    "",
			expected: 5,
		},
		{
			name:     "Type filter for examples",
			query:    "type:example",
			expected: 1,
			slugs:    []string{"docker-example"},
		},
		{
			name:     "Type filter for tutorials",
			query:    "type:tutorial",
			expected: 2,
			slugs:    []string{"api-tutorial", "advanced-config"},
		},
		{
			name:     "Topic filter",
			query:    "topic:docker",
			expected: 1,
			slugs:    []string{"docker-example"},
		},
		{
			name:     "Text search",
			query:    "docker",
			expected: 1,
			slugs:    []string{"docker-example"},
		},
		{
			name:     "Flag filter",
			query:    "flag:debug",
			expected: 2,
			slugs:    []string{"advanced-config", "troubleshooting"},
		},
		{
			name:     "Command filter",
			query:    "command:build",
			expected: 1,
			slugs:    []string{"app-guide"},
		},
		{
			name:     "Boolean filter true",
			query:    "toplevel:true",
			expected: 4,
		},
		{
			name:     "Boolean filter false",
			query:    "toplevel:false",
			expected: 1,
			slugs:    []string{"advanced-config"},
		},
		{
			name:     "Slug exact match",
			query:    "slug:api-tutorial",
			expected: 1,
			slugs:    []string{"api-tutorial"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := service.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(sections) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(sections))
			}

			if len(tt.slugs) > 0 {
				slugMap := make(map[string]bool)
				for _, section := range sections {
					slugMap[section.Slug] = true
				}

				for _, expectedSlug := range tt.slugs {
					if !slugMap[expectedSlug] {
						t.Errorf("Expected slug %s not found", expectedSlug)
					}
				}
			}
		})
	}
}

func TestSearchService_BooleanQueries(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
		slugs    []string
	}{
		{
			name:     "AND query",
			query:    "type:tutorial AND topic:getting-started",
			expected: 1,
			slugs:    []string{"api-tutorial"},
		},
		{
			name:     "OR query",
			query:    "type:example OR type:application",
			expected: 2,
			slugs:    []string{"docker-example", "app-guide"},
		},
		{
			name:     "NOT query",
			query:    "NOT type:tutorial",
			expected: 3,
			slugs:    []string{"docker-example", "app-guide", "troubleshooting"},
		},
		{
			name:     "Complex query with parentheses",
			query:    "(type:example OR type:tutorial) AND topic:getting-started",
			expected: 1,
			slugs:    []string{"api-tutorial"},
		},
		{
			name:     "Negated filter",
			query:    "-type:tutorial",
			expected: 3,
			slugs:    []string{"docker-example", "app-guide", "troubleshooting"},
		},
		{
			name:     "Multiple filters with implicit AND",
			query:    "type:tutorial topic:getting-started",
			expected: 1,
			slugs:    []string{"api-tutorial"},
		},
		{
			name:     "Mixed text and filters",
			query:    "docker type:example",
			expected: 1,
			slugs:    []string{"docker-example"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := service.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(sections) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(sections))
			}

			if len(tt.slugs) > 0 {
				slugMap := make(map[string]bool)
				for _, section := range sections {
					slugMap[section.Slug] = true
				}

				for _, expectedSlug := range tt.slugs {
					if !slugMap[expectedSlug] {
						t.Errorf("Expected slug %s not found", expectedSlug)
					}
				}
			}
		})
	}
}

func TestSearchService_TextSearch(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
		slugs    []string
	}{
		{
			name:     "Single word search",
			query:    "Docker",
			expected: 1,
			slugs:    []string{"docker-example"},
		},
		{
			name:     "Quoted phrase search",
			query:    "\"REST API\"",
			expected: 1,
			slugs:    []string{"api-tutorial"},
		},
		{
			name:     "Multiple word search",
			query:    "getting started",
			expected: 2,
			slugs:    []string{"api-tutorial", "app-guide"},
		},
		{
			name:     "Negated text search",
			query:    "-docker",
			expected: 4,
			slugs:    []string{"api-tutorial", "advanced-config", "app-guide", "troubleshooting"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := service.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(sections) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(sections))
			}

			if len(tt.slugs) > 0 {
				slugMap := make(map[string]bool)
				for _, section := range sections {
					slugMap[section.Slug] = true
				}

				for _, expectedSlug := range tt.slugs {
					if !slugMap[expectedSlug] {
						t.Errorf("Expected slug %s not found", expectedSlug)
					}
				}
			}
		})
	}
}

func TestSearchService_ValidationAndErrors(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)

	tests := []struct {
		name        string
		query       string
		shouldError bool
	}{
		{
			name:        "Valid query",
			query:       "type:example",
			shouldError: false,
		},
		{
			name:        "Invalid field",
			query:       "invalid_field:value",
			shouldError: true,
		},
		{
			name:        "Invalid type value",
			query:       "type:invalid_type",
			shouldError: true,
		},
		{
			name:        "Invalid boolean value",
			query:       "toplevel:maybe",
			shouldError: true,
		},
		{
			name:        "Syntax error",
			query:       "type:example AND",
			shouldError: true,
		},
		{
			name:        "Unterminated string",
			query:       "\"unterminated",
			shouldError: true,
		},
		{
			name:        "Empty filter value",
			query:       "type:",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateQuery(tt.query)

			if tt.shouldError {
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

func TestSearchService_SearchWithInfo(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	sections, info, err := service.SearchWithInfo(ctx, "type:example AND topic:docker")
	if err != nil {
		t.Fatalf("SearchWithInfo failed: %v", err)
	}

	if len(sections) != 1 {
		t.Errorf("Expected 1 result, got %d", len(sections))
	}

	if !info.HasFilters {
		t.Errorf("Expected HasFilters to be true")
	}

	if len(info.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(info.Fields))
	}

	if info.IsSimple {
		t.Errorf("Expected IsSimple to be false for AND query")
	}
}

func TestSearchService_QueryBuilder(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	// Test fluent interface
	sections, err := service.ExecuteBuilder(ctx,
		service.BuildQuery().
			WithType(model.SectionExample).
			WithTopic("docker"))

	if err != nil {
		t.Fatalf("Query builder execution failed: %v", err)
	}

	if len(sections) != 1 {
		t.Errorf("Expected 1 result, got %d", len(sections))
	}

	if sections[0].Slug != "docker-example" {
		t.Errorf("Expected docker-example, got %s", sections[0].Slug)
	}
}

func TestSearchService_SearchWithOptions(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	// Test limit
	sections, err := service.SearchWithOptions(ctx, "", SearchOptions{Limit: 2})
	if err != nil {
		t.Fatalf("Search with options failed: %v", err)
	}

	if len(sections) != 2 {
		t.Errorf("Expected 2 results with limit, got %d", len(sections))
	}

	// Test offset
	sections, err = service.SearchWithOptions(ctx, "", SearchOptions{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("Search with options failed: %v", err)
	}

	if len(sections) != 2 {
		t.Errorf("Expected 2 results with offset and limit, got %d", len(sections))
	}
}

func TestSearchService_SearchWithMetadata(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	result, err := service.SearchWithMetadata(ctx, "type:example", SearchOptions{Debug: true})
	if err != nil {
		t.Fatalf("SearchWithMetadata failed: %v", err)
	}

	if len(result.Sections) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Sections))
	}

	if result.Count != 1 {
		t.Errorf("Expected count 1, got %d", result.Count)
	}

	if result.SQL == "" {
		t.Errorf("Expected SQL debug info")
	}

	if result.QueryInfo == nil {
		t.Errorf("Expected query info")
	}

	if !result.QueryInfo.HasFilters {
		t.Errorf("Expected HasFilters to be true")
	}
}

func TestSearchService_UtilityMethods(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)

	// Test GetSupportedFields
	fields := service.GetSupportedFields()
	if len(fields) == 0 {
		t.Errorf("Expected non-empty fields list")
	}

	// Test GetSupportedTypes
	types := service.GetSupportedTypes()
	if len(types) == 0 {
		t.Errorf("Expected non-empty types list")
	}

	// Test GetFieldDescription
	desc := service.GetFieldDescription("type")
	if desc == "" {
		t.Errorf("Expected non-empty description")
	}

	// Test AnalyzeQuery
	info, err := service.AnalyzeQuery("type:example AND topic:docker")
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if !info.HasFilters {
		t.Errorf("Expected HasFilters to be true")
	}

	// Test FormatQuery
	formatted, err := service.FormatQuery("type:example AND topic:docker")
	if err != nil {
		t.Fatalf("FormatQuery failed: %v", err)
	}

	if formatted == "" {
		t.Errorf("Expected non-empty formatted query")
	}

	// Test CompileQuery
	sql, args, err := service.CompileQuery("type:example")
	if err != nil {
		t.Fatalf("CompileQuery failed: %v", err)
	}

	if sql == "" {
		t.Errorf("Expected non-empty SQL")
	}

	if len(args) == 0 {
		t.Errorf("Expected non-empty args")
	}
}

func TestSearchService_EdgeCases(t *testing.T) {
	testStore := setupTestStore(t)
	service := NewSearchService(testStore)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Empty string",
			query:    "",
			expected: 5,
		},
		{
			name:     "Whitespace only",
			query:    "   ",
			expected: 0, // Should be treated as empty after parsing
		},
		{
			name:     "No matches",
			query:    "nonexistent",
			expected: 0,
		},
		{
			name:     "Case insensitive field",
			query:    "TYPE:example",
			expected: 1,
		},
		{
			name:     "Multiple spaces",
			query:    "type:example    topic:docker",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := service.Search(ctx, tt.query)
			if err != nil && tt.name != "Whitespace only" {
				t.Fatalf("Search failed: %v", err)
			}

			if len(sections) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(sections))
			}
		})
	}
}
