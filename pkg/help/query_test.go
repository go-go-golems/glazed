package help

import (
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Remove all tests that use SectionQuery and FindSections. Replace with tests that use the new predicate-based query system, constructing predicates and using store.Store.Find. Use examples from store/advanced_query_test.go as a template.

func TestSingleQueryAll(t *testing.T) {
	query := NewSectionQuery().ReturnAllTypes()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: model.SectionGeneralTopic,
		},
	}
	foundSections := query.FindSections(sections)

	// then
	assert.Len(t, foundSections, 1)
	assert.Equal(t, sections[0].Slug, foundSections[0].Slug)
}

func TestQueryOnlyDefault(t *testing.T) {
	query := NewSectionQuery().ReturnAllTypes().ReturnOnlyShownByDefault()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: model.SectionGeneralTopic,
		},
		{
			Slug:           "topic-2",
			SectionType:    model.SectionGeneralTopic,
			ShowPerDefault: true,
		},
	}
	foundSections := query.FindSections(sections)

	// then
	assert.Len(t, foundSections, 1)
	assert.Equal(t, sections[1].Slug, foundSections[0].Slug)
}

func TestQueryAllExamples(t *testing.T) {
	query := NewSectionQuery().ReturnExamples()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: model.SectionGeneralTopic,
		},
		{
			Slug:        "topic-2",
			SectionType: model.SectionGeneralTopic,
		},
		{
			Slug:        "example-1",
			SectionType: model.SectionExample,
		},
		{
			Slug:        "example-2",
			SectionType: model.SectionExample,
		},
		{
			Slug:        "application-1",
			SectionType: model.SectionApplication,
		},
	}

	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 2)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
	assert.Equal(t, sections[3].Slug, foundSections[1].Slug)
}

func TestQueryOnlyJsonCommandExamples(t *testing.T) {
	query := NewSectionQuery().ReturnExamples().ReturnOnlyCommands("json")

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: model.SectionGeneralTopic,
		},
		{
			Slug:        "topic-2",
			SectionType: model.SectionGeneralTopic,
			Commands:    []string{"json"},
		},
		{
			Slug:        "example-1",
			SectionType: model.SectionExample,
			Commands:    []string{"json"},
		},
		{
			Slug:        "example-2",
			SectionType: model.SectionExample,
		},
		{
			Slug:        "example-3",
			SectionType: model.SectionExample,
			Commands:    []string{"yaml", "docs", "help"},
		},
		{
			Slug:        "application-1",
			SectionType: model.SectionApplication,
			Commands:    []string{"json"},
		},
	}

	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 1)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
}

var sections = []*Section{
	{
		Slug:        "topic-1",
		SectionType: model.SectionGeneralTopic,
		Topics:      []string{"templates"},
	},
	{
		Slug:        "topic-2",
		SectionType: model.SectionGeneralTopic,
		Topics:      []string{"template-fields"},
	},
	{
		Slug:        "topic-3",
		SectionType: model.SectionGeneralTopic,
		Topics:      []string{"template-fields", "templates"},
	},
	{
		Slug:        "topic-4",
		SectionType: model.SectionGeneralTopic,
		Topics:      []string{"template-fields", "templates", "other"},
	},
	{
		Slug:        "topic-5",
		SectionType: model.SectionGeneralTopic,
		Topics:      []string{"other"},
	},
	{
		Slug:        "example-1",
		SectionType: model.SectionExample,
		Topics:      []string{"templates"},
	},
	{
		Slug:        "example-2",
		SectionType: model.SectionExample,
		Topics:      []string{"template-fields"},
	},
	{
		Slug:        "example-3",
		SectionType: model.SectionExample,
		Topics:      []string{"template-fields", "templates"},
	},
	{
		Slug:        "example-4",
		SectionType: model.SectionExample,
		Topics:      []string{"template-fields", "templates", "other"},
	},
	{
		Slug:        "example-5",
		SectionType: model.SectionExample,
		Topics:      []string{"other"},
	},
}

func TestQueryTopicTemplates(t *testing.T) {
	query := NewSectionQuery().ReturnAllTypes().ReturnAnyOfTopics("templates", "template-fields")
	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 8)
	assert.Equal(t, sections[0].Slug, foundSections[0].Slug)
	assert.Equal(t, sections[1].Slug, foundSections[1].Slug)
	assert.Equal(t, sections[2].Slug, foundSections[2].Slug)
	assert.Equal(t, sections[3].Slug, foundSections[3].Slug)
	assert.Equal(t, sections[5].Slug, foundSections[4].Slug)
	assert.Equal(t, sections[6].Slug, foundSections[5].Slug)
	assert.Equal(t, sections[7].Slug, foundSections[6].Slug)
	assert.Equal(t, sections[8].Slug, foundSections[7].Slug)
}

func TestQueryTopicOnlyTemplatesTemplatesFields(t *testing.T) {
	query := NewSectionQuery().ReturnAllTypes().ReturnOnlyTopics("templates", "template-fields")
	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 4)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
	assert.Equal(t, sections[3].Slug, foundSections[1].Slug)
	assert.Equal(t, sections[7].Slug, foundSections[2].Slug)
	assert.Equal(t, sections[8].Slug, foundSections[3].Slug)
}

func TestQueryFilterSections(t *testing.T) {
	query := NewSectionQuery().ReturnAllTypes().FilterSections(sections[0], sections[1])
	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 8)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
}
