# Changelog

## 2026-05-02

- Initial workspace created


## 2026-05-02

Created multi-package docs.yolo.scapegoat.dev deployment research ticket, wrote intern-facing design guide, and started investigation diary.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml — Current production Deployment pattern to copy for docs-yolo.
- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml — Current Traefik/cert-manager Ingress pattern to adapt for docs.yolo.scapegoat.dev.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/loader/sources.go — Existing SQLiteDirLoader package/version directory contract.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go — Existing public API routes and package/version filtering.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go — Current startup-loaded serve command and external source flags.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/store/store.go — Existing composite package/version/slug store identity.


## 2026-05-02

Validated ticket and uploaded design/diary bundle to reMarkable at /ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE.


## 2026-05-02

Added Phase 1 static Vault-token publishing addendum and split Phase 2/3 Vault OIDC auth into ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH.


## 2026-05-02

Uploaded updated Phase 1 Vault token addendum bundle to reMarkable.


## 2026-05-02

Prepared complete updated Phase 1 ticket bundle for reMarkable upload.


## 2026-05-02

Uploaded complete updated Phase 1 ticket bundle to reMarkable and verified remote listing.


## 2026-05-02

Expanded current ticket into detailed multi-phase Phase 1 implementation tasks.


## 2026-05-02

Step 6: added docsctl CLI root entrypoint with help and version wiring.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/cmd/docsctl/main.go — New docs publishing CLI entrypoint.


## 2026-05-02

Step 7: added package/version validation helpers for safe docs publishing path construction.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/validation.go — Package/version/path validation helpers for docs publishing.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/validation_test.go — Validation helper tests.


## 2026-05-02

Step 8: added read-only SQLite help DB validator and unit tests for publish safety.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/sqlite_validator.go — Read-only Glazed help DB validator.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/sqlite_validator_test.go — Validator tests for valid and invalid uploads.


## 2026-05-02

Step 9: wired docsctl validate with JSON output, command tests, and real DB smoke validation.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/cmd/docsctl/validate.go — docsctl validate command implementation.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/cmd/docsctl/validate_test.go — Command-level validate tests.


## 2026-05-02

Step 10: added PublisherAuth interface, static token hash auth, and package-scope authorization tests.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/auth.go — Static package-token authorization core.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/auth_test.go — Auth tests for package scoping and token hashing.

