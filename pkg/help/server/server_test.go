package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

func setupTestServer(t *testing.T) (*store.Store, http.Handler) {
	t.Helper()
	st, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("store.NewInMemory: %v", err)
	}

	ctx := context.Background()
	sections := []*model.Section{
		{
			Slug:        "intro",
			SectionType: model.SectionGeneralTopic,
			Title:       "Introduction",
			Short:       "Welcome to the help system",
			Content:     "# Introduction\n\nWelcome!",
			Topics:      []string{"getting-started"},
		},
		{
			Slug:        "database",
			SectionType: model.SectionExample,
			Title:       "Database Example",
			Short:       "How to connect to a database",
			Content:     "# Database\n\nConnect using --db-url.",
			Topics:      []string{"database"},
			Flags:       []string{"--db-url"},
		},
		{
			Slug:        "config",
			SectionType: model.SectionGeneralTopic,
			Title:       "Configuration",
			Short:       "Configure the application",
			Content:     "# Configuration\n\nUse the config file.",
			Topics:      []string{"config"},
		},
	}
	for _, s := range sections {
		if err := st.Insert(ctx, s); err != nil {
			t.Fatalf("store.Insert: %v", err)
		}
	}

	deps := HandlerDeps{Store: st}
	return st, NewHandler(deps)
}

func TestHandleHealth(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Sections != 3 {
		t.Errorf("expected 3 sections, got %d", resp.Sections)
	}
}

func TestHandleListSections(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Total)
	}
	if len(resp.Sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(resp.Sections))
	}
	// Content should be omitted in list view (validated by type: no Content field).
}

func TestHandleListSections_IncludesHeadingMetadata(t *testing.T) {
	st, handler := setupTestServer(t)
	ctx := context.Background()
	section, err := st.GetBySlug(ctx, "intro")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	section.Content = "# Introduction\n\n## First Steps\n\n```\n## Ignored\n```\n\n### Details"
	if err := st.Update(ctx, section); err != nil {
		t.Fatalf("Update: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sections", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	var intro SectionSummary
	for _, section := range resp.Sections {
		if section.Slug == "intro" {
			intro = section
		}
	}
	if len(intro.Headings) != 2 {
		t.Fatalf("expected 2 headings, got %#v", intro.Headings)
	}
	if intro.Headings[0].ID != "first-steps" || intro.Headings[1].ID != "details" {
		t.Fatalf("unexpected headings: %#v", intro.Headings)
	}
}

func TestHandleListSections_FilterByType(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections?type=Example", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total=1 (Example only), got %d", resp.Total)
	}
	if resp.Sections[0].Slug != "database" {
		t.Errorf("expected slug=database, got %s", resp.Sections[0].Slug)
	}
}

func TestHandleListSections_FilterByTopic(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections?topic=database", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total=1 (topic=database), got %d", resp.Total)
	}
	if resp.Sections[0].Slug != "database" {
		t.Errorf("expected slug=database, got %s", resp.Sections[0].Slug)
	}
}

func TestHandleListSections_FilterBySearch(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections?q=connect", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total=1 (search 'connect'), got %d", resp.Total)
	}
	if resp.Sections[0].Slug != "database" {
		t.Errorf("expected slug=database, got %s", resp.Sections[0].Slug)
	}
}

func TestHandleListSections_Pagination(t *testing.T) {
	_, handler := setupTestServer(t)

	// Request with limit=1, offset=0.
	req := httptest.NewRequest(http.MethodGet, "/api/sections?limit=1&offset=0", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	var resp ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Total)
	}
	if len(resp.Sections) != 1 {
		t.Errorf("expected 1 section (limit=1), got %d", len(resp.Sections))
	}
	if resp.Limit != 1 {
		t.Errorf("expected limit=1, got %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Errorf("expected offset=0, got %d", resp.Offset)
	}
}

func TestHandleGetSection(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections/intro", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	var resp SectionDetail
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Slug != "intro" {
		t.Errorf("expected slug=intro, got %s", resp.Slug)
	}
	if resp.Type != "GeneralTopic" {
		t.Errorf("expected type=GeneralTopic, got %s", resp.Type)
	}
	if resp.Content == "" {
		t.Error("expected content to be present in detail view")
	}
}

func TestHandleGetSection_NotFound(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections/nonexistent", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rw.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if resp.Error != "not_found" {
		t.Errorf("expected error='not_found', got '%s'", resp.Error)
	}
}

func TestHandleGetSection_PathValueRouting(t *testing.T) {
	// Verify that the handler correctly extracts the slug from the URL.
	_, handler := setupTestServer(t)

	tests := []struct {
		url        string
		wantSlug   string
		wantStatus int
	}{
		{"/api/sections/config", "config", http.StatusOK},
		{"/api/sections/database", "database", http.StatusOK},
		{"/api/sections/intro", "intro", http.StatusOK},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, tc.url, nil)
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		if rw.Code != tc.wantStatus {
			t.Errorf("%s: expected status %d, got %d", tc.url, tc.wantStatus, rw.Code)
		}
		if tc.wantStatus == http.StatusOK {
			var resp SectionDetail
			if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
				t.Errorf("%s: json.Unmarshal: %v", tc.url, err)
			}
			if resp.Slug != tc.wantSlug {
				t.Errorf("%s: expected slug=%s, got %s", tc.url, tc.wantSlug, resp.Slug)
			}
		}
	}
}

func TestNewHandler_PanicsNilStore(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil store")
		}
	}()
	NewHandler(HandlerDeps{})
}

func TestNewHandler_PanicsNilLogger(t *testing.T) {
	// Should NOT panic — nil logger is replaced with slog.Default().
	st, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("store.NewInMemory: %v", err)
	}
	NewHandler(HandlerDeps{Store: st})
}

func TestCORSHeaders(t *testing.T) {
	_, handler := setupTestServer(t)

	// CORS preflight request.
	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", rw.Code)
	}
	if got := rw.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got %s", got)
	}
}

func TestContentTypeJSON(t *testing.T) {
	_, handler := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if got := rw.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %s", got)
	}
}

func TestHandlePackagesAndPackageFilters(t *testing.T) {
	st, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("store.NewInMemory: %v", err)
	}
	ctx := context.Background()
	for _, section := range []*model.Section{
		{PackageName: "glazed", Slug: "intro", Title: "Glazed Intro", SectionType: model.SectionGeneralTopic},
		{PackageName: "pinocchio", PackageVersion: "v1", Slug: "intro", Title: "Pinocchio Intro", SectionType: model.SectionTutorial},
	} {
		if err := st.Upsert(ctx, section); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}
	handler := NewHandler(HandlerDeps{Store: st})

	req := httptest.NewRequest(http.MethodGet, "/api/packages", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("packages status = %d: %s", rw.Code, rw.Body.String())
	}
	var packages ListPackagesResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &packages); err != nil {
		t.Fatalf("json.Unmarshal packages: %v", err)
	}
	if len(packages.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %#v", packages.Packages)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/sections?package=pinocchio&version=v1", nil)
	rw = httptest.NewRecorder()
	handler.ServeHTTP(rw, req)
	var list ListSectionsResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &list); err != nil {
		t.Fatalf("json.Unmarshal list: %v", err)
	}
	if list.Total != 1 || list.Sections[0].Title != "Pinocchio Intro" {
		t.Fatalf("unexpected filtered list: %#v", list)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/sections/intro", nil)
	rw = httptest.NewRecorder()
	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected ambiguous slug 400, got %d", rw.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/sections/intro?package=glazed", nil)
	rw = httptest.NewRecorder()
	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected package-qualified detail 200, got %d", rw.Code)
	}
	var detail SectionDetail
	if err := json.Unmarshal(rw.Body.Bytes(), &detail); err != nil {
		t.Fatalf("json.Unmarshal detail: %v", err)
	}
	if detail.Title != "Glazed Intro" {
		t.Fatalf("expected Glazed Intro, got %q", detail.Title)
	}
}
