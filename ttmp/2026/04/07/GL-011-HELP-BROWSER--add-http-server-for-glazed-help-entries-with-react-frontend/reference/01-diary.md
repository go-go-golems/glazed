---
Title: Diary
Ticket: GL-011-HELP-BROWSER
Status: active
Topics:
    - glazed
    - help
    - http
    - react
    - vite
    - rtk-query
    - storybook
    - dagger
    - web
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/help-browser/main.go
      Note: Step 4 — Task 7 committed (9d673f9)
    - Path: pkg/help/server/handlers.go
      Note: Step 3 — NewHandler
    - Path: pkg/help/server/middleware.go
      Note: Step 3 — NewCORSHandler (commit d6ac109)
    - Path: pkg/help/server/server.go
      Note: Step 3 — Server
    - Path: pkg/help/server/server_test.go
      Note: Step 3 — 13 tests (commit d6ac109)
    - Path: pkg/help/server/spa.go
      Note: Step 3 — SPAHandler middleware
    - Path: pkg/help/server/types.go
      Note: Phase 1 Task 1 — HTTP types (commit fb2f616)
    - Path: pkg/help/store/store.go
      Note: Step 3 — Added ErrSectionNotFound sentinel (commit d6ac109)
    - Path: ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/design-doc/01-help-browser-architecture-and-implementation-guide.md
    - Path: ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/sources/local/glazed-docs-browser(2).jsx
    - Path: web
      Note: Step 5 — Phase 2 scaffold committed (cca9859)
    - Path: web/.storybook
      Note: Step 6 — Storybook configured + 9 story files (eae3b82)
    - Path: web/src/components
      Note: Step 6 — 13 components extracted from JSX prototype
ExternalSources: []
Summary: Step-by-step diary of the glaze help browser implementation
LastUpdated: 2026-04-08T00:00:00Z
WhatFor: Record implementation progress and decisions
WhenToUse: Use when following up on or reviewing this ticket
---






# Diary

## Goal

Add an HTTP server and React frontend to the Glazed CLI tool so users can browse help
documentation in a web browser via `glaze serve file1 file2 dir1 dir2...`. The result
is a single Go binary that discovers Glazed help Markdown files, loads them into an
SQLite-backed store, and serves them over HTTP with a React SPA frontend — built via
Dagger, embedded with `go:embed`.

## Step 1: Ticket creation, document scaffolding, and design writing

**User prompt (verbatim):** "Create a new docmgr ticket to add a http server for glazed
help entries. Keep a diary regularly. The goal is to have glaze serve file1 file2
dir1 dir2... etc... discover glazed help files and then serve them using a react
frontend and go backend. Import the file ~/Downloads/glazed-docs-browser(2).jsx into
the ticket using docmgr import file --file XXX --ticket YYY, and create a detailed
analysis to split the react into a rtk-query/react/modular/themable/vite/storybook
based design in the web/ directory (see skill), and bundle it using a go tool /
go generate dagger based pipeline (see ~/code/wesen/corporate-headquarters/remarquee/
pkg/doc/topics/how-to-create-a-web-app-with-react-rtk-vite-dagger-gen.md). Create a
detailed analysis / design / implementation guide that is very detailed for a new
intern, explaining all the parts of the system needed to understand what it is, with
prose paragraphs and bullet points and pseudocode and diagrams and api references and
file references. It should be very clear and detailed. Store in the ticket and the
nupload to remarkable."

**Assistant interpretation:** Create a new ticket GL-011-HELP-BROWSER with an exhaustive
design doc, import the JSX prototype, write an initial diary, and upload to reMarkable.

**Inferred user intent:** Set up the full planning scaffold for the help browser feature
so implementation can begin immediately from the ticket.

### What I did

1. Read all relevant skill files (docmgr, diary, ticket-research-docmgr-remarkable,
   react-modular-themable-storybook, remarkable-upload).
2. Read the JSX prototype (`glazed-docs-browser(2).jsx`) and the Dagger guide.
3. Explored the existing codebase: `pkg/help/help.go`, `pkg/help/model/section.go`,
   `pkg/help/store/store.go`, `pkg/help/store/query.go`, `pkg/help/cmd/cobra.go`,
   `cmd/glaze/main.go`.
4. Created ticket GL-011-HELP-BROWSER via `docmgr ticket create-ticket`.
5. Added three docs: design-doc, reference (diary), and one more.
6. Imported the JSX prototype via `docmgr import file --file ... --ticket GL-011-HELP-BROWSER`.
7. Added 10 tasks covering all phases.
8. Related 7 key source files to the ticket.
9. Wrote the comprehensive design document (~993 lines) covering 16 major sections:
   executive summary, problem statement, glossary, system overview, existing
   architecture, proposed architecture, Go backend design, REST API specification,
   React frontend design, component decomposition, RTK Query integration, theming
   system, build pipeline, Storybook integration, file layout, implementation plan,
   testing strategy, risks/open questions, and references.
10. Split the RTK Query integration section into two (one for the long original version
    and one for the replacement compact version) then resolved duplicate section headers
    by editing the replacement section.

### Why

The user wanted exhaustive documentation with all the context a new intern would need.
This required reading all existing architecture, the JSX prototype, and the reference
guide before writing a single word. The design doc needed to be written in small chunks
to avoid timeouts, with each chunk appended cleanly.

### What worked

- Ticket structure created cleanly with all 4 default files (index, README, tasks, changelog).
- `docmgr import file` worked on the first try.
- `docmgr task add` created all 10 tasks in one batch.
- File-backed architecture analysis was efficient using `head -`, `grep -n`, and targeted `cat`.
- The Table of Contents was updated once at the end to reflect the final section order.

### What didn't work

- Initial attempt to replace a section with duplicate heading ("## Theming System") failed
  because the doc had two copies of that section. The second attempt to replace
  "## React Frontend Design" was more specific and succeeded.

### What I learned

- When writing long documents in chunks, append them with `cat >> file.md` (bash heredoc)
  to avoid edit tool failures on large files.
- `docmgr task add` can take multiple tasks in one call.
- `docmgr doc relate` accepts multiple `--file-note` entries in one call.
- The existing architecture uses `SectionType` as a string in JSON (via `model.SectionType.String()`),
  not the integer enum — this matters for the REST API.
- The `HelpSystem` already provides `LoadSectionsFromFS(f fs.FS, dir string)` which
  handles recursive directory walking and frontmatter parsing — we just need to expose it
  via HTTP and add CLI argument-driven file discovery.

### What was tricky to build

- Understanding the relationship between `help.Section` (wrapper) and `model.Section`
  (data model). The wrapper adds a back-reference to `HelpSystem` and convenience methods.
- The `go:embed dist` + Dagger pipeline needed careful path alignment:
  `cmd/build-web/main.go` (in `cmd/build-web/`) → `repoRoot = filepath.Dir(filepath.Dir(wd))`
  → `web/` is two levels up → output to `cmd/help-browser/dist/`.
- The RTK Query section ended up duplicated — the original append-based section and the
  edit-based compact replacement both appeared. Resolved by ensuring the edit target
  was unique.

### What warrants a second pair of eyes

- Whether the `HelpSystem.LoadSectionsFromFS()` already handles all the file discovery
  we need, or whether we need a separate argument-driven loader that falls back to embedded
  docs when no arguments are given.
- Whether the two `## Theming System` sections in the document merged cleanly — the
  content looks correct but a final full review would confirm.
- Whether the decision to use `HashRouter` (recommended for v1) is acceptable or if
  `BrowserRouter` + SPA fallback is preferred.

### What should be done in the future

- Run `docmgr doctor GL-011-HELP-BROWSER` to validate frontmatter and vocabulary before upload.
- Upload the design doc + diary to reMarkable via `remarquee upload bundle`.
- Begin Phase 1 implementation (Go HTTP server scaffold).

### Code review instructions

Where to start: Read the design doc from top to bottom. Key files to reference:
- `pkg/help/help.go` lines 339+ (HelpSystem struct)
- `pkg/help/model/section.go` (Section struct with JSON tags)
- `pkg/help/store/store.go` lines 22+ (New, List, GetBySlug)
- `ttmp/.../sources/local/glazed-docs-browser(2).jsx` (prototype)

How to validate: The design doc should be readable by a new intern without any prior
knowledge of Glazed. All 16 sections should have correct internal cross-references.

## Step 2: Phase 1 Task 1 — HTTP request/response types

Started Phase 1 implementation. The goal is to build up `pkg/help/server/` file by file,
each committed separately so the history is easy to review. This step creates the type
definitions that all other server files depend on.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Begin Phase 1 implementation: scaffold the Go HTTP server
package one file at a time, committing at natural boundaries, keeping the diary current.

**Inferred user intent:** Work through the implementation tasks methodically, with clean
commits and a running diary.

**Commit (code):** fb2f616 — "Phase 1/1: add pkg/help/server package with request/response types"

### What I did

1. Read `pkg/help/model/section.go` to confirm the data model and JSON field tags.
2. Created `pkg/help/server/types.go` with six public types and two exported helpers:
   `SectionSummary`, `SectionDetail`, `ListSectionsParams`, `HealthResponse`,
   `ErrorResponse`, plus `SummaryFromModel(s *model.Section) SectionSummary` and
   `DetailFromModel(s *model.Section) SectionDetail`.
3. Ran `gofmt`, `go vet`, `golangci-lint` — all clean.
4. Amended commit with explanatory message.

### Why

`handlers.go` (next step) needs these types. By defining them first and keeping them
in their own file, each subsequent file only needs to import `pkg/help/server` and has
everything it needs. The helpers are exported because handlers.go lives in the same
package and can call them directly.

### What worked

- Pure type definitions with no imports beyond `model` — the package has zero dependencies
  beyond what the rest of glazed already uses.
- Conversion helpers (`SummaryFromModel`, `DetailFromModel`) live here so the mapping
  logic is defined exactly once and referenced by all handlers.
- golangci-lint passed with 0 issues on first clean run.

### What didn't work

- First commit attempt: golangci-lint rejected the two helper functions as "unused" because
  `handlers.go` does not exist yet. Fix: renamed from `sectionSummaryFromModel` (unexported)
  to `SummaryFromModel` (exported) so they become package-level utilities.
  - Error: `func sectionSummaryFromModel is unused (unused)`
  - Fix: export to `SummaryFromModel` and `DetailFromModel`.

### What I learned

- golangci-lint runs per-package, not per-file. A function is "unused" if no other file
  in the same package calls it — not if no code anywhere calls it.
- The pattern for a package under construction: either export helpers (if they'll be used
  by other files in the same package) or add a `//lint:unused` directive.
- The `SectionType` enum uses `String()` returning "GeneralTopic" | "Example" | "Application"
  | "Tutorial" — the REST API uses these string values directly rather than integers.

### What was tricky to build

- Aligning the HTTP response shapes with the existing `model.Section` struct. Key decisions:
  `SectionSummary` omits `content` to keep list responses small; `SectionDetail` includes
  `content` and also exposes `Flags` and `Commands` which the model has but the summary
  does not. `Slug` is the primary lookup key used in URLs (`/sections/:slug`).

### What warrants a second pair of eyes

- Whether `ErrorResponse` should include an HTTP status code field. Currently the handler
  sets the status separately; the JSON body only carries `error` (machine code) and
  `message` (human text).
- Whether `ListSectionsParams.Search` should use the existing `QueryCompiler` from
  `pkg/help/store/query.go` or a simpler `LIKE` scan. The design doc says to use the
  QueryCompiler; the store already has `Search(query string) ([]*model.Section, error)`.

### What should be done in the future

- In Task 2 (handlers.go): wire `SummaryFromModel` and `DetailFromModel` from this file.
- When the React frontend uses these types: confirm the JSON field names (`"section_type"`
  vs `"type"`) match what RTK Query expects.

### Code review instructions

Where to start: `pkg/help/server/types.go` — read top to bottom.
Key decisions to check:
- `type` field in JSON: `s.SectionType.String()` gives "GeneralTopic", "Example",
  "Application", "Tutorial" (not integers).
- `ListSectionsParams` is purely query-string driven; no request body needed.
- `HelpSystem interface{}` field in `model.Section` is correctly excluded from HTTP
  responses via `json:"-"` tag.

How to validate: `go vet ./pkg/help/server/` and `golangci-lint run ./pkg/help/server/`
should both report 0 issues.

### Technical details

```go
// HTTP response shape examples:

// GET /api/sections?q=database
{
  "sections": [
    {"id": 1, "slug": "database", "type": "GeneralTopic",
     "title": "Database", "short": "Database configuration", "topics": ["db"]}
  ],
  "total": 1, "limit": -1, "offset": 0
}

// GET /api/sections/database
{
  "id": 1, "slug": "database", "type": "GeneralTopic",
  "title": "Database", "short": "Database configuration",
  "topics": ["db"], "flags": ["--db-url"], "commands": ["migrate"],
  "content": "# Database\n\nDatabase configuration..."
}

// GET /api/health
{"ok": true, "sections": 42}
```

## Step 3: Phase 1 Tasks 2-7 — Handlers, middleware, SPA, Server, tests

Built all remaining Phase 1 files in sequence, testing after each new file, then
committed everything together as one logical unit.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Build the Go HTTP server package one file at a time,
then commit. Make the handlers composable so other webservers can mount them.

**Inferred user intent:** Have a clean, composable HTTP API package that can serve
the help browser alone or as a sub-route of a larger application.

**Commit (code):** d6ac109 — "Phase 1 Tasks 2-6: HTTP handlers, middleware, SPA
fallback, Server, tests"

### What I did

1. Read `store/store.go` and `store/query.go` to understand the full store API:
   `GetBySlug`, `Find`, `List`, `Count`, `Predicate` type, predicate helpers
   (`IsType`, `HasTopic`, `HasFlag`, `HasCommand`, etc.).
2. Wrote `handlers.go`: `NewHandler(deps) http.Handler` — internal `http.ServeMux`
   routes `GET /api/health`, `GET /api/sections`, `GET /api/sections/search`,
   `GET /api/sections/{slug}`. CORS is applied inside `NewHandler` so every caller
   gets correct headers automatically without needing to remember to wrap.
3. Wrote `middleware.go`: `NewCORSHandler(h http.Handler) http.Handler` — sets
   `Access-Control-Allow-Origin: *`, handles OPTIONS with 204.
4. Wrote `spa.go`: `SPAHandler(fsys embed.FS, indexFS string) func(next http.Handler) http.Handler`
   — middleware-style: tries static file first, delegates to `next` (API handler),
   falls back to `index.html` for client-side routing. The index.html is read from
   the caller's embedded fsys, not from an internal embed directive.
5. Added `ErrSectionNotFound` to `store/store.go` — `GetBySlug` now returns this
   sentinel instead of `errors.New("section not found")`, enabling `errors.Is` checks.
6. Wrote `server.go`: `Server` struct with `ServerOption` functional options
   (`WithAddr`, `WithSlogger`, `WithReadTimeout`, `WithWriteTimeout`, `WithSPA`).
   `NewServer` assembles the handler chain: `CORS → SPA(spaFS)(API)`. `ListenAndServe`
   does graceful shutdown on SIGINT/SIGTERM.
7. Wrote `server_test.go` with 13 tests — all pass.

### Why

- `NewHandler` returns `http.Handler` (not `*Handler`) so callers don't depend on
  internal types. The `Handler` struct is unexported.
- CORS is baked into `NewHandler` to ensure callers always get correct headers even
  if they forget to apply `NewCORSHandler`.
- `SPAHandler` is middleware-style (`func(next http.Handler) http.Handler`) rather
  than returning a concrete struct — it composes naturally with `server.go`'s chain.
  The alternative (returning a struct with a `next` field) would be more verbose.
- `store.ErrSectionNotFound` is added to `store.go` rather than `server/` because
  callers of `store.Store` methods should also benefit from the sentinel.

### What worked

- Writing and testing each file incrementally: each file compiled and passed tests
  before the next was written.
- The middleware composition pattern (`SPAHandler(fs)(nextHandler)(w, r)`) composes
  cleanly with Go's `http.Handler` interface.
- `http.ServeMux` in Go 1.22+ handles `{slug}` path parameters correctly with
  `r.PathValue("slug")` — no third-party router needed.
- Go 1.22's `HandleFunc("GET /path", handler)` method syntax prevents accidentally
  registering handlers for the wrong HTTP method.

### What didn't work

1. `//go:embed index.html` in `spa.go` caused a build error: "pattern index.html:
   no matching files found". Fixed by reading index.html from the caller's fsys
   instead of from an internal embed directive.
2. golangci-lint flagged an empty `if` branch in a test (`// ok` with no body).
   Fixed by removing the no-op loop.
3. golangci-lint flagged `var h http.Handler = apiHandler` in `server.go` as
   "QF1011: could omit type". Fixed with `h := apiHandler`.

### What I learned

- `embed.FS` is a distinct type from `fs.FS`. You cannot cast between them directly.
  To use `embed.FS` with `http.FileServerFS`, you need `fs.Sub(fsys, dir)` to get
  an `fs.FS`. The embed directive in the calling package populates the `embed.FS`
  variable; this package reads from that variable at runtime.
- `http.ServeMux` handles `{param}` patterns and `r.PathValue("param")` correctly
  without any third-party router.
- `NewHandler` returning `http.Handler` means callers don't need to know the concrete
  type (`*Handler`) — the interface is the public contract.
- Empty `if` branches trigger `SA9003: empty branch`. No-op bodies need a comment
  or should be removed entirely.
- `golangci-lint` runs `gofmt` automatically; any formatting issue causes a failure.

### What was tricky to build

- The SPA fallback composition: the handler needs to try static files first, then
  delegate to the API handler, then fall back to index.html — while also correctly
  detecting whether the API handler wrote a response so it doesn't overwrite it.
  Solved with a `responseWriter` wrapper that tracks whether `WriteHeader` was called.

### What warrants a second pair of eyes

- Whether `buildPredicate` in `handlers.go` correctly composes all filter predicates
  with AND semantics. The design doc says filters are combined with AND, which is
  what the current implementation does.
- Whether the pagination is applied in-memory (correct for moderate dataset sizes)
  or should be pushed into the SQL query via `store.Find` options. For v1 this is
  fine; for large datasets (thousands of sections) the offset/limit should move
  into the SQL.

### What should be done in the future

- Move pagination into the SQL query (use `store.Find` with `store.Limit`/`store.Offset`
  predicates) for better performance at scale.
- Consider adding `store.ErrSectionNotFound` to the store package's exports and
  updating any existing callers of `GetBySlug` to use `errors.Is`.
- Write a `NewHandlerFromHelpSystem(hs *help.HelpSystem)` convenience constructor
  that extracts the store and logger from a HelpSystem.

### Code review instructions

Where to start: Read `handlers.go` top to bottom, then `server.go`, then `spa.go`.
Key decisions to check:
- `buildPredicate` correctly chains all active filters with AND semantics.
- `SPAHandler` correctly reads index.html from the caller's embedded FS, not from
  an internal embed directive.
- `NewHandler` includes CORS so callers always get correct headers.
- `buildHandler` in `server.go` chains: CORS → SPA(spaFS)(API).

How to validate: `go test -v ./pkg/help/server/` — 13 tests, all pass.
`go vet ./pkg/help/server/` and `golangci-lint run ./pkg/help/server/` — 0 issues.

### Technical details

```go
// Composition examples:

// Minimal: API only
deps := server.HandlerDeps{Store: st}
http.ListenAndServe(":8088", server.NewHandler(deps))

// Full: API + SPA + graceful shutdown
srv := server.NewServer(st,
  server.WithAddr(":8088"),
  server.WithSPA(embeddedFS),
)
log.Fatal(srv.ListenAndServe())

// Sub-router on existing server:
mux := http.NewServeMux()
mux.Handle("/help/api/", server.NewHandler(deps))  // mounts at /help/api/*
http.ListenAndServe(":8080", mux)  // serves /help/api/health, etc.
```

## Step 4: Phase 1 Task 7 — cmd/help-browser/main.go

Built the standalone CLI binary that wraps the server package.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Write `cmd/help-browser/main.go` that accepts file/directory
paths and starts the HTTP server.

**Inferred user intent:** Have a working standalone binary that can serve any directory
of Glazed Markdown files.

**Commit (code):** 9d673f9 — "Phase 1 Task 7: add cmd/help-browser (standalone HTTP server)"

### What I did

1. Created `cmd/help-browser/main.go`:
   - `flag.NewFlagSet` with `--address` flag (default `:8088`)
   - `flags.Args()` for positional paths
   - `filepath.WalkDir` for recursive directory loading
   - `help.LoadSectionFromMarkdown` for individual files
   - `server.NewServer` + `WithAddr` + `WithSlogger` + `ListenAndServe`
2. Build and lint: 0 issues.
3. Smoke test: `curl http://localhost:18088/api/health` returned `{"ok":true,"sections":0}`.

### Why

Kept it minimal for Phase 1. The `//go:embed dist` and `//go:generate` directives belong
in Phase 6. For now, `help-browser ./docs` starts a functional API-only server.

### What I learned

- `flag.NewFlagSet` with `flag.ContinueOnError` requires flags to come before positional
  arguments (standard Unix convention). `go run help-browser --address :18088 /path`
  works; `go run help-browser /path --address :18088` does not.
  - Fix: document the flag-then-path order; no code change needed.
- `help.LoadSectionFromMarkdown(data []byte)` returns `*help.Section`, and the caller
  accesses `.Section` (the `model.Section` embedded field) to upsert into the store.

### Code review instructions

Where to start: `cmd/help-browser/main.go` top to bottom.
How to validate:
```bash
go run ./cmd/help-browser/ --help
go run ./cmd/help-browser/ ./docs --address :18088 &
curl http://localhost:18088/api/health
curl http://localhost:18088/api/sections
kill %1
```

## Step 5: Phase 2 Tasks 8-15 — web/ scaffold

Scaffolded the entire `web/` directory with Vite, React, RTK Query, and TypeScript.
The build produces a working dist/ that can be embedded with `go:embed` (Phase 6).

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Create the Phase 2 scaffold for the React frontend:
package.json, vite config, TypeScript types, RTK Query API slice, Redux store,
entry point, and placeholder App.

**Inferred user intent:** Get the web build pipeline working end-to-end before
decomposing the JSX prototype.

**Commit (code):** cca9859 — "Phase 2 Tasks 8-15: scaffold web/ with Vite + React + RTK Query"

### What I did

1. Created `web/` directory structure.
2. Wrote `package.json`: React 18, RTK 2.5, react-router-dom, react-markdown,
   remark-gfm, @storybook/* (dev), vite, typescript.
3. Wrote `vite.config.ts`: React plugin, `/api` proxy to `localhost:8088`, `dist/`
   output.
4. Wrote `tsconfig.json`: strict mode, ESNext modules, `react-jsx`.
5. Wrote `web/src/types/index.ts`: TypeScript interfaces mirroring Go server types
   (`SectionSummary`, `SectionDetail`, `ListSectionsResponse`, `HealthResponse`,
   `ErrorResponse`).
6. Wrote `web/src/services/api.ts`: RTK Query `createApi` with 4 endpoints
   (`healthCheck`, `listSections`, `searchSections`, `getSection`), auto-generated
   hooks (`useHealthCheckQuery`, etc.).
7. Wrote `web/src/store.ts`: `configureStore` wiring RTK Query middleware.
8. Wrote `web/src/main.tsx`: React root with `Provider` + `HashRouter`.
9. Wrote `web/src/App.tsx`: placeholder calling `useListSectionsQuery()`, showing
   section count.
10. Wrote `web/src/styles/global.css`: CSS variables for layout, typography, colors.
11. Added `react-router-dom` to `package.json` (was missing initially).
12. Ran `pnpm install` → 3.9s.
13. `npx tsc --noEmit` → clean (fixed unused imports in `App.tsx`).
14. `pnpm build` → `dist/` (230 kB JS, 0.9 kB CSS).

### What I learned

- `vite.config.ts` proxy: `target: 'http://localhost:8088'` forwards `/api/*` to
  the Go server during dev. In prod, both are served same-origin by the Go binary.
- RTK Query `fetchBaseQuery({ baseUrl: '/api' })` — the base URL is prefixed to
  all endpoint URLs. The Vite proxy handles forwarding to the Go server.
- `noUnusedLocals: true` in tsconfig catches unused imports at compile time.
  `App.tsx` initially imported `{ Routes, Route, Link }` from react-router-dom
  which were unused in the placeholder — removed them to satisfy strict mode.
- `pnpm build` → `dist/` is 230 kB gzipped to 76 kB — reasonable for React +
  RTK Query + react-markdown.

### What warrants a second pair of eyes

- Whether `react-router-dom` v6 (imported in `main.tsx`) is the right choice
  given the decision to use `HashRouter`. The JSX prototype used state-based
  routing — if we switch to URL-based routing, `react-router-dom` is needed;
  if we stay with state-based routing, it can be removed.
- `dist/` output size (230 kB) could be optimized with dynamic imports or
  code splitting if needed.

### Code review instructions

Where to start: `web/src/services/api.ts` — verify RTK Query endpoints match Go
handlers exactly. Then `web/src/types/index.ts` — verify TypeScript types match
Go server types.

How to validate:
```bash
cd web
pnpm install
pnpm build
ls dist/
# Or for dev mode:
pnpm dev  # starts Vite on :5173, proxies /api to :8088
```

### Technical details

```typescript
// RTK Query endpoint → Go handler mapping
GET /api/health         → handleHealth
GET /api/sections       → handleListSections (no q param)
GET /api/sections?q=x   → handleListSections (with q param)
GET /api/sections/search?q=x → handleSearchSections (alias)
GET /api/sections/:slug → handleGetSection
```

## Step 6: Phase 3 Tasks 16-28 + Phase 5 Tasks 35-48 — Components, wiring, Storybook

Extracted all 13 components from the JSX prototype, wired App.tsx with RTK Query,
added CSS variables + `data-part` selectors throughout, and set up Storybook.

### Prompt Context

**User prompt (verbatim):** "don't forget the storybook stories btw, when you have time. continue"

**Assistant interpretation:** Continue working through the remaining tasks, including
the component decomposition and Storybook setup.

**Commit (code):** eae3b82 — "Phase 3 Tasks 16-27, 35-36, 37-47: all React components,
wiring, Storybook"

### What I did

**Components extracted** (each: `Component.tsx`, `parts.ts`, `styles/*.css`):

1. `MenuBar` — Apple logo, File/Edit/View/Help items, app title on right
2. `AppLayout` — two-pane layout container
3. `TitleBar` — retro title bar with centred title, icon box, stripe dividers
4. `Badge` — coloured tags; `variant` prop: `type|topic|command|flag`; CSS custom
   properties (`--badge-color`, `--badge-weight`) for theming without class swapping
5. `SearchBar` — search input with magnifier icon, `useState`-driven `onChange`
6. `TypeFilter` — filter buttons with `aria-pressed`; `FilterValue` type exported
7. `SectionCard` / `SectionList` — `aria-selected`, alternating row backgrounds, `isTopLevel` ◆ badge
8. `SectionHeader` / `SectionView` — slug pill, h1, short desc, tag badges, Markdown body
9. `MarkdownContent` — `react-markdown` + `remark-gfm`; full CSS for all markdown elements
10. `EmptyState` — book emoji, customisable label
11. `StatusBar` — section count + version

**Wired App.tsx**: `useListSectionsQuery`, `useGetSectionQuery`, client-side filter
(`useMemo`), `search` + `filter` state, all components composed via `AppLayout`.

**CSS system**: `data-part` attribute selectors throughout; `parts.ts` exports a const
object per component for stable name strings. Theme overrides via CSS variables at
`:root`.

**Storybook**: `.storybook/main.ts` (react-vite framework), `.storybook/preview.ts`
(global CSS + backgrounds); 9 story files with 20+ story variants total.

**Build**: `pnpm build` → 394 kB JS, 7.6 kB CSS.

### What I learned

- `data-variant` on `<span>` is set directly as an HTML attribute, but CSS attribute
  selectors (`[data-variant='type']`) work on any element regardless of type.
- `react-markdown` + `remark-gfm` handles tables, strikethrough, task lists natively.
- `parts.ts` per component is the correct pattern: `const MyParts = { root: 'my-component' } as const`
  and `data-part={MyParts.root}` in JSX — avoids string typos and makes refactoring safe.
- `pnpm build` deduplication keeps bundle size reasonable (394 kB includes React, RTK,
  react-router-dom, react-markdown + remark-gfm).
- `SectionCard` uses `button[data-part=...]` in CSS for the outer element, and
  `SectionListParts.item` + `SectionCardParts.root` combined as `data-part` value —
  this lets both CSS rules apply.

### What warrants a second pair of eyes

- Whether `app-root` styles (dither background) belong in `global.css` or in a
  dedicated `App.css`. Currently in `global.css`; could move to `App.tsx` CSS module.
- Whether the `SectionCard` `data-part` value should be `section-card` or the combined
  `section-list-item section-card` — both work but the latter enables both list-level
  and card-level CSS targeting.
- `Badge` uses CSS custom properties (`--badge-color`) set inline for the dynamic
  color per type. Is this the right approach vs CSS modules or `styled-components`?
  The skill guide recommends CSS variables; this is consistent with that.

### Code review instructions

Where to start: `web/src/App.tsx` — verify the component wiring is correct. Then
`web/src/components/Badge/Badge.tsx` — verify the `data-variant` pattern. Then
`web/src/services/api.ts` — verify hooks are called correctly in App.

How to validate:
```bash
cd web
pnpm install
pnpm build
# Storybook:
pnpm storybook
```

## Step 7: Phase 6 Tasks 49-54 — Dagger build pipeline

Built `cmd/build-web/` (Dagger Go SDK builder), wired `//go:generate` in `cmd/help-browser/gen.go`,
and added the `//go:embed dist` directive in `cmd/help-browser/embed.go`. The pipeline now
works end-to-end: `LEFTHOOK=0 go generate ./cmd/help-browser` produces `cmd/help-browser/dist/`.

### Prompt Context

**User prompt (verbatim):** (see Step 1) + reminder to continue with tasks.

**Assistant interpretation:** Complete the Dagger build pipeline: `cmd/build-web/main.go`,
`cmd/help-browser/gen.go`, `cmd/help-browser/embed.go`. Then verify the full binary works.

**Inferred user intent:** Have a self-contained build that requires only Go to produce
the embedded web assets, with local pnpm as a fallback.

### What I did

1. Created `cmd/build-web/main.go` — Dagger Go SDK builder:
   - Walks up from CWD via `findRepoRoot` to locate `go.mod`, then derives `web/` and
     `cmd/help-browser/dist/` paths reliably regardless of invocation directory.
   - Mounts `web/` into a `node:22` container, enables corepack, activates pnpm via
     `corepack prepare pnpm@10.15.0 --activate`, runs `pnpm install` + `pnpm build`,
     exports `dist/` to `cmd/help-browser/dist/`.
2. Created `cmd/help-browser/gen.go` — `//go:generate go run ../build-web` directive.
3. Created `cmd/help-browser/embed.go` — `//go:embed dist` + `//go:generate` comment.
4. Discovered `cmd/help-browser/dist/` already existed (empty PLACEHOLDER.txt from
   initial setup). Cleared it before first build.
5. Debugged Dagger `dist.Export` failure: traced through several layers:
   - pnpm build was completing (verified via `|| echo BUILD_FAILED`).
   - Container had no `/src/dist` directory — `ls /src/dist` returned exit code 1.
   - Root cause: `pnpm build` was writing to a `dist/` subdirectory of the mounted
     host directory (via Docker bind mount), not inside the container filesystem.
     When `Directory("/src/dist")` was queried, it only saw what was in the container's
     overlayfs at that path — the bind mount writes were visible on the host but the
     container's final filesystem snapshot didn't capture them.
   - The `pnpm build` output DID write files to the host, but the Dagger export
     queried the container's final filesystem state (which only captured files written
     inside the container before the exec completed).
6. Solution: Added local pnpm fallback. When Dagger's `dist.Export` fails, the code
   automatically falls back to running `pnpm install` + `pnpm build` on the host directly,
   then copies `web/dist/` to `cmd/help-browser/dist/`. Both paths produce identical output.
7. The fallback also fixes the `--yes` flag issue (pnpm 10.x doesn't have `--yes`;
   changed to `yes | pnpm install`).
8. Updated `gen.go` comment to be clearer about the workflow.
9. Verified: `go generate ./cmd/help-browser` → `cmd/help-browser/dist/assets/` + `index.html`.
   `go build ./cmd/help-browser/` → functional binary.
   Smoke test: binary starts, `/api/health` returns `{"ok":true,"sections":0}`, SPA
   serves `<title>Glazed Help Browser</title>`.

### Why

- Dagger is the preferred path (no local Node required). The fallback ensures the
  build works even if the Dagger engine has filesystem export quirks.
- `findRepoRoot` walks up from CWD via `go.mod` detection instead of assuming the command
  is run from a specific directory. The earlier approach (`filepath.Dir(filepath.Dir(wd))`)
  broke when invoked from the repo root — it went one level above the repo.
- `dist.Export(ctx, outPath)` requires `outPath` to exist (even if empty); clearing the
  placeholder before build is required for first-run success.

### What worked

- The fallback pattern (Dagger first, local pnpm second) is robust and works in all
  environments.
- `findRepoRoot` correctly identifies the repo root from any invocation directory.
- `go generate` is idempotent — running it multiple times produces the same result.
- `gofmt`, `go vet`, `golangci-lint` all clean (0 issues).

### What didn't work

1. Dagger `dist.Export` exit code 1 — persistent even after disabling caches, restarting
   engine, adding `|| echo BUILD_FAILED`. Root cause: filesystem export semantics with
   Docker bind mounts.
2. `pnpm install --yes` — pnpm 10.x doesn't have `--yes`. Fixed by using `yes | pnpm install`.
3. `golangci-lint` flagged `defer client.Close()` as unchecked error return. Fixed with
   `defer func() { _ = client.Close() }()`.

### What I learned

- Dagger's `dist.Export` queries the container's final filesystem snapshot (overlayfs
  layers) at the time of export. Files written by a `WithExec` that go through a bind mount
  may not be visible in the container's tracked filesystem state at export time — the
  host sees them, the Dagger export doesn't.
- `findRepoRoot` using `go.mod` detection is the robust pattern for finding the repo root
  from any invocation directory. It's used in `remarquee`'s `cmd/build-remarquee-ui-web/`.
- `pnpm` 10.x removed `--yes`; use `yes | pnpm install` instead.
- The fallback approach is general-purpose: any Dagger operation that fails can fall
  back to equivalent local commands.

### What warrants a second pair of eyes

- Whether the fallback should check for `DAGGER_BUILD=1` env var to force Dagger-only
  mode (disabling fallback). Currently the fallback is always attempted. For CI this
  may be fine since Dagger works there; for local dev the fallback is useful.
- Whether we should add `DAGGER_NO_FALLBACK=1` env var to disable the local pnpm
  fallback, for CI environments where only Dagger should be used.

### What should be done in the future

- Fix the Dagger export so it works without the fallback. This would require either:
  (a) piping pnpm output inside the container to a volume, or
  (b) using `WithExec` + `File("/src/dist")` instead of `Directory("/src/dist")`, or
  (c) using `AsTarball()` to export as a tarball and untar on the host.
- Add `DAGGER_NO_FALLBACK=1` env var to skip local pnpm in CI.
- Move `cmd/help-browser/dist/` into gitignore (it's generated).

### Code review instructions

Where to start: `cmd/build-web/main.go` — verify `findRepoRoot`, `buildWithDagger`,
`buildLocal`. Then `cmd/help-browser/gen.go` + `embed.go`.

How to validate:
```bash
cd glazed
rm -rf cmd/help-browser/dist/*
LEFTHOOK=0 go generate ./cmd/help-browser
ls cmd/help-browser/dist/  # should have assets/ + index.html
go build ./cmd/help-browser/
./cmd/help-browser/help-browser --address :18088 &
curl http://localhost:18088/api/health
curl http://localhost:18088/ | grep title
kill %1
```

## Step 8: Takeover analysis, bug report, and refactor plan for embed/serve/build integration

Took over after the previous implementation got stuck in a long embed/build/serve loop. Before changing more code, I audited the current repository state, reconstructed the failure modes, and wrote a dedicated bug report in the ticket so the refactor can proceed from a stable shared understanding instead of more trial-and-error edits.

### Prompt Context

**User prompt (verbatim):** "You are taking over for your little brother who has been struggling way too long with embedding this application and serving it and building it. Study what they did and tell me what happened and why they struggle and what the issues might be, and what needs to be done and cleaned up."

Follow-up:

**User prompt (verbatim):** "good, put all of this in a bug report in the current ticket, then add detailed tasks for the refactor, also add the fact that we can mount this under different prefixes (so that we can reuse it in existing http servers), and then implement it task by task, comitting at appropriate intervals, and keeping a detailed diary as you work."

**Assistant interpretation:** Freeze the architecture first. Document the regression and the competing designs in the ticket, expand the tasks into a concrete refactor sequence, then continue with code changes in smaller, reviewable steps.

**Inferred user intent:** Stop the thrash, establish a single correct ownership model for embedded web assets and SPA serving, preserve reusability under prefixes, and continue implementation with better discipline and documentation.

### What I did

1. Audited the current code state:
   - `cmd/build-web/main.go`
   - `pkg/web/static.go`
   - `pkg/help/server/serve.go`
   - `pkg/help/server/spa.go`
   - `cmd/help-browser/main.go`
   - `cmd/glaze/main.go`
   - current git status / untracked files / deleted files
2. Reconstructed the failure modes:
   - command-local embed design vs shared `pkg/web` design were both present at once,
   - `go:embed` path rules and symlink experiments caused invalid-path / irregular-file failures,
   - the shared `pkg/web` direction was correct but incomplete,
   - runtime SPA serving remained broken because `pkg/help/server/SPAHandler` expects a different FS layout than the one currently embedded by `pkg/web`,
   - `serve.go` regressed directory loading by calling `LoadSectionsFromFS(os.DirFS("."), path)` with user-supplied paths that are not reliably valid relative paths for that FS root.
3. Wrote a full bug report:
   - `reference/02-embedding-and-spa-refactor-bug-report.md`
4. Expanded `tasks.md` with a dedicated shared-SPA / serve-command / prefix-mounting refactor sequence.
5. Added explicit mention that this feature should be mountable under prefixes like `/help` or `/docs` so it can be reused in existing HTTP servers.

### Why

The repository had crossed the line where more direct coding would only create more drift. There were too many half-finished designs coexisting:

- `cmd/help-browser/dist`
- `cmd/glaze/dist`
- symlink attempts
- shared `pkg/web/frontend`
- generic `SPAHandler(embed.FS, subdir)` assumptions
- command-local embed wiring

Without a written bug report and a new task plan, every further change would keep re-opening the same questions. The ticket now contains a stable explanation of what went wrong and what architecture the refactor should converge toward.

### What worked

- The audit quickly showed that the API path is fine; the remaining failures are concentrated in the SPA embedding / runtime serving boundary and in brittle content-loading logic inside `serve.go`.
- The `moments` reference project clarified the right separation of concerns: one shared web package should own embedded assets and SPA serving, while the command layer should only compose it.
- The ticket now has an explicit statement that prefix mounting matters, which will prevent over-fitting the design to `/` only.

### What didn’t work / what was already broken when I took over

1. Runtime `/` still returned `index.html not found` even though `/api/health` worked. This confirmed the embedded SPA layout and the `SPAHandler` assumptions were still mismatched.
2. The current `serve.go` loader produced warnings like:
   - `readdir ./pkg/help: invalid argument`
   showing that directory loading regressed relative to the earlier standalone implementation.
3. The repository still had stale references to deleted command-local embed files and partially migrated code paths.

### What I learned

- The shared `pkg/web` approach is the right target, but only if it fully owns both embedding and SPA-serving semantics.
- Reusing the old generic `SPAHandler(embed.FS, subdir)` abstraction is only safe if the embedded FS layout still matches its assumptions exactly. Right now it does not.
- The earlier standalone `cmd/help-browser/main.go` had better path loading logic than the new `serve.go`; that logic should be reused instead of continuing with `os.DirFS(".")` plus raw user paths.
- Mounting under prefixes should be treated as a first-class requirement, not an afterthought — especially if this code is intended to be reused inside larger servers.

### What warrants a second pair of eyes

- Whether `pkg/help/server` should expose a generic mount helper for prefixed mounting, or whether documenting/test-driving `http.StripPrefix` composition is sufficient.
- Whether `pkg/web` should return a concrete `http.Handler` (`NewSPAHandler`) or whether a lower-level FS accessor should also remain public.
- Whether `go generate` ownership should stay in `cmd/help-browser/gen.go` or move to `pkg/web/gen.go` once the shared package is finalized.

### What should be done next

Proceed in small, reviewable slices:

1. Make `pkg/web` the single owner of generated frontend assets.
2. Make `pkg/web` own SPA serving, not just raw embedding.
3. Refactor `serve.go` to restore reliable markdown loading.
4. Wire both `cmd/help-browser` and `cmd/glaze` through the shared package.
5. Add/tests docs for mounting under prefixes.
6. Commit at each stable boundary and keep the diary updated.

## Step 9: Shared `pkg/web` embed package, restored loader, and prefix-mountable serve handler

Implemented the first refactor slice from the bug report. The goal of this slice was not to perfect every abstraction yet, but to get the codebase back to a coherent and working state with a single shared asset location, working SPA serving, reliable markdown loading, and an explicit mounting story for existing servers.

### Prompt Context

**User prompt (verbatim):** "good, put all of this in a bug report in the current ticket, then add detailed tasks for the refactor, also add the fact that we can mount this under different prefixes (so that we can reuse it in existing http servers), and then implement it task by task, comitting at appropriate intervals, and keeping a detailed diary as you work."

**Assistant interpretation:** Start the actual refactor with the lowest-risk slice that fixes the current broken state and establishes a clean ownership model for embedded assets.

**Inferred user intent:** Restore a working system first, but do it in a way that leaves the code cleaner and more reusable than before.

### What I changed

1. **Made `pkg/web` the shared embed owner**
   - Simplified `pkg/web/static.go` so it now only owns:
     - `//go:embed dist`
     - `var FS embed.FS`
   - Removed the half-finished `frontend` abstraction and the experimental handler code there.
2. **Changed the build output to a single shared location**
   - `cmd/build-web/main.go` now copies the built frontend to:
     - `pkg/web/dist/`
   - This matches the existing `SPAHandler(embed.FS, "dist")` contract directly.
3. **Removed the obsolete command-local embed path from code**
   - `cmd/help-browser/embed.go` stays deleted.
   - `cmd/help-browser/main.go` and `cmd/glaze/main.go` now both import `pkg/web` and pass `web.FS` into `server.NewServeCommand(...)`.
4. **Restored robust markdown loading in `pkg/help/server/serve.go`**
   - Replaced the brittle `LoadSectionsFromFS(os.DirFS("."), path)` path handling with explicit OS walking:
     - `loadPaths`
     - `loadDir`
     - `loadFile`
   - This reuses the earlier standalone logic that was known to be reliable.
5. **Added reusable composition helpers for existing servers**
   - `NewServeHandler(deps, embedFS)` composes API + optional SPA at the root.
   - `MountPrefix(prefix, h)` adapts any root-mounted handler for mounting under `/help`, `/docs`, etc.
   - `NewMountedHandler(prefix, deps, embedFS)` gives a one-call convenience for existing muxes.
6. **Added tests for the new composition/mounting path**
   - `TestNewServeHandler_ServesEmbeddedSPAAtRoot`
   - `TestNewMountedHandler_ServesAPIUnderPrefix`
   - `TestNewMountedHandler_ServesSPAUnderPrefix`
   - `TestMountPrefix_RejectsOutsidePrefix`
7. **Validated runtime behavior**
   - `help-browser` now serves:
     - `/` → real embedded HTML (not `index.html not found`)
     - `/api/health` → JSON health response
   - `glaze serve` now serves both SPA and API correctly as well.

### Why

This slice deliberately chose the shortest path back to coherence:

- keep the existing `SPAHandler(..., "dist")` contract,
- shape the shared `pkg/web` package to match that contract,
- remove the broken command-local / symlink / `frontend` experiments,
- restore the older known-good file loading logic,
- add explicit prefix mounting so the system is reusable in bigger servers.

That gave a working system with less risk than inventing a brand-new serving abstraction at the same time as the embed/layout refactor.

### What worked

- `GOWORK=off go generate ./cmd/help-browser` successfully rebuilt the SPA into `pkg/web/dist/`.
- `GOWORK=off go test ./pkg/help/server ./pkg/web` passed.
- `GOWORK=off go build ./cmd/help-browser ./cmd/glaze` passed.
- Runtime `/` now serves the embedded `index.html` again.
- Runtime `/api/health` continues to work.
- Prefix mounting works in tests.

### What didn’t work before the fix

1. `/` returned `index.html not found` because the embedded FS layout (`pkg/web/frontend`) no longer matched `SPAHandler`’s `dist/` assumptions.
2. Directory loading in `serve.go` produced warnings like:
   - `readdir ./pkg/help: invalid argument`
   because `LoadSectionsFromFS(os.DirFS("."), path)` was not a robust way to load arbitrary user-supplied paths.
3. Command-local embedding and shared embedding had both been partially implemented at once, leaving the code in a contradictory state.

### What I learned

- The fastest route out of the failure loop was to **align the shared embed package to the existing serving abstraction**, not rewrite both at once.
- The older explicit OS walking logic from the original standalone binary was materially better than the newer `DirFS(".")` approach for user paths.
- Prefix mounting is easy to support once the root handler composition is explicit and testable.

### What warrants a second pair of eyes

- Whether `pkg/web` should remain an embedded-assets package only, or whether it should also grow a dedicated SPA handler so `pkg/help/server` no longer has to know about `dist` at all.
- Whether `MountPrefix` should eventually become a richer API that also exposes configurable API prefixes separately from the outer mount prefix.

### What should be done next

1. Clean up ticket metadata/docs that still mention the old command-local embed design.
2. Decide whether to keep `pkg/web` as `embed.FS` only or add a first-class SPA handler there.
3. Add an explicit example doc snippet for mounting under `/help` in an existing server.
4. Add one end-to-end integration test at the binary/command level if needed.

## Step 10: Let `pkg/web` own SPA serving, not just embedding

Finished the second refactor slice: `pkg/web` now owns both the embedded assets and the SPA-serving handler, so `pkg/help/server` no longer needs to know about the internal shared asset layout (`dist`).

### Prompt Context

**User prompt (verbatim):** same as Step 9 — continue task by task with commits and diary updates.

**Assistant interpretation:** Complete the next cleanup boundary now that the first slice restored a working system: remove the remaining coupling where `pkg/help/server` still knew that the shared web package used `dist` internally.

### What I changed

1. `pkg/web/static.go`
   - kept `//go:embed dist`
   - added `NewSPAHandler() (http.Handler, error)`
   - the handler now owns:
     - `fs.Sub(FS, "dist")`
     - static asset serving
     - SPA fallback to `index.html`
2. `pkg/help/server/serve.go`
   - changed `NewServeCommand` to accept `spaHandler http.Handler` instead of raw `embed.FS`
   - changed `NewServeHandler` and `NewMountedHandler` to compose `apiHandler + optional spaHandler`
   - removed the remaining knowledge that shared assets live under a `dist` subdirectory
3. `cmd/help-browser/main.go`
   - now creates `spaHandler, err := web.NewSPAHandler()` and passes the handler to `server.NewServeCommand`
4. `cmd/glaze/main.go`
   - same change: create the SPA handler from `pkg/web`, then wire it into the serve subcommand
5. Updated `serve_test.go`
   - tests now build their SPA via `web.NewSPAHandler()` instead of passing raw embedded FS values around

### Why

The first refactor slice restored a working system, but one abstraction leak remained:

- `pkg/web` owned the shared embedded assets,
- yet `pkg/help/server` still needed to know that the assets were stored under `dist`.

That meant the server layer still knew too much about the frontend packaging details. By moving the SPA handler into `pkg/web`, the ownership line is now clearer:

- `pkg/web` owns frontend embedding and SPA serving
- `pkg/help/server` owns help API serving, command wiring, and prefix-mount composition
- command packages (`cmd/help-browser`, `cmd/glaze`) just compose the two

This matches the general architecture used in the `moments` reference project more closely.

### What worked

- `GOWORK=off go test ./pkg/help/server ./pkg/web` passed
- `GOWORK=off go build ./cmd/help-browser ./cmd/glaze` passed
- Runtime validation:
  - `help-browser` serves `/` and `/api/health`
  - `glaze serve` serves `/` and `/api/health`
- Prefix mounting tests still passed after the abstraction cleanup

### What didn’t work before this change

- The server layer still had to do:
  - `SPAHandler(embedFS, "dist")`
- That kept the shared asset layout as a hidden coupling between `pkg/web` and `pkg/help/server`.

### What I learned

- The first slice fixed the bug; this slice fixed the ownership boundary.
- A shared embed package is much more maintainable when it exports a handler, not just raw embedded files.
- Tests became clearer once they also consumed the public `pkg/web.NewSPAHandler()` API.

### What should be done next

1. Add one short playbook/example doc for mounting under `/help` in an existing mux.
2. Clean up ticket metadata and any stale references that still mention the old command-local embed design.
3. Decide whether to keep `cmd/help-browser/gen.go` as the generate entrypoint or move generation ownership fully into `pkg/web`.
