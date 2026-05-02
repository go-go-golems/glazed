package publish

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestValidateSQLiteHelpDB_Valid(t *testing.T) {
	path := createHelpDB(t, []testSection{{Slug: "intro", Title: "Intro"}, {Slug: "usage", Title: "Usage"}})

	result, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{PackageName: "pinocchio", Version: "v1.2.3"})
	if err != nil {
		t.Fatalf("ValidateSQLiteHelpDB: %v", err)
	}
	if result.SectionCount != 2 {
		t.Fatalf("expected 2 sections, got %d", result.SectionCount)
	}
	if result.SlugCount != 2 {
		t.Fatalf("expected 2 slugs, got %d", result.SlugCount)
	}
	if result.PackageName != "pinocchio" || result.Version != "v1.2.3" {
		t.Fatalf("unexpected package/version: %#v", result)
	}
}

func TestValidateSQLiteHelpDB_MissingFile(t *testing.T) {
	_, err := ValidateSQLiteHelpDB(context.Background(), filepath.Join(t.TempDir(), "missing.db"), SQLiteValidationOptions{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateSQLiteHelpDB_NonSQLiteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not.db")
	if err := os.WriteFile(path, []byte("not sqlite"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateSQLiteHelpDB_MissingSectionsTable(t *testing.T) {
	path := createSQLiteDB(t, `CREATE TABLE other (id INTEGER PRIMARY KEY);`)
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	assertErrContains(t, err, "missing required sections table")
}

func TestValidateSQLiteHelpDB_EmptySections(t *testing.T) {
	path := createHelpDB(t, nil)
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	assertErrContains(t, err, "contains no sections")
}

func TestValidateSQLiteHelpDB_MissingSlugColumn(t *testing.T) {
	path := createSQLiteDB(t, `CREATE TABLE sections (title TEXT NOT NULL); INSERT INTO sections (title) VALUES ('Intro');`)
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	assertErrContains(t, err, `missing required column "slug"`)
}

func TestValidateSQLiteHelpDB_EmptySlug(t *testing.T) {
	path := createHelpDB(t, []testSection{{Slug: "", Title: "No slug"}})
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	assertErrContains(t, err, "empty slugs")
}

func TestValidateSQLiteHelpDB_DuplicateSlug(t *testing.T) {
	path := createHelpDB(t, []testSection{{Slug: "intro", Title: "Intro"}, {Slug: "intro", Title: "Intro again"}})
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{})
	assertErrContains(t, err, "duplicate slugs")
}

func TestValidateSQLiteHelpDB_InvalidPackageVersionOptions(t *testing.T) {
	path := createHelpDB(t, []testSection{{Slug: "intro", Title: "Intro"}})
	_, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{PackageName: "../bad", Version: "v1"})
	assertErrContains(t, err, "path separators")
}

func TestValidateSQLiteHelpDB_MetadataWarnings(t *testing.T) {
	path := createHelpDB(t, []testSection{{PackageName: "source", PackageVersion: "v0", Slug: "intro", Title: "Intro"}})
	result, err := ValidateSQLiteHelpDB(context.Background(), path, SQLiteValidationOptions{PackageName: "target", Version: "v1"})
	if err != nil {
		t.Fatalf("ValidateSQLiteHelpDB: %v", err)
	}
	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %#v", result.Warnings)
	}
}

type testSection struct {
	PackageName    string
	PackageVersion string
	Slug           string
	Title          string
}

func createHelpDB(t *testing.T, sections []testSection) string {
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
			package_name TEXT NOT NULL DEFAULT '',
			package_version TEXT NOT NULL DEFAULT '',
			slug TEXT,
			title TEXT NOT NULL,
			content TEXT
		);
	`)
	if err != nil {
		t.Fatalf("create sections: %v", err)
	}
	for _, section := range sections {
		_, err := db.Exec(`INSERT INTO sections (package_name, package_version, slug, title, content) VALUES (?, ?, ?, ?, '')`, section.PackageName, section.PackageVersion, section.Slug, section.Title)
		if err != nil {
			t.Fatalf("insert section: %v", err)
		}
	}
	return path
}

func createSQLiteDB(t *testing.T, sqlText string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(sqlText); err != nil {
		t.Fatalf("exec fixture SQL: %v", err)
	}
	return path
}

func assertErrContains(t *testing.T, err error, text string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", text)
	}
	if !strings.Contains(err.Error(), text) {
		t.Fatalf("expected error containing %q, got %v", text, err)
	}
}
