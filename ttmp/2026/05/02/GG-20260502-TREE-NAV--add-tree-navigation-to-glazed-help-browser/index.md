---
Title: Add tree navigation to Glazed help browser
Ticket: GG-20260502-TREE-NAV
Status: active
Topics:
    - glazed
    - help
    - frontend
    - react
    - ui
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: web/src/App.test.tsx
      Note: Regression test for subsection tree clicks invoking scrollIntoView.
    - Path: web/src/App.tsx
      Note: Route/hash-driven Markdown pane scrolling for document and subsection tree selection.
    - Path: web/src/components/DocumentationTree/DocumentationTree.stories.tsx
      Note: Storybook examples for tree grouping
    - Path: web/src/components/DocumentationTree/DocumentationTree.tsx
      Note: Tree component markup now separates row part from data-kind so subsection rows inherit tree styling.
    - Path: web/src/components/DocumentationTree/styles/documentation-tree.css
      Note: Compact subsection row styling
    - Path: web/src/components/NavigationModeToggle/NavigationModeToggle.stories.tsx
      Note: Storybook examples for Tree and Search toggle states.
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T12:17:05.601847782-04:00
WhatFor: ""
WhenToUse: ""
---



# Add tree navigation to Glazed help browser

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
- frontend
- react
- ui

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
