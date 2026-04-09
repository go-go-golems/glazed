---
Title: Project Report - Glazed Serve
Ticket: GLAZE-HELP-REVIEW
Status: active
Topics:
    - help-browser
    - code-review
    - glazed
    - serve
    - react
    - web
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/glaze/main.go:Root command wiring and built-in doc preload"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/build-web/main.go:Frontend build pipeline with Dagger and local pnpm fallback"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/doc/doc.go:Embedded built-in docs loader"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/help.go:Top-level help orchestration after wrapper cleanup"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/model/parse.go:Canonical markdown/frontmatter parser"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve.go:Serve command path semantics and server composition"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/handlers.go:REST API handlers and predicate construction"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/query.go:Canonical predicate/query layer"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static.go:Embedded SPA handler"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/web/src/App.tsx:Frontend root wiring and current client-side filtering"
ExternalSources: []
Summary: "Detailed project-level report on the Glazed serve/help-browser stack, including backend, SPA embedding, build pipeline, and cleanup outcomes"
LastUpdated: 2026-04-08T20:00:00-04:00
WhatFor: "Provide a self-contained architectural project report inside the review ticket"
WhenToUse: "When reviewing the help-browser subsystem as a whole, beyond individual code-review findings"
---

# Project Report - Glazed Serve

This project adds a browser-facing help system to Glazed. The result is that the `glaze` binary can now serve its help documentation over HTTP, expose a small REST API backed by the existing SQLite help store, and render the documentation through an embedded React SPA instead of only through terminal help and the Bubble Tea TUI.

What started as a help-browser feature turned into a broader serving stack: a `serve` command, a store-backed API, a Vite/React frontend embedded into the Go binary, a Dagger-powered frontend build pipeline, and then a second cleanup phase that removed the legacy wrapper and query layers that had accumulated during implementation.

> [!summary]
> The current system has three important identities:
> 1. a **Glazed help-serving feature** inside the main `glaze` binary
> 2. a **single-binary web app architecture** where a Go backend serves an embedded SPA
> 3. a **cleanup case study** in taking an AI-built subsystem from “working” to “architecturally clean”

## Why this project exists

Before this work, Glazed documentation was rich in content but awkward to browse. The help system already had:

- Markdown help pages with YAML frontmatter
- a `pkg/help/store` SQLite store
- querying by slug, topic, command, and DSL
- terminal rendering through Glazed/Cobra templates and glamour
- an interactive TUI help browser

What it did not have was a browser-native way to explore those docs.

The `serve` work exists to close that gap. The goal is for Glazed to serve its documentation as a web application from the same Go binary, without introducing a separate deployment surface, while still preserving the terminal help system and the existing Markdown-based doc workflow.

A second reason this project matters is architectural: it is now a reference implementation for how to add an embedded SPA to an existing Go CLI while keeping the backend self-contained and the build flow reproducible.

## Current project status

The serving stack is now **implemented and cleaned up**.

### What exists today

- `glaze serve` starts an HTTP server for browsing help docs
- a store-backed REST API under `/api/*`
- an embedded React SPA served from `pkg/web`
- explicit path loading for `.md` files and directories
- mounted-prefix support for embedding under `/help`, `/docs`, etc.
- a shared SQLite-backed help store for CLI help and server help
- a Dagger-based frontend builder with local pnpm fallback
- a full cleanup pass that removed:
  - dead compatibility wrappers
  - duplicate markdown parsing paths
  - the `help.Section` wrapper
  - the `SectionQuery` builder layer

### What is still rough

- the frontend still filters section lists client-side rather than pushing all filtering into API requests
- the build currently still has one stale `go:generate` sharp edge in `pkg/help/server/types.go` (`set-zerolog-level`) that can break `make build` unless removed
- the server/help integration semantics were subtle enough that they needed a follow-up clarification: when explicit paths are provided to `serve`, those paths should be authoritative rather than additive

## Problem Statement

Glazed already had a help system strong enough to power CLI help and a TUI, but not one strong enough to present itself naturally in a browser. The implementation challenge was not just “start an HTTP server” — it was to preserve the existing markdown/store/query architecture, expose it cleanly over HTTP, embed a modern frontend into the same binary, and avoid leaving the result in a half-migrated state full of wrappers and compatibility code.

The project therefore had two linked problems:

1. **Feature problem** — add a browser-serving path for Glazed help docs
2. **Architecture problem** — make the final code read like one coherent system rather than an accumulation of interim migration layers

## Project shape

There are four layers in the current implementation:

1. **Documentation source layer**
   - Markdown files with YAML frontmatter
   - embedded docs in `pkg/doc`
   - optionally user-supplied files/directories at serve time

2. **Go help/storage layer**
   - `pkg/help/model`
   - `pkg/help/store`
   - `pkg/help/dsl`
   - `pkg/help/server`

3. **Web application layer**
   - `web/` Vite + React app
   - RTK Query for API access
   - embedded static assets in `pkg/web/dist`

4. **Binary/build layer**
   - `cmd/glaze/main.go`
   - `cmd/build-web/main.go`
   - `go:generate` + `go:embed`

## Architecture

```mermaid
flowchart TD
    A[Markdown docs\npkg/doc and explicit paths] --> B[model.ParseSectionFromMarkdown]
    B --> C[help.Store / SQLite]
    C --> D[help server handlers]
    D --> E[/api/health]
    D --> F[/api/sections]
    D --> G[/api/sections/{slug}]
    D --> H[/api/sections/search]

    I[React SPA\nweb/] --> J[RTK Query hooks]
    J --> F
    J --> G
    J --> E

    K[cmd/build-web] --> L[pkg/web/dist]
    L --> M[go:embed in pkg/web]
    M --> N[SPAHandler]

    D --> O[NewServeHandler]
    N --> O
    O --> P[glaze serve]
```

The important architectural decision is that the browser app is not a separate service. The same Go process loads the docs, exposes the API, and serves the frontend assets.

## Current user-facing commands

The main user-facing entry point is now:

```bash
go run -tags sqlite_fts5 ./cmd/glaze serve
```

That serves the built-in documentation already loaded into the Glazed help system.

You can also explicitly serve docs from paths:

```bash
go run -tags sqlite_fts5 ./cmd/glaze serve ./pkg/doc/topics --log-level debug
```

Current semantics:

- **no paths** → serve the built-in docs already loaded by `cmd/glaze/main.go`
- **one or more paths** → clear preloaded sections and serve **only** the sections discovered from those paths

This “explicit paths are authoritative” behavior was added after implementation because the earlier additive behavior was surprising.

## Proposed Solution

The solution that now exists in code is:

- preload built-in docs for the normal `glaze` experience
- expose a focused HTTP API over the same help store
- embed a Vite/React SPA into the Go binary
- allow `serve` to either reuse built-in docs or replace them with explicit path-based content
- keep the query/storage/parsing model unified so CLI help, HTTP API, and SPA all share one backend model

The final cleaned-up solution deliberately avoids parallel abstractions. There is now:

- one canonical section model (`model.Section`)
- one canonical markdown parser (`model.ParseSectionFromMarkdown()`)
- one canonical query layer (`store.Predicate`)
- one store-backed serving layer (`pkg/help/server`)

## Backend architecture

The important Go files are:

- `cmd/glaze/main.go` — constructs the root command, loads built-in docs, wires `serve`
- `pkg/doc/doc.go` — embeds and loads built-in Glazed docs into the help system
- `pkg/help/help.go` — top-level help system orchestration
- `pkg/help/model/section.go` — canonical section model
- `pkg/help/model/parse.go` — canonical markdown/frontmatter parser
- `pkg/help/store/store.go` — SQLite-backed store
- `pkg/help/store/query.go` — predicate/query compiler
- `pkg/help/dsl/` — boolean help query DSL
- `pkg/help/server/handlers.go` — HTTP handlers and predicate construction
- `pkg/help/server/serve.go` — CLI-facing `serve` command and path loading semantics
- `pkg/web/static.go` — embedded SPA handler with index fallback

### The current mental model

The cleanest way to think about the backend is:

- **Section model**: one canonical `model.Section`
- **Section parsing**: one canonical `model.ParseSectionFromMarkdown()`
- **Section storage/querying**: one canonical `store.Predicate` + SQLite store
- **Section serving**: one HTTP handler layer that translates URL params into predicates

The cleanup phase was largely about making this mental model true in the code.

## Implementation details

The most important technical story in this project is the transition from “feature added” to “feature made coherent.”

### 1. Loading and serving documentation

The `serve` command is built in `pkg/help/server/serve.go`. It does not itself know how to render docs; it just ensures the help system is populated, then composes the API handler and SPA handler.

At a high level, the runtime flow is:

```text
cmd/glaze/main.go
  -> help.NewHelpSystem()
  -> doc.AddDocToHelpSystem(helpSystem)
  -> web.NewSPAHandler()
  -> server.NewServeCommand(helpSystem, spaHandler)
  -> glaze serve [paths...]
```

Inside `Run()` for the serve command:

```text
if no explicit paths:
    use preloaded built-in docs
else:
    clear store
    walk files/directories
    parse markdown into model.Section
    upsert into store

compose API handler + SPA handler
start HTTP server
```

The key subtlety here is that `cmd/glaze/main.go` preloads built-in docs before the `serve` command even runs. That means path handling has to decide whether user-supplied paths are additive or authoritative. The final implementation makes them authoritative.

### 2. The API surface

The HTTP server is intentionally small. The core routes are:

- `GET /api/health`
- `GET /api/sections`
- `GET /api/sections/search`
- `GET /api/sections/{slug}`

The list endpoint supports:

- `type`
- `topic`
- `command`
- `flag`
- `q` / search
- `limit`
- `offset`

The important cleanup here was that search now goes through `store.TextSearch()` instead of hand-written `LIKE` SQL in the handler layer. That matters because the store abstraction already knows how to pick FTS5 vs non-FTS behavior.

### 3. The SPA embedding flow

The frontend lives in `web/`, but the binary serves `pkg/web/dist`.

That means the build flow is:

```text
web/ source
  -> cmd/build-web
  -> web/dist
  -> copied to pkg/web/dist
  -> embedded with //go:embed in pkg/web/static.go
  -> served by SPAHandler()
```

`pkg/web/static.go` serves real asset paths when they exist and falls back to `index.html` for unknown browser routes. That gives standard SPA deep-link behavior without a separate reverse proxy.

### 4. The cleanup phase

The second half of this project was a code-review-driven cleanup based on both the final code and the 21-hour AI coding transcript that produced it.

The major cleanup results were:

- **duplicated markdown parsing removed**
  - both old parsers replaced by `model.ParseSectionFromMarkdown()`

- **`help.Section` removed**
  - the code now uses `*model.Section` directly
  - this removed a wrapper/back-reference pattern that was no longer architecturally necessary

- **`SectionQuery` removed**
  - the store’s predicate system is now the one actual query layer
  - `cobra.go` and `render.go` build/store explicit predicates directly

- **DSL bridge fixed**
  - an O(N) pattern that created a temporary in-memory SQLite store per section was replaced with a direct `Store.Find()`

- **server search fixed**
  - inline `LIKE` logic replaced with `store.TextSearch()`

This cleanup is important because it changed the serving stack from something that merely worked into something that is understandable.

## Frontend shape

The frontend is a classic small SPA:

- `App.tsx` wires the whole experience together
- `useListSectionsQuery()` loads all sections
- `useGetSectionQuery(activeSlug)` loads the selected section
- component tree includes title bar, search bar, type filter, section list, section view, empty state, and status bar

The main design tradeoff right now is that the browser still does client-side filtering over the list response instead of pushing all filtering into the API. That is fine for the current doc volume but is worth revisiting if the doc corpus grows.

## Build pipeline

The build pipeline is one of the most interesting parts of the project because it tries to preserve the “single Go binary” experience while still having a modern frontend.

`cmd/build-web/main.go`:

- discovers the repo root by walking up to `go.mod`
- builds `web/` in a `node:22` Dagger container
- falls back to local `pnpm` if Dagger fails
- exports the built `dist/` to `pkg/web/dist`

This is a pragmatic design:

- **Dagger path** gives reproducibility and avoids requiring Node.js locally
- **local pnpm fallback** makes development less brittle when Dagger is unavailable

The main open issue here is not architectural but operational: a stray `//go:generate set-zerolog-level Info` directive in `pkg/help/server/types.go` can still break `go generate ./...` / `make build`. That should be removed.

## Design decisions

### Decision 1: Keep a single binary
The serving stack was embedded into `glaze` rather than deployed as a separate frontend/backend pair. This keeps the distribution model simple and matches the nature of CLI help documentation.

### Decision 2: Make explicit paths authoritative
When explicit paths are passed to `serve`, the command now clears preloaded sections first. This is less surprising than additive semantics and makes ad hoc serving/debugging easier.

### Decision 3: Use the store predicate layer directly
The cleanup removed `SectionQuery` and routed all query composition through `store.Predicate`. This reduced duplication and made the actual execution model visible in the code.

### Decision 4: Use one canonical parser
Markdown/frontmatter parsing now lives in `model.ParseSectionFromMarkdown()` so the CLI/server/store all interpret docs the same way.

## Alternatives considered

### Keep the compatibility layers
Rejected. The review showed that the compatibility wrappers (`store.HelpSystem`, `help.Section`, `SectionQuery`) added complexity without preserving any essential behavior.

### Treat explicit serve paths as additive
Rejected. This made `glaze serve ./some/path` behave as “embedded docs plus these extra docs,” which was surprising and made targeted serving/debugging harder.

### Split the browser app into a separate service
Rejected for now. The single-binary model is the main ergonomic win of this feature, and the API surface is simple enough that a co-located embedded SPA is a better fit.

### Push all filtering server-side immediately
Deferred. The current doc corpus is small enough that the frontend’s client-side list filtering is acceptable. It is a good future optimization, not a blocking architectural flaw.

## Important project docs

The most useful repo-local docs for understanding this subsystem are:

- `/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/design-doc/01-help-browser-architecture-and-implementation-guide.md`
- `/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/reference/01-diary.md`
- `/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/reference/02-embedding-and-spa-refactor-bug-report.md`
- `/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/08/GLAZE-HELP-REVIEW--pre-pr-code-review-glazed-help-browser/design-doc/01-code-review-findings.md`
- `/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/08/GLAZE-HELP-REVIEW--pre-pr-code-review-glazed-help-browser/reference/01-diary.md`

These two ticket clusters together tell the whole story: first implementation, then cleanup.

## Current repo shape

Relevant top-level locations now are:

- `cmd/glaze/main.go`
- `cmd/build-web/main.go`
- `pkg/help/`
- `pkg/help/server/`
- `pkg/help/store/`
- `pkg/help/model/`
- `pkg/web/`
- `web/`
- `pkg/doc/`

Notably, the old standalone `cmd/help-browser` path is gone. The browser-serving feature is now centered in the main `glaze` binary.

## What this project taught

This project ended up being as much about **cleaning up an AI-built subsystem** as about adding a web server.

The most durable lessons are:

1. **single-binary Go + SPA is a good fit** for this kind of help/documentation surface
2. **the store/predicate layer was the right abstraction all along**; the wrapper/query-builder compatibility layers only obscured it
3. **reviewing the agent transcript was useful** because it accurately pointed to the high-churn files that later needed cleanup
4. **mechanical refactors are where AST tooling would help most** — this project produced a very concrete case for a Go AST refactoring tool that can do type substitution and wrapper removal safely

## Implementation Plan

The implementation plan is effectively complete, but the remaining near-term operational tasks are:

1. remove the stale `set-zerolog-level` `go:generate` line that breaks `make build`
2. decide whether to keep or remove the simple-query fallback path in `pkg/help/dsl_bridge.go`
3. add a small integration test matrix for `glaze serve` path-loading behavior and Cobra help filtering behavior
4. consider moving the SPA list filtering to API-backed filtering
5. write a short developer note describing the final serve semantics (“embedded docs when no paths, explicit paths replace preloaded docs when provided”)

## Open Questions

- Should the frontend move from client-side filtering to server-side filtering for everything except small local UI refinements?
- Should the simple non-DSL query fallback in `pkg/help/dsl_bridge.go` remain, or should the DSL parser become the only query entry point?
- Should the server eventually expose richer grouping/navigation endpoints, or is the current flat list/detail API enough?
- Should the Dagger build step remain the default long-term, or should the repo bias toward simpler local-first frontend builds?
- Should `glaze serve` eventually support exporting a static site as well as serving dynamically?

## Project working rule

> [!important]
> Keep the serving stack centered on one canonical model, one canonical parser, and one canonical query system. If a new layer exists only to preserve an older mental model, remove it rather than teaching future readers two architectures.

## References

- Ticket findings: `design-doc/01-code-review-findings.md`
- Ticket diary: `reference/01-diary.md`
- Original implementation ticket: `ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/`
- Vault note: `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/08/PROJ - Glazed Serve - Help Browser, Embedded Docs, and SPA.md`
