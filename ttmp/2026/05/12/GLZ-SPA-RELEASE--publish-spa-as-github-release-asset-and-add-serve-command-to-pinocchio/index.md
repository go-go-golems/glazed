---
Title: Publish SPA as GitHub Release Asset and Add Serve Command to Pinocchio
Ticket: GLZ-SPA-RELEASE
Status: active
Topics:
    - help
    - serve
    - http
    - spa
    - release
    - goreleaser
    - pinocchio
    - distribution
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: .github/workflows/release.yaml
      Note: GoReleaser split release
    - Path: .goreleaser.yaml
      Note: Needs tar czf in before hooks and release.extra_files
    - Path: pkg/web/embed.go
      Note: Current embed mechanism
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-12T15:54:40.048118715-04:00
WhatFor: ""
WhenToUse: ""
---











# Publish SPA as GitHub Release Asset and Add Serve Command to Pinocchio

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
- release
- goreleaser
- pinocchio
- distribution

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
