---
Title: Add static website rendering to cmd/glaze
Ticket: GL-012-STATIC-RENDER
Status: active
Topics:
    - glazed
    - help
    - http
    - static-render
    - web
    - site-generator
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/build-web/main.go
      Note: Frontend build pipeline that the static export will reuse
    - Path: cmd/glaze/main.go
      Note: Root CLI wiring; the new static export command should be added alongside serve
    - Path: pkg/doc/topics/25-serving-help-over-http.md
      Note: Current end-user documentation for the live serve path
    - Path: pkg/help/help.go
      Note: Canonical HelpSystem lifecycle and FS loading behavior
    - Path: pkg/help/model/parse.go
      Note: Canonical markdown plus frontmatter parser
    - Path: pkg/help/model/section.go
      Note: Canonical section metadata contract for any exported snapshot
    - Path: pkg/help/server/serve.go
      Note: Current serve semantics and current home of shared path-loading helpers
    - Path: pkg/help/store/query.go
      Note: Existing ordering and filtering helpers to reuse for snapshot generation
    - Path: pkg/help/store/store.go
      Note: Canonical store and source of deterministic exported lists
    - Path: pkg/web/gen.go
      Note: go generate ownership for the shared web build
    - Path: pkg/web/static.go
      Note: Shared embedded SPA owner and static asset serving boundary
    - Path: ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/design-doc/01-help-browser-architecture-and-implementation-guide.md
      Note: Prior implementation ticket that established the current help-browser architecture
    - Path: ttmp/2026/04/08/GLAZE-HELP-REVIEW--pre-pr-code-review-glazed-help-browser/design-doc/02-project-report-glazed-serve.md
      Note: Follow-up review that explicitly raised static export as a future direction
    - Path: web/src/App.tsx
      Note: Current activeSlug local state highlights the deep-linking gap
    - Path: web/src/main.tsx
      Note: HashRouter mount makes static hosting easier
    - Path: web/src/services/api.ts
      Note: Current frontend transport layer and likely switch point for static mode
    - Path: web/src/types/index.ts
      Note: Frontend mirror of Go response types
    - Path: web/vite.config.ts
      Note: Current frontend build and dev proxy assumptions
ExternalSources: []
Summary: Research and design workspace for adding a static help-site export path to glaze alongside the existing live HTTP server.
LastUpdated: 2026-04-09T22:16:24.535135575-04:00
WhatFor: Track the architecture, tasks, diary, and delivery artifacts for static help-site rendering.
WhenToUse: Use when planning or implementing a static exported website for Glazed help content.
---


# Add static website rendering to cmd/glaze

## Overview

This ticket extends the recently-added `glaze serve` work into a second delivery mode: exporting the same help content as a static website that can be hosted without a Go server.

The working design in this ticket is to keep one canonical help model, one canonical markdown parser, and one canonical frontend, then add a new export pipeline that writes static JSON and frontend assets to disk.

## Key Links

- **Design guide**: `design-doc/01-static-help-website-rendering-architecture-and-implementation-guide.md`
- **Diary**: `reference/01-diary.md`
- **Tasks**: `tasks.md`
- **Related prior ticket**: `ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/`

## Status

Current status: **active**

## Topics

- glazed
- help
- http
- static-render
- web
- site-generator

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and implementation notes
- reference/ - Diary and quick-reference supporting material
- playbooks/ - Operational procedures and validation guides
- scripts/ - Temporary code and tooling
- various/ - Scratch notes and research
- archive/ - Deprecated or reference-only artifacts
