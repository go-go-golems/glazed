---
Title: Migrating to the New Facade Packages (schema/fields/values/sources)
Slug: migrating-to-facade-packages
Short: Step-by-step guide to migrate Glazed code from layers/parameters/middlewares vocabulary to the new facade packages (schema/fields/values/sources)
Topics:
- tutorial
- migration
- api-design
- schema
- values
- sources
- commands
Commands:
- none
Flags:
- none
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## Overview

Glazed introduced **additive** facade packages under `github.com/go-go-golems/glazed/pkg/cmds/`:

- `schema` — schema sections (previously “layers”)
- `fields` — field definitions and field types (previously “parameters”)
- `values` — resolved values + decoding helpers (previously “parsed layers”)
- `sources` — value sources / resolver chain helpers (previously “cmds/middlewares”)

These packages are implemented using **type aliases** plus small wrapper functions. That means:

- There is no behavioral change in the underlying engine.
- Existing code keeps working.
- You can migrate incrementally and mix old/new imports where needed.

## Quick mapping (old → new)

### Schema and fields

- `pkg/cmds/schema.Section` → `pkg/cmds/schema.Section`
- `pkg/cmds/schema.Schema` → `pkg/cmds/schema.Schema`
- `pkg/cmds/fields.Definition` → `pkg/cmds/fields.Definition`
- `pkg/cmds/fields.Type*` → `pkg/cmds/fields.Type*`

### Resolved values

- `pkg/cmds/values.Values` → `pkg/cmds/values.Values`
- `pkg/cmds/values.SectionValues` → `pkg/cmds/values.SectionValues`
- `parsed.InitializeStruct(slug, &dst)` → `values.DecodeSectionInto(vals, slug, &dst)`

### Sources / middleware chain

- `sources.FromCobra` → `sources.FromCobra`
- `sources.FromArgs` → `sources.FromArgs`
- `sources.FromEnv` → `sources.FromEnv`
- `sources.FromDefaults` → `sources.FromDefaults`
- `sources.FromFile(s)` → `sources.FromFile` / `sources.FromFiles`
- `sources.FromMap` → `sources.FromMap`
- `sources.Execute` → `sources.Execute`

## Important: aliases and interface signatures

The facade types are **aliases** of the original types. In practice this means you can write:

```go
func (c *MyCmd) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    // ...
}
```

…and it still satisfies interfaces that mention `*values.Values`, because `values.Values` is an alias for `values.Values`.

## Step-by-step migration recipe

### Step 1: Switch imports

Replace:

- `github.com/go-go-golems/glazed/pkg/cmds/layers`
- `github.com/go-go-golems/glazed/pkg/cmds/parameters`
- `github.com/go-go-golems/glazed/pkg/cmds/middlewares`

With (where applicable):

- `github.com/go-go-golems/glazed/pkg/cmds/schema`
- `github.com/go-go-golems/glazed/pkg/cmds/fields`
- `github.com/go-go-golems/glazed/pkg/cmds/values`
- `github.com/go-go-golems/glazed/pkg/cmds/sources`

You can keep old imports for advanced/legacy types (for example `parameters.FileData`) until you’re ready to refactor them.

### Step 2: Replace parameter definitions

Before:

```go
fields.New("limit", fields.TypeInteger, fields.WithDefault(10))
```

After:

```go
fields.New("limit", fields.TypeInteger, fields.WithDefault(10))
```

### Step 3: Replace schema construction (optional)

If you currently build explicit layers:

Before:

```go
demoLayer, _ := schema.NewSection("demo", "Demo",
    layers.WithPrefix("demo-"),
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
    ),
)
```

After:

```go
demoSection, _ := schema.NewSection("demo", "Demo",
    schema.WithPrefix("demo-"),
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
    ),
)
```

### Step 4: Decode values into settings structs

Before:

```go
settings := &MySettings{}
_ = parsedLayers.InitializeStruct(schema.DefaultSlug, settings)
```

After:

```go
settings := &MySettings{}
_ = values.DecodeSectionInto(vals, schema.DefaultSlug, settings)
```

### Step 5: Use `sources` for explicit precedence chains

If you manually build a resolver chain, prefer the `sources` wrappers:

```go
vals := values.New()
schema_ := schema.NewSchema(schema.WithSections(demoSection))

err := sources.Execute(schema_, vals,
    sources.FromCobra(cmd, sources.WithSource("flags")),
    sources.FromEnv("MYAPP", sources.WithSource("env")),
    sources.FromFile("config.yaml", sources.WithParseOptions(sources.WithSource("config"))),
    sources.FromDefaults(sources.WithSource("defaults")),
)
```

## Glazed “output flags” layer: what to do now

- If your command implements `cmds.GlazeCommand`, `cli.BuildCobraCommand(...)` will ensure the glazed output layer exists, so you usually don’t need to add it manually.
- If you do want to add it explicitly (e.g. when building a schema yourself), prefer `settings.NewGlazedSchema()` (wrapper around `settings.NewGlazedParameterLayers()`).

## When you still need the old packages

The goal is “new vocabulary at the API edges”, not “eliminate old packages everywhere”.

Common reasons to keep old imports:

- Cobra-only plumbing: attaching layers to Cobra uses `layers.CobraParameterLayer`.
- Some helper types/functions still live in `parameters` (e.g. `parameters.FileData`, `parameters.RenderValue`).
- Some config mapping utilities still live under `cmds/middlewares/*`.

Migrating piecemeal is fine; because facade types are aliases, interoperability is zero-friction.

## Validation checklist

- `gofmt -w <changed files>`
- `go test ./...`
- Run one of the examples with env + flags and confirm precedence:
  - defaults < config files < env < flags
