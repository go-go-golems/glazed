package store

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadFromMarkdown(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)

	markdown := `---
Slug: test-section
Title: Test Section
SubTitle: A test section
Short: This is a test
SectionType: Example
Topics: [topic1, topic2]
Flags: [--flag1, --flag2]
Commands: [cmd1, cmd2]
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
Order: 1
---

# Test Section

This is the content of the test section.

## Example

Some example content here.
`

	section, err := loader.LoadFromMarkdown([]byte(markdown))
	require.NoError(t, err)

	assert.Equal(t, "test-section", section.Slug)
	assert.Equal(t, "Test Section", section.Title)
	assert.Equal(t, "A test section", section.SubTitle)
	assert.Equal(t, "This is a test", section.Short)
	assert.Equal(t, model.SectionExample, section.SectionType)
	assert.Equal(t, []string{"topic1", "topic2"}, section.Topics)
	assert.Equal(t, []string{"--flag1", "--flag2"}, section.Flags)
	assert.Equal(t, []string{"cmd1", "cmd2"}, section.Commands)
	assert.True(t, section.IsTopLevel)
	assert.False(t, section.IsTemplate)
	assert.True(t, section.ShowPerDefault)
	assert.Equal(t, 1, section.Order)
	assert.Contains(t, section.Content, "# Test Section")
	assert.Contains(t, section.Content, "This is the content")
}

func TestLoader_LoadFromMarkdownMinimal(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)

	markdown := `---
Slug: minimal-section
Title: Minimal Section
---

Just some content.
`

	section, err := loader.LoadFromMarkdown([]byte(markdown))
	require.NoError(t, err)

	assert.Equal(t, "minimal-section", section.Slug)
	assert.Equal(t, "Minimal Section", section.Title)
	assert.Equal(t, "", section.SubTitle)
	assert.Equal(t, "", section.Short)
	assert.Equal(t, model.SectionGeneralTopic, section.SectionType) // default
	assert.Equal(t, []string{}, section.Topics)
	assert.Equal(t, []string{}, section.Flags)
	assert.Equal(t, []string{}, section.Commands)
	assert.False(t, section.IsTopLevel)     // default
	assert.False(t, section.IsTemplate)     // default
	assert.False(t, section.ShowPerDefault) // default
	assert.Equal(t, 0, section.Order)       // default
	assert.Contains(t, section.Content, "Just some content")
}

func TestLoader_LoadFromMarkdownValidation(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)

	// Missing slug
	markdown := `---
Title: Test Section
---

Content here.
`

	_, err = loader.LoadFromMarkdown([]byte(markdown))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")

	// Missing title
	markdown = `---
Slug: test-section
---

Content here.
`

	_, err = loader.LoadFromMarkdown([]byte(markdown))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")

	// Invalid section type
	markdown = `---
Slug: test-section
Title: Test Section
SectionType: InvalidType
---

Content here.
`

	_, err = loader.LoadFromMarkdown([]byte(markdown))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section type")
}

func TestLoader_LoadFromFS(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)
	ctx := context.Background()

	// Create a test filesystem
	testFS := fstest.MapFS{
		"help/example1.md": &fstest.MapFile{
			Data: []byte(`---
Slug: example-1
Title: Example 1
SectionType: Example
Topics: [topic1]
---

Example 1 content.
`),
		},
		"help/tutorial1.md": &fstest.MapFile{
			Data: []byte(`---
Slug: tutorial-1
Title: Tutorial 1
SectionType: Tutorial
Topics: [topic1]
---

Tutorial 1 content.
`),
		},
		"help/subdir/example2.md": &fstest.MapFile{
			Data: []byte(`---
Slug: example-2
Title: Example 2
SectionType: Example
Topics: [topic2]
---

Example 2 content.
`),
		},
		"help/README.md": &fstest.MapFile{
			Data: []byte("This is a README file that should be ignored."),
		},
		"help/notmarkdown.txt": &fstest.MapFile{
			Data: []byte("This is not a markdown file."),
		},
	}

	err = loader.LoadFromFS(ctx, testFS, "help")
	require.NoError(t, err)

	// Check that sections were loaded
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count) // Should have loaded 3 sections, ignoring README and txt file

	// Check specific sections
	section, err := store.GetBySlug(ctx, "example-1")
	require.NoError(t, err)
	assert.Equal(t, "Example 1", section.Title)
	assert.Equal(t, model.SectionExample, section.SectionType)

	section, err = store.GetBySlug(ctx, "tutorial-1")
	require.NoError(t, err)
	assert.Equal(t, "Tutorial 1", section.Title)
	assert.Equal(t, model.SectionTutorial, section.SectionType)

	section, err = store.GetBySlug(ctx, "example-2")
	require.NoError(t, err)
	assert.Equal(t, "Example 2", section.Title)
	assert.Equal(t, []string{"topic2"}, section.Topics)
}

func TestLoader_SyncFromFS(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)
	ctx := context.Background()

	// First, add some existing data
	existingSection := &model.Section{
		Slug:        "existing-section",
		Title:       "Existing Section",
		SectionType: model.SectionGeneralTopic,
		Content:     "Existing content",
	}
	err = store.Insert(ctx, existingSection)
	require.NoError(t, err)

	// Verify existing data
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Create test filesystem
	testFS := fstest.MapFS{
		"help/new-section.md": &fstest.MapFile{
			Data: []byte(`---
Slug: new-section
Title: New Section
SectionType: Example
---

New section content.
`),
		},
	}

	// Sync from filesystem (should clear existing data)
	err = loader.SyncFromFS(ctx, testFS, "help")
	require.NoError(t, err)

	// Check that only new data exists
	count, err = store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Check that old section is gone
	_, err = store.GetBySlug(ctx, "existing-section")
	assert.Error(t, err)

	// Check that new section exists
	section, err := store.GetBySlug(ctx, "new-section")
	require.NoError(t, err)
	assert.Equal(t, "New Section", section.Title)
}

func TestLoader_LoadSections(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)
	ctx := context.Background()

	markdownFiles := map[string][]byte{
		"section1.md": []byte(`---
Slug: section-1
Title: Section 1
SectionType: Example
---

Section 1 content.
`),
		"section2.md": []byte(`---
Slug: section-2
Title: Section 2
SectionType: Tutorial
---

Section 2 content.
`),
		"invalid.md": []byte(`---
Title: Invalid Section
# Missing slug
---

Invalid content.
`),
	}

	err = loader.LoadSections(ctx, markdownFiles)
	require.NoError(t, err)

	// Check that valid sections were loaded
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count) // Should have loaded 2 valid sections

	// Check specific sections
	section, err := store.GetBySlug(ctx, "section-1")
	require.NoError(t, err)
	assert.Equal(t, "Section 1", section.Title)

	section, err = store.GetBySlug(ctx, "section-2")
	require.NoError(t, err)
	assert.Equal(t, "Section 2", section.Title)
}

func TestLoader_BatchUpsert(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)
	ctx := context.Background()

	sections := []*model.Section{
		{
			Slug:        "batch-1",
			Title:       "Batch 1",
			SectionType: model.SectionExample,
			Content:     "Batch 1 content",
		},
		{
			Slug:        "batch-2",
			Title:       "Batch 2",
			SectionType: model.SectionTutorial,
			Content:     "Batch 2 content",
		},
	}

	err = loader.BatchUpsert(ctx, sections)
	require.NoError(t, err)

	// Check that sections were loaded
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Check specific sections
	section, err := store.GetBySlug(ctx, "batch-1")
	require.NoError(t, err)
	assert.Equal(t, "Batch 1", section.Title)

	section, err = store.GetBySlug(ctx, "batch-2")
	require.NoError(t, err)
	assert.Equal(t, "Batch 2", section.Title)
}

func TestLoader_GetSectionStats(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)
	ctx := context.Background()

	// Add test sections
	sections := []*model.Section{
		{
			Slug:           "example-1",
			Title:          "Example 1",
			SectionType:    model.SectionExample,
			IsTopLevel:     true,
			ShowPerDefault: true,
		},
		{
			Slug:           "example-2",
			Title:          "Example 2",
			SectionType:    model.SectionExample,
			IsTopLevel:     false,
			ShowPerDefault: false,
		},
		{
			Slug:           "tutorial-1",
			Title:          "Tutorial 1",
			SectionType:    model.SectionTutorial,
			IsTopLevel:     true,
			ShowPerDefault: true,
		},
		{
			Slug:           "general-1",
			Title:          "General 1",
			SectionType:    model.SectionGeneralTopic,
			IsTopLevel:     false,
			ShowPerDefault: false,
		},
	}

	for _, section := range sections {
		err = store.Insert(ctx, section)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := loader.GetSectionStats(ctx)
	require.NoError(t, err)

	// Check stats
	assert.Equal(t, int64(4), stats["total"])
	assert.Equal(t, int64(2), stats["Example"])
	assert.Equal(t, int64(1), stats["Tutorial"])
	assert.Equal(t, int64(1), stats["GeneralTopic"])
	assert.Equal(t, int64(0), stats["Application"])
	assert.Equal(t, int64(2), stats["top_level"])
	assert.Equal(t, int64(2), stats["shown_by_default"])
}

func TestLoader_OrderConversion(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	loader := NewLoader(store)

	// Test with integer order
	markdown := `---
Slug: int-order
Title: Int Order
Order: 42
---

Content.
`

	section, err := loader.LoadFromMarkdown([]byte(markdown))
	require.NoError(t, err)
	assert.Equal(t, 42, section.Order)

	// Test with float order
	markdown = `---
Slug: float-order
Title: Float Order
Order: 42.5
---

Content.
`

	section, err = loader.LoadFromMarkdown([]byte(markdown))
	require.NoError(t, err)
	assert.Equal(t, 42, section.Order) // Should be truncated to int
}
