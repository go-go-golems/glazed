package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_BasicOperations(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test section
	section := &model.Section{
		Slug:        "test-section",
		Title:       "Test Section",
		Subtitle:    "A test section",
		Short:       "Short description",
		Content:     "This is the content of the test section",
		SectionType: model.SectionExample,
		IsTopLevel:  true,
		ShowDefault: true,
		Order:       1,
		Topics:      []string{"testing", "example"},
		Flags:       []string{"verbose", "debug"},
		Commands:    []string{"test", "example"},
	}

	// Test upsert
	err = store.Upsert(ctx, section)
	require.NoError(t, err)
	assert.NotZero(t, section.ID)

	// Test get by slug
	retrieved, err := store.GetBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, section.Slug, retrieved.Slug)
	assert.Equal(t, section.Title, retrieved.Title)
	assert.Equal(t, section.SectionType, retrieved.SectionType)
	assert.Equal(t, section.Topics, retrieved.Topics)
	assert.Equal(t, section.Flags, retrieved.Flags)
	assert.Equal(t, section.Commands, retrieved.Commands)

	// Test update
	section.Title = "Updated Test Section"
	err = store.Upsert(ctx, section)
	require.NoError(t, err)

	updated, err := store.GetBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Section", updated.Title)

	// Test delete
	err = store.Delete(ctx, "test-section")
	require.NoError(t, err)

	_, err = store.GetBySlug(ctx, "test-section")
	assert.Error(t, err)
}

func TestStore_Find(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test sections
	sections := []*model.Section{
		{
			Slug:        "example-1",
			Title:       "Example 1",
			SectionType: model.SectionExample,
			IsTopLevel:  true,
			ShowDefault: true,
			Topics:      []string{"topic1", "common"},
			Flags:       []string{"flag1"},
			Commands:    []string{"cmd1"},
		},
		{
			Slug:        "example-2",
			Title:       "Example 2",
			SectionType: model.SectionExample,
			IsTopLevel:  false,
			ShowDefault: false,
			Topics:      []string{"topic2", "common"},
			Flags:       []string{"flag2"},
			Commands:    []string{"cmd2"},
		},
		{
			Slug:        "tutorial-1",
			Title:       "Tutorial 1",
			SectionType: model.SectionTutorial,
			IsTopLevel:  true,
			ShowDefault: true,
			Topics:      []string{"topic1"},
			Flags:       []string{"flag1"},
			Commands:    []string{"cmd1"},
		},
	}

	// Insert test sections
	for _, section := range sections {
		err = store.Upsert(ctx, section)
		require.NoError(t, err)
	}

	// Test finding by type
	examples, err := store.Find(ctx, query.IsType(model.SectionExample))
	require.NoError(t, err)
	assert.Len(t, examples, 2)

	tutorials, err := store.Find(ctx, query.IsType(model.SectionTutorial))
	require.NoError(t, err)
	assert.Len(t, tutorials, 1)

	// Test finding by topic
	topic1Sections, err := store.Find(ctx, query.HasTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, topic1Sections, 2)

	// Test finding by flag
	flag1Sections, err := store.Find(ctx, query.HasFlag("flag1"))
	require.NoError(t, err)
	assert.Len(t, flag1Sections, 2)

	// Test finding by command
	cmd1Sections, err := store.Find(ctx, query.HasCommand("cmd1"))
	require.NoError(t, err)
	assert.Len(t, cmd1Sections, 2)

	// Test finding top level
	topLevelSections, err := store.Find(ctx, query.IsTopLevel())
	require.NoError(t, err)
	assert.Len(t, topLevelSections, 2)

	// Test finding shown by default
	defaultSections, err := store.Find(ctx, query.ShownByDefault())
	require.NoError(t, err)
	assert.Len(t, defaultSections, 2)

	// Test complex query
	complexSections, err := store.Find(ctx, query.And(
		query.IsType(model.SectionExample),
		query.HasTopic("common"),
		query.IsTopLevel(),
	))
	require.NoError(t, err)
	assert.Len(t, complexSections, 1)
	assert.Equal(t, "example-1", complexSections[0].Slug)

	// Test OR query
	orSections, err := store.Find(ctx, query.Or(
		query.IsType(model.SectionExample),
		query.IsType(model.SectionTutorial),
	))
	require.NoError(t, err)
	assert.Len(t, orSections, 3)

	// Test NOT query
	notSections, err := store.Find(ctx, query.Not(query.IsType(model.SectionExample)))
	require.NoError(t, err)
	assert.Len(t, notSections, 1)
	assert.Equal(t, model.SectionTutorial, notSections[0].SectionType)
}

func TestStore_Stats(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test sections
	sections := []*model.Section{
		{
			Slug:        "example-1",
			Title:       "Example 1",
			SectionType: model.SectionExample,
			IsTopLevel:  true,
			ShowDefault: true,
		},
		{
			Slug:        "example-2",
			Title:       "Example 2",
			SectionType: model.SectionExample,
			IsTopLevel:  false,
			ShowDefault: false,
		},
		{
			Slug:        "tutorial-1",
			Title:       "Tutorial 1",
			SectionType: model.SectionTutorial,
			IsTopLevel:  true,
			ShowDefault: true,
		},
	}

	// Insert test sections
	for _, section := range sections {
		err = store.Upsert(ctx, section)
		require.NoError(t, err)
	}

	// Test stats
	stats, err := store.Stats(ctx)
	require.NoError(t, err)

	assert.Equal(t, 3, stats["total_sections"])
	assert.Equal(t, 2, stats["type_Example"])
	assert.Equal(t, 1, stats["type_Tutorial"])
	assert.Equal(t, 2, stats["top_level"])
	assert.Equal(t, 2, stats["show_default"])
}

func TestStore_TextSearch(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test sections with different content
	sections := []*model.Section{
		{
			Slug:    "search-1",
			Title:   "Search Test 1",
			Content: "This section contains information about databases and SQL queries",
		},
		{
			Slug:    "search-2",
			Title:   "Search Test 2",
			Content: "This section talks about web development and HTML",
		},
		{
			Slug:    "search-3",
			Title:   "Database Guide",
			Content: "A comprehensive guide to working with databases",
		},
	}

	// Insert test sections
	for _, section := range sections {
		err = store.Upsert(ctx, section)
		require.NoError(t, err)
	}

	// Rebuild FTS index to ensure it's up to date
	err = store.RebuildFTS(ctx)
	require.NoError(t, err)

	// Test text search - use wildcard to match both "database" and "databases"
	results, err := store.Find(ctx, query.TextSearch("database*"))
	require.NoError(t, err)
	assert.Len(t, results, 2)

	results, err = store.Find(ctx, query.TextSearch("web"))
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-2", results[0].Slug)

	results, err = store.Find(ctx, query.TextSearch("nonexistent"))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}
