package help

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSectionQueryWithStore(t *testing.T) {
	// Create an in-memory store
	st, err := store.NewInMemory()
	require.NoError(t, err)
	defer st.Close()

	ctx := context.Background()

	// Add some test sections
	sections := []*model.Section{
		{
			Slug:           "test-example",
			SectionType:    model.SectionExample,
			Title:          "Test Example",
			Content:        "This is a test example",
			Topics:         []string{"testing", "example"},
			ShowPerDefault: true,
		},
		{
			Slug:           "test-tutorial",
			SectionType:    model.SectionTutorial,
			Title:          "Test Tutorial",
			Content:        "This is a test tutorial",
			Topics:         []string{"testing", "tutorial"},
			ShowPerDefault: false,
		},
		{
			Slug:           "another-example",
			SectionType:    model.SectionExample,
			Title:          "Another Example",
			Content:        "This is another example",
			Topics:         []string{"demo"},
			ShowPerDefault: true,
		},
	}

	for _, section := range sections {
		err = st.Upsert(ctx, section)
		require.NoError(t, err)
	}

	t.Run("Query examples", func(t *testing.T) {
		query := NewSectionQuery().ReturnExamples()
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 2)
		assert.Equal(t, "test-example", results[0].Slug)
		assert.Equal(t, "another-example", results[1].Slug)
	})

	t.Run("Query tutorials", func(t *testing.T) {
		query := NewSectionQuery().ReturnTutorials()
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, "test-tutorial", results[0].Slug)
	})

	t.Run("Query by topic", func(t *testing.T) {
		query := NewSectionQuery().ReturnAllTypes().ReturnAnyOfTopics("testing")
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 2)
		// Results should contain both test-example and test-tutorial
		slugs := make([]string, len(results))
		for i, r := range results {
			slugs[i] = r.Slug
		}
		assert.Contains(t, slugs, "test-example")
		assert.Contains(t, slugs, "test-tutorial")
	})

	t.Run("Query shown by default", func(t *testing.T) {
		query := NewSectionQuery().ReturnAllTypes().ReturnOnlyShownByDefault()
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 2)
		// Results should contain both examples (shown by default)
		slugs := make([]string, len(results))
		for i, r := range results {
			slugs[i] = r.Slug
		}
		assert.Contains(t, slugs, "test-example")
		assert.Contains(t, slugs, "another-example")
	})

	t.Run("Query not shown by default", func(t *testing.T) {
		query := NewSectionQuery().ReturnAllTypes().ReturnOnlyNotShownByDefault()
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, "test-tutorial", results[0].Slug)
	})

	t.Run("Query by slug", func(t *testing.T) {
		query := NewSectionQuery().ReturnAllTypes().SearchForSlug("test-example")
		results, err := query.FindSections(ctx, st)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, "test-example", results[0].Slug)
	})
}
