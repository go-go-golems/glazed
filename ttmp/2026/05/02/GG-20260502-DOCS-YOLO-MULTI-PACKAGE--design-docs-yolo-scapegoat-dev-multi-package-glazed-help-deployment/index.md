---
Title: Design docs.yolo.scapegoat.dev multi-package Glazed help deployment
Ticket: GG-20260502-DOCS-YOLO-MULTI-PACKAGE
Status: active
Topics:
    - glazed
    - docs
    - deploy
    - kubernetes
    - gitops
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/docs-registry/main.go
      Note: Standalone registry server entrypoint for Phase 1 direct uploads.
    - Path: cmd/docsctl/main.go
      Note: New docsctl CLI entrypoint for Phase 1 validation and publish commands.
    - Path: cmd/docsctl/validate.go
      Note: docsctl validate command uses shared SQLite publish validator.
    - Path: cmd/docsctl/validate_test.go
      Note: Command tests for text output
    - Path: pkg/help/loader/sources.go
      Note: SQLiteDirLoader already maps package/version directory layouts into package metadata.
    - Path: pkg/help/publish/auth.go
      Note: PublisherAuth interface and static token hash authorization implementation.
    - Path: pkg/help/publish/auth_test.go
      Note: Tests proving token/package scoping behavior.
    - Path: pkg/help/publish/catalog.go
      Note: Vault-shaped token records and reloadable publisher auth catalog.
    - Path: pkg/help/publish/catalog_test.go
      Note: Tests for revoked records
    - Path: pkg/help/publish/registry.go
      Note: HTTP registry skeleton for authorized SQLite uploads.
    - Path: pkg/help/publish/registry_test.go
      Note: Tests for health
    - Path: pkg/help/publish/sqlite_validator.go
      Note: Read-only SQLite help DB validator shared by docsctl and registry publishing.
    - Path: pkg/help/publish/sqlite_validator_test.go
      Note: Tests for valid DBs
    - Path: pkg/help/publish/validation.go
      Note: Package/version/path validation helpers for safe docs publishing.
    - Path: pkg/help/publish/validation_test.go
      Note: Tests for safe and unsafe package/version names.
    - Path: pkg/help/server/handlers.go
      Note: Public API exposes health
    - Path: pkg/help/server/serve.go
      Note: Serve command supports external JSON/SQLite/SQLite directory sources but loads them only at startup today.
    - Path: pkg/help/server/types.go
      Note: API response contracts for package summaries
    - Path: pkg/help/store/store.go
      Note: SQLite schema stores package_name/package_version and unique package-version-slug identity.
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T13:02:08.7165699-04:00
WhatFor: ""
WhenToUse: ""
---










# Design docs.yolo.scapegoat.dev multi-package Glazed help deployment

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
