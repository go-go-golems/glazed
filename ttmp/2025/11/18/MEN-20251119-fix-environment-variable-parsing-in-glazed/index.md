---
Title: Fix environment variable parsing in glazed
Ticket: MEN-20251119
Status: complete
Topics:
    - glazed
    - parameters
    - middleware
    - env
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/middlewares/update.go
      Note: Env update middleware to parse typed values
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/middlewares/update_test.go
      Note: New test file for env parsing behavior
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/parameters/gather-parameters.go
      Note: Map inputs string->typed conversion example
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/parameters/parameters.go
      Note: Type checking and SetValue helpers
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/parameters/parse.go
      Note: String parsing/typing entry points for various ParameterTypes
    - Path: /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/parameters/strings.go
      Note: Flag parsing from string lists (reference for parsing semantics)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:37:18.894741716-05:00
---




# Fix environment variable parsing in glazed

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

## Topics

- glazed
- parameters
- middleware
- env

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
