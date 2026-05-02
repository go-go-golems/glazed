package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ContentLoader loads help sections from an external source into a HelpSystem.
type ContentLoader interface {
	Load(ctx context.Context, hs *help.HelpSystem) error
	String() string
}

// PackageRef identifies the help package/version assigned while importing a
// source. Version is empty for unversioned package exports.
type PackageRef struct {
	Name    string
	Version string
	Source  string
}

// DiscoveredSQLitePackage is one SQLite help export found under a directory
// scanned by SQLiteDirLoader.
type DiscoveredSQLitePackage struct {
	Path    string
	Package string
	Version string
}

func applyPackageRef(section *model.Section, ref PackageRef) {
	if section.PackageName == "" {
		section.PackageName = ref.Name
	}
	if section.PackageVersion == "" {
		section.PackageVersion = ref.Version
	}
}

func packageNameFromPath(path string) string {
	if path == "" || path == "-" {
		return ""
	}
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func packageNameFromBinary(binary string) string {
	return strings.TrimSuffix(filepath.Base(binary), filepath.Ext(binary))
}

// NormalizeStringList trims values and expands comma-separated entries.
func NormalizeStringList(values []string) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				ret = append(ret, part)
			}
		}
	}
	return ret
}

// MarkdownPathLoader loads markdown help sections from files or directories.
type MarkdownPathLoader struct {
	Paths []string
}

func (l *MarkdownPathLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	paths := NormalizeStringList(l.Paths)
	if err := LoadPaths(ctx, hs, paths); err != nil {
		return err
	}
	if len(paths) > 0 {
		if err := hs.Store.SetDefaultPackage(ctx, packageNameFromPath(paths[0]), ""); err != nil {
			return err
		}
	}
	return nil
}

func (l *MarkdownPathLoader) String() string {
	return "markdown paths: " + strings.Join(NormalizeStringList(l.Paths), ", ")
}

// JSONFileLoader loads help sections from JSON export files. A path of "-" reads stdin.
type JSONFileLoader struct {
	Paths []string
}

func (l *JSONFileLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	paths := NormalizeStringList(l.Paths)
	stdinCount := 0
	for _, path := range paths {
		if path == "-" {
			stdinCount++
		}
	}
	if stdinCount > 1 {
		return errors.New("--from-json - may only be used once")
	}

	for _, path := range paths {
		var r io.ReadCloser
		if path == "-" {
			r = io.NopCloser(os.Stdin)
		} else {
			f, err := os.Open(path)
			if err != nil {
				return errors.Wrapf(err, "open JSON source %s", path)
			}
			r = f
		}

		sections, err := DecodeSectionsJSON(r)
		closeErr := r.Close()
		if err != nil {
			return errors.Wrapf(err, "decode JSON source %s", path)
		}
		if closeErr != nil {
			return errors.Wrapf(closeErr, "close JSON source %s", path)
		}

		ref := PackageRef{Name: packageNameFromPath(path), Source: "json: " + path}
		for _, section := range sections {
			applyPackageRef(section, ref)
			if err := upsertWithCollisionLog(ctx, hs, section, "json: "+path); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *JSONFileLoader) String() string {
	return "json files: " + strings.Join(NormalizeStringList(l.Paths), ", ")
}

// SQLiteLoader loads help sections from exported Glazed help SQLite databases.
type SQLiteLoader struct {
	Paths []string
}

func (l *SQLiteLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	for _, path := range NormalizeStringList(l.Paths) {
		ref := PackageRef{Name: packageNameFromPath(path), Source: "sqlite: " + path}
		if err := loadSQLitePath(ctx, hs, path, ref); err != nil {
			return err
		}
	}
	return nil
}

func loadSQLitePath(ctx context.Context, hs *help.HelpSystem, path string, ref PackageRef) error {
	sourceStore, err := store.New(path)
	if err != nil {
		return errors.Wrapf(err, "open SQLite source %s", path)
	}

	sections, listErr := sourceStore.List(ctx, "")
	closeErr := sourceStore.Close()
	if listErr != nil {
		return errors.Wrapf(listErr, "list sections from SQLite source %s", path)
	}
	if closeErr != nil {
		return errors.Wrapf(closeErr, "close SQLite source %s", path)
	}

	for _, section := range sections {
		applyPackageRef(section, ref)
		if err := upsertWithCollisionLog(ctx, hs, section, ref.Source); err != nil {
			return err
		}
	}
	return nil
}

func (l *SQLiteLoader) String() string {
	return "sqlite files: " + strings.Join(NormalizeStringList(l.Paths), ", ")
}

// SQLiteDirLoader recursively scans directories for package/versioned SQLite
// help exports. Accepted layouts relative to each root are X.db, X/X.db, and
// X/Y/X.db, where X is the package name and Y is the optional version.
type SQLiteDirLoader struct {
	Roots []string
}

func (l *SQLiteDirLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	for _, root := range NormalizeStringList(l.Roots) {
		discovered, err := DiscoverSQLitePackages(root)
		if err != nil {
			return err
		}
		for _, d := range discovered {
			ref := PackageRef{Name: d.Package, Version: d.Version, Source: "sqlite-dir: " + d.Path}
			if err := loadSQLitePath(ctx, hs, d.Path, ref); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *SQLiteDirLoader) String() string {
	return "sqlite dirs: " + strings.Join(NormalizeStringList(l.Roots), ", ")
}

func DiscoverSQLitePackages(root string) ([]DiscoveredSQLitePackage, error) {
	var discovered []DiscoveredSQLitePackage
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isSQLiteHelpDB(path) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := splitPath(rel)
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		switch len(parts) {
		case 1:
			discovered = append(discovered, DiscoveredSQLitePackage{Path: path, Package: stem})
		case 2:
			pkg := parts[0]
			if stem == pkg {
				discovered = append(discovered, DiscoveredSQLitePackage{Path: path, Package: pkg})
			}
		case 3:
			pkg := parts[0]
			version := parts[1]
			if stem == pkg {
				discovered = append(discovered, DiscoveredSQLitePackage{Path: path, Package: pkg, Version: version})
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "scan SQLite help directory %s", root)
	}
	sort.Slice(discovered, func(i, j int) bool {
		if discovered[i].Package != discovered[j].Package {
			return discovered[i].Package < discovered[j].Package
		}
		if discovered[i].Version != discovered[j].Version {
			return discovered[i].Version < discovered[j].Version
		}
		return discovered[i].Path < discovered[j].Path
	})
	return discovered, nil
}

func isSQLiteHelpDB(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".db", ".sqlite":
		return true
	default:
		return false
	}
}

func splitPath(path string) []string {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
	ret := parts[:0]
	for _, part := range parts {
		if part != "." && part != "" {
			ret = append(ret, part)
		}
	}
	return ret
}

// GlazedCommandLoader runs '<binary> help export --output json' for each binary.
type GlazedCommandLoader struct {
	Binaries []string
}

func (l *GlazedCommandLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	for _, binary := range NormalizeStringList(l.Binaries) {
		data, err := runGlazedHelpExport(ctx, binary)
		if err != nil {
			return err
		}
		ref := PackageRef{Name: packageNameFromBinary(binary), Source: "glazed command: " + binary}
		if err := importJSONBytes(ctx, hs, data, ref); err != nil {
			return err
		}
	}
	return nil
}

func (l *GlazedCommandLoader) String() string {
	return "glazed commands: " + strings.Join(NormalizeStringList(l.Binaries), ", ")
}

func runGlazedHelpExport(ctx context.Context, binary string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, "help", "export", "--with-content=true", "--output", "json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "%s help export failed; stderr: %s", binary, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func importJSONBytes(ctx context.Context, hs *help.HelpSystem, data []byte, ref PackageRef) error {
	sections, err := DecodeSectionsJSON(bytes.NewReader(data))
	if err != nil {
		return errors.Wrapf(err, "decode %s", ref.Source)
	}
	for _, section := range sections {
		applyPackageRef(section, ref)
		if err := upsertWithCollisionLog(ctx, hs, section, ref.Source); err != nil {
			return err
		}
	}
	return nil
}

func upsertWithCollisionLog(ctx context.Context, hs *help.HelpSystem, section *model.Section, source string) error {
	if err := section.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section from %s", source)
	}
	if existing, err := hs.Store.GetByPackageSlug(ctx, section.PackageName, section.PackageVersion, section.Slug); err == nil && existing != nil {
		log.Warn().
			Str("package", section.PackageName).
			Str("version", section.PackageVersion).
			Str("slug", section.Slug).
			Str("source", source).
			Msg("Replacing existing help section")
	}
	if err := hs.Store.Upsert(ctx, section); err != nil {
		return errors.Wrapf(err, "upsert section %s from %s", section.Slug, source)
	}
	return nil
}

// DecodeSectionsJSON decodes Glazed help export JSON into model sections.
func DecodeSectionsJSON(r io.Reader) ([]*model.Section, error) {
	var rows []sectionImportRow
	if err := json.NewDecoder(r).Decode(&rows); err != nil {
		return nil, err
	}
	sections := make([]*model.Section, 0, len(rows))
	for i, row := range rows {
		section, err := row.ToSection()
		if err != nil {
			return nil, errors.Wrapf(err, "row %d", i)
		}
		if err := section.Validate(); err != nil {
			return nil, errors.Wrapf(err, "row %d", i)
		}
		sections = append(sections, section)
	}
	return sections, nil
}

type sectionImportRow struct {
	ID             int64           `json:"id,omitempty"`
	Slug           string          `json:"slug"`
	Type           json.RawMessage `json:"type,omitempty"`
	SectionType    json.RawMessage `json:"section_type,omitempty"`
	Title          string          `json:"title"`
	SubTitle       string          `json:"sub_title"`
	Short          string          `json:"short"`
	Content        string          `json:"content"`
	Topics         []string        `json:"topics"`
	Flags          []string        `json:"flags"`
	Commands       []string        `json:"commands"`
	PackageName    string          `json:"package_name,omitempty"`
	PackageNameAlt string          `json:"packageName,omitempty"`
	PackageVersion string          `json:"package_version,omitempty"`
	PackageVerAlt  string          `json:"packageVersion,omitempty"`
	IsTopLevel     bool            `json:"is_top_level"`
	IsTemplate     bool            `json:"is_template"`
	ShowPerDefault bool            `json:"show_per_default"`
	Order          int             `json:"order"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

func (r sectionImportRow) ToSection() (*model.Section, error) {
	st, err := parseSectionType(r.Type, r.SectionType)
	if err != nil {
		return nil, err
	}
	packageName := r.PackageName
	if packageName == "" {
		packageName = r.PackageNameAlt
	}
	packageVersion := r.PackageVersion
	if packageVersion == "" {
		packageVersion = r.PackageVerAlt
	}
	return &model.Section{
		ID:             r.ID,
		Slug:           r.Slug,
		SectionType:    st,
		PackageName:    packageName,
		PackageVersion: packageVersion,
		Title:          r.Title,
		SubTitle:       r.SubTitle,
		Short:          r.Short,
		Content:        r.Content,
		Topics:         r.Topics,
		Flags:          r.Flags,
		Commands:       r.Commands,
		IsTopLevel:     r.IsTopLevel,
		IsTemplate:     r.IsTemplate,
		ShowPerDefault: r.ShowPerDefault,
		Order:          r.Order,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}, nil
}

func parseSectionType(typeRaw, sectionTypeRaw json.RawMessage) (model.SectionType, error) {
	if len(typeRaw) > 0 && string(typeRaw) != "null" {
		return parseSectionTypeRaw(typeRaw)
	}
	if len(sectionTypeRaw) > 0 && string(sectionTypeRaw) != "null" {
		return parseSectionTypeRaw(sectionTypeRaw)
	}
	return model.SectionGeneralTopic, errors.New("missing type or section_type")
}

func parseSectionTypeRaw(raw json.RawMessage) (model.SectionType, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return model.SectionTypeFromString(s)
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		st := model.SectionType(n)
		switch st {
		case model.SectionGeneralTopic, model.SectionExample, model.SectionApplication, model.SectionTutorial:
			return st, nil
		default:
			return model.SectionGeneralTopic, fmt.Errorf("unknown section type number %d", n)
		}
	}

	return model.SectionGeneralTopic, fmt.Errorf("invalid section type %s", string(raw))
}
