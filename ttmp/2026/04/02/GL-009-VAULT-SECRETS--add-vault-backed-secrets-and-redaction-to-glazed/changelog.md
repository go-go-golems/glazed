# Changelog

## 2026-04-02

- Initial workspace created


## 2026-04-02

Added an evidence-backed intern guide for vault-backed secrets, credentials aliasing, centralized redaction, and vault bootstrap design in Glazed; imported the provided sketch and patch files and recorded the investigation diary.

### Related Files

- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/ttmp/2026/04/02/GL-009-VAULT-SECRETS--add-vault-backed-secrets-and-redaction-to-glazed/design-doc/01-intern-guide-vault-backed-secrets-credentials-aliases-and-redaction-in-glazed.md — Primary analysis/design/implementation guide
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/ttmp/2026/04/02/GL-009-VAULT-SECRETS--add-vault-backed-secrets-and-redaction-to-glazed/reference/01-diary.md — Chronological investigation record


## 2026-04-02

Implemented Phase 1 centralized secret redaction in Glazed core: TypeSecret values are now masked in serializable field output, parsed-field debug printing, parse-log metadata rendering, and Cobra help default displays; added focused regression tests. (commit c4445fa780898da9b3e4612409968ceac5e5e99a)

### Related Files

- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go — Parsed-field printing reuses the redacted serializable representation
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go — Secret defaults are masked in Cobra help text
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/sensitive.go — Shared redaction helpers introduced for TypeSecret handling
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go — Serializable output now redacts sensitive values and parse logs


## 2026-04-02

Implemented Phase 2 Vault support in pkg/cmds/sources: added a reusable vault-settings section, token/path helpers, Vault overlay middleware that only hydrates TypeSecret fields, and a bootstrap helper plus precedence tests. No credentials alias was added. (commit b18ccb6)

### Related Files

- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault.go — Vault client
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_settings.go — Reusable vault-settings schema and decoding helper
- /home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_test.go — Focused coverage for section decode

