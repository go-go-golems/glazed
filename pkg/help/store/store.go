package store

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// Store represents a SQLite-backed help section store
type Store struct {
	db *sql.DB
}

// New creates a new Store with the given database connection
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// NewInMemory creates a new in-memory SQLite store
func NewInMemory() (*Store, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open in-memory database")
	}
	
	// For in-memory databases, we need exactly one connection to maintain state
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(-1) // No max lifetime
	
	store := New(db)
	if err := store.Init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize store")
	}
	
	return store, nil
}

// NewFile creates a new file-based SQLite store
func NewFile(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database file")
	}
	
	store := New(db)
	if err := store.Init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize store")
	}
	
	return store, nil
}

// Init initializes the database schema
func (s *Store) Init() error {
	if _, err := s.db.Exec(schema); err != nil {
		return errors.Wrap(err, "failed to create database schema")
	}
	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Find finds sections matching the given predicate
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
		
		section.SectionType, err = model.SectionTypeFromString(sectionTypeStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse section type")
		}
		
		// Initialize slices to avoid nil pointer issues
		section.Topics = make([]string, 0)
		section.Flags = make([]string, 0)
		section.Commands = make([]string, 0)
		
		sections = append(sections, section)
	}
	
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row iteration error")
	}
	
	// Load relations for all sections
	if err := s.loadSectionRelationsForAll(ctx, sections); err != nil {
		return nil, errors.Wrap(err, "failed to load section relations")
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
	
	// Insert or replace section
	_, err = tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO sections (slug, title, subtitle, short, content, sectionType, isTopLevel, isTemplate, showDefault, ord)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, section.Slug, section.Title, section.Subtitle, section.Short, section.Content,
		section.SectionType.String(), section.IsTopLevel, section.IsTemplate, section.ShowDefault, section.Order)
	if err != nil {
		return errors.Wrap(err, "failed to upsert section")
	}
	
	// Get the section ID
	var sectionID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM sections WHERE slug = ?", section.Slug).Scan(&sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to get section ID")
	}
	
	section.ID = int(sectionID)
	
	// Delete existing relationships
	_, err = tx.ExecContext(ctx, "DELETE FROM section_topics WHERE section_id = ?", sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing topics")
	}
	
	_, err = tx.ExecContext(ctx, "DELETE FROM section_flags WHERE section_id = ?", sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing flags")
	}
	
	_, err = tx.ExecContext(ctx, "DELETE FROM section_commands WHERE section_id = ?", sectionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing commands")
	}
	
	// Insert new relationships
	for _, topic := range section.Topics {
		_, err = tx.ExecContext(ctx, "INSERT INTO section_topics (section_id, topic) VALUES (?, ?)", sectionID, topic)
		if err != nil {
			return errors.Wrap(err, "failed to insert topic")
		}
	}
	
	for _, flag := range section.Flags {
		_, err = tx.ExecContext(ctx, "INSERT INTO section_flags (section_id, flag) VALUES (?, ?)", sectionID, flag)
		if err != nil {
			return errors.Wrap(err, "failed to insert flag")
		}
	}
	
	for _, command := range section.Commands {
		_, err = tx.ExecContext(ctx, "INSERT INTO section_commands (section_id, command) VALUES (?, ?)", sectionID, command)
		if err != nil {
			return errors.Wrap(err, "failed to insert command")
		}
	}
	
	return tx.Commit()
}

// LoadSectionsFromFS loads sections from a filesystem
func (s *Store) LoadSectionsFromFS(ctx context.Context, fsys fs.FS, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		log.Warn().Err(err).Str("dir", dir).Msg("Failed to read directory")
		return nil
	}
	
	for _, entry := range entries {
		filePath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err = s.LoadSectionsFromFS(ctx, fsys, filePath)
			if err != nil {
				log.Warn().Err(err).Str("dir", filePath).Msg("Failed to load sections from directory")
				continue
			}
		} else {
			// Skip non-markdown files and readme files
			if !strings.HasSuffix(entry.Name(), ".md") || strings.ToLower(entry.Name()) == "readme.md" {
				continue
			}
			
			b, err := fs.ReadFile(fsys, filePath)
			if err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to read file")
				continue
			}
			
			section, err := model.LoadSectionFromMarkdown(b)
			if err != nil {
				log.Debug().Err(err).Str("file", filePath).Msg("Failed to load section from file")
				continue
			}
			
			if err := s.Upsert(ctx, section); err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to upsert section")
				continue
			}
		}
	}
	
	return nil
}

// loadSectionRelations loads the topics, flags, and commands for a section
func (s *Store) loadSectionRelations(ctx context.Context, section *model.Section) error {
	// Load topics
	topicRows, err := s.db.QueryContext(ctx, "SELECT topic FROM section_topics WHERE section_id = ?", section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to query topics")
	}
	defer topicRows.Close()
	
	for topicRows.Next() {
		var topic string
		if err := topicRows.Scan(&topic); err != nil {
			return errors.Wrap(err, "failed to scan topic")
		}
		section.Topics = append(section.Topics, topic)
	}
	
	// Load flags
	flagRows, err := s.db.QueryContext(ctx, "SELECT flag FROM section_flags WHERE section_id = ?", section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to query flags")
	}
	defer flagRows.Close()
	
	for flagRows.Next() {
		var flag string
		if err := flagRows.Scan(&flag); err != nil {
			return errors.Wrap(err, "failed to scan flag")
		}
		section.Flags = append(section.Flags, flag)
	}
	
	// Load commands
	cmdRows, err := s.db.QueryContext(ctx, "SELECT command FROM section_commands WHERE section_id = ?", section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to query commands")
	}
	defer cmdRows.Close()
	
	for cmdRows.Next() {
		var command string
		if err := cmdRows.Scan(&command); err != nil {
			return errors.Wrap(err, "failed to scan command")
		}
		section.Commands = append(section.Commands, command)
	}
	
	return nil
}

// loadSectionRelationsForAll loads relations for multiple sections in batch
func (s *Store) loadSectionRelationsForAll(ctx context.Context, sections []*model.Section) error {
	if len(sections) == 0 {
		return nil
	}
	
	// Create a map for quick section lookup
	sectionMap := make(map[int]*model.Section)
	for _, section := range sections {
		sectionMap[section.ID] = section
	}
	
	// Load all topics for these sections
	topicRows, err := s.db.QueryContext(ctx, "SELECT section_id, topic FROM section_topics WHERE section_id IN ("+s.placeholders(len(sections))+")", s.sectionIDs(sections)...)
	if err != nil {
		return errors.Wrap(err, "failed to query topics")
	}
	defer topicRows.Close()
	
	for topicRows.Next() {
		var sectionID int
		var topic string
		if err := topicRows.Scan(&sectionID, &topic); err != nil {
			return errors.Wrap(err, "failed to scan topic")
		}
		if section, ok := sectionMap[sectionID]; ok {
			section.Topics = append(section.Topics, topic)
		}
	}
	
	// Load all flags for these sections
	flagRows, err := s.db.QueryContext(ctx, "SELECT section_id, flag FROM section_flags WHERE section_id IN ("+s.placeholders(len(sections))+")", s.sectionIDs(sections)...)
	if err != nil {
		return errors.Wrap(err, "failed to query flags")
	}
	defer flagRows.Close()
	
	for flagRows.Next() {
		var sectionID int
		var flag string
		if err := flagRows.Scan(&sectionID, &flag); err != nil {
			return errors.Wrap(err, "failed to scan flag")
		}
		if section, ok := sectionMap[sectionID]; ok {
			section.Flags = append(section.Flags, flag)
		}
	}
	
	// Load all commands for these sections
	cmdRows, err := s.db.QueryContext(ctx, "SELECT section_id, command FROM section_commands WHERE section_id IN ("+s.placeholders(len(sections))+")", s.sectionIDs(sections)...)
	if err != nil {
		return errors.Wrap(err, "failed to query commands")
	}
	defer cmdRows.Close()
	
	for cmdRows.Next() {
		var sectionID int
		var command string
		if err := cmdRows.Scan(&sectionID, &command); err != nil {
			return errors.Wrap(err, "failed to scan command")
		}
		if section, ok := sectionMap[sectionID]; ok {
			section.Commands = append(section.Commands, command)
		}
	}
	
	return nil
}

// placeholders generates SQL placeholders for IN clauses
func (s *Store) placeholders(count int) string {
	if count == 0 {
		return ""
	}
	result := "?"
	for i := 1; i < count; i++ {
		result += ",?"
	}
	return result
}

// sectionIDs extracts section IDs as interfaces for SQL queries
func (s *Store) sectionIDs(sections []*model.Section) []interface{} {
	ids := make([]interface{}, len(sections))
	for i, section := range sections {
		ids[i] = section.ID
	}
	return ids
}
