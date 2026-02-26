package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/help/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Store represents the SQLite-backed help system storage
type Store struct {
	db *sql.DB
}

// New creates a new SQLite store
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	store := &Store{db: db}
	if err := store.createTables(); err != nil {
		return nil, errors.Wrap(err, "failed to create tables")
	}

	return store, nil
}

// NewInMemory creates a new in-memory SQLite store
func NewInMemory() (*Store, error) {
	return New(":memory:")
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables for the help system
func (s *Store) createTables() error {
	// Create sections table
	sectionsTable := `
		CREATE TABLE IF NOT EXISTS sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			section_type INTEGER NOT NULL,
			title TEXT NOT NULL,
			sub_title TEXT,
			short TEXT,
			content TEXT,
			topics TEXT,
			flags TEXT,
			commands TEXT,
			is_top_level BOOLEAN DEFAULT FALSE,
			is_template BOOLEAN DEFAULT FALSE,
			show_per_default BOOLEAN DEFAULT FALSE,
			order_num INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := s.db.Exec(sectionsTable); err != nil {
		return errors.Wrap(err, "failed to create sections table")
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_sections_slug ON sections(slug);",
		"CREATE INDEX IF NOT EXISTS idx_sections_type ON sections(section_type);",
		"CREATE INDEX IF NOT EXISTS idx_sections_top_level ON sections(is_top_level);",
		"CREATE INDEX IF NOT EXISTS idx_sections_show_default ON sections(show_per_default);",
		"CREATE INDEX IF NOT EXISTS idx_sections_order ON sections(order_num);",
	}

	for _, index := range indexes {
		if _, err := s.db.Exec(index); err != nil {
			return errors.Wrap(err, "failed to create index")
		}
	}

	// Create FTS tables if enabled (build tag dependent)
	if err := s.createFTSTables(); err != nil {
		return err
	}

	return nil
}

// Insert adds a new section to the store
func (s *Store) Insert(ctx context.Context, section *model.Section) error {
	if err := section.Validate(); err != nil {
		return errors.Wrap(err, "section validation failed")
	}

	query := `
		INSERT INTO sections (
			slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		section.Slug,
		section.SectionType.ToInt(),
		section.Title,
		section.SubTitle,
		section.Short,
		section.Content,
		section.TopicsAsString(),
		section.FlagsAsString(),
		section.CommandsAsString(),
		section.IsTopLevel,
		section.IsTemplate,
		section.ShowPerDefault,
		section.Order,
	)

	if err != nil {
		return errors.Wrap(err, "failed to insert section")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to get last insert id")
	}

	section.ID = id
	return nil
}

// Update modifies an existing section in the store
func (s *Store) Update(ctx context.Context, section *model.Section) error {
	if err := section.Validate(); err != nil {
		return errors.Wrap(err, "section validation failed")
	}

	query := `
		UPDATE sections SET
			slug = ?, section_type = ?, title = ?, sub_title = ?, short = ?, content = ?,
			topics = ?, flags = ?, commands = ?, is_top_level = ?, is_template = ?,
			show_per_default = ?, order_num = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		section.Slug,
		section.SectionType.ToInt(),
		section.Title,
		section.SubTitle,
		section.Short,
		section.Content,
		section.TopicsAsString(),
		section.FlagsAsString(),
		section.CommandsAsString(),
		section.IsTopLevel,
		section.IsTemplate,
		section.ShowPerDefault,
		section.Order,
		section.ID,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update section")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("no section found with given ID")
	}

	return nil
}

// Upsert inserts or updates a section based on its slug
func (s *Store) Upsert(ctx context.Context, section *model.Section) error {
	if err := section.Validate(); err != nil {
		return errors.Wrap(err, "section validation failed")
	}

	query := `
		INSERT INTO sections (
			slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			section_type = excluded.section_type,
			title = excluded.title,
			sub_title = excluded.sub_title,
			short = excluded.short,
			content = excluded.content,
			topics = excluded.topics,
			flags = excluded.flags,
			commands = excluded.commands,
			is_top_level = excluded.is_top_level,
			is_template = excluded.is_template,
			show_per_default = excluded.show_per_default,
			order_num = excluded.order_num,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := s.db.ExecContext(ctx, query,
		section.Slug,
		section.SectionType.ToInt(),
		section.Title,
		section.SubTitle,
		section.Short,
		section.Content,
		section.TopicsAsString(),
		section.FlagsAsString(),
		section.CommandsAsString(),
		section.IsTopLevel,
		section.IsTemplate,
		section.ShowPerDefault,
		section.Order,
	)

	if err != nil {
		return errors.Wrap(err, "failed to upsert section")
	}

	if section.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "failed to get last insert id")
		}
		section.ID = id
	}

	return nil
}

// Delete removes a section from the store
func (s *Store) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM sections WHERE id = ?"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete section")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("no section found with given ID")
	}

	return nil
}

// GetBySlug retrieves a section by its slug
func (s *Store) GetBySlug(ctx context.Context, slug string) (*model.Section, error) {
	query := `
		SELECT id, slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num, created_at, updated_at
		FROM sections WHERE slug = ?
	`

	row := s.db.QueryRowContext(ctx, query, slug)
	return s.scanSection(row)
}

// GetByID retrieves a section by its ID
func (s *Store) GetByID(ctx context.Context, id int64) (*model.Section, error) {
	query := `
		SELECT id, slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num, created_at, updated_at
		FROM sections WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanSection(row)
}

// List retrieves all sections with optional ordering
func (s *Store) List(ctx context.Context, orderBy string) ([]*model.Section, error) {
	query := `
		SELECT id, slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num, created_at, updated_at
		FROM sections
	`

	orderByClause, err := sanitizeOrderByClause(orderBy)
	if err != nil {
		return nil, errors.Wrap(err, "invalid order by clause")
	}
	if orderByClause != "" {
		// #nosec G202 -- orderByClause is built only by sanitizeOrderByClause from a fixed column/direction allow-list.
		query += orderByClause
	}

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sections")
	}
	defer func() {
		_ = rows.Close()
	}()

	var sections []*model.Section
	for rows.Next() {
		section, err := s.scanSection(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan section")
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func sanitizeOrderByClause(orderBy string) (string, error) {
	trimmed := strings.TrimSpace(orderBy)
	if trimmed == "" {
		return "", nil
	}

	parts := strings.Fields(trimmed)
	if len(parts) == 0 || len(parts) > 2 {
		return "", fmt.Errorf("unsupported order by format")
	}

	column := strings.ToLower(parts[0])
	column = strings.TrimPrefix(column, "s.")
	allowedColumns := map[string]string{
		"id":         "id",
		"slug":       "slug",
		"title":      "title",
		"order_num":  "order_num",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	resolvedColumn, ok := allowedColumns[column]
	if !ok {
		return "", fmt.Errorf("unsupported order by column %q", parts[0])
	}

	direction := "ASC"
	if len(parts) == 2 {
		d := strings.ToUpper(parts[1])
		if d != "ASC" && d != "DESC" {
			return "", fmt.Errorf("unsupported order by direction %q", parts[1])
		}
		direction = d
	}

	return " ORDER BY " + resolvedColumn + " " + direction, nil
}

// scanSection scans a database row into a Section struct
func (s *Store) scanSection(scanner interface{}) (*model.Section, error) {
	var section model.Section
	var sectionType int
	var topics, flags, commands string
	var createdAt, updatedAt time.Time

	var err error
	switch v := scanner.(type) {
	case *sql.Row:
		err = v.Scan(
			&section.ID,
			&section.Slug,
			&sectionType,
			&section.Title,
			&section.SubTitle,
			&section.Short,
			&section.Content,
			&topics,
			&flags,
			&commands,
			&section.IsTopLevel,
			&section.IsTemplate,
			&section.ShowPerDefault,
			&section.Order,
			&createdAt,
			&updatedAt,
		)
	case *sql.Rows:
		err = v.Scan(
			&section.ID,
			&section.Slug,
			&sectionType,
			&section.Title,
			&section.SubTitle,
			&section.Short,
			&section.Content,
			&topics,
			&flags,
			&commands,
			&section.IsTopLevel,
			&section.IsTemplate,
			&section.ShowPerDefault,
			&section.Order,
			&createdAt,
			&updatedAt,
		)
	default:
		return nil, errors.New("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("section not found")
		}
		return nil, errors.Wrap(err, "failed to scan section")
	}

	section.SectionType = model.SectionType(sectionType)
	section.SetTopicsFromString(topics)
	section.SetFlagsFromString(flags)
	section.SetCommandsFromString(commands)
	section.CreatedAt = createdAt.Format(time.RFC3339)
	section.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &section, nil
}

// Count returns the total number of sections
func (s *Store) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM sections"
	row := s.db.QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count sections")
	}

	return count, nil
}

// Clear removes all sections from the store
func (s *Store) Clear(ctx context.Context) error {
	query := "DELETE FROM sections"
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to clear sections")
	}

	log.Debug().Msg("Cleared all sections from store")
	return nil
}
