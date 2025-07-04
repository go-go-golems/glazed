package store

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

// UpsertSection inserts or updates a Section and its related data.
func (s *Store) UpsertSection(sec *model.Section) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Upsert section
	res, err := tx.Exec(`
		INSERT INTO sections (slug, title, subtitle, short, content, sectionType, isTopLevel, isTemplate, showDefault, ord)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title=excluded.title, subtitle=excluded.subtitle, short=excluded.short, content=excluded.content,
			sectionType=excluded.sectionType, isTopLevel=excluded.isTopLevel, isTemplate=excluded.isTemplate,
			showDefault=excluded.showDefault, ord=excluded.ord
	`, sec.Slug, sec.Title, sec.Subtitle, sec.Short, sec.Content, sec.SectionType.String(), sec.IsTopLevel, sec.IsTemplate, sec.ShowPerDefault, sec.Order)
	if err != nil {
		return fmt.Errorf("upsert section: %w", err)
	}
	var sectionID int64
	if sec.ID > 0 {
		sectionID = sec.ID
	} else {
		id, err := res.LastInsertId()
		if err != nil {
			// fallback: fetch by slug
			row := tx.QueryRow("SELECT id FROM sections WHERE slug = ?", sec.Slug)
			if err := row.Scan(&sectionID); err != nil {
				return fmt.Errorf("get section id: %w", err)
			}
		} else {
			sectionID = id
		}
	}

	// Remove old topics/flags/commands
	for _, tbl := range []string{"section_topics", "section_flags", "section_commands"} {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE section_id = ?", tbl), sectionID); err != nil {
			return fmt.Errorf("delete %s: %w", tbl, err)
		}
	}

	// Insert topics
	for _, topic := range sec.Topics {
		if _, err := tx.Exec("INSERT INTO section_topics (section_id, topic) VALUES (?, ?)", sectionID, topic); err != nil {
			return fmt.Errorf("insert topic: %w", err)
		}
	}
	// Insert flags
	for _, flag := range sec.Flags {
		if _, err := tx.Exec("INSERT INTO section_flags (section_id, flag) VALUES (?, ?)", sectionID, flag); err != nil {
			return fmt.Errorf("insert flag: %w", err)
		}
	}
	// Insert commands
	for _, cmd := range sec.Commands {
		if _, err := tx.Exec("INSERT INTO section_commands (section_id, command) VALUES (?, ?)", sectionID, cmd); err != nil {
			return fmt.Errorf("insert command: %w", err)
		}
	}

	// Update FTS
	if _, err := tx.Exec(`INSERT OR REPLACE INTO section_fts(rowid, slug, title, subtitle, short, content) VALUES (?, ?, ?, ?, ?, ?)`,
		sectionID, sec.Slug, sec.Title, sec.Subtitle, sec.Short, sec.Content); err != nil {
		return fmt.Errorf("upsert fts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
