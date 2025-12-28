---
Title: Honor default profile name in missing-file semantics
Ticket: 004-PROFILE-DEFAULT-NAME
Status: active
Topics:
    - profiles
    - config
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: .ttmp.yaml
      Note: docmgr root path update
    - Path: pkg/appconfig/options.go
      Note: WithProfile wiring for default profile name
    - Path: pkg/cmds/middlewares/profiles.go
      Note: default profile missing-file logic
    - Path: pkg/doc/topics/12-profiles-use-code.md
      Note: profile usage examples
    - Path: pkg/doc/topics/15-profiles.md
      Note: profile error semantics
    - Path: pkg/doc/topics/21-cmds-middlewares.md
      Note: middleware signature documentation
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T18:23:00.267979649Z
---


# Honor default profile name in missing-file semantics

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- profiles
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
