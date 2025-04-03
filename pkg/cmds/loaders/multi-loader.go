package loaders

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// BaseCommand represents the minimal structure needed to determine the type of a command
type BaseCommand struct {
	Type string `yaml:"type" json:"type"`
}

// MultiLoader implements CommandLoader and dispatches to registered loaders based on the Type field
type MultiLoader struct {
	loaders map[string]CommandLoader
	// defaultLoader is used when no Type field is present
	defaultLoader CommandLoader
}

// NewMultiLoader creates a new MultiLoader instance
func NewMultiLoader() *MultiLoader {
	return &MultiLoader{
		loaders: make(map[string]CommandLoader),
	}
}

// RegisterLoader registers a new loader for a specific type
func (m *MultiLoader) RegisterLoader(typeName string, loader CommandLoader) {
	m.loaders[typeName] = loader
}

// SetDefaultLoader sets the default loader to use when no Type field is present
func (m *MultiLoader) SetDefaultLoader(loader CommandLoader) {
	m.defaultLoader = loader
}

// findSupportedLoader tries each registered loader to find one that supports the file
func (m *MultiLoader) findSupportedLoader(f fs.FS, fileName string) CommandLoader {
	for _, loader := range m.loaders {
		if loader.IsFileSupported(f, fileName) {
			return loader
		}
	}
	return nil
}

// getLoaderForFile determines which loader to use for a given file
func (m *MultiLoader) getLoaderForFile(f fs.FS, fileName string, content []byte) (CommandLoader, error) {
	// Try to parse as YAML to get the type
	var base BaseCommand
	if err := yaml.Unmarshal(content, &base); err == nil {
		// If we have a type field, try to get the corresponding loader
		if base.Type != "" {
			if loader, ok := m.loaders[base.Type]; ok {
				return loader, nil
			}
		}
	}

	// If we have a default loader and it supports the file, use it
	if m.defaultLoader != nil && m.defaultLoader.IsFileSupported(f, fileName) {
		return m.defaultLoader, nil
	}

	// Try to find a loader that supports this file
	if loader := m.findSupportedLoader(f, fileName); loader != nil {
		return loader, nil
	}

	// No suitable loader found
	if base.Type != "" {
		return nil, fmt.Errorf("no loader registered for type: %s and no loader supports the file", base.Type)
	}
	return nil, fmt.Errorf("no type field found in command file %s and no loader supports the file", fileName)
}

// LoadCommands implements the CommandLoader interface
func (m *MultiLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) (commands []cmds.Command, err error) {
	// First, read the file
	file, err := f.Open(entryName)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "error closing file")
		}
	}()

	// Read the content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}

	// Get the appropriate loader
	loader, err := m.getLoaderForFile(f, entryName, content)
	if err != nil {
		return nil, err
	}

	// Use the loader to load the commands
	return loader.LoadCommands(f, entryName, options, aliasOptions)
}

// IsFileSupported implements the CommandLoader interface
func (m *MultiLoader) IsFileSupported(f fs.FS, fileName string) (supported bool) {
	// Try to open and read the file
	file, err := f.Open(fileName)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			supported = false
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	// Try to get a loader for this file
	loader, err := m.getLoaderForFile(f, fileName, content)
	return err == nil && loader != nil
}
