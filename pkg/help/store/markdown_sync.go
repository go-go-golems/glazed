package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

// SyncMarkdownDir walks a directory, loads all .md files, and upserts them into the store.
func (s *Store) SyncMarkdownDir(dir string) ([]*model.Section, error) {
	var sections []*model.Section
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		sec, err := LoadSectionFromMarkdown(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		if err := s.UpsertSection(sec); err != nil {
			return fmt.Errorf("upsert %s: %w", path, err)
		}
		sections = append(sections, sec)
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Rebuild FTS index after all inserts
	if _, err := s.db.Exec(`INSERT INTO section_fts(section_fts) VALUES ('rebuild')`); err != nil {
		return nil, fmt.Errorf("rebuild fts: %w", err)
	}
	return sections, nil
} 