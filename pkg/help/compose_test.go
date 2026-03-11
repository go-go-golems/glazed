package help

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
)

type testPagePart struct {
	kind     string
	order    int
	markdown string
}

func (p testPagePart) Kind() string { return p.kind }
func (p testPagePart) Order() int   { return p.order }
func (p testPagePart) RenderMarkdown(_ context.Context) (string, error) {
	return p.markdown, nil
}

type testComposer struct {
	pages map[string]*ComposedPage
}

func (c testComposer) ComposePage(_ context.Context, slug string, _ *HelpSystem) (*ComposedPage, error) {
	page, ok := c.pages[slug]
	if !ok {
		return nil, ErrPageNotComposed
	}
	return page, nil
}

func TestComposedPageRenderMarkdownOrdersParts(t *testing.T) {
	page := &ComposedPage{
		Slug: "test",
		Parts: []PagePart{
			testPagePart{kind: "second", order: 20, markdown: "Second"},
			testPagePart{kind: "first", order: 10, markdown: "First"},
		},
	}

	markdown, err := page.RenderMarkdown(context.Background())
	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v", err)
	}

	if markdown != "First\n\nSecond" {
		t.Fatalf("unexpected markdown: %q", markdown)
	}
}

func TestRenderTopicHelpUsesComposerWhenAvailable(t *testing.T) {
	hs := NewHelpSystem()
	hs.AddSection(&Section{
		Section: &model.Section{
			Slug:    "topic-1",
			Title:   "Topic 1",
			Content: "Legacy content",
		},
	})
	hs.SetPageComposer(testComposer{
		pages: map[string]*ComposedPage{
			"topic-1": {
				Slug: "topic-1",
				Parts: []PagePart{
					testPagePart{kind: "body", order: 10, markdown: "# Composed Topic\n\nComposed content"},
				},
			},
		},
	})

	section, err := hs.GetSectionWithSlug("topic-1")
	if err != nil {
		t.Fatalf("GetSectionWithSlug() error = %v", err)
	}

	var out bytes.Buffer
	rendered, err := hs.RenderTopicHelpWithWriter(section, &RenderOptions{
		Query: NewSectionQuery().ReturnAllTypes(),
	}, &out)
	if err != nil {
		t.Fatalf("RenderTopicHelpWithWriter() error = %v", err)
	}

	if !strings.Contains(rendered, "Composed Topic") {
		t.Fatalf("expected composed output, got %q", rendered)
	}
	if strings.Contains(rendered, "Legacy content") {
		t.Fatalf("expected composer to override legacy rendering, got %q", rendered)
	}
}

func TestRenderTopicHelpFallsBackWhenPageNotComposed(t *testing.T) {
	hs := NewHelpSystem()
	hs.AddSection(&Section{
		Section: &model.Section{
			Slug:           "topic-1",
			Title:          "Topic 1",
			Content:        "Legacy content",
			ShowPerDefault: true,
			SectionType:    model.SectionGeneralTopic,
		},
	})
	hs.SetPageComposer(testComposer{pages: map[string]*ComposedPage{}})

	section, err := hs.GetSectionWithSlug("topic-1")
	if err != nil {
		t.Fatalf("GetSectionWithSlug() error = %v", err)
	}

	var out bytes.Buffer
	rendered, err := hs.RenderTopicHelpWithWriter(section, &RenderOptions{
		Query:           NewSectionQuery().ReturnAllTypes(),
		ShowShortTopic:  true,
		ShowAllSections: true,
	}, &out)
	if err != nil {
		t.Fatalf("RenderTopicHelpWithWriter() error = %v", err)
	}

	if !strings.Contains(rendered, "Topic 1") {
		t.Fatalf("expected fallback title in output, got %q", rendered)
	}
	if strings.Contains(rendered, "Composed Topic") {
		t.Fatalf("expected fallback output, got composed output %q", rendered)
	}
}
