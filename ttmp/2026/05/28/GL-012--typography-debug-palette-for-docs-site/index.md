---
Title: Typography Debug Palette for Docs Site
Ticket: GL-012
Status: active
Topics:
    - typography
    - css
    - frontend
    - debug-tooling
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: web/src/App.tsx
      Note: Root app component — where the debug palette toggle would be wired
    - Path: web/src/components/Markdown/styles/markdown.css
      Note: Prose typography styles (h1-h3
    - Path: web/src/components/SectionView/styles/section-view.css
      Note: Section view layout and header typography (heading
    - Path: web/src/components/TypographyPalette/TypographyPalette.tsx
      Note: Main palette component
    - Path: web/src/components/TypographyPalette/css-override-engine.ts
      Note: CSS override and export engine
    - Path: web/src/services/api.ts
      Note: RTK Query API — existing state management pattern to follow
    - Path: web/src/store.ts
      Note: Redux store — where palette state slice would live
    - Path: web/src/store/typographyPaletteSlice.ts
      Note: Redux slice with persistence
    - Path: web/src/styles/global.css
      Note: Root CSS variables for typography (font-family
    - Path: web/src/types/index.ts
      Note: TypeScript interfaces — where palette types would be added
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-28T09:32:18.962371224-04:00
WhatFor: ""
WhenToUse: ""
---



# Typography Debug Palette for Docs Site

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- typography
- css
- frontend
- debug-tooling

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
