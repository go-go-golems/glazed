package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/stretchr/testify/require"
)

// ExampleUsage demonstrates how to use the SQLite-based help system
func TestExampleUsage(t *testing.T) {
	// Create an in-memory SQLite help system
	hs, err := store.NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()

	ctx := context.Background()

	// Add some example sections
	sections := []*help.Section{
		{
			Slug:        "getting-started",
			SectionType: help.SectionTutorial,
			Title:       "Getting Started",
			Content:     "Learn how to get started with our CLI tool",
			Topics:      []string{"basics", "setup"},
			IsTopLevel:  true,
			ShowPerDefault: true,
			Order:       1,
		},
		{
			Slug:        "database-example",
			SectionType: help.SectionExample,
			Title:       "Database Connection Example",
			Content:     "This example shows how to connect to a database using our CLI",
			Topics:      []string{"database", "examples"},
			Flags:       []string{"--db-url", "--timeout"},
			Commands:    []string{"connect", "query"},
			IsTopLevel:  false,
			ShowPerDefault: false,
			Order:       2,
		},
		{
			Slug:        "advanced-usage",
			SectionType: help.SectionApplication,
			Title:       "Advanced Usage Patterns",
			Content:     "Advanced patterns for power users",
			Topics:      []string{"advanced", "patterns"},
			IsTopLevel:  true,
			ShowPerDefault: false,
			Order:       3,
		},
	}

	// Add sections to the help system
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}

	// Example 1: Find all top-level sections
	fmt.Println("=== Top-level sections ===")
	topLevel, err := hs.Find(ctx, store.IsTopLevel())
	require.NoError(t, err)
	for _, section := range topLevel {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 2: Find examples related to databases
	fmt.Println("\n=== Database examples ===")
	dbExamples, err := hs.Find(ctx, store.And(
		store.IsExample(),
		store.HasTopic("database"),
	))
	require.NoError(t, err)
	for _, section := range dbExamples {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 3: Find sections for a specific flag
	fmt.Println("\n=== Sections using --db-url flag ===")
	flagSections, err := hs.Find(ctx, store.HasFlag("--db-url"))
	require.NoError(t, err)
	for _, section := range flagSections {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 4: Complex query - tutorials or examples that are shown by default
	fmt.Println("\n=== Default tutorials and examples ===")
	defaultContent, err := hs.Find(ctx, store.And(
		store.Or(
			store.IsTutorial(),
			store.IsExample(),
		),
		store.ShownByDefault(),
	))
	require.NoError(t, err)
	for _, section := range defaultContent {
		fmt.Printf("- %s (%s): %s\n", section.Slug, section.SectionType.String(), section.Title)
	}

	// Example 5: Full-text search
	fmt.Println("\n=== Full-text search for 'database' ===")
	searchResults, err := hs.Find(ctx, store.TextSearch("database"))
	require.NoError(t, err)
	for _, section := range searchResults {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 6: Get statistics
	fmt.Println("\n=== Help system statistics ===")
	stats, err := hs.GetStats(ctx)
	require.NoError(t, err)
	for key, value := range stats {
		fmt.Printf("- %s: %d\n", key, value)
	}
}
