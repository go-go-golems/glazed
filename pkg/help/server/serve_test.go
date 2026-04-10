package server

import (
	"context"
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
