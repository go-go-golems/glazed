---
title: Root Cause Analysis and Fix Strategy for Help SPA Empty Sections Bug
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

# Root Cause Analysis and Fix Strategy for Help SPA Empty Sections Bug (Issue #571)

## Executive Summary

GitHub issue [#571](https://github.com/go-go-golems/glazed/issues/571) reports that developers who follow the documented "Reuse the API and SPA in your own server" examples get a Help SPA that shows **0 sections** in the sidebar — even though the underlying `/api/sections` endpoint returns data correctly. This is a **documentation bug combined with an API ergonomics problem**. The root cause is a missing `SetDefaultPackage()` call that `glaze serve` performs internally but that the programmatic examples never mention. The right fix is a two-pronged approach:

1. **Fix the API to be harder to misuse** — make `NewServeHandler` auto-assign a default package name when sections have `package_name = ""`, so the programmatic path "just works."
2. **Fix the documentation** — update the help entry `serve-help-over-http` to show `SetDefaultPackage()` as a best-practice call, and add godoc to the relevant types.

This document is written as an intern-level guide: it explains every layer of the system you need to understand, walks through the bug in detail, and provides a step-by-step implementation plan.

---

## Part 1: System Architecture Overview

### What is the Glazed Help System?

The Glazed help system is a documentation engine embedded in Go CLIs built with the Glazed framework. It solves a specific problem: **CLI help text is limited, but rich browsable documentation makes tools much more usable.** Instead of only showing `--help` output, a Glazed-based CLI can serve an interactive web browser for its documentation.

The system has four major layers, each in its own package:

```
┌──────────────────────────────────────────────────────────┐
│                    React SPA (browser)                    │
│                  pkg/help/web/src/                        │
│              User-facing documentation browser            │
├──────────────────────────────────────────────────────────┤
│                  HTTP Server Layer                        │
│               pkg/help/server/                            │
│           REST API: /api/sections, /api/packages          │
├──────────────────────────────────────────────────────────┤
│                SQLite Store Layer                         │
│              pkg/help/store/                              │
│         Query engine, predicates, FTS5 search            │
├──────────────────────────────────────────────────────────┤
│             Help Domain Model                             │
│          pkg/help/ + pkg/help/model/                      │
│    HelpSystem, Section, SectionType, Loaders              │
└──────────────────────────────────────────────────────────┘
```

### Layer 1: Help Domain Model (`pkg/help/`, `pkg/help/model/`)

**Key files:**
- `pkg/help/help.go` — `HelpSystem` struct, `LoadSectionsFromFS()`, `AddSection()`
- `pkg/help/model/section.go` — `Section` struct, `SectionType` enum
- `pkg/help/model/parse.go` — YAML frontmatter parser for markdown files

The domain model defines what a "help section" is:

```go
// pkg/help/model/section.go
type Section struct {
    ID             int64       
    Slug           string      // unique identifier like "serve-help-over-http"
    SectionType    SectionType // GeneralTopic | Example | Application | Tutorial
    PackageName    string      // which tool owns this section (e.g. "glazed", "pinocchio")
    PackageVersion string      // optional version
    Title          string
    Short          string      // one-line description
    Content        string      // full markdown body
    
    Topics         []string    // tags like "help", "http", "serve"
    Flags          []string    // CLI flags this doc relates to
    Commands       []string    // CLI commands this doc relates to
    
    IsTopLevel     bool        // show in main listing
    ShowPerDefault bool        // show without --all
    Order          int         // sort position
}
```

A `HelpSystem` is the top-level container:

```go
// pkg/help/help.go
type HelpSystem struct {
    Store *store.Store  // SQLite backend
}
```

The `HelpSystem` is created with an in-memory SQLite database. You load markdown files into it, and it stores them as `Section` rows. The `LoadSectionsFromFS()` method walks an `embed.FS` or `fs.FS`, parses markdown files with YAML frontmatter, and inserts them.

**Critical detail for the bug:** When sections are loaded via `LoadSectionsFromFS()`, they get `PackageName = ""` and `PackageVersion = ""`. The markdown frontmatter doesn't have a `package_name` field — that field is set externally after loading.

### Layer 2: SQLite Store (`pkg/help/store/`)

**Key files:**
- `pkg/help/store/store.go` — `Store` struct, CRUD operations, `SetDefaultPackage()`
- `pkg/help/store/query.go` — `Predicate` system, `QueryCompiler`, SQL generation
- `pkg/help/store/loader.go` — `Loader` struct for filesystem-to-store sync

The store wraps a SQLite database with a `sections` table:

```sql
CREATE TABLE sections (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name    TEXT NOT NULL DEFAULT '',
    package_version TEXT NOT NULL DEFAULT '',
    slug            TEXT NOT NULL,
    section_type    INTEGER NOT NULL,
    title           TEXT NOT NULL,
    -- ... more columns ...
    UNIQUE(package_name, package_version, slug)
);
```

The unique constraint is on `(package_name, package_version, slug)`. This means two different packages can have sections with the same slug — they're distinguished by package name.

**The `SetDefaultPackage()` method** is the crux of the bug:

```go
// pkg/help/store/store.go
func (s *Store) SetDefaultPackage(ctx context.Context, packageName, packageVersion string) error {
    _, err := s.db.ExecContext(ctx, `
        UPDATE sections
        SET package_name = ?, package_version = ?
        WHERE COALESCE(package_name, '') = ''
    `, packageName, packageVersion)
    return err
}
```

This method updates all sections that have an empty `package_name`, assigning them the given package name. Without this call, sections loaded from embedded markdown files sit in the database with `package_name = ""`.

The **Predicate system** is how queries are built:

```go
// pkg/help/store/query.go
type Predicate func(*QueryCompiler)

// Example predicates:
func InPackage(packageName string) Predicate { ... }
func IsType(sectionType model.SectionType) Predicate { ... }
func HasTopic(topic string) Predicate { ... }
```

When the SPA requests `GET /api/sections?package=glazed`, the handler builds a predicate:

```go
store.InPackageVersion("glazed", "")
```

which compiles to:

```sql
WHERE s.package_name = 'glazed' AND s.package_version = ''
```

If sections have `package_name = ""`, this query returns zero rows.

### Layer 3: HTTP Server (`pkg/help/server/`)

**Key files:**
- `pkg/help/server/handlers.go` — HTTP handlers for all `/api/*` routes
- `pkg/help/server/serve.go` — `ServeCommand`, `NewServeHandler()`, `NewMountedHandler()`
- `pkg/help/server/types.go` — request/response types
- `pkg/help/server/middleware.go` — CORS middleware

The server layer exposes these endpoints:

| Endpoint | Purpose |
|----------|---------|
| `GET /api/health` | Health check, returns section count |
| `GET /api/packages` | Lists all package names with section counts |
| `GET /api/sections` | Lists sections, filterable by `package`, `type`, `topic`, `q` |
| `GET /api/sections/{slug}` | Returns full section detail |

The important type for the bug is `HandlerDeps`:

```go
// pkg/help/server/types.go
type HandlerDeps struct {
    Store  *store.Store
    Logger *slog.Logger
}
```

This is all you need to create an API handler:

```go
handler := server.NewServeHandler(
    server.HandlerDeps{Store: hs.Store},
    spaHandler,
)
```

**The `/api/packages` endpoint** has a subtle normalization:

```go
// pkg/help/server/handlers.go — handleListPackages
for _, info := range infos {
    name := info.Name
    if name == "" {
        name = "default"  // <-- normalizes empty to "default"
    }
    // ...
}
```

So when sections have `package_name = ""`, the `/api/packages` response says the package is called `"default"`. But the actual database still has `""`. When the SPA then queries `GET /api/sections?package=default`, the predicate `InPackageVersion("default", "")` searches for `package_name = 'default'`, which matches nothing.

### Layer 4: React SPA (`web/src/`)

**Key files:**
- `web/src/App.tsx` — root component, wires everything together
- `web/src/services/api.ts` — RTK Query API slice
- `web/src/store.ts` — Redux store configuration
- `web/src/types/index.ts` — TypeScript types mirroring Go types
- `web/src/components/PackageSelector/PackageSelector.tsx` — package dropdown

The SPA uses Redux Toolkit Query (RTK Query) to fetch data. On startup:

1. `App.tsx` calls `useListPackagesQuery()` → `GET /api/packages`
2. The response includes `defaultPackage` (e.g., `"default"` for unassigned sections)
3. `App.tsx` auto-selects this package and sets `selectedPackage`
4. `App.tsx` calls `useListSectionsQuery({ packageName: selectedPackage })` → `GET /api/sections?package=default`
5. The server runs `InPackageVersion("default", "")` which returns 0 rows
6. **The sidebar shows 0 sections**

Here's the relevant code flow in `App.tsx`:

```typescript
// web/src/App.tsx
const { data: packageData } = useListPackagesQuery();
const packages = packageData?.packages ?? [];

useEffect(() => {
    if (!packageData || selectedPackage) return;
    const initialPackage = packageData.defaultPackage || packageData.packages[0]?.name || '';
    setSelectedPackage(initialPackage);
}, [packageData, selectedPackage]);

// This sends GET /api/sections?package=<selectedPackage>
const { data: listData } = useListSectionsQuery(
    selectedPackage ? { packageName: selectedPackage, version: effectiveVersion } : undefined,
);
```

---

## Part 2: Root Cause — The Full Bug Chain

Here is the exact sequence of events that causes the bug:

```
Developer writes:
    hs := help.NewHelpSystem()
    doc.AddDocsToHelpSystem(hs)     // loads .md files, all get package_name = ""
    // MISSING: hs.Store.SetDefaultPackage(ctx, "myapp", "")
    handler := helpserver.NewServeHandler(
        helpserver.HandlerDeps{Store: hs.Store},
        spaHandler,
    )

Database state:
    sections table has 65 rows, ALL with package_name = ""

SPA startup:
    1. GET /api/packages
       → store.ListPackages() returns [{Name: "", SectionCount: 65}]
       → handleListPackages normalizes "" to "default"
       → Response: {packages: [{name: "default", sectionCount: 65}], defaultPackage: "default"}

    2. SPA auto-selects package "default"
       → setSelectedPackage("default")

    3. GET /api/sections?package=default
       → buildPredicate creates store.InPackageVersion("default", "")
       → SQL: WHERE s.package_name = 'default' AND s.package_version = ''
       → Returns 0 rows (DB has package_name = '', not 'default')

    4. Sidebar shows "0 sections" ✗
```

The **asymmetry** is the bug: `/api/packages` normalizes `""` to `"default"` for display, but `/api/sections` uses the raw predicate which searches for the literal string `"default"` in the database. The database still has `""`.

### Why does `glaze serve` work?

Because `ServeCommand.Run()` in `pkg/help/server/serve.go` explicitly calls `SetDefaultPackage()`:

```go
// pkg/help/server/serve.go — ServeCommand.Run()
loaders := buildServeLoaders(s)
if len(loaders) == 0 || s.WithEmbedded {
    if err := hs.Store.SetDefaultPackage(ctx, "glazed", ""); err != nil {
        return fmt.Errorf("assigning embedded package metadata: %w", err)
    }
}
```

This updates all `package_name = ""` rows to `package_name = "glazed"`. The SPA then queries `GET /api/sections?package=glazed`, which works because the database actually contains `"glazed"`.

---

## Part 3: Fix Strategy — API Improvement vs. Documentation-Only

### Assessment: This is primarily an API ergonomics problem, not just a docs problem

A documentation-only fix would mean: "Tell developers to call `SetDefaultPackage()`." But this has several problems:

1. **The API violates the principle of least surprise.** If you create a `HelpSystem`, load docs, and create a `ServeHandler`, you expect it to work. The fact that it silently shows 0 sections is a bad API smell.

2. **The bug is invisible.** There's no error, no warning, no log message. The SPA just shows an empty sidebar. The unfiltered `/api/sections` endpoint works fine, which makes it even more confusing.

3. **The naming is misleading.** `SetDefaultPackage` sounds optional — like setting a preference. In reality, it's required for the SPA to function at all.

4. **The normalization asymmetry is a genuine bug.** The `/api/packages` handler normalizes `""` to `"default"`, but the `/api/sections` handler doesn't normalize when filtering. This inconsistency should be fixed regardless.

### Recommended two-pronged approach

**Fix A (API): Auto-assign default package in `NewServeHandler`**

When `NewServeHandler` is called, it should check if the store has sections with `package_name = ""` and auto-assign them a default package name. This makes the programmatic path "just work" without requiring callers to know about the package name pitfall.

```go
// pkg/help/server/serve.go — NewServeHandler
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    if deps.Store == nil {
        panic("server.NewHandler: deps.Store must not be nil")
    }
    
    // Auto-assign default package to sections that don't have one.
    // This prevents the SPA from showing 0 sections when used programmatically.
    ctx := context.Background()
    _ = deps.Store.SetDefaultPackage(ctx, "default", "")
    
    apiHandler := NewHandler(deps)
    // ... rest unchanged
}
```

**Fix B (Docs): Update help entry and godoc**

Even with Fix A, the documentation should explain the package system properly:

- Update `pkg/doc/topics/25-serving-help-over-http.md` to mention `SetDefaultPackage()` as a best practice
- Add godoc to `HandlerDeps` and `SetDefaultPackage()` explaining the package name requirement
- Fix the normalization inconsistency in `/api/packages` vs `/api/sections`

---

## Part 4: Detailed Implementation Guide

### Fix A: Auto-assign default package

**File: `pkg/help/server/serve.go`**

Change `NewServeHandler` to auto-assign:

```go
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    if deps.Store == nil {
        panic("server.NewHandler: deps.Store must not be nil")
    }
    if deps.Logger == nil {
        deps.Logger = slog.Default()
    }

    // Auto-assign a default package name to sections that have package_name = "".
    // Sections loaded via LoadSectionsFromFS get package_name = "", but the SPA's
    // package filter requires a non-empty name. Without this, the SPA shows
    // "0 sections" even though the data is in the store.
    ctx := context.Background()
    if err := deps.Store.SetDefaultPackage(ctx, "default", ""); err != nil {
        deps.Logger.Warn("failed to auto-assign default package", "error", err)
        // Non-fatal: the API still works, just the SPA might show 0 sections.
    }

    apiHandler := NewHandler(deps)
    if spaHandler == nil {
        return apiHandler
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cleanPath := stdpath.Clean("/" + r.URL.Path)
        if cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/") {
            apiHandler.ServeHTTP(w, r)
            return
        }
        spaHandler.ServeHTTP(w, r)
    })
}
```

**Consideration: Is this a behavior change?**

The `ServeCommand.Run()` already calls `SetDefaultPackage(ctx, "glazed", "")` before `NewServeHandler`. So for the `glaze serve` path:
- Sections get `package_name = "glazed"` first
- Then `NewServeHandler` calls `SetDefaultPackage(ctx, "default", "")`
- But `SetDefaultPackage` only updates rows WHERE `package_name = ''`, which is none
- So it's a no-op for the existing `glaze serve` path ✓

For the programmatic path (issue #571):
- Sections have `package_name = ""`
- `NewServeHandler` auto-assigns them `package_name = "default"`
- SPA queries `GET /api/sections?package=default` → works ✓

**Alternative: Fix the normalization inconsistency instead**

Instead of auto-assigning, fix the `/api/sections` handler to normalize `package_name = ""` when querying, similar to how `/api/packages` normalizes for display:

```go
// In buildPredicate or handleListSections:
if params.PackageName == "default" {
    // Treat "default" as matching both "" and "default"
    preds = append(preds, store.Or(
        store.InPackage(""),
        store.InPackage("default"),
    ))
}
```

This is more complex and introduces a "magic" package name. The auto-assign approach is cleaner.

### Fix B: Documentation updates

**File: `pkg/doc/topics/25-serving-help-over-http.md`**

Add a section about package assignment. In the "Mount at the server root" example, add a comment:

```go
func main() {
    hs := help.NewHelpSystem()

    // Load your own help pages here.
    // err := hs.LoadSectionsFromFS(...)
    // Or: err := doc.AddDocToHelpSystem(hs)

    spaHandler, err := web.NewSPAHandler()
    if err != nil {
        panic(err)
    }

    handler := helpserver.NewServeHandler(
        helpserver.HandlerDeps{Store: hs.Store},
        spaHandler,
    )

    _ = http.ListenAndServe(":18100", handler)
}
```

Add a new troubleshooting row:

| Problem | Cause | Solution |
| --- | --- | --- |
| SPA shows "0 sections" but `/api/sections` returns data | Sections have no package name; the SPA filters by package | NewServeHandler auto-assigns a default package. For manual Store use, call `hs.Store.SetDefaultPackage(ctx, "myapp", "")` after loading docs. |

**File: `pkg/help/server/serve.go`** — Add godoc to `NewServeHandler`:

```go
// NewServeHandler composes the API handler and optional SPA handler for use at
// the server root (/). The returned handler already includes CORS because
// NewHandler applies it internally.
//
// If the Store contains sections with empty package_name (as happens when loading
// via LoadSectionsFromFS), this function automatically assigns them a default
// package name so the SPA's package filter can find them.
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
```

**File: `pkg/help/store/store.go`** — Improve `SetDefaultPackage` godoc:

```go
// SetDefaultPackage assigns package metadata to sections that do not have it yet.
// It updates all rows where package_name is empty, setting them to the given
// packageName and packageVersion.
//
// This is necessary because sections loaded from embedded markdown files (via
// LoadSectionsFromFS) get package_name = "". The SPA's package filter queries
// by name, so sections without a package name won't appear in the sidebar.
// NewServeHandler calls this automatically, but direct Store users should call
// it after loading sections.
func (s *Store) SetDefaultPackage(ctx context.Context, packageName, packageVersion string) error {
```

### Fix C: Add a test for the programmatic path

**File: `pkg/help/server/serve_test.go`**

Add a test that reproduces issue #571:

```go
func TestNewServeHandler_AutoAssignsDefaultPackage(t *testing.T) {
    // Create a help system and load sections WITHOUT calling SetDefaultPackage
    hs := help.NewHelpSystem()
    hs.AddSection(&model.Section{
        Slug:        "test-topic",
        Title:       "Test Topic",
        SectionType: model.SectionGeneralTopic,
        Short:       "A test section",
        // PackageName intentionally left empty — this is the bug scenario
    })

    spaHandler, err := web.NewSPAHandler()
    if err != nil {
        t.Fatalf("web.NewSPAHandler: %v", err)
    }

    handler := NewServeHandler(HandlerDeps{Store: hs.Store}, spaHandler)

    // The SPA should be able to list sections via the default package
    req := httptest.NewRequest(http.MethodGet, "/api/sections?package=default", nil)
    rw := httptest.NewRecorder()
    handler.ServeHTTP(rw, req)

    if rw.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", rw.Code)
    }

    var resp ListSectionsResponse
    if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
        t.Fatalf("decode response: %v", err)
    }

    if resp.Total != 1 {
        t.Fatalf("expected 1 section after auto-assign, got %d", resp.Total)
    }
}
```

---

## Part 5: Key File Reference Map

For the intern working on this fix, here are all the files you need to understand and modify:

### Files to modify

| File | What to change |
|------|---------------|
| `pkg/help/server/serve.go` | Add `SetDefaultPackage` call in `NewServeHandler`. Update godoc on `NewServeHandler`. |
| `pkg/help/store/store.go` | Improve godoc on `SetDefaultPackage`. |
| `pkg/doc/topics/25-serving-help-over-http.md` | Add troubleshooting row, mention `SetDefaultPackage` best practice. |
| `pkg/help/server/serve_test.go` | Add `TestNewServeHandler_AutoAssignsDefaultPackage` test. |

### Files to read (no changes needed)

| File | Why it matters |
|------|---------------|
| `pkg/help/help.go` | Defines `HelpSystem`, `LoadSectionsFromFS()`, `AddSection()` |
| `pkg/help/model/section.go` | Defines `Section` struct with `PackageName` field |
| `pkg/help/store/store.go` | Defines `Store`, `SetDefaultPackage()`, `ListPackages()`, SQL schema |
| `pkg/help/store/query.go` | Defines `Predicate`, `InPackage()`, `InPackageVersion()` |
| `pkg/help/server/handlers.go` | HTTP handlers, `handleListPackages` normalizes "" to "default" |
| `pkg/help/server/types.go` | `HandlerDeps`, request/response types |
| `pkg/help/loader/sources.go` | Content loaders, `MarkdownPathLoader` calls `SetDefaultPackage` |
| `web/src/App.tsx` | SPA startup logic, auto-selects `defaultPackage` |
| `web/src/services/api.ts` | RTK Query API, resolves base URL for mounted deployments |
| `cmd/glaze/main.go` | Shows how `glaze serve` wires everything together |

### Data flow diagram

```
                    ┌─────────────────┐
                    │  embed.FS / fs   │
                    │  (markdown files)│
                    └────────┬────────┘
                             │ LoadSectionsFromFS()
                             │ parses YAML frontmatter
                             │ sections get package_name=""
                             ▼
                    ┌─────────────────┐
                    │   HelpSystem    │
                    │  (help.go)      │
                    │                 │
                    │  Store *Store ──┼──► SQLite (in-memory)
                    └────────┬────────┘    sections table:
                             │             package_name = ""
                             │
         ┌───────────────────┼───────────────────────┐
         │                   │                       │
    glaze serve         programmatic            direct Store
    (serve.go)           path (#571)             usage
         │                   │                       │
         ▼                   ▼                       │
  SetDefaultPackage    MISSING THIS CALL              │
  ("glazed", "")             │                       │
  package_name →        package_name                  │
  "glazed" ✓            stays "" ✗                    │
         │                   │                       │
         ▼                   ▼                       ▼
  ┌──────────────────────────────────────────────────────┐
  │              NewServeHandler(deps, spa)               │
  │                                                       │
  │   GET /api/packages  → normalizes "" to "default"     │
  │   GET /api/sections?package=default → searches        │
  │       WHERE package_name='default' → 0 rows ✗         │
  │       WHERE package_name='glazed'  → N rows ✓         │
  └──────────────────────────────────────────────────────┘
```

---

## Part 6: Implementation Checklist

### Step 1: Add auto-assign to `NewServeHandler`

- [ ] Open `pkg/help/server/serve.go`
- [ ] In `NewServeHandler`, add `deps.Store.SetDefaultPackage(context.Background(), "default", "")` after the nil check
- [ ] Log a warning if it fails but don't panic (non-fatal degradation)
- [ ] Update the godoc on `NewServeHandler` to document this behavior

### Step 2: Improve godoc on `SetDefaultPackage`

- [ ] Open `pkg/help/store/store.go`
- [ ] Expand the doc comment to explain why it's needed and when it's called automatically

### Step 3: Add regression test

- [ ] Open `pkg/help/server/serve_test.go`
- [ ] Add `TestNewServeHandler_AutoAssignsDefaultPackage` (see pseudocode above)
- [ ] Run `go test ./pkg/help/server/...` to verify

### Step 4: Update help documentation

- [ ] Open `pkg/doc/topics/25-serving-help-over-http.md`
- [ ] Add a troubleshooting row for "SPA shows 0 sections"
- [ ] Add a brief note in the programmatic examples about `SetDefaultPackage` as a best practice
- [ ] Mention that `NewServeHandler` auto-assigns, so most users don't need to call it manually

### Step 5: Verify end-to-end

- [ ] Run `go test ./pkg/help/...` — all tests pass
- [ ] Run `glaze serve` — still works, shows correct sections
- [ ] Build a test program using the programmatic API (no `SetDefaultPackage` call) — SPA should now show sections

---

## Part 7: Alternatives Considered

### Alternative 1: Documentation-only fix

Just update the help entry and godoc. Leave the API as-is.

**Pros:** No code changes, no risk of regression.
**Cons:** Every developer using the programmatic API hits this bug first. It's a paper-cut that damages trust in the framework. The normalization asymmetry remains.

**Verdict:** Insufficient. The API should be harder to misuse.

### Alternative 2: Fix the SPA to handle empty package names

Modify the SPA to not filter by package when there's only one package (the implicit "default").

**Pros:** Works without any backend changes.
**Cons:** Makes the SPA more complex. Breaks multi-package serving. Doesn't fix the root cause (sections with empty package_name).

**Verdict:** Wrong layer. The backend should have consistent data.

### Alternative 3: Make `LoadSectionsFromFS` accept a package name parameter

Change the signature to `LoadSectionsFromFS(f fs.FS, dir string, packageName string)`.

**Pros:** Sections get their package name at load time, which is the right place.
**Cons:** Breaking API change. All callers need to be updated. `LoadSectionsFromFS` is called from many places.

**Verdict:** Too invasive for a bug fix. Consider for a future v2 API cleanup.

### Alternative 4 (recommended): Auto-assign in `NewServeHandler`

Call `SetDefaultPackage` inside `NewServeHandler`.

**Pros:** No API changes. Existing code starts working. No risk to `glaze serve` path (it's a no-op there). Fixes the bug at the boundary where it matters.
**Cons:** Slightly magical behavior — the handler mutates the store. But it's a reasonable default and documented.

**Verdict:** Best cost-benefit ratio. Safe, minimal, and effective.

---

## Part 8: Open Questions

1. **Should the auto-assign package name be configurable?** Currently we'd use `"default"`. Should `HandlerDeps` get a `DefaultPackageName string` field? This adds complexity but gives control. Recommendation: Start with hardcoded `"default"`, add the field later if requested.

2. **Should `handleListSections` also normalize empty package names?** The `/api/packages` handler normalizes `""` to `"default"`. Should the sections handler do the same? With Fix A, this becomes moot since sections won't have empty names after auto-assign. But for robustness, adding normalization in the query layer would be defensive.

3. **Should there be a log message when auto-assign happens?** A debug-level log like "auto-assigned N sections to package 'default'" would help troubleshooting. Yes, add this.

---

## Appendix A: Pseudocode for the Complete Fix

### serve.go — NewServeHandler

```go
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    if deps.Store == nil {
        panic("server.NewHandler: deps.Store must not be nil")
    }
    if deps.Logger == nil {
        deps.Logger = slog.Default()
    }

    // BUGFIX (#571): Auto-assign a default package name to sections loaded
    // without one. Sections loaded via LoadSectionsFromFS get package_name = "".
    // The SPA's package filter queries by name, so without this, the SPA shows
    // "0 sections" even though /api/sections (unfiltered) returns data correctly.
    // This is a no-op when sections already have a package name (e.g., from
    // ServeCommand.Run which calls SetDefaultPackage explicitly).
    ctx := context.Background()
    if err := deps.Store.SetDefaultPackage(ctx, "default", ""); err != nil {
        deps.Logger.Warn("failed to auto-assign default package", "error", err)
    }

    apiHandler := NewHandler(deps)
    if spaHandler == nil {
        return apiHandler
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cleanPath := stdpath.Clean("/" + r.URL.Path)
        if cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/") {
            apiHandler.ServeHTTP(w, r)
            return
        }
        spaHandler.ServeHTTP(w, r)
    })
}
```

### serve_test.go — New test

```go
func TestNewServeHandler_AutoAssignsDefaultPackage_Issue571(t *testing.T) {
    // Reproduce issue #571: sections loaded without a package name should
    // still appear in the SPA after NewServeHandler auto-assigns one.
    hs := help.NewHelpSystem()
    hs.AddSection(&model.Section{
        Slug:        "example-topic",
        Title:       "Example Topic",
        SectionType: model.SectionGeneralTopic,
        Short:       "An example section",
        // PackageName intentionally empty — this is the bug scenario
    })

    // Verify section has no package name
    sections, err := hs.Store.List(context.Background(), "")
    if err != nil {
        t.Fatalf("list: %v", err)
    }
    if len(sections) != 1 || sections[0].PackageName != "" {
        t.Fatalf("expected section with empty package_name")
    }

    spaHandler, err := web.NewSPAHandler()
    if err != nil {
        t.Fatalf("web.NewSPAHandler: %v", err)
    }

    handler := NewServeHandler(HandlerDeps{Store: hs.Store}, spaHandler)

    // Step 1: GET /api/packages should return the default package
    req := httptest.NewRequest(http.MethodGet, "/api/packages", nil)
    rw := httptest.NewRecorder()
    handler.ServeHTTP(rw, req)

    var pkgResp ListPackagesResponse
    json.NewDecoder(rw.Body).Decode(&pkgResp)
    if len(pkgResp.Packages) != 1 || pkgResp.Packages[0].Name != "default" {
        t.Fatalf("expected package 'default', got %v", pkgResp.Packages)
    }

    // Step 2: GET /api/sections?package=default should return 1 section
    req = httptest.NewRequest(http.MethodGet, "/api/sections?package=default", nil)
    rw = httptest.NewRecorder()
    handler.ServeHTTP(rw, req)

    var secResp ListSectionsResponse
    json.NewDecoder(rw.Body).Decode(&secResp)
    if secResp.Total != 1 {
        t.Fatalf("expected 1 section for package 'default', got %d", secResp.Total)
    }
}
```

---

## Appendix B: Key API Reference Quick-Look

| Type | Package | Purpose |
|------|---------|---------|
| `HelpSystem` | `pkg/help` | Top-level container, holds `Store` |
| `Store` | `pkg/help/store` | SQLite-backed CRUD for sections |
| `Section` | `pkg/help/model` | A single help document with metadata |
| `SectionType` | `pkg/help/model` | Enum: GeneralTopic, Example, Application, Tutorial |
| `HandlerDeps` | `pkg/help/server` | Dependencies for HTTP handlers (just `Store`) |
| `Predicate` | `pkg/help/store` | Query builder function type |
| `ContentLoader` | `pkg/help/loader` | Interface for loading from external sources |
| `PackageInfo` | `pkg/help/store` | Summary of a package/version group |

### Key methods

| Method | Location | What it does |
|--------|----------|-------------|
| `NewHelpSystem()` | `pkg/help/help.go` | Creates HelpSystem with in-memory SQLite |
| `LoadSectionsFromFS(f, dir)` | `pkg/help/help.go` | Loads .md files into Store, sections get `package_name = ""` |
| `AddSection(section)` | `pkg/help/help.go` | Upserts a section into Store |
| `Store.SetDefaultPackage(ctx, name, ver)` | `pkg/help/store/store.go` | Updates all `package_name = ""` rows |
| `Store.ListPackages(ctx)` | `pkg/help/store/store.go` | Returns package/version groups with counts |
| `Store.Find(ctx, predicate)` | `pkg/help/store/query.go` | Queries sections with predicate system |
| `NewServeHandler(deps, spa)` | `pkg/help/server/serve.go` | Creates HTTP handler for API + SPA |
| `NewMountedHandler(prefix, deps, spa)` | `pkg/help/server/serve.go` | Creates prefix-mounted handler |
| `NewSPAHandler()` | `pkg/help/web/static.go` | Creates SPA handler from embedded assets |
| `AddDocToHelpSystem(hs)` | `pkg/doc/doc.go` | Loads embedded Glazed docs into HelpSystem |
