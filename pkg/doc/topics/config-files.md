---
Title: Config Files and Overlays
Slug: config-files
Short: Load one or more config files with clear precedence and traceable parse steps
Topics:
- configuration
- middlewares
- flags
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Config Files and Overlays

Glazed supports loading configuration from one or more files using middlewares. Each file is applied in sequence (low → high precedence) and recorded as a separate parse step, so you can inspect where values came from with `--print-parsed-parameters`.

## Core Concepts

- **Single file**: `LoadParametersFromFile(path)` reads a YAML or JSON file and maps values into layers/parameters.
- **Multiple files**: `LoadParametersFromFiles([]string)` applies files in order; later files overwrite earlier ones.
- **Cobra integration**: `LoadParametersFromResolvedFilesForCobra` accepts a resolver callback that returns the file list to load.
- **Precedence model**: Defaults < Config files (low→high) < Env < Positional Args < Flags.
- **Traceability**: Every update adds a parse step with `source: config` and metadata like `config_file` and `index`.

## Single File

```go
middlewares.ExecuteMiddlewares(layers, parsed,
    middlewares.LoadParametersFromFile("/etc/myapp/config.yaml"),
)
```

## Multiple Files (Overlays)

```go
files := []string{ "base.yaml", "env.yaml", "local.yaml" }
middlewares.ExecuteMiddlewares(layers, parsed,
    middlewares.LoadParametersFromFiles(files),
)
```

## Integrating with Cobra

Use a resolver to programmatically decide which files to load (e.g., based on `--config-file`, current directory, or environment):

```go
resolver := func(parsed *layers.ParsedLayers, cmd *cobra.Command, args []string) ([]string, error) {
    // build low->high precedence list
    return []string{ "base.yaml", "local.yaml" }, nil
}

mw := middlewares.LoadParametersFromResolvedFilesForCobra(cmd, args, resolver)
```

## Flags and Patterns

- `--config-file` (on the `command-settings` layer) can opt-in an explicit file.
- Common overlay patterns: `<base>.override.yaml`, `<app>.local.yaml`.
- Implement such patterns inside your resolver and return the resulting list.

## Inspecting Parse Steps

Run with `--print-parsed-parameters` to see a parameter’s history:

```yaml
demo:
  api-key:
    log:
      - source: config
        metadata:
          config_file: base.yaml
          index: 0
        value: base
      - source: config
        metadata:
          config_file: local.yaml
          index: 1
        value: local
    value: local
```

## Deprecated: Viper Integration

Legacy Viper-based middlewares like `GatherFlagsFromViper` and per-command `--load-parameters-from-file` are deprecated. Prefer config middlewares with resolvers and `--config-file`.


