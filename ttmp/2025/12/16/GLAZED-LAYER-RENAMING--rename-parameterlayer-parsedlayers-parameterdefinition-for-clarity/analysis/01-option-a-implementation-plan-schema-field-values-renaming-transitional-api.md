---
Title: 'Option A implementation plan: Schema/Field/Values renaming + transitional API'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
    - backwards-compatibility
DocType: analysis
Intent: working-document
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: ParameterLayer/ParameterLayers are the schema-ish core we’re renaming
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: ParsedLayer/ParsedLayers are the values core; InitializeStruct verb is key
    - Path: glazed/pkg/cmds/middlewares/layers.go
      Note: ExecuteMiddlewares and source middlewares naming (sources/resolvers)
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: ParameterDefinition is the field spec; candidate for FieldDefinition
    - Path: glazed/pkg/doc/topics/layers-guide.md
      Note: User-facing docs that would need vocabulary updates
    - Path: glazed/pkg/doc/tutorials/custom-layer.md
      Note: Tutorial that embeds current names
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/brainstorm/01-brainstorm-renaming-layers-parameters-api-for-clarity.md
      Note: Option A is grounded in this brainstorm; keep aligned
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/reference/01-debate-prep-candidates-and-questions-for-renaming-layers-parameters.md
      Note: Debate agenda/candidates for evaluating Option A
ExternalSources: []
Summary: Deep plan for implementing Bundle A (Schema/Field/Values), including compatibility layer design
LastUpdated: 2025-12-17T08:27:08.272682492-05:00
---


# Option A implementation plan: Schema/Field/Values renaming + transitional API

## Executive summary

Bundle A (“Schema / Field / Values”) is a vocabulary upgrade:

- “what is allowed?” → **schema**
- “one named thing in the schema” → **field**
- “what the user actually provided (merged from sources)” → **values**

The goal is to make Glazed’s API read naturally for *both* CLI users (flags/args) and config-driven users (env/config files), without breaking existing programs abruptly.

This document focuses on: **how to implement Bundle A in a migration-friendly way**, including a transitional API that lets legacy code keep compiling while new code adopts clearer names.

## Non-goals

- Not deciding Bundle A vs other bundles (that’s in the brainstorm doc).
- Not attempting a giant “rename everything in one PR”.
- Not changing the semantics of parsing/precedence/hydration (naming only).

## What we’re renaming (semantic mapping)

Core concepts today:

- `parameters.ParameterDefinition`: one **field spec** (type/default/help/required; flag or arg).
- `layers.ParameterLayer`: one **schema section** (group of field specs; slug/prefix/help metadata).
- `layers.ParsedLayer`: **values for one schema section**.
- `layers.ParsedLayers`: **values across all schema sections**.

Bundle A target vocabulary:

- **Schema**: “definition-time” objects (`Section`, `Sections`, `Field`, etc.)
- **Values**: “runtime resolved” objects (`SectionValues`, `Values`, etc.)

## Constraints and design choices

### Constraint 1: import-path renames are the most breaking

Changing `pkg/cmds/layers` → `pkg/cmds/schema` breaks every downstream import. Even with `go fix`, this is heavy.

### Constraint 2: Go type aliases help for names, but not for method renames

Type aliases (`type Section = layers.ParameterLayer`) are great for “new nouns”, but:

- You can’t add methods on an alias type.
- You can’t rename existing methods by aliasing.

So the transitional API needs a strategy for *method* vocabulary (`InitializeStruct`, etc.).

### Constraint 3: tags are part of the public API

Struct tags like `glazed.parameter:"host"` are “stringly-typed” but extremely sticky.

Recommendation: **do not rename tags** in the first phase. If we ever do, it should be additive (support both tags) with clear docs.

## Proposed implementation strategy (phased)

### Phase 0 (immediately): add “new names” as compatibility synonyms inside existing packages

This is the lowest-risk option. It improves readability without changing import paths.

Examples:

- In `pkg/cmds/parameters`:
  - `type FieldDefinition = ParameterDefinition`
  - `type FieldDefinitions = ParameterDefinitions` (if exists)
  - `func NewFieldDefinition(...) *FieldDefinition` (wrapper)
- In `pkg/cmds/layers`:
  - `type SchemaSection = ParameterLayer`
  - `type SchemaSections = ParameterLayers`
  - `type SectionValues = ParsedLayer`
  - `type Values = ParsedLayers`

Add `// Deprecated:` comments on the old names only after we’re confident.

**Pros**
- No import churn.
- Existing code continues to compile unchanged.

**Cons**
- Package names remain `layers`/`parameters`, so “schema vs values” still feels muddled unless type names are used consistently.

### Phase 1: add new, nicer packages as thin façades (optional but recommended)

Create new packages that re-export the existing implementation with Bundle A vocabulary:

- `pkg/cmds/schema` (façade over `cmds/layers` for schema concepts)
- `pkg/cmds/fields` (façade over `cmds/parameters` for field definitions)
- `pkg/cmds/values` (façade over `cmds/layers` parsed types)
- `pkg/cmds/sources` (façade over `cmds/middlewares` source middlewares)

Each façade should be *boring*:

- Prefer `type X = otherpkg.Y` aliases where possible.
- Provide minimal constructors/wrappers for the most common entry points.
- Avoid cycles: façades should depend “downward” (schema → layers/parameters), not vice versa.

**Pros**
- New code reads very clearly:
  - `schema.Section(...)`, `fields.String(...)`, `sources.Env(...)`, `values.DecodeInto(...)`
- Old code remains untouched.

**Cons**
- Adds maintenance surface (two packages worth of docs and names).
- Doesn’t automatically migrate everyone; it’s opt-in.

### Phase 2: migrate internal docs/examples to Bundle A vocabulary

Update tutorials and guides to prefer Bundle A names:

- `pkg/doc/topics/layers-guide.md`
- `pkg/doc/tutorials/custom-layer.md`
- examples under `cmd/examples/`

This is where the new vocabulary becomes “real” for users.

### Phase 3: deprecate old names (long tail)

Once Bundle A has baked:

- Add `// Deprecated:` on old names.
- Keep them for at least one major version (or indefinitely if Glazed is effectively “no major breaks”).

## Transitional API design (big section)

We want two things simultaneously:

1. **Legacy programs** keep compiling with `layers.ParameterLayer`, `parameters.ParameterDefinition`, `parsedLayers.InitializeStruct`, etc.
2. **New programs** can write code using Bundle A nouns and (ideally) verbs.

### Transitional API, level 1: noun-only transition (type synonyms)

This is the simplest and most realistic “first win”.

Implement in-place synonyms (Phase 0):

- New names are aliases to old names.
- New constructors are thin wrappers.

Example (conceptually):

```go
// in pkg/cmds/layers
type SchemaSection = ParameterLayer
type Values = ParsedLayers
```

**Compatibility**: perfect; no behavior changes.

**Downside**: call sites still use old verbs (`InitializeStruct`, `NewParameterLayer`).

### Transitional API, level 2: verb transition via helper functions (recommended)

Since we can’t rename methods cheaply, provide *new verbs as free functions* in façade packages.

Examples:

- `values.DecodeSectionInto(v *layers.ParsedLayers, section string, dst any) error`
  - internally calls `v.InitializeStruct(section, dst)`
- `schema.AddToCobra(cmd, sections)` wraps `ParameterLayers.AddToCobraCommand`

This gives a readable API without breaking anything.

**Key insight**: functions can be added without affecting method sets or requiring wrapper types.

### Transitional API, level 3: wrapper types for fluent methods (optional)

If we really want method names to change (e.g. `DecodeInto`), introduce *new defined wrapper types*:

```go
// pkg/cmds/values
type Values struct{ inner *layers.ParsedLayers }
func (v Values) DecodeInto(section string, dst any) error { return v.inner.InitializeStruct(section, dst) }
```

This adds extra ceremony (wrapping/unwrapping), but enables method naming.

Recommendation: only do this if we strongly prefer fluent OO-style APIs; otherwise stick to helper functions.

### Transitional API, level 4: package-level re-exports to migrate imports gradually

Provide new packages (`schema`, `fields`, `values`, `sources`) that users can adopt incrementally:

- Old code: unchanged imports.
- New code: uses new imports.
- Mixed codebases: can migrate file-by-file.

This is the best “real-world migration” story for large repos.

### Backwards compatibility checklist

For every rename we propose, check:

- **Import path changes?** If yes, must be additive (new pkg) first.
- **Type name changes?** Can be done via alias (additive).
- **Method name changes?** Prefer helper functions; wrapper types only if needed.
- **Struct tags changes?** Avoid; if desired, support both tags.
- **Serialized formats** (YAML, JSON): do not change keys by rename-only work.

## Concrete “Option A” naming mapping (recommended target)

This is a suggested mapping to implement first (minimal + high impact):

### Schema nouns

- `layers.ParameterLayer` → `layers.SchemaSection` (alias)
- `layers.ParameterLayers` → `layers.SchemaSections` (alias)
- `parameters.ParameterDefinition` → `parameters.FieldDefinition` (alias)

### Values nouns

- `layers.ParsedLayer` → `layers.SectionValues` (alias)
- `layers.ParsedLayers` → `layers.Values` (alias)

### Verbs (via helpers)

- `InitializeStruct` → `DecodeInto` / `DecodeSectionInto` (helper funcs)

## Work breakdown (suggested tasks)

1. Add alias types + constructors in existing packages (Phase 0).
2. Add façade packages (Phase 1) if we want import-path cleanliness.
3. Update docs/examples to prefer new names (Phase 2).
4. Add deprecations (Phase 3).
5. Provide a migration guide: “old → new” table and examples.

## Concrete transitional API proposal (recommended “starter set”)

This is a very specific, implementable slice that provides immediate clarity while preserving compatibility.

### 1) In-place aliases (no new packages)

Add these aliases in the existing packages. This gives new nouns without import churn.

#### `pkg/cmds/parameters`

- Add:
  - `type FieldDefinition = ParameterDefinition`
  - `type FieldDefinitionOption = ParameterDefinitionOption`
  - `func NewFieldDefinition(name string, t ParameterType, opts ...FieldDefinitionOption) *FieldDefinition`
- (Optional) add convenience constructors:
  - `func String(name string, opts ...FieldDefinitionOption) *FieldDefinition`
  - `func Int(name string, opts ...FieldDefinitionOption) *FieldDefinition`

#### `pkg/cmds/layers`

- Add:
  - `type SchemaSection = ParameterLayer`
  - `type SchemaSections = ParameterLayers`
  - `type SectionValues = ParsedLayer`
  - `type Values = ParsedLayers`

### 2) Helper functions for verbs (no wrapper types)

Add free functions (either in the same packages or in new façade packages) to improve verb clarity without changing method sets.

Examples:

- `layers.DecodeSectionInto(values *layers.ParsedLayers, sectionKey string, dst any) error`
  - implementation: `return values.InitializeStruct(sectionKey, dst)`
- `layers.DecodeDefaultInto(values *layers.ParsedLayers, dst any) error`
  - implementation: `return values.InitializeStruct(layers.DefaultSlug, dst)`

This pattern can be mirrored in a new package `pkg/cmds/values` if we want imports to reflect “values”.

### 3) Optional façade packages for clean imports

If we decide import paths should match semantics, create:

- `pkg/cmds/fields` → aliases/wrappers over `pkg/cmds/parameters`
- `pkg/cmds/schema` → aliases/wrappers over `pkg/cmds/layers` (schema side)
- `pkg/cmds/values` → aliases/wrappers over `pkg/cmds/layers` (values side)
- `pkg/cmds/sources` → wrappers over `pkg/cmds/middlewares`

These packages should:

- be tiny
- avoid introducing new behavior
- primarily exist to improve names and import readability

## Compatibility matrix (what breaks vs what doesn’t)

| Change | Legacy code impact | Notes |
|---|---|---|
| Add alias type names | None | Safe additive change |
| Add wrapper constructors | None | Safe additive change |
| Add helper functions | None | Safe additive change |
| Add façade packages | None | Safe additive change; new imports optional |
| Rename existing types | Breaking | Avoid until a major version (if ever) |
| Rename packages (move code) | Breaking | Avoid; prefer façades |
| Rename struct tag keys | Breaking | Avoid; if needed, support both |

## Risks

- “Schema” might imply strict validation; we should clarify in docs that Glazed schema is descriptive, not necessarily validating.
- If we create new packages, we must be careful about cyclic imports and duplicated docs.
- Two vocabularies will coexist for a while; docs must be explicit about “preferred names”.
