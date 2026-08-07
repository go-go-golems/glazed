package cmds

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/types"
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
	_, _ = settings.NewJqSection()
}
`

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "legacy.go"), []byte(fixtureSource), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// collectProcessor captures emitted rows without serialization.
type collectProcessor struct {
	rows []types.Row
}

func (c *collectProcessor) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}
func (c *collectProcessor) Close(context.Context) error { return nil }

// parsedWithPaths builds parsed values holding only the paths positional
// argument, mirroring what the Cobra parser would produce.
func parsedWithPaths(t *testing.T, description *cmds.CommandDescription, dir string) *values.Values {
	t.Helper()
	section, ok := description.GetDefaultSection()
	if !ok {
		t.Fatal("command description has no default section")
	}
	sv, err := values.NewSectionValues(section)
	if err != nil {
		t.Fatal(err)
	}
	sv.Fields.Set("paths", &fields.FieldValue{Value: []string{dir}})
	return values.New(values.WithSectionValues(schema.DefaultSlug, sv))
}

func rowGet(t *testing.T, row types.Row, key string) interface{} {
	t.Helper()
	v, ok := row.Get(key)
	if !ok {
		t.Fatalf("row missing key %q", key)
	}
	return v
}

func TestCheckCommandEmitsDiagnosticRows(t *testing.T) {
	dir := writeFixture(t)

	cmd := NewCheckCommand()
	processor := &collectProcessor{}
	err := cmd.RunIntoGlazeProcessor(context.Background(), parsedWithPaths(t, cmd.Description(), dir), processor)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(processor.rows) != 4 {
		t.Fatalf("emitted %d rows, want 4 findings", len(processor.rows))
	}
	autoFixable := 0
	for _, row := range processor.rows {
		file := rowGet(t, row, "file").(string)
		if !strings.HasSuffix(file, "legacy.go") {
			t.Errorf("row file = %v, want legacy.go", file)
		}
		if rowGet(t, row, "line").(int) < 1 || rowGet(t, row, "column").(int) < 1 {
			t.Errorf("row has invalid position %v:%v", rowGet(t, row, "line"), rowGet(t, row, "column"))
		}
		if rowGet(t, row, "fixes_available").(int) > 0 {
			autoFixable++
		}
	}
	if autoFixable != 3 {
		t.Errorf("%d auto-fixable findings, want 3 (jq section finding is report-only)", autoFixable)
	}
}

func TestFixCommandAppliesMigrationsAndReportsManual(t *testing.T) {
	dir := writeFixture(t)
	target := filepath.Join(dir, "legacy.go")

	cmd := NewFixCommand()
	processor := &collectProcessor{}
	err := cmd.RunIntoGlazeProcessor(context.Background(), parsedWithPaths(t, cmd.Description(), dir), processor)
	if err != nil {
		t.Fatalf("fix: %v", err)
	}
	if len(processor.rows) != 2 {
		t.Fatalf("emitted %d rows, want 1 modified-file row + 1 manual row", len(processor.rows))
	}
	if rowGet(t, processor.rows[0], "edits_applied").(int) != 3 {
		t.Errorf("edits_applied = %v, want 3", rowGet(t, processor.rows[0], "edits_applied"))
	}
	if !strings.Contains(rowGet(t, processor.rows[1], "message").(string), "manual migration required") {
		t.Errorf("second row should flag the manual jq migration: %v", processor.rows[1])
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{"settings.NewStructuredOutputSection(", `"format": "json"`, "settings.StructuredOutputSlug"} {
		if !strings.Contains(text, want) {
			t.Errorf("rewritten fixture does not contain %q:\n%s", want, text)
		}
	}
}

func TestCheckCommandPropagatesScanError(t *testing.T) {
	cmd := NewCheckCommand()
	processor := &collectProcessor{}
	err := cmd.RunIntoGlazeProcessor(context.Background(),
		parsedWithPaths(t, cmd.Description(), filepath.Join(t.TempDir(), "missing")), processor)
	if err == nil {
		t.Fatal("check returned nil error for missing path")
	}
}

func TestCobraBuilderAddsUniversalOutputFlagsOnly(t *testing.T) {
	checkCobra, err := cli.BuildCobraCommandFromCommand(NewCheckCommand())
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"format", "output-fields", "max-output-rows"} {
		if checkCobra.Flags().Lookup(flag) == nil {
			t.Errorf("check command missing universal flag --%s", flag)
		}
	}
	fixCobra, err := cli.BuildCobraCommandFromCommand(NewFixCommand())
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"format", "output-fields", "max-output-rows"} {
		if fixCobra.Flags().Lookup(flag) == nil {
			t.Errorf("fix command missing universal flag --%s", flag)
		}
	}
	// Removed legacy output flags must not reappear.
	for _, flag := range []string{"output", "template", "fields", "sort-columns", "with-headers"} {
		if checkCobra.Flags().Lookup(flag) != nil {
			t.Errorf("check command unexpectedly exposes removed flag --%s", flag)
		}
	}
	// paths is a positional argument, not a flag.
	if checkCobra.Flags().Lookup("paths") != nil {
		t.Error("paths should be a positional argument, not a flag")
	}
}
