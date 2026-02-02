---
Title: API Change Analysis (origin/main → HEAD)
Ticket: ""
Status: active
Topics:
    - glazed
    - migration
    - api-design
    - docs
DocType: ""
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/doc/tutorials/migrating-to-facade-packages.md
      Note: Updated migration guide with latest API changes
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/analysis_queries.sql
      Note: Analysis SQL queries
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/import_git_diff_to_sqlite.py
      Note: Diff import script
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main-summary.json
      Note: Exported symbol summary
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main.sqlite
      Note: Queryable diff database
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# API Change Analysis (origin/main → HEAD)

## Goal

Provide a concise, copy/paste-ready reference of public API changes between `origin/main` and `HEAD` to support migration docs and upgrade work.

## Context

- Repo: `/home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed`
- Base ref: `origin/main` @ `3ea1f83d2fe4d30472559cf4f63f36152daa9937`
- Head ref: `HEAD` @ `f1327e8665fdb24b024cbc2460f17831bd09d73d`
- Diff DB: `glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main.sqlite`
- Summary JSON: `glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main-summary.json`

## Quick Reference

### Exported symbol additions

| Package | Symbol | Kind | Path |
| --- | --- | --- | --- |
| schema | FlagGroupUsage | type | `pkg/cmds/schema/schema.go` |
| schema | CommandFlagGroupUsage | type | `pkg/cmds/schema/schema.go` |
| schema | ComputeCommandFlagGroupUsage | func | `pkg/cmds/schema/schema.go` |
| schema | NewSectionFromYAML | func | `pkg/cmds/schema/schema.go` |
| sources | SourceDefaults | const | `pkg/cmds/sources/sources.go` |
| sources | FromMapFirst | func | `pkg/cmds/sources/sources.go` |
| sources | FromMapAsDefault | func | `pkg/cmds/sources/sources.go` |
| sources | FromMapAsDefaultFirst | func | `pkg/cmds/sources/sources.go` |
| values | SectionValuesOption | type | `pkg/cmds/values/values.go` |
| values | NewSectionValues | func | `pkg/cmds/values/values.go` |
| values | WithParameters | func | `pkg/cmds/values/values.go` |
| values | WithParameterValue | func | `pkg/cmds/values/values.go` |
| settings | NewGlazedSchema | func | `pkg/settings/glazed_layer.go` |

### Exported symbol removals

| Package | Symbol | Kind | Path |
| --- | --- | --- | --- |
| cmds | CommandDefinition | type | `pkg/cmds/cmds.go` |
| cmds | CommandDefinitionOption | type | `pkg/cmds/cmds.go` |
| schema | NewGlazedSchema | func | `pkg/cmds/schema/schema.go` |

### Signature / type shifts (public surface)

- `cmds.CommandDescription.Layers`: `*layers.ParameterLayers` → `*schema.Schema`.
- `layers.ParameterLayer` interface: `AddFlags(...*parameters.ParameterDefinition)` → `AddFlags(...*fields.Definition)`; `GetParameterDefinitions() *parameters.ParameterDefinitions` → `*fields.Definitions`.
- Command execution interfaces now use `*values.Values`:
  - `cmds.BareCommand`, `cmds.WriterCommand`, `cmds.GlazeCommand`, `cmds.CommandWithMetadata`.
  - `cli.CobraRunFunc`, `cli.CobraParser.Parse`, `cli.ParseCommandSettingsLayer`.
  - `middlewares.HandlerFunc` (and middleware signatures now used by `sources.Execute`).
- Cobra parsing and sources:
  - `middlewares.ParseFromCobraCommand`, `GatherArguments`, `UpdateFromEnv`, `SetFromDefaults`, `LoadParametersFromFile(s)` → `sources.FromCobra`, `FromArgs`, `FromEnv`, `FromDefaults`, `FromFile(s)`.
  - `parameters.WithParseStepSource(...)` → `sources.WithSource(...)`.
- Glazed settings layer:
  - `schema.NewGlazedSchema` removed, `settings.NewGlazedSchema` added.
  - Output helper functions in `pkg/settings` now accept `*values.SectionValues`.

### Migration hot spots (call sites changed)

- `pkg/cmds/cmds.go` (command interfaces + CommandDefinition removal)
- `pkg/cmds/layers/layer.go` (ParameterLayer method types)
- `pkg/cmds/schema/schema.go` (flag group usage + section-from-YAML + removal of NewGlazedSchema)
- `pkg/cmds/values/values.go` (new SectionValues builders)
- `pkg/cmds/sources/sources.go` (new Map-first/default helpers + SourceDefaults)
- `pkg/cli/*` (Cobra parser and command settings layers now use schema/values/fields)
- `pkg/settings/glazed_layer.go` (NewGlazedSchema + values.SectionValues in helper signatures)
- `pkg/appconfig/*` and `pkg/lua/*` (middleware/sources and values decoding changes)

## Usage Examples

### Update a command signature

```go
func (c *MyCmd) Run(ctx context.Context, parsedLayers *values.Values) error {
    // ...
    return nil
}
```

### Build schema sections using facade packages

```go
demoSection, _ := schema.NewSection("demo", "Demo",
    schema.WithPrefix("demo-"),
    schema.WithFields(
        fields.New("api-key", fields.TypeString),
    ),
)
```

### Execute a sources chain

```go
vals := values.New()
err := sources.Execute(schema_, vals,
    sources.FromCobra(cmd, sources.WithSource("flags")),
    sources.FromEnv("MYAPP", sources.WithSource("env")),
    sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
)
```

### Load a section from YAML

```go
section, _ := schema.NewSectionFromYAML(rawYAML, schema.WithPrefix("demo-"))
```
