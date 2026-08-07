package glazedmigration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSource = `package fixture

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
)

func build() {
	_, _ = settings.NewGlazedSection(schema.WithDefaults(map[string]interface{}{
		"output": "json",
	}))
	_ = settings.GlazedSlug
}
`

// writeFixture copies the fixture source into a scratch directory tree so
// fix application never touches the repository.
func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	sub := filepath.Join(dir, "pkg", "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(fixtureSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.go"), []byte("package sub\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestScanFindsDiagnostics(t *testing.T) {
	dir := writeFixture(t)

	diagnostics, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("Scan returned no diagnostics, want findings for removed APIs")
	}
	for _, d := range diagnostics {
		if !strings.HasSuffix(d.File, "a.go") {
			t.Errorf("diagnostic in unexpected file %s", d.File)
		}
		if d.Line < 1 || d.Column < 1 {
			t.Errorf("diagnostic has invalid position %d:%d", d.Line, d.Column)
		}
	}
	totalFixes := 0
	for _, d := range diagnostics {
		totalFixes += d.FixCount
	}
	if totalFixes == 0 {
		t.Error("Scan diagnostics carry no suggested fixes")
	}
	t.Logf("scan found %d diagnostics with %d edits", len(diagnostics), totalFixes)
}

func TestScanReportsMissingPath(t *testing.T) {
	_, err := Scan(context.Background(), []string{filepath.Join(t.TempDir(), "does-not-exist")})
	if err == nil {
		t.Fatal("Scan returned nil error for missing path")
	}
}

func TestApplyFixesRewritesFile(t *testing.T) {
	dir := writeFixture(t)
	target := filepath.Join(dir, "a.go")

	diagnostics, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	result, err := ApplyFixes(context.Background(), diagnostics)
	if err != nil {
		t.Fatalf("ApplyFixes: %v", err)
	}
	if result.AppliedPerFile[target] == 0 {
		t.Fatal("ApplyFixes applied no edits")
	}
	if result.Skipped != 0 {
		t.Errorf("ApplyFixes skipped %d edits, want none for non-overlapping fixes", result.Skipped)
	}
	if len(result.AppliedPerFile) != 1 {
		t.Errorf("AppliedPerFile = %v, want exactly the fixture file", result.AppliedPerFile)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{"settings.NewStructuredOutputSection(", `"format": "json"`, "settings.StructuredOutputSlug"} {
		if !strings.Contains(text, want) {
			t.Errorf("rewritten file does not contain %q:\n%s", want, text)
		}
	}
	for _, gone := range []string{"NewGlazedSection(", `"output": "json"`, "settings.GlazedSlug"} {
		if strings.Contains(text, gone) {
			t.Errorf("rewritten file still contains %q:\n%s", gone, text)
		}
	}
}

func TestApplyFixesIsIdempotentWhenRescanned(t *testing.T) {
	dir := writeFixture(t)

	diagnostics, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if _, err := ApplyFixes(context.Background(), diagnostics); err != nil {
		t.Fatalf("ApplyFixes: %v", err)
	}

	// After fixing, a rescan must find nothing left to migrate.
	after, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("rescan: %v", err)
	}
	if len(after) != 0 {
		t.Errorf("rescan found %d diagnostics after fix application, want 0", len(after))
	}
}

func TestScanExpandsDotDotDotAndDeduplicatesOverlappingRoots(t *testing.T) {
	if got := normalizeScanPath("./..."); got != "." {
		t.Fatalf("normalizeScanPath(./...) = %q, want .", got)
	}
	dir := writeFixture(t)
	baseline, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("baseline scan: %v", err)
	}

	target := filepath.Join(dir, "a.go")
	got, err := Scan(context.Background(), []string{dir + string(filepath.Separator) + "...", dir, target})
	if err != nil {
		t.Fatalf("overlapping ./... scan: %v", err)
	}
	if len(got) != len(baseline) {
		t.Fatalf("overlapping scan emitted %d diagnostics, want deduplicated baseline %d", len(got), len(baseline))
	}
}

func TestScanAndFixRecognizeDotImportedSlugWithoutTypeChecking(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dot.go")
	source := `package fixture
import . "github.com/go-go-golems/glazed/pkg/settings"
var _ = GlazedSlug
func localShadow() {
	GlazedSlug := "local"
	_ = GlazedSlug
}
`
	if err := os.WriteFile(target, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	diagnostics, err := Scan(context.Background(), []string{target})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(diagnostics) != 1 || diagnostics[0].FixCount != 1 {
		t.Fatalf("dot-import diagnostics = %#v, want one fixable GlazedSlug finding", diagnostics)
	}
	if _, err := ApplyFixes(context.Background(), diagnostics); err != nil {
		t.Fatalf("ApplyFixes: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "StructuredOutputSlug") {
		t.Fatalf("dot-import fix not applied:\n%s", content)
	}
}

func TestApplyFixesStopsBeforeWritesWhenCanceled(t *testing.T) {
	dir := writeFixture(t)
	target := filepath.Join(dir, "a.go")
	diagnostics, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := ApplyFixes(ctx, diagnostics)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ApplyFixes error = %v, want context.Canceled", err)
	}
	if len(result.AppliedPerFile) != 0 {
		t.Fatalf("canceled apply modified files: %v", result.AppliedPerFile)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("canceled apply changed the source file")
	}
}

func TestApplyFixesReturnsPartialResultsOnLaterFailure(t *testing.T) {
	dir := writeFixture(t)
	first := filepath.Join(dir, "a.go")
	later := filepath.Join(dir, "z.go")
	if err := os.WriteFile(later, []byte(fixtureSource), 0o644); err != nil {
		t.Fatal(err)
	}
	diagnostics, err := Scan(context.Background(), []string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if err := os.Remove(later); err != nil {
		t.Fatal(err)
	}

	result, err := ApplyFixes(context.Background(), diagnostics)
	if err == nil {
		t.Fatal("ApplyFixes returned nil error after a diagnosed file disappeared")
	}
	if result.AppliedPerFile[first] == 0 {
		t.Fatalf("partial result = %v, want earlier modified file reported", result.AppliedPerFile)
	}
}
