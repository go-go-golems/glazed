---
Title: Config Files Quickstart
Slug: config-files-quickstart
Short: Minimal examples to load single and multiple config files with clear precedence
Topics:
- tutorial
- configuration
- middlewares
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Config Files Quickstart

This tutorial shows how to load configuration from one or more files using Glazed middlewares. You’ll see a simple single-file setup and a multi-file overlay with deterministic precedence. We’ll also show how to inspect parse steps using `--print-parsed-parameters`.

## Prerequisites

- Go 1.19+
- Familiarity with Cobra commands and Glazed layers

## 1. Single File

Create a minimal command with a single custom layer and an explicit config file path:

```go
demoLayer, _ := schema.NewSection(
    "demo", "Demo settings",
    layers.WithPrefix("demo-"),
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
        fields.New("threshold", fields.TypeInteger, fields.WithDefault(10)),
    ),
)

desc := cmds.NewCommandDescription("demo", cmds.WithLayersList(demoLayer))

cmd := &DemoBareCommand{CommandDescription: desc}

cobraCmd, _ := cli.BuildCobraCommandFromCommand(cmd,
    cli.WithParserConfig(cli.CobraParserConfig{
        SkipCommandSettingsLayer: true,
        ConfigPath:               "./config.yaml",
    }),
)
```

Example `config.yaml`:

```yaml
demo:
  api-key: cfg-one
  threshold: 33
```

Run:

```bash
go run ./cmd/examples/config-single demo
```

Expected output:

```text
api_key=cfg-one threshold=33
```

## 2. Multiple Files (Overlays)

Use a resolver to return an ordered list of files (low → high precedence):

```go
resolver := func(_ *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
    return []string{"base.yaml", "env.yaml", "local.yaml"}, nil
}

cobraCmd, _ := cli.BuildCobraCommandFromCommand(cmd,
    cli.WithParserConfig(cli.CobraParserConfig{
        SkipCommandSettingsLayer: true,
        ConfigFilesFunc:          resolver,
    }),
)
```

Example files (low → high):

```yaml
# base.yaml
demo:
  api-key: base
  threshold: 5
```

```yaml
# env.yaml
demo:
  api-key: env-file
  threshold: 12
```

```yaml
# local.yaml
demo:
  api-key: local
  threshold: 20
```

Run:

```bash
go run ./cmd/examples/config-overlay overlay
```

Expected output:

```text
api_key=local threshold=20
```

## 3. Inspect Parse Steps

Add `--print-parsed-parameters` to see each config file applied in sequence:

```bash
go run ./cmd/examples/config-overlay overlay --print-parsed-parameters
```

Excerpt:

```yaml
demo:
  threshold:
    log:
      - source: config
        metadata: { config_file: base.yaml, index: 0 }
        value: 5
      - source: config
        metadata: { config_file: env.yaml, index: 1 }
        value: 12
      - source: config
        metadata: { config_file: local.yaml, index: 2 }
        value: 20
    value: 20
```

## 4. Override with Env and Flags

Precedence remains: Defaults < Config Files < Env < Args < Flags.

```bash
# Env override
DEMO_API_KEY=env-var go run ./cmd/examples/config-overlay overlay

# Flag override
go run ./cmd/examples/config-overlay overlay --demo-threshold 77
```

## 5. Pattern: base + .override.yaml

To layer `<base>.override.yaml` automatically on top of `--config-file`:

```go
resolver := func(parsed *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
    cs := &cli.CommandSettings{}
    _ = parsed.InitializeStruct(cli.CommandSettingsSlug, cs)
    files := []string{}
    if cs.ConfigFile != "" {
        files = append(files, cs.ConfigFile)
        dir := filepath.Dir(cs.ConfigFile)
        base := filepath.Base(cs.ConfigFile)
        ext := filepath.Ext(base)
        stem := strings.TrimSuffix(base, ext)
        override := filepath.Join(dir, fmt.Sprintf("%s.override.yaml", stem))
        if _, err := os.Stat(override); err == nil {
            files = append(files, override)
        }
    }
    return files, nil
}
```

Run:

```bash
go run ./cmd/examples/overlay-override overlay-override --config-file ./base.yaml
```

## 6. Pattern-Based Mapping (Optional)

Map arbitrary config structures to parameters without custom Go by using the pattern-based config mapper. Works with YAML or JSON files.

```go
// Define a layer
demoLayer, _ := schema.NewSection("demo", "Demo",
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
        fields.New("dev-api-key", fields.TypeString),
        fields.New("prod-api-key", fields.TypeString),
    ),
)
paramLayers := schema.NewSchema(layers.WithLayers(demoLayer))

// Create a mapper using a named capture {env}
mapper, _ := patternmapper.NewConfigMapper(paramLayers,
    patternmapper.MappingRule{
        Source:      "app.{env}.settings",
        TargetLayer: "demo",
        Rules: []patternmapper.MappingRule{
            {Source: "api_key", TargetParameter: "{env}-api-key"},
        },
    },
)

// Use the mapper when loading the file
mw := sources.FromFile("config.yaml",
    middlewares.WithConfigMapper(mapper),
)
_ = sources.Execute(paramLayers, values.New(), mw)
```

Builder API (fluent):

```go
b := patternmapper.NewConfigMapperBuilder(paramLayers).
    MapObject("app.{env}.settings", "demo", []patternmapper.MappingRule{
        patternmapper.Child("api_key", "{env}-api-key"),
    })
mapper, _ := b.Build()
```

## 7. Direct Middleware API (Alternative to Cobra)

You can execute middlewares directly without relying on the Cobra parser config:

```go
err := sources.Execute(layers_, parsed,
    sources.FromDefaults(),
    sources.FromFiles([]string{"base.yaml", "local.yaml"}),
    sources.FromEnv("APP"),
    sources.FromCobra(cmd), // flags & args
)
```

## 8. Validation and Ambiguity (Gotchas)

- Required patterns: mark mappings as required so missing keys error out.

  ```go
  patternmapper.MappingRule{
      Source:          "app.settings.api_key",
      TargetLayer:     "demo",
      TargetParameter: "api-key",
      Required:        true,
  }
  ```

- Ambiguity: wildcard patterns that match multiple different values or rules that resolve to the same target parameter cause errors. Prefer named captures (e.g., `app.{env}.api_key`) when collecting multiple values.

- Missing parameters: mapping to a non-existent parameter errors (prefix-aware), helping catch typos early.

## 9. Deprecated: Viper Integration

Legacy Viper-based config parsing (e.g., `GatherFlagsFromViper`) is deprecated. Prefer config file middlewares plus env and flags:

```go
err := sources.Execute(layers_, parsed,
    sources.FromDefaults(),
    sources.FromFiles([]string{"base.yaml", "env.yaml", "local.yaml"}),
    sources.FromEnv("APP"),
    sources.FromCobra(cmd),
)
```

## 10. Validate Config Files

See the validation guides for full explanations and code snippets:

```
glaze help config-files
glaze help pattern-based-config-mapping
```

## Next Steps

- See topic pages and examples for deeper coverage:

```
glaze help config-files
glaze help pattern-based-config-mapping
glaze help cmds-middlewares
```

- Examples: `cmd/examples/config-overlay`, `cmd/examples/config-pattern-mapper`


