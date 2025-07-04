package store

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"os"
	"testing"
)

func TestFindSections(t *testing.T) {
	dbPath := "test_help.db"
	defer os.Remove(dbPath)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	sec := &model.Section{
		Slug:        "test-section",
		Title:       "Test Section",
		SectionType: model.SectionTutorial,
		IsTopLevel:  true,
		Topics:      []string{"foo", "bar"},
		Content:     "This is a test section.",
	}
	if err := s.UpsertSection(sec); err != nil {
		t.Fatalf("failed to upsert section: %v", err)
	}

	// Debug: Query FTS table directly
	row := s.db.QueryRow("SELECT slug, title, content FROM section_fts WHERE rowid = (SELECT id FROM sections WHERE slug = ?)", sec.Slug)
	var ftsSlug, ftsTitle, ftsContent string
	if err := row.Scan(&ftsSlug, &ftsTitle, &ftsContent); err != nil {
		t.Fatalf("FTS table missing row for section: %v", err)
	}
	t.Logf("FTS row: slug=%q title=%q content=%q", ftsSlug, ftsTitle, ftsContent)

	q := query.And(query.IsType(model.SectionTutorial), query.HasTopic("foo"))
	results, err := s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != sec.Slug {
		t.Errorf("expected slug %q, got %q", sec.Slug, results[0].Slug)
	}
}

func TestFindSectionsTextSearch(t *testing.T) {
	dbPath := "test_help.db"
	defer os.Remove(dbPath)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	sec := &model.Section{
		Slug:        "fts-section",
		Title:       "FTS Section",
		SectionType: model.SectionExample,
		Content:     "This is a unique full-text search test.",
	}
	if err := s.UpsertSection(sec); err != nil {
		t.Fatalf("failed to upsert section: %v", err)
	}

	// Rebuild FTS index to ensure the new section is searchable
	if _, err := s.db.Exec(`INSERT INTO section_fts(section_fts) VALUES ('rebuild')`); err != nil {
		t.Fatalf("failed to rebuild FTS index: %v", err)
	}

	// Debug: Query FTS table directly
	row := s.db.QueryRow("SELECT slug, title, content FROM section_fts WHERE rowid = (SELECT id FROM sections WHERE slug = ?)", sec.Slug)
	var ftsSlug, ftsTitle, ftsContent string
	if err := row.Scan(&ftsSlug, &ftsTitle, &ftsContent); err != nil {
		t.Fatalf("FTS table missing row for section: %v", err)
	}
	t.Logf("FTS row: slug=%q title=%q content=%q", ftsSlug, ftsTitle, ftsContent)

	// Print the generated SQL for debugging
	c := &query.Compiler{}
	q := query.TextSearch("unique")
	q(c)
	sqlStr, args := c.SQL()
	t.Logf("TextSearch SQL: %s, args: %v", sqlStr, args)

	// FTS5 index may need a rebuild in some SQLite setups, but should work for this test
	results, err := s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != sec.Slug {
		t.Errorf("expected slug %q, got %q", sec.Slug, results[0].Slug)
	}
}

func TestFTSBooleanAndPunctuation(t *testing.T) {
	dbPath := "fts_bool.db"
	defer os.Remove(dbPath)
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	sec1 := &model.Section{Slug: "fts-a", Title: "Alpha", Content: "foo bar baz"}
	sec2 := &model.Section{Slug: "fts-b", Title: "Beta", Content: "foo qux"}
	sec3 := &model.Section{Slug: "fts-c", Title: "Gamma", Content: "quux corge"}
	for _, sec := range []*model.Section{sec1, sec2, sec3} {
		if err := s.UpsertSection(sec); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}
	s.db.Exec(`INSERT INTO section_fts(section_fts) VALUES ('rebuild')`)

	// FTS AND
	q := query.TextSearch("foo AND qux")
	results, err := s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("FTS AND failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "fts-b" {
		t.Errorf("expected only fts-b for 'foo AND qux', got %+v", results)
	}

	// FTS OR
	q = query.TextSearch("baz OR quux")
	results, err = s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("FTS OR failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'baz OR quux', got %+v", results)
	}

	// FTS NOT
	q = query.TextSearch("foo NOT bar")
	results, err = s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("FTS NOT failed: %v", err)
	}
	if len(results) != 1 || results[0].Slug != "fts-b" {
		t.Errorf("expected only fts-b for 'foo NOT bar', got %+v", results)
	}

	// FTS with punctuation
	sec4 := &model.Section{Slug: "fts-d", Title: "Delta", Content: "hello, world!"}
	if err := s.UpsertSection(sec4); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	s.db.Exec(`INSERT INTO section_fts(section_fts) VALUES ('rebuild')`)
	q = query.TextSearch("hello")
	results, err = s.Find(context.Background(), q)
	if err != nil {
		t.Fatalf("FTS punctuation failed: %v", err)
	}
	if len(results) == 0 || results[0].Slug != "fts-d" {
		t.Errorf("expected fts-d for 'hello', got %+v", results)
	}
}
