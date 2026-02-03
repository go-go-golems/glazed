---
Title: Migrating to the New Facade Packages (schema/fields/values/sources)
Slug: migrating-to-facade-packages
Short: Step-by-step guide to migrate Glazed code from sections/fields/middlewares vocabulary to the new facade packages (schema/fields/values/sources)
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

- `schema` — schema sections (previously “sections”)
- `fields` — field definitions and field types (previously “fields”)
- `values` — resolved values + decoding helpers (previously “parsed sections”)
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
- `pkg/cmds/fields.Definitions` → `pkg/cmds/fields.Definitions`
- `pkg/cmds/fields.Type*` → `pkg/cmds/fields.Type*`

### Resolved values

- `pkg/cmds/schema.Values` → `pkg/cmds/values.Values`
- `pkg/cmds/schema.SectionValues` → `pkg/cmds/values.SectionValues`
- `sections.NewValues()` → `values.New()`
- `sections.NewSectionValues(section, ...)` → `values.NewSectionValues(section, ...)`
- `sections.WithFieldValues(...)` → `values.WithFields(...)`
- `sections.WithFieldValueValue(...)` → `values.WithFieldValue(...)`
- `parsedSections.DecodeSectionInto(slug, &dst)` → `values.DecodeSectionInto(parsedSections, slug, &dst)`

### Sources / middleware chain

- `middlewares.ParseFromCobraCommand` → `sources.FromCobra`
- `middlewares.GatherArguments` → `sources.FromArgs`
- `middlewares.UpdateFromEnv` → `sources.FromEnv`
- `middlewares.SetFromDefaults` → `sources.FromDefaults`
- `middlewares.LoadFieldsFromFile(s)` → `sources.FromFile` / `sources.FromFiles`
- `middlewares.UpdateFromMap` → `sources.FromMap` / `sources.FromMapFirst`
- `middlewares.ExecuteMiddlewares` → `sources.Execute`
- `fields.WithParseStepSource(...)` → `sources.WithSource(...)`

## Important: aliases and interface signatures

The facade types are **aliases** of the original types. In practice this means you can write:

```go
func (c *MyCmd) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    // ...
}
```

…and it still satisfies interfaces that mention `*values.Values`, because `values.Values` is an alias for `values.Values`.

That said, most public interfaces now use the new names (`*values.Values`, `schema.Section`, `fields.Definition`). Update your method signatures to match for clarity and to reduce confusion.

## Recent API updates to account for

These changes landed alongside the facade packages. If you compile against `origin/main` vs `HEAD`, expect to update the following:

- `cmds.CommandDefinition` and `cmds.CommandDefinitionOption` were removed. Use `cmds.CommandDescription` and `cmds.CommandDescriptionOption` instead.
- `cmds.CommandDescription.Sections` is now `*schema.Schema` (was `*schema.Schema`).
- `schema.Section` interface methods now use `*fields.Definition` / `*fields.Definitions`:
  - `AddFields(...*fields.Definition)`
  - `GetDefinitions() *fields.Definitions`
- Command execution interfaces now accept `*values.Values`:
  - `cmds.BareCommand`, `cmds.WriterCommand`, `cmds.GlazeCommand`, `cmds.CommandWithMetadata`
  - `cli.CobraRunFunc`, `cli.CobraParser.Parse`, `cli.ParseCommandSettingsSection`
  - `middlewares.HandlerFunc` / `middlewares.ExecuteMiddlewares` (now `sources.Execute`)
- `settings.NewGlazedSchema` moved to `pkg/settings`. `schema.NewGlazedSchema` was removed.
- `sources` additions: `FromMapFirst`, `FromMapAsDefault`, `FromMapAsDefaultFirst`, and `SourceDefaults`.
- `schema` additions: `NewSectionFromYAML`, `ComputeCommandFlagGroupUsage`, `FlagGroupUsage`, `CommandFlagGroupUsage`.
- `values` additions: `NewSectionValues`, `SectionValuesOption`, `WithFields`, `WithFieldValue`.

## Step-by-step migration recipe

### Step 1: Switch imports

Replace:

- `github.com/go-go-golems/glazed/pkg/cmds/schema`
- `github.com/go-go-golems/glazed/pkg/cmds/fields`
- `github.com/go-go-golems/glazed/pkg/cmds/middlewares`

With (where applicable):

- `github.com/go-go-golems/glazed/pkg/cmds/schema`
- `github.com/go-go-golems/glazed/pkg/cmds/fields`
- `github.com/go-go-golems/glazed/pkg/cmds/values`
- `github.com/go-go-golems/glazed/pkg/cmds/sources`

You can keep old imports for advanced/legacy types (for example `fields.FileData`) until you’re ready to refactor them.

### Step 2: Replace field definitions

Before:

```go
fields.NewDefinition("limit", fields.TypeInteger, fields.WithDefault(10))
```

After:

```go
fields.New("limit", fields.TypeInteger, fields.WithDefault(10))
```

### Step 3: Replace schema construction (optional)

If you currently build explicit sections:

Before:

```go
demoSection, _ := sections.NewSection("demo", "Demo",
    sections.WithPrefix("demo-"),
    sections.WithDefinitions(
        fields.NewDefinition("api-key", fields.TypeString),
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

### Step 4: Update command interfaces to accept `*values.Values`

Before:

```go
func (c *MyCmd) Run(ctx context.Context, parsedSections *values.Values) error {
    // ...
}
```

After:

```go
func (c *MyCmd) Run(ctx context.Context, parsedSections *values.Values) error {
    // ...
}
```

### Step 5: Decode values into settings structs

Before:

```go
settings := &MySettings{}
_ = parsedSections.DecodeSectionInto(schema.DefaultSlug, settings)
```

After:

```go
settings := &MySettings{}
_ = values.DecodeSectionInto(vals, schema.DefaultSlug, settings)
```

### Step 6: Use `sources` for explicit precedence chains

If you manually build a resolver chain, prefer the `sources` wrappers:

```go
vals := values.New()
schema_ := schema.NewSchema(schema.WithSections(demoSection))

err := sources.Execute(schema_, vals,
    sources.FromCobra(cmd, sources.WithSource("flags")),
    sources.FromEnv("MYAPP", sources.WithSource("env")),
    sources.FromFile("config.yaml", sources.WithParseOptions(sources.WithSource("config"))),
    sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
)
```

## Glazed “output flags” section: what to do now

- If your command implements `cmds.GlazeCommand`, `cli.BuildCobraCommand(...)` will ensure the glazed output section exists, so you usually don’t need to add it manually.
- If you do want to add it explicitly (e.g. when building a schema yourself), prefer `settings.NewGlazedSchema()` (wrapper around `settings.NewGlazedSchema()`).

## When you still need the old packages

The goal is “new vocabulary at the API edges”, not “eliminate old packages everywhere”.

Common reasons to keep old imports:

- Cobra-only plumbing: attaching sections to Cobra uses `sections.CobraSection`.
- Some helper types/functions still live in `fields` (e.g. `fields.FileData`, `fields.RenderValue`).
- Some config mapping utilities still live under `cmds/middlewares/*`.

Migrating piecemeal is fine; because facade types are aliases, interoperability is zero-friction.

## Validation checklist

- `gofmt -w <changed files>`
- `go test ./...`
- Run one of the examples with env + flags and confirm precedence:
  - defaults < config files < env < flags
