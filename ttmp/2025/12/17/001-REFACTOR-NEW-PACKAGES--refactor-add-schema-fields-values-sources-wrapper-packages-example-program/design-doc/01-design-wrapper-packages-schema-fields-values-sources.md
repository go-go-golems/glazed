---
Title: 'Design: wrapper packages (schema/fields/values/sources)'
Ticket: 001-REFACTOR-NEW-PACKAGES
Status: active
Topics:
    - glazed
    - api-design
    - refactor
    - backwards-compatibility
    - migration
    - schema
    - examples
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/layers/layer-impl.go
      Note: NewParameterLayer/ParseLayerFromCobraCommand and prefix semantics referenced in the design
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/debate/02-debate-round-1-independent-composer-analysis.md
      Note: Debate synthesis that motivated this ticket
ExternalSources: []
Summary: Design for additive facade packages (schema/fields/values/sources) built from type aliases + wrapper functions, plus an example program validating env+cobra parsing and struct decoding.
LastUpdated: 2025-12-17T09:01:05.389827519-05:00
---


# Design: wrapper packages (schema/fields/values/sources)

## Executive Summary

We will introduce four **new, additive** packages under `glazed/pkg/cmds/`:

- `schema`: schema sections (today: `schema.Section` + `schema.Schema`)
- `fields`: field definitions and field types (today: `fields.Definition`, `fields.Type`)
- `values`: resolved values (today: `values.SectionValues` + `values.Values`) + “decode” helpers
- `sources`: value sources / resolvers (today: `cmds/middlewares`) + “execute chain” helpers

These packages will be implemented using **Go type aliases** (`type X = Y`) and **wrapper functions** (not wrapper types), so:

- Existing code remains unchanged and compatible.
- New code can opt into cleaner vocabulary and import paths.
- There is **no runtime overhead** (aliases are zero-cost; wrappers are thin).

We will also add an **example program** under `glazed/cmd/examples/` that defines a command with multiple schema sections, resolves values from **defaults + config + env + cobra flags/args**, and decodes resolved values into Go structs. This example doubles as an acceptance test for the new packages.

## Problem Statement

Glazed currently exposes core concepts via packages and names that are semantically overloaded or hard to discover:

- “Schema” is expressed as `schema.Section(s)`, even though many users think in terms of **schema sections** and **fields**.
- Resolved values are expressed as `values.SectionValues(s)`, which does not communicate “values” at the API edge.
- Resolution logic lives in `cmds/middlewares`, but these are primarily **value sources/resolvers** (cobra, env, config, defaults), not generic middleware.

The result is a higher onboarding cost and lower API clarity, especially in tutorials and example programs.

At the same time, the existing APIs are widely used, so “rename in place” is expensive and risky. We need an approach that improves clarity **without breaking** downstream code.

## Proposed Solution

### 1) Add façade packages with Option A vocabulary

We add new packages that re-export the current implementation with better names.

#### `glazed/pkg/cmds/schema`

Primary goal: make it easy to talk about “schema sections” without importing `layers`.

Proposed surface (illustrative):

- **Type aliases**
  - `type Section = schema.Section`
  - `type Schema = schema.Schema`
  - `type SectionImpl = schema.SectionImpl` (the common concrete impl)
  - `type SectionOption = schema.SectionOptions`
  - `type SchemaOption = schema.SchemaOption`
- **Constructor wrappers**
  - `func NewSection(slug, name string, opts ...SectionOption) (*SectionImpl, error)` → `schema.NewSection`
  - `func NewSchema(opts ...SchemaOption) *Schema` → `schema.NewSchema`
  - `func WithSections(sections ...Section) SchemaOption` → `layers.WithLayers`
  - `func NewGlazedSchema(opts ...settings.GlazeParameterLayerOption) (Section, error)` → `settings.NewGlazedParameterLayers`
- **Constants**
  - `const DefaultSlug = schema.DefaultSlug`

#### `glazed/pkg/cmds/fields`

Primary goal: clarify that `ParameterDefinition` is a **field definition**, and group field/type helpers under a “fields” concept.

Proposed surface:

- **Type aliases**
  - `type Definition = fields.Definition`
  - `type Definitions = fields.Definitions`
  - `type Type = fields.Type`
  - `type Option = fields.DefinitionOption`
- **Constructor wrappers**
  - `func New(name string, t Type, opts ...Option) *Definition` → `fields.New`
  - `func NewDefinitions(opts ...func(*Definitions)) *Definitions` (optional sugar; can be deferred)
- **Option re-exports**
  - `WithHelp`, `WithShortFlag`, `WithDefault`, `WithChoices`, `WithRequired`, `WithIsArgument`
- **Type constant re-exports**
  - `TypeString`, `TypeBool`, `TypeInteger`, … (mirror `fields.Type*`)

#### `glazed/pkg/cmds/values`

Primary goal: rename “parsed layers” to “values” and provide “decode” verbs without renaming methods.

Proposed surface:

- **Type aliases**
  - `type SectionValues = values.SectionValues`
  - `type Values = values.Values`
  - `type ValuesOption = values.ValuesOption`
- **Constructor wrappers**
  - `func New(opts ...ValuesOption) *Values` → `values.New`
- **Verb wrappers (functions)**
  - `func DecodeInto(v *SectionValues, dst any) error` → `v.InitializeStruct(dst)`
  - `func DecodeSectionInto(vs *Values, sectionSlug string, dst any) error` → `vs.InitializeStruct(sectionSlug, dst)`
  - (optional) `func AsMap(vs *Values) map[string]any` → `vs.GetDataMap()`

#### `glazed/pkg/cmds/sources`

Primary goal: expose the “source chain” under a name that matches semantics, and make it ergonomic to execute a precedence chain.

Proposed surface:

- **Type aliases**
  - `type Middleware = middlewares.Middleware`
- **Wrapper functions for common sources**
  - `FromCobra(cmd *cobra.Command, opts ...parameters.ParseStepOption) Middleware` → `sources.FromCobra`
  - `FromArgs(args []string, opts ...parameters.ParseStepOption) Middleware` → `sources.FromArgs`
  - `FromEnv(prefix string, opts ...parameters.ParseStepOption) Middleware` → `sources.FromEnv`
  - `FromDefaults(opts ...parameters.ParseStepOption) Middleware` → `sources.FromDefaults`
  - `FromConfigFilesForCobra(...) Middleware` → `middlewares.LoadParametersFromResolvedFilesForCobra` (optional; phase 2)
- **Execution helper**
  - `func Execute(schema *schema.Schema, vals *values.Values, ms ...Middleware) error` → `sources.Execute`

### 2) Add an example program as an acceptance test

We add `glazed/cmd/examples/refactor-new-packages/` with a runnable program that:

- Defines a command with **2–3 schema sections** using `schema` + `fields`
- Configures cobra parsing using the existing Glazed cobra bridge
- Resolves values from:
  - defaults (lowest precedence)
  - optional config file(s)
  - env (with `AppName` prefix)
  - cobra flags + args (highest precedence)
- Decodes resolved values into section-specific structs using `values.DecodeSectionInto` (or underlying method)

This example intentionally mirrors the existing precedence chain in the cobra parser (`cli.CobraParserConfig`), which already composes `sources.FromCobra`, `sources.FromArgs`, `sources.FromEnv`, config file loading, and defaults.

## Design Decisions

1. **Use type aliases, not new wrapper types**  
   Aliases preserve identity and method sets without conversions and without runtime overhead.

2. **Use wrapper functions to introduce new verbs**  
   Go does not allow adding methods to an alias type; free functions like `values.DecodeInto` provide the improved vocabulary without breaking callers.

3. **Keep existing packages untouched (additive-only)**  
   This prevents import cycles and avoids destabilizing existing users. The new packages depend on `layers/parameters/middlewares`, never the other way around.

4. **Keep struct tags unchanged**  
   We keep `glazed` as-is; this ticket is about package/type vocabulary, not tags.

5. **Start with a minimal but coherent surface**  
   We focus on the “happy path” constructors and source chain pieces needed by the example program first, then expand coverage as needed.

## Alternatives Considered

1. **Rename existing packages/types in place**  
   Rejected: high breakage risk, large downstream churn, and a poor incremental adoption story.

2. **Wrapper structs to add methods (non-alias types)**  
   Rejected (for phase 1): forces explicit conversions, complicates interop with the existing API, and adds cognitive overhead.

3. **Only add aliases inside existing packages**  
   Rejected: improves names but does not improve import discoverability (`layers` still doesn’t read as “schema”).

## Implementation Plan

See: [Implementation plan: wrapper packages + example program](../planning/01-implementation-plan-wrapper-packages-example-program.md)

## Open Questions

1. **Exact naming of the “all values” container**  
   Keep it as `values.Values` (simple) vs something more explicit like `values.All` or `values.Set`.

2. **How much of `fields.Type*` to re-export in `fields`**  
   All for completeness vs a minimal subset initially.

3. **Should `sources` provide a “default chain builder”**  
   e.g. `sources.DefaultCobraChain(appName, cmd, args, ...)` vs keeping wrappers primitive and letting `cli` remain the canonical chain builder.

## References

- Debate prep: `glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/reference/01-debate-prep-candidates-and-questions-for-renaming-layers-parameters.md`
- Independent debate round: `glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/debate/02-debate-round-1-independent-composer-analysis.md`
- Existing implementations to wrap:
  - `glazed/pkg/cmds/layers/*`
  - `glazed/pkg/cmds/parameters/*`
  - `glazed/pkg/cmds/middlewares/*`
  - `glazed/pkg/cli/cobra-parser.go`
