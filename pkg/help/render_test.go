package help

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

func TestComputeRenderData_RelaxesNoQueryFallbackWhenPredicateIsNil(t *testing.T) {
	hs := NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:           "intro",
		Title:          "Intro",
		SectionType:    model.SectionGeneralTopic,
		ShowPerDefault: true,
		Order:          1,
	})
	hs.AddSection(&model.Section{
		Slug:        "example-install",
		Title:       "Install Example",
		SectionType: model.SectionExample,
		Order:       2,
	})

	data, noResultsFound := hs.ComputeRenderData(&RenderOptions{
		Predicate:             store.HasTopic("missing-topic"),
		RelaxNoQueryPredicate: nil,
		HasOnlyQueries:        true,
		QueryString:           "topics missing-topic",
	})

	if !noResultsFound {
		t.Fatalf("expected noResultsFound to be true")
	}

	helpPage, ok := data["Help"].(*HelpPage)
	if !ok {
		t.Fatalf("expected Help to be a *HelpPage, got %T", data["Help"])
	}

	if got := len(helpPage.AllGeneralTopics) + len(helpPage.AllExamples) + len(helpPage.AllApplications) + len(helpPage.AllTutorials); got != 2 {
		t.Fatalf("expected relaxed fallback to return all sections, got %d", got)
	}
}

func TestComputeRenderData_RelaxesNoTypesFallbackWhenPredicateIsNil(t *testing.T) {
	hs := NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:           "intro",
		Title:          "Intro",
		SectionType:    model.SectionGeneralTopic,
		ShowPerDefault: true,
		Order:          1,
	})
	hs.AddSection(&model.Section{
		Slug:        "example-install",
		Title:       "Install Example",
		SectionType: model.SectionExample,
		Order:       2,
	})

	data, noResultsFound := hs.ComputeRenderData(&RenderOptions{
		Predicate:             store.IsTutorial(),
		RelaxNoTypesPredicate: nil,
		HasRestrictedTypes:    true,
		RequestedTypes:        "tutorials",
	})

	if !noResultsFound {
		t.Fatalf("expected noResultsFound to be true")
	}

	helpPage, ok := data["Help"].(*HelpPage)
	if !ok {
		t.Fatalf("expected Help to be a *HelpPage, got %T", data["Help"])
	}

	if got := len(helpPage.AllGeneralTopics) + len(helpPage.AllExamples) + len(helpPage.AllApplications) + len(helpPage.AllTutorials); got != 2 {
		t.Fatalf("expected relaxed fallback to return all sections, got %d", got)
	}
}
