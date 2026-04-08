package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSPAHandler_Root(t *testing.T) {
	h, err := NewSPAHandler()
	if err != nil {
		t.Fatalf("NewSPAHandler: %v", err)
	}

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

func TestNewSPAHandler_Fallback(t *testing.T) {
	h, err := NewSPAHandler()
	if err != nil {
		t.Fatalf("NewSPAHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/some/client/route", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	if !strings.Contains(rw.Body.String(), "Glazed Help Browser") {
		t.Fatalf("expected SPA fallback content, got %q", rw.Body.String())
	}
}
