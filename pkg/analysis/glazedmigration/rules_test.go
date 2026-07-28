package glazedmigration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/settings"
	"golang.org/x/tools/go/analysis"
)

// runAnalyzer parses source and runs the analyzer, returning the diagnostics.
func runAnalyzer(t *testing.T, source string) []analysis.Diagnostic {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "input.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Analyzer: Analyzer,
		Fset:     fset,
		Files:    []*ast.File{file},
		TypesInfo: &types.Info{
			Uses: map[*ast.Ident]types.Object{},
		},
		Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
	}
	if _, err := run(pass); err != nil {
		t.Fatal(err)
	}
	return diagnostics
}

// countFixes returns the total number of suggested fixes across diagnostics.
func countFixes(diagnostics []analysis.Diagnostic) int {
	n := 0
	for _, d := range diagnostics {
		n += len(d.SuggestedFixes)
	}
	return n
}

// fixTexts returns the NewText of the first edit of each suggested fix.
func fixTexts(diagnostics []analysis.Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		for _, f := range d.SuggestedFixes {
			for _, e := range f.TextEdits {
				out = append(out, string(e.NewText))
			}
		}
	}
	return out
}

// TestR2NewGlazedSectionRename verifies that NewGlazedSection (the missing twin
// of NewGlazedSchema) is renamed to NewStructuredOutputSection.
func TestR2NewGlazedSectionRename(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantDiag  int
		wantFixes int
		wantText  string
	}{
		{
			name:      "qualified zero-arg",
			source:    `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { settings.NewGlazedSection() }`,
			wantDiag:  1,
			wantFixes: 1,
			wantText:  newConstructor,
		},
		{
			name:      "aliased zero-arg",
			source:    `package p; import gs "github.com/go-go-golems/glazed/pkg/settings"; func f() { gs.NewGlazedSection() }`,
			wantDiag:  1,
			wantFixes: 1,
			wantText:  newConstructor,
		},
		{
			name:      "dot-import zero-arg",
			source:    `package p; import . "github.com/go-go-golems/glazed/pkg/settings"; func f() { NewGlazedSection() }`,
			wantDiag:  1,
			wantFixes: 1,
			wantText:  newConstructor,
		},
		{
			name:      "dot-import shadowed replacement",
			source:    `package p; import . "github.com/go-go-golems/glazed/pkg/settings"; func NewStructuredOutputSection() {}; func f() { NewGlazedSection() }`,
			wantDiag:  1,
			wantFixes: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := runAnalyzer(t, tt.source)
			if len(diag) != tt.wantDiag {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(diag), tt.wantDiag, diag)
			}
			if got := countFixes(diag); got != tt.wantFixes {
				t.Fatalf("got %d fixes, want %d", got, tt.wantFixes)
			}
			if tt.wantFixes > 0 {
				texts := fixTexts(diag)
				if !contains(texts, tt.wantText) {
					t.Fatalf("fix texts %v do not contain %q", texts, tt.wantText)
				}
			}
		})
	}
}

// TestR3WrapperUnwrap verifies that With*SectionOptions wrappers are unwrapped.
func TestR3WrapperUnwrap(t *testing.T) {
	source := `package p
import (
	"github.com/go-go-golems/glazed/pkg/settings"
"github.com/go-go-golems/glazed/pkg/cmds/schema"
)
func f() {
	_ = settings.NewGlazedSection(
		settings.WithOutputSectionOptions(
			schema.WithDefaults(map[string]interface{}{"output": "json"}),
		),
	)
}`
	diag := runAnalyzer(t, source)
	// Expect: one constructor diagnostic (rename, no fix because args present),
	// one wrapper-unwrap diagnostic (with fix), one key-rename diagnostic (with fix).
	if len(diag) < 2 {
		t.Fatalf("got %d diagnostics, want at least 2: %+v", len(diag), diag)
	}
	// The wrapper unwrap should produce a fix whose text contains the inner arg.
	foundUnwrap := false
	for _, d := range diag {
		for _, f := range d.SuggestedFixes {
			for _, e := range f.TextEdits {
				if strings.Contains(string(e.NewText), "WithDefaults") {
					foundUnwrap = true
				}
			}
		}
	}
	if !foundUnwrap {
		t.Fatalf("expected a wrapper-unwrap fix containing the inner WithDefaults arg, got: %+v", diag)
	}
}

// TestR4KeyRename verifies that the "output" default-map key is renamed to
// "format", and that unsupported values are reported.
func TestR4KeyRename(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		wantKeyFix  bool
		wantValueDiag bool
	}{
		{"json supported", "json", true, false},
		{"table supported", "table", true, false},
		{"yaml supported", "yaml", true, false},
		{"sql unsupported", "sql", true, true},
		{"excel unsupported", "excel", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := `package p
import "github.com/go-go-golems/glazed/pkg/cmds/schema"
func f() {
	_ = schema.WithDefaults(map[string]interface{}{"output": "` + tt.value + `"})
}`
			diag := runAnalyzer(t, source)
			foundKeyFix := false
			foundValueDiag := false
			for _, d := range diag {
				if strings.Contains(d.Message, "rename default-map key") && len(d.SuggestedFixes) > 0 {
					foundKeyFix = true
				}
				if strings.Contains(d.Message, "not a supported structured-output format") {
					foundValueDiag = true
				}
			}
			if foundKeyFix != tt.wantKeyFix {
				t.Fatalf("key fix: got %v, want %v", foundKeyFix, tt.wantKeyFix)
			}
			if foundValueDiag != tt.wantValueDiag {
				t.Fatalf("value diagnostic: got %v, want %v", foundValueDiag, tt.wantValueDiag)
			}
		})
	}
}

// TestR5SlugRename verifies that settings.GlazedSlug is renamed to
// StructuredOutputSlug.
func TestR5SlugRename(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "qualified",
			source: `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { _ = settings.GlazedSlug }`,
		},
		{
			name:   "aliased",
			source: `package p; import s "github.com/go-go-golems/glazed/pkg/settings"; func f() { _ = s.GlazedSlug }`,
		},
		{
			name:   "dot-import",
			source: `package p; import . "github.com/go-go-golems/glazed/pkg/settings"; func f() { _ = GlazedSlug }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := runAnalyzer(t, tt.source)
			if len(diag) != 1 {
				t.Fatalf("got %d diagnostics, want 1: %+v", len(diag), diag)
			}
			if countFixes(diag) != 1 {
				t.Fatalf("got %d fixes, want 1", countFixes(diag))
			}
			texts := fixTexts(diag)
			if !contains(texts, newSlug) {
				t.Fatalf("fix texts %v do not contain %q", texts, newSlug)
			}
		})
	}
}

// TestR6R7R8SetupHelpers verifies that the removed Setup* helpers produce
// report-only diagnostics (no fixes).
func TestR6R7R8SetupHelpers(t *testing.T) {
	helpers := []string{
		"SetupTableProcessor",
		"SetupProcessorOutput",
		"SetupTableOutputFormatter",
		"SetupRowOutputFormatter",
		"SetupSimpleTableProcessor",
		"NewOutputFormatterSettings",
	}
	for _, h := range helpers {
		t.Run(h, func(t *testing.T) {
			source := `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { _, _ = settings.` + h + `(nil) }`
			diag := runAnalyzer(t, source)
			if len(diag) != 1 {
				t.Fatalf("got %d diagnostics, want 1: %+v", len(diag), diag)
			}
			if countFixes(diag) != 0 {
				t.Fatalf("expected no fixes for report-only helper %s, got %d", h, countFixes(diag))
			}
			if !strings.Contains(diag[0].Message, setupHelpers[h]) {
				t.Fatalf("diagnostic %q does not mention replacement %q", diag[0].Message, setupHelpers[h])
			}
		})
	}
}

// TestR9RemovedFeatureSections verifies that removed feature section
// constructors produce report-only diagnostics.
func TestR9RemovedFeatureSections(t *testing.T) {
	for ctor := range removedFeatureSectionConstructors {
		t.Run(ctor, func(t *testing.T) {
			source := `package p; import "github.com/go-go-golems/glazed/pkg/settings"; func f() { _, _ = settings.` + ctor + `() }`
			diag := runAnalyzer(t, source)
			if len(diag) != 1 {
				t.Fatalf("got %d diagnostics, want 1: %+v", len(diag), diag)
			}
			if countFixes(diag) != 0 {
				t.Fatalf("expected no fixes for removed feature %s, got %d", ctor, countFixes(diag))
			}
			if !strings.Contains(diag[0].Message, "no mechanical migration") {
				t.Fatalf("diagnostic %q does not mention manual redesign", diag[0].Message)
			}
		})
	}
}

// TestOtherPackageNotMigrated verifies that selectors from a non-Glazed
// settings package are not touched.
func TestOtherPackageNotMigrated(t *testing.T) {
	source := `package p; import other "example.com/settings"; func f() { other.NewGlazedSection(); other.GlazedSlug; other.SetupTableProcessor(nil) }`
	diag := runAnalyzer(t, source)
	if len(diag) != 0 {
		t.Fatalf("got %d diagnostics for a foreign package, want 0: %+v", len(diag), diag)
	}
}

// TestNoSettingsImport verifies that files without the settings import are
// skipped entirely.
func TestNoSettingsImport(t *testing.T) {
	source := `package p; import "fmt"; func f() { fmt.Println("hi") }`
	diag := runAnalyzer(t, source)
	if len(diag) != 0 {
		t.Fatalf("got %d diagnostics, want 0: %+v", len(diag), diag)
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// TestStructuredOutputFormatsExported verifies the settings accessor returns
// the expected format set, guarding against drift in the R4 allowlist.
func TestStructuredOutputFormatsExported(t *testing.T) {
	got := settings.StructuredOutputFormats()
	want := []string{"table", "json", "jsonl", "csv", "tsv", "yaml"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StructuredOutputFormats() = %v, want %v", got, want)
	}
}
