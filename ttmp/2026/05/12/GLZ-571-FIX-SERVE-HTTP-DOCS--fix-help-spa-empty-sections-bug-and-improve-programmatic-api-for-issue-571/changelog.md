# Changelog

## 2026-05-12

- Initial workspace created


## 2026-05-12

Created comprehensive analysis and design doc for issue #571. Root cause identified: normalization asymmetry between /api/packages and /api/sections combined with missing SetDefaultPackage in programmatic API path. Recommended fix: auto-assign default package in NewServeHandler + documentation updates.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/server/handlers.go — handleListPackages normalizes empty to default
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/server/serve.go — NewServeHandler needs SetDefaultPackage auto-assign


## 2026-05-12

Discovered second systemic issue: SPA assets cannot be embedded by external Go binaries that depend on glazed as a library. The go:generate + Dagger pipeline only runs in the glazed repo, and generated assets are excluded from the module cache by .gitignore. External consumers get a placeholder page. Created design doc 02 with analysis and solutions.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/cmd/build-web/main.go — Dagger pipeline not transitive across module boundaries
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/web/embed.go — SPA embedding only works within glazed repo


## 2026-05-12

Created comprehensive SPA distribution architecture patterns reference (ref/02). Covers 5 patterns: separate module, release artifact fetch, runtime fetch, npm package, consumer-side build. Includes decision matrix, decision framework, and step-by-step implementation guide for Pattern A. Uploaded to reMarkable.


## 2026-05-12

Implemented fix: (1) NewServeHandler auto-assigns 'default' package to sections with empty package_name. (2) Updated godoc on HandlerDeps, SetDefaultPackage, NewServeHandler. (3) Rewrote help entry 25-serving-help-over-http.md with API-only mode as default, SPA limitation documented, new troubleshooting rows. (4) Added 3 regression tests: auto-assign, API-only, no-overwrite. All 27 server tests pass, all 10 help package test suites pass.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/doc/topics/25-serving-help-over-http.md — Rewrote with API-only default
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/server/handlers.go — Improved HandlerDeps godoc
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/server/serve.go — Added SetDefaultPackage auto-assign in NewServeHandler
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/server/serve_test.go — Added 3 regression tests for issue 571
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/pkg/help/store/store.go — Improved SetDefaultPackage godoc

