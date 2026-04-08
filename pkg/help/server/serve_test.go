package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/web"
)

func TestNewServeHandler_ServesEmbeddedSPAAtRoot(t *testing.T) {
	st, _ := setupTestServer(t)
	h := NewServeHandler(HandlerDeps{Store: st}, web.FS)

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
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, web.FS)

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
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, web.FS)

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
	h := NewMountedHandler("/help", HandlerDeps{Store: st}, web.FS)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.Code)
	}
}
