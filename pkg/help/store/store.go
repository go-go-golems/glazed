package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed storage for help sections
type Store struct {
	db *sql.DB
}

// New creates a new Store instance
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, errors.Wrap(err, "failed to enable foreign keys")
	}

	store := &Store{db: db}

	// Create tables and indexes
	if err := store.createTables(); err != nil {
		return nil, errors.Wrap(err, "failed to create tables")
	}

	return store, nil
}

// NewInMemory creates a new in-memory Store instance for testing
func NewInMemory() (*Store, error) {
	return New(":memory:")
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// createTables creates the database schema
func (s *Store) createTables() error {
	_, err := s.db.Exec(CreateTablesSQL)
	return err
}

// Find executes a query and returns matching sections
func (s *Store) Find(ctx context.Context, pred query.Predicate) ([]*model.Section, error) {
	sql, args := query.Compile(pred)
	
	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	// First, collect all sections without loading relations
	var sections []*model.Section
	for rows.Next() {
		section, err := s.scanSection(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan section")
		}
		sections = append(sections, section)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row iteration error")
	}

	// Close the rows before loading relations to avoid SQLite connection issues
	rows.Close()

	// Now load related data for each section
	for _, section := range sections {
		if err := s.loadSectionRelations(ctx, section); err != nil {
			return nil, errors.Wrap(err, "failed to load section relations")
		}
	}

	return sections, nil
}

// GetBySlug retrieves a section by its slug
func (s *Store) GetBySlug(ctx context.Context, slug string) (*model.Section, error) {
	sections, err := s.Find(ctx, query.SlugEquals(slug))
	if err != nil {
		return nil, err
	}
	
	if len(sections) == 0 {
		return nil, errors.New("section not found")
	}
	
	return sections[0], nil
}

// Upsert inserts or updates a section
func (s *Store) Upsert(ctx context.Context, section *model.Section) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Insert or update main section
	err = s.upsertSection(ctx, tx, section)
	if err != nil {
		return errors.Wrap(err, "failed to upsert section")
	}

	// Get the section ID
	var sectionID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM sections WHERE slug = ?", section.Slug).Scan(&sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to get section ID")
	}
	section.ID = sectionID

	// Delete existing relations
	if err := s.deleteRelations(ctx, tx, sectionID); err != nil {
		return errors.Wrap(err, "failed to delete existing relations")
	}

	// Insert new relations
	if err := s.insertRelations(ctx, tx, sectionID, section); err != nil {
		return errors.Wrap(err, "failed to insert relations")
	}

	return tx.Commit()
}

// Delete removes a section by slug
func (s *Store) Delete(ctx context.Context, slug string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Get section ID
	var sectionID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM sections WHERE slug = ?", slug).Scan(&sectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("section not found")
		}
		return errors.Wrap(err, "failed to get section ID")
	}

	// Delete relations (CASCADE should handle this, but let's be explicit)
	if err := s.deleteRelations(ctx, tx, sectionID); err != nil {
		return errors.Wrap(err, "failed to delete relations")
	}

	// Delete main section
	_, err = tx.ExecContext(ctx, "DELETE FROM sections WHERE id = ?", sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete section")
	}

	return tx.Commit()
}

// scanSection scans a database row into a Section struct
func (s *Store) scanSection(rows *sql.Rows) (*model.Section, error) {
	var section model.Section
	var sectionTypeStr string

	err := rows.Scan(
		&section.ID,
		&section.Slug,
		&section.Title,
		&section.Subtitle,
		&section.Short,
		&section.Content,
		&sectionTypeStr,
		&section.IsTopLevel,
		&section.IsTemplate,
		&section.ShowDefault,
		&section.Order,
	)
	if err != nil {
		return nil, err
	}

	section.SectionType, err = model.SectionTypeFromString(sectionTypeStr)
	if err != nil {
		return nil, err
	}

	return &section, nil
}

// loadSectionRelations loads topics, flags, and commands for a section
func (s *Store) loadSectionRelations(ctx context.Context, section *model.Section) error {
	// Initialize slices to avoid nil issues
	section.Topics = []string{}
	section.Flags = []string{}
	section.Commands = []string{}

	// Load topics
	rows, err := s.db.QueryContext(ctx, "SELECT topic FROM section_topics WHERE section_id = ?", section.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			return err
		}
		section.Topics = append(section.Topics, topic)
	}

	// Load flags
	rows, err = s.db.QueryContext(ctx, "SELECT flag FROM section_flags WHERE section_id = ?", section.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var flag string
		if err := rows.Scan(&flag); err != nil {
			return err
		}
		section.Flags = append(section.Flags, flag)
	}

	// Load commands
	rows, err = s.db.QueryContext(ctx, "SELECT command FROM section_commands WHERE section_id = ?", section.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var command string
		if err := rows.Scan(&command); err != nil {
			return err
		}
		section.Commands = append(section.Commands, command)
	}

	return nil
}

// upsertSection inserts or updates the main section record
func (s *Store) upsertSection(ctx context.Context, tx *sql.Tx, section *model.Section) error {
	query := `
		INSERT OR REPLACE INTO sections 
		(slug, title, subtitle, short, content, sectionType, isTopLevel, isTemplate, showDefault, ord)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := tx.ExecContext(ctx, query,
		section.Slug,
		section.Title,
		section.Subtitle,
		section.Short,
		section.Content,
		section.SectionType.String(),
		section.IsTopLevel,
		section.IsTemplate,
		section.ShowDefault,
		section.Order,
	)
	
	return err
}

// deleteRelations removes all relation records for a section
func (s *Store) deleteRelations(ctx context.Context, tx *sql.Tx, sectionID int64) error {
	queries := []string{
		"DELETE FROM section_topics WHERE section_id = ?",
		"DELETE FROM section_flags WHERE section_id = ?",
		"DELETE FROM section_commands WHERE section_id = ?",
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query, sectionID); err != nil {
			return err
		}
	}

	return nil
}

// insertRelations inserts relation records for a section
func (s *Store) insertRelations(ctx context.Context, tx *sql.Tx, sectionID int64, section *model.Section) error {
	// Insert topics
	for _, topic := range section.Topics {
		_, err := tx.ExecContext(ctx, "INSERT INTO section_topics (section_id, topic) VALUES (?, ?)", sectionID, topic)
		if err != nil {
			return err
		}
	}

	// Insert flags
	for _, flag := range section.Flags {
		_, err := tx.ExecContext(ctx, "INSERT INTO section_flags (section_id, flag) VALUES (?, ?)", sectionID, flag)
		if err != nil {
			return err
		}
	}

	// Insert commands
	for _, command := range section.Commands {
		_, err := tx.ExecContext(ctx, "INSERT INTO section_commands (section_id, command) VALUES (?, ?)", sectionID, command)
		if err != nil {
			return err
		}
	}

	return nil
}

// RebuildFTS rebuilds the full-text search index
func (s *Store) RebuildFTS(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO section_fts(section_fts) VALUES('rebuild')")
	return err
}

// Stats returns statistics about the store
func (s *Store) Stats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// Total sections
	var totalSections int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections").Scan(&totalSections)
	if err != nil {
		return nil, err
	}
	stats["total_sections"] = totalSections

	// Sections by type
	rows, err := s.db.QueryContext(ctx, "SELECT sectionType, COUNT(*) FROM sections GROUP BY sectionType")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sectionType string
		var count int
		if err := rows.Scan(&sectionType, &count); err != nil {
			return nil, err
		}
		stats[fmt.Sprintf("type_%s", sectionType)] = count
	}

	// Top level sections
	var topLevel int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE isTopLevel = 1").Scan(&topLevel)
	if err != nil {
		return nil, err
	}
	stats["top_level"] = topLevel

	// Default sections
	var showDefault int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE showDefault = 1").Scan(&showDefault)
	if err != nil {
		return nil, err
	}
	stats["show_default"] = showDefault

	return stats, nil
}
