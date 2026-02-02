---
Title: 'No-Return Rename: Removing Alias Facades and Migration Tool'
Ticket: GL-001-ADD-MIGRATION-DOCS
Status: active
Topics:
    - glazed
    - migration
    - api-design
    - docs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/cmds/fields/parameters.go
      Note: Field definitions (former parameters)
    - Path: pkg/cmds/schema/layer.go
      Note: Schema sections and collections
    - Path: pkg/cmds/sources/middlewares.go
      Note: Sources middleware chain
    - Path: pkg/cmds/values/parsed-layer.go
      Note: Values/parsed sections
    - Path: ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/rename_glazed_api.go
      Note: AST migration tool implementation
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# No-Return Rename: Removing Alias Facades and Migration Tool

## Executive Summary

We will remove the compatibility facade packages (`schema`, `fields`, `values`, `sources`) that currently rely on **type aliases** to legacy packages (`layers`, `parameters`, `middlewares`). The goal is a **no‑return** migration where the new vocabulary is the only public API. This requires:

1) Moving and renaming the real implementations into the new packages.
2) Eliminating old packages (or converting them to non-exported/internal helpers).
3) Migrating all code, tests, and documentation to the new names.
4) Providing a Go AST migration tool to update imports and identifiers safely and repeatably.

## Problem Statement

The repo still exposes type aliases and wrappers that preserve backwards compatibility with old package names and types. This blocks a clean break to the new API vocabulary and makes the public surface ambiguous (multiple names for the same concept). We need to remove the aliases, move the actual implementations into the new packages, and update all call sites. We also need a repeatable tool to migrate downstream codebases and future changes without hand-editing every file.

## Proposed Solution

### 1) Remove alias facades

- Replace alias-only files in `pkg/cmds/schema`, `pkg/cmds/fields`, `pkg/cmds/values`, and `pkg/cmds/sources` with **real implementations**.
- Delete or move the legacy packages to non-exported/internal locations (prefer delete for true no-return).

### 2) Move and rename packages

**Target package layout (post‑refactor):**

- `pkg/cmds/schema` → contains schema section and schema collection types.
- `pkg/cmds/fields` → contains field definitions and field types.
- `pkg/cmds/values` → contains resolved values / parsed values types.
- `pkg/cmds/sources` → contains parsing/source middlewares, config file loading, pattern mapper, etc.

**Moves:**

- `pkg/cmds/parameters/**` → `pkg/cmds/fields/**`
- `pkg/cmds/layers/**` → `pkg/cmds/schema/**` (section/collection types)
- `pkg/cmds/layers/parsed-*.go`, `serialize_parsed.go` → `pkg/cmds/values/**`
- `pkg/cmds/middlewares/**` → `pkg/cmds/sources/**`
- `pkg/cmds/middlewares/patternmapper/**` → `pkg/cmds/sources/patternmapper/**`

### 3) Rename exported identifiers to the new vocabulary

The following renames are **API-level** (public):

**Schema layer → Sections**

- `ParameterLayer` → `Section`
- `ParameterLayers` → `Schema`
- `ParameterLayerImpl` → `SectionImpl`
- `ParameterLayerOptions` → `SectionOption`
- `ParameterLayersOption` → `SchemaOption`
- `NewParameterLayer` → `NewSection`
- `NewParameterLayers` → `NewSchema`
- `WithLayers` → `WithSections`
- `WithParameterDefinitions` → `WithFields`

**Fields / parameters**

- `ParameterDefinition` → `Definition`
- `ParameterDefinitions` → `Definitions`
- `ParameterDefinitionOption` → `Option`
- `ParameterTypeX` → `TypeX`
- `NewParameterDefinition` → `New`
- `NewParameterDefinitions` → `NewDefinitions`

**Values / parsed layers**

- `ParsedLayer` → `SectionValues`
- `ParsedLayers` → `Values`
- `ParsedLayerOption` → `SectionValuesOption`
- `ParsedLayersOption` → `ValuesOption`
- `NewParsedLayer` → `NewSectionValues`
- `NewParsedLayers` → `New`
- `WithParsedParameters` → `WithParameters`
- `WithParsedParameterValue` → `WithParameterValue`

**Sources / middlewares**

- `ExecuteMiddlewares` → `Execute`
- `ParseFromCobraCommand` → `FromCobra`
- `GatherArguments` → `FromArgs`
- `UpdateFromEnv` → `FromEnv`
- `SetFromDefaults` → `FromDefaults`
- `LoadParametersFromFile(s)` → `FromFile` / `FromFiles`
- `UpdateFromMap` → `FromMap`
- `UpdateFromMapFirst` → `FromMapFirst`
- `UpdateFromMapAsDefault` → `FromMapAsDefault`
- `UpdateFromMapAsDefaultFirst` → `FromMapAsDefaultFirst`
- `WithParseStepSource` → `WithSource`

### 4) Go AST migration tool

A Go AST tool will:

- Re‑write **imports** from legacy paths to new paths.
- Re‑write **selector identifiers** for known package imports.
- Re‑write **package declarations** for moved files (optional via `--rewrite-package` mode).
- Support **dry‑run** and **diff output** for auditing.

This tool will be kept in the ticket under `scripts/` and can be reused for downstream migrations.

## Design Decisions

1) **No backwards compatibility.** Old packages are removed entirely (or converted to internal if needed). This ensures downstream users must migrate.
2) **Canonical new names only.** All exported names will use schema/fields/values/sources vocabulary.
3) **AST-based rewrite.** Avoids brittle regex changes and only updates code that truly references the old packages.
4) **Strict mapping.** Only explicit mappings are renamed; anything else is flagged for manual review.

## Alternatives Considered

- **Keep old packages with deprecation warnings.** Rejected: still backward compatible and undermines the “no return” goal.
- **Use gofmt + regex replacements.** Rejected: fragile and can corrupt code or docs.
- **Large manual migration.** Rejected: error‑prone and hard to repeat for downstream users.

## Implementation Plan

1) **Prepare migration tool**
   - Build AST rewrite tool with mapping tables (imports + identifiers).
   - Add dry‑run and report output (JSON/Markdown).

2) **Move packages**
   - Move files to `schema`, `fields`, `values`, `sources` with `git mv`.
   - Update package names and adjust imports in moved files.

3) **Run migration tool**
   - Apply to repo to fix imports and identifiers.
   - Review tool output and fix remaining edge cases manually.

4) **Delete old packages**
   - Remove `pkg/cmds/layers`, `pkg/cmds/parameters`, `pkg/cmds/middlewares`.

5) **Update docs and examples**
   - Ensure all references to legacy names are replaced.

6) **Validate**
   - `go test ./...`
   - Run examples that exercise parsing and config to ensure sources chain still works.

## Migration Tool Design (Detailed)

### CLI

```
cmd/glazed-refactor rename \
  --root /path/to/repo \
  --map ./mappings.yaml \
  --write \
  --report ./various/rename-report.json
```

### Mapping Model

```yaml
imports:
  "github.com/go-go-golems/glazed/pkg/cmds/layers": "github.com/go-go-golems/glazed/pkg/cmds/schema"
  "github.com/go-go-golems/glazed/pkg/cmds/parameters": "github.com/go-go-golems/glazed/pkg/cmds/fields"
  "github.com/go-go-golems/glazed/pkg/cmds/middlewares": "github.com/go-go-golems/glazed/pkg/cmds/sources"

idents:
  schema:
    ParameterLayer: Section
    ParameterLayers: Schema
    ParameterLayerImpl: SectionImpl
    ParameterLayerOptions: SectionOption
    ParameterLayersOption: SchemaOption
    NewParameterLayer: NewSection
    NewParameterLayers: NewSchema
    WithLayers: WithSections
    WithParameterDefinitions: WithFields

  fields:
    ParameterDefinition: Definition
    ParameterDefinitions: Definitions
    ParameterDefinitionOption: Option
    ParameterType: Type
    NewParameterDefinition: New
    NewParameterDefinitions: NewDefinitions

  values:
    ParsedLayer: SectionValues
    ParsedLayers: Values
    ParsedLayerOption: SectionValuesOption
    ParsedLayersOption: ValuesOption
    NewParsedLayer: NewSectionValues
    NewParsedLayers: New
    WithParsedParameters: WithParameters
    WithParsedParameterValue: WithParameterValue

  sources:
    ExecuteMiddlewares: Execute
    ParseFromCobraCommand: FromCobra
    GatherArguments: FromArgs
    UpdateFromEnv: FromEnv
    SetFromDefaults: FromDefaults
    LoadParametersFromFile: FromFile
    LoadParametersFromFiles: FromFiles
    UpdateFromMap: FromMap
    UpdateFromMapFirst: FromMapFirst
    UpdateFromMapAsDefault: FromMapAsDefault
    UpdateFromMapAsDefaultFirst: FromMapAsDefaultFirst
    WithParseStepSource: WithSource
```

### Rewrite Algorithm

1) Parse `.go` files with `go/parser`.
2) Build import map and update import paths.
3) For each selector expression `pkg.Symbol`:
   - If `pkg` matches renamed import, rewrite `Symbol` per mapping.
4) Emit `gofmt` output.
5) Produce a report listing:
   - Files changed
   - Imports rewritten
   - Identifiers renamed
   - Unmapped legacy identifiers (warnings)

### Safety Features

- **Dry-run** mode that outputs unified diff without rewriting.
- **Audit report** (JSON) for review.
- **Explicit allowlist** of renames to prevent accidental mutation.
- **Skip generated files** by default (`// Code generated ... DO NOT EDIT.`).

### Testing Strategy

- Unit tests for mapping rewrites (synthetic AST fixtures).
- Integration test running tool against a small sample package with known expected output.

## Open Questions

- Should any legacy packages be preserved as internal wrappers for extremely low‑level usage? (Default: no.)
- Do we want to rename `middlewares` to `sources` for non‑parsing helpers or keep a thin internal package?

## Rollout / Migration Guidance

- Communicate that the API is **breaking** and requires updates to import paths and type names.
- Provide the migration tool and a short CLI usage guide in docs.

