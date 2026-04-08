// Package server provides an HTTP API surface for browsing Glazed help sections.
// It is self-contained and has no dependencies on the CLI or Cobra — it only needs
// a HelpSystem instance to operate.
package server

import (
	"github.com/go-go-golems/glazed/pkg/help/model"
)

// ---------------------------------------------------------------------------
// Request types (URL/query params, decoded from JSON)
// ---------------------------------------------------------------------------

// ListSectionsParams describes optional filters for GET /api/sections.
// All fields are optional; zero values mean "no filter".
//
//go:generate set-zerolog-level Info
type ListSectionsParams struct {
	// SectionType filters by the section type (GeneralTopic, Example, Application, Tutorial).
	// Zero value means "all types".
	SectionType string `json:"section_type,omitempty"`
	// Topic filters by topic name (exact match, case-insensitive).
	Topic string `json:"topic,omitempty"`
	// Command filters by command name (exact match).
	Command string `json:"command,omitempty"`
	// Flag filters by flag name (exact match).
	Flag string `json:"flag,omitempty"`
	// Search performs a full-text search over title, subtitle, short description, and content.
	Search string `json:"search,omitempty"`
	// Limit caps the number of results. Zero or negative means "no limit".
	Limit int `json:"limit,omitempty"`
	// Offset skips the first N results for pagination. Zero means start at beginning.
	Offset int `json:"offset,omitempty"`
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

// SectionSummary is the public shape for a section in list/search results.
// It intentionally omits the full `content` field to keep responses small.
type SectionSummary struct {
	ID         int64    `json:"id"`
	Slug       string   `json:"slug"`
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Short      string   `json:"short"`
	Topics     []string `json:"topics"`
	IsTopLevel bool     `json:"isTopLevel"`
}

// SummaryFromModel converts a model.Section into a SectionSummary.
// It is the only place where this conversion is defined.
func SummaryFromModel(s *model.Section) SectionSummary {
	return SectionSummary{
		ID:         s.ID,
		Slug:       s.Slug,
		Type:       s.SectionType.String(), // "GeneralTopic" | "Example" | "Application" | "Tutorial"
		Title:      s.Title,
		Short:      s.Short,
		Topics:     s.Topics,
		IsTopLevel: s.IsTopLevel,
	}
}

// ListSectionsResponse is the shape of GET /api/sections and GET /api/sections/search.
type ListSectionsResponse struct {
	// Sections is the list of matching sections (summary shape, no content).
	Sections []SectionSummary `json:"sections"`
	// Total is the total number of matching sections (before pagination).
	Total int `json:"total"`
	// Limit reflects the requested limit, or -1 if no limit was applied.
	Limit int `json:"limit"`
	// Offset reflects the requested offset.
	Offset int `json:"offset"`
}

// SectionDetail is the full shape returned by GET /api/sections/:slug.
type SectionDetail struct {
	ID         int64    `json:"id"`
	Slug       string   `json:"slug"`
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Short      string   `json:"short"`
	Topics     []string `json:"topics"`
	Flags      []string `json:"flags"`
	Commands   []string `json:"commands"`
	IsTopLevel bool     `json:"isTopLevel"`
	// Content is the full rendered Markdown body.
	Content string `json:"content"`
}

// DetailFromModel converts a model.Section into a SectionDetail.
func DetailFromModel(s *model.Section) SectionDetail {
	return SectionDetail{
		ID:         s.ID,
		Slug:       s.Slug,
		Type:       s.SectionType.String(),
		Title:      s.Title,
		Short:      s.Short,
		Topics:     s.Topics,
		Flags:      s.Flags,
		Commands:   s.Commands,
		IsTopLevel: s.IsTopLevel,
		Content:    s.Content,
	}
}

// HealthResponse is the shape of GET /api/health.
type HealthResponse struct {
	OK       bool `json:"ok"`
	Sections int  `json:"sections"`
}

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

// ErrorResponse is the shape of all error responses (4xx/5xx).
type ErrorResponse struct {
	// Error is a short machine-readable code, e.g. "not_found" or "bad_request".
	Error string `json:"error"`
	// Message is a human-readable description.
	Message string `json:"message"`
}
