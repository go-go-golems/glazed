package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/spf13/cobra"
)

//go:embed docs
var docsFS embed.FS

func main() {
	fmt.Println("🔧 Help System Test")
	fmt.Println("==================")

	// Test 1: Create new help system
	fmt.Println("\n1. Creating new help system...")
	hs := help.NewHelpSystem()
	defer func() { _ = hs.Store.Close() }()

	// Test 2: Load documentation from embedded filesystem
	fmt.Println("\n2. Loading documentation from embedded filesystem...")
	err := hs.LoadSectionsFromFS(docsFS, "docs")
	if err != nil {
		log.Printf("Error loading from embedded FS: %v", err)
	}

	// Test 3: Add sections programmatically
	fmt.Println("\n3. Adding sections programmatically...")
	testSection := &model.Section{
		Slug:           "test-section",
		Title:          "Test Section",
		Short:          "A test section for demonstration",
		Content:        "# Test Section\n\nThis is a test section with **markdown** content.",
		SectionType:    model.SectionExample,
		Topics:         []string{"testing", "example"},
		Commands:       []string{"test"},
		Flags:          []string{"--verbose"},
		IsTopLevel:     true,
		ShowPerDefault: true,
		Order:          1,
	}

	err = hs.Store.Insert(context.Background(), testSection)
	if err != nil {
		log.Printf("Error adding section: %v", err)
	}

	// Test 4: Query by section type
	fmt.Println("\n4. Testing DSL queries...")

	examples, err := hs.QuerySections("type:example")
	if err != nil {
		log.Printf("Error querying examples: %v", err)
	} else {
		fmt.Printf("Found %d examples:\n", len(examples))
		for _, section := range examples {
			fmt.Printf("  - %s: %s\n", section.Slug, section.Short)
		}
	}

	// Test 5: Boolean logic queries
	fmt.Println("\n5. Testing boolean logic...")

	results, err := hs.QuerySections("type:example AND topic:testing")
	if err != nil {
		log.Printf("Error with boolean query: %v", err)
	} else {
		fmt.Printf("Found %d examples about testing:\n", len(results))
		for _, section := range results {
			fmt.Printf("  - %s: %s\n", section.Slug, section.Short)
		}
	}

	// Test 6: Individual section retrieval
	fmt.Println("\n6. Testing individual section retrieval...")
	section, err := hs.GetSectionWithSlug("test-section")
	if err != nil {
		log.Printf("Error getting section: %v", err)
	} else {
		fmt.Printf("Retrieved section: %s\n", section.Title)
		fmt.Printf("  Type: %s\n", section.SectionType.String())
		fmt.Printf("  Topics: %v\n", section.Topics)
		fmt.Printf("  Commands: %v\n", section.Commands)
	}

	// Test 7: Text search
	fmt.Println("\n7. Testing text search...")
	searchResults, err := hs.QuerySections(`"markdown"`)
	if err != nil {
		log.Printf("Error with text search: %v", err)
	} else {
		fmt.Printf("Found %d sections containing 'markdown':\n", len(searchResults))
		for _, section := range searchResults {
			fmt.Printf("  - %s: %s\n", section.Slug, section.Short)
		}
	}

	// Test 8: List all sections
	fmt.Println("\n8. Listing all sections...")
	allSections, err := hs.QuerySections("")
	if err != nil {
		log.Printf("Error listing all sections: %v", err)
	} else {
		fmt.Printf("Total sections: %d\n", len(allSections))
		for _, section := range allSections {
			fmt.Printf("  - %s (%s): %s\n", section.Slug, section.SectionType.String(), section.Short)
		}
	}

	// Test 9: Cobra integration example
	fmt.Println("\n9. Testing Cobra integration...")
	setupCobraExample(hs)

	fmt.Println("\n✅ Help system test completed!")
}

func setupCobraExample(hs *help.HelpSystem) {
	rootCmd := &cobra.Command{
		Use:   "help-test",
		Short: "Help system test CLI",
	}

	// Add help command with query support
	helpCmd := &cobra.Command{
		Use:   "help [topic]",
		Short: "Help about any command or topic",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				// Show all topics
				sections, err := hs.QuerySections("")
				if err != nil {
					fmt.Printf("Error querying sections: %v\n", err)
					return
				}

				fmt.Println("Available help topics:")
				for _, section := range sections {
					fmt.Printf("  %s - %s\n", section.Slug, section.Short)
				}
				return
			}

			// Look up specific section
			section, err := hs.GetSectionWithSlug(args[0])
			if err != nil {
				fmt.Printf("Help topic '%s' not found\n", args[0])
				return
			}

			fmt.Printf("# %s\n\n%s\n", section.Title, section.Content)
		},
	}

	rootCmd.AddCommand(helpCmd)

	fmt.Println("Cobra integration example set up (commands not executed in test)")
}
