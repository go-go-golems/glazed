package store

import (
	"bytes"
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/glazed/pkg/help/model"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Loader handles loading and syncing markdown files into the SQLite store
type Loader struct {
	store *Store
}

// NewLoader creates a new loader for the given store
func NewLoader(store *Store) *Loader {
	return &Loader{store: store}
}

// LoadFromMarkdown parses a markdown file and returns a Section
func (l *Loader) LoadFromMarkdown(markdownBytes []byte) (*model.Section, error) {
	var metaData map[string]interface{}

	inputReader := bytes.NewReader(markdownBytes)
	rest, err := frontmatter.Parse(inputReader, &metaData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse frontmatter")
	}

	section := &model.Section{}

	// Parse metadata
	if title, ok := metaData["Title"]; ok {
		section.Title = title.(string)
	}
	if subTitle, ok := metaData["SubTitle"]; ok {
		section.SubTitle = subTitle.(string)
	}
	if short, ok := metaData["Short"]; ok {
		section.Short = short.(string)
	}

	// Parse section type
	if sectionType, ok := metaData["SectionType"]; ok {
		section.SectionType, err = model.SectionTypeFromString(sectionType.(string))
		if err != nil {
			return nil, errors.Wrap(err, "invalid section type")
		}
	} else {
		section.SectionType = model.SectionGeneralTopic
	}

	// Parse slug
	if slug, ok := metaData["Slug"]; ok {
		section.Slug = slug.(string)
	}

	// Set content
	section.Content = string(rest)

	// Parse arrays - initialize to empty slices if not present
	if topics, ok := metaData["Topics"]; ok {
		section.Topics = strings2.InterfaceToStringList(topics)
	} else {
		section.Topics = []string{}
	}
	if flags, ok := metaData["Flags"]; ok {
		section.Flags = strings2.InterfaceToStringList(flags)
	} else {
		section.Flags = []string{}
	}
	if commands, ok := metaData["Commands"]; ok {
		section.Commands = strings2.InterfaceToStringList(commands)
	} else {
		section.Commands = []string{}
	}

	// Parse boolean flags
	if isTopLevel, ok := metaData["IsTopLevel"]; ok {
		section.IsTopLevel = isTopLevel.(bool)
	}
	if isTemplate, ok := metaData["IsTemplate"]; ok {
		section.IsTemplate = isTemplate.(bool)
	}
	if showPerDefault, ok := metaData["ShowPerDefault"]; ok {
		section.ShowPerDefault = showPerDefault.(bool)
	}

	// Parse order
	if order, ok := metaData["Order"]; ok {
		switch v := order.(type) {
		case int:
			section.Order = v
		case float64:
			section.Order = int(v)
		}
	}

	// Validate required fields
	if section.Slug == "" || section.Title == "" {
		return nil, errors.New("missing required fields: slug and title")
	}

	return section, nil
}

// LoadFromFS loads all markdown files from a filesystem into the store
func (l *Loader) LoadFromFS(ctx context.Context, filesystem fs.FS, rootDir string) error {
	log.Info().Str("dir", rootDir).Msg("Loading sections from filesystem")

	entries, err := fs.ReadDir(filesystem, rootDir)
	if err != nil {
		log.Warn().Err(err).Str("dir", rootDir).Msg("Failed to read directory")
		return errors.Wrap(err, "failed to read directory")
	}

	for _, entry := range entries {
		filePath := filepath.Join(rootDir, entry.Name())

		if entry.IsDir() {
			// Recursively load from subdirectories
			err = l.LoadFromFS(ctx, filesystem, filePath)
			if err != nil {
				log.Warn().Err(err).Str("dir", filePath).Msg("Failed to load sections from subdirectory")
				continue
			}
		} else {
			// Skip non-markdown files and README.md
			if !strings.HasSuffix(entry.Name(), ".md") || strings.ToLower(entry.Name()) == "readme.md" {
				continue
			}

			err = l.loadFileFromFS(ctx, filesystem, filePath)
			if err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to load section from file")
				continue
			}
		}
	}

	return nil
}

// loadFileFromFS loads a single markdown file from the filesystem
func (l *Loader) loadFileFromFS(ctx context.Context, filesystem fs.FS, filePath string) error {
	log.Debug().Str("file", filePath).Msg("Loading section from file")

	fileBytes, err := fs.ReadFile(filesystem, filePath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	section, err := l.LoadFromMarkdown(fileBytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse markdown")
	}

	// Upsert the section (insert or update based on slug)
	err = l.store.Upsert(ctx, section)
	if err != nil {
		return errors.Wrap(err, "failed to upsert section")
	}

	log.Debug().Str("slug", section.Slug).Str("title", section.Title).Msg("Loaded section")
	return nil
}

// SyncFromFS synchronizes the store with markdown files from a filesystem
// This will clear the store and reload all sections
func (l *Loader) SyncFromFS(ctx context.Context, filesystem fs.FS, rootDir string) error {
	log.Info().Str("dir", rootDir).Msg("Syncing sections from filesystem")

	// Clear existing sections
	err := l.store.Clear(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to clear store")
	}

	// Load all sections from filesystem
	err = l.LoadFromFS(ctx, filesystem, rootDir)
	if err != nil {
		return errors.Wrap(err, "failed to load from filesystem")
	}

	count, err := l.store.Count(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to count sections")
	}

	log.Info().Int64("count", count).Msg("Synced sections from filesystem")
	return nil
}

// LoadSections loads multiple sections from markdown bytes
func (l *Loader) LoadSections(ctx context.Context, markdownFiles map[string][]byte) error {
	log.Info().Int("count", len(markdownFiles)).Msg("Loading sections from markdown files")

	for filename, content := range markdownFiles {
		section, err := l.LoadFromMarkdown(content)
		if err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Failed to parse markdown file")
			continue
		}

		err = l.store.Upsert(ctx, section)
		if err != nil {
			log.Warn().Err(err).Str("file", filename).Str("slug", section.Slug).Msg("Failed to upsert section")
			continue
		}

		log.Debug().Str("file", filename).Str("slug", section.Slug).Msg("Loaded section")
	}

	return nil
}

// BatchUpsert performs bulk upsert of sections
func (l *Loader) BatchUpsert(ctx context.Context, sections []*model.Section) error {
	log.Info().Int("count", len(sections)).Msg("Batch upserting sections")

	// TODO: Implement actual batch operation for better performance
	// For now, just iterate through sections
	for _, section := range sections {
		err := l.store.Upsert(ctx, section)
		if err != nil {
			log.Warn().Err(err).Str("slug", section.Slug).Msg("Failed to upsert section")
			continue
		}
	}

	return nil
}

// GetSectionStats returns statistics about the loaded sections
func (l *Loader) GetSectionStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count total sections
	total, err := l.store.Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count total sections")
	}
	stats["total"] = total

	// Count by type
	for _, sectionType := range []model.SectionType{
		model.SectionGeneralTopic,
		model.SectionExample,
		model.SectionApplication,
		model.SectionTutorial,
	} {
		sections, err := l.store.Find(ctx, IsType(sectionType))
		if err != nil {
			return nil, errors.Wrap(err, "failed to count sections by type")
		}
		stats[sectionType.String()] = int64(len(sections))
	}

	// Count top-level sections
	topLevel, err := l.store.Find(ctx, IsTopLevel())
	if err != nil {
		return nil, errors.Wrap(err, "failed to count top-level sections")
	}
	stats["top_level"] = int64(len(topLevel))

	// Count default sections
	defaults, err := l.store.Find(ctx, ShownByDefault())
	if err != nil {
		return nil, errors.Wrap(err, "failed to count default sections")
	}
	stats["shown_by_default"] = int64(len(defaults))

	return stats, nil
}
