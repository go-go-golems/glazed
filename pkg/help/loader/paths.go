package loader

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
)

// ReplaceStoreWithPaths clears any preloaded sections and reloads the help
// store from the provided files or directories.
func ReplaceStoreWithPaths(ctx context.Context, hs *help.HelpSystem, paths []string) error {
	if err := hs.Store.Clear(ctx); err != nil {
		return errors.Wrap(err, "clearing preloaded sections")
	}
	return LoadPaths(ctx, hs, paths)
}

// LoadPaths loads one or more markdown files or directories into the help
// store. Directories are walked recursively.
func LoadPaths(ctx context.Context, hs *help.HelpSystem, paths []string) error {
	for _, input := range paths {
		info, err := os.Stat(input)
		if err != nil {
			return errors.Wrapf(err, "stat %q", input)
		}
		if info.IsDir() {
			if err := LoadDir(ctx, hs, input); err != nil {
				return errors.Wrapf(err, "loading directory %q", input)
			}
			continue
		}
		if err := LoadFile(ctx, hs, input); err != nil {
			return errors.Wrapf(err, "loading file %q", input)
		}
	}
	return nil
}

// LoadDir walks a local directory recursively and loads all non-README markdown
// files into the help store.
func LoadDir(ctx context.Context, hs *help.HelpSystem, dir string) error {
	return filepath.WalkDir(dir, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if name == "readme.md" || !strings.HasSuffix(name, ".md") {
			return nil
		}
		return LoadFile(ctx, hs, filePath)
	})
}

// LoadFile loads one markdown file from disk and upserts the parsed section.
func LoadFile(ctx context.Context, hs *help.HelpSystem, filePath string) error {
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	section, err := help.LoadSectionFromMarkdown(data)
	if err != nil {
		return err
	}
	return hs.Store.Upsert(ctx, section)
}
