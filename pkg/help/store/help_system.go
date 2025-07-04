package store

import (
	"context"
	"io/fs"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// HelpSystem represents a SQLite-backed help system
type HelpSystem struct {
	store *Store
	// Sections is maintained for compatibility with the existing interface
	Sections []*help.Section
}

// NewHelpSystem creates a new SQLite-backed help system
func NewHelpSystem(dbPath string) (*HelpSystem, error) {
	store, err := NewStore(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store")
	}
	
	return &HelpSystem{
		store: store,
		Sections: []*help.Section{},
	}, nil
}

// NewInMemoryHelpSystem creates a new in-memory SQLite-backed help system
func NewInMemoryHelpSystem() (*HelpSystem, error) {
	store, err := NewInMemoryStore()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in-memory store")
	}
	
	return &HelpSystem{
		store: store,
		Sections: []*help.Section{},
	}, nil
}

// Close closes the help system
func (hs *HelpSystem) Close() error {
	return hs.store.Close()
}

// Find executes a query using the predicate DSL
func (hs *HelpSystem) Find(ctx context.Context, predicate Predicate) ([]*help.Section, error) {
	return hs.store.Find(ctx, predicate)
}

// AddSection adds a section to the help system
func (hs *HelpSystem) AddSection(ctx context.Context, section *help.Section) error {
	err := hs.store.AddSection(ctx, section)
	if err != nil {
		return err
	}
	
	// Update the in-memory sections list for compatibility
	hs.Sections = append(hs.Sections, section)
	section.HelpSystem = &help.HelpSystem{Sections: hs.Sections}
	
	return nil
}

// GetSectionBySlug retrieves a section by its slug
func (hs *HelpSystem) GetSectionBySlug(ctx context.Context, slug string) (*help.Section, error) {
	return hs.store.GetSectionBySlug(ctx, slug)
}

// LoadSectionsFromFS loads sections from a filesystem
func (hs *HelpSystem) LoadSectionsFromFS(ctx context.Context, f fs.FS, dir string) error {
	// Create a traditional help system to parse the files
	tempHelpSystem := help.NewHelpSystem()
	if err := tempHelpSystem.LoadSectionsFromFS(f, dir); err != nil {
		return errors.Wrap(err, "failed to load sections from filesystem")
	}
	
	// Add all sections to the store
	for _, section := range tempHelpSystem.Sections {
		if err := hs.AddSection(ctx, section); err != nil {
			log.Warn().
				Err(err).
				Str("slug", section.Slug).
				Msg("Failed to add section to store")
		}
	}
	
	return nil
}

// GetTopLevelSections returns all top-level sections
func (hs *HelpSystem) GetTopLevelSections(ctx context.Context) ([]*help.Section, error) {
	return hs.Find(ctx, IsTopLevel())
}

// GetSectionsByType returns sections of a specific type
func (hs *HelpSystem) GetSectionsByType(ctx context.Context, sectionType help.SectionType) ([]*help.Section, error) {
	return hs.Find(ctx, IsType(sectionType.String()))
}

// GetSectionsByTopic returns sections related to a specific topic
func (hs *HelpSystem) GetSectionsByTopic(ctx context.Context, topic string) ([]*help.Section, error) {
	return hs.Find(ctx, HasTopic(topic))
}

// GetSectionsByFlag returns sections related to a specific flag
func (hs *HelpSystem) GetSectionsByFlag(ctx context.Context, flag string) ([]*help.Section, error) {
	return hs.Find(ctx, HasFlag(flag))
}

// GetSectionsByCommand returns sections related to a specific command
func (hs *HelpSystem) GetSectionsByCommand(ctx context.Context, command string) ([]*help.Section, error) {
	return hs.Find(ctx, HasCommand(command))
}

// SearchSections performs full-text search
func (hs *HelpSystem) SearchSections(ctx context.Context, query string) ([]*help.Section, error) {
	return hs.Find(ctx, TextSearch(query))
}

// GetDefaultSections returns sections shown by default
func (hs *HelpSystem) GetDefaultSections(ctx context.Context) ([]*help.Section, error) {
	return hs.Find(ctx, ShownByDefault())
}

// GetStats returns statistics about the help system
func (hs *HelpSystem) GetStats(ctx context.Context) (map[string]int, error) {
	return hs.store.GetStats(ctx)
}

// GetTopLevelHelpPage returns a help page for top-level sections
func (hs *HelpSystem) GetTopLevelHelpPage(ctx context.Context) (*help.HelpPage, error) {
	sections, err := hs.GetTopLevelSections(ctx)
	if err != nil {
		return nil, err
	}
	
	return help.NewHelpPage(sections), nil
}

// GetHelpPageForTopic returns a help page for a specific topic
func (hs *HelpSystem) GetHelpPageForTopic(ctx context.Context, topic string) (*help.HelpPage, error) {
	sections, err := hs.GetSectionsByTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	
	return help.NewHelpPage(sections), nil
}

// GetHelpPageForCommand returns a help page for a specific command
func (hs *HelpSystem) GetHelpPageForCommand(ctx context.Context, command string) (*help.HelpPage, error) {
	sections, err := hs.GetSectionsByCommand(ctx, command)
	if err != nil {
		return nil, err
	}
	
	return help.NewHelpPage(sections), nil
}

// GetHelpPageForFlag returns a help page for a specific flag
func (hs *HelpSystem) GetHelpPageForFlag(ctx context.Context, flag string) (*help.HelpPage, error) {
	sections, err := hs.GetSectionsByFlag(ctx, flag)
	if err != nil {
		return nil, err
	}
	
	return help.NewHelpPage(sections), nil
}

// Example queries using the predicate DSL

// GetExampleSectionsForTopic returns example sections for a topic
func (hs *HelpSystem) GetExampleSectionsForTopic(ctx context.Context, topic string) ([]*help.Section, error) {
	return hs.Find(ctx, And(
		IsExample(),
		HasTopic(topic),
	))
}

// GetTutorialsAndExamples returns both tutorials and examples
func (hs *HelpSystem) GetTutorialsAndExamples(ctx context.Context) ([]*help.Section, error) {
	return hs.Find(ctx, Or(
		IsTutorial(),
		IsExample(),
	))
}

// GetTopLevelDefaultSections returns top-level sections shown by default
func (hs *HelpSystem) GetTopLevelDefaultSections(ctx context.Context) ([]*help.Section, error) {
	return hs.Find(ctx, And(
		IsTopLevel(),
		ShownByDefault(),
	))
}

// GetNonDefaultExamples returns examples not shown by default
func (hs *HelpSystem) GetNonDefaultExamples(ctx context.Context) ([]*help.Section, error) {
	return hs.Find(ctx, And(
		IsExample(),
		NotShownByDefault(),
	))
}

// GetSectionsForTopicAndCommand returns sections that match both topic and command
func (hs *HelpSystem) GetSectionsForTopicAndCommand(ctx context.Context, topic, command string) ([]*help.Section, error) {
	return hs.Find(ctx, And(
		HasTopic(topic),
		HasCommand(command),
	))
}

// SearchInExamples performs full-text search within example sections only
func (hs *HelpSystem) SearchInExamples(ctx context.Context, query string) ([]*help.Section, error) {
	return hs.Find(ctx, And(
		IsExample(),
		TextSearch(query),
	))
}

// Compatibility methods for the existing interface
func (hs *HelpSystem) GetSectionWithSlug(slug string) (*help.Section, error) {
	section, err := hs.GetSectionBySlug(context.Background(), slug)
	if err != nil {
		return nil, help.ErrSectionNotFound
	}
	return section, nil
}

func (hs *HelpSystem) GetTopLevelHelpPage() *help.HelpPage {
	sections, err := hs.GetTopLevelSections(context.Background())
	if err != nil {
		return help.NewHelpPage([]*help.Section{})
	}
	return help.NewHelpPage(sections)
}

func (hs *HelpSystem) SetupCobraRootCommand(cmd *cobra.Command) {
	// Create a wrapper that provides the traditional help system interface
	wrapper := &traditionalHelpSystemWrapper{
		sqliteHelpSystem: hs,
		traditionalHS:    &help.HelpSystem{Sections: hs.Sections},
	}
	
	helpFunc, usageFunc := help.GetCobraHelpUsageFuncs(wrapper.traditionalHS)
	helpTemplate, usageTemplate := help.GetCobraHelpUsageTemplates(wrapper.traditionalHS)

	cmd.PersistentFlags().Bool("long-help", false, "Show long help")

	cmd.SetHelpFunc(helpFunc)
	cmd.SetUsageFunc(usageFunc)
	cmd.SetHelpTemplate(helpTemplate)
	cmd.SetUsageTemplate(usageTemplate)

	helpCmd := help.NewCobraHelpCommand(wrapper.traditionalHS)
	cmd.SetHelpCommand(helpCmd)
}

// traditionalHelpSystemWrapper provides compatibility with the traditional help system
type traditionalHelpSystemWrapper struct {
	sqliteHelpSystem *HelpSystem
	traditionalHS    *help.HelpSystem
}

func (hs *HelpSystem) ComputeRenderData(userQuery *help.SectionQuery) (map[string]interface{}, bool) {
	// Convert the traditional query to our predicate system
	// This is a simplified version - you might need to implement a full converter
	
	// For now, just return all sections - this can be improved
	allSections, err := hs.Find(context.Background(), All())
	if err != nil {
		return map[string]interface{}{}, true
	}
	
	// Use the traditional query system to filter sections
	filteredSections := userQuery.FindSections(allSections)
	
	return map[string]interface{}{
		"Sections": filteredSections,
	}, len(filteredSections) == 0
}

func (hs *HelpSystem) RenderTopicHelp(section *help.Section, options *help.RenderOptions) (string, error) {
	// This is a simplified implementation
	// You might need to implement full rendering logic here
	return section.Content, nil
}
