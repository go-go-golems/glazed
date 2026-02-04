---
Title: Further cleanup and renaming plan (no-compat)
Ticket: GL-002-FURTHER-CLEANUP
Status: active
Topics:
    - glazed
    - api-design
    - renaming
    - cleanup
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cmds
      Note: Primary API surface area targeted for renames
    - Path: ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/rename_glazed_api.go
      Note: |-
        Existing migration tooling to leverage
        Existing migration tooling
    - Path: ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/sources/01-glazed-cleanup-notes.md
      Note: |-
        Imported notes that drive the naming decisions
        Imported renaming rationale
    - Path: ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/various/01-parameter-layer-mentions.txt
      Note: |-
        Raw exhaustive inventory of Parameter/Layer mentions
        Raw Parameter/Layer inventory
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: Define the no-compat renaming cleanup after the schema/fields/values/sources shift
WhenToUse: When planning the next rename wave and tooling
---


# Further cleanup and renaming plan (no-compat)

## Executive Summary
The current public API still leaks the old vocabulary (Layer/Parameter) alongside the new schema/fields/values/sources terminology. This is a top-level usability issue: users cannot form a stable mental model while synonyms remain. This plan proposes a **no-backward-compat** rename sweep that eliminates the `Layer` and `Parameter` terms from exported identifiers, unifies the “Section” concept, and replaces “Parsed*” with value-oriented names. It uses the imported notes (`sources/01-glazed-cleanup-notes.md`) as the guiding rationale and leverages existing migration tooling to execute at scale.

## Problem Statement
- **Synonyms remain exported**: `Layer` and `Parameter` appear in public identifiers even after the no-return refactor.
- **Duplicated concepts**: `schema.Section` and `values.Section` are both exported and refer to different interfaces, creating ambiguity.
- **Misleading value naming**: `ParsedParameters` implies “intermediate raw parsing” even though these are resolved values with provenance.
- **Verb confusion**: `InitializeStruct` is used for multiple, distinct actions (defaults in schema, values into struct), making the API ambiguous.

## Imported Guidance (summary)
The notes in `sources/01-glazed-cleanup-notes.md` assert:
- One canonical noun per concept, **no synonyms** in exported API.
- **Layer must disappear** from exported identifiers.
- Fix the dual Section concept by removing or renaming `values.Section`.
- Replace “Parsed*” with “*Value(s)” and `InitializeStruct` with `Decode` (clearer verb).
- If “field” becomes the canonical noun, remove “parameter” from exported names and tags.

## Constraints & Non-Goals
- **No backward compatibility**: do not keep aliases or shims.
- Keep public API vocabulary stable once updated.
- This plan **does not** define unrelated feature work.

---

## Rename Inventory (exhaustive source of truth)
The raw list of all **Parameter** and **Layer** occurrences (case-insensitive) is stored at:
- `various/01-parameter-layer-mentions.txt`

That file is intentionally exhaustive and unfiltered. It includes:
- code identifiers (exported + internal)
- documentation strings and tutorials
- ticket artifacts under `ttmp/`

Use it as the master checklist for cleanup. Filtering should happen on top (e.g., “exported only”, “docs only”).

---

## Proposed Solution

### 1) Remove **Layer** from exported identifiers
**Goal:** ensure there is no exported symbol containing `Layer`.

#### CommandDescription (entry point)
- `CommandDescription.Layers` → `CommandDescription.Schema`
- `WithLayers` / `WithLayersList` / `WithLayersMap` → `WithSchema*` variants only

Example (current):
```go
cmds.WithLayers(schema.NewSchema(...))
```
Proposed:
```go
cmds.WithSchema(schema.NewSchema(...))
```

#### schema package
- `AppendLayers` / `PrependLayers` → `AppendSections` / `PrependSections`
- `ChildLayers` → `ChildSections`
- `AddLayerToCobraCommand` → `AddToCobraCommand` (or `AddSectionToCobraCommand`)
- `ParseLayerFromCobraCommand` → `ParseSectionFromCobraCommand`

Example:
```go
schema.NewSchema(schema.WithSections(sectionA, sectionB))
```

#### values package
- `SectionValues.Layer` → `SectionValues.Section`
- `GetDefaultParameterLayer` → `DefaultSectionValues` (or `DefaultSection`)

Example:
```go
vals.DefaultSectionValues()
```

#### sources package
Rename all `*Layer*` helpers to `*Section*` equivalents:
- `WhitelistLayers` → `WhitelistSections`
- `WrapWithLayerModifyingHandler` → `WrapWithSchemaModifyingHandler`

Example:
```go
sources.WhitelistSections(schema_, []string{"default"})
```

### 2) Remove **Parameter** from exported identifiers
This is optional but strongly recommended if “field” is the canonical noun.

#### fields package
- `ParsedParameter` → `FieldValue`
- `ParsedParameters` → `FieldValues`
- `ParameterType` → `FieldType` (if exported)

Example:
```go
vals := fields.NewFieldValues()
```

#### values package
- `WithParameterValue` → `WithFieldValue`
- Any error strings that say “parameter” → “field”

Example:
```go
values.WithFieldValue("name", "demo")
```

#### tags
- `glazed:"..."` → `glazed:"..."` or `glazed.field:"..."`

Example:
```go
// preferred
Name string `glazed:"name"`
```

### 3) Resolve the **dual Section** problem in values
**Preferred path:** remove `values.Section` by breaking the cycle.

Plan:
- Move cobra parsing/default initialization helpers out of `schema` into `sources`.
- Let `values` import `schema` directly.
- Replace `values.Section` interface with `schema.Section` in public types.

Example:
```go
// after cycle break
 type SectionValues struct {
     Section schema.Section
     Fields  *fields.FieldValues
 }
```

Fallback (if cycle remains):
- rename `values.Section` → `values.SchemaSection` and ensure it’s not promoted in docs.

### 4) Rename `InitializeStruct` → `DecodeSectionInto`
User direction prefers decode semantics, and the current verb is ambiguous.

Example:
```go
if err := values.DecodeSectionInto(vals, schema.DefaultSectionSlug, &cfg); err != nil { ... }
```

### 5) Rename `Parsed*` → `*Value(s)`
Reflects that these are resolved, typed values with provenance.

Example (current):
```go
parsed := fields.NewParsedParameters()
```
Proposed:
```go
values := fields.NewFieldValues()
```

---

## Design Decisions
1) **No synonyms in exported names**: once “Schema/Section/Field/Values” is chosen, all exported API must follow it.
2) **No backward compatibility**: no alias packages, no deprecated wrappers.
3) **Decode is the canonical verb**: all struct hydration uses `Decode*` naming.
4) **Single Section concept**: remove the duplicate `values.Section` export (preferred path).

---

## Alternatives Considered
- **Keep Parameters as noun**: rejected because package `fields` and `WithFields` already push “field” vocabulary, causing mixed terminology.
- **Keep Layer in specific contexts**: rejected because it communicates a distinct concept that no longer exists.
- **Keep InitializeStruct**: rejected; “initialize” collides with defaults initialization and is too vague.

---

## Migration Tooling Plan
We already have a Go-based rename tool from GL-001:
- `glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/rename_glazed_api.go`

### Improvements to leverage it further
1) **Add mapping tables for new renames**
   - Layer → Section: `AppendLayers` → `AppendSections`, etc.
   - Parameter → Field: `ParsedParameters` → `FieldValues`, etc.
2) **AST-first, text-second**
   - Use `go/ast` to rename identifiers and selectors.
   - Use targeted text replacement only for docs/strings.
3) **Export a report**
   - Keep JSON report of replaced identifiers.
   - Attach it to ticket under `various/` for audit.

### Suggested pipeline (no-compat)
1) Update rename tool mapping.
2) Run tool across `pkg/`, `cmd/`, `examples/`.
3) Run `rg -n -i "parameter|layer"` to verify removal.
4) Update docs.
5) `go test ./...`.
6) Commit in a single no-compat change set.

---

## Implementation Plan (step-by-step)
1) **Inventory + classify**
   - Use `various/01-parameter-layer-mentions.txt` to classify exported identifiers vs doc strings.
2) **Decide final vocabulary**
   - Confirm “field” vs “parameter” for public nouns.
3) **Resolve Section duplication**
   - Break schema/values cycle by moving cobra parsing into sources.
4) **Apply renames**
   - Schema/CommandDescription layer rename
   - Parameter → Field rename
   - InitializeStruct → DecodeSectionInto
5) **Docs + examples**
   - Replace mentions in `pkg/doc` and examples.
6) **Migration tooling**
   - Update rename tool and store its report.
7) **Validation**
   - `go test ./...`
   - `rg -n -i "parameter|layer" pkg cmd doc` to confirm removal.

---

## Open Questions
- Is “field” officially the canonical noun, or should we keep “parameter”? (Strong recommendation: field.)
- Are we willing to break the schema/values cycle now to eliminate the duplicate Section export?
- Should `values.Values` be renamed to `values.CommandValues` to reduce ambiguity?

---

## Code Snippet (happy path)
```go
sch := schema.NewSchema(schema.WithSections(defaultSection, loggingSection))
vals := values.New()

err := sources.Resolve(
    sch, vals,
    sources.FromDefaults(),
    sources.FromEnv("APP"),
    sources.FromCobra(cmd),
)

var cfg Config
err = values.DecodeSectionInto(vals, schema.DefaultSectionSlug, &cfg)
```

