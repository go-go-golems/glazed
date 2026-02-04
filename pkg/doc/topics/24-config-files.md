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

Glazed provides first-class support for reading configuration from one or more YAML/JSON files. Files are applied from low → high precedence, every file is recorded as its own parse step, and the result integrates cleanly with environment variables, positional args, and flags.

- Precedence: Defaults < Config files (low→high) < Env < Positional Args < Flags
- Traceability: Each config file write is logged with `source: config` and `{ config_file, index }` metadata and can be inspected with `--print-parsed-fields`.

This guide shows how to load single and multiple files, integrate with Cobra, implement app-level file resolution patterns, use pattern- and custom-mappers, inspect parse steps, and validate config files.

## Option A: Direct middlewares (library-only)

Use this approach when you’re embedding Glazed into a service or library and you want explicit, programmatic control over where configuration comes from. You decide the exact order of sources and call the middleware execution yourself. This makes the precedence rules obvious in code and easy to unit test.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
    // Define sections
    demo, _ := schema.NewSection(
        "demo", "Demo",
        schema.WithFields(
            fields.New("api-key", fields.TypeString),
            fields.New("threshold", fields.TypeInteger, fields.WithDefault(10)),
        ),
    )
    pls := schema.NewSchema(sections.WithSections(demo))
    parsed := values.New()

    // Apply middlewares in reverse-precedence order so later sources override earlier ones
    err := sources.Execute(pls, parsed,
        sources.FromDefaults(),                                // Defaults
        sources.FromFiles([]string{"base.yaml"}),  // Config (low → high)
        sources.FromEnv("MYAPP"),                           // Env (prefix)
        sources.FromArgs(args),                            // Positional args
        sources.FromCobra(cmd),                       // Flags
    )
    return err
}
```

## Option B: Cobra integration (recommended for CLIs)

If you’re building a CLI, the Cobra integration wires configuration, environment variables, positional arguments, and flags into a predictable pipeline with minimal boilerplate. `CobraParserConfig` lets you enable app-wide env prefixes, resolve config files, or inject your own resolver logic. This keeps your command code focused on business logic while Glazed handles the parsing pipeline and debug flags (like `--print-parsed-fields`).

Use `github.com/go-go-golems/glazed/pkg/cli` to build Cobra commands and attach config processing. The parser config can auto-wire env and config discovery.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
)

func build() (*cobra.Command, error) {
    demo, _ := schema.NewSection(
        "demo", "Demo",
        schema.WithFields(
            fields.New("api-key", fields.TypeString),
            fields.New("threshold", fields.TypeInteger, fields.WithDefault(10)),
        ),
    )
    desc := cmds.NewCommandDescription("demo", cmds.WithSectionsList(demo))

    // AppName enables env overrides (prefix = APPNAME_) and config discovery
    // ConfigPath uses an explicit path if provided
    // ConfigFilesFunc can return multiple files low → high precedence
    return cli.BuildCobraCommandFromCommand(&DemoBare{desc},
        cli.WithParserConfig(cli.CobraParserConfig{
            AppName:       "myapp",
            ConfigPath:    "", // optional explicit file (can come from --config-file too)
            ConfigFilesFunc: func(_ *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
                return []string{"base.yaml", "local.yaml"}, nil
            },
        }),
    )
}
```

## Single file

Load a single YAML/JSON file when your application’s configuration is centralized. The file is parsed into your field sections, and each value update is recorded as a `config` parse step, making it clear where settings came from when debugging.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

schema_ := schema.NewSchema(/* ... */)
parsed := values.New()
_ = sources.Execute(schema_, parsed,
    sources.FromFile("/etc/myapp/config.yaml"),
)
```

## Multiple files (overlays)

Overlays let you compose configuration from multiple files with deterministic precedence. A common pattern is a base file committed to version control plus a local developer override. Glazed applies files in the order you provide (low → high), and the last writer wins. Each file is recorded as its own parse step with `config_file` and `index` metadata so you can audit how a value evolved.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

files := []string{"base.yaml", "env.yaml", "local.yaml"} // low → high precedence
_ = sources.Execute(schema_, parsed,
    sources.FromFiles(files),
)
```

## App-level config discovery and patterns

Many CLIs have a conventional config location (XDG, home dotdir, or `/etc`). `ResolveAppConfigPath` encapsulates that search so your app can “just find” a config without hardcoding paths. Pair it with a `--config-file` flag (already provided by the `command-settings` section) so power users can override discovery. For overlays, a resolver can add optional files like `<base>.override.yaml` if they exist, keeping configuration flexible without hidden magic.

Use `github.com/go-go-golems/glazed/pkg/config.ResolveAppConfigPath` to discover a per-app config file:

```go
package main

import (
    appconfig "github.com/go-go-golems/glazed/pkg/config"
)

configPath, err := appconfig.ResolveAppConfigPath("myapp", "")
// Search order (first existing wins):
// 1) $XDG_CONFIG_HOME/myapp/config.yaml
// 2) $HOME/.myapp/config.yaml
// 3) /etc/myapp/config.yaml
```

You can also implement overlay patterns like `<base>.override.yaml` or `<app>.local.yaml` in a resolver:

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
)

resolver := func(parsed *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
    cs := &cli.CommandSettings{}
    _ = parsed.DecodeSectionInto(cli.CommandSettingsSlug, cs)
    if cs.ConfigFile == "" { return nil, nil }
    files := []string{cs.ConfigFile}
    dir, base := filepath.Dir(cs.ConfigFile), filepath.Base(cs.ConfigFile)
    stem := strings.TrimSuffix(base, filepath.Ext(base))
    override := filepath.Join(dir, fmt.Sprintf("%s.override.yaml", stem))
    if _, err := os.Stat(override); err == nil {
        files = append(files, override)
    }
    return files, nil
}
```

## Config file formats and mapping strategies

Glazed supports both “default-shaped” configs (where the file mirrors your sections and fields) and mappers (which translate arbitrary structures to field updates). Use the default structure for greenfield projects and simple cases—it’s the most transparent. Reach for mappers when you must consume legacy formats, have nested structures that don’t match your field layout, or need to derive multiple fields from one subtree.

Glazed supports two ways to map config file data into field sections:

1) Default structure (no mapper): your config matches the section/field shapes directly

```yaml
# default structure
demo:
  api-key: cfg
  threshold: 33
```

2) Mappers: use a mapper to transform arbitrary config shapes to section/field assignments.

### Pattern-based mapper (declarative)

Pattern mappers describe how to traverse a config tree and map matched values into fields. Patterns support exact segments, wildcards, and named captures (for environment-like keys such as `{env}`). Validation happens both at construction time (syntax, capture references, static targets) and at runtime (required matches, ambiguity, collisions). Prefer named captures over wildcards when you expect multiple values to be collected.

Use `github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper` to declare mapping rules and pass the mapper to `LoadFieldsFromFile`.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
)

demo, _ := schema.NewSection("demo", "Demo",
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
        fields.New("dev-api-key", fields.TypeString),
        fields.New("prod-api-key", fields.TypeString),
    ),
)
pls := schema.NewSchema(sections.WithSections(demo))

mapper, _ := pm.NewConfigMapper(pls,
    pm.MappingRule{
        Source:      "app.{env}.settings",
        TargetSection: "demo",
        Rules: []pm.MappingRule{
            {Source: "api_key", TargetField: "{env}-api-key"},
        },
    },
)

_ = sources.Execute(pls, values.New(),
    sources.FromFile("config.yaml", middlewares.WithConfigMapper(mapper)),
)
```

### Custom mapper (Go function)

Use a custom function when you need full control: conditional logic, array handling, value transformations, or cross-field validation that’s not practical to express with patterns. The function receives the unmarshaled config as `interface{}` and returns a standard `map[sectionSlug]map[paramName]any` for Glazed to apply.

Provide a `ConfigFileMapper` function to `WithConfigFileMapper` to transform raw config into a `map[sectionSlug]map[paramName]any`:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

var mapper middlewares.ConfigFileMapper = func(raw interface{}) (map[string]map[string]interface{}, error) {
    // inspect raw (unmarshaled YAML/JSON) and build the section map
    return map[string]map[string]interface{}{
        "demo": {"api-key": "secret", "threshold": 5},
    }, nil
}

_ = sources.Execute(pls, values.New(),
    sources.FromFile("config.yaml", middlewares.WithConfigFileMapper(mapper)),
)
```

## Inspecting parse steps (`--print-parsed-fields`)

Parsing is not a black box—every write records its source and any relevant metadata. Enable `--print-parsed-fields` to see the exact sequence of updates for each field. This is invaluable when debugging precedence issues (for example, “why didn’t my local override win?”) or auditing where a value originated.

Add the `command-settings` section (done automatically by the Cobra parser unless disabled) and run with `--print-parsed-fields` to see where a value came from:

```yaml
demo:
  api-key:
    log:
      - source: config
        metadata: { config_file: base.yaml, index: 0 }
        value: base
      - source: config
        metadata: { config_file: local.yaml, index: 1 }
        value: local
    value: local
```

## Validation

Validate configs early to catch mistakes before runtime. For default-shaped files, check for unknown sections/fields and type errors. For pattern-based configs, instantiate a mapper and call `Map` in a validate-only pass; the mapper will fail fast on missing required matches, ambiguous patterns, or invalid targets. These validators are small enough to run in CI and provide crisp error messages for contributors.

You can validate config files before applying them.

### Default-structured validator (unknown keys + type checks)

Apply this validator to YAML/JSON files that mirror your sections. It’s conservative by design: any unexpected section or field is flagged, and values are type-checked against your field definitions. This keeps configs tidy and prevents silent drift as fields evolve.

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "gopkg.in/yaml.v3"
)

func validateConfigFile(schema_ *schema.Schema, path string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    var raw map[string]interface{}
    if err := yaml.Unmarshal(b, &raw); err != nil { return err }

    issues := []string{}
    for sectionSlug, v := range raw {
        section, ok := schema_.Get(sectionSlug)
        if !ok { issues = append(issues, fmt.Sprintf("unknown section %s", sectionSlug)); continue }
        kv, ok := v.(map[string]interface{})
        if !ok { issues = append(issues, fmt.Sprintf("section %s must be an object", sectionSlug)); continue }
        pds := section.GetDefinitions()
        known := map[string]bool{}
        pds.ForEach(func(pd *fields.Definition) { known[pd.Name] = true })
        for key, val := range kv {
            if !known[key] { issues = append(issues, fmt.Sprintf("unknown field %s.%s", sectionSlug, key)); continue }
            pd, _ := pds.Get(key)
            if _, err := pd.CheckValueValidity(val); err != nil {
                issues = append(issues, fmt.Sprintf("invalid value for %s.%s: %v", sectionSlug, key, err))
            }
        }
    }
    if len(issues) > 0 { return fmt.Errorf(strings.Join(issues, "\n")) }
    return nil
}
```

Validate overlays per-file or implement overlay-aware required semantics if needed.

### Pattern-based validator (`mapper.Map`)

For declarative mappings, the mapper is your validator. Build it once per app (construction validates static aspects) and call `Map` on the raw config (runtime semantics validate dynamic aspects). Error messages include path hints and prefix-aware field names to accelerate debugging.

```go
package main

import (
    "os"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
    "gopkg.in/yaml.v3"
)

rules, _ := pm.LoadRulesFromFile("mappings.yaml")
mapper, _ := pm.NewConfigMapper(pls, rules...)
b, _ := os.ReadFile("config.yaml")
var raw map[string]interface{}
_ = yaml.Unmarshal(b, &raw)
if _, err := mapper.Map(raw); err != nil {
    // invalid config (required missing, collisions, unknown target params, ...)
}
```

## Tips & best practices

These guidelines help keep your configuration predictable across environments and easy to reason about during reviews and incident response.

- Keep overlays small and ordered: `base.yaml`, `env.yaml`, `local.yaml`.
- Prefer named captures over wildcards in pattern rules when collecting multiple values.
- Use `AppName` in `CobraParserConfig` to enable env overrides automatically.
- Record parse sources with `sources.WithSource("config")` (done for you by the config middlewares).

## Example projects and scripts

Use these as templates. Each example shows a minimal, focused scenario you can copy-paste and expand within your own app. The validation script runs “happy path” and negative tests to demonstrate failure modes and ensure your validators stay effective over time.

- Single-file: `cmd/examples/config-single`
- Overlays: `cmd/examples/config-overlay`
- Pattern mapper: `cmd/examples/config-pattern-mapper`
- Custom mapper: `cmd/examples/config-custom-mapper`
- Validation script: `glazed/ttmp/2025-11-03/validate-config-examples.sh`

## Deprecated: Viper integration

If you’re migrating from Viper-based setups, replace per-command file injection and env parsing with Glazed middlewares and `CobraParserConfig`. This typically reduces glue code while improving observability (traceable parse steps) and testability (deterministic precedence).

Legacy Viper-based middlewares like `GatherFlagsFromViper` and per-command `--load-fields-from-file` are deprecated. Prefer config middlewares (`LoadFieldsFromFiles`) with resolvers and `--config-file`.


