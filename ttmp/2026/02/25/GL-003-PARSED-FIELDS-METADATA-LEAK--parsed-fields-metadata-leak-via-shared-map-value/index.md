---
Title: Parsed-Fields Metadata Leak Via Shared map-value
Ticket: GL-003-PARSED-FIELDS-METADATA-LEAK
Status: active
Topics:
    - glazed
    - security
    - metadata
    - config
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/cli/helpers.go
      Note: Exposes parse metadata in --print-parsed-fields
    - Path: glazed/pkg/cmds/fields/field-value.go
      Note: Root-cause metadata aliasing
    - Path: glazed/pkg/cmds/fields/gather-fields.go
      Note: Root-cause raw map-value metadata capture
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-25T20:03:32.2159977-05:00
WhatFor: ""
WhenToUse: ""
---


# Parsed-Fields Metadata Leak Via Shared map-value

## Overview

This ticket documents a security/correctness bug where `--print-parsed-fields` can expose raw secret values via `metadata.map-value` and can smear those values across unrelated fields due to metadata map aliasing.

Deliverables in this workspace:

1. Design-doc bug report with root cause and fix proposal.
2. Investigation diary with reproduction commands and sanitized evidence.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- security
- metadata
- config

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
