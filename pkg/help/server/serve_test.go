package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	helploader "github.com/go-go-golems/glazed/pkg/help/loader"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/web"
)

func TestNewServeHandler_ServesEmbeddedSPAAtRoot(t *testing.T) {
	st, _ := setupTestServer(t)
	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		t.Fatalf("web.NewSPAHandler: %v", err)
	}
	h := NewServeHandler(HandlerDeps{Store: st}, spaHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if ct := rw.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected html content-type, got %q", ct)
	}
	if !strings.Contains(rw.Body.String(), "Glazed Help Browser") {
		t.Fatalf("expected embedded index.html content, got %q", rw.Body.String())
	}
}

func TestNewMountedHandler_ServesAPIUnderPrefix(t *testing.T) {
	st, _ := setupTestServer(t)
	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		t.Fatalf("web.NewSPAHandler: %v", err)
	}
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, spaHandler)

	req := httptest.NewRequest(http.MethodGet, "/help/api/health", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if !strings.Contains(rw.Body.String(), `"ok":true`) {
		t.Fatalf("expected health response, got %q", rw.Body.String())
	}
}

func TestNewMountedHandler_ServesSPAUnderPrefix(t *testing.T) {
	st, _ := setupTestServer(t)
	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		t.Fatalf("web.NewSPAHandler: %v", err)
	}
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, spaHandler)

	req := httptest.NewRequest(http.MethodGet, "/help/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if !strings.Contains(rw.Body.String(), "Glazed Help Browser") {
		t.Fatalf("expected embedded index.html content, got %q", rw.Body.String())
	}
}

func TestMountPrefix_RejectsOutsidePrefix(t *testing.T) {
	st, _ := setupTestServer(t)
	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		t.Fatalf("web.NewSPAHandler: %v", err)
	}
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, spaHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}
}

func TestBuildServeLoaders_UsesAllExternalSources(t *testing.T) {
	settings := &ServeSettings{
		Paths:         []string{"./docs"},
		FromJSON:      []string{"a.json,b.json"},
		FromSQLite:    []string{"a.db"},
		FromGlazedCmd: []string{"pinocchio,sqleton"},
	}

	loaders := buildServeLoaders(settings)
	if len(loaders) != 4 {
		t.Fatalf("expected 4 loaders, got %d", len(loaders))
	}
	if !strings.Contains(loaders[3].String(), "pinocchio") || !strings.Contains(loaders[3].String(), "sqleton") {
		t.Fatalf("expected glazed command loader to include normalized binary list, got %q", loaders[3].String())
	}
}

type testContentLoader struct {
	name string
	fn   func(context.Context, *help.HelpSystem) error
}

func (l testContentLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	return l.fn(ctx, hs)
}

func (l testContentLoader) String() string { return l.name }

func TestLoadServeSources_DoesNotClearExistingSectionsWhenStagingFails(t *testing.T) {
	hs := help.NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:        "existing-topic",
		Title:       "Existing Topic",
		SectionType: model.SectionGeneralTopic,
	})

	err := loadServeSources(context.Background(), hs, []helploader.ContentLoader{
		testContentLoader{
			name: "failing loader",
			fn: func(context.Context, *help.HelpSystem) error {
				return fmt.Errorf("transient read failure")
			},
		},
	}, true, nil)
	if err == nil {
		t.Fatalf("expected failing staged reload to return an error")
	}

	count, err := hs.Store.Count(context.Background())
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected existing section to survive failed reload, got count %d", count)
	}
	if _, err := hs.GetSectionWithSlug("existing-topic"); err != nil {
		t.Fatalf("expected existing-topic to survive failed reload: %v", err)
	}
}

func TestReplaceStoreWithPaths_ClearsPreloadedSections(t *testing.T) {
	hs := help.NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:        "embedded-topic",
		Title:       "Embedded Topic",
		SectionType: model.SectionGeneralTopic,
	})

	tmpDir := t.TempDir()
	markdown := `---
Title: Explicit Topic
Slug: explicit-topic
SectionType: GeneralTopic
---

Loaded from explicit path.
`
	filePath := filepath.Join(tmpDir, "explicit.md")
	if err := os.WriteFile(filePath, []byte(markdown), 0o644); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	if err := helploader.ReplaceStoreWithPaths(context.Background(), hs, []string{tmpDir}); err != nil {
		t.Fatalf("ReplaceStoreWithPaths: %v", err)
	}

	count, err := hs.Store.Count(context.Background())
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 section after replace, got %d", count)
	}

	if _, err := hs.GetSectionWithSlug("explicit-topic"); err != nil {
		t.Fatalf("expected explicit-topic to be loaded: %v", err)
	}
	if _, err := hs.GetSectionWithSlug("embedded-topic"); err == nil {
		t.Fatalf("expected embedded-topic to be cleared when explicit paths are provided")
	}
}

func TestNewServeHandler_AutoAssignsDefaultPackage_Issue571(t *testing.T) {
	// Reproduce issue #571: sections loaded without a package name should
	// still appear in the API after NewServeHandler auto-assigns one.
	hs := help.NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:        "example-topic",
		Title:       "Example Topic",
		SectionType: model.SectionGeneralTopic,
		Short:       "An example section",
		// PackageName intentionally left empty — this is the bug scenario.
	})

	// Verify section has no package name before the fix.
	sections, err := hs.Store.List(context.Background(), "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sections) != 1 || sections[0].PackageName != "" {
		t.Fatalf("expected section with empty package_name, got %q", sections[0].PackageName)
	}

	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		t.Fatalf("web.NewSPAHandler: %v", err)
	}

	handler := NewServeHandler(HandlerDeps{Store: hs.Store}, spaHandler)

	// Step 1: GET /api/packages should return the default package.
	req := httptest.NewRequest(http.MethodGet, "/api/packages", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("packages: expected status 200, got %d", rw.Code)
	}

	var pkgResp ListPackagesResponse
	if err := json.NewDecoder(rw.Body).Decode(&pkgResp); err != nil {
		t.Fatalf("decode packages: %v", err)
	}
	if len(pkgResp.Packages) != 1 || pkgResp.Packages[0].Name != "default" {
		var names []string
		for _, p := range pkgResp.Packages {
			names = append(names, p.Name)
		}
		t.Fatalf("expected package 'default', got %v", names)
	}

	// Step 2: GET /api/sections?package=default should return 1 section.
	req = httptest.NewRequest(http.MethodGet, "/api/sections?package=default", nil)
	rw = httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("sections: expected status 200, got %d", rw.Code)
	}

	var secResp ListSectionsResponse
	if err := json.NewDecoder(rw.Body).Decode(&secResp); err != nil {
		t.Fatalf("decode sections: %v", err)
	}
	if secResp.Total != 1 {
		t.Fatalf("expected 1 section for package 'default', got %d", secResp.Total)
	}
	if len(secResp.Sections) != 1 || secResp.Sections[0].Slug != "example-topic" {
		t.Fatalf("expected section 'example-topic', got %v", secResp.Sections)
	}
}

func TestNewServeHandler_APIOnlyWithNoSPA(t *testing.T) {
	// API-only mode (nil SPA handler) should work and auto-assign packages.
	hs := help.NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:        "api-only-topic",
		Title:       "API Only Topic",
		SectionType: model.SectionGeneralTopic,
	})

	handler := NewServeHandler(HandlerDeps{Store: hs.Store}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/sections?package=default", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}

	var secResp ListSectionsResponse
	if err := json.NewDecoder(rw.Body).Decode(&secResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if secResp.Total != 1 {
		t.Fatalf("expected 1 section in API-only mode, got %d", secResp.Total)
	}
}

func TestNewServeHandler_DoesNotOverwriteExistingPackage(t *testing.T) {
	// If sections already have a package name, auto-assign should be a no-op.
	hs := help.NewHelpSystem()
	hs.AddSection(&model.Section{
		Slug:        "glazed-topic",
		Title:       "Glazed Topic",
		SectionType: model.SectionGeneralTopic,
		PackageName: "glazed",
	})

	_ = NewServeHandler(HandlerDeps{Store: hs.Store}, nil)

	sections, err := hs.Store.List(context.Background(), "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sections) != 1 || sections[0].PackageName != "glazed" {
		t.Fatalf("expected package_name to remain 'glazed', got %q", sections[0].PackageName)
	}
}
