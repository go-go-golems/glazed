package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ExportCommand exports help sections as structured data or to disk.
type ExportCommand struct {
	*cmds.CommandDescription
	helpSystem *help.HelpSystem
}

var _ cmds.BareCommand = (*ExportCommand)(nil)

// ExportSettings holds parsed flag values for the export command.
type ExportSettings struct {
	Type        string `glazed:"type"`
	Topic       string `glazed:"topic"`
	Command     string `glazed:"command"`
	Flag        string `glazed:"flag"`
	Slug        string `glazed:"slug"`
	WithContent bool   `glazed:"with-content"`
	Format      string `glazed:"format"`
	OutputPath  string `glazed:"output-path"`
	FlattenDirs bool   `glazed:"flatten-dirs"`
}

// NewExportCommand creates a new export command.
func NewExportCommand(hs *help.HelpSystem) (*ExportCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create glazed schema")
	}

	return &ExportCommand{
		CommandDescription: cmds.NewCommandDescription(
			"export",
			cmds.WithShort("Export help sections to external formats"),
			cmds.WithLong(`Export help sections as structured data or to disk.

By default, exports all matching sections as JSON/CSV/table/YAML via the Glazed
processor, including full markdown content (--with-content defaults to true).

Use --format files to write individual .md files, or --format sqlite to
produce a portable SQLite database. Use --with-content=false for lightweight
metadata-only exports.
`),
			cmds.WithFlags(
				fields.New("type", fields.TypeString, fields.WithHelp("Filter by section type"), fields.WithDefault("")),
				fields.New("topic", fields.TypeString, fields.WithHelp("Filter by topic"), fields.WithDefault("")),
				fields.New("command", fields.TypeString, fields.WithHelp("Filter by command"), fields.WithDefault("")),
				fields.New("flag", fields.TypeString, fields.WithHelp("Filter by flag"), fields.WithDefault("")),
				fields.New("slug", fields.TypeString, fields.WithHelp("Filter by slug(s), comma-separated"), fields.WithDefault("")),
				fields.New("with-content", fields.TypeBool, fields.WithHelp("Include content field in tabular output"), fields.WithDefault(true)),
				fields.New("format", fields.TypeString, fields.WithHelp("Export mode: glazed, files, sqlite"), fields.WithDefault("glazed")),
				fields.New("output-path", fields.TypeString, fields.WithHelp("Output path for files/sqlite mode"), fields.WithDefault("")),
				fields.New("flatten-dirs", fields.TypeBool, fields.WithHelp("Flatten directory structure in files mode"), fields.WithDefault(false)),
			),
			cmds.WithSchema(schema.NewSchema(schema.WithSections(glazedSection))),
		),
		helpSystem: hs,
	}, nil
}

// Run executes the export command.
func (c *ExportCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &ExportSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to decode export settings")
	}

	// Validate format
	switch s.Format {
	case "glazed", "files", "sqlite":
		// OK
	default:
		return errors.Errorf("unknown format: %s (expected glazed, files, or sqlite)", s.Format)
	}

	predicate := buildExportPredicate(s)
	sections, err := c.helpSystem.Store.Find(ctx, predicate)
	if err != nil {
		return errors.Wrap(err, "failed to query sections")
	}

	switch s.Format {
	case "glazed":
		return c.runGlazed(ctx, parsedValues, sections, s)
	case "files":
		return exportToFiles(sections, s)
	case "sqlite":
		return exportToSQLite(ctx, sections, s)
	}

	return nil
}

// runGlazed emits sections through the Glaze processor.
func (c *ExportCommand) runGlazed(ctx context.Context, parsedValues *values.Values, sections []*model.Section, s *ExportSettings) error {
	glazedValues, ok := parsedValues.Get(settings.GlazedSlug)
	if !ok {
		return errors.New("glazed section not found in parsed values")
	}

	gp, err := settings.SetupTableProcessor(glazedValues)
	if err != nil {
		return errors.Wrap(err, "failed to setup table processor")
	}

	_, err = settings.SetupProcessorOutput(gp, glazedValues, os.Stdout)
	if err != nil {
		return errors.Wrap(err, "failed to setup processor output")
	}

	for _, section := range sections {
		row := types.NewRow(
			types.MRP("id", section.ID),
			types.MRP("slug", section.Slug),
			types.MRP("type", section.SectionType.String()),
			types.MRP("title", section.Title),
			types.MRP("short", section.Short),
			types.MRP("topics", section.Topics),
			types.MRP("flags", section.Flags),
			types.MRP("commands", section.Commands),
			types.MRP("is_top_level", section.IsTopLevel),
			types.MRP("show_per_default", section.ShowPerDefault),
			types.MRP("order", section.Order),
			types.MRP("created_at", section.CreatedAt),
			types.MRP("updated_at", section.UpdatedAt),
		)
		if s.WithContent {
			row.Set("content", section.Content)
		}
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "failed to add row")
		}
	}

	if err := gp.Close(ctx); err != nil {
		return errors.Wrap(err, "failed to close processor")
	}

	return nil
}

// buildExportPredicate constructs a store.Predicate from export settings.
func buildExportPredicate(s *ExportSettings) store.Predicate {
	var preds []store.Predicate

	if s.Type != "" {
		st, err := model.SectionTypeFromString(s.Type)
		if err == nil {
			preds = append(preds, store.IsType(st))
		}
	}
	if s.Topic != "" {
		preds = append(preds, store.HasTopic(s.Topic))
	}
	if s.Command != "" {
		preds = append(preds, store.HasCommand(s.Command))
	}
	if s.Flag != "" {
		preds = append(preds, store.HasFlag(s.Flag))
	}
	if s.Slug != "" {
		slugs := strings.Split(s.Slug, ",")
		for i := range slugs {
			slugs[i] = strings.TrimSpace(slugs[i])
		}
		if len(slugs) == 1 {
			preds = append(preds, store.SlugEquals(slugs[0]))
		} else {
			preds = append(preds, store.SlugIn(slugs))
		}
	}

	base := store.OrderByOrder()
	if len(preds) == 0 {
		return base
	}
	return store.And(store.And(preds...), base)
}

// exportToFiles writes each section as an individual markdown file.
func exportToFiles(sections []*model.Section, s *ExportSettings) error {
	outDir := s.OutputPath
	if outDir == "" {
		outDir = "./help-export"
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	for _, section := range sections {
		dir := outDir
		if !s.FlattenDirs {
			dir = filepath.Join(outDir, slugify(section.SectionType.String()))
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.Wrapf(err, "failed to create directory %s", dir)
			}
		}

		path := filepath.Join(dir, section.Slug+".md")
		data, err := reconstructMarkdown(section)
		if err != nil {
			return errors.Wrapf(err, "failed to reconstruct markdown for %s", section.Slug)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return errors.Wrapf(err, "failed to write file %s", path)
		}
	}

	return nil
}

// exportToSQLite writes sections to a new SQLite database.
func exportToSQLite(ctx context.Context, sections []*model.Section, s *ExportSettings) error {
	dbPath := s.OutputPath
	if dbPath == "" {
		dbPath = "./help-export.sqlite"
	}
	if !strings.HasSuffix(dbPath, ".sqlite") && !strings.HasSuffix(dbPath, ".db") {
		dbPath += ".sqlite"
	}

	// Remove existing file to start fresh
	_ = os.Remove(dbPath)

	exportStore, err := store.New(dbPath)
	if err != nil {
		return errors.Wrap(err, "failed to create SQLite store")
	}
	defer func() {
		_ = exportStore.Close()
	}()

	for _, section := range sections {
		if err := exportStore.Upsert(ctx, section); err != nil {
			return errors.Wrapf(err, "failed to upsert section %s", section.Slug)
		}
	}

	return nil
}

// reconstructMarkdown rebuilds a markdown file from a Section.
func reconstructMarkdown(section *model.Section) ([]byte, error) {
	frontmatter := map[string]interface{}{
		"Title":          section.Title,
		"Slug":           section.Slug,
		"Short":          section.Short,
		"SectionType":    section.SectionType.String(),
		"Topics":         section.Topics,
		"Flags":          section.Flags,
		"Commands":       section.Commands,
		"IsTopLevel":     section.IsTopLevel,
		"IsTemplate":     section.IsTemplate,
		"ShowPerDefault": section.ShowPerDefault,
		"Order":          section.Order,
	}
	if section.SubTitle != "" {
		frontmatter["SubTitle"] = section.SubTitle
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(frontmatter); err != nil {
		return nil, err
	}
	_ = enc.Close()
	buf.WriteString("---\n")
	buf.WriteString(section.Content)
	return buf.Bytes(), nil
}

// slugify converts a string to a filesystem-friendly slug.
func slugify(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

// AddExportCommand wires the export command into the help subcommand tree.
func AddExportCommand(helpCmd *cobra.Command, hs *help.HelpSystem) error {
	exportGlazedCmd, err := NewExportCommand(hs)
	if err != nil {
		return err
	}
	exportCobraCmd, err := cli.BuildCobraCommand(exportGlazedCmd)
	if err != nil {
		return err
	}
	helpCmd.AddCommand(exportCobraCmd)
	return nil
}
