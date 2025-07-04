package help

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

// ExampleStoreUsage demonstrates how to use the new store-based query functionality
func ExampleStoreUsage() {
	// Create an in-memory store
	st, err := store.NewInMemory()
	if err != nil {
		panic(err)
	}
	defer st.Close()

	ctx := context.Background()

	// Add some example sections
	sections := []*model.Section{
		{
			Slug:           "getting-started",
			SectionType:    model.SectionTutorial,
			Title:          "Getting Started",
			Content:        "This tutorial will help you get started with the application.",
			Topics:         []string{"basics", "introduction"},
			ShowPerDefault: true,
		},
		{
			Slug:           "advanced-usage",
			SectionType:    model.SectionExample,
			Title:          "Advanced Usage Examples",
			Content:        "These examples show advanced usage patterns.",
			Topics:         []string{"advanced", "patterns"},
			Commands:       []string{"create", "update"},
			ShowPerDefault: false,
		},
		{
			Slug:           "troubleshooting",
			SectionType:    model.SectionGeneralTopic,
			Title:          "Troubleshooting",
			Content:        "Common issues and how to solve them.",
			Topics:         []string{"troubleshooting", "debugging"},
			ShowPerDefault: true,
		},
	}

	// Insert sections into the store
	for _, section := range sections {
		if err := st.Upsert(ctx, section); err != nil {
			panic(err)
		}
	}

	// Example 1: Query all tutorials
	fmt.Println("=== All Tutorials ===")
	query := NewSectionQuery().ReturnTutorials()
	results, err := query.FindSectionsWithStore(ctx, st)
	if err != nil {
		panic(err)
	}
	for _, section := range results {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 2: Query sections by topic
	fmt.Println("\n=== Sections about 'advanced' ===")
	query = NewSectionQuery().ReturnAllTypes().ReturnAnyOfTopics("advanced")
	results, err = query.FindSectionsWithStore(ctx, st)
	if err != nil {
		panic(err)
	}
	for _, section := range results {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 3: Query sections shown by default
	fmt.Println("\n=== Sections shown by default ===")
	query = NewSectionQuery().ReturnAllTypes().ReturnOnlyShownByDefault()
	results, err = query.FindSectionsWithStore(ctx, st)
	if err != nil {
		panic(err)
	}
	for _, section := range results {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 4: Query sections related to specific commands
	fmt.Println("\n=== Sections related to 'create' command ===")
	query = NewSectionQuery().ReturnAllTypes().ReturnAnyOfCommands("create")
	results, err = query.FindSectionsWithStore(ctx, st)
	if err != nil {
		panic(err)
	}
	for _, section := range results {
		fmt.Printf("- %s: %s\n", section.Slug, section.Title)
	}

	// Example 5: Using the store-compatible HelpSystem
	fmt.Println("\n=== Using HelpSystem with store ===")
	helpSystem := NewHelpSystemWithStore(st)
	
	// This would work with any existing code that uses the HelpSystem
	// but now benefits from the store backend performance
	if helpSystem.Store != nil {
		fmt.Println("HelpSystem is backed by a store!")
		count, err := helpSystem.Store.Count(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Total sections in store: %d\n", count)
	}
}
