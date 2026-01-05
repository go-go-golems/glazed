//go:build sqlite_fts5

package store

import (
	"github.com/pkg/errors"
)

// createFTSTables creates FTS5 tables and triggers
func (s *Store) createFTSTables() error {
	// Create FTS5 virtual table for full-text search
	ftsTable := `
		CREATE VIRTUAL TABLE IF NOT EXISTS sections_fts USING fts5(
			slug,
			title,
			sub_title,
			short,
			content,
			topics,
			flags,
			commands,
			content='sections',
			content_rowid='id'
		);
	`

	if _, err := s.db.Exec(ftsTable); err != nil {
		return errors.Wrap(err, "failed to create FTS table")
	}

	// Create triggers to keep FTS table in sync
	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS sections_fts_insert AFTER INSERT ON sections BEGIN
			INSERT INTO sections_fts(rowid, slug, title, sub_title, short, content, topics, flags, commands)
			VALUES (new.id, new.slug, new.title, new.sub_title, new.short, new.content, new.topics, new.flags, new.commands);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS sections_fts_delete AFTER DELETE ON sections BEGIN
			INSERT INTO sections_fts(sections_fts, rowid, slug, title, sub_title, short, content, topics, flags, commands)
			VALUES ('delete', old.id, old.slug, old.title, old.sub_title, old.short, old.content, old.topics, old.flags, old.commands);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS sections_fts_update AFTER UPDATE ON sections BEGIN
			INSERT INTO sections_fts(sections_fts, rowid, slug, title, sub_title, short, content, topics, flags, commands)
			VALUES ('delete', old.id, old.slug, old.title, old.sub_title, old.short, old.content, old.topics, old.flags, old.commands);
			INSERT INTO sections_fts(rowid, slug, title, sub_title, short, content, topics, flags, commands)
			VALUES (new.id, new.slug, new.title, new.sub_title, new.short, new.content, new.topics, new.flags, new.commands);
		END;`,
	}

	for _, trigger := range triggers {
		if _, err := s.db.Exec(trigger); err != nil {
			return errors.Wrap(err, "failed to create trigger")
		}
	}

	// log.Trace().Msg("Created FTS5 tables and triggers")
	return nil
}
