package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestValidateCommandSuccessText(t *testing.T) {
	path := createDocsctlHelpDB(t, "intro")
	stdout, stderr, err := executeDocsctl("validate", "--package", "pinocchio", "--version", "v1", "--file", path)
	if err != nil {
		t.Fatalf("execute validate: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "OK:") || !strings.Contains(stdout, "pinocchio@v1") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestValidateCommandSuccessJSON(t *testing.T) {
	path := createDocsctlHelpDB(t, "intro")
	stdout, stderr, err := executeDocsctl("validate", "--package", "pinocchio", "--version", "v1", "--file", path, "--json")
	if err != nil {
		t.Fatalf("execute validate: %v\nstderr=%s", err, stderr)
	}
	var payload struct {
		PackageName  string `json:"packageName"`
		Version      string `json:"version"`
		SectionCount int    `json:"sectionCount"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal stdout %q: %v", stdout, err)
	}
	if payload.PackageName != "pinocchio" || payload.Version != "v1" || payload.SectionCount != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestValidateCommandValidationFailure(t *testing.T) {
	path := createDocsctlHelpDB(t, "")
	_, _, err := executeDocsctl("validate", "--package", "pinocchio", "--version", "v1", "--file", path)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "empty slugs") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCommandRequiresFlags(t *testing.T) {
	_, _, err := executeDocsctl("validate")
	if err == nil {
		t.Fatalf("expected required flag error")
	}
}

func executeDocsctl(args ...string) (string, string, error) {
	cmd := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func createDocsctlHelpDB(t *testing.T, slug string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "help.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.Exec(`
		CREATE TABLE sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT,
			title TEXT NOT NULL
		);
		INSERT INTO sections (slug, title) VALUES (?, 'Intro');
	`, slug)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	return path
}
