package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryHelpSystem(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Test that help system is empty initially
	stats, err := hs.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats["sections"])
}

func TestHelpSystemAddSection(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	section := &help.Section{
		Slug:        "test-section",
		SectionType: help.SectionExample,
		Title:       "Test Section",
		Content:     "Test content",
		Topics:      []string{"testing"},
		IsTopLevel:  true,
	}
	
	err = hs.AddSection(ctx, section)
	require.NoError(t, err)
	
	// Retrieve the section
	retrieved, err := hs.GetSectionBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, section.Slug, retrieved.Slug)
	assert.Equal(t, section.Title, retrieved.Title)
}

func TestHelpSystemGetSectionsByType(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections of different types
	sections := []*help.Section{
		{
			Slug:        "example-1",
			SectionType: help.SectionExample,
			Title:       "Example 1",
			Content:     "Example content",
		},
		{
			Slug:        "example-2",
			SectionType: help.SectionExample,
			Title:       "Example 2",
			Content:     "Example content",
		},
		{
			Slug:        "tutorial-1",
			SectionType: help.SectionTutorial,
			Title:       "Tutorial 1",
			Content:     "Tutorial content",
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting sections by type
	examples, err := hs.GetSectionsByType(ctx, help.SectionExample)
	require.NoError(t, err)
	assert.Len(t, examples, 2)
	
	tutorials, err := hs.GetSectionsByType(ctx, help.SectionTutorial)
	require.NoError(t, err)
	assert.Len(t, tutorials, 1)
	
	applications, err := hs.GetSectionsByType(ctx, help.SectionApplication)
	require.NoError(t, err)
	assert.Len(t, applications, 0)
}

func TestHelpSystemGetSectionsByTopic(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with different topics
	sections := []*help.Section{
		{
			Slug:        "section-1",
			SectionType: help.SectionExample,
			Title:       "Section 1",
			Content:     "Content 1",
			Topics:      []string{"topic1", "topic2"},
		},
		{
			Slug:        "section-2",
			SectionType: help.SectionExample,
			Title:       "Section 2",
			Content:     "Content 2",
			Topics:      []string{"topic2", "topic3"},
		},
		{
			Slug:        "section-3",
			SectionType: help.SectionTutorial,
			Title:       "Section 3",
			Content:     "Content 3",
			Topics:      []string{"topic1"},
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting sections by topic
	topic1Sections, err := hs.GetSectionsByTopic(ctx, "topic1")
	require.NoError(t, err)
	assert.Len(t, topic1Sections, 2)
	
	topic2Sections, err := hs.GetSectionsByTopic(ctx, "topic2")
	require.NoError(t, err)
	assert.Len(t, topic2Sections, 2)
	
	topic3Sections, err := hs.GetSectionsByTopic(ctx, "topic3")
	require.NoError(t, err)
	assert.Len(t, topic3Sections, 1)
	
	nonExistentTopic, err := hs.GetSectionsByTopic(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Len(t, nonExistentTopic, 0)
}

func TestHelpSystemGetSectionsByFlag(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with different flags
	sections := []*help.Section{
		{
			Slug:        "section-1",
			SectionType: help.SectionExample,
			Title:       "Section 1",
			Content:     "Content 1",
			Flags:       []string{"--verbose", "--debug"},
		},
		{
			Slug:        "section-2",
			SectionType: help.SectionExample,
			Title:       "Section 2",
			Content:     "Content 2",
			Flags:       []string{"--verbose", "--output"},
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting sections by flag
	verboseSections, err := hs.GetSectionsByFlag(ctx, "--verbose")
	require.NoError(t, err)
	assert.Len(t, verboseSections, 2)
	
	debugSections, err := hs.GetSectionsByFlag(ctx, "--debug")
	require.NoError(t, err)
	assert.Len(t, debugSections, 1)
	
	outputSections, err := hs.GetSectionsByFlag(ctx, "--output")
	require.NoError(t, err)
	assert.Len(t, outputSections, 1)
}

func TestHelpSystemGetSectionsByCommand(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with different commands
	sections := []*help.Section{
		{
			Slug:        "section-1",
			SectionType: help.SectionExample,
			Title:       "Section 1",
			Content:     "Content 1",
			Commands:    []string{"build", "test"},
		},
		{
			Slug:        "section-2",
			SectionType: help.SectionExample,
			Title:       "Section 2",
			Content:     "Content 2",
			Commands:    []string{"build", "deploy"},
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting sections by command
	buildSections, err := hs.GetSectionsByCommand(ctx, "build")
	require.NoError(t, err)
	assert.Len(t, buildSections, 2)
	
	testSections, err := hs.GetSectionsByCommand(ctx, "test")
	require.NoError(t, err)
	assert.Len(t, testSections, 1)
	
	deploySections, err := hs.GetSectionsByCommand(ctx, "deploy")
	require.NoError(t, err)
	assert.Len(t, deploySections, 1)
}

func TestHelpSystemSearchSections(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with searchable content
	sections := []*help.Section{
		{
			Slug:        "section-1",
			SectionType: help.SectionExample,
			Title:       "Database Connection",
			Content:     "This section shows how to connect to a database",
		},
		{
			Slug:        "section-2",
			SectionType: help.SectionExample,
			Title:       "File Processing",
			Content:     "Learn how to process files efficiently",
		},
		{
			Slug:        "section-3",
			SectionType: help.SectionTutorial,
			Title:       "Advanced Database Queries",
			Content:     "A comprehensive guide to advanced database operations",
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test text search
	databaseResults, err := hs.SearchSections(ctx, "database")
	require.NoError(t, err)
	assert.Len(t, databaseResults, 2)
	
	fileResults, err := hs.SearchSections(ctx, "files")
	require.NoError(t, err)
	assert.Len(t, fileResults, 1)
	assert.Equal(t, "section-2", fileResults[0].Slug)
	
	processingResults, err := hs.SearchSections(ctx, "processing")
	require.NoError(t, err)
	assert.Len(t, processingResults, 1)
	assert.Equal(t, "section-2", processingResults[0].Slug)
}

func TestHelpSystemGetTopLevelSections(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with different top-level settings
	sections := []*help.Section{
		{
			Slug:        "section-1",
			SectionType: help.SectionExample,
			Title:       "Section 1",
			Content:     "Content 1",
			IsTopLevel:  true,
		},
		{
			Slug:        "section-2",
			SectionType: help.SectionExample,
			Title:       "Section 2",
			Content:     "Content 2",
			IsTopLevel:  false,
		},
		{
			Slug:        "section-3",
			SectionType: help.SectionTutorial,
			Title:       "Section 3",
			Content:     "Content 3",
			IsTopLevel:  true,
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting top-level sections
	topLevelSections, err := hs.GetTopLevelSections(ctx)
	require.NoError(t, err)
	assert.Len(t, topLevelSections, 2)
	
	// Verify the correct sections are returned
	slugs := make([]string, len(topLevelSections))
	for i, section := range topLevelSections {
		slugs[i] = section.Slug
	}
	assert.Contains(t, slugs, "section-1")
	assert.Contains(t, slugs, "section-3")
}

func TestHelpSystemGetDefaultSections(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add sections with different default settings
	sections := []*help.Section{
		{
			Slug:           "section-1",
			SectionType:    help.SectionExample,
			Title:          "Section 1",
			Content:        "Content 1",
			ShowPerDefault: true,
		},
		{
			Slug:           "section-2",
			SectionType:    help.SectionExample,
			Title:          "Section 2",
			Content:        "Content 2",
			ShowPerDefault: false,
		},
		{
			Slug:           "section-3",
			SectionType:    help.SectionTutorial,
			Title:          "Section 3",
			Content:        "Content 3",
			ShowPerDefault: true,
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test getting default sections
	defaultSections, err := hs.GetDefaultSections(ctx)
	require.NoError(t, err)
	assert.Len(t, defaultSections, 2)
	
	// Verify the correct sections are returned
	slugs := make([]string, len(defaultSections))
	for i, section := range defaultSections {
		slugs[i] = section.Slug
	}
	assert.Contains(t, slugs, "section-1")
	assert.Contains(t, slugs, "section-3")
}

func TestHelpSystemExampleQueries(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add comprehensive test data
	sections := []*help.Section{
		{
			Slug:           "example-1",
			SectionType:    help.SectionExample,
			Title:          "Example 1",
			Content:        "Example content",
			Topics:         []string{"topic1"},
			IsTopLevel:     true,
			ShowPerDefault: true,
		},
		{
			Slug:           "example-2",
			SectionType:    help.SectionExample,
			Title:          "Example 2",
			Content:        "Example content",
			Topics:         []string{"topic1"},
			IsTopLevel:     false,
			ShowPerDefault: false,
		},
		{
			Slug:           "tutorial-1",
			SectionType:    help.SectionTutorial,
			Title:          "Tutorial 1",
			Content:        "Tutorial content",
			Topics:         []string{"topic1"},
			IsTopLevel:     true,
			ShowPerDefault: true,
		},
		{
			Slug:           "tutorial-2",
			SectionType:    help.SectionTutorial,
			Title:          "Tutorial 2",
			Content:        "Tutorial content",
			Topics:         []string{"topic2"},
			IsTopLevel:     false,
			ShowPerDefault: false,
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test GetExampleSectionsForTopic
	exampleSections, err := hs.GetExampleSectionsForTopic(ctx, "topic1")
	require.NoError(t, err)
	assert.Len(t, exampleSections, 2)
	
	// Test GetTutorialsAndExamples
	tutorialsAndExamples, err := hs.GetTutorialsAndExamples(ctx)
	require.NoError(t, err)
	assert.Len(t, tutorialsAndExamples, 4)
	
	// Test GetTopLevelDefaultSections
	topLevelDefault, err := hs.GetTopLevelDefaultSections(ctx)
	require.NoError(t, err)
	assert.Len(t, topLevelDefault, 2)
	
	// Test GetNonDefaultExamples
	nonDefaultExamples, err := hs.GetNonDefaultExamples(ctx)
	require.NoError(t, err)
	assert.Len(t, nonDefaultExamples, 1)
	assert.Equal(t, "example-2", nonDefaultExamples[0].Slug)
}

func TestHelpSystemGetHelpPages(t *testing.T) {
	hs, err := NewInMemoryHelpSystem()
	require.NoError(t, err)
	defer hs.Close()
	
	ctx := context.Background()
	
	// Add test sections
	sections := []*help.Section{
		{
			Slug:           "example-1",
			SectionType:    help.SectionExample,
			Title:          "Example 1",
			Content:        "Example content",
			Topics:         []string{"topic1"},
			Commands:       []string{"cmd1"},
			Flags:          []string{"--flag1"},
			IsTopLevel:     true,
			ShowPerDefault: true,
			Order:          1,
		},
		{
			Slug:           "tutorial-1",
			SectionType:    help.SectionTutorial,
			Title:          "Tutorial 1",
			Content:        "Tutorial content",
			Topics:         []string{"topic1"},
			Commands:       []string{"cmd1"},
			IsTopLevel:     true,
			ShowPerDefault: false,
			Order:          2,
		},
	}
	
	for _, section := range sections {
		err = hs.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test GetTopLevelHelpPage
	topLevelHelpPage, err := hs.GetTopLevelHelpPage(ctx)
	require.NoError(t, err)
	assert.NotNil(t, topLevelHelpPage)
	assert.Len(t, topLevelHelpPage.DefaultExamples, 1)
	assert.Len(t, topLevelHelpPage.OtherTutorials, 1)
	
	// Test GetHelpPageForTopic
	topicHelpPage, err := hs.GetHelpPageForTopic(ctx, "topic1")
	require.NoError(t, err)
	assert.NotNil(t, topicHelpPage)
	assert.Len(t, topicHelpPage.AllExamples, 1)
	assert.Len(t, topicHelpPage.AllTutorials, 1)
	
	// Test GetHelpPageForCommand
	commandHelpPage, err := hs.GetHelpPageForCommand(ctx, "cmd1")
	require.NoError(t, err)
	assert.NotNil(t, commandHelpPage)
	assert.Len(t, commandHelpPage.AllExamples, 1)
	assert.Len(t, commandHelpPage.AllTutorials, 1)
	
	// Test GetHelpPageForFlag
	flagHelpPage, err := hs.GetHelpPageForFlag(ctx, "--flag1")
	require.NoError(t, err)
	assert.NotNil(t, flagHelpPage)
	assert.Len(t, flagHelpPage.AllExamples, 1)
	assert.Len(t, flagHelpPage.AllTutorials, 0)
}
