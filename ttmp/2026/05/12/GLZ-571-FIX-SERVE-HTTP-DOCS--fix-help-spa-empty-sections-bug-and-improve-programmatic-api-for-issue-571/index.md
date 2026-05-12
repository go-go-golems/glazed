---
Title: 'Fix Help SPA Empty Sections Bug and Improve Programmatic API for Issue #571'
Ticket: GLZ-571-FIX-SERVE-HTTP-DOCS
Status: active
Topics:
    - help
    - serve
    - http
    - spa
    - api
    - bug
    - paper-cut
    - documentation
    - intern-guide
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/build-web/main.go
      Note: Dagger-based SPA build tool
    - Path: pkg/doc/topics/25-serving-help-over-http.md
      Note: Help entry with programmatic examples missing SetDefaultPackage
    - Path: pkg/help/server/handlers.go
      Note: handleListPackages normalizes empty package_name to default
    - Path: pkg/help/server/serve.go
      Note: NewServeHandler needs auto-assign SetDefaultPackage call
    - Path: pkg/help/store/store.go
      Note: SetDefaultPackage assigns package_name to sections without one
    - Path: pkg/web/embed.go
      Note: SPA assets embedded via go:embed
    - Path: pkg/web/embed_none.go
      Note: Fallback for non-embed builds
    - Path: pkg/web/gen.go
      Note: go:generate directive for building SPA frontend
    - Path: web/src/App.tsx
      Note: SPA auto-selects defaultPackage on startup
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-12T10:01:34.960586841-04:00
WhatFor: ""
WhenToUse: ""
---










# Fix Help SPA Empty Sections Bug and Improve Programmatic API for Issue #571

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- help
- serve
- http
- spa
- api
- bug
- paper-cut
- documentation
- intern-guide

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
