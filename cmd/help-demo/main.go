package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/rs/zerolog"
)

func main() {
	// Set up logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	// Create a new SQLite help store
	helpStore, err := store.NewInMemory()
	if err != nil {
		log.Fatalf("Failed to create help store: %v", err)
	}
	defer helpStore.Close()

	// Create some sample sections
	sections := []*model.Section{
		{
			Slug:        "getting-started",
			Title:       "Getting Started Guide",
			Subtitle:    "Learn the basics",
			Short:       "A quick introduction to the CLI tool",
			Content:     "This guide covers the fundamental concepts and basic usage patterns. You'll learn how to use templates, configure settings, and run your first commands.",
			SectionType: model.SectionTutorial,
			IsTopLevel:  true,
			ShowDefault: true,
			Order:       1,
			Topics:      []string{"basics", "tutorial"},
			Flags:       []string{"help", "version"},
			Commands:    []string{"init", "help"},
		},
		{
			Slug:        "template-example",
			Title:       "Basic Template Example",
			Short:       "Shows how to create a simple template",
			Content:     "Here's how to create your first template:\n\n```yaml\nname: my-template\ndata:\n  message: Hello World\n```\n\nThis template demonstrates basic variable substitution and YAML structure.",
			SectionType: model.SectionExample,
			IsTopLevel:  false,
			ShowDefault: true,
			Order:       2,
			Topics:      []string{"templates", "examples"},
			Flags:       []string{"output", "format"},
			Commands:    []string{"process", "template"},
		},
		{
			Slug:        "advanced-templates",
			Title:       "Advanced Template Techniques",
			Short:       "Complex template patterns and best practices",
			Content:     "Advanced templating features include loops, conditionals, and custom functions. This section covers performance optimization and debugging techniques.",
			SectionType: model.SectionTutorial,
			IsTopLevel:  false,
			ShowDefault: false,
			Order:       3,
			Topics:      []string{"templates", "advanced"},
			Flags:       []string{"debug", "verbose"},
			Commands:    []string{"process", "template", "validate"},
		},
		{
			Slug:        "api-integration",
			Title:       "API Integration Application",
			Short:       "Real-world example of API data processing",
			Content:     "This application shows how to fetch data from REST APIs, transform it using templates, and output structured results. Includes error handling and rate limiting.",
			SectionType: model.SectionApplication,
			IsTopLevel:  false,
			ShowDefault: false,
			Order:       4,
			Topics:      []string{"api", "integration", "examples"},
			Flags:       []string{"timeout", "retry"},
			Commands:    []string{"fetch", "process"},
		},
	}

	// Load sections into the store
	fmt.Println("Loading help sections...")
	for _, section := range sections {
		if err := helpStore.Upsert(ctx, section); err != nil {
			log.Fatalf("Failed to upsert section %s: %v", section.Slug, err)
		}
		fmt.Printf("  ✓ Loaded: %s\n", section.Title)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SQLite Help System Demo")
	fmt.Println(strings.Repeat("=", 60))

	// Demo 1: Find all top-level sections
	fmt.Println("\n1. Top-level help sections:")
	topLevelSections, err := helpStore.Find(ctx, query.IsTopLevel())
	if err != nil {
		log.Fatalf("Failed to find top-level sections: %v", err)
	}
	for _, section := range topLevelSections {
		fmt.Printf("  • %s - %s\n", section.Title, section.Short)
	}

	// Demo 2: Find all template-related content
	fmt.Println("\n2. Template-related content:")
	templateSections, err := helpStore.Find(ctx, query.HasTopic("templates"))
	if err != nil {
		log.Fatalf("Failed to find template sections: %v", err)
	}
	for _, section := range templateSections {
		fmt.Printf("  • %s (%s) - %s\n", section.Title, section.SectionType.String(), section.Short)
	}

	// Demo 3: Find examples shown by default
	fmt.Println("\n3. Default examples:")
	defaultExamples, err := helpStore.Find(ctx, query.And(
		query.IsType(model.SectionExample),
		query.ShownByDefault(),
	))
	if err != nil {
		log.Fatalf("Failed to find default examples: %v", err)
	}
	for _, section := range defaultExamples {
		fmt.Printf("  • %s - %s\n", section.Title, section.Short)
	}

	// Demo 4: Complex query - tutorials or examples about templates
	fmt.Println("\n4. Tutorials or examples about templates:")
	complexSections, err := helpStore.Find(ctx, query.And(
		query.Or(
			query.IsType(model.SectionTutorial),
			query.IsType(model.SectionExample),
		),
		query.HasTopic("templates"),
	))
	if err != nil {
		log.Fatalf("Failed to execute complex query: %v", err)
	}
	for _, section := range complexSections {
		fmt.Printf("  • %s (%s) - %s\n", section.Title, section.SectionType.String(), section.Short)
	}

	// Demo 5: Full-text search
	fmt.Println("\n5. Full-text search for 'template':")
	searchResults, err := helpStore.Find(ctx, query.TextSearch("template*"))
	if err != nil {
		log.Fatalf("Failed to perform text search: %v", err)
	}
	for _, section := range searchResults {
		fmt.Printf("  • %s - %s\n", section.Title, section.Short)
	}

	// Demo 6: Find sections for specific command
	fmt.Println("\n6. Help for 'process' command:")
	processHelp, err := helpStore.Find(ctx, query.HasCommand("process"))
	if err != nil {
		log.Fatalf("Failed to find process command help: %v", err)
	}
	for _, section := range processHelp {
		fmt.Printf("  • %s (%s) - %s\n", section.Title, section.SectionType.String(), section.Short)
	}

	// Demo 7: Statistics
	fmt.Println("\n7. Help system statistics:")
	stats, err := helpStore.Stats(ctx)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	for key, value := range stats {
		fmt.Printf("  • %s: %d\n", key, value)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Demo completed successfully! 🎉")
	fmt.Println(strings.Repeat("=", 60))
}
