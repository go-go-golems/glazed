---
Title: Proper docsctl URL structure with package/version/page/section navigation
Ticket: DOCSCTL-URL-NAV
Status: active
Topics:
    - docsctl
    - urls
    - routing
    - spa
    - go
    - typescript
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Replace HashRouter with BrowserRouter and add package/version/page/section to URLs so that docs.yolo.scapegoat.dev supports deep linking, bookmarking, and heading navigation."
LastUpdated: 2026-05-25T20:20:45.580778188-04:00
WhatFor: ""
WhenToUse: ""
---

# Proper docsctl URL structure with package/version/page/section navigation

## Overview

The docs.yolo.scapegoat.dev documentation site currently uses hash-based routing that prevents deep linking and bookmarking. This ticket redesigns the URL structure to encode package, version, page slug, and heading section in the URL path: `/{package}/{version}/sections/{slug}#{heading-id}`. See the design doc for the full analysis and 3-phase implementation plan.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docsctl
- urls
- routing
- spa
- go
- typescript

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
