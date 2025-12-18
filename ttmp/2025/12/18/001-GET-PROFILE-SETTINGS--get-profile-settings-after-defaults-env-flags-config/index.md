---
Title: Get profile settings after defaults/env/flags/config
Ticket: 001-GET-PROFILE-SETTINGS
Status: complete
Topics:
    - glazed
    - cobra
    - profiles
    - config
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../geppetto/pkg/layers/layers.go
      Note: Reference implementation of middleware ordering and current profile-loading circularity note
    - Path: pkg/cli/cli.go
      Note: Defines ProfileSettingsSlug and related CLI settings structs
    - Path: pkg/cli/cobra-parser.go
      Note: Primary orchestration of Cobra parsing + middleware execution in Glazed
    - Path: pkg/cmds/middlewares/cobra.go
      Note: ParseFromCobraCommand implementation (flag+arg parsing)
    - Path: pkg/cmds/middlewares/middlewares.go
      Note: Middleware chain execution model and handler signatures
    - Path: pkg/cmds/middlewares/profiles.go
      Note: GatherFlagsFromProfiles implementation
    - Path: pkg/doc/topics/12-profiles-use-code.md
      Note: Docs describing intended profile usage and ordering
    - Path: pkg/doc/topics/21-cmds-middlewares.md
      Note: Docs describing middleware ordering/precedence and examples
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:37:19.128398489-05:00
---




# Get profile settings after defaults/env/flags/config

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- cobra
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
