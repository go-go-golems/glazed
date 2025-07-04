package help

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore creates a test store with sample sections
func setupTestStore(t *testing.T) *store.Store {
	st, err := store.NewInMemory()
	require.NoError(t, err)

	ctx := context.Background()

	sections := []*model.Section{
		{
			Slug:        "topic-1",
			SectionType: SectionGeneralTopic,
			Title:       "Topic 1",
			Content:     "Content 1",
			Topics:      []string{"templates"},
		},
		{
			Slug:           "topic-2",
			SectionType:    SectionGeneralTopic,
			Title:          "Topic 2",
			Content:        "Content 2",
			Topics:         []string{"template-fields"},
			ShowPerDefault: true,
		},
		{
			Slug:        "topic-3",
			SectionType: SectionGeneralTopic,
			Title:       "Topic 3",
			Content:     "Content 3",
			Topics:      []string{"template-fields", "templates"},
		},
		{
			Slug:        "example-1",
			SectionType: SectionExample,
			Title:       "Example 1",
			Content:     "Example content 1",
			Topics:      []string{"templates"},
		},
		{
			Slug:        "example-2",
			SectionType: SectionExample,
			Title:       "Example 2",
			Content:     "Example content 2",
			Commands:    []string{"json"},
		},
		{
			Slug:        "example-3",
			SectionType: SectionExample,
			Title:       "Example 3",
			Content:     "Example content 3",
			Commands:    []string{"yaml", "docs", "help"},
		},
	}

	for _, section := range sections {
		err = st.Upsert(ctx, section)
		require.NoError(t, err)
	}

	return st
}

func TestQueryAllTypes(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnAllTypes()
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 6) // 3 topics + 3 examples
}

func TestQueryOnlyDefault(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnAllTypes().ReturnOnlyShownByDefault()
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, "topic-2", results[0].Slug)
}

func TestQueryExamples(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnExamples()
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 3)
	slugs := make([]string, len(results))
	for i, r := range results {
		slugs[i] = r.Slug
	}
	assert.Contains(t, slugs, "example-1")
	assert.Contains(t, slugs, "example-2")
	assert.Contains(t, slugs, "example-3")
}

func TestQueryByCommand(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnExamples().ReturnOnlyCommands("json")
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, "example-2", results[0].Slug)
}

func TestQueryByTopic(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnAllTypes().ReturnAnyOfTopics("templates")
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 3) // topic-1, topic-3, example-1
	slugs := make([]string, len(results))
	for i, r := range results {
		slugs[i] = r.Slug
	}
	assert.Contains(t, slugs, "topic-1")
	assert.Contains(t, slugs, "topic-3")
	assert.Contains(t, slugs, "example-1")
}

func TestQueryOnlyTopics(t *testing.T) {
	st := setupTestStore(t)
	defer func() { _ = st.Close() }()

	ctx := context.Background()

	query := NewSectionQuery().ReturnAllTypes().ReturnOnlyTopics("templates", "template-fields")
	results, err := query.FindSections(ctx, st)
	require.NoError(t, err)

	assert.Len(t, results, 1) // Only topic-3 has both topics
	assert.Equal(t, "topic-3", results[0].Slug)
}
