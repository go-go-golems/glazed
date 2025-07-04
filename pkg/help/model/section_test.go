package model

import (
	"testing"
)

func TestSectionTypeString(t *testing.T) {
	if SectionExample.String() != "Example" {
		t.Errorf("expected 'Example', got '%s'", SectionExample.String())
	}
}

func TestSectionStruct(t *testing.T) {
	s := &Section{
		ID:          1,
		Slug:        "test-section",
		Title:       "Test Section",
		SectionType: SectionTutorial,
		IsTopLevel:  true,
		Topics:      []string{"foo", "bar"},
	}
	if s.Slug != "test-section" {
		t.Errorf("expected slug 'test-section', got '%s'", s.Slug)
	}
	if s.SectionType != SectionTutorial {
		t.Errorf("expected SectionType 'Tutorial', got '%s'", s.SectionType)
	}
} 