package query

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
)

func TestSimplePredicates(t *testing.T) {
	tests := []struct {
		name     string
		pred     Predicate
		wantSQL  string
		wantArgs []interface{}
	}{
		{
			name: "IsType",
			pred: IsType(model.SectionExample),
			wantSQL: "SELECT DISTINCT s.id, s.slug, s.title, s.subtitle, s.short, s.content, s.sectionType, s.isTopLevel, s.isTemplate, s.showDefault, s.ord FROM sections s WHERE s.sectionType = ? ORDER BY s.ord",
			wantArgs: []interface{}{"Example"},
		},
		{
			name: "HasTopic",
			pred: HasTopic("foo"),
			wantSQL: "SELECT DISTINCT s.id, s.slug, s.title, s.subtitle, s.short, s.content, s.sectionType, s.isTopLevel, s.isTemplate, s.showDefault, s.ord FROM sections s JOIN section_topics st ON st.section_id = s.id WHERE st.topic = ? ORDER BY s.ord",
			wantArgs: []interface{}{"foo"},
		},
		{
			name: "IsTopLevel",
			pred: IsTopLevel(),
			wantSQL: "SELECT DISTINCT s.id, s.slug, s.title, s.subtitle, s.short, s.content, s.sectionType, s.isTopLevel, s.isTemplate, s.showDefault, s.ord FROM sections s WHERE s.isTopLevel = ? ORDER BY s.ord",
			wantArgs: []interface{}{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := Compile(tt.pred)
			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestAndPredicate(t *testing.T) {
	pred := And(
		IsType(model.SectionExample),
		HasTopic("foo"),
		IsTopLevel(),
	)
	
	sql, args := Compile(pred)
	
	// Check that all conditions are present
	assert.Contains(t, sql, "s.sectionType = ?")
	assert.Contains(t, sql, "st.topic = ?")
	assert.Contains(t, sql, "s.isTopLevel = ?")
	assert.Contains(t, sql, "JOIN section_topics st ON st.section_id = s.id")
	
	expectedArgs := []interface{}{"Example", "foo", true}
	assert.Equal(t, expectedArgs, args)
}

func TestOrPredicate(t *testing.T) {
	pred := Or(
		IsType(model.SectionExample),
		IsType(model.SectionTutorial),
	)
	
	sql, args := Compile(pred)
	
	assert.Contains(t, sql, "s.sectionType = ?")
	assert.Contains(t, sql, "OR")
	
	expectedArgs := []interface{}{"Example", "Tutorial"}
	assert.Equal(t, expectedArgs, args)
}

func TestNotPredicate(t *testing.T) {
	pred := Not(IsType(model.SectionExample))
	
	sql, args := Compile(pred)
	
	assert.Contains(t, sql, "NOT")
	assert.Contains(t, sql, "s.sectionType = ?")
	
	expectedArgs := []interface{}{"Example"}
	assert.Equal(t, expectedArgs, args)
}

func TestComplexQuery(t *testing.T) {
	pred := And(
		Or(IsType(model.SectionExample), IsType(model.SectionTutorial)),
		HasTopic("foo"),
		IsTopLevel(),
	)
	
	sql, args := Compile(pred)
	
	// Should have the join for topics
	assert.Contains(t, sql, "JOIN section_topics st ON st.section_id = s.id")
	
	// Should have all the conditions
	assert.Contains(t, sql, "s.sectionType = ?")
	assert.Contains(t, sql, "st.topic = ?")
	assert.Contains(t, sql, "s.isTopLevel = ?")
	assert.Contains(t, sql, "OR")
	assert.Contains(t, sql, "AND")
	
	expectedArgs := []interface{}{"Example", "Tutorial", "foo", true}
	assert.Equal(t, expectedArgs, args)
}
