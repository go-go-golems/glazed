package store

import (
	"os"
	"testing"
)

func TestOpenAndSchema(t *testing.T) {
	dbPath := "test_help.db"
	defer os.Remove(dbPath)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("failed to close store: %v", err)
	}
}
