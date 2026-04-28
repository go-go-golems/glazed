package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHelpSystem(t *testing.T) *help.HelpSystem {
	hs := help.NewHelpSystem()

	sections := []*model.Section{
		{
			Slug:           "help-system",
			Title:          "Help System",
			Short:          "Overview of the help system",
			SectionType:    model.SectionGeneralTopic,
			Topics:         []string{"help", "documentation"},
			Commands:       []string{"help"},
			IsTopLevel:     true,
			ShowPerDefault: true,
			Order:          0,
			Content:        "The help system allows...",
		},
		{
			Slug:           "json-example",
			Title:          "JSON Output Example",
			Short:          "How to format JSON output",
			SectionType:    model.SectionExample,
			Topics:         []string{"json", "output"},
			Commands:       []string{"json"},
			IsTopLevel:     false,
			ShowPerDefault: true,
			Order:          1,
			Content:        "Use the json command with...",
		},
		{
			Slug:           "csv-tutorial",
			Title:          "Working with CSV",
			Short:          "Tutorial on CSV processing",
			SectionType:    model.SectionTutorial,
			Topics:         []string{"csv", "data"},
			Commands:       []string{"csv"},
			Flags:          []string{"separator"},
			IsTopLevel:     false,
			ShowPerDefault: false,
			Order:          2,
			Content:        "CSV files can be processed...",
		},
	}

	ctx := context.Background()
	for _, s := range sections {
		err := hs.Store.Upsert(ctx, s)
		require.NoError(t, err)
	}

	return hs
}

func TestBuildExportPredicate_NoFilters(t *testing.T) {
	s := &ExportSettings{}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	require.NotNil(t, pred)
}

func TestBuildExportPredicate_ByType(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Type: "Example"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "json-example", sections[0].Slug)
}

func TestBuildExportPredicate_InvalidType(t *testing.T) {
	s := &ExportSettings{Type: "tutorial"}
	_, err := buildExportPredicate(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section type")
}

func TestBuildExportPredicate_ByTopic(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Topic: "json"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "json-example", sections[0].Slug)
}

func TestBuildExportPredicate_ByCommand(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Command: "csv"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "csv-tutorial", sections[0].Slug)
}

func TestBuildExportPredicate_ByFlag(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Flag: "separator"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "csv-tutorial", sections[0].Slug)
}

func TestBuildExportPredicate_BySlug(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Slug: "help-system"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "help-system", sections[0].Slug)
}

func TestBuildExportPredicate_ByMultipleSlugs(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	s := &ExportSettings{Slug: "help-system, json-example"}
	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)
	assert.Len(t, sections, 2)
}

func TestExportToFiles(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	s := &ExportSettings{
		Format:      "files",
		OutputPath:  tmpDir,
		FlattenDirs: false,
	}

	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)

	err = exportToFiles(sections, s)
	require.NoError(t, err)

	// Verify directory structure
	assert.DirExists(t, filepath.Join(tmpDir, "generaltopic"))
	assert.DirExists(t, filepath.Join(tmpDir, "example"))
	assert.DirExists(t, filepath.Join(tmpDir, "tutorial"))

	// Verify a file exists and can be parsed back
	helpSystemPath := filepath.Join(tmpDir, "generaltopic", "help-system.md")
	require.FileExists(t, helpSystemPath)

	data, err := os.ReadFile(helpSystemPath)
	require.NoError(t, err)
	parsed, err := model.ParseSectionFromMarkdown(data)
	require.NoError(t, err)
	assert.Equal(t, "help-system", parsed.Slug)
	assert.Equal(t, "Help System", parsed.Title)
	assert.Equal(t, "The help system allows...", parsed.Content)
}

func TestExportToFiles_Flatten(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	s := &ExportSettings{
		Format:      "files",
		OutputPath:  tmpDir,
		FlattenDirs: true,
	}

	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)

	err = exportToFiles(sections, s)
	require.NoError(t, err)

	// All files should be in the root
	assert.FileExists(t, filepath.Join(tmpDir, "help-system.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "json-example.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "csv-tutorial.md"))
}

func TestExportToFiles_RejectsUnsafeSlug(t *testing.T) {
	tmpDir := t.TempDir()
	sections := []*model.Section{{
		Slug:        "../escape",
		Title:       "Unsafe",
		SectionType: model.SectionGeneralTopic,
	}}

	err := exportToFiles(sections, &ExportSettings{Format: "files", OutputPath: tmpDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe section slug")
	assert.NoFileExists(t, filepath.Join(tmpDir, "..", "escape.md"))
}

func TestSafeSectionFilePath_AllowsNormalSlug(t *testing.T) {
	tmpDir := t.TempDir()
	path, err := safeSectionFilePath(tmpDir, "help-system")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "help-system.md"), path)
}

func TestExportToSQLite(t *testing.T) {
	hs := setupTestHelpSystem(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	s := &ExportSettings{
		Format:     "sqlite",
		OutputPath: dbPath,
	}

	pred, err := buildExportPredicate(s)
	require.NoError(t, err)
	sections, err := hs.Store.Find(ctx, pred)
	require.NoError(t, err)

	err = exportToSQLite(ctx, sections, s)
	require.NoError(t, err)

	require.FileExists(t, dbPath)

	// Re-open and query
	exportStore, err := store.New(dbPath)
	require.NoError(t, err)
	defer func() {
		_ = exportStore.Close()
	}()

	count, err := exportStore.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	section, err := exportStore.GetBySlug(ctx, "help-system")
	require.NoError(t, err)
	assert.Equal(t, "Help System", section.Title)
	assert.Equal(t, "The help system allows...", section.Content)
}

func TestReconstructMarkdown(t *testing.T) {
	section := &model.Section{
		Slug:           "test-section",
		Title:          "Test Section",
		Short:          "A test section",
		SectionType:    model.SectionGeneralTopic,
		Topics:         []string{"test", "docs"},
		Commands:       []string{"test"},
		Flags:          []string{"verbose"},
		IsTopLevel:     true,
		ShowPerDefault: true,
		Order:          5,
		Content:        "This is the content.",
	}

	data, err := reconstructMarkdown(section)
	require.NoError(t, err)

	// Should be parseable back
	parsed, err := model.ParseSectionFromMarkdown(data)
	require.NoError(t, err)
	assert.Equal(t, section.Slug, parsed.Slug)
	assert.Equal(t, section.Title, parsed.Title)
	assert.Equal(t, section.Short, parsed.Short)
	assert.Equal(t, section.SectionType, parsed.SectionType)
	assert.Equal(t, section.Topics, parsed.Topics)
	assert.Equal(t, section.Commands, parsed.Commands)
	assert.Equal(t, section.Flags, parsed.Flags)
	assert.Equal(t, section.IsTopLevel, parsed.IsTopLevel)
	assert.Equal(t, section.ShowPerDefault, parsed.ShowPerDefault)
	assert.Equal(t, section.Order, parsed.Order)
	assert.Equal(t, section.Content, parsed.Content)
}
