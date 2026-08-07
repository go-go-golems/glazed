package glazedmigration

import (
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

	diagnostics, err := Scan([]string{dir})
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
	_, err := Scan([]string{filepath.Join(t.TempDir(), "does-not-exist")})
	if err == nil {
		t.Fatal("Scan returned nil error for missing path")
	}
}

func TestApplyFixesRewritesFile(t *testing.T) {
	dir := writeFixture(t)
	target := filepath.Join(dir, "a.go")

	diagnostics, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	appliedPerFile, skipped, err := ApplyFixes(diagnostics)
	if err != nil {
		t.Fatalf("ApplyFixes: %v", err)
	}
	if appliedPerFile[target] == 0 {
		t.Fatal("ApplyFixes applied no edits")
	}
	if skipped != 0 {
		t.Errorf("ApplyFixes skipped %d edits, want none for non-overlapping fixes", skipped)
	}
	if len(appliedPerFile) != 1 {
		t.Errorf("appliedPerFile = %v, want exactly the fixture file", appliedPerFile)
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

	diagnostics, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if _, _, err := ApplyFixes(diagnostics); err != nil {
		t.Fatalf("ApplyFixes: %v", err)
	}

	// After fixing, a rescan must find nothing left to migrate.
	after, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("rescan: %v", err)
	}
	if len(after) != 0 {
		t.Errorf("rescan found %d diagnostics after fix application, want 0", len(after))
	}
}
