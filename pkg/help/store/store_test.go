package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_BasicOperations(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	ctx := context.Background()

	// Test Insert
	section := &model.Section{
		Slug:        "test-section",
		Title:       "Test Section",
		SectionType: model.SectionExample,
		Content:     "This is a test section",
		Topics:      []string{"test", "example"},
		Flags:       []string{"--verbose"},
		Commands:    []string{"test"},
		IsTopLevel:  true,
		Order:       1,
	}

	err = store.Insert(ctx, section)
	require.NoError(t, err)
	assert.NotZero(t, section.ID)

	// Test GetBySlug
	retrieved, err := store.GetBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, section.Slug, retrieved.Slug)
	assert.Equal(t, section.Title, retrieved.Title)
	assert.Equal(t, section.SectionType, retrieved.SectionType)
	assert.Equal(t, section.Topics, retrieved.Topics)
	assert.Equal(t, section.Flags, retrieved.Flags)
	assert.Equal(t, section.Commands, retrieved.Commands)
	assert.Equal(t, section.IsTopLevel, retrieved.IsTopLevel)

	// Test GetByID
	retrieved2, err := store.GetByID(ctx, section.ID)
	require.NoError(t, err)
	assert.Equal(t, section.ID, retrieved2.ID)

	// Test Update
	section.Title = "Updated Test Section"
	section.Content = "Updated content"
	err = store.Update(ctx, section)
	require.NoError(t, err)

	updated, err := store.GetBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Section", updated.Title)
	assert.Equal(t, "Updated content", updated.Content)

	// Test Count
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test Delete
	err = store.Delete(ctx, section.ID)
	require.NoError(t, err)

	_, err = store.GetBySlug(ctx, "test-section")
	assert.Error(t, err)

	// Test count after delete
	count, err = store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestStore_Upsert(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	ctx := context.Background()

	section := &model.Section{
		Slug:        "upsert-test",
		Title:       "Upsert Test",
		SectionType: model.SectionGeneralTopic,
		Content:     "Original content",
	}

	// First upsert (insert)
	err = store.Upsert(ctx, section)
	require.NoError(t, err)
	assert.NotZero(t, section.ID)

	// Second upsert (update)
	section.Title = "Updated Upsert Test"
	section.Content = "Updated content"
	err = store.Upsert(ctx, section)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.GetBySlug(ctx, "upsert-test")
	require.NoError(t, err)
	assert.Equal(t, "Updated Upsert Test", retrieved.Title)
	assert.Equal(t, "Updated content", retrieved.Content)

	// Verify only one record exists
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestStore_List(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	ctx := context.Background()

	// Insert multiple sections
	sections := []*model.Section{
		{
			Slug:        "section-1",
			Title:       "Section 1",
			SectionType: model.SectionExample,
			Order:       2,
		},
		{
			Slug:        "section-2",
			Title:       "Section 2",
			SectionType: model.SectionTutorial,
			Order:       1,
		},
		{
			Slug:        "section-3",
			Title:       "Section 3",
			SectionType: model.SectionGeneralTopic,
			Order:       3,
		},
	}

	for _, section := range sections {
		err = store.Insert(ctx, section)
		require.NoError(t, err)
	}

	// Test list without ordering
	result, err := store.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Test list with ordering
	result, err = store.List(ctx, "order_num ASC")
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "section-2", result[0].Slug) // order 1
	assert.Equal(t, "section-1", result[1].Slug) // order 2
	assert.Equal(t, "section-3", result[2].Slug) // order 3
}

func TestStore_Clear(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	ctx := context.Background()

	// Insert a section
	section := &model.Section{
		Slug:        "clear-test",
		Title:       "Clear Test",
		SectionType: model.SectionExample,
	}

	err = store.Insert(ctx, section)
	require.NoError(t, err)

	// Verify it exists
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Clear the store
	err = store.Clear(ctx)
	require.NoError(t, err)

	// Verify it's empty
	count, err = store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestStore_StringFields(t *testing.T) {
	store, err := NewInMemory()
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	ctx := context.Background()

	section := &model.Section{
		Slug:        "string-test",
		Title:       "String Test",
		SectionType: model.SectionExample,
		Topics:      []string{"topic1", "topic2", "topic3"},
		Flags:       []string{"--flag1", "--flag2"},
		Commands:    []string{"cmd1", "cmd2"},
	}

	err = store.Insert(ctx, section)
	require.NoError(t, err)

	retrieved, err := store.GetBySlug(ctx, "string-test")
	require.NoError(t, err)

	assert.Equal(t, []string{"topic1", "topic2", "topic3"}, retrieved.Topics)
	assert.Equal(t, []string{"--flag1", "--flag2"}, retrieved.Flags)
	assert.Equal(t, []string{"cmd1", "cmd2"}, retrieved.Commands)
}

func TestSection_StringConversion(t *testing.T) {
	section := &model.Section{
		Topics:   []string{"topic1", "topic2"},
		Flags:    []string{"--flag1", "--flag2"},
		Commands: []string{"cmd1", "cmd2"},
	}

	// Test conversion to string
	assert.Equal(t, "topic1,topic2", section.TopicsAsString())
	assert.Equal(t, "--flag1,--flag2", section.FlagsAsString())
	assert.Equal(t, "cmd1,cmd2", section.CommandsAsString())

	// Test conversion from string
	section.SetTopicsFromString("new1,new2,new3")
	assert.Equal(t, []string{"new1", "new2", "new3"}, section.Topics)

	section.SetFlagsFromString("--new1,--new2")
	assert.Equal(t, []string{"--new1", "--new2"}, section.Flags)

	section.SetCommandsFromString("newcmd1,newcmd2")
	assert.Equal(t, []string{"newcmd1", "newcmd2"}, section.Commands)

	// Test empty string
	section.SetTopicsFromString("")
	assert.Equal(t, []string{}, section.Topics)
}

func TestSection_Validation(t *testing.T) {
	// Valid section
	section := &model.Section{
		Slug:        "valid-section",
		Title:       "Valid Section",
		SectionType: model.SectionExample,
	}
	assert.NoError(t, section.Validate())

	// Missing slug
	section.Slug = ""
	assert.Error(t, section.Validate())

	// Missing title
	section.Slug = "valid-section"
	section.Title = ""
	assert.Error(t, section.Validate())
}

func TestSectionType_Conversion(t *testing.T) {
	// Test string to type conversion
	sectionType, err := model.SectionTypeFromString("Example")
	require.NoError(t, err)
	assert.Equal(t, model.SectionExample, sectionType)

	sectionType, err = model.SectionTypeFromString("Tutorial")
	require.NoError(t, err)
	assert.Equal(t, model.SectionTutorial, sectionType)

	// Test invalid type
	_, err = model.SectionTypeFromString("Invalid")
	assert.Error(t, err)

	// Test type to string conversion
	assert.Equal(t, "Example", model.SectionExample.String())
	assert.Equal(t, "Tutorial", model.SectionTutorial.String())
	assert.Equal(t, "GeneralTopic", model.SectionGeneralTopic.String())
	assert.Equal(t, "Application", model.SectionApplication.String())

	// Test type to int conversion
	assert.Equal(t, 1, model.SectionExample.ToInt())
	assert.Equal(t, 3, model.SectionTutorial.ToInt())
}
