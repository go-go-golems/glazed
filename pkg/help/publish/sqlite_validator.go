package publish

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// SQLiteValidationOptions controls validation of a Glazed help SQLite export.
type SQLiteValidationOptions struct {
	PackageName string
	Version     string
}

// SQLiteValidationResult summarizes a validated help database.
type SQLiteValidationResult struct {
	Path         string   `json:"path"`
	PackageName  string   `json:"packageName,omitempty"`
	Version      string   `json:"version,omitempty"`
	SectionCount int      `json:"sectionCount"`
	SlugCount    int      `json:"slugCount"`
	Warnings     []string `json:"warnings,omitempty"`
}

// ValidateSQLiteHelpDB opens path read-only and verifies that it looks like a
// Glazed help SQLite export that is safe to publish for the requested package
// and version.
func ValidateSQLiteHelpDB(ctx context.Context, path string, opts SQLiteValidationOptions) (*SQLiteValidationResult, error) {
	if path == "" {
		return nil, errors.New("file path must not be empty")
	}
	if opts.PackageName != "" {
		if err := ValidatePackageName(opts.PackageName); err != nil {
			return nil, err
		}
	}
	if opts.Version != "" {
		if err := ValidateVersion(opts.Version); err != nil {
			return nil, err
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "resolve database path")
	}

	db, err := sql.Open("sqlite3", readOnlySQLiteDSN(absPath))
	if err != nil {
		return nil, errors.Wrap(err, "open SQLite database")
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "open SQLite database read-only")
	}

	if err := requireSectionsTable(ctx, db); err != nil {
		return nil, err
	}
	if err := requireSectionColumns(ctx, db); err != nil {
		return nil, err
	}

	result := &SQLiteValidationResult{Path: absPath, PackageName: opts.PackageName, Version: opts.Version}
	sectionCount, err := countSections(ctx, db)
	if err != nil {
		return nil, err
	}
	if sectionCount == 0 {
		return nil, errors.New("help database contains no sections")
	}
	result.SectionCount = sectionCount

	slugCount, err := countDistinctSlugs(ctx, db)
	if err != nil {
		return nil, err
	}
	result.SlugCount = slugCount

	if err := rejectEmptySlugs(ctx, db); err != nil {
		return nil, err
	}
	if err := rejectDuplicateSlugs(ctx, db); err != nil {
		return nil, err
	}
	warnings, err := packageMetadataWarnings(ctx, db, opts)
	if err != nil {
		return nil, err
	}
	result.Warnings = warnings

	return result, nil
}

func readOnlySQLiteDSN(absPath string) string {
	u := url.URL{Scheme: "file", Path: absPath}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("_query_only", "true")
	u.RawQuery = q.Encode()
	return u.String()
}

func requireSectionsTable(ctx context.Context, db *sql.DB) error {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'sections'`).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "inspect SQLite schema")
	}
	if count == 0 {
		return errors.New("missing required sections table")
	}
	return nil
}

func requireSectionColumns(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(sections)`)
	if err != nil {
		return errors.Wrap(err, "inspect sections columns")
	}
	defer func() { _ = rows.Close() }()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return errors.Wrap(err, "scan sections column")
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate sections columns")
	}

	for _, required := range []string{"slug", "title"} {
		if !columns[required] {
			return fmt.Errorf("sections table missing required column %q", required)
		}
	}
	return nil
}

func countSections(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sections`).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "count sections")
	}
	return count, nil
}

func countDistinctSlugs(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT slug) FROM sections`).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "count section slugs")
	}
	return count, nil
}

func rejectEmptySlugs(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sections WHERE slug IS NULL OR TRIM(slug) = ''`).Scan(&count); err != nil {
		return errors.Wrap(err, "count empty slugs")
	}
	if count > 0 {
		return fmt.Errorf("help database contains %d section(s) with empty slugs", count)
	}
	return nil
}

func rejectDuplicateSlugs(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `
		SELECT slug, COUNT(*)
		FROM sections
		GROUP BY slug
		HAVING COUNT(*) > 1
		ORDER BY slug
		LIMIT 5
	`)
	if err != nil {
		return errors.Wrap(err, "find duplicate slugs")
	}
	defer func() { _ = rows.Close() }()

	var duplicates []string
	for rows.Next() {
		var slug string
		var count int
		if err := rows.Scan(&slug, &count); err != nil {
			return errors.Wrap(err, "scan duplicate slug")
		}
		duplicates = append(duplicates, fmt.Sprintf("%s (%d)", slug, count))
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "iterate duplicate slugs")
	}
	if len(duplicates) > 0 {
		return fmt.Errorf("help database contains duplicate slugs: %s", strings.Join(duplicates, ", "))
	}
	return nil
}

func packageMetadataWarnings(ctx context.Context, db *sql.DB, opts SQLiteValidationOptions) ([]string, error) {
	columns, err := tableColumns(ctx, db, "sections")
	if err != nil {
		return nil, err
	}
	var warnings []string
	if columns["package_name"] && opts.PackageName != "" {
		values, err := distinctNonEmptyValues(ctx, db, "package_name")
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if value != opts.PackageName {
				warnings = append(warnings, fmt.Sprintf("database contains package_name %q; publisher will serve it as %q", value, opts.PackageName))
			}
		}
	}
	if columns["package_version"] && opts.Version != "" {
		values, err := distinctNonEmptyValues(ctx, db, "package_version")
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if value != opts.Version {
				warnings = append(warnings, fmt.Sprintf("database contains package_version %q; publisher will serve it as %q", value, opts.Version))
			}
		}
	}
	return warnings, nil
}

func tableColumns(ctx context.Context, db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return nil, errors.Wrapf(err, "inspect %s columns", table)
	}
	defer func() { _ = rows.Close() }()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return nil, errors.Wrapf(err, "scan %s column", table)
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

func distinctNonEmptyValues(ctx context.Context, db *sql.DB, column string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT DISTINCT `+column+` FROM sections WHERE COALESCE(`+column+`, '') != '' ORDER BY `+column)
	if err != nil {
		return nil, errors.Wrapf(err, "query distinct %s values", column)
	}
	defer func() { _ = rows.Close() }()
	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, errors.Wrapf(err, "scan %s value", column)
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
