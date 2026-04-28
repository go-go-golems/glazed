package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
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

// NormalizeCommandList trims command values without comma-splitting them.
// Command strings may legitimately contain commas inside arguments.
func NormalizeCommandList(values []string) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			ret = append(ret, value)
		}
	}
	return ret
}

// MarkdownPathLoader loads markdown help sections from files or directories.
type MarkdownPathLoader struct {
	Paths []string
}

func (l *MarkdownPathLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	return LoadPaths(ctx, hs, NormalizeStringList(l.Paths))
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

		for _, section := range sections {
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
			if err := upsertWithCollisionLog(ctx, hs, section, "sqlite: "+path); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *SQLiteLoader) String() string {
	return "sqlite files: " + strings.Join(NormalizeStringList(l.Paths), ", ")
}

// CommandJSONLoader runs commands whose stdout is a JSON help export.
type CommandJSONLoader struct {
	Commands []string
}

func (l *CommandJSONLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
	for _, command := range NormalizeCommandList(l.Commands) {
		data, err := runCommandForJSON(ctx, command)
		if err != nil {
			return err
		}
		if err := importJSONBytes(ctx, hs, data, "command: "+command); err != nil {
			return err
		}
	}
	return nil
}

func (l *CommandJSONLoader) String() string {
	return "commands: " + strings.Join(NormalizeCommandList(l.Commands), ", ")
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
		if err := importJSONBytes(ctx, hs, data, "glazed command: "+binary); err != nil {
			return err
		}
	}
	return nil
}

func (l *GlazedCommandLoader) String() string {
	return "glazed commands: " + strings.Join(NormalizeStringList(l.Binaries), ", ")
}

func runCommandForJSON(ctx context.Context, command string) ([]byte, error) {
	args, err := tokenizeCommand(command)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, errors.New("empty command")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "command %q failed; stderr: %s", command, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
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

func importJSONBytes(ctx context.Context, hs *help.HelpSystem, data []byte, source string) error {
	sections, err := DecodeSectionsJSON(bytes.NewReader(data))
	if err != nil {
		return errors.Wrapf(err, "decode %s", source)
	}
	for _, section := range sections {
		if err := upsertWithCollisionLog(ctx, hs, section, source); err != nil {
			return err
		}
	}
	return nil
}

func upsertWithCollisionLog(ctx context.Context, hs *help.HelpSystem, section *model.Section, source string) error {
	if err := section.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section from %s", source)
	}
	if existing, err := hs.Store.GetBySlug(ctx, section.Slug); err == nil && existing != nil {
		log.Warn().Str("slug", section.Slug).Str("source", source).Msg("Replacing existing help section")
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
	return &model.Section{
		ID:             r.ID,
		Slug:           r.Slug,
		SectionType:    st,
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

func tokenizeCommand(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case ' ', '\t', '\n', '\r':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote %s", strconv.QuoteRune(quote))
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}
