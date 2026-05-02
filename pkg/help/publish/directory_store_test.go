package publish

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestDirectoryPackageStorePublishAndList(t *testing.T) {
	root := t.TempDir()
	store := NewDirectoryPackageStore(root)
	store.Now = func() time.Time { return time.Date(2026, 5, 2, 18, 0, 0, 0, time.UTC) }
	db := createDirectoryStoreDB(t, "intro")
	result, err := ValidateSQLiteHelpDB(context.Background(), db, SQLiteValidationOptions{PackageName: "pinocchio", Version: "v1"})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	published, err := store.Publish(context.Background(), "pinocchio", "v1", db, result, &PublisherIdentity{Subject: "tester"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if published.Path != "pinocchio/v1/pinocchio.db" || published.PublishedBy != "tester" || published.SHA256 == "" {
		t.Fatalf("unexpected published: %#v", published)
	}
	if _, err := os.Stat(filepath.Join(root, "pinocchio", "v1", "pinocchio.db")); err != nil {
		t.Fatalf("stat db: %v", err)
	}
	packages, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(packages) != 1 || packages[0].PackageName != "pinocchio" {
		t.Fatalf("unexpected packages: %#v", packages)
	}
}

func TestDirectoryPackageStoreRejectsTraversal(t *testing.T) {
	store := NewDirectoryPackageStore(t.TempDir())
	db := createDirectoryStoreDB(t, "intro")
	result := &SQLiteValidationResult{SectionCount: 1, SlugCount: 1}
	_, err := store.Publish(context.Background(), "../bad", "v1", db, result, nil)
	if err == nil || !strings.Contains(err.Error(), "path separators") {
		t.Fatalf("expected path validation error, got %v", err)
	}
}

func TestDirectoryPackageStoreReplacesAtomically(t *testing.T) {
	root := t.TempDir()
	store := NewDirectoryPackageStore(root)
	first := createDirectoryStoreDB(t, "first")
	firstResult, _ := ValidateSQLiteHelpDB(context.Background(), first, SQLiteValidationOptions{PackageName: "pinocchio", Version: "v1"})
	if _, err := store.Publish(context.Background(), "pinocchio", "v1", first, firstResult, nil); err != nil {
		t.Fatalf("publish first: %v", err)
	}
	second := createDirectoryStoreDB(t, "second")
	secondResult, _ := ValidateSQLiteHelpDB(context.Background(), second, SQLiteValidationOptions{PackageName: "pinocchio", Version: "v1"})
	if _, err := store.Publish(context.Background(), "pinocchio", "v1", second, secondResult, nil); err != nil {
		t.Fatalf("publish second: %v", err)
	}
	publishedDB := filepath.Join(root, "pinocchio", "v1", "pinocchio.db")
	validated, err := ValidateSQLiteHelpDB(context.Background(), publishedDB, SQLiteValidationOptions{PackageName: "pinocchio", Version: "v1"})
	if err != nil {
		t.Fatalf("validate published: %v", err)
	}
	if validated.SectionCount != 1 {
		t.Fatalf("unexpected count: %d", validated.SectionCount)
	}
}

func createDirectoryStoreDB(t *testing.T, slug string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "help.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.Exec(`CREATE TABLE sections (id INTEGER PRIMARY KEY, slug TEXT, title TEXT NOT NULL); INSERT INTO sections (slug,title) VALUES (?, 'Title')`, slug)
	if err != nil {
		t.Fatalf("create db: %v", err)
	}
	return path
}
