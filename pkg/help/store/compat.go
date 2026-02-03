package store

import (
	"context"
	"io/fs"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/pkg/errors"
)

// HelpSystem provides a compatibility shim with the existing help system interface
type HelpSystem struct {
	store  *Store
	loader *Loader
}

// NewHelpSystem creates a new HelpSystem with SQLite backing
func NewHelpSystem(dbPath string) (*HelpSystem, error) {
	store, err := New(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store")
	}

	loader := NewLoader(store)

	return &HelpSystem{
		store:  store,
		loader: loader,
	}, nil
}

// NewInMemoryHelpSystem creates a new in-memory HelpSystem
func NewInMemoryHelpSystem() (*HelpSystem, error) {
	store, err := NewInMemory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store")
	}

	loader := NewLoader(store)

	return &HelpSystem{
		store:  store,
		loader: loader,
	}, nil
}

// Close closes the underlying database connection
func (hs *HelpSystem) Close() error {
	return hs.store.Close()
}

// LoadSectionsFromFS loads sections from a filesystem, compatible with existing interface
func (hs *HelpSystem) LoadSectionsFromFS(filesystem fs.FS, dir string) error {
	ctx := context.Background()
	return hs.loader.LoadFromFS(ctx, filesystem, dir)
}

// AddSection adds a single section to the help system
func (hs *HelpSystem) AddSection(section *model.Section) error {
	ctx := context.Background()
	return hs.store.Upsert(ctx, section)
}

// GetSectionWithSlug retrieves a section by its slug
func (hs *HelpSystem) GetSectionWithSlug(slug string) (*model.Section, error) {
	ctx := context.Background()
	return hs.store.GetBySlug(ctx, slug)
}

// GetSections returns all sections (for compatibility with existing interface)
func (hs *HelpSystem) GetSections() ([]*model.Section, error) {
	ctx := context.Background()
	return hs.store.List(ctx, "order_num ASC")
}

// Find executes a predicate-based query
func (hs *HelpSystem) Find(predicate Predicate) ([]*model.Section, error) {
	ctx := context.Background()
	return hs.store.Find(ctx, predicate)
}

// Clear removes all sections from the help system
func (hs *HelpSystem) Clear() error {
	ctx := context.Background()
	return hs.store.Clear(ctx)
}

// Count returns the total number of sections
func (hs *HelpSystem) Count() (int64, error) {
	ctx := context.Background()
	return hs.store.Count(ctx)
}

// Compatibility methods that mimic the existing SectionQuery behavior

// GetTopLevelSections returns all top-level sections
func (hs *HelpSystem) GetTopLevelSections() ([]*model.Section, error) {
	return hs.Find(And(IsTopLevel(), OrderByOrder()))
}

// GetExamplesForTopic returns examples for a specific topic
func (hs *HelpSystem) GetExamplesForTopic(topic string) ([]*model.Section, error) {
	return hs.Find(And(IsExample(), HasTopic(topic), OrderByOrder()))
}

// GetTutorialsForTopic returns tutorials for a specific topic
func (hs *HelpSystem) GetTutorialsForTopic(topic string) ([]*model.Section, error) {
	return hs.Find(And(IsTutorial(), HasTopic(topic), OrderByOrder()))
}

// GetApplicationsForTopic returns applications for a specific topic
func (hs *HelpSystem) GetApplicationsForTopic(topic string) ([]*model.Section, error) {
	return hs.Find(And(IsApplication(), HasTopic(topic), OrderByOrder()))
}

// GetDefaultExamplesForTopic returns default examples for a specific topic
func (hs *HelpSystem) GetDefaultExamplesForTopic(topic string) ([]*model.Section, error) {
	return hs.Find(DefaultExamplesForTopic(topic))
}

// GetDefaultTutorialsForTopic returns default tutorials for a specific topic
func (hs *HelpSystem) GetDefaultTutorialsForTopic(topic string) ([]*model.Section, error) {
	return hs.Find(DefaultTutorialsForTopic(topic))
}

// GetSectionsForCommand returns sections related to a specific command
func (hs *HelpSystem) GetSectionsForCommand(command string) ([]*model.Section, error) {
	return hs.Find(And(HasCommand(command), OrderByOrder()))
}

// GetSectionsForFlag returns sections related to a specific flag
func (hs *HelpSystem) GetSectionsForFlag(flag string) ([]*model.Section, error) {
	return hs.Find(And(HasFlag(flag), OrderByOrder()))
}

// SearchSections performs a text search across sections
func (hs *HelpSystem) SearchSections(term string) ([]*model.Section, error) {
	return hs.Find(And(TextSearch(term), OrderByOrder()))
}

// GetSectionsByType returns sections of a specific type
func (hs *HelpSystem) GetSectionsByType(sectionType model.SectionType) ([]*model.Section, error) {
	return hs.Find(And(IsType(sectionType), OrderByOrder()))
}

// GetStats returns statistics about the help system
func (hs *HelpSystem) GetStats() (map[string]int64, error) {
	ctx := context.Background()
	return hs.loader.GetSectionStats(ctx)
}

// SyncFromFS synchronizes the help system with markdown files from a filesystem
func (hs *HelpSystem) SyncFromFS(filesystem fs.FS, dir string) error {
	ctx := context.Background()
	return hs.loader.SyncFromFS(ctx, filesystem, dir)
}

// Store returns the underlying store (for advanced usage)
func (hs *HelpSystem) Store() *Store {
	return hs.store
}

// Loader returns the underlying loader (for advanced usage)
func (hs *HelpSystem) Loader() *Loader {
	return hs.loader
}
