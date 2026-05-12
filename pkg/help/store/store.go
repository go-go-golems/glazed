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

// ErrSectionNotFound is returned by GetBySlug and GetByID when no matching
// section exists in the store. Callers should use errors.Is to test for it.
var ErrSectionNotFound = errors.New("section not found")

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
	// A plain :memory: SQLite database is per connection. The store is often used
	// in-memory by the help server, so keep one connection to avoid requests
	// seeing a fresh empty database without the sections table.
	db.SetMaxOpenConns(1)

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
	exists, err := s.tableExists("sections")
	if err != nil {
		return errors.Wrap(err, "failed to inspect sections table")
	}

	if exists {
		if err := s.migrateSectionsTable(); err != nil {
			return err
		}
	} else if err := s.createSectionsTable(); err != nil {
		return err
	}

	// Create indexes for better performance. The composite unique index is the
	// package-aware identity used by multi-package serving.
	indexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_sections_package_version_slug ON sections(package_name, package_version, slug);",
		"CREATE INDEX IF NOT EXISTS idx_sections_package ON sections(package_name);",
		"CREATE INDEX IF NOT EXISTS idx_sections_package_version ON sections(package_name, package_version);",
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

func (s *Store) tableExists(name string) (bool, error) {
	row := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) createSectionsTable() error {
	sectionsTable := `
		CREATE TABLE IF NOT EXISTS sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			package_name TEXT NOT NULL DEFAULT '',
			package_version TEXT NOT NULL DEFAULT '',
			slug TEXT NOT NULL,
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(package_name, package_version, slug)
		);
	`

	if _, err := s.db.Exec(sectionsTable); err != nil {
		return errors.Wrap(err, "failed to create sections table")
	}
	return nil
}

func (s *Store) migrateSectionsTable() error {
	columns, err := s.sectionColumns()
	if err != nil {
		return errors.Wrap(err, "failed to inspect sections columns")
	}
	if columns["package_name"] && columns["package_version"] && !s.hasLegacySlugUniqueIndex() {
		return nil
	}

	legacyName := fmt.Sprintf("sections_legacy_%d", time.Now().UnixNano())
	tx, err := s.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin sections migration")
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("ALTER TABLE sections RENAME TO " + legacyName); err != nil {
		return errors.Wrap(err, "failed to rename legacy sections table")
	}
	if _, err := tx.Exec(`
		CREATE TABLE sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			package_name TEXT NOT NULL DEFAULT '',
			package_version TEXT NOT NULL DEFAULT '',
			slug TEXT NOT NULL,
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(package_name, package_version, slug)
		);
	`); err != nil {
		return errors.Wrap(err, "failed to create migrated sections table")
	}

	packageNameExpr := "''"
	packageVersionExpr := "''"
	if columns["package_name"] {
		packageNameExpr = "COALESCE(package_name, '')"
	}
	if columns["package_version"] {
		packageVersionExpr = "COALESCE(package_version, '')"
	}

	// #nosec G201 -- package column expressions and legacy table name are generated internally during migration, not from user input.
	copySQL := fmt.Sprintf(`
		INSERT INTO sections (
			id, package_name, package_version, slug, section_type, title, sub_title,
			short, content, topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num, created_at, updated_at
		)
		SELECT id, %s, %s, slug, section_type, title, sub_title, short, content,
			topics, flags, commands, is_top_level, is_template, show_per_default,
			order_num, created_at, updated_at
		FROM %s
	`, packageNameExpr, packageVersionExpr, legacyName)
	if _, err := tx.Exec(copySQL); err != nil {
		return errors.Wrap(err, "failed to copy legacy sections rows")
	}
	if _, err := tx.Exec("DROP TABLE " + legacyName); err != nil {
		return errors.Wrap(err, "failed to drop legacy sections table")
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit sections migration")
	}
	return nil
}

func (s *Store) sectionColumns() (map[string]bool, error) {
	rows, err := s.db.Query("PRAGMA table_info(sections)")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

func (s *Store) hasLegacySlugUniqueIndex() bool {
	rows, err := s.db.Query("PRAGMA index_list(sections)")
	if err != nil {
		return false
	}

	var uniqueIndexes []string
	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil || unique == 0 {
			continue
		}
		uniqueIndexes = append(uniqueIndexes, name)
	}
	_ = rows.Close()

	for _, name := range uniqueIndexes {
		infoRows, err := s.db.Query("PRAGMA index_info(" + name + ")")
		if err != nil {
			continue
		}
		var cols []string
		for infoRows.Next() {
			var seqno, cid int
			var col string
			if err := infoRows.Scan(&seqno, &cid, &col); err == nil {
				cols = append(cols, col)
			}
		}
		_ = infoRows.Close()
		if len(cols) == 1 && cols[0] == "slug" {
			return true
		}
	}
	return false
}

// Insert adds a new section to the store
func (s *Store) Insert(ctx context.Context, section *model.Section) error {
	if err := section.Validate(); err != nil {
		return errors.Wrap(err, "section validation failed")
	}

	query := `
		INSERT INTO sections (
			package_name, package_version, slug, section_type, title, sub_title,
			short, content, topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		section.PackageName,
		section.PackageVersion,
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
			package_name = ?, package_version = ?, slug = ?, section_type = ?,
			title = ?, sub_title = ?, short = ?, content = ?, topics = ?, flags = ?,
			commands = ?, is_top_level = ?, is_template = ?, show_per_default = ?,
			order_num = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		section.PackageName,
		section.PackageVersion,
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

// Upsert inserts or updates a section based on its package, version, and slug.
func (s *Store) Upsert(ctx context.Context, section *model.Section) error {
	if err := section.Validate(); err != nil {
		return errors.Wrap(err, "section validation failed")
	}

	query := `
		INSERT INTO sections (
			package_name, package_version, slug, section_type, title, sub_title,
			short, content, topics, flags, commands, is_top_level, is_template,
			show_per_default, order_num
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(package_name, package_version, slug) DO UPDATE SET
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
		section.PackageName,
		section.PackageVersion,
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
		SELECT id, package_name, package_version, slug, section_type, title,
			sub_title, short, content, topics, flags, commands, is_top_level,
			is_template, show_per_default, order_num, created_at, updated_at
		FROM sections WHERE slug = ?
		ORDER BY package_name ASC, package_version ASC
		LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, query, slug)
	return s.scanSection(row)
}

// GetByPackageSlug retrieves a section by package, optional version, and slug.
func (s *Store) GetByPackageSlug(ctx context.Context, packageName, packageVersion, slug string) (*model.Section, error) {
	query := `
		SELECT id, package_name, package_version, slug, section_type, title,
			sub_title, short, content, topics, flags, commands, is_top_level,
			is_template, show_per_default, order_num, created_at, updated_at
		FROM sections WHERE package_name = ? AND package_version = ? AND slug = ?
	`

	row := s.db.QueryRowContext(ctx, query, packageName, packageVersion, slug)
	return s.scanSection(row)
}

// GetByID retrieves a section by its ID
func (s *Store) GetByID(ctx context.Context, id int64) (*model.Section, error) {
	query := `
		SELECT id, package_name, package_version, slug, section_type, title,
			sub_title, short, content, topics, flags, commands, is_top_level,
			is_template, show_per_default, order_num, created_at, updated_at
		FROM sections WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanSection(row)
}

// List retrieves all sections with optional ordering
func (s *Store) List(ctx context.Context, orderBy string) ([]*model.Section, error) {
	query := `
		SELECT id, package_name, package_version, slug, section_type, title,
			sub_title, short, content, topics, flags, commands, is_top_level,
			is_template, show_per_default, order_num, created_at, updated_at
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
			&section.PackageName,
			&section.PackageVersion,
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
			&section.PackageName,
			&section.PackageVersion,
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
			return nil, ErrSectionNotFound
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

// PackageInfo summarizes one package/version group in the store.
type PackageInfo struct {
	Name         string
	Version      string
	SectionCount int
}

// ListPackages returns package/version groups with section counts.
func (s *Store) ListPackages(ctx context.Context) ([]PackageInfo, error) {
	query := `
		SELECT package_name, package_version, COUNT(*)
		FROM sections
		GROUP BY package_name, package_version
		ORDER BY package_name ASC, package_version DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list packages")
	}
	defer func() { _ = rows.Close() }()

	var packages []PackageInfo
	for rows.Next() {
		var pkg PackageInfo
		if err := rows.Scan(&pkg.Name, &pkg.Version, &pkg.SectionCount); err != nil {
			return nil, errors.Wrap(err, "failed to scan package")
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

// SetDefaultPackage assigns package metadata to sections that do not have it yet.
// It updates all rows where package_name is empty, setting them to the given
// packageName and packageVersion.
//
// This is necessary because sections loaded from embedded markdown files (via
// LoadSectionsFromFS) get package_name = "". The SPA's package filter queries
// by name, so sections without a package name won't appear in the sidebar.
// NewServeHandler calls this automatically with package "default", so most
// callers do not need to invoke it directly.
func (s *Store) SetDefaultPackage(ctx context.Context, packageName, packageVersion string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE sections
		SET package_name = ?, package_version = ?
		WHERE COALESCE(package_name, '') = ''
	`, packageName, packageVersion)
	if err != nil {
		return errors.Wrap(err, "failed to set default package")
	}
	return nil
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
