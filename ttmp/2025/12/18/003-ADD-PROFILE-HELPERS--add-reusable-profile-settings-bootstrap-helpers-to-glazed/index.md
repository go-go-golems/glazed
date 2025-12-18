---
Title: Add reusable profile-settings bootstrap helpers to Glazed
Ticket: 003-ADD-PROFILE-HELPERS
Status: active
Topics:
    - glazed
    - profiles
    - cobra
    - middleware
    - refactor
    - docs
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../geppetto/pkg/doc/topics/01-profiles.md
      Note: Downstream docs that should reference helper/pattern
    - Path: ../../../../../../geppetto/pkg/layers/layers.go
      Note: Current bootstrap parse implementation to refactor into Glazed helpers
    - Path: pkg/cli/cobra-parser.go
      Note: CobraParser.Parse + ParseCommandSettingsLayer bootstrap limitations
    - Path: pkg/cmds/middlewares/cobra.go
      Note: LoadParametersFromResolvedFilesForCobra helper (potential building block)
    - Path: pkg/cmds/middlewares/profiles.go
      Note: GatherFlagsFromProfiles behavior and error semantics
    - Path: pkg/doc/topics/15-profiles.md
      Note: New canonical docs page describing bootstrap/circularity
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:17:08.274426142-05:00
---


# Add reusable profile-settings bootstrap helpers to Glazed

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- profiles
- cobra
- middleware
- refactor
- docs

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
