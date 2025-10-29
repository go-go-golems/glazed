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
demoLayer, _ := layers.NewParameterLayer(
    "demo", "Demo settings",
    layers.WithPrefix("demo-"),
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
        parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger, parameters.WithDefault(10)),
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
resolver := func(_ *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
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
resolver := func(parsed *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
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

## Next Steps

- See the topic page for deeper coverage:

```
glaze help config-files
```


