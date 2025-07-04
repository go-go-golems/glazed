package store_test

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

// Example demonstrates basic usage of the SQLite help system
func Example() {
	// Create an in-memory help system for demonstration
	hs, err := store.NewInMemoryHelpSystem()
	if err != nil {
		log.Fatal(err)
	}
	defer hs.Close()

	// Add some example sections
	sections := []*model.Section{
		{
			Slug:           "getting-started",
			Title:          "Getting Started",
			SectionType:    model.SectionGeneralTopic,
			Content:        "This section covers the basics of getting started.",
			Topics:         []string{"basics", "introduction"},
			IsTopLevel:     true,
			ShowPerDefault: true,
			Order:          1,
		},
		{
			Slug:           "example-hello",
			Title:          "Hello World Example",
			SectionType:    model.SectionExample,
			Content:        "A simple hello world example.",
			Topics:         []string{"basics", "examples"},
			Commands:       []string{"hello"},
			ShowPerDefault: true,
			Order:          2,
		},
		{
			Slug:           "advanced-tutorial",
			Title:          "Advanced Tutorial",
			SectionType:    model.SectionTutorial,
			Content:        "Advanced concepts and techniques.",
			Topics:         []string{"advanced", "tutorial"},
			ShowPerDefault: false,
			Order:          3,
		},
	}

	// Add sections to the help system
	for _, section := range sections {
		if err := hs.AddSection(section); err != nil {
			log.Fatal(err)
		}
	}

	// Query examples using the predicate system

	// 1. Find all top-level sections
	topLevel, err := hs.Find(store.IsTopLevel())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Top-level sections: %d\n", len(topLevel))

	// 2. Find examples related to "basics"
	basicExamples, err := hs.Find(store.And(
		store.IsExample(),
		store.HasTopic("basics"),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Basic examples: %d\n", len(basicExamples))

	// 3. Complex query: Find sections that are either examples OR tutorials AND relate to "basics"
	complexQuery, err := hs.Find(store.And(
		store.Or(store.IsExample(), store.IsTutorial()),
		store.HasTopic("basics"),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Examples or tutorials about basics: %d\n", len(complexQuery))

	// 4. Text search (uses LIKE fallback when FTS5 is not available)
	searchResults, err := hs.Find(store.TextSearch("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Search results for 'hello': %d\n", len(searchResults))

	// 5. Get sections shown by default, ordered by order field
	defaultSections, err := hs.Find(store.And(
		store.ShownByDefault(),
		store.OrderByOrder(),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Default sections: %d\n", len(defaultSections))

	// 6. Convenience methods
	stats, err := hs.GetStats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total sections: %d\n", stats["total"])

	// Output:
	// Top-level sections: 1
	// Basic examples: 1
	// Examples or tutorials about basics: 1
	// Search results for 'hello': 1
	// Default sections: 2
	// Total sections: 3
}

// Example_advancedQuery demonstrates more complex querying capabilities
func Example_advancedQuery() {
	hs, err := store.NewInMemoryHelpSystem()
	if err != nil {
		log.Fatal(err)
	}
	defer hs.Close()

	// Add test data
	sections := []*model.Section{
		{
			Slug:        "cmd-help",
			Title:       "Command Help",
			SectionType: model.SectionExample,
			Topics:      []string{"commands"},
			Commands:    []string{"help", "man"},
			Flags:       []string{"--verbose", "--help"},
			Order:       1,
		},
		{
			Slug:        "config-tutorial",
			Title:       "Configuration Tutorial",
			SectionType: model.SectionTutorial,
			Topics:      []string{"configuration", "setup"},
			Commands:    []string{"config"},
			Flags:       []string{"--config"},
			Order:       2,
		},
	}

	for _, section := range sections {
		if err := hs.AddSection(section); err != nil {
			log.Fatal(err)
		}
	}

	// Complex query: Find sections that have the "help" command OR the "--help" flag
	results, err := hs.Find(store.Or(
		store.HasCommand("help"),
		store.HasFlag("--help"),
	))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sections with 'help' command or '--help' flag: %d\n", len(results))

	// Query with negation: Find all sections that are NOT tutorials
	nonTutorials, err := hs.Find(store.Not(store.IsTutorial()))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Non-tutorial sections: %d\n", len(nonTutorials))

	// Pagination example: Get first 1 result, then skip 1 and get next 1
	firstResult, err := hs.Find(store.And(
		store.OrderByOrder(),
		store.Limit(1),
	))
	if err != nil {
		log.Fatal(err)
	}

	secondResult, err := hs.Find(store.And(
		store.OrderByOrder(),
		store.Limit(1),
		store.Offset(1),
	))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("First result: %s\n", firstResult[0].Slug)
	fmt.Printf("Second result: %s\n", secondResult[0].Slug)

	// Output:
	// Sections with 'help' command or '--help' flag: 1
	// Non-tutorial sections: 1
	// First result: cmd-help
	// Second result: config-tutorial
}

// Example_legacyCompatibility shows how to use the system as a drop-in replacement
func Example_legacyCompatibility() {
	// Create help system (compatible with existing interface)
	hs, err := store.NewInMemoryHelpSystem()
	if err != nil {
		log.Fatal(err)
	}
	defer hs.Close()

	// Add a section using the legacy-compatible method
	section := &model.Section{
		Slug:        "example-section",
		Title:       "Example Section",
		SectionType: model.SectionExample,
		Topics:      []string{"examples"},
	}

	if err := hs.AddSection(section); err != nil {
		log.Fatal(err)
	}

	// Use legacy-compatible methods
	retrieved, err := hs.GetSectionWithSlug("example-section")
	if err != nil {
		log.Fatal(err)
	}

	count, err := hs.Count()
	if err != nil {
		log.Fatal(err)
	}

	// Get examples for a topic (legacy-compatible)
	examples, err := hs.GetExamplesForTopic("examples")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved section: %s\n", retrieved.Title)
	fmt.Printf("Total sections: %d\n", count)
	fmt.Printf("Examples for 'examples' topic: %d\n", len(examples))

	// Output:
	// Retrieved section: Example Section
	// Total sections: 1
	// Examples for 'examples' topic: 1
}
