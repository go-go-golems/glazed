---
Title: Add HTTP server for glazed help entries with React frontend
Ticket: GL-011-HELP-BROWSER
Status: active
Topics:
    - glazed
    - help
    - http
    - react
    - vite
    - rtk-query
    - storybook
    - dagger
    - web
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/build-web/main.go
      Note: Shared frontend builder; outputs to pkg/web/dist
    - Path: cmd/glaze/main.go
      Note: Main CLI entrypoint; `glaze serve` is the supported runtime surface
    - Path: pkg/help/cmd/cobra.go
      Note: Existing Cobra help command integration - pattern for adding serve subcommand
    - Path: pkg/help/server/serve.go
      Note: Shared serve command, loader, root-handler composition, and prefix mounting helpers
    - Path: pkg/help/help.go
      Note: Core HelpSystem struct and API - the Go backend is exposed via HTTP
    - Path: pkg/help/model/section.go
      Note: Section model with JSON tags - maps directly to API response shape
    - Path: pkg/help/store/query.go
      Note: QueryCompiler for advanced filtering/search
    - Path: pkg/help/store/store.go
      Note: SQLite store with List/GetBySlug/Query - backend for the API
    - Path: pkg/web/gen.go
      Note: go:generate entrypoint that rebuilds the shared embedded frontend
    - Path: pkg/web/static.go
      Note: Shared embedded frontend package and SPA handler (single source of truth)
    - Path: ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/sources/local/glazed-docs-browser(2).jsx
      Note: Source JSX prototype to decompose into modular React
ExternalSources:
    - local:glazed-docs-browser(2).jsx
Summary: ""
LastUpdated: 2026-04-07T20:26:14.571308584-04:00
WhatFor: ""
WhenToUse: ""
---




# Add HTTP server for glazed help entries with React frontend

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- help
- http
- react
- vite
- rtk-query
- storybook
- dagger
- web

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
