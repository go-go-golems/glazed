package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishCommandDryRun(t *testing.T) {
	path := createDocsctlHelpDB(t, "intro")
	stdout, _, err := executeDocsctl("publish", "--server", "http://example.test", "--package", "pinocchio", "--version", "v1", "--file", path, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("publish dry-run: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["dryRun"] != true {
		t.Fatalf("unexpected payload: %s", stdout)
	}
}

func TestPublishCommandDryRunValidationFailure(t *testing.T) {
	path := createDocsctlHelpDB(t, "")
	_, _, err := executeDocsctl("publish", "--server", "http://example.test", "--package", "pinocchio", "--version", "v1", "--file", path, "--dry-run")
	if err == nil || !strings.Contains(err.Error(), "empty slugs") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestPublishCommandMissingToken(t *testing.T) {
	t.Setenv("DOCS_YOLO_PUBLISH_TOKEN", "")
	path := createDocsctlHelpDB(t, "intro")
	_, _, err := executeDocsctl("publish", "--server", "http://example.test", "--package", "pinocchio", "--version", "v1", "--file", path)
	if err == nil || !strings.Contains(err.Error(), "publish token is required") {
		t.Fatalf("expected token error, got %v", err)
	}
}

func TestPublishCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusForbidden) }))
	defer server.Close()
	path := createDocsctlHelpDB(t, "intro")
	_, _, err := executeDocsctl("publish", "--server", server.URL, "--package", "pinocchio", "--version", "v1", "--file", path, "--token", "secret")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected server error, got %v", err)
	}
}

func TestPublishCommandPrintSchemaDoesNotUpload(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.Error(w, "unexpected upload", http.StatusInternalServerError)
	}))
	defer server.Close()
	path := createDocsctlHelpDB(t, "intro")
	stdout, _, err := executeDocsctl("publish", "--server", server.URL, "--package", "pinocchio", "--version", "v1", "--file", path, "--token", "secret", "--print-schema")
	if err != nil {
		t.Fatalf("publish --print-schema: %v", err)
	}
	if called {
		t.Fatalf("expected --print-schema to return before upload")
	}
	if !strings.Contains(stdout, "properties") || !strings.Contains(stdout, "package") {
		t.Fatalf("unexpected schema stdout: %s", stdout)
	}
}

func TestPublishCommandPrintParsedFieldsDoesNotRequireTokenOrUpload(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.Error(w, "unexpected upload", http.StatusInternalServerError)
	}))
	defer server.Close()
	path := createDocsctlHelpDB(t, "intro")
	stdout, _, err := executeDocsctl("publish", "--server", server.URL, "--package", "pinocchio", "--version", "v1", "--file", path, "--print-parsed-fields")
	if err != nil {
		t.Fatalf("publish --print-parsed-fields: %v", err)
	}
	if called {
		t.Fatalf("expected --print-parsed-fields to return before upload")
	}
	if !strings.Contains(stdout, "default") || !strings.Contains(stdout, "pinocchio") {
		t.Fatalf("unexpected parsed-fields stdout: %s", stdout)
	}
}

func TestPublishCommandSuccess(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	path := createDocsctlHelpDB(t, "intro")
	stdout, _, err := executeDocsctl("publish", "--server", server.URL, "--package", "pinocchio", "--version", "v1", "--file", path, "--token", "secret")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("unexpected auth: %s", gotAuth)
	}
	if !strings.Contains(stdout, "OK: published") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}
