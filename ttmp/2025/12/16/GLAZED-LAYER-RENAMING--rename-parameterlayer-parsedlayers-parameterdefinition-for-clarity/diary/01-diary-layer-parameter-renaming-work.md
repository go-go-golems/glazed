---
Title: 'Diary: layer/parameter renaming work'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
DocType: diary
Intent: working-document
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/analysis/01-option-a-implementation-plan-schema-field-values-renaming-transitional-api.md
      Note: Deep Option A plan referenced by diary
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/brainstorm/01-brainstorm-renaming-layers-parameters-api-for-clarity.md
      Note: Brainstorm snapshot referenced by diary
    - Path: glazed/ttmp/vocabulary.yaml
      Note: Tracks new docTypes/intents/topics used by this ticket
ExternalSources: []
Summary: Chronological log of work on naming/renaming ParameterLayer/ParsedLayers/ParameterDefinition
LastUpdated: 2025-12-17T08:30:37.551566127-05:00
---


# Diary: layer/parameter renaming work

## 2025-12-17

### Ticket created + initial brainstorm

- Created ticket workspace: `GLAZED-LAYER-RENAMING`.
- Wrote a naming brainstorm exploring multiple coherent bundles (Schema/Field/Values, Options, Inputs, Config) and noting peripheral naming (`middlewares`, `InitializeStruct`) also contributes to confusion.

### Vocabulary updates (docmgr)

Updated `glazed/ttmp/vocabulary.yaml`:

- Added glazed-relevant topics (commands/cli/cobra/layers/parameters/schema/values/sources/parsing/hydration/migration/etc.).
- Added doctypes: `brainstorm`, `diary`.
- Added intents: `working-document`, `reviewed`, `deprecated`.

### Brainstorm reclassified (doctype + intent)

- Moved brainstorm doc from `design/` → `brainstorm/`.
- Updated frontmatter to `DocType: brainstorm` and `Intent: reviewed`.
- Preserved `RelatedFiles` entries pointing at core code (`layers`, `parameters`) and user-facing docs (guides/tutorials).

### Deep analysis: Option A implementation plan (working document)

Added a detailed analysis doc describing how to implement Bundle A (Schema/Field/Values), with a focus on **transitional compatibility**:

- Phase 0: in-place alias names (no import churn).
- Phase 1: optional façade packages (`schema`, `fields`, `values`, `sources`) that re-export the old implementation.
- Verb renames via helper functions (since method renames aren’t feasible with pure type aliases).
- Compatibility matrix clarifying what is additive vs breaking.

### Commits (traceability)

- `3662240`: initial ticket + brainstorm
- `61ebe5d`: vocabulary expansion + Option A analysis doc + updated links
- `7808a2e`: finalize brainstorm doctype move (delete old `design/` doc + restore RelatedFiles)
