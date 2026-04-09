---
Title: 'Pre-PR Code Review: Glazed Help Browser'
Ticket: GLAZE-HELP-REVIEW
Status: active
Topics:
    - help-browser
    - code-review
    - glazed
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/help/dsl_bridge.go
      Note: O(N) temp-store-per-section performance bug in evaluatePredicate
    - Path: pkg/help/help.go
      Note: Backward-compat re-exports + Section wrapper + duplicated parsing
    - Path: pkg/help/model/section.go
      Note: Dead HelpSystem field
    - Path: pkg/help/server/handlers.go
      Note: Search bypasses FTS5 abstraction
    - Path: pkg/help/store/loader.go
      Note: Duplicated LoadFromMarkdown vs help.go
ExternalSources: []
Summary: 'Pre-PR review of the glazed help browser: session analysis + code review findings'
LastUpdated: 2026-04-08T17:46:51.654644641-04:00
WhatFor: Review the help browser code before PR submission
WhenToUse: When reviewing or continuing work on the help browser
---



# Pre-PR Code Review: Glazed Help Browser

## Overview

Pre-PR code review of the glazed help browser feature, which adds:
- HTTP server (`pkg/help/server/`) for browsing help sections via REST API
- React SPA frontend (`web/`) with RTK Query
- Build pipeline (`cmd/build-web/`) using Dagger + pnpm
- Embed pipeline (`pkg/web/`) for serving frontend from Go binary

The review analyzed both the 21-hour coding session (via go-minitrace) and the resulting code.
Found 2 critical bugs, 3 dead/duplicated code patterns, and several legacy patterns.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- help-browser
- code-review
- glazed

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
