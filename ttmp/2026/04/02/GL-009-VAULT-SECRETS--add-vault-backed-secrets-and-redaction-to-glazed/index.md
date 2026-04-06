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

This ticket now captures both the design work and the completed implementation for bringing Vault-backed secret loading and consistent sensitive-value handling into Glazed. The final implementation stayed intentionally smaller than the imported generalized patch: keep `TypeSecret` as the only first-pass sensitivity semantic, centralize output redaction in `pkg/cmds/fields`, and port a minimal Vault overlay middleware plus bootstrap helper based on the existing profile-style precedence model.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: `design-doc/01-intern-guide-vault-backed-secrets-credentials-aliases-and-redaction-in-glazed.md`
- **Technical Project Report**: `design-doc/02-technical-project-report-glazed-secret-redaction-and-vault-bootstrap.md`
- **Diary**: `reference/01-diary.md`

## Status

Current status: **active**

Completed in this ticket so far:

1. Created the ticket workspace.
2. Imported the provided sketch and patch files.
3. Audited the current Glazed code and the migrated `vault-envrc-generator` implementation.
4. Wrote the intern-oriented analysis/design/implementation guide.
5. Implemented redaction hardening.
6. Implemented the Vault settings section, middleware, and bootstrap helper.
7. Added focused tests and a technical project report.

Still open:

1. Optional downstream example/help-page adoption work, outside the core framework merge.

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
