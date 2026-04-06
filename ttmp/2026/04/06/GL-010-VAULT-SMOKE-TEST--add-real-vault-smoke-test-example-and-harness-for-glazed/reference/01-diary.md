---
Title: Diary
Ticket: GL-010-VAULT-SMOKE-TEST
Status: active
Topics:
    - glazed
    - vault
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Working diary for the real Vault smoke-test example and shell harness.
LastUpdated: 2026-04-06T16:10:00-04:00
WhatFor: Capture the implementation steps, decisions, validations, and follow-up notes for GL-010.
WhenToUse: Use this diary when reviewing what changed or rerunning the smoke harness work.
---

# Diary

## Goal

Capture the implementation of a real local Vault smoke-test harness for Glazed, including the command wiring, the smoke script workflow, the validation steps, and the reasoning behind any adjustments made during implementation.

## Context

GL-009 added the Vault middleware and bootstrap helper, but stopped short of a developer-facing end-to-end harness. This ticket adds that missing layer so maintainers can run a real command against a real local Vault server and verify precedence and redaction behavior without reconstructing the setup from source code alone.

## Quick Reference

- Ticket: `GL-010-VAULT-SMOKE-TEST`
- Target example directory: `cmd/examples/vault-smoke-test`
- Expected proof points:
  - secret fields hydrate from Vault
  - non-secret fields do not hydrate from Vault
  - env overrides Vault
  - flags override env
  - parsed-field output redacts secrets

## Usage Examples

### Step 1: Create the ticket and plan the smoke-test harness

User request:

> alright, add a ticket for the smoke test, and then add an implementation plan document and tasks, and then implement task by task, committing at approprivte intervals, adding A README.md and a test script in the examples directory, keep a diary

Actions:

- Created ticket `GL-010-VAULT-SMOKE-TEST`.
- Added an implementation plan document.
- Added this diary document.
- Audited the existing example patterns and the new Vault APIs before writing code.

Notes:

- The example should use the same public Vault helpers added in GL-009 instead of test-only seams.
- `printParsedFields` is package-private inside `pkg/cli`, so the example should rely on the normal Cobra command-settings path rather than trying to call that helper directly.
- `apply_patch` is failing in this workspace, so documentation and code edits are being done with a shell fallback.

Commit:

- Not committed yet at this stage.

## Related

- `../design-doc/01-implementation-plan-for-real-vault-smoke-test-harness.md`
- `../tasks.md`
