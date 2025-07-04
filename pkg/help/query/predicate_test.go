package query

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
)

func TestBasicPredicates(t *testing.T) {
	tests := []struct {
		name     string
		pred     Predicate
		expected string
		args     []any
	}{
		{
			name:     "IsType",
			pred:     IsType(model.SectionExample),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE s.sectionType = ? ORDER BY s.ord",
			args:     []any{"Example"},
		},
		{
			name:     "HasTopic",
			pred:     HasTopic("templates"),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE st.topic = ? ORDER BY s.ord",
			args:     []any{"templates"},
		},
		{
			name:     "HasFlag",
			pred:     HasFlag("verbose"),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_flags sf ON sf.section_id = s.id WHERE sf.flag = ? ORDER BY s.ord",
			args:     []any{"verbose"},
		},
		{
			name:     "HasCommand",
			pred:     HasCommand("help"),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_commands sc ON sc.section_id = s.id WHERE sc.command = ? ORDER BY s.ord",
			args:     []any{"help"},
		},
		{
			name:     "IsTopLevel",
			pred:     IsTopLevel(),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE s.isTopLevel = ? ORDER BY s.ord",
			args:     []any{true},
		},
		{
			name:     "ShownByDefault",
			pred:     ShownByDefault(),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE s.showDefault = ? ORDER BY s.ord",
			args:     []any{true},
		},
		{
			name:     "SlugEquals",
			pred:     SlugEquals("test-slug"),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE s.slug = ? ORDER BY s.ord",
			args:     []any{"test-slug"},
		},
		{
			name:     "TextSearch",
			pred:     TextSearch("search term"),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_fts fts ON fts.rowid = s.id WHERE section_fts MATCH ? ORDER BY s.ord",
			args:     []any{"search term"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.pred)
			assert.Equal(t, tt.expected, sql)
			assert.Equal(t, tt.args, args)
		})
	}
}

func TestBooleanCombiners(t *testing.T) {
	tests := []struct {
		name     string
		pred     Predicate
		expected string
		args     []any
	}{
		{
			name: "And",
			pred: And(
				IsType(model.SectionExample),
				HasTopic("templates"),
			),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE (s.sectionType = ? AND st.topic = ?) ORDER BY s.ord",
			args:     []any{"Example", "templates"},
		},
		{
			name: "Or",
			pred: Or(
				IsType(model.SectionExample),
				IsType(model.SectionTutorial),
			),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE (s.sectionType = ? OR s.sectionType = ?) ORDER BY s.ord",
			args:     []any{"Example", "Tutorial"},
		},
		{
			name: "Not",
			pred: Not(IsType(model.SectionExample)),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE NOT (s.sectionType = ?) ORDER BY s.ord",
			args:     []any{"Example"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.pred)
			assert.Equal(t, tt.expected, sql)
			assert.Equal(t, tt.args, args)
		})
	}
}

func TestComplexQueries(t *testing.T) {
	tests := []struct {
		name     string
		pred     Predicate
		expected string
		args     []any
	}{
		{
			name: "AndOr",
			pred: And(
				Or(IsType(model.SectionExample), IsType(model.SectionTutorial)),
				HasTopic("templates"),
				IsTopLevel(),
			),
			expected: "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE ((s.sectionType = ? OR s.sectionType = ?) AND st.topic = ? AND s.isTopLevel = ?) ORDER BY s.ord",
			args:     []any{"Example", "Tutorial", "templates", true},
		},
		{
			name: "NotOr",
			pred: Not(Or(IsType(model.SectionExample), IsType(model.SectionTutorial))),
			expected: "SELECT DISTINCT s.* FROM sections s WHERE NOT ((s.sectionType = ? OR s.sectionType = ?)) ORDER BY s.ord",
			args:     []any{"Example", "Tutorial"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.pred)
			assert.Equal(t, tt.expected, sql)
			assert.Equal(t, tt.args, args)
		})
	}
}

func TestEmptyPredicates(t *testing.T) {
	tests := []struct {
		name     string
		pred     Predicate
		expected string
	}{
		{
			name:     "EmptyAnd",
			pred:     And(),
			expected: "SELECT DISTINCT s.* FROM sections s ORDER BY s.ord",
		},
		{
			name:     "EmptyOr",
			pred:     Or(),
			expected: "SELECT DISTINCT s.* FROM sections s ORDER BY s.ord",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.pred)
			assert.Equal(t, tt.expected, sql)
			assert.Empty(t, args)
		})
	}
}

func TestJoinDeduplication(t *testing.T) {
	// Test that duplicate JOINs are properly deduplicated
	pred := And(
		HasTopic("topic1"),
		HasTopic("topic2"),
	)

	sql, args := Compile(pred)
	
	// Should only have one JOIN for section_topics
	expected := "SELECT DISTINCT s.* FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE (st.topic = ? AND st.topic = ?) ORDER BY s.ord"
	assert.Equal(t, expected, sql)
	assert.Equal(t, []any{"topic1", "topic2"}, args)
}
