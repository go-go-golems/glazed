package publish

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilePublisherCatalogSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "publishers.json")
	data := `{"publishers":[{"package":"pinocchio","subject":"repo:pinocchio","tokenHash":"` + HashPublishToken("secret") + `"}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	source := FilePublisherCatalogSource{Path: path}
	records, err := source.LoadPublisherTokenRecords(context.Background())
	if err != nil {
		t.Fatalf("LoadPublisherTokenRecords: %v", err)
	}
	if len(records) != 1 || records[0].PackageName != "pinocchio" || records[0].Subject != "repo:pinocchio" {
		t.Fatalf("unexpected records: %#v", records)
	}
}
