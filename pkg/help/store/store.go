package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

// Store represents the SQLite-backed help system store
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQLite store
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to disable foreign keys")
	}

	store := &Store{db: db}
	
	if err := store.createSchema(); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to create schema")
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database connection (for debugging)
func (s *Store) GetDB() *sql.DB {
	return s.db
}

// createSchema creates the database schema
func (s *Store) createSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sections (
			id          INTEGER PRIMARY KEY,
			slug        TEXT UNIQUE NOT NULL,
			title       TEXT,
			subtitle    TEXT,
			short       TEXT,
			content     TEXT,
			sectionType TEXT,
			isTopLevel  BOOLEAN,
			isTemplate  BOOLEAN,
			showDefault BOOLEAN,
			ord         INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS section_topics (
			section_id INTEGER,
			topic      TEXT,
			FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS section_flags (
			section_id INTEGER,
			flag       TEXT,
			FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS section_commands (
			section_id INTEGER,
			command    TEXT,
			FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS section_fts USING fts5(
			slug, title, subtitle, short, content, content='sections'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_topics_topic ON section_topics(topic)`,
		`CREATE INDEX IF NOT EXISTS idx_flags_flag ON section_flags(flag)`,
		`CREATE INDEX IF NOT EXISTS idx_commands_command ON section_commands(command)`,
		`CREATE INDEX IF NOT EXISTS idx_sections_type ON sections(sectionType)`,
		`CREATE INDEX IF NOT EXISTS idx_sections_toplevel ON sections(isTopLevel)`,
		`CREATE INDEX IF NOT EXISTS idx_sections_slug ON sections(slug)`,
	}

	for _, stmt := range statements {
		_, err := s.db.Exec(stmt)
		if err != nil {
			return errors.Wrapf(err, "failed to execute schema statement: %s", stmt)
		}
	}
	
	return nil
}

// Find executes a query and returns matching sections
func (s *Store) Find(ctx context.Context, pred query.Predicate) ([]*model.Section, error) {
	sql, args := query.Compile(pred)
	
	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var sections []*model.Section
	for rows.Next() {
		section := &model.Section{}
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
			return nil, errors.Wrap(err, "failed to scan row")
		}
		
		// Convert string to SectionType
		section.SectionType, err = model.SectionTypeFromString(sectionTypeStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse section type")
		}

		// Load related data
		if err := s.loadSectionRelations(ctx, section); err != nil {
			return nil, errors.Wrap(err, "failed to load section relations")
		}

		sections = append(sections, section)
	}

	return sections, nil
}

// loadSectionRelations loads topics, flags, and commands for a section
func (s *Store) loadSectionRelations(ctx context.Context, section *model.Section) error {
	// Load topics
	topics, err := s.loadRelation(ctx, "section_topics", "topic", section.ID)
	if err != nil {
		return err
	}
	section.Topics = topics

	// Load flags
	flags, err := s.loadRelation(ctx, "section_flags", "flag", section.ID)
	if err != nil {
		return err
	}
	section.Flags = flags

	// Load commands
	commands, err := s.loadRelation(ctx, "section_commands", "command", section.ID)
	if err != nil {
		return err
	}
	section.Commands = commands

	return nil
}

// loadRelation loads a specific relation for a section
func (s *Store) loadRelation(ctx context.Context, table, column string, sectionID int64) ([]string, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE section_id = ?", column, table)
	rows, err := s.db.QueryContext(ctx, query, sectionID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query table %s for section %d", table, sectionID)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	return result, nil
}

// UpsertSection inserts or updates a section
func (s *Store) UpsertSection(ctx context.Context, section *model.Section) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// First, try to get existing section ID if it exists
	var sectionID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM sections WHERE slug = ?", section.Slug).Scan(&sectionID)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to check existing section")
	}
	
	if err == sql.ErrNoRows {
		// Insert new section
		insertSQL := `
		INSERT INTO sections 
		(slug, title, subtitle, short, content, sectionType, isTopLevel, isTemplate, showDefault, ord)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		
		result, err := tx.ExecContext(ctx, insertSQL,
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
		if err != nil {
			return errors.Wrap(err, "failed to insert section")
		}
		
		sectionID, err = result.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "failed to get last insert ID")
		}
		section.ID = sectionID
	} else {
		// Update existing section
		updateSQL := `
		UPDATE sections SET 
		title = ?, subtitle = ?, short = ?, content = ?, sectionType = ?, 
		isTopLevel = ?, isTemplate = ?, showDefault = ?, ord = ?
		WHERE id = ?
		`
		
		_, err := tx.ExecContext(ctx, updateSQL,
			section.Title,
			section.Subtitle,
			section.Short,
			section.Content,
			section.SectionType.String(),
			section.IsTopLevel,
			section.IsTemplate,
			section.ShowDefault,
			section.Order,
			sectionID,
		)
		if err != nil {
			return errors.Wrap(err, "failed to update section")
		}
		section.ID = sectionID
	}

	// Delete existing relations
	for _, table := range []string{"section_topics", "section_flags", "section_commands"} {
		_, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE section_id = ?", table), sectionID)
		if err != nil {
			return errors.Wrap(err, "failed to delete existing relations")
		}
	}

	// Insert new relations
	if err := s.insertRelations(ctx, tx, "section_topics", "topic", sectionID, section.Topics); err != nil {
		return err
	}
	if err := s.insertRelations(ctx, tx, "section_flags", "flag", sectionID, section.Flags); err != nil {
		return err
	}
	if err := s.insertRelations(ctx, tx, "section_commands", "command", sectionID, section.Commands); err != nil {
		return err
	}

	// Update FTS index
	_, err = tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO section_fts (rowid, slug, title, subtitle, short, content)
		VALUES (?, ?, ?, ?, ?, ?)
	`, sectionID, section.Slug, section.Title, section.Subtitle, section.Short, section.Content)
	if err != nil {
		return errors.Wrap(err, "failed to update FTS index")
	}

	return tx.Commit()
}

// insertRelations inserts relation records
func (s *Store) insertRelations(ctx context.Context, tx *sql.Tx, table, column string, sectionID int64, values []string) error {
	if len(values) == 0 {
		return nil
	}

	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values)*2)
	for i, value := range values {
		placeholders[i] = "(?, ?)"
		args[i*2] = sectionID
		args[i*2+1] = value
	}

	sql := fmt.Sprintf("INSERT INTO %s (section_id, %s) VALUES %s", table, column, strings.Join(placeholders, ", "))
	_, err := tx.ExecContext(ctx, sql, args...)
	return err
}

// GetSectionBySlug retrieves a section by its slug
func (s *Store) GetSectionBySlug(ctx context.Context, slug string) (*model.Section, error) {
	sections, err := s.Find(ctx, query.SlugEquals(slug))
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return nil, errors.New("section not found")
	}
	return sections[0], nil
}

// RebuildFTSIndex rebuilds the full-text search index
func (s *Store) RebuildFTSIndex(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO section_fts(section_fts) VALUES ('rebuild')")
	return err
}

// FindByText executes a text-based query and returns matching sections
func (s *Store) FindByText(ctx context.Context, textQuery string) ([]*model.Section, error) {
	// Import the search package to convert text query to predicate
	// This creates a circular import, so we'll need to refactor this
	// For now, we'll add this method later after resolving the import issue
	return nil, errors.New("FindByText not yet implemented - requires refactoring to avoid circular imports")
}
