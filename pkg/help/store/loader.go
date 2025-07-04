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

// Loader handles loading markdown files into the store
type Loader struct {
	store *Store
}

// NewLoader creates a new Loader instance
func NewLoader(store *Store) *Loader {
	return &Loader{store: store}
}

// LoadFromFS loads all markdown files from a filesystem directory
func (l *Loader) LoadFromFS(ctx context.Context, f fs.FS, dir string) error {
	entries, err := fs.ReadDir(f, dir)
	if err != nil {
		log.Warn().Err(err).Str("dir", dir).Msg("Failed to read directory")
		return nil
	}

	for _, entry := range entries {
		filePath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := l.LoadFromFS(ctx, f, filePath); err != nil {
				log.Warn().Err(err).Str("dir", filePath).Msg("Failed to load sections from directory")
				continue
			}
		} else {
			// Skip non-markdown files and readme.md
			if !strings.HasSuffix(entry.Name(), ".md") || strings.ToLower(entry.Name()) == "readme.md" {
				continue
			}

			if err := l.loadMarkdownFile(ctx, f, filePath); err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to load markdown file")
				continue
			}
		}
	}

	// Rebuild FTS index after loading all files
	if err := l.store.RebuildFTS(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to rebuild FTS index")
	}

	return nil
}

// loadMarkdownFile loads a single markdown file
func (l *Loader) loadMarkdownFile(ctx context.Context, f fs.FS, filePath string) error {
	data, err := fs.ReadFile(f, filePath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	section, err := l.parseMarkdown(data)
	if err != nil {
		return errors.Wrap(err, "failed to parse markdown")
	}

	// If no slug is provided, derive it from the file path
	if section.Slug == "" {
		section.Slug = l.deriveSlugFromPath(filePath)
	}

	if err := l.store.Upsert(ctx, section); err != nil {
		return errors.Wrap(err, "failed to upsert section")
	}

	log.Debug().Str("slug", section.Slug).Str("file", filePath).Msg("Loaded section")
	return nil
}

// parseMarkdown parses markdown content into a Section
func (l *Loader) parseMarkdown(markdownBytes []byte) (*model.Section, error) {
	var metaData map[string]interface{}

	inputReader := bytes.NewReader(markdownBytes)
	rest, err := frontmatter.Parse(inputReader, &metaData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse frontmatter")
	}

	section := &model.Section{
		Content: string(rest),
	}

	// Parse metadata
	if title, ok := metaData["Title"]; ok {
		section.Title = title.(string)
	}
	if subtitle, ok := metaData["SubTitle"]; ok {
		section.Subtitle = subtitle.(string)
	}
	if short, ok := metaData["Short"]; ok {
		section.Short = short.(string)
	}
	if slug, ok := metaData["Slug"]; ok {
		section.Slug = slug.(string)
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

	// Parse boolean flags
	if isTopLevel, ok := metaData["IsTopLevel"]; ok {
		section.IsTopLevel = isTopLevel.(bool)
	}
	if isTemplate, ok := metaData["IsTemplate"]; ok {
		section.IsTemplate = isTemplate.(bool)
	}
	if showDefault, ok := metaData["ShowPerDefault"]; ok {
		section.ShowDefault = showDefault.(bool)
	}

	// Parse order
	if order, ok := metaData["Order"]; ok {
		section.Order = order.(int)
	}

	// Parse string arrays
	if topics, ok := metaData["Topics"]; ok {
		section.Topics = strings2.InterfaceToStringList(topics)
	}
	if flags, ok := metaData["Flags"]; ok {
		section.Flags = strings2.InterfaceToStringList(flags)
	}
	if commands, ok := metaData["Commands"]; ok {
		section.Commands = strings2.InterfaceToStringList(commands)
	}

	// Validate required fields
	if section.Slug == "" || section.Title == "" {
		return nil, errors.New("missing required slug or title")
	}

	return section, nil
}

// deriveSlugFromPath creates a slug from a file path
func (l *Loader) deriveSlugFromPath(filePath string) string {
	// Remove extension and directory
	base := filepath.Base(filePath)
	slug := strings.TrimSuffix(base, filepath.Ext(base))
	
	// Convert to lowercase and replace spaces/underscores with hyphens
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	return slug
}

// SyncDirectory synchronizes a directory with the store
// This is a more advanced version that can handle incremental updates
func (l *Loader) SyncDirectory(ctx context.Context, f fs.FS, dir string) error {
	// For now, we'll just reload everything
	// In a production system, you might want to check file modification times
	// and only reload changed files
	return l.LoadFromFS(ctx, f, dir)
}
