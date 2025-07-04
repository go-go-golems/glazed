package store

import (
	"github.com/go-go-golems/glazed/pkg/help/model"
	"os"
	"testing"
)

func TestUpsertSection(t *testing.T) {
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
		Flags:       []string{"flag1"},
		Commands:    []string{"cmd1"},
		Content:     "This is a test section.",
	}
	if err := s.UpsertSection(sec); err != nil {
		t.Fatalf("failed to upsert section: %v", err)
	}

	// Check that the section is in the DB
	row := s.db.QueryRow("SELECT slug, title, sectionType, isTopLevel FROM sections WHERE slug = ?", sec.Slug)
	var slug, title, sectionType string
	var isTopLevel bool
	if err := row.Scan(&slug, &title, &sectionType, &isTopLevel); err != nil {
		t.Fatalf("failed to query section: %v", err)
	}
	if slug != sec.Slug || title != sec.Title || sectionType != sec.SectionType.String() || isTopLevel != sec.IsTopLevel {
		t.Errorf("section data mismatch: got %v, want %v", []any{slug, title, sectionType, isTopLevel}, []any{sec.Slug, sec.Title, sec.SectionType.String(), sec.IsTopLevel})
	}
}
