---
Title: Minimal Structured Output and Machine-Readable Command Manifests
Ticket: GLAZED-DESCRIBE-MANIFESTS
Status: active
Topics:
    - glazed
    - commands
    - cli
    - cobra
    - settings
    - formatters
    - api-design
    - migration
    - help-system
    - intern-guide
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cmds/cmds.go
      Note: Command authoring model extended by the proposal.
    - Path: pkg/cli/cobra.go
      Note: Current automatic output injection and registration path.
    - Path: pkg/settings/glazed_section.go
      Note: Current rich structured-output settings aggregate.
    - Path: pkg/settings/flags
      Note: Current 44-flag automatic structured-data surface.
ExternalSources: []
Summary: Design ticket for reducing Glazed commands to one output-format flag and adding versioned static command manifests exposed through a standard describe command.
LastUpdated: 2026-07-09T17:30:00-04:00
WhatFor: Track the design, evidence, implementation plan, and GitHub handoff for the framework change.
WhenToUse: Start here when implementing or reviewing minimal structured output and machine-readable command discovery.
---

# Minimal Structured Output and Machine-Readable Command Manifests

## Overview

This ticket defines a Glazed framework hard cut: stop automatically mounting the 44-flag rich structured-data transformation surface on every `GlazeCommand`, retain one stable `--format table|json|jsonl|csv|tsv` boundary, and add a static versioned command catalog exposed by `app describe [command path...]`.

The design also introduces atomic command compilation so duplicate paths, aliases, and application/framework input collisions fail before the Cobra tree is mutated. Rich processors and formatters remain available as libraries or explicit transformation tools.

## Key documents

- [Design: Minimal Structured Output and Machine-Readable Command Manifests](./design-doc/01-design-minimal-structured-output-and-machine-readable-command-manifests.md)
- [Implementation Diary](./diary/01-implementation-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Status

The design document is drafted and awaiting repository validation and GitHub issue publication.

## Central decisions

- Keep only `--format` as the automatic structured-output flag.
- Delegate post-processing and destination routing to callers or explicit tools.
- Keep source-affecting business flags in individual commands.
- Compile a static immutable command catalog before Cobra mounting.
- Implement root-level `describe`; do not make leaf `--describe` the primary API.
- Treat undeclared effects as unknown.
- Make the migration a major-version hard cut with no compatibility aliases.
