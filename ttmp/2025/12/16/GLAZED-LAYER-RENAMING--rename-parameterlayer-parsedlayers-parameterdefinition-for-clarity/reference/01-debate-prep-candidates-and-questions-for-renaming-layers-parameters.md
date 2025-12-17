---
Title: 'Debate prep: candidates and questions for renaming layers/parameters'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
DocType: reference
Intent: working-document
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: Debate touches ParameterLayer/ParameterLayers naming
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: Debate touches ParsedLayer/ParsedLayers naming + InitializeStruct
    - Path: glazed/pkg/cmds/middlewares/layers.go
      Note: Debate touches middlewares vs sources naming
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: Debate touches ParameterDefinition naming
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Example of current CLI registration pattern (BuildCobraCommand/CobraParserConfig) discussed in debate
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/analysis/01-option-a-implementation-plan-schema-field-values-renaming-transitional-api.md
      Note: Option A plan referenced by debate
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/brainstorm/01-brainstorm-renaming-layers-parameters-api-for-clarity.md
      Note: Source of naming bundles
ExternalSources: []
Summary: 'Debate setup: candidate roster and questions for the naming/renaming effort (no perf/security)'
LastUpdated: 2025-12-17T08:33:08.627392269-05:00
---



# Debate prep: candidates and questions for renaming layers/parameters

## Goal

Set up a presidential-style debate for **GLAZED-LAYER-RENAMING** by defining:

- **Debate questions / resolutions** (what we’re actually trying to decide)
- **Candidates** (human personas + personified code entities)
- **Candidate research questions** (what evidence each needs)

Scope constraints:

- **Exclude** performance and security arguments.
- **Focus** on API clarity, semantics, compatibility, migration cost, docs, and developer experience.

## Context

We’re exploring renames like:

- `layers.ParameterLayer` → “schema section”
- `parameters.ParameterDefinition` → “field definition”
- `layers.ParsedLayers` → “values / resolved values”

But there are multiple viable paths:

- Rename types only vs rename packages too
- Provide aliases vs add façade packages vs add wrapper types
- Improve nouns only vs also improve verbs (`InitializeStruct` → `DecodeInto`)

## Quick Reference

### Debate questions (resolutions)

#### Primary resolutions

1. **Scope of rename**: Should we rename **types only**, or also introduce **new packages** (`schema`, `fields`, `values`, `sources`)?
2. **Transitional strategy**: What is the preferred compatibility path?
   - type aliases only
   - helper functions for new verbs
   - wrapper types for fluent methods
   - façade packages re-exporting old implementation
3. **Verb cleanup**: Should we introduce “decode/bind” vocabulary (`DecodeInto`, `DecodeSectionInto`) to complement the new nouns?
4. **Peripheral naming**: Should `cmds/middlewares` be conceptually renamed (at least at API level) to **sources/resolvers**, even if we don’t change the import path?
5. **Command definition surface**: How should we deal with Glazed command definitions (e.g. `cmds.CommandDescription`, layers/flags/args composition) under the new vocabulary?
   - Do we keep “layer” terminology in command descriptions, or rename to “schema sections” at the API edge?
   - What is the desired end-state for “flags vs args” definition APIs?
6. **CLI registration / Cobra wrapping**: What is a better high-level CLI registration story for Glazed commands?
   - Today, tutorials like `glazed/pkg/doc/tutorials/05-build-first-command.md` show `cli.BuildCobraCommand` + `CobraParserConfig`.
   - Should we introduce a higher-level “app/registry” or “CLI builder” that standardizes registration, parsing, and help wiring?

#### Secondary questions (tie-breakers)

7. **Where do new names live?**
   - in-place (within `cmds/layers` and `cmds/parameters`)
   - in new packages (recommended for cleaner imports)
8. **What is the “unit” of values?**
   - `ParsedLayer` is “section values”
   - `ParsedLayers` is “all values”
   - do we also want a cross-layer merged view name?
9. **What do we call “layer” itself?**
   - keep “layer” as the grouping word
   - or replace with “section”, “group”, or “namespace” (docs vs code)?
10. **Do we keep the struct tag `glazed.parameter`?**
   - (Debate should cover naming clarity, but the current plan is “do not rename tags”; confirm.)

### Candidate roster

#### Human personas

- **The Pragmatist**
  - **Cares about**: minimal breakage, incremental adoption, low cognitive load
  - **Default stance**: aliases + doc updates; avoid package moves

- **The Architect**
  - **Cares about**: coherent domain model, consistent vocabulary across modules
  - **Default stance**: Bundle A + façade packages (`schema/fields/values/sources`) for clean imports

- **The Migration Engineer**
  - **Cares about**: compatibility guarantees, deprecation policy, upgrade path for downstream repos
  - **Default stance**: additive approach, explicit matrix of breaking vs non-breaking changes

- **The Doc Maintainer**
  - **Cares about**: tutorial clarity, searchability, naming consistency in docs
  - **Default stance**: standardize docs/examples first; ensure glossary is coherent and discoverable

- **The New Developer**
  - **Cares about**: learnability, intuitive names, “what do I call what?”
  - **Default stance**: strongest preference for Schema/Field/Values (or Options) with clearer verbs

#### Code entity personas (personified modules)

- **`pkg/cmds/layers` (“The Layer Librarian”)**
  - **Key surfaces**: `ParameterLayer`, `ParameterLayers`, `ParsedLayer`, `ParsedLayers`, `DefaultSlug`
  - **Likely argument**: “I already encode schema+values; splitting me into schema vs values might be cleaner, but don’t break everyone.”

- **`pkg/cmds/parameters` (“The Field Spec”)**
  - **Key surfaces**: `ParameterDefinition`, parsing/validation, struct init
  - **Likely argument**: “I’m a field spec, not a runtime value; my name should reflect that.”

- **`pkg/cmds/middlewares` (“The Source Chain”)**
  - **Key surfaces**: `SetFromDefaults`, `UpdateFromEnv`, `LoadParametersFromFiles`, `ParseFromCobraCommand`, `GatherArguments`, `ExecuteMiddlewares`
  - **Likely argument**: “I’m fundamentally about sources/resolution, not generic middleware.”

- **`pkg/cli` (“The Cobra Bridge”)**
  - **Key surfaces**: Cobra parser config, bridge functions, integration patterns
  - **Likely argument**: “Names must map cleanly to CLI mental model, but still support config/env.”

- **`pkg/appconfig` (“The AppConfig Boundary”)**
  - **Key surfaces**: `appconfig.Parser`, layer registration, parsing chain, hydration semantics
  - **Likely argument**: “Clear schema/values naming makes my API easier to explain; I’ll benefit from new nouns.”

### Candidate research questions (no perf/security)

#### Pragmatist

- What’s the minimum additive set of aliases/wrappers that yields a noticeable clarity win?
- How much downstream breakage happens if we rename packages vs only types?
- Which docs/examples are the top confusion points today?

#### Architect

- Which vocabulary bundle yields the most coherent end-to-end story across `layers/parameters/middlewares/cli/appconfig`?
- Can we design façade packages without import cycles and with minimal duplication?
- What are the “canonical” verbs we want (`DecodeInto` vs `BindInto`)?

#### Migration Engineer

- Can we implement noun renames with `type X = Y` aliases everywhere we need?
- Where do method renames force helper funcs or wrapper types?
- What deprecation policy is realistic (keep old forever vs timed removal)?

#### Doc Maintainer

- Where in docs/tutorials do the names appear most often? What should the glossary become?
- Should we update doc types/intent usage (working-document vs reviewed vs deprecated) consistently?

#### New Developer

- If you see `ParameterLayer`, what do you think it means? Same for `ParsedLayers`?
- Are “schema/field/values” discoverable without reading the implementation?
- Which of the bundles is most intuitive without prior context?

## Usage Examples

- Use the **Primary resolutions** list as the agenda for Debate Round 1.
- Use the **candidate research questions** to drive concrete `grep`/code reading before writing opening statements.
- Keep arguments grounded in compatibility and clarity; **do not** introduce perf/security.

## Related

- Brainstorm: `brainstorm/01-brainstorm-renaming-layers-parameters-api-for-clarity.md`
- Option A plan: `analysis/01-option-a-implementation-plan-schema-field-values-renaming-transitional-api.md`
