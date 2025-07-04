package store

import (
	"os"
	"path/filepath"
	"testing"
)

const md1 = `---
slug: section-1
title: Section 1
sectionType: Example
---
Content 1.`
const md2 = `---
slug: section-2
title: Section 2
sectionType: Tutorial
---
Content 2.`

func TestSyncMarkdownDir(t *testing.T) {
	dir := "test_md_dir"
	os.Mkdir(dir, 0755)
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "1.md"), []byte(md1), 0644)
	os.WriteFile(filepath.Join(dir, "2.md"), []byte(md2), 0644)

	s, err := Open("test_sync.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { s.Close(); os.Remove("test_sync.db") }()

	sections, err := s.SyncMarkdownDir(dir)
	if err != nil {
		t.Fatalf("sync dir: %v", err)
	}
	if len(sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(sections))
	}
	// Query to check both slugs are present
	for _, slug := range []string{"section-1", "section-2"} {
		found := false
		for _, sec := range sections {
			if sec.Slug == slug {
				found = true
			}
		}
		if !found {
			t.Errorf("section %s not found in loaded sections", slug)
		}
	}
} 