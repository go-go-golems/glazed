# Tasks

## Done

- [x] Create the Glazed ticket workspace under `glazed/ttmp`.
- [x] Import `glazed-vault-bootstrap-example.md`, `glazed-secret-redaction.patch`, and `glazed-implemented-clean.patch` into the ticket.
- [x] Audit the current Glazed `TypeSecret`, parse log, serialization, and parsed-field-printing code paths.
- [x] Audit the migrated `vault-envrc-generator` Vault middleware, section definition, and example usage.
- [x] Write the intern-oriented analysis/design/implementation guide.
- [x] Write the ticket diary for the investigation.

## Phase 1: Secret Redaction

- [x] Add shared sensitivity/redaction helpers in `pkg/cmds/fields`.
- [x] Redact secret values and parse logs in `ToSerializableFieldValue` and parsed-values JSON/YAML serialization.
- [x] Make `printParsedFields` reuse the redacted serializable representation instead of printing raw values.
- [x] Redact Cobra display defaults for secret fields so help output does not expose sensitive defaults.
- [x] Add focused tests covering secret redaction in serialization, parsed-field printing paths, and Cobra default display.
- [x] Update GL-009 diary, changelog, and task bookkeeping for the redaction phase.
- [x] Commit the Phase 1 redaction work to git.

## Phase 2: Vault Middleware And Bootstrap

- [x] Add a reusable Vault settings section and decoding helper in Glazed.
- [x] Mark sensitive Vault configuration fields such as `vault-token` as `TypeSecret`.
- [x] Add a minimal Vault client/helper layer in Glazed sufficient for reading KV secrets and resolving templated paths.
- [x] Add a Vault overlay middleware that hydrates only `TypeSecret` fields after `next(...)`.
- [x] Add bootstrap parsing support or a tested recipe for `vault-settings` so provider settings can come from config/env/flags before the main parse.
- [x] Add focused tests for Vault field hydration, section decoding, and precedence/override behavior.
- [x] Update GL-009 diary, changelog, and task bookkeeping for the Vault/bootstrap phase.
- [x] Commit the Phase 2 Vault/bootstrap work to git.

## Final Validation

- [x] Run the relevant focused tests for both phases.
- [x] Re-run `docmgr doctor --ticket GL-009-VAULT-SECRETS --stale-after 30`.
- [x] Refresh the ticket docs if the implementation diverged from the design.
