# Changelog

## 2026-04-07

- Initial workspace created


## 2026-04-07

Step 1: Create ticket, scaffold docs, import JSX prototype, write 993-line design doc with 16 sections + RTK Query integration section (compact replacement), create diary with step 1

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/design-doc/01-help-browser-architecture-and-implementation-guide.md — Primary design doc — 993 lines
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/reference/01-diary.md — Diary with Step 1 record
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/sources/local/glazed-docs-browser(2).jsx — Imported JSX prototype


## 2026-04-07

Replace placeholder tasks.md with 64 granular tasks across 8 phases: Phase 1 Go server (7 tasks), Phase 2 React scaffold (8), Phase 3 component decomposition (13), Phase 4 theming (6), Phase 5 Storybook (14), Phase 6 Dagger (6), Phase 7 Cobra (4), Phase 8 testing (6)

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/tasks.md — 64 granular tasks across 8 phases


## 2026-04-07

Phase 1 Task 1: Create pkg/help/server/types.go with SectionSummary, SectionDetail, ListSectionsParams, HealthResponse, ErrorResponse, SummaryFromModel, DetailFromModel (commit fb2f616)

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/types.go — Phase 1 Task 1 — HTTP request/response types


## 2026-04-07

Phase 1 Tasks 2-6: pkg/help/server/ — handlers, middleware, SPA fallback, Server, tests (commit d6ac109)

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/handlers.go — Tasks 2-6 — HTTP handler functions
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/middleware.go — Tasks 2-6 — CORS middleware
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/server.go — Tasks 2-6 — Server struct
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/server_test.go — Tasks 2-6 — 13 integration tests
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/spa.go — Tasks 2-6 — SPA fallback handler (middleware-style)
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/store.go — Task 6 — Added ErrSectionNotFound sentinel


## 2026-04-07

Phase 1 Task 7: add cmd/help-browser/main.go (commit 9d673f9)

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/help-browser/main.go — Task 7 — standalone CLI


## 2026-04-07

Phase 2 Tasks 8-15: scaffold web/ with Vite + React + RTK Query + TypeScript (commit cca9859)

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/web/ — Phase 2 Tasks 8-15 — full web scaffold


## 2026-04-07

Phase 3 Tasks 16-28: all 13 React components extracted from JSX prototype (commit eae3b82)


## 2026-04-07

Phase 4 + Phase 5 Tasks 29-36, 37-48: Storybook configured + stories for all components (commit eae3b82)


## 2026-04-07

cmd/build-web: Dagger Go SDK builder with local pnpm fallback. cmd/help-browser: added //go:generate + //go:embed dist. go generate ./cmd/help-browser now produces embedded web assets.


## 2026-04-08

Shared SPA refactor slice 1: move generated frontend ownership to `pkg/web/dist`, embed via `pkg/web/static.go`, restore robust markdown path loading in `pkg/help/server/serve.go`, wire both `cmd/help-browser` and `cmd/glaze` to the shared embedded assets, and add prefix-mount helpers/tests.

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/build-web/main.go — Builder now outputs to pkg/web/dist
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static.go — Shared embedded asset package
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve.go — Robust loader + root handler + prefix mounting helpers
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve_test.go — Tests for root SPA serving and prefix mounting
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/help-browser/main.go — Standalone binary wired to pkg/web
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/glaze/main.go — Main CLI wired to pkg/web

## 2026-04-08

Shared SPA refactor slice 2: move SPA-serving ownership into `pkg/web.NewSPAHandler()`, make `pkg/help/server` compose API + optional SPA handlers instead of raw embed FS values, and add a playbook showing how to mount the help browser under `/help` or another prefix.

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static.go — Shared embedded asset package now also exposes `NewSPAHandler()`
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve.go — Serve command now composes API + optional SPA handlers
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve_test.go — Tests updated to consume `web.NewSPAHandler()`
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/help-browser/main.go — Standalone binary now builds SPA handler from `pkg/web`
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/glaze/main.go — Main CLI now builds SPA handler from `pkg/web`
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/playbooks/01-mount-help-browser-under-prefix.md — Prefix-mount playbook for existing servers

## 2026-04-08

Cleanup slice: remove the redundant `cmd/help-browser` runtime wrapper, move `go generate` ownership to `pkg/web/gen.go`, add direct `pkg/web.NewSPAHandler()` tests, and delete the now-dead generic server/SPA wrapper files.

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/gen.go — `go generate` ownership moved to the shared web package
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static.go — Shared SPA handler and embed owner remain the single source of truth
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static_test.go — Direct SPA handler tests
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/build-web/main.go — Builder docs updated to match the new ownership model
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/glaze/main.go — `glaze serve` is now the only supported runtime entrypoint

## 2026-04-08

Final polish: ignore `web/node_modules/`, keep `web/pnpm-lock.yaml` in version control for reproducible frontend builds, and add a new Glazed help topic (`serve-help-over-http`) that explains `glaze serve`, the HTTP API, and how to mount the help browser/API in existing servers.

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/.gitignore — Ignore `web/node_modules/`
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/web/pnpm-lock.yaml — Frontend dependency lockfile kept in version control
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/doc/topics/25-serving-help-over-http.md — New help topic for `glaze serve` and programmatic API usage

## 2026-04-08 (Step 13)

UI fixes and serve behavior improvements.

### Changes

- Removed MenuBar from the help browser UI
- Added `isTopLevel` to API response types (Go + TypeScript) and fixed ◆ TOP indicator
- Made `glaze serve` paths optional — no arguments now serves the built-in embedded documentation
- Fixed initial section selection race condition (no auto-select on load)
- Updated CSS with classic Mac font stack and retro scrollbar styling
- Updated help topic to document no-args behavior

### Commit

- `4966ed1` — Fix UI issues and make serve paths optional
