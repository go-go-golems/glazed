package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// Store manages the SQLite database for help sections.
type Store struct {
	db *sql.DB
}

// Open opens (and initializes if needed) the SQLite database at the given path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// createSchema creates the required tables and indexes if they don't exist.
func createSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
