package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryStore(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Test that store is empty initially
	sections, err := store.GetAllSections(ctx)
	require.NoError(t, err)
	assert.Empty(t, sections)
}

func TestAddSection(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	section := &help.Section{
		Slug:        "test-section",
		SectionType: help.SectionExample,
		Title:       "Test Section",
		SubTitle:    "A test section",
		Short:       "Short description",
		Content:     "This is test content",
		Topics:      []string{"testing", "example"},
		Flags:       []string{"--test", "--verbose"},
		Commands:    []string{"test", "run"},
		IsTopLevel:  true,
		ShowPerDefault: true,
		Order:       1,
	}
	
	err = store.AddSection(ctx, section)
	require.NoError(t, err)
	
	// Retrieve the section
	retrieved, err := store.GetSectionBySlug(ctx, "test-section")
	require.NoError(t, err)
	
	assert.Equal(t, section.Slug, retrieved.Slug)
	assert.Equal(t, section.SectionType, retrieved.SectionType)
	assert.Equal(t, section.Title, retrieved.Title)
	assert.Equal(t, section.SubTitle, retrieved.SubTitle)
	assert.Equal(t, section.Short, retrieved.Short)
	assert.Equal(t, section.Content, retrieved.Content)
	assert.Equal(t, section.Topics, retrieved.Topics)
	assert.Equal(t, section.Flags, retrieved.Flags)
	assert.Equal(t, section.Commands, retrieved.Commands)
	assert.Equal(t, section.IsTopLevel, retrieved.IsTopLevel)
	assert.Equal(t, section.ShowPerDefault, retrieved.ShowPerDefault)
	assert.Equal(t, section.Order, retrieved.Order)
}

func TestFindWithPredicates(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Add test sections
	sections := []*help.Section{
		{
			Slug:        "example-1",
			SectionType: help.SectionExample,
			Title:       "Example 1",
			Content:     "Example content 1",
			Topics:      []string{"topic1", "topic2"},
			IsTopLevel:  true,
			ShowPerDefault: true,
		},
		{
			Slug:        "example-2",
			SectionType: help.SectionExample,
			Title:       "Example 2",
			Content:     "Example content 2",
			Topics:      []string{"topic2", "topic3"},
			IsTopLevel:  false,
			ShowPerDefault: false,
		},
		{
			Slug:        "tutorial-1",
			SectionType: help.SectionTutorial,
			Title:       "Tutorial 1",
			Content:     "Tutorial content 1",
			Topics:      []string{"topic1"},
			IsTopLevel:  true,
			ShowPerDefault: true,
		},
	}
	
	for _, section := range sections {
		err = store.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test IsType predicate
	examples, err := store.Find(ctx, IsExample())
	require.NoError(t, err)
	assert.Len(t, examples, 2)
	
	tutorials, err := store.Find(ctx, IsTutorial())
	require.NoError(t, err)
	assert.Len(t, tutorials, 1)
	
	// Test HasTopic predicate
	topic1Sections, err := store.Find(ctx, HasTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, topic1Sections, 2)
	
	topic2Sections, err := store.Find(ctx, HasTopic("topic2"))
	require.NoError(t, err)
	assert.Len(t, topic2Sections, 2)
	
	topic3Sections, err := store.Find(ctx, HasTopic("topic3"))
	require.NoError(t, err)
	assert.Len(t, topic3Sections, 1)
	
	// Test IsTopLevel predicate
	topLevelSections, err := store.Find(ctx, IsTopLevel())
	require.NoError(t, err)
	assert.Len(t, topLevelSections, 2)
	
	// Test ShownByDefault predicate
	defaultSections, err := store.Find(ctx, ShownByDefault())
	require.NoError(t, err)
	assert.Len(t, defaultSections, 2)
	
	// Test NotShownByDefault predicate
	nonDefaultSections, err := store.Find(ctx, NotShownByDefault())
	require.NoError(t, err)
	assert.Len(t, nonDefaultSections, 1)
	
	// Test SlugEquals predicate
	specificSection, err := store.Find(ctx, SlugEquals("example-1"))
	require.NoError(t, err)
	assert.Len(t, specificSection, 1)
	assert.Equal(t, "example-1", specificSection[0].Slug)
}

func TestBooleanCombinators(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Add test sections
	sections := []*help.Section{
		{
			Slug:        "example-1",
			SectionType: help.SectionExample,
			Title:       "Example 1",
			Content:     "Example content 1",
			Topics:      []string{"topic1"},
			IsTopLevel:  true,
			ShowPerDefault: true,
		},
		{
			Slug:        "example-2",
			SectionType: help.SectionExample,
			Title:       "Example 2",
			Content:     "Example content 2",
			Topics:      []string{"topic2"},
			IsTopLevel:  false,
			ShowPerDefault: false,
		},
		{
			Slug:        "tutorial-1",
			SectionType: help.SectionTutorial,
			Title:       "Tutorial 1",
			Content:     "Tutorial content 1",
			Topics:      []string{"topic1"},
			IsTopLevel:  true,
			ShowPerDefault: true,
		},
	}
	
	for _, section := range sections {
		err = store.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test And combinator
	topLevelExamples, err := store.Find(ctx, And(
		IsExample(),
		IsTopLevel(),
	))
	require.NoError(t, err)
	assert.Len(t, topLevelExamples, 1)
	assert.Equal(t, "example-1", topLevelExamples[0].Slug)
	
	// Test Or combinator
	examplesOrTutorials, err := store.Find(ctx, Or(
		IsExample(),
		IsTutorial(),
	))
	require.NoError(t, err)
	assert.Len(t, examplesOrTutorials, 3)
	
	// Test Not combinator
	nonExamples, err := store.Find(ctx, Not(IsExample()))
	require.NoError(t, err)
	assert.Len(t, nonExamples, 1)
	assert.Equal(t, "tutorial-1", nonExamples[0].Slug)
	
	// Test complex query
	complex, err := store.Find(ctx, And(
		Or(
			IsExample(),
			IsTutorial(),
		),
		HasTopic("topic1"),
		IsTopLevel(),
	))
	require.NoError(t, err)
	assert.Len(t, complex, 2)
}

func TestTextSearch(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Add test sections with searchable content
	sections := []*help.Section{
		{
			Slug:        "example-1",
			SectionType: help.SectionExample,
			Title:       "Database Example",
			Content:     "This example shows how to connect to a database",
			Topics:      []string{"database"},
		},
		{
			Slug:        "example-2",
			SectionType: help.SectionExample,
			Title:       "File Processing",
			Content:     "Learn how to process files efficiently",
			Topics:      []string{"files"},
		},
		{
			Slug:        "tutorial-1",
			SectionType: help.SectionTutorial,
			Title:       "Getting Started",
			Content:     "A comprehensive guide to getting started with databases",
			Topics:      []string{"database", "beginner"},
		},
	}
	
	for _, section := range sections {
		err = store.AddSection(ctx, section)
		require.NoError(t, err)
	}
	
	// Test text search
	databaseResults, err := store.Find(ctx, TextSearch("database"))
	require.NoError(t, err)
	assert.Len(t, databaseResults, 2)
	
	fileResults, err := store.Find(ctx, TextSearch("files"))
	require.NoError(t, err)
	assert.Len(t, fileResults, 1)
	assert.Equal(t, "example-2", fileResults[0].Slug)
	
	// Test combining text search with other predicates
	databaseExamples, err := store.Find(ctx, And(
		TextSearch("database"),
		IsExample(),
	))
	require.NoError(t, err)
	assert.Len(t, databaseExamples, 1)
	assert.Equal(t, "example-1", databaseExamples[0].Slug)
}

func TestGetStats(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Initially empty
	stats, err := store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats["sections"])
	assert.Equal(t, 0, stats["topics"])
	assert.Equal(t, 0, stats["flags"])
	assert.Equal(t, 0, stats["commands"])
	
	// Add a section
	section := &help.Section{
		Slug:        "test-section",
		SectionType: help.SectionExample,
		Title:       "Test Section",
		Content:     "Test content",
		Topics:      []string{"topic1", "topic2"},
		Flags:       []string{"--flag1", "--flag2"},
		Commands:    []string{"cmd1", "cmd2"},
	}
	
	err = store.AddSection(ctx, section)
	require.NoError(t, err)
	
	// Check stats
	stats, err = store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats["sections"])
	assert.Equal(t, 2, stats["topics"])
	assert.Equal(t, 2, stats["flags"])
	assert.Equal(t, 2, stats["commands"])
}

func TestUpdateSection(t *testing.T) {
	store, err := NewInMemoryStore()
	require.NoError(t, err)
	defer store.Close()
	
	ctx := context.Background()
	
	// Add initial section
	section := &help.Section{
		Slug:        "test-section",
		SectionType: help.SectionExample,
		Title:       "Test Section",
		Content:     "Original content",
		Topics:      []string{"topic1"},
	}
	
	err = store.AddSection(ctx, section)
	require.NoError(t, err)
	
	// Update the section
	section.Title = "Updated Test Section"
	section.Content = "Updated content"
	section.Topics = []string{"topic1", "topic2"}
	
	err = store.AddSection(ctx, section)
	require.NoError(t, err)
	
	// Verify update
	retrieved, err := store.GetSectionBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Section", retrieved.Title)
	assert.Equal(t, "Updated content", retrieved.Content)
	assert.Equal(t, []string{"topic1", "topic2"}, retrieved.Topics)
	
	// Verify only one section exists
	allSections, err := store.GetAllSections(ctx)
	require.NoError(t, err)
	assert.Len(t, allSections, 1)
}
