package patternmapper_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

func buildTestLayers(t *testing.T, defs ...*fields.Definition) *schema.Schema {
	t.Helper()
	l, err := schema.NewSection("demo", "Demo",
		schema.WithFields(defs...),
	)
	if err != nil {
		t.Fatalf("failed to create layer: %v", err)
	}
	return schema.NewSchema(schema.WithSections(l))
}

func TestLoadRulesFromYAML_Object(t *testing.T) {
	data := []byte(`
mappings:
  - source: "app.settings"
    target_layer: "demo"
    rules:
      - source: "api_key"
        target_parameter: "api-key"
      - source: "threshold"
        target_parameter: "threshold"
`)

	rules, err := pm.LoadRulesFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("LoadRulesFromReader failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 top-level rule, got %d", len(rules))
	}
	if rules[0].Source != "app.settings" || rules[0].TargetLayer != "demo" {
		t.Fatalf("unexpected top-level rule: %+v", rules[0])
	}
	if len(rules[0].Rules) != 2 {
		t.Fatalf("expected 2 child rules, got %d", len(rules[0].Rules))
	}
}

func TestLoadRulesFromYAML_Array(t *testing.T) {
	data := []byte(`
- source: "app.settings.api_key"
  target_layer: "demo"
  target_parameter: "api-key"
- source: "app.settings.threshold"
  target_layer: "demo"
  target_parameter: "threshold"
`)

	rules, err := pm.LoadRulesFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("LoadRulesFromReader failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestLoadMapperFromFile_E2E(t *testing.T) {
	// Prepare YAML mapping file
	content := []byte(`
mappings:
  - source: "app.{env}.settings"
    target_layer: "demo"
    rules:
      - source: "api_key"
        target_parameter: "{env}-api-key"
`)
	dir := t.TempDir()
	f := filepath.Join(dir, "mappings.yaml")
	if err := os.WriteFile(f, content, 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Layers with expected params
	defs := []*fields.Definition{
		fields.New("dev-api-key", fields.TypeString),
		fields.New("prod-api-key", fields.TypeString),
	}
	pls := buildTestLayers(t, defs...)

	mapper, err := pm.LoadMapperFromFile(pls, f)
	if err != nil {
		t.Fatalf("LoadMapperFromFile failed: %v", err)
	}

	// Config to map
	raw := map[string]interface{}{
		"app": map[string]interface{}{
			"dev": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "dev-secret",
				},
			},
			"prod": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "prod-secret",
				},
			},
		},
	}

	out, err := mapper.Map(raw)
	if err != nil {
		t.Fatalf("Map failed: %v", err)
	}
	if out["demo"]["dev-api-key"] != "dev-secret" {
		t.Fatalf("expected demo.dev-api-key to be dev-secret, got %v", out["demo"]["dev-api-key"])
	}
	if out["demo"]["prod-api-key"] != "prod-secret" {
		t.Fatalf("expected demo.prod-api-key to be prod-secret, got %v", out["demo"]["prod-api-key"])
	}
}
