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


## 2026-05-02

Step 11: added Vault-shaped publisher token records and reloadable static auth catalog.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/catalog.go — Publisher token record and reloadable catalog implementation.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/catalog_test.go — Catalog validation and reload behavior tests.


## 2026-05-02

Step 12: added docs-registry skeleton with health/list/upload routes, auth, validation, and handler tests.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/cmd/docs-registry/main.go — New docs registry command skeleton.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/registry.go — Registry HTTP handler and upload route.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/registry_test.go — Registry route tests.


## 2026-05-02

Step 13: added DirectoryPackageStore, docsctl publish, file-backed publisher catalog, and registry wiring.

### Related Files

- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/cmd/docsctl/publish.go — docsctl publish client command.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/catalog_file.go — File-backed publisher catalog source for Phase 1 Vault-shaped token hashes.
- /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/publish/directory_store.go — Atomic directory publisher for package/version SQLite DBs.


## 2026-05-02

Step 14: added docs-yolo GitOps scaffold in k3s repo (commit fb821557652c9a5225499976d19d5a4ac3658fc0).


## 2026-05-02

Step 15: completed local end-to-end Phase 1 smoke validation for docsctl publish, docs-registry, package root materialization, and glaze serve.


## 2026-05-02

Step 16: uploaded Phase 1 implementation complete bundle to reMarkable and completed final ticket validation tasks.


## 2026-05-02

Step 17: added glaze serve --reload-interval and enabled 30s package directory reload polling in the docs-yolo GitOps deployment.


## 2026-05-02

Step 18: smoke-tested reload polling by publishing Pinocchio into an initially empty package root while the browser was already running.


## 2026-05-02

Uploaded Phase 1 reload update bundle to reMarkable after adding and smoke-testing help source reload polling.


## 2026-05-02

Step 20: pushed Glazed branch, built GHCR image sha-cca4dbf, synced docs-yolo in Argo CD, bootstrapped Glazed/Pinocchio vtest DBs, and opened PR #561.


## 2026-05-02

Step 21: addressed PR review comments by staging reloads before clearing live docs and generating unique repeated heading IDs across backend/frontend.


## 2026-05-02

Step 22: refactored docsctl validate/publish and docs-registry from raw Cobra flag definitions into Glazed command definitions and settings decoding.


## 2026-05-02

Step 23: aligned build-web with generated-at-build Dagger/pnpm strategy, removed committed pkg/web/dist assets, and switched production embedding to -tags embed.


## 2026-05-02

Step 24: honored docsctl Glazed --print-schema/--print-yaml/--print-parsed-fields flags before command execution and added no-upload regression tests.


## 2026-05-02

Step 25: moved docsctl --print-* handling into exported Glazed CLI helpers and reused them from both docsctl and the default Cobra runner.

