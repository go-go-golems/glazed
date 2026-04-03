---
Title: Add vault-backed secrets and redaction to Glazed
Ticket: GL-009-VAULT-SECRETS
Status: active
Topics:
    - glazed
    - security
    - config
    - vault
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - local:glazed-implemented-clean.patch
    - local:glazed-secret-redaction.patch
    - local:glazed-vault-bootstrap-example.md
Summary: ""
LastUpdated: 2026-04-02T19:20:42.706868306-04:00
WhatFor: ""
WhenToUse: ""
---




# Add vault-backed secrets and redaction to Glazed

## Overview

This ticket captures the design work for bringing Vault-backed secret loading and consistent sensitive-value handling into Glazed. The current recommendation is intentionally smaller than the imported generalized patch: keep `TypeSecret` as the first-pass behavior, accept `credentials` only as an alias, centralize output redaction in `pkg/cmds/fields`, and port a minimal Vault overlay middleware plus bootstrap recipe based on the existing `appconfig.WithProfile(...)` pattern.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: `design-doc/01-intern-guide-vault-backed-secrets-credentials-aliases-and-redaction-in-glazed.md`
- **Diary**: `reference/01-diary.md`

## Status

Current status: **active**

Completed in this ticket so far:

1. Created the ticket workspace.
2. Imported the provided sketch and patch files.
3. Audited the current Glazed code and the migrated `vault-envrc-generator` implementation.
4. Wrote the intern-oriented analysis/design/implementation guide.

Still open:

1. Implement redaction hardening.
2. Add the `credentials` alias.
3. Add the Vault section and middleware.
4. Add tests and documentation examples.

## Topics

- glazed
- security
- config
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
