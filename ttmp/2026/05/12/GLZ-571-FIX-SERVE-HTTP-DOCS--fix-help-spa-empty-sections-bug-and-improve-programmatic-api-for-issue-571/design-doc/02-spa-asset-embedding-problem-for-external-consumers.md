---
title: SPA Asset Embedding Problem for External Consumers
doc_type: design-doc
status: active
intent: long-term
topics:
  - help
  - serve
  - http
  - spa
  - api
  - bug
  - paper-cut
  - documentation
  - intern-guide
owners:
  - manuel
ticket: GLZ-571-FIX-SERVE-HTTP-DOCS
created: "2026-05-12"
---

# SPA Asset Embedding Problem for External Consumers

## Executive Summary

While investigating issue #571, a second systemic problem emerged: **the Help SPA frontend assets cannot be embedded by external Go binaries that depend on glazed as a library.** The current build pipeline uses `go:generate` + Dagger to build the React frontend and copy it to `pkg/web/embed/public/`, then a `//go:embed` directive with the `embed` build tag bakes it into the binary. This works for `glaze` itself, but any other binary (pinocchio, sqleton, etc.) that imports `pkg/web` will get a **placeholder HTML page** instead of the real SPA — the generated assets simply don't exist in the Go module cache.

This document explains the full asset pipeline, why it breaks for external consumers, and evaluates potential fixes.

---

## Part 1: The Current Asset Pipeline

### Build flow for `glaze` itself

```
┌──────────────────────────────────────────────────────────────────┐
│ Step 1: go generate ./pkg/web                                    │
│                                                                  │
│   pkg/web/gen.go:                                                │
│     //go:generate go run ../../cmd/build-web                     │
│                                                                  │
│   cmd/build-web/main.go:                                         │
│     1. Find repo root (walk up to go.mod)                        │
│     2. Use Dagger (or local pnpm) to build web/                  │
│     3. Copy web/dist/ → pkg/web/embed/public/                    │
│                                                                  │
│   Result: pkg/web/embed/public/index.html, *.js, *.css           │
├──────────────────────────────────────────────────────────────────┤
│ Step 2: go build -tags embed ./cmd/glaze                         │
│                                                                  │
│   pkg/web/embed.go  (build tag: embed):                          │
│     //go:embed embed/public                                      │
│     var embeddedFS embed.FS                                      │
│     var PublicFS = mustSubFS(embeddedFS, "embed/public")         │
│                                                                  │
│   → index.html, JS, CSS are baked into the glaze binary         │
├──────────────────────────────────────────────────────────────────┤
│ Step 3: Runtime                                                  │
│                                                                  │
│   web.NewSPAHandler() → reads PublicFS → serves real SPA         │
└──────────────────────────────────────────────────────────────────┘
```

**Key files:**

| File | Role |
|------|------|
| `pkg/web/gen.go` | `//go:generate` directive that triggers the build |
| `cmd/build-web/main.go` | Dagger-based build tool, copies output to `embed/public/` |
| `pkg/web/embed.go` | `//go:build embed` — uses `//go:embed` to bake assets |
| `pkg/web/embed_none.go` | `//go:build !embed` — fallback for dev/test builds |
| `pkg/web/static.go` | `NewSPAHandler()` — serves from `PublicFS` |
| `.gitignore` | Excludes `pkg/web/embed/public/` — assets never committed |

### The `embed_none.go` fallback

When the binary is built **without** `-tags embed` (the default), `embed_none.go` kicks in:

```go
// pkg/web/embed_none.go
//go:build !embed

var PublicFS = findPublicFS()

func findPublicFS() fs.FS {
    // 1. Walk up from CWD looking for repo root (go.mod)
    // 2. Check pkg/web/embed/public/index.html on disk
    // 3. Check web/dist/index.html on disk
    // 4. Fall back to a tiny placeholder HTML
    return fallbackPublicFS()
}

func fallbackPublicFS() fs.FS {
    return fstest.MapFS{
        "index.html": &fstest.MapFile{Data: []byte(`<!doctype html>
<html lang="en">
  <head><meta charset="UTF-8" /><title>Glazed Help Browser</title></head>
  <body><div id="root">Glazed Help Browser assets have not been generated.
    Run go generate ./pkg/web for the full browser.</div></body>
</html>`)},
    }
}
```

This fallback is fine for **development** (running `go test` in the glazed repo). But for **external consumers** it produces a broken SPA.

---

## Part 2: Why External Consumers Get No SPA

### What happens when pinocchio does `web.NewSPAHandler()`

```
pinocchio's go.mod:
    require github.com/go-go-golems/glazed v0.38.0

pinocchio's main.go:
    import "github.com/go-go-golems/glazed/pkg/web"
    spaHandler, err := web.NewSPAHandler()

Build:
    go build -o pinocchio ./cmd/pinocchio
```

**Scenario A: No `embed` build tag (default)**

1. `embed_none.go` is compiled (since no `-tags embed`)
2. `findPublicFS()` walks up from CWD looking for `go.mod`
3. Finds pinocchio's `go.mod` (not glazed's)
4. Checks `pkg/web/embed/public/index.html` relative to pinocchio's root — doesn't exist
5. Checks `web/dist/index.html` relative to pinocchio's root — doesn't exist
6. Falls back to `fallbackPublicFS()` — **placeholder page**
7. `NewSPAHandler()` succeeds but serves a useless page

**Scenario B: With `-tags embed`**

1. `embed.go` is compiled
2. `//go:embed embed/public` looks for `pkg/web/embed/public/` relative to the **glazed module source**
3. In the Go module cache (`$GOMODCACHE/github.com/go-go-golems/glazed@v0.38.0/`), that directory **doesn't exist** because:
   - `.gitignore` excludes `embed/public/` — the generated files are never committed
   - `go mod download` only gets committed files
   - The Go module zip deliberately excludes files matching `.gitignore`
4. **Build fails** with: `no matching files found for pattern embed/public`

**Neither scenario works.** External consumers cannot get the SPA assets.

### Why `go generate` doesn't help

`go generate` only runs on source in your own module. When pinocchio runs `go generate ./...`, it does NOT run `go generate` inside glazed's module cache. The `//go:generate` directive in `pkg/web/gen.go` is part of the glazed module, and Go's generate tool doesn't cross module boundaries.

Even if it did, the Dagger build tool (`cmd/build-web/main.go`) walks up from CWD to find the repo root by looking for `go.mod`. In the module cache, this would find glazed's `go.mod` but there's no `web/` directory in the module cache with the React source — only the compiled `.go` files and committed assets.

---

## Part 3: How Multi-Tool Help Serving Actually Works Today

The **current working pattern** for serving help from multiple tools is:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton
```

This works because:

1. Only `glaze` has the SPA embedded in its binary
2. `--from-glazed-cmd` runs `pinocchio help export --output json` to get section metadata
3. The metadata (JSON) is loaded into glaze's help store
4. `glaze` serves its own SPA with all the imported sections

**The SPA is only ever served from `glaze` itself.** Other binaries don't need the SPA — they just export their help data as JSON.

This means the "Reuse the API and SPA in your own server" documentation in the help entry is actually **misleading for external binaries**. The code example works (after the #571 fix), but only if you're building within the glazed repo. An external binary would need the SPA assets, which it can't get.

---

## Part 4: Solutions

### Solution A: Commit the generated assets to the repository

**What:** Remove `pkg/web/embed/public/` from `.gitignore` and commit the built JS/CSS/HTML.

**Pros:**
- External consumers get assets automatically when they `go mod download`
- Works with `//go:embed` as-is
- No pipeline changes needed
- The most straightforward fix

**Cons:**
- Bloats the git repository with generated binary blobs
- Every frontend change creates large diffs
- Need a CI step to ensure committed assets are up to date
- Go module zip size increases

**Implementation:**
1. Remove `/pkg/web/embed/public/` from `.gitignore`
2. Run `go generate ./pkg/web` and commit the output
3. Add a CI check that verifies committed assets match a fresh build
4. Update GoReleaser `before: hooks` to always regenerate (already runs `go generate ./...`)

### Solution B: Make the SPA an opt-in import

**What:** Move the SPA assets to a separate Go module or package that external consumers can import if they want the frontend.

```go
// External consumer who wants the SPA:
import _ "github.com/go-go-golems/glazed/pkg/web/spa-assets"

// External consumer who only wants the API:
import "github.com/go-go-golems/glazed/pkg/web"  // no SPA assets
```

**Pros:**
- Clean separation between API and SPA
- External consumers opt in explicitly
- Binary size stays small for API-only users

**Cons:**
- Requires a separate module or package with its own embedding
- Same problem: the SPA assets still need to be in the module cache somehow
- More complex import path

### Solution C: Serve the SPA from a CDN or static file server

**What:** Don't embed the SPA at all. Instead, serve it from a CDN or let the user provide their own.

```go
// NewSPAHandler accepts an optional URL or fs.FS
func NewSPAHandler(assets fs.FS) (http.Handler, error)
```

**Pros:**
- Completely decouples SPA distribution from Go builds
- Always up-to-date frontend
- External consumers can use their own frontend

**Cons:**
- Requires network access to load the SPA
- Breaks the "single binary" deployment model
- More configuration for users

### Solution D: API-only mode as the documented default for external consumers

**What:** Document that external consumers should use the API-only mode (pass `nil` as SPA handler) and build their own frontend or use `glaze serve --from-glazed-cmd` for the browser.

```go
// Recommended for external binaries:
handler := helpserver.NewServeHandler(
    helpserver.HandlerDeps{Store: hs.Store},
    nil,  // no SPA — use the JSON API directly
)
```

**Pros:**
- No code changes needed
- The JSON API works perfectly for external consumers
- Consistent with the `--from-glazed-cmd` pattern
- Single binary still works

**Cons:**
- External consumers can't get the browser UI without running `glaze serve`
- Limits the "compose in your own server" story to API-only

### Recommendation

**Short term (now): Solution D** — Update the documentation to be honest about the limitation. The "Reuse the API and SPA" section should clearly state that the SPA embedding only works within the glazed repo, and recommend API-only mode or `glaze serve --from-glazed-cmd` for external binaries.

**Medium term: Solution A** — Commit the generated assets. Yes, it bloats the repo, but it's the simplest way to make `//go:embed` work transitively. Add a CI check to verify they're fresh. This is a common pattern (many Go projects commit generated protobuf code, for example).

**Long term: Consider Solution B or C** if the SPA grows large or if there's a need for custom frontends.

---

## Part 5: Updated Architecture Diagram

```
Current state (broken for external consumers):

┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│       glaze         │     │     pinocchio        │     │      sqleton        │
│                     │     │                     │     │                     │
│  Has SPA ✓          │     │  No SPA ✗            │     │  No SPA ✗            │
│  Has help docs ✓    │     │  Has help docs ✓    │     │  Has help docs ✓    │
│                     │     │                     │     │                     │
│  Can serve browser  │     │  Can export JSON    │     │  Can export JSON    │
│  for itself +       │     │  for glaze to       │     │  for glaze to       │
│  imported packages  │     │  import via          │     │  import via          │
│                     │     │  --from-glazed-cmd  │     │  --from-glazed-cmd  │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘

The working pattern:
    glaze serve --from-glazed-cmd pinocchio,sqleton
    → glaze loads pinocchio's and sqleton's help as JSON
    → glaze serves its own SPA with all three packages

The broken pattern (issue #571):
    pinocchio tries:
        handler := NewServeHandler(deps, spaHandler)
    → spaHandler serves placeholder HTML
    → Even if SPA worked, sections show 0 (SetDefaultPackage bug)
```

---

## Part 6: Action Items

1. **Update `pkg/doc/topics/25-serving-help-over-http.md`** — Add a clear note that the SPA embedding only works within the glazed repo. Recommend API-only mode for external consumers.

2. **Consider committing generated assets** — Evaluate the repo size impact. If the SPA is small (likely under 500KB gzipped), committing it is practical.

3. **Add a runtime check** — In `NewSPAHandler()`, if `PublicFS` only contains the placeholder (no JS/CSS), log a warning explaining why the SPA won't work and suggesting alternatives.

4. **Cross-reference with issue #571** — The SetDefaultPackage fix from #571 is still needed regardless. Even if external consumers use API-only mode, they need correct package assignment for their own API clients.
