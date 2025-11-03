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

## Validation

Glazed expects default-structured config files to be maps of `layer-slug -> { parameter-name: value }`. You can validate such files against your command’s parameter layers by scanning for unknown layers/parameters and type-checking values using parameter definitions.

### Default-structured validator (no mapper)

```go
func validateConfigFile(layers_ *layers.ParameterLayers, path string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    var raw map[string]interface{}
    if err := yaml.Unmarshal(b, &raw); err != nil { return err }

    issues := []string{}
    for layerSlug, v := range raw {
        layer, ok := layers_.Get(layerSlug)
        if !ok {
            issues = append(issues, fmt.Sprintf("unknown layer %s", layerSlug))
            continue
        }
        kv, ok := v.(map[string]interface{})
        if !ok {
            issues = append(issues, fmt.Sprintf("layer %s must be an object", layerSlug))
            continue
        }
        pds := layer.GetParameterDefinitions()
        known := map[string]bool{}
        pds.ForEach(func(pd *parameters.ParameterDefinition) { known[pd.Name] = true })
        for key, val := range kv {
            if !known[key] {
                issues = append(issues, fmt.Sprintf("unknown parameter %s.%s", layerSlug, key))
                continue
            }
            pd, _ := pds.Get(key)
            if _, err := pd.CheckValueValidity(val); err != nil {
                issues = append(issues, fmt.Sprintf("invalid value for %s.%s: %v", layerSlug, key, err))
            }
        }
    }
    if len(issues) > 0 {
        return fmt.Errorf(strings.Join(issues, "\n"))
    }
    return nil
}

// Validate overlays (low -> high precedence) per-file
func validateOverlay(layers_ *layers.ParameterLayers, files []string) error {
    for _, f := range files {
        if err := validateConfigFile(layers_, f); err != nil {
            return fmt.Errorf("%s: %w", f, err)
        }
    }
    return nil
}
```

Notes:
- This validator is conservative and per-file. It’s often preferable for CI since it flags extraneous keys early.
- Required parameters are enforced by the parser when applying values; a post-parse check can assert all requireds are satisfied across all sources.

### Pattern-based mapping validator

When using pattern rules, validate by constructing a mapper and calling `Map` on the raw config. The mapper performs rich checks (required patterns, missing target parameters with prefix-awareness, ambiguity, collisions).

```go
rules, err := patternmapper.LoadRulesFromFile("mappings.yaml")
if err != nil { return err }
mapper, err := patternmapper.NewConfigMapper(layers_, rules...)
if err != nil { return err }

data, err := os.ReadFile("config.yaml")
if err != nil { return err }
var raw map[string]interface{}
if err := yaml.Unmarshal(data, &raw); err != nil { return err }

// Validate-only: result is discarded; any error indicates invalid config
if _, err := mapper.Map(raw); err != nil {
    return err
}
```

Overlays: validate each file individually, or aggregate matches if you want overlay-aware required semantics.

## Deprecated: Viper Integration

Legacy Viper-based middlewares like `GatherFlagsFromViper` and per-command `--load-parameters-from-file` are deprecated. Prefer config middlewares with resolvers and `--config-file`.


