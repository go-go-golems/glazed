---
Title: Implementation plan for real Vault smoke-test harness
Ticket: GL-010-VAULT-SMOKE-TEST
Status: active
Topics:
    - glazed
    - vault
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Add a dedicated example command and a real local Vault smoke script to validate the GL-009 Vault middleware end to end.
LastUpdated: 2026-04-06T16:10:00-04:00
WhatFor: Explain the shape, rationale, and staged implementation plan for the real Vault smoke-test harness.
WhenToUse: Use this document when implementing or reviewing the example command and smoke script.
---

# Implementation plan for real Vault smoke-test harness

## Executive Summary

GL-009 added the core Vault middleware and bootstrap parsing support, but it still lacks a practical end-to-end proof that a real Glazed command can consume those features against a real Vault server. This ticket closes that gap by adding a self-contained example command and a smoke-test script that exercises the real public APIs with a local `vault server -dev` process.

The example should make the important behaviors visible and reproducible: only `TypeSecret` fields should hydrate from Vault, later sources should still override Vault, and parsed and debug output should remain redacted.

## Problem Statement

The unit tests in `pkg/cmds/sources/vault_test.go` are useful, but they do not prove the entire CLI path. They do not show how a real command wires `vault-settings` into Cobra, how bootstrap parsing interacts with environment variables and config files, or how an operator should run and inspect the behavior locally.

Without a real smoke harness, regressions can hide in the seams between:

- command definition
- Cobra wiring
- bootstrap parsing of `vault-settings`
- main source-chain precedence
- secret redaction in parsed-field output

This ticket adds that missing integration layer.

## Proposed Solution

Add a dedicated example command under `cmd/examples/vault-smoke-test/` with three companion assets:

- `main.go`
- `README.md`
- `smoke-test.sh`

The example command will define a minimal app section with:

- one non-secret field such as `host`
- two secret fields such as `password` and `api-key`

It will then:

1. add `vault-settings` and `command-settings`
2. bootstrap `vault-settings` using the existing public helper
3. build the main source chain with the intended precedence
4. print the resolved values in a grep-friendly way

The smoke-test script will start a real local Vault dev server in `tmux`, seed secrets, create a config file, run the example multiple ways, and assert:

- secret fields hydrate from Vault
- non-secret fields do not
- env overrides Vault
- flags override env
- parsed-field output redacts secrets

## Design Decisions

### Use a dedicated example command instead of only more tests

A runnable example has two jobs that unit tests do not:

- it documents the intended integration pattern for future commands
- it gives maintainers a fast manual verification path when debugging source precedence issues

### Use the real GL-009 public APIs

The example should exercise the real shape added in GL-009, not internal test seams:

- `sources.NewVaultSettingsSection()`
- `sources.BootstrapVaultSettings(...)`
- `sources.FromVaultSettings(...)`

That keeps the example aligned with the actual supported integration surface.

### Put process orchestration in the shell script, not in the Go example

The Go example should remain focused on command wiring and source precedence. Starting and killing Vault belongs in the smoke harness. That keeps `main.go` easy to read and keeps the shell script responsible for external process lifecycle.

### Require a local Vault binary for the real smoke test

This ticket is explicitly for a real smoke path. If `vault` is not installed, the smoke script should fail clearly rather than silently falling back to mocks.

### Make output grep-friendly

The script should be able to assert behavior with plain shell tools. The example therefore needs stable, plain-text output lines such as:

- `host=...`
- `password=...`
- `api-key=...`

## Alternatives Considered

### Only add more Go tests

Rejected because it still would not give a copy-paste operator workflow or a documented integration example in `cmd/examples/`.

### Only add a shell script without a dedicated example

Rejected because the whole point is to demonstrate how a real command should integrate the Vault middleware. Reusing an unrelated existing example would make the smoke path harder to understand and maintain.

### Start Vault from Go integration tests

Deferred. That might be worthwhile later, but for this ticket the lower-friction and more inspectable path is a shell smoke script that developers can run directly.

## Implementation Plan

1. Replace the ticket stubs with a concrete plan, tasks, and diary baseline.
2. Add `cmd/examples/vault-smoke-test/main.go`.
3. Add `cmd/examples/vault-smoke-test/README.md`.
4. Run focused Go validation for the example package.
5. Commit the example harness baseline.
6. Add `cmd/examples/vault-smoke-test/smoke-test.sh`.
7. Run the smoke-test script against the local Vault binary and fix any issues.
8. Commit the smoke harness.
9. Update the diary, changelog, and ticket metadata.
10. Run `docmgr doctor` and commit the final bookkeeping.

## Open Questions

### Should the smoke script also assert JSON or YAML output?

The first pass can focus on the plain-text resolved-value output plus `--print-parsed-fields`. If JSON or YAML becomes part of the smoke surface, the script can grow later.

### Should this eventually run in CI?

Possibly, but this ticket is scoped to a developer-facing smoke harness first.

### How configurable should the secret path be?

The example should be simple by default, but the command and script should expose enough knobs to make local debugging easy.

## References

- `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault.go`
- `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_settings.go`
- `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_test.go`
- `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/cmd/examples/appconfig-profiles/main.go`
- `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/cmd/examples/sources-example/main.go`
