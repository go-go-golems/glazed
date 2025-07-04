package query

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicPredicates(t *testing.T) {
	tests := []struct {
		name          string
		predicate     Predicate
		expectedSQL   string
		expectedArgs  []interface{}
	}{
		{
			name:      "IsType",
			predicate: IsType(model.SectionExample),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE s.sectionType = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"Example"},
		},
		{
			name:      "HasTopic",
			predicate: HasTopic("foo"),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE st.topic = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"foo"},
		},
		{
			name:      "HasFlag",
			predicate: HasFlag("verbose"),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_flags sf ON sf.section_id = s.id WHERE sf.flag = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"verbose"},
		},
		{
			name:      "HasCommand",
			predicate: HasCommand("list"),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_commands sc ON sc.section_id = s.id WHERE sc.command = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"list"},
		},
		{
			name:      "IsTopLevel",
			predicate: IsTopLevel(),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE s.isTopLevel = 1 ORDER BY s.ord",
			expectedArgs: []interface{}{},
		},
		{
			name:      "ShownByDefault",
			predicate: ShownByDefault(),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE s.showDefault = 1 ORDER BY s.ord",
			expectedArgs: []interface{}{},
		},
		{
			name:      "SlugEquals",
			predicate: SlugEquals("example-slug"),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE s.slug = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"example-slug"},
		},
		{
			name:      "TextSearch",
			predicate: TextSearch("search term"),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_fts fts ON fts.rowid = s.id WHERE section_fts MATCH ? ORDER BY s.ord",
			expectedArgs: []interface{}{"search term"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.predicate)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestBooleanCombinators(t *testing.T) {
	tests := []struct {
		name          string
		predicate     Predicate
		expectedSQL   string
		expectedArgs  []interface{}
	}{
		{
			name: "And with two predicates",
			predicate: And(
				IsType(model.SectionExample),
				HasTopic("foo"),
			),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE s.sectionType = ? AND st.topic = ? ORDER BY s.ord",
			expectedArgs: []interface{}{"Example", "foo"},
		},
		{
			name: "Or with two predicates",
			predicate: Or(
				IsType(model.SectionExample),
				IsType(model.SectionTutorial),
			),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE (s.sectionType = ? OR s.sectionType = ?) ORDER BY s.ord",
			expectedArgs: []interface{}{"Example", "Tutorial"},
		},
		{
			name: "Not with single predicate",
			predicate: Not(IsTopLevel()),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s WHERE NOT (s.isTopLevel = 1) ORDER BY s.ord",
			expectedArgs: []interface{}{},
		},
		{
			name: "Complex query with And/Or",
			predicate: And(
				Or(IsType(model.SectionExample), IsType(model.SectionTutorial)),
				HasTopic("foo"),
				IsTopLevel(),
			),
			expectedSQL: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE (s.sectionType = ? OR s.sectionType = ?) AND st.topic = ? AND s.isTopLevel = 1 ORDER BY s.ord",
			expectedArgs: []interface{}{"Example", "Tutorial", "foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.predicate)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestUniqueAliases(t *testing.T) {
	// Test that unique aliases are generated for multiple joins of the same table
	predicate := And(
		HasTopic("foo"),
		HasTopic("bar"),
	)
	
	sql, args := Compile(predicate)
	
	// Should have unique aliases st and st2
	assert.Contains(t, sql, "JOIN section_topics st ON st.section_id = s.id")
	assert.Contains(t, sql, "JOIN section_topics st2 ON st2.section_id = s.id")
	assert.Contains(t, sql, "st.topic = ?")
	assert.Contains(t, sql, "st2.topic = ?")
	
	require.Equal(t, 2, len(args))
	assert.Equal(t, "foo", args[0])
	assert.Equal(t, "bar", args[1])
}

func TestEmptyPredicates(t *testing.T) {
	tests := []struct {
		name      string
		predicate Predicate
	}{
		{
			name:      "Empty And",
			predicate: And(),
		},
		{
			name:      "Empty Or",
			predicate: Or(),
		},
		{
			name:      "Single predicate And",
			predicate: And(IsTopLevel()),
		},
		{
			name:      "Single predicate Or",
			predicate: Or(IsTopLevel()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.predicate)
			assert.NotEmpty(t, sql)
			assert.NotNil(t, args)
		})
	}
}
