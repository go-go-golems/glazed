package store

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// LoadSectionsFromFS loads sections from a filesystem and stores them in the database
func (s *Store) LoadSectionsFromFS(ctx context.Context, f fs.FS, dir string) error {
	return s.loadSectionsFromFS(ctx, f, dir)
}

// loadSectionsFromFS recursively loads sections from a directory
func (s *Store) loadSectionsFromFS(ctx context.Context, f fs.FS, dir string) error {
	entries, err := fs.ReadDir(f, dir)
	if err != nil {
		log.Warn().Err(err).Str("dir", dir).Msg("Failed to read directory")
		return nil
	}

	for _, entry := range entries {
		filePath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err = s.loadSectionsFromFS(ctx, f, filePath)
			if err != nil {
				log.Warn().Err(err).Str("dir", filePath).Msg("Failed to load sections from directory")
				continue
			}
		} else {
			// make an explicit exception for readme.md
			if !strings.HasSuffix(entry.Name(), ".md") || strings.ToLower(entry.Name()) == "readme.md" {
				continue
			}
			
			b, err := fs.ReadFile(f, filePath)
			if err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to read file")
				continue
			}
			
			section, err := model.LoadSectionFromMarkdown(b)
			if err != nil {
				log.Debug().Err(err).Str("file", filePath).Msg("Failed to load section from file")
				continue
			}
			
			err = s.UpsertSection(ctx, section)
			if err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to upsert section")
				continue
			}
			
			log.Debug().Str("slug", section.Slug).Str("file", filePath).Msg("Loaded section")
		}
	}

	return nil
}

// ClearAllSections removes all sections from the database
func (s *Store) ClearAllSections(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Clear all tables
	tables := []string{"section_commands", "section_flags", "section_topics", "section_fts", "sections"}
	for _, table := range tables {
		_, err := tx.ExecContext(ctx, "DELETE FROM "+table)
		if err != nil {
			return errors.Wrapf(err, "failed to clear table %s", table)
		}
	}

	return tx.Commit()
}

// SyncSectionsFromFS completely replaces all sections with content from the filesystem
func (s *Store) SyncSectionsFromFS(ctx context.Context, f fs.FS, dir string) error {
	log.Info().Str("dir", dir).Msg("Starting sync of sections from filesystem")

	// Clear existing sections
	if err := s.ClearAllSections(ctx); err != nil {
		return errors.Wrap(err, "failed to clear existing sections")
	}

	// Load new sections
	if err := s.LoadSectionsFromFS(ctx, f, dir); err != nil {
		return errors.Wrap(err, "failed to load sections from filesystem")
	}

	// Rebuild FTS index
	if err := s.RebuildFTSIndex(ctx); err != nil {
		return errors.Wrap(err, "failed to rebuild FTS index")
	}

	log.Info().Msg("Completed sync of sections from filesystem")
	return nil
}

// GetSectionCount returns the total number of sections in the database
func (s *Store) GetSectionCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections").Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count sections")
	}
	return count, nil
}

// GetSectionStats returns statistics about the sections in the database
func (s *Store) GetSectionStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// Count by section type
	rows, err := s.db.QueryContext(ctx, "SELECT sectionType, COUNT(*) FROM sections GROUP BY sectionType")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get section type stats")
	}
	defer rows.Close()

	for rows.Next() {
		var sectionType string
		var count int
		if err := rows.Scan(&sectionType, &count); err != nil {
			return nil, errors.Wrap(err, "failed to scan section type stats")
		}
		stats["type_"+sectionType] = count
	}

	// Count top level sections
	var topLevel int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE isTopLevel = 1").Scan(&topLevel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count top level sections")
	}
	stats["top_level"] = topLevel

	// Count default sections
	var showDefault int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE showDefault = 1").Scan(&showDefault)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count default sections")
	}
	stats["show_default"] = showDefault

	// Count template sections
	var templates int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE isTemplate = 1").Scan(&templates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count template sections")
	}
	stats["templates"] = templates

	return stats, nil
}
