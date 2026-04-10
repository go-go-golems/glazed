package site

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	helploader "github.com/go-go-golems/glazed/pkg/help/loader"
	"github.com/go-go-golems/glazed/pkg/help/server"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/go-go-golems/glazed/pkg/web"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	DefaultOutputDir = "glaze-site"
	DefaultSiteTitle = "Glazed Help Browser"
	DefaultDataDir   = "site-data"
)

type RenderSiteCommand struct {
	*cmds.CommandDescription
	helpSystem *help.HelpSystem
}

var _ cmds.BareCommand = (*RenderSiteCommand)(nil)

type RenderSiteSettings struct {
	OutputDir string   `glazed:"output-dir"`
	SiteTitle string   `glazed:"site-title"`
	BasePath  string   `glazed:"base-path"`
	DataDir   string   `glazed:"data-dir"`
	Overwrite bool     `glazed:"overwrite"`
	Paths     []string `glazed:"paths"`
}

type RuntimeConfig struct {
	Mode        string `json:"mode"`
	APIBaseURL  string `json:"apiBaseUrl,omitempty"`
	DataBaseURL string `json:"dataBasePath,omitempty"`
	SiteTitle   string `json:"siteTitle"`
}

type Manifest struct {
	GeneratedBy   string   `json:"generatedBy"`
	SiteTitle     string   `json:"siteTitle"`
	TotalSections int      `json:"totalSections"`
	DataBasePath  string   `json:"dataBasePath"`
	TopLevelSlugs []string `json:"topLevelSlugs"`
	DefaultSlugs  []string `json:"defaultSlugs"`
}

type SlugIndex struct {
	Items map[string][]string `json:"items"`
}

func NewRenderSiteCommand(hs *help.HelpSystem) (*RenderSiteCommand, error) {
	return &RenderSiteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"render-site",
			cmds.WithShort("Render help documentation as a static website"),
			cmds.WithLong(`Discover Glazed Markdown files from the given paths and render a static
website to disk instead of starting an HTTP server.

When no paths are given, the command exports the built-in Glazed documentation
already loaded into the help system. When one or more paths are given, the
command clears any preloaded sections and exports only the sections discovered
from those explicit paths.

The output directory contains the frontend assets plus a static JSON data tree
that the SPA can browse without a live /api server.`),
			cmds.WithFlags(
				fields.New(
					"output-dir",
					fields.TypeString,
					fields.WithShortFlag("o"),
					fields.WithHelp("Directory where the static site will be written"),
					fields.WithDefault(DefaultOutputDir),
				),
				fields.New(
					"site-title",
					fields.TypeString,
					fields.WithHelp("Title written into the exported runtime config"),
					fields.WithDefault(DefaultSiteTitle),
				),
				fields.New(
					"base-path",
					fields.TypeString,
					fields.WithHelp("Optional base path to bake into the exported runtime config (for example /docs)"),
					fields.WithDefault(""),
				),
				fields.New(
					"data-dir",
					fields.TypeString,
					fields.WithHelp("Relative directory under the output root that stores exported JSON payloads"),
					fields.WithDefault(DefaultDataDir),
				),
				fields.New(
					"overwrite",
					fields.TypeBool,
					fields.WithHelp("Allow writing into a non-empty output directory"),
					fields.WithDefault(false),
				),
			),
			cmds.WithArguments(
				fields.New(
					"paths",
					fields.TypeStringList,
					fields.WithHelp("Markdown files or directories to load (default: embedded docs)"),
				),
			),
		),
		helpSystem: hs,
	}, nil
}

func (c *RenderSiteCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &RenderSiteSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to decode render-site settings")
	}

	return RenderSite(ctx, c.helpSystem, s)
}

func RenderSite(ctx context.Context, hs *help.HelpSystem, settings *RenderSiteSettings) error {
	if hs.Store == nil {
		return errors.New("HelpSystem.Store is nil")
	}
	if len(settings.Paths) > 0 {
		if err := helploader.ReplaceStoreWithPaths(ctx, hs, settings.Paths); err != nil {
			return err
		}
	}

	sections, err := hs.Store.Find(ctx, store.OrderByOrder())
	if err != nil {
		return errors.Wrap(err, "listing sections for site export")
	}
	sort.SliceStable(sections, func(i, j int) bool {
		if sections[i].Order != sections[j].Order {
			return sections[i].Order < sections[j].Order
		}
		return sections[i].Slug < sections[j].Slug
	})

	if err := prepareOutputDir(settings.OutputDir, settings.Overwrite); err != nil {
		return err
	}
	if err := exportFrontend(settings.OutputDir); err != nil {
		return err
	}

	dataRoot := filepath.Join(settings.OutputDir, settings.DataDir)
	if err := os.MkdirAll(filepath.Join(dataRoot, "sections"), 0o755); err != nil {
		return errors.Wrap(err, "creating site-data directory")
	}
	if err := os.MkdirAll(filepath.Join(dataRoot, "indexes"), 0o755); err != nil {
		return errors.Wrap(err, "creating site-data indexes directory")
	}

	dataBasePath := computeDataBasePath(settings.BasePath, settings.DataDir)
	summaries := make([]server.SectionSummary, 0, len(sections))
	topLevelSlugs := make([]string, 0)
	defaultSlugs := make([]string, 0)
	topics := map[string][]string{}
	commands := map[string][]string{}
	flags := map[string][]string{}

	for _, section := range sections {
		summaries = append(summaries, server.SummaryFromModel(section))
		if section.IsTopLevel {
			topLevelSlugs = append(topLevelSlugs, section.Slug)
		}
		if section.ShowPerDefault {
			defaultSlugs = append(defaultSlugs, section.Slug)
		}
		indexSlugs(topics, section.Topics, section.Slug)
		indexSlugs(commands, section.Commands, section.Slug)
		indexSlugs(flags, section.Flags, section.Slug)

		detail := server.DetailFromModel(section)
		if err := writeJSON(filepath.Join(dataRoot, "sections", section.Slug+".json"), detail); err != nil {
			return errors.Wrapf(err, "writing section payload for %q", section.Slug)
		}
	}

	sort.Strings(topLevelSlugs)
	sort.Strings(defaultSlugs)
	sortIndex(topics)
	sortIndex(commands)
	sortIndex(flags)

	listResponse := server.ListSectionsResponse{
		Sections: summaries,
		Total:    len(summaries),
		Limit:    -1,
		Offset:   0,
	}
	if err := writeJSON(filepath.Join(dataRoot, "sections.json"), listResponse); err != nil {
		return errors.Wrap(err, "writing sections list payload")
	}
	if err := writeJSON(filepath.Join(dataRoot, "health.json"), server.HealthResponse{
		OK:       true,
		Sections: len(summaries),
	}); err != nil {
		return errors.Wrap(err, "writing health payload")
	}
	if err := writeJSON(filepath.Join(dataRoot, "indexes", "topics.json"), SlugIndex{Items: topics}); err != nil {
		return errors.Wrap(err, "writing topic index")
	}
	if err := writeJSON(filepath.Join(dataRoot, "indexes", "commands.json"), SlugIndex{Items: commands}); err != nil {
		return errors.Wrap(err, "writing command index")
	}
	if err := writeJSON(filepath.Join(dataRoot, "indexes", "flags.json"), SlugIndex{Items: flags}); err != nil {
		return errors.Wrap(err, "writing flag index")
	}
	if err := writeJSON(filepath.Join(dataRoot, "indexes", "top-level.json"), topLevelSlugs); err != nil {
		return errors.Wrap(err, "writing top-level index")
	}
	if err := writeJSON(filepath.Join(dataRoot, "indexes", "defaults.json"), defaultSlugs); err != nil {
		return errors.Wrap(err, "writing defaults index")
	}
	if err := writeJSON(filepath.Join(dataRoot, "manifest.json"), Manifest{
		GeneratedBy:   "glaze render-site",
		SiteTitle:     settings.SiteTitle,
		TotalSections: len(summaries),
		DataBasePath:  dataBasePath,
		TopLevelSlugs: topLevelSlugs,
		DefaultSlugs:  defaultSlugs,
	}); err != nil {
		return errors.Wrap(err, "writing manifest")
	}

	if err := writeRuntimeConfig(filepath.Join(settings.OutputDir, "site-config.js"), RuntimeConfig{
		Mode:        "static",
		DataBaseURL: dataBasePath,
		SiteTitle:   settings.SiteTitle,
	}); err != nil {
		return err
	}

	log.Info().
		Str("output_dir", settings.OutputDir).
		Int("sections", len(summaries)).
		Msg("Rendered static help site")
	return nil
}

func prepareOutputDir(outputDir string, overwrite bool) error {
	info, err := os.Stat(outputDir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return os.MkdirAll(outputDir, 0o755)
	case err != nil:
		return errors.Wrap(err, "stat output directory")
	case !info.IsDir():
		return errors.Errorf("output path %q is not a directory", outputDir)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return errors.Wrap(err, "reading output directory")
	}
	if len(entries) > 0 && !overwrite {
		return errors.Errorf("output directory %q is not empty; rerun with --overwrite", outputDir)
	}
	if overwrite {
		for _, entry := range entries {
			if err := os.RemoveAll(filepath.Join(outputDir, entry.Name())); err != nil {
				return errors.Wrapf(err, "removing %q from output directory", entry.Name())
			}
		}
	}
	return nil
}

func exportFrontend(outputDir string) error {
	sub, err := fs.Sub(web.FS, "dist")
	if err != nil {
		return errors.Wrap(err, "opening embedded web dist")
	}

	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		target := filepath.Join(outputDir, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func writeRuntimeConfig(path string, config RuntimeConfig) error {
	payload, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshalling runtime config")
	}
	content := fmt.Sprintf("window.__GLAZE_SITE_CONFIG__ = %s;\n", string(payload))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return errors.Wrap(err, "writing runtime config")
	}
	return nil
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func indexSlugs(index map[string][]string, keys []string, slug string) {
	for _, key := range keys {
		index[key] = append(index[key], slug)
	}
}

func sortIndex(index map[string][]string) {
	for key, slugs := range index {
		sort.Strings(slugs)
		index[key] = slugs
	}
}

func computeDataBasePath(basePath, dataDir string) string {
	trimmedDataDir := strings.Trim(dataDir, "/")
	trimmedBasePath := strings.TrimSpace(basePath)

	if trimmedBasePath == "" || trimmedBasePath == "." {
		return "./" + trimmedDataDir
	}
	if strings.HasPrefix(trimmedBasePath, "/") {
		return strings.TrimRight(trimmedBasePath, "/") + "/" + trimmedDataDir
	}
	return strings.TrimRight(trimmedBasePath, "/") + "/" + trimmedDataDir
}
