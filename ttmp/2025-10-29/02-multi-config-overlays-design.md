### Multi-config overlays for Glazed middlewares (design)

Goal: support loading multiple config files where later files override earlier ones, with each file recorded as its own parse step for traceability. Provide both explicit file lists and pattern-based discovery (e.g., app-local overrides) without relying on Viper.

## Requirements

- Load N config files in a defined precedence order: defaults < config[0] < ... < config[N-1] < env < args < flags
- Track each file as a distinct parse step (source="config", metadata includes config_file and index)
- Support explicit files via CLI and API, plus discovery via patterns and common locations
- Backward-compatible with current single-file `--config-file` behavior

## Precedence Model

- Execution order (lowest → highest precedence):
  - SetFromDefaults
  - LoadParametersFromFiles(filesLowToHigh)
  - UpdateFromEnv(APP_...)
  - GatherArguments(args)
  - ParseFromCobraCommand(cmd)

Note: the middleware builder appends in reverse precedence so execution results in the order shown above.

## API Surface

1) New middleware to apply multiple files in sequence (low→high precedence):

```go
// Applies files in the order provided. Each file becomes a separate parse step
// with metadata { config_file: <path>, index: <i> } and source "config".
func LoadParametersFromFiles(files []string, options ...parameters.ParseStepOption) Middleware
```

2) Optional convenience resolver that takes patterns and returns concrete file list:

```go
// Expands patterns into an ordered list of existing files. Patterns are processed in order.
// Example patterns: 
//   "/etc/%s/config.yaml", "$XDG_CONFIG_HOME/%s/config.yaml", "$HOME/.%s/config.yaml",
//   "./%s.yaml", "./%s.override.yaml", "./%s.local.yaml"
func ResolveConfigFiles(appName string, explicit []string, patterns []string) ([]string, error)
```

3) Cobra integration via parser config (no code yet, proposed):

```go
type CobraParserConfig struct {
    AppName            string
    ConfigPath         string         // single explicit file (kept)
    ConfigPaths        []string       // NEW: ordered explicit files (low→high)
    ConfigPatterns     []string       // NEW: discovery patterns, processed in order
    IncludeDefaultSearch bool         // NEW: include XDG/Home/etc base list
    IncludeCWDOverrides bool          // NEW: include ./<app>.yaml, ./<app>.override.yaml, ./<app>.local.yaml
    UseEnvConfigList   bool           // NEW: read ${APPNAME}_CONFIG_FILES (":" separated) as extra entries
}
```

4) Command Settings flags (proposed, design only):

- `--config-file string` (existing) → highest priority single file
- `--config-files string[]` (NEW) → ordered list (low→high)

Resolver merges in this sequence (lowest→highest):
1. Default search (if enabled)
2. Patterns (in order)
3. Env var `${APPNAME}_CONFIG_FILES` (split by ":")
4. `--config-files` (in order)
5. `--config-file` (single highest)

Each entry that exists is added to a list; the final list is passed to `LoadParametersFromFiles`.

## Middleware Behavior (sketch)

```go
func LoadParametersFromFiles(files []string, options ...parameters.ParseStepOption) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(layers_ *schema.Schema, pl *values.Values) error {
            if err := next(layers_, pl); err != nil { return err }
            for i, f := range files {
                m, err := readConfigFileToLayerMap(f) // map[string]map[string]interface{}
                if err != nil { return err }
                opts := append(options,
                    sources.WithSource("config"),
                    sources.WithMetadata(map[string]interface{}{
                        "config_file": f,
                        "index":       i,
                    }),
                )
                if err := updateFromMap(layers_, pl, m, opts...); err != nil { return err }
            }
            return nil
        }
    }
}
```

Notes:
- `readConfigFileToLayerMap` reuses YAML/JSON logic from current single-file middleware
- Separate `index` and `config_file` metadata allow inspection via `--print-parsed-parameters`

## Discovery Patterns

Default search order (lowest→highest), if enabled:
- `/etc/<appName>/config.yaml`
- `$XDG_CONFIG_HOME/<appName>/config.yaml` (or `os.UserConfigDir()`)
- `$HOME/.<appName>/config.yaml`

Optional CWD overrides (added after defaults, lowest→highest in this subset):
- `./<appName>.yaml`
- `./<appName>.override.yaml`
- `./<appName>.local.yaml`

Env-driven list (if enabled):
- `${APPNAME}_CONFIG_FILES` → split by `:` and append in given order

CLI-provided files:
- `--config-files f1 --config-files f2 ...` appended in given order
- `--config-file fN` appended last

The final concatenated list is passed to `LoadParametersFromFiles` and becomes a chain of parse steps.

## Parser Wiring (conceptual)

In `CobraParserConfig`-driven builder:

1. Build `[]string filesLowToHigh` from resolver with config above
2. Append `LoadParametersFromFiles(filesLowToHigh)` into the default chain right above defaults and below env/args/flags
3. Keep `UpdateFromEnv(AppName)` and flag parsing as-is for higher precedence

## Step Tracking Example

`--print-parsed-parameters` would show for a given parameter:

```yaml
demo:
  api-key:
    log:
      - source: config
        metadata:
          config_file: /etc/myapp/config.yaml
          index: 0
        value: abc
      - source: config
        metadata:
          config_file: /home/user/.myapp/config.yaml
          index: 1
        value: def
      - source: env
        metadata:
          env_key: MYAPP_DEMO_API_KEY
        value: ghi
      - source: cobra
        metadata:
          flag: demo-api-key
          parsed-strings: ["final"]
        value: final
    value: final
```

## Edge Cases

- Missing or unreadable files are skipped (log a debug step if desired); `Required` could be supported later per-pattern
- File format: YAML/JSON as today
- Per-layer filtering can wrap `LoadParametersFromFiles` using `WrapWithWhitelistedLayers`

## Migration Strategy

- Keep `--config-file` for single-file override (highest precedence among configs)
- Introduce `--config-files` for ordered lists; deprecate legacy `--load-parameters-from-file`
- Update documentation to explain resolver order and how to inspect parse history

## Examples (user-facing)

- Resolve XDG/Home + local overrides and then apply an explicit file:
  - `myapp cmd --config-file ./local.yaml`
- Ordered list of overlays:
  - `myapp cmd --config-files base.yaml --config-files env.yaml --config-files secrets.yaml`
- Env-controlled list:
  - `export MYAPP_CONFIG_FILES=base.yaml:dev.yaml:.secrets.yaml`
  - `myapp cmd`


