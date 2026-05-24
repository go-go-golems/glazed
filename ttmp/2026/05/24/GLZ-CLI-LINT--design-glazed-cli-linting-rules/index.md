---
Title: Design Glazed CLI linting rules
Ticket: GLZ-CLI-LINT
Status: active
Topics:
    - glazed
    - linting
    - cli
    - cobra
    - intern-onboarding
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/design-doc/01-glazed-cli-linting-rules-analysis-and-implementation-guide.md
      Note: Primary design and implementation guide
    - Path: glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/reference/01-investigation-diary.md
      Note: Chronological diary for this investigation
ExternalSources:
    - https://pkg.go.dev/golang.org/x/tools/go/analysis
    - https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest
Summary: Ticket for designing a Glazed-specific go/analysis linter for raw env reads, misplaced Glazed output flags, and raw CLI flag definitions.
LastUpdated: 2026-05-24T12:35:00-04:00
WhatFor: Track the linting design package and follow-up implementation tasks.
WhenToUse: When planning or implementing the Glazed CLI policy analyzer.
---


# Design Glazed CLI linting rules

## Overview

This ticket defines a new Glazed-specific linting tool that should be implemented as a `go/analysis` vettool. The requested policy checks are:

- direct `os.Getenv` usage in CLI/application code;
- Glazed output sections added to commands that do not output structured rows through the Glazed framework;
- raw Cobra, pflag, or standard-library `flag` APIs used to define CLI verb flags instead of Glazed `fields.New` / `cmds.WithFlags` schemas.

The primary deliverable is an intern-oriented implementation guide, not a code implementation.

## Key Links

- **Primary design**: [design-doc/01-glazed-cli-linting-rules-analysis-and-implementation-guide.md](design-doc/01-glazed-cli-linting-rules-analysis-and-implementation-guide.md)
- **Investigation diary**: [reference/01-investigation-diary.md](reference/01-investigation-diary.md)
- **Tasks**: [tasks.md](tasks.md)
- **Changelog**: [changelog.md](changelog.md)

## Status

Current status: **active**.

The design package is complete and ready for implementation review. Follow-up tasks remain for implementing the analyzer and wiring it into Glazed/Pinocchio lint workflows.

## Topics

- glazed
- linting
- cli
- cobra
- intern-onboarding

## Structure

- `design-doc/` - Architecture and implementation guide.
- `reference/` - Investigation diary.
- `playbooks/` - Reserved for future command sequences.
- `scripts/` - Reserved for future experiments.
- `various/` - Reserved for working notes.
- `archive/` - Reserved for deprecated/reference-only artifacts.
