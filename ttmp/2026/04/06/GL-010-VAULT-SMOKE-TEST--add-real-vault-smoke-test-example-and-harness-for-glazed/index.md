---
Title: Add real Vault smoke-test example and harness for Glazed
Ticket: GL-010-VAULT-SMOKE-TEST
Status: active
Topics:
    - glazed
    - vault
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Add a real end-to-end example command and smoke-test harness for the Vault support added in GL-009.
LastUpdated: 2026-04-06T16:10:00-04:00
WhatFor: Provide a practical, reproducible local validation path for Vault-backed secret hydration and redaction behavior in Glazed.
WhenToUse: Use this ticket when implementing, running, or reviewing the real Vault smoke-test example and shell harness.
---

# Add real Vault smoke-test example and harness for Glazed

## Overview

This ticket adds a real end-to-end smoke-test harness for the Vault support introduced in GL-009. The goal is not more abstract unit coverage, but a runnable example and script that prove the full command path against a local `vault server -dev` instance.

The deliverable is a small example command under `cmd/examples/` plus a `README.md` and a smoke-test script. Together they should show that only `TypeSecret` fields hydrate from Vault, that later sources still override Vault as intended, and that parsed-field output stays redacted.

## Key Links

- [Implementation plan](./design-doc/01-implementation-plan-for-real-vault-smoke-test-harness.md)
- [Diary](./reference/01-diary.md)
- [Tasks](./tasks.md)
- **Related Files**: See frontmatter `RelatedFiles`
- **External Sources**: See frontmatter `ExternalSources`

## Status

Current status: **active**

Work in progress:

- Ticket scaffold created.
- Concrete implementation plan and tasks are in place.
- Next implementation phase is a dedicated example command in `cmd/examples/vault-smoke-test`.
- After the example is in place, a smoke-test script will start local Vault in `tmux`, seed secrets, run the example, and assert precedence and redaction behavior.

## Topics

- glazed
- vault

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design/` - Architecture and design documents
- `reference/` - Prompt packs, API contracts, context summaries
- `playbooks/` - Command sequences and test procedures
- `scripts/` - Temporary code and tooling
- `various/` - Working notes and research
- `archive/` - Deprecated or reference-only artifacts
