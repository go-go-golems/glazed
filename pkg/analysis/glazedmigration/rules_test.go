package glazedmigration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
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
		name          string
		value         string
		wantKeyFix    bool
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
import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
)
func f() {
	_, _ = settings.NewStructuredOutputSection(schema.WithDefaults(map[string]interface{}{"output": "` + tt.value + `"}))
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
			// Dot imports are recognized in degraded (no type-checker) mode via
			// go/parser's same-file object resolution; local declarations still
			// carry Ident.Obj and remain protected.
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
			// Each helper recommends a specific replacement. SetupTableProcessor
			// standalone (no SetupProcessorOutput in scope) recommends
			// SetupStructuredProcessor; the pair case recommends SetupStructuredOutput.
			want := setupHelpers[h]
			if h == "SetupTableProcessor" {
				want = "SetupStructuredProcessor"
			}
			if !strings.Contains(diag[0].Message, want) {
				t.Fatalf("diagnostic %q does not mention replacement %q", diag[0].Message, want)
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
// TestR4FieldNameGetField verifies that GetField(GlazedSlug, "output") is
// renamed to GetField(StructuredOutputSlug, "format").
func TestR4FieldNameGetField(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
func f(parsedValues interface{}) {
	_, _ = parsedValues.(interface{ GetField(string, string) interface{} }).GetField(settings.GlazedSlug, "output")
}`
	diag := runAnalyzer(t, source)
	found := false
	for _, d := range diag {
		if strings.Contains(d.Message, "rename field name \"output\" to \"format\"") && len(d.SuggestedFixes) > 0 {
			found = true
			// Verify the fix text is "format".
			for _, f := range d.SuggestedFixes {
				for _, e := range f.TextEdits {
					if string(e.NewText) != strconv.Quote("format") {
						t.Fatalf("fix text = %q, want %q", string(e.NewText), strconv.Quote("format"))
					}
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected a field-name rename diagnostic, got: %+v", diag)
	}
}

// TestR4FieldNameGetParameter verifies GetParameter(GlazedSlug, "output") is
// renamed.
func TestR4FieldNameGetParameter(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
func f(layers interface{}) {
	_, _ = layers.(interface{ GetParameter(string, string) interface{} }).GetParameter(settings.GlazedSlug, "output")
}`
	diag := runAnalyzer(t, source)
	found := false
	for _, d := range diag {
		if strings.Contains(d.Message, "rename field name") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a GetParameter field-name rename diagnostic, got: %+v", diag)
	}
}

// TestR4FieldNameUpdateExistingValue verifies UpdateExistingValue("output", ...)
// is renamed.
func TestR4FieldNameUpdateExistingValue(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
var _ = settings.GlazedSlug
func f(glazedLayer interface{}) {
	_, _ = glazedLayer.(interface{ UpdateExistingValue(string, string, ...interface{}) interface{} }).UpdateExistingValue("output", "table")
}`
	diag := runAnalyzer(t, source)
	found := false
	for _, d := range diag {
		if strings.Contains(d.Message, "UpdateExistingValue") && strings.Contains(d.Message, "format") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an UpdateExistingValue field-name rename diagnostic, got: %+v", diag)
	}
}

// TestR4FieldNameFromMap verifies the sources.FromMap(map{GlazedSlug: {"output":...}})
// pattern renames the inner "output" key to "format".
func TestR4FieldNameFromMap(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
func f() {
	_ = map[string]map[string]interface{}{
		settings.GlazedSlug: {
			"output": "json",
		},
	}
}`
	diag := runAnalyzer(t, source)
	found := false
	for _, d := range diag {
		if strings.Contains(d.Message, "rename default-map key \"output\" to \"format\"") && len(d.SuggestedFixes) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a FromMap inner-key rename diagnostic, got: %+v", diag)
	}
}

// TestR4StandaloneWithDefaultsNotRenamed verifies that a schema.WithDefaults
// call with an "output" key is NOT renamed when it is not an argument to a
// structured-output constructor, because application-defined sections may
// legitimately define a field named "output".
func TestR4StandaloneWithDefaultsNotRenamed(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/cmds/schema"
func f() {
	_ = schema.WithDefaults(map[string]interface{}{"output": "json"})
}`
	diag := runAnalyzer(t, source)
	for _, d := range diag {
		if strings.Contains(d.Message, "rename default-map key") {
			t.Fatalf("standalone WithDefaults should not be renamed, got: %s", d.Message)
		}
	}
}

// TestR3FeatureWrapperReportedNotUnwrapped verifies that feature-section
// option wrappers are reported for manual redesign and NOT auto-unwrapped.
func TestR3FeatureWrapperReportedNotUnwrapped(t *testing.T) {
	for w := range featureSectionOptionWrappers {
		t.Run(w, func(t *testing.T) {
			source := `package p
import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
)
func f() {
	_, _ = settings.NewGlazedSection(settings.` + w + `(schema.WithDefaults(map[string]interface{}{"x": 1})))
}`
			diag := runAnalyzer(t, source)
			foundReport := false
			for _, d := range diag {
				if strings.Contains(d.Message, "targets a removed feature subsection") {
					foundReport = true
				}
				if strings.Contains(d.Message, "unwrap removed settings") && len(d.SuggestedFixes) > 0 {
					t.Fatalf("feature wrapper %s should not be auto-unwrapped", w)
				}
			}
			if !foundReport {
				t.Fatalf("expected a report-only diagnostic for feature wrapper %s, got: %+v", w, diag)
			}
		})
	}
}

// TestR2NewOutputSectionRename verifies that the removed NewOutputSection
// constructor is also migrated to NewStructuredOutputSection.
func TestR2NewOutputSectionRename(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
func f() { _, _ = settings.NewOutputSection() }`
	diag := runAnalyzer(t, source)
	if len(diag) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %+v", len(diag), diag)
	}
	if countFixes(diag) != 1 {
		t.Fatalf("got %d fixes, want 1", countFixes(diag))
	}
	if !contains(fixTexts(diag), newConstructor) {
		t.Fatalf("fix does not rename to %s", newConstructor)
	}
}

// TestR4MapIndexReported verifies that expr["output"] map index expressions are
// reported (with a fix) when they may reference the structured-output field.
func TestR4MapIndexReported(t *testing.T) {
	source := `package p
import "github.com/go-go-golems/glazed/pkg/settings"
func f() {
	var m map[string]interface{}
	_ = m["output"]
}`
	diag := runAnalyzer(t, source)
	found := false
	for _, d := range diag {
		if strings.Contains(d.Message, "map key \"output\"") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a map-index diagnostic, got: %+v", diag)
	}
}

func TestNoSettingsImport(t *testing.T) {
	source := `package p; import "fmt"; func f() { fmt.Println("hi") }`
	diag := runAnalyzer(t, source)
	if len(diag) != 0 {
		t.Fatalf("got %d diagnostics, want 0: %+v", len(diag), diag)
	}
}

// TestR1DeletionWithheldForMultipleUses verifies that R1 deletion is withheld
// (falling back to rename) when the variable is used in multiple WithSections
// calls within the same function, which would leave dangling references.
func TestR1DeletionWithheldForMultipleUses(t *testing.T) {
	source := "package p\n" +
		"import (\n" +
		"\t\"github.com/go-go-golems/glazed/pkg/cmds\"\n" +
		"\t\"github.com/go-go-golems/glazed/pkg/settings\"\n" +
		")\n" +
		"type C struct{ *cmds.CommandDescription }\n" +
		"func (c *C) RunIntoGlazeProcessor() {}\n" +
		"func f() (*C, error) {\n" +
		"\tglazedLayer, err := settings.NewGlazedSection()\n" +
		"\tif err != nil { return nil, err }\n" +
		"\t_ = cmds.WithSections(glazedLayer)\n" +
		"\treturn &C{CommandDescription: cmds.NewCommandDescription(\"x\", cmds.WithSections(glazedLayer))}, nil\n" +
		"}\n"
	diag := runAnalyzer(t, source)
	foundDeletion := false
	for _, d := range diag {
		if strings.Contains(d.Message, "redundant and can be deleted") {
			foundDeletion = true
		}
	}
	if foundDeletion {
		t.Fatalf("R1 deletion should be withheld when the variable is used in multiple WithSections calls")
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
