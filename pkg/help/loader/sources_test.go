package loader

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

func TestNormalizeStringList(t *testing.T) {
	got := NormalizeStringList([]string{"pinocchio,sqleton", " xxx ", "", "a, b"})
	want := []string{"pinocchio", "sqleton", "xxx", "a", "b"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestDecodeSectionsJSON_CurrentExportTypeField(t *testing.T) {
	input := `[{
		"slug":"help-system",
		"title":"Help System",
		"type":"GeneralTopic",
		"content":"body",
		"topics":["help"],
		"flags":["topic"],
		"commands":["help"],
		"is_top_level":true,
		"show_per_default":true,
		"order":7
	}]`
	sections, err := DecodeSectionsJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("DecodeSectionsJSON: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	s := sections[0]
	if s.SectionType != model.SectionGeneralTopic {
		t.Fatalf("expected GeneralTopic, got %s", s.SectionType.String())
	}
	if s.Slug != "help-system" || s.Title != "Help System" || s.Order != 7 {
		t.Fatalf("unexpected section: %#v", s)
	}
	if len(s.Topics) != 1 || s.Topics[0] != "help" {
		t.Fatalf("topics not preserved: %#v", s.Topics)
	}
}

func TestDecodeSectionsJSON_SectionTypeStringField(t *testing.T) {
	input := `[{"slug":"tutorial","title":"Tutorial","section_type":"Tutorial"}]`
	sections, err := DecodeSectionsJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("DecodeSectionsJSON: %v", err)
	}
	if sections[0].SectionType != model.SectionTutorial {
		t.Fatalf("expected Tutorial, got %s", sections[0].SectionType.String())
	}
}

func TestDecodeSectionsJSON_SectionTypeNumericField(t *testing.T) {
	input := `[{"slug":"application","title":"Application","section_type":2}]`
	sections, err := DecodeSectionsJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("DecodeSectionsJSON: %v", err)
	}
	if sections[0].SectionType != model.SectionApplication {
		t.Fatalf("expected Application, got %s", sections[0].SectionType.String())
	}
}

func TestDecodeSectionsJSON_MissingTypeFails(t *testing.T) {
	input := `[{"slug":"missing","title":"Missing"}]`
	_, err := DecodeSectionsJSON(strings.NewReader(input))
	if err == nil {
		t.Fatalf("expected missing type error")
	}
}

func TestJSONFileLoader_LoadsSections(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()
	path := filepath.Join(t.TempDir(), "help.json")
	input := `[{"slug":"json-topic","title":"JSON Topic","type":"Example"}]`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("write JSON: %v", err)
	}

	loader := &JSONFileLoader{Paths: []string{path}}
	if err := loader.Load(ctx, hs); err != nil {
		t.Fatalf("loader.Load: %v", err)
	}
	section, err := hs.Store.GetBySlug(ctx, "json-topic")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if section.SectionType != model.SectionExample {
		t.Fatalf("expected Example, got %s", section.SectionType.String())
	}
}

func TestSQLiteLoader_LoadsSections(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "help.db")
	source, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	if err := source.Upsert(ctx, &model.Section{Slug: "sqlite-topic", Title: "SQLite Topic", SectionType: model.SectionApplication}); err != nil {
		t.Fatalf("source.Upsert: %v", err)
	}
	if err := source.Close(); err != nil {
		t.Fatalf("source.Close: %v", err)
	}

	hs := help.NewHelpSystem()
	loader := &SQLiteLoader{Paths: []string{dbPath}}
	if err := loader.Load(ctx, hs); err != nil {
		t.Fatalf("loader.Load: %v", err)
	}
	section, err := hs.Store.GetBySlug(ctx, "sqlite-topic")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if section.SectionType != model.SectionApplication {
		t.Fatalf("expected Application, got %s", section.SectionType.String())
	}
}

func TestCommandJSONLoader_LoadsSections(t *testing.T) {
	ctx := context.Background()
	hs := help.NewHelpSystem()
	jsonPayload, err := json.Marshal([]map[string]any{{
		"slug":  "cmd-topic",
		"title": "Command Topic",
		"type":  "Tutorial",
	}})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	loader := &CommandJSONLoader{Commands: []string{"printf " + strconvQuote(string(jsonPayload))}}
	if err := loader.Load(ctx, hs); err != nil {
		t.Fatalf("loader.Load: %v", err)
	}
	section, err := hs.Store.GetBySlug(ctx, "cmd-topic")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if section.SectionType != model.SectionTutorial {
		t.Fatalf("expected Tutorial, got %s", section.SectionType.String())
	}
}

func TestTokenizeCommand_Quotes(t *testing.T) {
	args, err := tokenizeCommand(`tool help export --query "topic:hello world"`)
	if err != nil {
		t.Fatalf("tokenizeCommand: %v", err)
	}
	want := []string{"tool", "help", "export", "--query", "topic:hello world"}
	if len(args) != len(want) {
		t.Fatalf("expected %v, got %v", want, args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, args)
		}
	}
}

func TestTokenizeCommand_UnterminatedQuote(t *testing.T) {
	_, err := tokenizeCommand(`tool "oops`)
	if err == nil {
		t.Fatalf("expected unterminated quote error")
	}
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
