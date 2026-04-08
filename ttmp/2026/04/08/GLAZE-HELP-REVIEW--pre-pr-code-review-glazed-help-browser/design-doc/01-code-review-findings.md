---
Title: Code Review Findings
Ticket: GLAZE-HELP-REVIEW
Status: active
Topics:
    - help-browser
    - code-review
    - glazed
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/help.go:Duplicated markdown parsing + backward-compat re-exports + Section wrapper"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/model/section.go:Section model with dead HelpSystem field"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/compat.go:Dead compat HelpSystem wrapper"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/loader.go:Duplicated LoadFromMarkdown"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/query.go:Predicate system (good)"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/query.go:Legacy SectionQuery builder"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/dsl_bridge.go:O(N) temp-store-per-section bug"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/handlers.go:Clean handlers, minor search bypass"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/server/serve.go:ServeCommand with mount helpers"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/web/static.go:SPA handler with embed"
    - "/home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/build-web/main.go:Dagger-based build pipeline"
ExternalSources: []
Summary: "Pre-PR code review findings: bugs, duplication, dead code, and architectural issues"
LastUpdated: 2026-04-08T18:30:00-04:00
WhatFor: "Actionable list of issues to address before or during PR review"
WhenToUse: "When reviewing the help browser PR"
---

# Code Review Findings

## Executive Summary

The glazed help browser adds a HTTP server for browsing Glazed help documentation with a React SPA frontend. The implementation is functional and the server package is well-designed, but there are **2 bugs, 3 instances of dead/duplicated code, and several legacy patterns** that should be addressed before the PR.

## Critical Issues

### CRITICAL-1: DSL bridge creates temp SQLite database per section

**File**: `pkg/help/dsl_bridge.go:74-92`  
**Severity**: Performance bug  
**Impact**: O(N) database creates for DSL queries

The `evaluatePredicate()` method creates a new in-memory SQLite store, inserts one section, and queries it — for **every section** in the help system. With ~50 sections, this means 50 database creates/inserts/queries per DSL search.

```go
func (hs *HelpSystem) evaluatePredicate(predicate store.Predicate, section *model.Section) bool {
    memStore, err := store.NewInMemory()  // NEW DATABASE
    err = memStore.Insert(context.Background(), section)  // INSERT
    results, err := memStore.Find(context.Background(), predicate)  // QUERY
    return len(results) > 0
}
```

**Fix**: Pass the predicate directly to `store.Find()` on the main store, just like the legacy query path does:
```go
func (hs *HelpSystem) QuerySections(query string) ([]*Section, error) {
    predicate, err := dsl.ParseQuery(query)
    // ...
    return hs.Store.Find(ctx, predicate)  // Single query
}
```

### CRITICAL-2: Duplicated markdown parsing

**Files**: `pkg/help/help.go:LoadSectionFromMarkdown()` and `pkg/help/store/loader.go:LoadFromMarkdown()`  
**Severity**: Maintenance burden / bug source  
**Impact**: Two places to keep in sync when adding fields

Both functions parse the same YAML frontmatter with nearly identical logic. Differences:
- `help.go` doesn't initialize empty slices for missing arrays
- `loader.go` handles `float64` for Order (from YAML parsing)
- `help.go` returns `*Section` (wrapper), `loader.go` returns `*model.Section`

**Fix**: Extract to `model.ParseSectionFromMarkdown()` that both callers use.

## Significant Issues

### SIGNIFICANT-1: Dead `store.HelpSystem` compat wrapper

**File**: `pkg/help/store/compat.go`  
**Severity**: Dead code  
**Impact**: Confusion for future developers

The `store.HelpSystem` struct wraps `*Store` + `*Loader` and provides methods like `GetExamplesForTopic()`, `GetTutorialsForTopic()`, etc. It is **never used anywhere** in the codebase. `cmd/glaze/main.go` uses `help.HelpSystem` directly.

**Fix**: Delete `compat.go` entirely.

### SIGNIFICANT-2: Dead `model.Section.HelpSystem` field

**File**: `pkg/help/model/section.go:47`  
**Severity**: Dead code  
**Impact**: Confusion

```go
// Back-reference to the help system (not stored in database)
// Using interface{} to avoid circular dependency
HelpSystem interface{} `json:"-" yaml:"-"`
```

This field is never set and never read. It's a leftover from before the store refactor.

**Fix**: Remove the field.

### SIGNIFICANT-3: `help.Section` wrapper back-reference pattern is legacy

**File**: `pkg/help/help.go`  
**Severity**: Architectural debt  
**Impact**: Unnecessary complexity

The `Section` struct wraps `model.Section` and adds `HelpSystem *HelpSystem` so it can call methods like `DefaultGeneralTopics()`, `OtherExamples()`, etc. These methods each do a database query via `FindSections()`. The `HelpPage` type calls up to 12 of these methods.

Now that there's a proper store with predicates, the back-reference pattern is unnecessary — callers should query the store directly with predicates.

**Fix**: Mark `Section.DefaultGeneralTopics()` etc. as deprecated. Refactor `HelpPage` to use store predicates directly.

### SIGNIFICANT-4: `SectionQuery` builder is 300 lines of legacy API

**File**: `pkg/help/query.go`  
**Severity**: Code bloat  
**Impact**: Maintenance burden

`SectionQuery` is a fluent builder that converts to `store.Predicate` via `toPredicate()`. The `store.Predicate` system already provides all the same functionality with less code. `SectionQuery` exists only because `cmd/cobra.go` and `render.go` use it.

**Fix**: Gradually migrate `cobra.go` and `render.go` to use `store.Predicate` directly. Mark `SectionQuery` as deprecated.

### SIGNIFICANT-5: Server search bypasses FTS5 abstraction

**File**: `pkg/help/server/handlers.go:buildPredicate()`  
**Severity**: Bug (FTS5 never used for search)  
**Impact**: FTS5 build tag is effectively dead for the HTTP API

The `buildPredicate()` function implements search with raw `LIKE`:
```go
preds = append(preds, func(qc *store.QueryCompiler) {
    qc.AddWhere(
        "LOWER(s.title) LIKE ? OR LOWER(s.short) LIKE ? OR LOWER(s.content) LIKE ?",
        "%"+term+"%", "%"+term+"%", "%"+term+"%",
    )
})
```

But `store.TextSearch()` already abstracts FTS5 vs LIKE based on build tags. The server should use `store.TextSearch(term)` instead.

**Fix**: Replace inline LIKE with `store.TextSearch(params.Search)`.

## Minor Issues

### MINOR-1: Frontend client-side filtering duplicates server API

**File**: `web/src/App.tsx`  
The frontend fetches ALL sections and filters client-side. The server API already supports `?type=`, `?q=` params. Fine for v1 (small dataset) but worth noting.

### MINOR-2: Unused `searchSections` RTK Query endpoint

**File**: `web/src/services/api.ts`  
The `searchSections` endpoint is defined but never used in the UI. Remove or use it.

### MINOR-3: TypeScript types re-declare parent fields

**File**: `web/src/types/index.ts`  
```typescript
export interface SectionDetail extends SectionSummary {
  short: string;   // already in SectionSummary
  topics: string[]; // already in SectionSummary
```

### MINOR-4: Pagination is in-memory

**File**: `pkg/help/server/handlers.go:handleListSections()`  
Offset/limit is applied to a Go slice after fetching all results. Fine for small datasets.

### MINOR-5: `example_store_usage.go` is dead code

**File**: `pkg/help/example_store_usage.go`  
This is an `ExampleStoreUsage()` function that's never called. Should be removed or converted to a proper test.

## Architecture Diagram

```
cmd/glaze/main.go
├── help.HelpSystem (original, with Store field)
│   ├── pkg/help/help.go         — Section wrapper, LoadSectionFromMarkdown (DUP)
│   ├── pkg/help/query.go        — SectionQuery builder (LEGACY, 300 lines)
│   ├── pkg/help/dsl_bridge.go   — DSL query bridge (BUG: O(N) temp stores)
│   ├── pkg/help/render.go       — Terminal rendering with glamour
│   └── pkg/help/store/
│       ├── store.go             — SQLite-backed Store (GOOD)
│       ├── query.go             — Predicate system (GOOD)
│       ├── loader.go            — LoadFromMarkdown (DUP with help.go)
│       ├── compat.go            — HelpSystem compat (DEAD CODE)
│       ├── fts5.go/nofts.go     — FTS5 conditional build
│       └── query_fts5.go/query_nofts.go — TextSearch conditional
├── pkg/help/model/
│   └── section.go               — Section model (dead HelpSystem field)
├── pkg/help/server/
│   ├── handlers.go              — HTTP API (CLEAN, minor search bypass)
│   ├── serve.go                 — ServeCommand (GOOD)
│   ├── types.go                 — Request/response types (GOOD)
│   └── middleware.go            — CORS middleware (GOOD)
├── pkg/help/dsl/
│   ├── dsl.go                   — ParseQuery bridge
│   ├── compiler.go              — AST to Predicate compiler
│   ├── lexer.go / parser.go     — Query DSL parser
│   └── *_test.go                — DSL tests
├── pkg/help/cmd/
│   ├── cobra.go                 — Cobra help command integration
│   └── ui/                      — TUI help browser
├── pkg/web/
│   ├── static.go                — SPA handler + embed (GOOD)
│   └── gen.go                   — go:generate directive
└── cmd/build-web/
    └── main.go                  — Dagger/pnpm build pipeline
```

## Recommendations

### Before the PR
1. Fix CRITICAL-1 (DSL bridge O(N) bug)
2. Fix CRITICAL-2 (extract shared markdown parser)
3. Fix SIGNIFICANT-5 (use store.TextSearch in server)

### In the PR or immediately after
4. Delete `store/compat.go` (dead code)
5. Remove `model.Section.HelpSystem` field
6. Delete `example_store_usage.go`

### Future cleanup
7. Deprecate `SectionQuery` builder
8. Refactor `help.Section` wrapper out of existence
9. Move frontend to server-side filtering
10. Simplify build pipeline (drop Dagger, keep pnpm)
