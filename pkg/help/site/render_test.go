package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
)

func TestRenderSite_ExportsStaticJSONTree(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()

	tmpDocs := t.TempDir()
	markdown := `---
Title: Exported Topic
Slug: exported-topic
Short: Exported short description
SectionType: GeneralTopic
Topics: [export, static]
Commands: [serve]
Flags: [output-dir]
IsTopLevel: true
ShowPerDefault: true
---

Exported body.
`
	if err := os.WriteFile(filepath.Join(tmpDocs, "topic.md"), []byte(markdown), 0o644); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "site")
	err := RenderSite(ctx, hs, &RenderSiteSettings{
		OutputDir: outDir,
		SiteTitle: "Static Export",
		DataDir:   DefaultDataDir,
		Overwrite: false,
		Paths:     []string{tmpDocs},
	})
	if err != nil {
		t.Fatalf("RenderSite: %v", err)
	}

	checkFileContains(t, filepath.Join(outDir, "index.html"), "Glazed Help Browser")
	checkFileContains(t, filepath.Join(outDir, "site-config.js"), `"mode": "static"`)
	checkFileContains(t, filepath.Join(outDir, "site-config.js"), `"dataBasePath": "./site-data"`)
	checkFileContains(t, filepath.Join(outDir, "site-data", "sections.json"), `"slug": "exported-topic"`)
	checkFileContains(t, filepath.Join(outDir, "site-data", "sections", "exported-topic.json"), `Exported body.`)
	checkFileContains(t, filepath.Join(outDir, "site-data", "indexes", "topics.json"), `"exported-topic"`)
	checkFileContains(t, filepath.Join(outDir, "site-data", "manifest.json"), `"siteTitle": "Static Export"`)
}

func TestRenderSite_RejectsNonEmptyOutputDirWithoutOverwrite(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "existing.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	err := RenderSite(ctx, hs, &RenderSiteSettings{
		OutputDir: outDir,
		SiteTitle: DefaultSiteTitle,
		DataDir:   DefaultDataDir,
		Overwrite: false,
	})
	if err == nil {
		t.Fatalf("expected non-empty output dir error")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderSite_UsesEmbeddedDocsWhenNoPathsAreProvided(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()
	hs.AddSection(helpSection("embedded-topic", "Embedded Topic"))

	outDir := filepath.Join(t.TempDir(), "site")
	err := RenderSite(ctx, hs, &RenderSiteSettings{
		OutputDir: outDir,
		SiteTitle: DefaultSiteTitle,
		DataDir:   DefaultDataDir,
	})
	if err != nil {
		t.Fatalf("RenderSite: %v", err)
	}

	checkFileContains(t, filepath.Join(outDir, "site-data", "sections.json"), `"slug": "embedded-topic"`)
}

func TestRenderSite_ExplicitPathsReplacePreloadedDocs(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()
	hs.AddSection(helpSection("embedded-topic", "Embedded Topic"))

	tmpDocs := t.TempDir()
	markdown := `---
Title: Explicit Topic
Slug: explicit-topic
SectionType: GeneralTopic
---

Loaded from explicit path.
`
	if err := os.WriteFile(filepath.Join(tmpDocs, "topic.md"), []byte(markdown), 0o644); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "site")
	err := RenderSite(ctx, hs, &RenderSiteSettings{
		OutputDir: outDir,
		SiteTitle: DefaultSiteTitle,
		DataDir:   DefaultDataDir,
		Paths:     []string{tmpDocs},
	})
	if err != nil {
		t.Fatalf("RenderSite: %v", err)
	}

	checkFileContains(t, filepath.Join(outDir, "site-data", "sections.json"), `"slug": "explicit-topic"`)
	checkFileMissingSubstring(t, filepath.Join(outDir, "site-data", "sections.json"), `"slug": "embedded-topic"`)
}

func helpSection(slug, title string) *model.Section {
	return &model.Section{
		Slug:        slug,
		Title:       title,
		SectionType: model.SectionGeneralTopic,
	}
}

func checkFileContains(t *testing.T, path, substring string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), substring) {
		t.Fatalf("expected %s to contain %q, got %q", path, substring, string(data))
	}
}

func checkFileMissingSubstring(t *testing.T, path, substring string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(data), substring) {
		t.Fatalf("expected %s to not contain %q", path, substring)
	}
}
