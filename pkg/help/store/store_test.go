package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Upsert(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	section := &model.Section{
		Slug:        "test-section",
		Title:       "Test Section",
		Subtitle:    "A test section",
		Short:       "Short description",
		Content:     "This is test content",
		SectionType: model.SectionExample,
		IsTopLevel:  true,
		ShowDefault: true,
		Order:       1,
		Topics:      []string{"topic1", "topic2"},
		Flags:       []string{"--flag1", "--flag2"},
		Commands:    []string{"cmd1", "cmd2"},
	}

	err = store.Upsert(ctx, section)
	require.NoError(t, err)
	assert.NotZero(t, section.ID)
}

func TestStore_Find(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	// Insert test sections
	sections := []*model.Section{
		{
			Slug:        "example-1",
			Title:       "Example 1",
			SectionType: model.SectionExample,
			IsTopLevel:  true,
			Topics:      []string{"topic1"},
			Order:       1,
		},
		{
			Slug:        "tutorial-1",
			Title:       "Tutorial 1",
			SectionType: model.SectionTutorial,
			IsTopLevel:  false,
			Topics:      []string{"topic1", "topic2"},
			Order:       2,
		},
		{
			Slug:        "example-2",
			Title:       "Example 2",
			SectionType: model.SectionExample,
			IsTopLevel:  false,
			Topics:      []string{"topic2"},
			Order:       3,
		},
	}

	for _, section := range sections {
		err = store.Upsert(ctx, section)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		pred     query.Predicate
		wantLen  int
		wantSlug string
	}{
		{
			name:    "Find by type",
			pred:    query.IsType(model.SectionExample),
			wantLen: 2,
		},
		{
			name:    "Find by topic",
			pred:    query.HasTopic("topic1"),
			wantLen: 2,
		},
		{
			name:    "Find top level",
			pred:    query.IsTopLevel(),
			wantLen: 1,
		},
		{
			name:     "Find by slug",
			pred:     query.SlugEquals("example-1"),
			wantLen:  1,
			wantSlug: "example-1",
		},
		{
			name:    "Complex query",
			pred:    query.And(query.IsType(model.SectionExample), query.HasTopic("topic1")),
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.Find(ctx, tt.pred)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantLen)
			
			if tt.wantSlug != "" {
				assert.Equal(t, tt.wantSlug, results[0].Slug)
			}
		})
	}
}

func TestStore_GetBySlug(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemory()
	require.NoError(t, err)
	defer store.Close()

	section := &model.Section{
		Slug:        "test-section",
		Title:       "Test Section",
		SectionType: model.SectionExample,
		Topics:      []string{"topic1"},
		Flags:       []string{"--flag1"},
		Commands:    []string{"cmd1"},
	}

	err = store.Upsert(ctx, section)
	require.NoError(t, err)

	// Test successful retrieval
	result, err := store.GetBySlug(ctx, "test-section")
	require.NoError(t, err)
	assert.Equal(t, section.Slug, result.Slug)
	assert.Equal(t, section.Title, result.Title)
	assert.Equal(t, section.Topics, result.Topics)
	assert.Equal(t, section.Flags, result.Flags)
	assert.Equal(t, section.Commands, result.Commands)

	// Test not found
	_, err = store.GetBySlug(ctx, "nonexistent")
	assert.Error(t, err)
}
