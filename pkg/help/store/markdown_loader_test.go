package store

import (
	"os"
	"testing"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

const sampleMarkdown = `---
slug: test-md-section
title: Test MD Section
sectionType: Tutorial
isTopLevel: true
topics:
  - foo
  - bar
---
This is the markdown content.
`

func TestLoadSectionFromMarkdown(t *testing.T) {
	path := "test_section.md"
	if err := os.WriteFile(path, []byte(sampleMarkdown), 0644); err != nil {
		t.Fatalf("failed to write test markdown: %v", err)
	}
	defer os.Remove(path)

	sec, err := LoadSectionFromMarkdown(path)
	if err != nil {
		t.Fatalf("LoadSectionFromMarkdown failed: %v", err)
	}
	if sec.Slug != "test-md-section" || sec.Title != "Test MD Section" || sec.SectionType != "Tutorial" || !sec.IsTopLevel {
		t.Errorf("section fields not loaded correctly: %+v", sec)
	}
	if sec.Content != "This is the markdown content." {
		t.Errorf("content not loaded correctly: %q", sec.Content)
	}
	if len(sec.Topics) != 2 || sec.Topics[0] != "foo" || sec.Topics[1] != "bar" {
		t.Errorf("topics not loaded correctly: %+v", sec.Topics)
	}
}

func TestSectionNoTopicsFlagsCommands(t *testing.T) {
	path := "test_section_empty.md"
	md := `---
slug: empty-section
title: Empty Section
sectionType: Example
---
No extras.`
	if err := os.WriteFile(path, []byte(md), 0644); err != nil {
		t.Fatalf("failed to write test markdown: %v", err)
	}
	defer os.Remove(path)
	sec, err := LoadSectionFromMarkdown(path)
	if err != nil {
		t.Fatalf("LoadSectionFromMarkdown failed: %v", err)
	}
	if sec.Slug != "empty-section" || sec.Title != "Empty Section" || sec.SectionType != "Example" {
		t.Errorf("section fields not loaded correctly: %+v", sec)
	}
	if len(sec.Topics) != 0 || len(sec.Flags) != 0 || len(sec.Commands) != 0 {
		t.Errorf("expected empty arrays, got: topics=%v flags=%v commands=%v", sec.Topics, sec.Flags, sec.Commands)
	}
}

func TestDuplicateSlugUpsert(t *testing.T) {
	dbPath := "test_upsert.db"
	defer os.Remove(dbPath)
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	sec1 := &model.Section{Slug: "dup", Title: "First", SectionType: "Example"}
	sec2 := &model.Section{Slug: "dup", Title: "Second", SectionType: "Example"}
	if err := s.UpsertSection(sec1); err != nil {
		t.Fatalf("upsert 1: %v", err)
	}
	if err := s.UpsertSection(sec2); err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	row := s.db.QueryRow("SELECT title FROM sections WHERE slug = ?", "dup")
	var title string
	if err := row.Scan(&title); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if title != "Second" {
		t.Errorf("expected title 'Second', got %q", title)
	}
}

func TestMalformedFrontMatter(t *testing.T) {
	path := "test_bad.md"
	md := `---
slug: bad-section
title: Bad Section
sectionType Example  # missing colon
---
Bad front-matter.`
	if err := os.WriteFile(path, []byte(md), 0644); err != nil {
		t.Fatalf("failed to write test markdown: %v", err)
	}
	defer os.Remove(path)
	_, err := LoadSectionFromMarkdown(path)
	if err == nil {
		t.Errorf("expected error for malformed front-matter, got nil")
	}
} 