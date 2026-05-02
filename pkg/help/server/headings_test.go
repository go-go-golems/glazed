package server

import "testing"

func TestExtractHeadings(t *testing.T) {
	content := `# Test Section

## Overview
Some text.

` + "```" + `
## Ignored In Fence
` + "```" + `

### Details!
#### Deep Dive ###
##### Too Deep
#Another invalid heading
`

	headings := ExtractHeadings(content, "Test Section")
	want := []SectionHeading{
		{ID: "overview", Level: 2, Text: "Overview"},
		{ID: "details", Level: 3, Text: "Details!"},
		{ID: "deep-dive", Level: 4, Text: "Deep Dive"},
	}
	if len(headings) != len(want) {
		t.Fatalf("expected %d headings, got %d: %#v", len(want), len(headings), headings)
	}
	for i := range want {
		if headings[i] != want[i] {
			t.Fatalf("heading %d: expected %#v, got %#v", i, want[i], headings[i])
		}
	}
}

func TestExtractHeadingsMakesDuplicateIDsUnique(t *testing.T) {
	content := `## Install

### Install

## Install
`
	headings := ExtractHeadings(content, "Package")
	want := []SectionHeading{
		{ID: "install", Level: 2, Text: "Install"},
		{ID: "install-2", Level: 3, Text: "Install"},
		{ID: "install-3", Level: 2, Text: "Install"},
	}
	if len(headings) != len(want) {
		t.Fatalf("expected %d headings, got %d: %#v", len(want), len(headings), headings)
	}
	for i := range want {
		if headings[i] != want[i] {
			t.Fatalf("heading %d: expected %#v, got %#v", i, want[i], headings[i])
		}
	}
}

func TestSlugifyHeading(t *testing.T) {
	tests := map[string]string{
		"Events, Streaming, and Watermill": "events-streaming-and-watermill",
		"  Types & APIs  ":                 "types-apis",
		"NewEventRouter()":                 "neweventrouter",
		"---":                              "section",
	}
	for input, want := range tests {
		if got := SlugifyHeading(input); got != want {
			t.Fatalf("SlugifyHeading(%q): expected %q, got %q", input, want, got)
		}
	}
}
