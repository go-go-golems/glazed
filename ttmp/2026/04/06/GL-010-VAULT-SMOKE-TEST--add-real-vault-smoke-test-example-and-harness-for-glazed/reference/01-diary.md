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
RelatedFiles:
    - Path: cmd/examples/vault-smoke-test/README.md
      Note: Operator guide for running the example manually
    - Path: cmd/examples/vault-smoke-test/main.go
      Note: Example command implementing the real Vault smoke harness
    - Path: cmd/examples/vault-smoke-test/smoke-test.sh
      Note: Automated real Vault smoke-test script using tmux
ExternalSources: []
Summary: Working diary for the real Vault smoke-test example and shell harness.
LastUpdated: 2026-04-06T16:35:00-04:00
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

- `0d48330` `Add GL-010 smoke-test ticket plan and diary`

### Step 2: Implement the example command and README

Actions:

- Added `cmd/examples/vault-smoke-test/main.go`.
- Added `cmd/examples/vault-smoke-test/README.md`.
- Built the example through `cli.BuildCobraCommandFromCommand` so `--print-parsed-fields` uses the standard Glazed path.
- Added a custom middleware builder that bootstraps `vault-settings` first and then runs the main chain as `defaults -> config -> vault -> env -> args -> cobra`.
- Made the example print the resolved values and their winning source in a shell-friendly `key=value` format.

Validation:

- `go test ./cmd/examples/vault-smoke-test`
- `go run ./cmd/examples/vault-smoke-test --help`
- `go run ./cmd/examples/vault-smoke-test`
- `go run ./cmd/examples/vault-smoke-test --print-parsed-fields`

Observations:

- The plain run shows the default values and their source as `defaults`.
- The parsed-field output is redacted for `TypeSecret` values, including the default `vault-token` field.
- The help output already masks secret defaults as expected.

Commit:

- `ce85c11` `Add GL-010 Vault smoke-test example harness`

### Step 3: Implement and run the real smoke script

Actions:

- Added `cmd/examples/vault-smoke-test/smoke-test.sh`.
- Made the script start a dedicated `vault server -dev` instance inside a unique `tmux` session.
- Wrote a temporary config file that sets `vault-settings.secret-path` and a config-level `app.host` / `app.password`.
- Seeded Vault with `password`, `api-key`, and `host` so the smoke run could prove that only the `TypeSecret` fields hydrate from Vault.
- Added assertion helpers for shell output and a cleanup trap that captures the `tmux` pane on failure.

Validation:

- `./cmd/examples/vault-smoke-test/smoke-test.sh`

Observed cases:

- Case 1: config provided `host`, Vault replaced only `password` and `api-key`
- Case 2: `GLAZED_VAULT_SMOKE_TEST_PASSWORD` overrode the Vault value
- Case 3: `--password` overrode the environment value
- Case 4: `GLAZED_VAULT_SMOKE_TEST_SECRET_PATH` allowed bootstrap without a config file
- Case 5: `--print-parsed-fields` did not print the raw Vault secrets or the root token

Commit:

- Not committed yet at this stage.

### Step 4: Final bookkeeping

Actions:

- Updated the task list and changelog for the completed smoke harness work.
- Planned doc relations for the example files so the ticket metadata points back to the implementation.
- Planned a final `docmgr doctor --ticket GL-010-VAULT-SMOKE-TEST --stale-after 30` run before the closing commit.

Commit:

- Not committed yet at this stage.

## Related

- `../design-doc/01-implementation-plan-for-real-vault-smoke-test-harness.md`
- `../tasks.md`
