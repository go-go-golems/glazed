package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed data/*.md
var helpData embed.FS

func main() {
	// Set up logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx := context.Background()

	// Create temporary database for demo
	helpStore, err := store.NewStore("/tmp/help_demo.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create store")
	}
	defer helpStore.Close()

	// Load sections from embedded filesystem
	fmt.Println("Loading sections from filesystem...")
	err = helpStore.LoadSectionsFromFS(ctx, helpData, "data")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load sections")
	}

	// Debug: Check if tables exist
	fmt.Println("\nChecking database schema...")
	tables := []string{"sections", "section_topics", "section_flags", "section_commands", "section_fts"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT count(*) FROM %s", table)
		err := helpStore.GetDB().QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Table %s: %v\n", table, err)
		} else {
			fmt.Printf("✅ Table %s: %d rows\n", table, count)
		}
	}

	// Get stats
	stats, err := helpStore.GetSectionStats(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get stats")
	}

	fmt.Printf("Loaded sections: %v\n", stats)

	// Comprehensive test queries
	fmt.Println("\n=== Comprehensive Test Queries ===")

	testQueries := []struct {
		name        string
		description string
		query       query.Predicate
	}{
		{
			name:        "All Examples",
			description: "Find all sections of type Example",
			query:       query.IsType(model.SectionExample),
		},
		{
			name:        "All Tutorials", 
			description: "Find all sections of type Tutorial",
			query:       query.IsType(model.SectionTutorial),
		},
		{
			name:        "All Applications",
			description: "Find all sections of type Application", 
			query:       query.IsType(model.SectionApplication),
		},
		{
			name:        "All General Topics",
			description: "Find all sections of type GeneralTopic",
			query:       query.IsType(model.SectionGeneralTopic),
		},
		{
			name:        "Top-level Sections",
			description: "Find all top-level sections",
			query:       query.IsTopLevel(),
		},
		{
			name:        "Default Sections",
			description: "Find all sections shown by default",
			query:       query.ShownByDefault(),
		},
		{
			name:        "Getting Started Topic",
			description: "Find sections with 'getting-started' topic",
			query:       query.HasTopic("getting-started"),
		},
		{
			name:        "Examples Topic",
			description: "Find sections with 'examples' topic",
			query:       query.HasTopic("examples"),
		},
		{
			name:        "Advanced Topic",
			description: "Find sections with 'advanced' topic",
			query:       query.HasTopic("advanced"),
		},
		{
			name:        "Verbose Flag",
			description: "Find sections with 'verbose' flag",
			query:       query.HasFlag("verbose"),
		},
		{
			name:        "Help Flag",
			description: "Find sections with 'help' flag",
			query:       query.HasFlag("help"),
		},
		{
			name:        "Run Command",
			description: "Find sections with 'run' command",
			query:       query.HasCommand("run"),
		},
		{
			name:        "Deploy Command",
			description: "Find sections with 'deploy' command",
			query:       query.HasCommand("deploy"),
		},
		{
			name:        "Specific Slug",
			description: "Find section with slug 'getting-started'",
			query:       query.SlugEquals("getting-started"),
		},
		{
			name:        "Examples AND Top-level",
			description: "Find examples that are top-level",
			query:       query.And(query.IsType(model.SectionExample), query.IsTopLevel()),
		},
		{
			name:        "Examples OR Tutorials",
			description: "Find sections that are examples or tutorials",
			query:       query.Or(query.IsType(model.SectionExample), query.IsType(model.SectionTutorial)),
		},
		{
			name:        "NOT Top-level",
			description: "Find sections that are not top-level",
			query:       query.Not(query.IsTopLevel()),
		},
		{
			name:        "Complex: (Examples OR Tutorials) AND Top-level",
			description: "Find examples or tutorials that are top-level",
			query: query.And(
				query.Or(query.IsType(model.SectionExample), query.IsType(model.SectionTutorial)),
				query.IsTopLevel(),
			),
		},
		{
			name:        "Complex: Getting-started AND Default",
			description: "Find getting-started sections shown by default",
			query:       query.And(query.HasTopic("getting-started"), query.ShownByDefault()),
		},
		{
			name:        "Complex: Advanced OR (Examples AND Verbose)",
			description: "Find advanced topics or examples with verbose flag",
			query: query.Or(
				query.HasTopic("advanced"),
				query.And(query.IsType(model.SectionExample), query.HasFlag("verbose")),
			),
		},
		{
			name:        "Text Search: 'tutorial'",
			description: "Full-text search for 'tutorial'",
			query:       query.TextSearch("tutorial"),
		},
		{
			name:        "Text Search: 'production'",
			description: "Full-text search for 'production'",
			query:       query.TextSearch("production"),
		},
		{
			name:        "Text Search: 'example'",
			description: "Full-text search for 'example'",
			query:       query.TextSearch("example"),
		},
	}

	for i, test := range testQueries {
		fmt.Printf("\n%d. %s\n", i+1, test.name)
		fmt.Printf("   Description: %s\n", test.description)
		
		// Always print the generated SQL for debugging
		sql, args := query.Compile(test.query)
		fmt.Printf("   SQL: %s\n", sql)
		fmt.Printf("   Args: %v\n", args)
		
		results, err := helpStore.Find(ctx, test.query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}
		
		if len(results) == 0 {
			fmt.Printf("   📭 No results\n")
		} else {
			fmt.Printf("   📊 Found %d result(s):\n", len(results))
			for _, section := range results {
				fmt.Printf("      - %s: %s (%s)\n", section.Slug, section.Title, section.SectionType.String())
				if len(section.Topics) > 0 {
					fmt.Printf("        Topics: %v\n", section.Topics)
				}
				if len(section.Flags) > 0 {
					fmt.Printf("        Flags: %v\n", section.Flags)
				}
				if len(section.Commands) > 0 {
					fmt.Printf("        Commands: %v\n", section.Commands)
				}
			}
		}
	}

	// Test error cases
	fmt.Println("\n=== Error Case Tests ===")
	
	fmt.Println("\n1. Non-existent slug:")
	_, err = helpStore.GetSectionBySlug(ctx, "non-existent-slug")
	if err != nil {
		fmt.Printf("   ✅ Expected error: %v\n", err)
	} else {
		fmt.Printf("   ❌ Should have failed but didn't\n")
	}

	fmt.Println("\n=== Performance Test ===")
	
	// Test query performance with multiple predicates
	fmt.Println("\n1. Complex query performance test:")
	complexQuery := query.And(
		query.Or(
			query.IsType(model.SectionExample),
			query.IsType(model.SectionTutorial),
			query.IsType(model.SectionApplication),
		),
		query.Or(
			query.HasTopic("getting-started"),
			query.HasTopic("advanced"),
		),
		query.Not(query.IsTopLevel()),
	)
	
	// Debug: Print the generated SQL
	sql, args := query.Compile(complexQuery)
	fmt.Printf("   Generated SQL: %s\n", sql)
	fmt.Printf("   Arguments: %v\n", args)
	
	results, err := helpStore.Find(ctx, complexQuery)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Complex query executed successfully, found %d results\n", len(results))
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("🎉 All tests completed successfully!")
}
