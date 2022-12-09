package help

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleQueryAll(t *testing.T) {
	query := NewQueryBuilder().ReturnAllTypes().Build()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: SectionGeneralTopic,
		},
	}
	foundSections := query.FindSections(sections)

	// then
	assert.Len(t, foundSections, 1)
	assert.Equal(t, sections[0].Slug, foundSections[0].Slug)
}

func TestQueryOnlyDefault(t *testing.T) {
	query := NewQueryBuilder().ReturnAllTypes().OnlyDefault().Build()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: SectionGeneralTopic,
		},
		{
			Slug:           "topic-2",
			SectionType:    SectionGeneralTopic,
			ShowPerDefault: true,
		},
	}
	foundSections := query.FindSections(sections)

	// then
	assert.Len(t, foundSections, 1)
	assert.Equal(t, sections[1].Slug, foundSections[0].Slug)
}

func TestQueryAllExamples(t *testing.T) {
	query := NewQueryBuilder().ReturnExamples().Build()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: SectionGeneralTopic,
		},
		{
			Slug:        "topic-2",
			SectionType: SectionGeneralTopic,
		},
		{
			Slug:        "example-1",
			SectionType: SectionExample,
		},
		{
			Slug:        "example-2",
			SectionType: SectionExample,
		},
		{
			Slug:        "application-1",
			SectionType: SectionApplication,
		},
	}

	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 2)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
	assert.Equal(t, sections[3].Slug, foundSections[1].Slug)
}

func TestQueryOnlyJsonCommandExamples(t *testing.T) {
	query := NewQueryBuilder().ReturnExamples().OnlyCommands("json").Build()

	sections := []*Section{
		{
			Slug:        "topic-1",
			SectionType: SectionGeneralTopic,
		},
		{
			Slug:        "topic-2",
			SectionType: SectionGeneralTopic,
			Commands:    []string{"json"},
		},
		{
			Slug:        "example-1",
			SectionType: SectionExample,
			Commands:    []string{"json"},
		},
		{
			Slug:        "example-2",
			SectionType: SectionExample,
		},
		{
			Slug:        "example-3",
			SectionType: SectionExample,
			Commands:    []string{"yaml", "docs", "help"},
		},
		{
			Slug:        "application-1",
			SectionType: SectionApplication,
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
		SectionType: SectionGeneralTopic,
		Topics:      []string{"templates"},
	},
	{
		Slug:        "topic-2",
		SectionType: SectionGeneralTopic,
		Topics:      []string{"template-fields"},
	},
	{
		Slug:        "topic-3",
		SectionType: SectionGeneralTopic,
		Topics:      []string{"template-fields", "templates"},
	},
	{
		Slug:        "topic-4",
		SectionType: SectionGeneralTopic,
		Topics:      []string{"template-fields", "templates", "other"},
	},
	{
		Slug:        "topic-5",
		SectionType: SectionGeneralTopic,
		Topics:      []string{"other"},
	},
	{
		Slug:        "example-1",
		SectionType: SectionExample,
		Topics:      []string{"templates"},
	},
	{
		Slug:        "example-2",
		SectionType: SectionExample,
		Topics:      []string{"template-fields"},
	},
	{
		Slug:        "example-3",
		SectionType: SectionExample,
		Topics:      []string{"template-fields", "templates"},
	},
	{
		Slug:        "example-4",
		SectionType: SectionExample,
		Topics:      []string{"template-fields", "templates", "other"},
	},
	{
		Slug:        "example-5",
		SectionType: SectionExample,
		Topics:      []string{"other"},
	},
}

func TestQueryTopicTemplates(t *testing.T) {
	query := NewQueryBuilder().ReturnAllTypes().Topics("templates", "template-fields").Build()
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
	query := NewQueryBuilder().ReturnAllTypes().OnlyTopics("templates", "template-fields").Build()
	foundSections := query.FindSections(sections)
	assert.Len(t, foundSections, 4)
	assert.Equal(t, sections[2].Slug, foundSections[0].Slug)
	assert.Equal(t, sections[3].Slug, foundSections[1].Slug)
	assert.Equal(t, sections[7].Slug, foundSections[2].Slug)
	assert.Equal(t, sections[8].Slug, foundSections[3].Slug)
}
