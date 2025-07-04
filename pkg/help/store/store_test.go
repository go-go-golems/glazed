package store

import (
	"context"
	"os"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	// Create a temporary database
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create store
	store, err := NewStore(tmpfile.Name())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test data
	section1 := &model.Section{
		Slug:        "test-example",
		Title:       "Test Example",
		Subtitle:    "A test example",
		Short:       "This is a test",
		Content:     "This is the content of the test example",
		SectionType: model.SectionExample,
		IsTopLevel:  true,
		IsTemplate:  false,
		ShowDefault: true,
		Order:       1,
		Topics:      []string{"testing", "examples"},
		Flags:       []string{"verbose", "debug"},
		Commands:    []string{"test", "example"},
	}

	section2 := &model.Section{
		Slug:        "tutorial-basics",
		Title:       "Tutorial Basics",
		Subtitle:    "Basic tutorial",
		Short:       "Learn the basics",
		Content:     "This tutorial covers the basics",
		SectionType: model.SectionTutorial,
		IsTopLevel:  false,
		IsTemplate:  false,
		ShowDefault: false,
		Order:       2,
		Topics:      []string{"tutorials", "basics"},
		Flags:       []string{"help"},
		Commands:    []string{"tutorial"},
	}

	// Test UpsertSection
	err = store.UpsertSection(ctx, section1)
	require.NoError(t, err)
	assert.NotZero(t, section1.ID)

	err = store.UpsertSection(ctx, section2)
	require.NoError(t, err)
	assert.NotZero(t, section2.ID)

	// Test GetSectionBySlug
	retrieved, err := store.GetSectionBySlug(ctx, "test-example")
	require.NoError(t, err)
	assert.Equal(t, section1.Slug, retrieved.Slug)
	assert.Equal(t, section1.Title, retrieved.Title)
	assert.Equal(t, section1.Topics, retrieved.Topics)
	assert.Equal(t, section1.Flags, retrieved.Flags)
	assert.Equal(t, section1.Commands, retrieved.Commands)

	// Test Find with IsType predicate
	examples, err := store.Find(ctx, query.IsType(model.SectionExample))
	require.NoError(t, err)
	assert.Len(t, examples, 1)
	assert.Equal(t, "test-example", examples[0].Slug)

	// Test Find with HasTopic predicate
	testingSections, err := store.Find(ctx, query.HasTopic("testing"))
	require.NoError(t, err)
	assert.Len(t, testingSections, 1)
	assert.Equal(t, "test-example", testingSections[0].Slug)

	// Test Find with HasFlag predicate
	verboseSections, err := store.Find(ctx, query.HasFlag("verbose"))
	require.NoError(t, err)
	assert.Len(t, verboseSections, 1)
	assert.Equal(t, "test-example", verboseSections[0].Slug)

	// Test Find with HasCommand predicate
	testSections, err := store.Find(ctx, query.HasCommand("test"))
	require.NoError(t, err)
	assert.Len(t, testSections, 1)
	assert.Equal(t, "test-example", testSections[0].Slug)

	// Test Find with IsTopLevel predicate
	topLevelSections, err := store.Find(ctx, query.IsTopLevel())
	require.NoError(t, err)
	assert.Len(t, topLevelSections, 1)
	assert.Equal(t, "test-example", topLevelSections[0].Slug)

	// Test Find with ShownByDefault predicate
	defaultSections, err := store.Find(ctx, query.ShownByDefault())
	require.NoError(t, err)
	assert.Len(t, defaultSections, 1)
	assert.Equal(t, "test-example", defaultSections[0].Slug)

	// Test Find with complex query (And)
	complexSections, err := store.Find(ctx, query.And(
		query.IsType(model.SectionExample),
		query.HasTopic("testing"),
		query.IsTopLevel(),
	))
	require.NoError(t, err)
	assert.Len(t, complexSections, 1)
	assert.Equal(t, "test-example", complexSections[0].Slug)

	// Test Find with Or query
	orSections, err := store.Find(ctx, query.Or(
		query.IsType(model.SectionExample),
		query.IsType(model.SectionTutorial),
	))
	require.NoError(t, err)
	assert.Len(t, orSections, 2)

	// Test Find with Not query
	notTopLevelSections, err := store.Find(ctx, query.Not(query.IsTopLevel()))
	require.NoError(t, err)
	assert.Len(t, notTopLevelSections, 1)
	assert.Equal(t, "tutorial-basics", notTopLevelSections[0].Slug)

	// Test TextSearch (basic functionality)
	searchSections, err := store.Find(ctx, query.TextSearch("tutorial"))
	require.NoError(t, err)
	assert.Len(t, searchSections, 1)
	assert.Equal(t, "tutorial-basics", searchSections[0].Slug)

	// Test non-existent section
	_, err = store.GetSectionBySlug(ctx, "non-existent")
	assert.Error(t, err)

	// Test update (upsert existing)
	section1.Title = "Updated Test Example"
	err = store.UpsertSection(ctx, section1)
	require.NoError(t, err)

	updated, err := store.GetSectionBySlug(ctx, "test-example")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Example", updated.Title)
}

func TestStoreWithMarkdownSection(t *testing.T) {
	// Create a temporary database
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create store
	store, err := NewStore(tmpfile.Name())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test markdown parsing
	markdown := `---
Title: Markdown Test
Slug: markdown-test
SubTitle: Test parsing
Short: A markdown test
SectionType: Example
Topics: [markdown, test]
Flags: [verbose]
Commands: [parse]
IsTopLevel: true
ShowPerDefault: true
Order: 1
---

# Markdown Test

This is a test of markdown parsing.
`

	section, err := model.LoadSectionFromMarkdown([]byte(markdown))
	require.NoError(t, err)

	err = store.UpsertSection(ctx, section)
	require.NoError(t, err)

	// Verify it was stored correctly
	retrieved, err := store.GetSectionBySlug(ctx, "markdown-test")
	require.NoError(t, err)
	assert.Equal(t, "Markdown Test", retrieved.Title)
	assert.Equal(t, "Test parsing", retrieved.Subtitle)
	assert.Equal(t, model.SectionExample, retrieved.SectionType)
	assert.Contains(t, retrieved.Topics, "markdown")
	assert.Contains(t, retrieved.Flags, "verbose")
	assert.Contains(t, retrieved.Commands, "parse")
	assert.True(t, retrieved.IsTopLevel)
	assert.True(t, retrieved.ShowDefault)
}
