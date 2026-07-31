---
Title: Extend glazed-migrate to Handle All Removed Glazed Settings APIs
Ticket: GLZ-MIGRATE-ALL-APIS
Status: active
Topics:
    - glazed
    - cli
    - migration
    - settings
    - api-design
    - intern-guide
    - go-analysis
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Maps the full surface of Glazed settings APIs removed by GLZ-OUTPUT-FLAGS-CLEANUP and specifies the migration rules the glazed-migrate analyzer must learn beyond NewGlazedSchema, with an intern-ready design doc and implementation plan.
LastUpdated: 2026-07-28T12:00:00-04:00
WhatFor: Extend the glazed-migrate analyzer to cover every removed Glazed settings API so consuming repositories can be migrated automatically.
WhenToUse: Start here when extending glazed-migrate or migrating a consuming repository off removed Glazed settings APIs.
---

# Extend glazed-migrate to Handle All Removed Glazed Settings APIs

## Result (proposed)

The `glazed-migrate` analyzer currently rewrites exactly one removed API
(`NewGlazedSchema` → `NewStructuredOutputSection`). This ticket specifies the
nine rules (R1–R9) needed to cover the remaining 313 call sites across consuming
repositories, covering constructor deletion/rename, wrapper unwrapping, the
`output`→`format` flag rename, the `GlazedSlug`→`StructuredOutputSlug` constant
rename, and report-only diagnostics for the `Setup*` runtime helpers and
removed feature sections.

## Key documents

- [Intern guide and design doc](design-doc/01-intern-guide-extending-glazed-migrate-to-all-removed-glazed-settings-apis.md)
- [Investigation diary](reference/01-investigation-diary.md)
- [Tasks](tasks.md)
- [Changelog](changelog.md)

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- cli
- migration
- settings
- api-design
- intern-guide
- go-analysis

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
