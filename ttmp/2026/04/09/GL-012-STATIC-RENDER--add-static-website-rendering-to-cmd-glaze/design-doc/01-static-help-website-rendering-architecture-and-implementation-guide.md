---
Title: Static Help Website Rendering Architecture and Implementation Guide
Ticket: GL-012-STATIC-RENDER
Status: active
Topics:
    - glazed
    - help
    - http
    - static-render
    - web
    - site-generator
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/build-web/main.go
      Note: Build pipeline that should remain the single frontend artifact producer
    - Path: cmd/glaze/main.go
      Note: Primary CLI integration point for the proposed render-site command
    - Path: pkg/doc/doc.go
      Note: Built-in document loading behavior that must remain mirrored in static export
    - Path: pkg/doc/topics/25-serving-help-over-http.md
      Note: Current user-facing serve documentation that the static docs should complement
    - Path: pkg/help/help.go
      Note: Canonical HelpSystem API and recursive FS loading semantics
    - Path: pkg/help/model/parse.go
      Note: Canonical markdown parser; avoid parallel parsing logic
    - Path: pkg/help/model/section.go
      Note: Canonical section model that exported JSON must derive from
    - Path: pkg/help/server/serve.go
      Note: Live serve command and current location of reload semantics to share
    - Path: pkg/help/server/types.go
      Note: Existing API payload shapes to mirror in static JSON
    - Path: pkg/help/store/query.go
      Note: Ordering and filtering helpers to reuse instead of rebuilding indexing logic
    - Path: pkg/help/store/store.go
      Note: Canonical store used to build deterministic site snapshots
    - Path: pkg/web/gen.go
      Note: Generate entrypoint for shared frontend assets
    - Path: pkg/web/static.go
      Note: Embedded SPA serving boundary and shared frontend ownership
    - Path: ttmp/2026/04/08/GLAZE-HELP-REVIEW--pre-pr-code-review-glazed-help-browser/design-doc/02-project-report-glazed-serve.md
      Note: Review document that explicitly raised static export as a likely next feature
    - Path: web/src/App.tsx
      Note: Current deep-linking gap because state is local
    - Path: web/src/main.tsx
      Note: HashRouter use reduces static-hosting complexity
    - Path: web/src/services/api.ts
      Note: Current transport abstraction and likely insertion point for static mode
    - Path: web/src/types/index.ts
      Note: Frontend contract currently aligned to Go server types
    - Path: web/vite.config.ts
      Note: Current frontend output and proxy assumptions
ExternalSources: []
Summary: Detailed architecture, design, and implementation guide for exporting Glazed help content as a static website using the same model and frontend as glaze serve.
LastUpdated: 2026-04-09T22:16:24.576048519-04:00
WhatFor: Guide an intern through the current serve architecture, the static-site gap, and a concrete implementation path.
WhenToUse: Use when implementing or reviewing the static website rendering command for cmd/glaze.
---


# Static Help Website Rendering Architecture and Implementation Guide

## Executive Summary

`glaze` already knows how to load help sections from Markdown, store them in the canonical help store, expose them through an HTTP API, and render them in a browser-backed SPA. The missing capability is a second delivery mode that writes the same content to disk as a static website so that users can publish help pages on simple file hosting, object storage, or a static web host without needing a live Go process.

The recommended design is not to build a second documentation system. Instead, keep the current architecture centered on the existing `HelpSystem`, `model.Section`, SQLite-backed `store.Store`, and shared frontend in `web/`, then add a one-shot export command that:

1. loads help content exactly like `glaze serve`,
2. materializes a deterministic static data snapshot,
3. copies the already-built frontend assets, and
4. injects a small runtime configuration so the SPA reads local JSON instead of calling a live `/api` server.

That design keeps one parser, one content model, one browsing UI, and one mental model for future engineers. It also matches the review conclusion from the help-browser work: keep the serving stack centered on one canonical model instead of teaching the codebase two parallel architectures.

## Problem Statement

The current `glaze serve` feature is runtime-oriented. It starts an HTTP server, reads Markdown help pages, and serves both JSON API endpoints and the embedded SPA. That solves the interactive local-browsing use case, but it does not solve these adjacent needs:

- publishing help pages to a static host,
- generating a reviewable artifact during CI,
- shipping help content into environments where running a server is undesirable,
- preserving the exact content view as a directory tree that can be archived or attached elsewhere.

The new feature should feel like `glaze serve`, but instead of listening on `:8088`, it should write a self-contained site to disk.

The core requirement is to avoid re-implementing the help system in a new format. The static renderer should consume the same inputs and preserve the same semantics:

- default to built-in embedded docs when no paths are given,
- allow explicit Markdown files and directories to replace preloaded docs,
- keep the same section metadata,
- keep the same frontend browsing experience as much as practical.

## Scope

### In scope

- add a new static-site export command to `cmd/glaze`,
- reuse current help loading semantics from the `serve` path,
- emit a deterministic static data tree and frontend asset bundle,
- add enough frontend support to run without a live API server,
- document the design and implementation plan thoroughly.

### Out of scope for the first slice

- server-side rendering or SEO-first HTML generation,
- a second non-SPA theme or alternate frontend,
- public deployment automation,
- backward-compatibility shims for an old static export format, because there is no old format yet.

## Current-State Architecture

This section anchors the design to the code that exists today.

### 1. `cmd/glaze` is the runtime entrypoint

`cmd/glaze/main.go` wires the root Cobra command, loads built-in docs, sets up the help system, constructs the SPA handler, and adds the `serve` subcommand.

Relevant code:

- `cmd/glaze/main.go:31-46`

Important observations:

- the built-in docs are loaded before command execution,
- `web.NewSPAHandler()` is the single owner of embedded SPA serving,
- `server.NewServeCommand(helpSystem, spaHandler)` is already a reusable `cmds.BareCommand`,
- the static export command should follow that same wiring style rather than inventing a special-case Cobra path.

### 2. `pkg/doc` is the built-in document source

`pkg/doc/doc.go:8-12` embeds the docs tree and exposes `AddDocToHelpSystem`, which calls `helpSystem.LoadSectionsFromFS(docFS, ".")`.

This matters because the export command should mirror the same default behavior as `serve`:

- no explicit paths means export the embedded docs,
- explicit paths means replace the preloaded docs with those paths.

### 3. `HelpSystem` is the canonical in-memory API

`pkg/help/help.go` is the high-level orchestration layer.

Key points:

- `HelpSystem` owns a `*store.Store` (`pkg/help/help.go:99-117`),
- `LoadSectionsFromFS` recursively walks an `fs.FS`, reads Markdown, parses it, and upserts sections (`pkg/help/help.go:120-154`),
- `AddSection` upserts parsed sections into the store (`pkg/help/help.go:156-162`),
- `GetSectionWithSlug` and `QuerySections` are the canonical read APIs.

This is the first major architectural constraint:

- the static exporter should not parse frontmatter differently,
- the static exporter should not build a second section model,
- the static exporter should operate on the same `HelpSystem` or the same `store.Store`.

### 4. `model.Section` is the canonical content contract

`pkg/help/model/section.go:55-79` defines the shape of a help section:

- `Slug`,
- `SectionType`,
- `Title`, `SubTitle`, `Short`, `Content`,
- metadata arrays: `Topics`, `Flags`, `Commands`,
- display flags: `IsTopLevel`, `IsTemplate`, `ShowPerDefault`, `Order`.

`pkg/help/model/parse.go:11-96` is the canonical Markdown parser using YAML frontmatter.

This is the second major constraint:

- all emitted static JSON should derive from this model or thin view-models built from it,
- file-backed Markdown remains the source of truth,
- any static-site index should preserve slugs and metadata from `model.Section`.

### 5. `store.Store` is the canonical query and persistence layer

`pkg/help/store/store.go` and `pkg/help/store/query.go` define storage and query primitives.

Important details:

- `store.NewInMemory()` gives the default runtime store (`pkg/help/store/store.go:40-43`),
- sections are stored with stable fields and indexes in SQLite (`pkg/help/store/store.go:50-99`),
- list and find operations already exist (`pkg/help/store/store.go:296-320`, `pkg/help/store/query.go:98-121`),
- ordering and predicate helpers already exist (`pkg/help/store/query.go:124-320`).

This matters because a static exporter does not need a separate indexing engine for V1. It can derive stable lists from the current store queries.

### 6. `pkg/help/server` is the live-delivery adapter

`pkg/help/server/serve.go` is the current runtime composition layer.

Relevant responsibilities:

- defines `ServeCommand` and `ServeSettings` (`pkg/help/server/serve.go:27-40`),
- when explicit paths are passed, clears preloaded docs and reloads from the requested files (`pkg/help/server/serve.go:97-105`),
- composes API handler and SPA handler (`pkg/help/server/serve.go:115-170`),
- starts the HTTP server (`pkg/help/server/serve.go:230-259`),
- owns the current path-loading helper functions (`pkg/help/server/serve.go:173-228`).

Current API response shapes are defined in `pkg/help/server/types.go`.

Key live API contracts:

- `GET /api/health`
- `GET /api/sections`
- `GET /api/sections/{slug}`

The static exporter should reuse those shapes wherever possible, because the frontend already mirrors them in TypeScript.

### 7. `pkg/web` owns the embedded SPA

`pkg/web/static.go:11-50` embeds `pkg/web/dist` and serves the Vite build output with SPA fallback semantics.

`pkg/web/gen.go:1-13` and `cmd/build-web/main.go:1-148` define the build ownership:

- `go generate ./pkg/web` runs `cmd/build-web`,
- the Go tool builds `web/` with pnpm and exports to `pkg/web/dist`,
- `pkg/web` is the single source of truth for embedded browser assets.

This is important for the export command because it means the static site does not need a second frontend build product. It should copy or reuse the same `pkg/web/dist` artifact.

### 8. The frontend is already close to static-host-friendly

Three details in `web/` matter a lot:

- `web/src/main.tsx:7-17` uses `HashRouter`,
- `web/src/services/api.ts:11-24` centralizes how the SPA decides where to load data,
- `web/src/App.tsx:15-38` still holds active section state locally instead of deriving it from the URL.

That means:

- the app is already biased toward static hosting because it does not require server-side route rewriting,
- but it is not yet truly deep-linkable because the selected section is not route-driven,
- the API client abstraction is the natural place to switch between live-server mode and static-data mode.

## Gap Analysis

The system already has almost everything needed for static export except the final bridge.

### What already exists

- canonical Markdown parser,
- canonical section model,
- canonical store and query system,
- live API response shapes,
- shared frontend bundle,
- shared browser UI,
- Dagger/local pnpm build flow,
- a review recommendation to avoid parallel architectures.

### What is missing

- a one-shot command that writes site output instead of serving requests,
- a shared package for loading docs without tying the logic to `ServeCommand`,
- a static snapshot format,
- a frontend runtime mode that reads snapshot JSON from disk,
- route-backed section selection so exported links are durable,
- tests that assert exported site layout and determinism.

### The most important design constraint

Do not create:

- one content model for `serve`, and
- another content model for `render-site`.

That would repeat the same drift that the help-browser review explicitly warned about.

## Proposed Solution

The proposed feature is a new command, tentatively named:

```bash
glaze render-site [paths...] --output-dir ./dist/help
```

The command should:

1. create a `HelpSystem` exactly as `glaze serve` does,
2. load embedded docs by default,
3. if explicit paths are supplied, clear the preloaded store and load only those paths,
4. query the store to build a deterministic static snapshot,
5. copy the already-built frontend assets into the target directory,
6. write a runtime config file that tells the SPA to use static data mode,
7. write JSON payloads for the SPA to read without a live API server.

### Recommended package split

The cleanest split is:

- `pkg/help/loader`
  shared file and directory loading helpers currently trapped inside `pkg/help/server/serve.go`
- `pkg/help/site`
  static snapshot building, runtime config emission, and output writing
- `pkg/help/site/cmd.go`
  `RenderSiteCommand` implementing `cmds.BareCommand`

The package names are suggestions, but the responsibilities should stay separated.

### Why a dedicated command instead of `glaze serve --export`

Because these are different operational modes:

- `serve` is long-running, binds a port, and handles signals,
- `render-site` is a deterministic filesystem export that should exit non-interactively,
- the flags are different,
- the failure modes are different,
- the validation strategy is different.

This is one of the few places where a separate command improves clarity rather than adding duplication.

## Proposed Runtime and Data Architecture

### High-level flow

```text
Markdown files / embedded docs
           |
           v
     model.ParseSectionFromMarkdown
           |
           v
        HelpSystem
           |
           v
        store.Store
           |
           +--------------------+
           |                    |
           v                    v
      glaze serve         glaze render-site
           |                    |
           v                    v
   live /api + SPA        static JSON + SPA assets
```

### Static export flow

```text
glaze render-site
    |
    +--> load docs into HelpSystem
    |
    +--> snapshot builder
    |      |
    |      +--> sections list
    |      +--> section details
    |      +--> topic / command / flag indexes
    |      +--> site manifest
    |
    +--> copy pkg/web/dist/*
    |
    +--> write site-config.js
    |
    +--> write data/*.json
    |
    '--> output directory ready for static hosting
```

## Detailed Design Decisions

### Decision 1: Reuse the existing SPA instead of writing a second HTML renderer

Reasoning:

- the current browser experience already exists,
- the current frontend already renders Markdown bodies with `react-markdown`,
- `HashRouter` reduces static-hosting friction,
- the current TypeScript types already mirror Go response types,
- a second renderer would duplicate both the content model and the styling surface.

Consequence:

- the export command mostly writes data and config,
- frontend work is limited to static-mode loading and durable routing,
- future UI improvements automatically benefit both `serve` and `render-site`.

### Decision 2: Reuse live API shapes for static JSON where practical

Reasoning:

- `pkg/help/server/types.go` already defines the shapes the SPA expects,
- the frontend `web/src/types/index.ts` already mirrors those shapes,
- reusing those contracts reduces adapter code.

Suggested mapping:

- `site-data/health.json` shaped like `HealthResponse`
- `site-data/sections.json` shaped like `ListSectionsResponse`
- `site-data/sections/<slug>.json` shaped like `SectionDetail`

Consequence:

- the SPA can switch data sources without changing view components,
- most static-mode code becomes a transport concern instead of a data-model concern.

### Decision 3: Make static-mode runtime selection explicit

Do not infer mode from weak heuristics alone.

Recommended runtime config contract:

```ts
type SiteRuntimeConfig = {
  mode: 'server' | 'static';
  apiBaseUrl?: string;
  dataBasePath?: string;
  siteTitle: string;
};
```

Recommended output file:

- `site-config.js` or `site-config.json`

The frontend boot code should read this config first, then build the appropriate data adapter.

### Decision 4: Extract path-loading logic out of `pkg/help/server/serve.go`

Reasoning:

- current loader helpers are coupled to the serve command file,
- static export needs the same semantics,
- duplication here would be easy but would drift quickly.

Refactor target:

- move `replaceStoreWithPaths`, `loadPaths`, `loadDir`, and `loadFile` into a shared package,
- keep `ServeCommand` and `RenderSiteCommand` as thin orchestration layers.

### Decision 5: Add route-backed section selection

Current gap:

- `web/src/App.tsx:15-22` uses local state for `activeSlug`,
- `HashRouter` is mounted, but no route is driving the selected section.

For static export, this is not good enough because users need durable links.

Recommended route shape:

- `#/sections/<slug>`
- optional filtered views later:
  - `#/topics/<topic>`
  - `#/commands/<command>`

Consequence:

- static sites become bookmarkable,
- live `serve` benefits too,
- the export command does not need per-section HTML files in V1.

## Proposed Output Layout

Recommended exported directory tree:

```text
out/
  index.html
  assets/
    ...
  site-config.js
  site-data/
    health.json
    sections.json
    indexes/
      topics.json
      commands.json
      flags.json
      top-level.json
      defaults.json
    sections/
      help-system.json
      writing-help-entries.json
      ...
```

### Why `site-data/` instead of `/api/`

Using `site-data/` is clearer for a static export because:

- there is no actual API server,
- many static hosts are happier serving explicit `.json` files than extensionless pseudo-endpoints,
- it lets the live `serve` path keep `/api` while the static path keeps a clearly different transport.

This is one intentional divergence from the live HTTP URL surface. The payload shape stays the same; only the filesystem layout changes.

## Proposed Go APIs

### Command settings

```go
type RenderSiteSettings struct {
    OutputDir string   `glazed:"output-dir"`
    BasePath  string   `glazed:"base-path"`
    SiteTitle string   `glazed:"site-title"`
    Paths     []string `glazed:"paths"`
}
```

Suggested first-pass flags:

- `--output-dir`
- `--site-title`
- `--base-path`

Possible later flags:

- `--clean`
- `--overwrite`
- `--include-health`
- `--include-debug-metadata`

### Snapshot contract

```go
type SiteManifest struct {
    Version        string   `json:"version"`
    SiteTitle      string   `json:"siteTitle"`
    DataBasePath   string   `json:"dataBasePath"`
    TotalSections  int      `json:"totalSections"`
    TopLevelSlugs  []string `json:"topLevelSlugs"`
    DefaultSlugs   []string `json:"defaultSlugs"`
}
```

```go
type SiteSnapshot struct {
    Manifest      *SiteManifest
    Health        server.HealthResponse
    SectionList   server.ListSectionsResponse
    SectionBySlug map[string]server.SectionDetail
    TopicIndex    map[string][]string
    CommandIndex  map[string][]string
    FlagIndex     map[string][]string
}
```

These are not final type names, but they show the desired shape: thin, explicit, transport-ready data.

## Pseudocode

### 1. Shared loading

```go
func LoadHelpContent(ctx context.Context, hs *help.HelpSystem, paths []string) error {
    if len(paths) == 0 {
        return nil
    }

    if err := hs.Store.Clear(ctx); err != nil {
        return errors.Wrap(err, "clearing preloaded help store")
    }

    for _, input := range paths {
        if err := loadPath(ctx, hs, input); err != nil {
            return err
        }
    }

    return nil
}
```

### 2. Snapshot builder

```go
func BuildSnapshot(ctx context.Context, st *store.Store, siteTitle, dataBasePath string) (*SiteSnapshot, error) {
    sections, err := st.Find(ctx, store.OrderByOrder())
    if err != nil {
        return nil, errors.Wrap(err, "listing sections")
    }

    details := map[string]server.SectionDetail{}
    summaries := make([]server.SectionSummary, 0, len(sections))
    topics := map[string][]string{}
    commands := map[string][]string{}
    flags := map[string][]string{}

    for _, s := range sections {
        detail := server.DetailFromModel(s)
        summary := server.SummaryFromModel(s)

        details[s.Slug] = detail
        summaries = append(summaries, summary)
        indexMetadata(topics, s.Topics, s.Slug)
        indexMetadata(commands, s.Commands, s.Slug)
        indexMetadata(flags, s.Flags, s.Slug)
    }

    manifest := &SiteManifest{
        SiteTitle: siteTitle,
        DataBasePath: dataBasePath,
        TotalSections: len(sections),
    }

    return &SiteSnapshot{
        Manifest: manifest,
        Health: server.HealthResponse{OK: true, Sections: len(sections)},
        SectionList: server.ListSectionsResponse{
            Sections: summaries,
            Total: len(summaries),
            Limit: -1,
            Offset: 0,
        },
        SectionBySlug: details,
        TopicIndex: topics,
        CommandIndex: commands,
        FlagIndex: flags,
    }, nil
}
```

### 3. Export writer

```go
func RenderSite(ctx context.Context, settings *RenderSiteSettings, hs *help.HelpSystem) error {
    if err := LoadHelpContent(ctx, hs, settings.Paths); err != nil {
        return err
    }

    snapshot, err := BuildSnapshot(ctx, hs.Store, settings.SiteTitle, "./site-data")
    if err != nil {
        return err
    }

    if err := CopyEmbeddedFrontend(settings.OutputDir); err != nil {
        return err
    }

    if err := WriteRuntimeConfig(settings.OutputDir, RuntimeConfig{
        Mode: "static",
        DataBasePath: joinBasePath(settings.BasePath, "site-data"),
        SiteTitle: settings.SiteTitle,
    }); err != nil {
        return err
    }

    if err := WriteSnapshot(settings.OutputDir, snapshot); err != nil {
        return err
    }

    return nil
}
```

## Frontend Design for Static Mode

### Current frontend constraints

Observed in the current code:

- `web/src/services/api.ts` always builds an RTK Query `fetchBaseQuery`,
- `web/src/App.tsx` assumes section detail comes from a fetch call keyed by `activeSlug`,
- `HashRouter` is mounted but not actually used to derive page state.

### Recommended frontend refactor

Split transport from view state.

Suggested pieces:

- `web/src/runtime/config.ts`
  reads `window.__GLAZE_SITE_CONFIG__`
- `web/src/services/dataSource.ts`
  returns either a live server adapter or a static JSON adapter
- `web/src/routes.tsx`
  owns selected slug from the hash route
- `web/src/App.tsx`
  becomes mostly layout and composition

### Static data adapter

In static mode, the app should:

- fetch `site-data/sections.json` for the list,
- fetch `site-data/sections/<slug>.json` for detail pages,
- optionally fetch index files for secondary navigation.

Suggested client contract:

```ts
interface HelpDataSource {
  listSections(): Promise<ListSectionsResponse>;
  getSection(slug: string): Promise<SectionDetail>;
}
```

Then implement:

- `serverDataSource`
- `staticDataSource`

This is simpler than trying to make RTK Query speak both live API URLs and static file layout through a maze of conditionals.

If the team wants to preserve RTK Query, that is still possible by switching `baseQuery`, but the design should stay explicit.

## File-Level Implementation Plan

### Phase 1: Shared loading refactor

Files to touch:

- `pkg/help/server/serve.go`
- new shared package, likely `pkg/help/loader/*.go`

Goal:

- move document loading helpers into shared code,
- keep `ServeCommand` behavior identical.

### Phase 2: Static site package

Files to add:

- `pkg/help/site/cmd.go`
- `pkg/help/site/render.go`
- `pkg/help/site/snapshot.go`
- `pkg/help/site/write.go`
- `pkg/help/site/types.go`

Goal:

- own export-specific orchestration,
- keep write-path logic isolated from loader and frontend code.

### Phase 3: Root command wiring

Files to touch:

- `cmd/glaze/main.go`

Goal:

- build the new command using `cli.BuildCobraCommand(...)`,
- wire it next to `serve`.

### Phase 4: Frontend runtime mode

Files to touch:

- `web/src/main.tsx`
- `web/src/App.tsx`
- `web/src/services/api.ts`
- new `web/src/runtime/*`
- new `web/src/routes/*`

Goal:

- make the SPA load static JSON when exported,
- make section selection route-driven.

### Phase 5: Documentation

Files to touch:

- `pkg/doc/topics/25-serving-help-over-http.md`
- new help doc such as `pkg/doc/topics/26-rendering-help-as-a-static-site.md`
- new ticket playbook in this workspace if useful.

Goal:

- teach users how `serve` and `render-site` differ,
- keep the help system self-documenting.

## Testing Strategy

### Go unit tests

Add tests for:

- shared path loading helpers,
- snapshot contents,
- deterministic ordering,
- runtime config emission,
- writer output layout.

Suggested test style:

- use `t.TempDir()`,
- write a small fixture Markdown tree,
- export the site,
- assert on emitted JSON and file existence.

### Go integration tests

Add an end-to-end test that:

1. creates a temp output directory,
2. exports a site from 1-2 Markdown fixtures,
3. reads `site-data/sections.json`,
4. verifies the expected slug list and detail payloads,
5. verifies `index.html` and `site-config.js` exist.

### Frontend tests

Add tests for:

- static-mode adapter loads section list correctly,
- navigating to `#/sections/<slug>` selects the right page,
- missing slug shows a clear empty or error state.

### Optional browser validation

The simplest manual validation loop for an intern is:

```bash
go run ./cmd/glaze render-site ./pkg/doc --output-dir /tmp/glaze-site
python3 -m http.server 8080 --directory /tmp/glaze-site
```

Then open the served directory in a browser and verify:

- landing page loads,
- section list appears,
- selecting a section loads content,
- reloading a `#/sections/...` URL keeps the same page selected.

## Risks

### Risk 1: Duplicating the content model by accident

This is the main architectural risk.

Mitigation:

- derive static payloads from `model.Section` and `server.*FromModel(...)`,
- do not invent a parallel parser or a separate hand-maintained JSON schema.

### Risk 2: Frontend routing grows more complex than the export itself

The current app has a mounted router but not route-driven state.

Mitigation:

- keep the first route surface minimal,
- only add `#/sections/:slug` in the first pass.

### Risk 3: Static-path handling becomes brittle under nested hosting

If the exported site is hosted under a prefix, asset and data URLs can drift.

Mitigation:

- use a runtime config with explicit `basePath` / `dataBasePath`,
- avoid hardcoding `/api`.

### Risk 4: Exported output becomes nondeterministic

If section ordering depends on filesystem walk order, CI diffs will be noisy.

Mitigation:

- always sort exported lists deterministically,
- add tests that compare stable slug sequences.

## Alternatives Considered

### Alternative A: Pure Go HTML templates, no SPA

Pros:

- zero frontend runtime requirement,
- more crawlable HTML.

Cons:

- duplicates the browser presentation layer,
- duplicates Markdown rendering and styling logic,
- diverges from the live `serve` UI,
- increases maintenance cost immediately.

Decision:

- reject for V1.

### Alternative B: Extend `glaze serve` with `--export`

Pros:

- fewer command names.

Cons:

- conflates long-running server mode with one-shot export mode,
- complicates flags and help text,
- encourages a single command with two different operational lifecycles.

Decision:

- reject for V1; can be reconsidered later as a delegating shortcut.

### Alternative C: Full Vite SSG or SSR stack

Pros:

- stronger pre-rendered HTML story.

Cons:

- heavier build chain,
- drifts away from the existing Go-centric architecture,
- unnecessary for the first feature slice.

Decision:

- reject for V1.

## Recommended Order of Work for an Intern

1. Read the current live architecture first:
   - `cmd/glaze/main.go`
   - `pkg/help/server/serve.go`
   - `pkg/web/static.go`
   - `web/src/services/api.ts`
2. Refactor shared loading out of `serve.go` without changing behavior.
3. Add a tiny `pkg/help/site` package that can write a manifest and section JSON to a temp directory.
4. Add the new command to `cmd/glaze/main.go`.
5. Only after the Go exporter works, teach the frontend to read static data.
6. Add route-backed section selection.
7. Add tests.
8. Add user-facing docs last, once the command shape is stable.

That order keeps the work incremental and debuggable.

## Review Checklist

- Does the new command reuse the same loader semantics as `serve`?
- Does the static snapshot derive from canonical section and server response models?
- Does the frontend stay shared between live and static modes?
- Are exported files deterministic across runs?
- Can a reviewer read one code path and understand both delivery modes?

## Open Questions

- Final command name: `render-site`, `static-site`, or something shorter?
- Should the export path write only the SPA shell and JSON, or also a small host-ready `404.html`?
- Should the static runtime keep RTK Query, or should it use a small explicit data-source abstraction?
- Is `site-data/` the preferred directory name, or would Manuel prefer `data/`?
- Should we expose the snapshot builder as public Go API for embedding into other binaries?

## References

### Primary source files

- `cmd/glaze/main.go`
- `cmd/build-web/main.go`
- `pkg/doc/doc.go`
- `pkg/help/help.go`
- `pkg/help/model/section.go`
- `pkg/help/model/parse.go`
- `pkg/help/store/store.go`
- `pkg/help/store/query.go`
- `pkg/help/server/serve.go`
- `pkg/help/server/types.go`
- `pkg/help/server/handlers.go`
- `pkg/web/gen.go`
- `pkg/web/static.go`
- `web/src/main.tsx`
- `web/src/App.tsx`
- `web/src/services/api.ts`
- `web/src/types/index.ts`
- `web/vite.config.ts`

### Related documentation and prior work

- `pkg/doc/topics/25-serving-help-over-http.md`
- `ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/`
- `ttmp/2026/04/08/GLAZE-HELP-REVIEW--pre-pr-code-review-glazed-help-browser/design-doc/02-project-report-glazed-serve.md`
