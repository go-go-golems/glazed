---
Title: Design Vault GitHub OIDC docs publishing auth for Glazed packages
Ticket: GG-20260502-VAULT-OIDC-DOCS-PUBLISH
Status: active
Topics:
    - glazed
    - docs
    - deploy
    - kubernetes
    - gitops
    - security
    - vault
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/help/loader/sources.go
      Note: Current SQLite package directory contract that authorized publishes materialize into.
    - Path: pkg/help/server/serve.go
      Note: Current docs serving entrypoint that Phase 2/3 publishing auth eventually feeds via registry/reload.
    - Path: ttmp/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE--design-docs-yolo-scapegoat-dev-multi-package-glazed-help-deployment/design-doc/01-docs-yolo-scapegoat-dev-multi-package-glazed-help-deployment-design-and-implementation-guide.md
      Note: Parent Phase 1 multi-package deployment design that this Phase 2/3 auth guide extends.
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T14:56:20.465705204-04:00
WhatFor: ""
WhenToUse: ""
---


# Design Vault GitHub OIDC docs publishing auth for Glazed packages

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- docs
- deploy
- kubernetes
- gitops
- security
- vault

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
