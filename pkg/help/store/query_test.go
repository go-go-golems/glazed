package store

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) *Store {
	store, err := NewInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()

	// Create test sections
	sections := []*model.Section{
		{
			Slug:           "example-1",
			Title:          "Example 1",
			SectionType:    model.SectionExample,
			Content:        "This is example 1 content",
			Topics:         []string{"topic1", "topic2"},
			Flags:          []string{"--flag1"},
			Commands:       []string{"cmd1"},
			IsTopLevel:     true,
			ShowPerDefault: true,
			Order:          1,
		},
		{
			Slug:           "example-2",
			Title:          "Example 2",
			SectionType:    model.SectionExample,
			Content:        "This is example 2 content",
			Topics:         []string{"topic2", "topic3"},
			Flags:          []string{"--flag2"},
			Commands:       []string{"cmd2"},
			IsTopLevel:     false,
			ShowPerDefault: false,
			Order:          2,
		},
		{
			Slug:           "tutorial-1",
			Title:          "Tutorial 1",
			SectionType:    model.SectionTutorial,
			Content:        "This is tutorial 1 content",
			Topics:         []string{"topic1"},
			Flags:          []string{"--flag1"},
			Commands:       []string{"cmd1"},
			IsTopLevel:     true,
			ShowPerDefault: true,
			Order:          3,
		},
		{
			Slug:           "general-1",
			Title:          "General Topic 1",
			SectionType:    model.SectionGeneralTopic,
			Content:        "This is general topic 1 content",
			Topics:         []string{"topic3"},
			Flags:          []string{"--flag3"},
			Commands:       []string{"cmd3"},
			IsTopLevel:     false,
			ShowPerDefault: false,
			Order:          4,
		},
		{
			Slug:           "app-1",
			Title:          "Application 1",
			SectionType:    model.SectionApplication,
			Content:        "This is application 1 content",
			Topics:         []string{"topic1", "topic3"},
			Flags:          []string{"--flag1", "--flag3"},
			Commands:       []string{"cmd1", "cmd3"},
			IsTopLevel:     true,
			ShowPerDefault: false,
			Order:          5,
		},
	}

	for _, section := range sections {
		err := store.Insert(ctx, section)
		require.NoError(t, err)
	}

	return store
}

func TestQuery_IsType(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test IsExample
	results, err := store.Find(ctx, IsExample())
	require.NoError(t, err)
	assert.Len(t, results, 2)
	for _, result := range results {
		assert.Equal(t, model.SectionExample, result.SectionType)
	}

	// Test IsTutorial
	results, err = store.Find(ctx, IsTutorial())
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, model.SectionTutorial, results[0].SectionType)

	// Test IsGeneralTopic
	results, err = store.Find(ctx, IsGeneralTopic())
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, model.SectionGeneralTopic, results[0].SectionType)

	// Test IsApplication
	results, err = store.Find(ctx, IsApplication())
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, model.SectionApplication, results[0].SectionType)
}

func TestQuery_HasTopic(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test HasTopic with topic1
	results, err := store.Find(ctx, HasTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, tutorial-1, app-1

	// Test HasTopic with topic2
	results, err = store.Find(ctx, HasTopic("topic2"))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, example-2

	// Test HasTopic with non-existent topic
	results, err = store.Find(ctx, HasTopic("nonexistent"))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestQuery_HasFlag(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test HasFlag with --flag1
	results, err := store.Find(ctx, HasFlag("--flag1"))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, tutorial-1, app-1

	// Test HasFlag with --flag2
	results, err = store.Find(ctx, HasFlag("--flag2"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-2

	// Test HasFlag with non-existent flag
	results, err = store.Find(ctx, HasFlag("--nonexistent"))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestQuery_HasCommand(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test HasCommand with cmd1
	results, err := store.Find(ctx, HasCommand("cmd1"))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, tutorial-1, app-1

	// Test HasCommand with cmd2
	results, err = store.Find(ctx, HasCommand("cmd2"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-2

	// Test HasCommand with non-existent command
	results, err = store.Find(ctx, HasCommand("nonexistent"))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestQuery_IsTopLevel(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, IsTopLevel())
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, tutorial-1, app-1

	for _, result := range results {
		assert.True(t, result.IsTopLevel)
	}
}

func TestQuery_ShownByDefault(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, ShownByDefault())
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, tutorial-1

	for _, result := range results {
		assert.True(t, result.ShowPerDefault)
	}
}

func TestQuery_NotShownByDefault(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, NotShownByDefault())
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-2, general-1, app-1

	for _, result := range results {
		assert.False(t, result.ShowPerDefault)
	}
}

func TestQuery_SlugEquals(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, SlugEquals("example-1"))
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "example-1", results[0].Slug)

	// Test non-existent slug
	results, err = store.Find(ctx, SlugEquals("nonexistent"))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestQuery_SlugIn(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, SlugIn([]string{"example-1", "tutorial-1"}))
	require.NoError(t, err)
	assert.Len(t, results, 2)

	slugs := make([]string, len(results))
	for i, result := range results {
		slugs[i] = result.Slug
	}
	assert.Contains(t, slugs, "example-1")
	assert.Contains(t, slugs, "tutorial-1")

	// Test empty slice
	results, err = store.Find(ctx, SlugIn([]string{}))
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestQuery_TitleContains(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, TitleContains("Example"))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, example-2

	results, err = store.Find(ctx, TitleContains("Tutorial"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // tutorial-1

	// Test case-insensitive search
	results, err = store.Find(ctx, TitleContains("example"))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, example-2
}

func TestQuery_ContentContains(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	results, err := store.Find(ctx, ContentContains("example 1"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-1

	results, err = store.Find(ctx, ContentContains("content"))
	require.NoError(t, err)
	assert.Len(t, results, 5) // all sections contain "content"
}

func TestQuery_And(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test AND combination: Examples AND TopLevel
	results, err := store.Find(ctx, And(IsExample(), IsTopLevel()))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-1

	// Test AND combination: HasTopic AND HasFlag
	results, err = store.Find(ctx, And(HasTopic("topic1"), HasFlag("--flag1")))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, tutorial-1, app-1

	// Test AND combination with multiple predicates
	results, err = store.Find(ctx, And(IsExample(), HasTopic("topic2"), ShownByDefault()))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-1
}

func TestQuery_Or(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test OR combination: Examples OR Tutorials
	results, err := store.Find(ctx, Or(IsExample(), IsTutorial()))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-1, example-2, tutorial-1

	// Test OR combination: HasTopic topic1 OR topic3
	results, err = store.Find(ctx, Or(HasTopic("topic1"), HasTopic("topic3")))
	require.NoError(t, err)
	assert.Len(t, results, 5) // example-1, example-2, tutorial-1, general-1, app-1

	// Test OR combination with single predicate
	results, err = store.Find(ctx, Or(IsExample()))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, example-2
}

func TestQuery_Not(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test NOT combination: NOT Examples
	results, err := store.Find(ctx, Not(IsExample()))
	require.NoError(t, err)
	assert.Len(t, results, 3) // tutorial-1, general-1, app-1

	// Test NOT combination: NOT TopLevel
	results, err = store.Find(ctx, Not(IsTopLevel()))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-2, general-1

	// Test NOT combination: NOT ShownByDefault
	results, err = store.Find(ctx, Not(ShownByDefault()))
	require.NoError(t, err)
	assert.Len(t, results, 3) // example-2, general-1, app-1
}

func TestQuery_ComplexCombinations(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test complex query: (Examples OR Tutorials) AND TopLevel
	results, err := store.Find(ctx, And(Or(IsExample(), IsTutorial()), IsTopLevel()))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, tutorial-1

	// Test complex query: HasTopic(topic1) AND NOT ShownByDefault
	results, err = store.Find(ctx, And(HasTopic("topic1"), Not(ShownByDefault())))
	require.NoError(t, err)
	assert.Len(t, results, 1) // app-1

	// Test complex query: (HasTopic(topic1) OR HasTopic(topic2)) AND IsExample
	results, err = store.Find(ctx, And(Or(HasTopic("topic1"), HasTopic("topic2")), IsExample()))
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, example-2
}

func TestQuery_Ordering(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test OrderByOrder
	results, err := store.Find(ctx, And(OrderByOrder()))
	require.NoError(t, err)
	assert.Len(t, results, 5)
	assert.Equal(t, "example-1", results[0].Slug)  // order 1
	assert.Equal(t, "example-2", results[1].Slug)  // order 2
	assert.Equal(t, "tutorial-1", results[2].Slug) // order 3
	assert.Equal(t, "general-1", results[3].Slug)  // order 4
	assert.Equal(t, "app-1", results[4].Slug)      // order 5

	// Test OrderByTitle
	results, err = store.Find(ctx, And(OrderByTitle()))
	require.NoError(t, err)
	assert.Len(t, results, 5)
	assert.Equal(t, "app-1", results[0].Slug)      // "Application 1"
	assert.Equal(t, "example-1", results[1].Slug)  // "Example 1"
	assert.Equal(t, "example-2", results[2].Slug)  // "Example 2"
	assert.Equal(t, "general-1", results[3].Slug)  // "General Topic 1"
	assert.Equal(t, "tutorial-1", results[4].Slug) // "Tutorial 1"
}

func TestQuery_LimitOffset(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test Limit
	results, err := store.Find(ctx, And(OrderByOrder(), Limit(2)))
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "example-1", results[0].Slug)
	assert.Equal(t, "example-2", results[1].Slug)

	// Test Offset
	results, err = store.Find(ctx, And(OrderByOrder(), Offset(2)))
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "tutorial-1", results[0].Slug)
	assert.Equal(t, "general-1", results[1].Slug)
	assert.Equal(t, "app-1", results[2].Slug)

	// Test Limit and Offset
	results, err = store.Find(ctx, And(OrderByOrder(), Limit(2), Offset(1)))
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "example-2", results[0].Slug)
	assert.Equal(t, "tutorial-1", results[1].Slug)
}

func TestQuery_ConveniencePredicates(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Test TopLevelDefaults
	results, err := store.Find(ctx, TopLevelDefaults())
	require.NoError(t, err)
	assert.Len(t, results, 2) // example-1, tutorial-1

	// Test ExamplesForTopic
	results, err = store.Find(ctx, ExamplesForTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-1

	// Test TutorialsForTopic
	results, err = store.Find(ctx, TutorialsForTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // tutorial-1

	// Test DefaultExamplesForTopic
	results, err = store.Find(ctx, DefaultExamplesForTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // example-1

	// Test DefaultTutorialsForTopic
	results, err = store.Find(ctx, DefaultTutorialsForTopic("topic1"))
	require.NoError(t, err)
	assert.Len(t, results, 1) // tutorial-1
}

func TestQueryCompiler_BuildQuery(t *testing.T) {
	compiler := NewQueryCompiler()

	// Test basic query
	compiler.AddWhere("slug = ?", "test")
	query, args := compiler.BuildQuery()
	assert.Contains(t, query, "WHERE slug = ?")
	assert.Equal(t, []interface{}{"test"}, args)

	// Test with multiple conditions
	compiler = NewQueryCompiler()
	compiler.AddWhere("slug = ?", "test")
	compiler.AddWhere("section_type = ?", 1)
	query, args = compiler.BuildQuery()
	assert.Contains(t, query, "WHERE slug = ? AND section_type = ?")
	assert.Equal(t, []interface{}{"test", 1}, args)

	// Test with joins
	compiler = NewQueryCompiler()
	compiler.AddJoin("JOIN sections_fts ON sections_fts.rowid = s.id")
	compiler.AddWhere("sections_fts MATCH ?", "search term")
	query, args = compiler.BuildQuery()
	assert.Contains(t, query, "JOIN sections_fts ON sections_fts.rowid = s.id")
	assert.Contains(t, query, "WHERE sections_fts MATCH ?")
	assert.Equal(t, []interface{}{"search term"}, args)

	// Test with order by
	compiler = NewQueryCompiler()
	compiler.SetOrderBy("s.order_num ASC")
	query, _ = compiler.BuildQuery()
	assert.Contains(t, query, "ORDER BY s.order_num ASC")

	// Test with limit
	compiler = NewQueryCompiler()
	compiler.SetLimit(10)
	query, _ = compiler.BuildQuery()
	assert.Contains(t, query, "LIMIT 10")

	// Test with offset
	compiler = NewQueryCompiler()
	compiler.SetOffset(5)
	query, _ = compiler.BuildQuery()
	assert.Contains(t, query, "OFFSET 5")
}
