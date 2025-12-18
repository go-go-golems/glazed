---
Title: 'Refactor: add schema/fields/values/sources wrapper packages + example program'
Ticket: 001-REFACTOR-NEW-PACKAGES
Status: complete
Topics:
    - glazed
    - api-design
    - refactor
    - backwards-compatibility
    - migration
    - schema
    - examples
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Add new wrapper packages (schema/fields/values/sources) via type aliases + helpers, plus an example program exercising env+cobra parsing and struct decoding.
LastUpdated: 2025-12-18T15:37:24.726743209-05:00
---


# Refactor: add schema/fields/values/sources wrapper packages + example program

## Overview

This ticket introduces **new, future-facing Glazed packages** that expose the Option A vocabulary discussed in **GLAZED-LAYER-RENAMING**:

- **`schema`**: “what inputs exist” (sections + field definitions)
- **`values`**: “what values were resolved” (per-section values + all values)
- **`sources`**: “where values come from” (cobra flags/args, env, defaults, config files)
- **`fields`**: field/flag/arg definitions (light façade over `parameters`)

These packages are **additive** and implemented as **type aliases + wrapper functions**, so existing code using `layers`, `parameters`, and `middlewares` continues to work unchanged.

We also add an **example program** that defines a command with a few sections, parses from **env + cobra**, and decodes section values into structs—serving as a concrete acceptance test for the wrapper packages.

## Key Links

- Design doc: [Design: wrapper packages (schema/fields/values/sources)](./design-doc/01-design-wrapper-packages-schema-fields-values-sources.md)
- Implementation plan: [Implementation plan: wrapper packages + example program](./planning/01-implementation-plan-wrapper-packages-example-program.md)
- Diary: [Diary](./diary/01-diary.md)
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- api-design
- refactor
- backwards-compatibility
- migration
- schema
- examples

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
