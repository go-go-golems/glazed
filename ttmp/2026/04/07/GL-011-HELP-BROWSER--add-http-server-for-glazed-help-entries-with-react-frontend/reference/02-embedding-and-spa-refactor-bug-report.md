---
Title: Embedding and SPA Integration Bug Report
Ticket: GL-011-HELP-BROWSER
Status: active
Topics:
  - glazed
  - help
  - http
  - react
  - vite
  - embed
  - build
  - dagger
  - cobra
DocType: reference
Intent: long-term
Owners:
  - manuel
RelatedFiles:
  - Path: cmd/build-web/main.go
    Note: Current build pipeline; currently copied around multiple output locations during experiments
  - Path: cmd/glaze/main.go
    Note: Main CLI now partially wired to shared web package
  - Path: cmd/help-browser/main.go
    Note: Standalone binary now partially wired to shared web package
  - Path: pkg/help/server/serve.go
    Note: Current serve command; mixes content loading, API wiring, and SPA mounting
  - Path: pkg/help/server/spa.go
    Note: Existing SPA fallback middleware assumes embedded FS contains a named subdirectory like dist/
  - Path: pkg/web/static.go
    Note: New shared web package; current root/subdir assumptions are inconsistent with SPAHandler
Summary: Root-cause analysis of the embedding/serving/build struggle and a concrete refactor plan
LastUpdated: 2026-04-08T09:10:00-04:00
---

# Embedding and SPA Integration Bug Report

## Executive summary

The help browser implementation drifted into an inconsistent intermediate state while solving three separate problems at once:

1. how to build the web frontend,
2. how to embed the build output into Go binaries, and
3. how to serve the embedded assets at runtime with SPA fallback semantics.

Each of these was individually solvable, but they were changed simultaneously and repeatedly. The result is that the repository currently contains two partially-overlapping designs:

- an older **command-local embedding** design (`cmd/help-browser/dist`, `cmd/glaze/dist`), and
- a newer **shared package** design (`pkg/web/frontend`, `pkg/web/static.go`).

The newer shared-package direction is the correct one, but it has not been completed. Runtime API serving works; SPA serving is still broken due to a mismatch between the embedded filesystem layout and the assumptions in `pkg/help/server/spa.go`.

## What happened

### 1. The original command-local embedding design mostly worked

The earlier standalone binary design was simple:

- `cmd/build-web` built the SPA,
- `cmd/help-browser/dist` stored the generated files,
- `cmd/help-browser/embed.go` embedded `dist`, and
- `server.WithSPA(staticFS)` mounted the SPA fallback.

This matched the existing `SPAHandler(fsys embed.FS, indexFS string)` contract because `SPAHandler` expects an embedded filesystem whose root contains a named subdirectory such as `dist/`.

### 2. Dagger export problems triggered a redesign

The Dagger builder hit `dist.Export(...)` failures. Those were real and required investigation. The earlier work added:

- a robust `findRepoRoot()` helper, and
- a local pnpm fallback when Dagger export failed.

That part was valid and should be preserved.

### 3. The attempt to share assets between `cmd/help-browser` and `cmd/glaze` caused confusion

The next problem was: both `cmd/help-browser` and `cmd/glaze` need the same SPA assets.

Several approaches were tried:

- embedding from command-local `dist/` directories,
- embedding via `../../...` paths,
- symlinking `cmd/glaze/dist` to `cmd/help-browser/dist`,
- copying to multiple command-local locations,
- introducing a shared `pkg/web` package.

The shared package is the right answer, but the earlier symlink and dual-output experiments left the code in a mixed state.

### 4. The current shared package is conceptually right but wired incorrectly

The current builder now outputs to `pkg/web/frontend/` and `pkg/web/static.go` embeds `frontend`.

That is a good architectural direction.

However, the runtime serving path still routes through `pkg/help/server/SPAHandler`, which expects a different filesystem shape:

- `SPAHandler` assumes the embedded FS contains a subdirectory such as `dist/`,
- it computes `fs.Sub(fsys, indexFS)`, and
- it computes the index path as `filepath.Join(indexFS, "index.html")`.

That assumption no longer matches the new `pkg/web` package, where the assets are embedded differently. As a result, `/api/health` works but `/` returns `index.html not found`.

### 5. Content loading in `serve.go` also regressed

`pkg/help/server/serve.go` currently calls:

```go
hs.LoadSectionsFromFS(os.DirFS("."), path)
```

for a user-supplied directory path.

That is brittle because `LoadSectionsFromFS` expects an FS root and a directory path relative to that FS. Passing `os.DirFS(".")` plus paths like `./pkg/help` or absolute paths can yield invalid arguments and silent partial failures. This is why the runtime warning currently appears:

- `readdir ./pkg/help: invalid argument`

The older standalone `cmd/help-browser/main.go` implementation had more reliable directory walking and per-file loading. That logic should be reused.

## Why the struggle took so long

This was not one bug. It was a chain of interacting problems:

- **build problem**: Dagger export instability,
- **layout problem**: where generated assets should live,
- **embed problem**: `go:embed` path rules and directive placement,
- **runtime problem**: SPA handler expecting a different directory shape,
- **CLI problem**: Cobra integration plus standalone binary behavior,
- **environment problem**: parent `go.work` / toolchain friction adding noise during debugging.

Because those were all moving at once, each fix exposed a new class of issue.

## Root causes

### Root cause 1: shared ownership was unclear

The repository did not have a single obvious owner for the frontend assets.

A clean design needs exactly one package to own:

- where the generated frontend lives,
- how it is embedded, and
- how it is served.

That owner should be `pkg/web`.

### Root cause 2: `SPAHandler` abstraction no longer matched the asset layout

`SPAHandler` is designed for an embedded filesystem that contains a named subdirectory such as `dist/`.

The new `pkg/web` package embeds `frontend/`, but then tries to reuse `SPAHandler` without preserving that original assumption cleanly.

### Root cause 3: `serve.go` does too much

`pkg/help/server/serve.go` currently mixes:

- Cobra command wiring,
- user path loading,
- API handler construction,
- SPA mounting,
- and standalone process lifecycle.

That makes it harder to reuse in existing servers and harder to test in isolation.

## What the code should look like after refactor

## A. Single source of truth for web assets

Use a shared `pkg/web` package.

Responsibilities:

- own the generated frontend directory,
- own `//go:embed`,
- expose a `NewSPAHandler()` function returning `http.Handler`.

The build pipeline should copy `web/dist/` to exactly one place:

- `pkg/web/dist/`

No command-local embedding. No symlinks. No duplicated output locations.

## B. Keep API routing separate from SPA routing

`pkg/help/server` should continue to own the help API.

It should expose composable building blocks like:

- `NewHandler(deps) http.Handler` for API-only routing,
- optionally a small composition helper for combining API and SPA handlers,
- Cobra command wiring that accepts an optional SPA handler or SPA handler factory.

It should **not** need to understand how the frontend was embedded.

## C. Make mounting under prefixes explicit and first-class

The design should support reuse inside existing HTTP servers, for example:

- `/help/api/...` for the API,
- `/help/...` for the SPA,
- or mounting only the API without the SPA.

This can be implemented either by:

1. exposing prefix-aware composition helpers in `pkg/help/server`, or
2. documenting and testing how to mount the existing handlers using `http.StripPrefix` and an outer mux.

The key requirement is that the API and SPA can be reused under prefixes without baking assumptions about being mounted at `/`.

## D. Reuse the older robust content-loading logic

Directory loading should use explicit OS walking or equivalent well-defined path handling, not `os.DirFS(".")` plus user paths.

The prior standalone loader logic in the old `cmd/help-browser/main.go` is a good reference:

- `filepath.WalkDir` for directories,
- per-file `.md` checks,
- `help.LoadSectionFromMarkdown` + `Store.Upsert(...)`.

## Immediate cleanup needed

1. Remove the obsolete command-local embedding design.
2. Remove stale references to `cmd/help-browser/dist`, `cmd/glaze/dist`, and deleted `embed.go` files.
3. Make `pkg/web` the sole embedded asset package.
4. Fix `pkg/help/server/serve.go` to use a proper loader and cleaner composition boundary.
5. Replace ad hoc SPA mounting with a clear shared `pkg/web.NewSPAHandler()` path.
6. Add tests for:
   - shared embedded SPA works,
   - `glaze serve` serves the SPA,
   - mount-under-prefix behavior works.

## Recommended refactor sequence

1. **Documentation and tasks**
   - record the bug report,
   - update tasks and diary,
   - freeze the target architecture.
2. **Shared web package**
   - move the generated output to `pkg/web/dist/`,
   - embed it there,
   - add `pkg/web.NewSPAHandler()`.
3. **Server composition cleanup**
   - simplify `serve.go`,
   - restore robust path loading,
   - stop routing raw embed FS through mismatched abstractions.
4. **Command wiring**
   - wire both `cmd/help-browser` and `cmd/glaze` to shared `pkg/web`.
5. **Prefix mounting support**
   - add helper(s) and tests for mounting under `/help`, `/docs`, etc.
6. **Validation**
   - build both binaries,
   - verify `/`, `/assets/...`, `/api/health`, and prefixed mounting.

## Review checklist

A reviewer should confirm:

- there is only one frontend asset output location,
- only one package owns `//go:embed` for the SPA,
- `glaze serve` and `help-browser` both use the same SPA package,
- mounting under prefixes is either implemented or clearly documented and tested,
- no stale command-local embed paths remain,
- runtime `/` serves the real embedded `index.html`, not an error page.
