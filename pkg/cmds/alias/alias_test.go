package alias

import (
	"strings"
	"testing"
)

func TestNewCommandAliasFromYAML_ScalarAliasForRemainsRelative(t *testing.T) {
	a, err := NewCommandAliasFromYAML(strings.NewReader("name: concise-doc\naliasFor: go\n"), WithParents("code", "go"))
	if err != nil {
		t.Fatalf("NewCommandAliasFromYAML failed: %v", err)
	}

	if got := a.AliasFor.Segments(); len(got) != 1 || got[0] != "go" {
		t.Fatalf("expected scalar alias target [go], got %#v", got)
	}
	if got := a.ResolveAliasedCommandPath(); len(got) != 3 || got[0] != "code" || got[1] != "go" || got[2] != "go" {
		t.Fatalf("expected legacy relative path [code go go], got %#v", got)
	}
}

func TestNewCommandAliasFromYAML_SequenceAliasForResolvesAbsolutePath(t *testing.T) {
	a, err := NewCommandAliasFromYAML(strings.NewReader("name: concise-doc\naliasFor: [code, go]\n"), WithParents("code", "go"))
	if err != nil {
		t.Fatalf("NewCommandAliasFromYAML failed: %v", err)
	}

	if got := a.ResolveAliasedCommandPath(); len(got) != 2 || got[0] != "code" || got[1] != "go" {
		t.Fatalf("expected absolute alias path [code go], got %#v", got)
	}
}

func TestNewCommandAliasFromYAML_WhitespaceSequenceEntryNormalizesToPathSegments(t *testing.T) {
	a, err := NewCommandAliasFromYAML(strings.NewReader("name: concise-doc\naliasFor: [code go]\n"), WithParents("code", "go"))
	if err != nil {
		t.Fatalf("NewCommandAliasFromYAML failed: %v", err)
	}

	if got := a.ResolveAliasedCommandPath(); len(got) != 2 || got[0] != "code" || got[1] != "go" {
		t.Fatalf("expected normalized absolute alias path [code go], got %#v", got)
	}
}

func TestNewCommandAliasFromYAML_EmptyAliasForIsRejected(t *testing.T) {
	if _, err := NewCommandAliasFromYAML(strings.NewReader("name: concise-doc\naliasFor: []\n")); err == nil {
		t.Fatal("expected empty aliasFor to fail")
	}
}
