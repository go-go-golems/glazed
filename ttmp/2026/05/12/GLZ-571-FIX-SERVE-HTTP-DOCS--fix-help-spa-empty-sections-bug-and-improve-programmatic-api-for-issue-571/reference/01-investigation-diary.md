---
title: Investigation Diary
doc_type: reference
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

# Investigation Diary

## 2026-05-12 — Initial Analysis

### Context
GitHub issue #571 reports that the Help SPA shows 0 sections when using the programmatic API path (`NewServeHandler`/`NewMountedHandler`) without calling `SetDefaultPackage()` first.

### Investigation Steps

1. **Read the issue** — Clear description of the bug, reproduction steps, and suggested fixes.
2. **Read `pkg/help/server/serve.go`** — Found `ServeCommand.Run()` calls `SetDefaultPackage(ctx, "glazed", "")` at line ~155. This is the `glaze serve` path only.
3. **Read `pkg/help/server/handlers.go`** — Found `handleListPackages` normalizes `""` to `"default"` for display, but `handleListSections` uses raw `InPackageVersion` predicate which doesn't normalize.
4. **Read `pkg/help/store/store.go`** — Confirmed `SetDefaultPackage` updates rows WHERE `package_name = ''`.
5. **Read `pkg/help/help.go`** — Confirmed `LoadSectionsFromFS` doesn't set any package name.
6. **Read `web/src/App.tsx`** — Confirmed SPA auto-selects `defaultPackage` from `/api/packages` response, then filters sections by that name.
7. **Read the help doc** `pkg/doc/topics/25-serving-help-over-http.md` — Confirmed programmatic examples are missing `SetDefaultPackage` call.

### Root Cause Chain

```
LoadSectionsFromFS → package_name = ""
/api/packages → normalizes "" to "default" for display
SPA selects "default" as active package
/api/sections?package=default → SQL WHERE package_name = 'default' → 0 rows
```

### Analysis: Is this a docs bug or API bug?

**Both.** The documentation is incomplete, but the real problem is that the API has a trap:
- Creating a `HelpSystem`, loading docs, and creating a `ServeHandler` should Just Work
- The `SetDefaultPackage` requirement is invisible and unintuitive
- The normalization asymmetry in the API (`/api/packages` normalizes, `/api/sections` doesn't) is a genuine inconsistency

### Recommended Fix

Two-pronged approach:
1. **API fix**: Auto-assign default package in `NewServeHandler` — makes the programmatic path Just Work
2. **Docs fix**: Update help entry and godoc to explain the package system

The auto-assign is safe because `SetDefaultPackage` is idempotent (only updates rows with empty names). For the existing `glaze serve` path, it's a no-op since sections already have names.

### Design doc written

Created comprehensive design doc at `design-doc/01-root-cause-analysis-and-fix-strategy-for-help-spa-empty-sections-bug.md` covering:
- Full system architecture (4 layers)
- Root cause chain with code references
- Fix strategy with pseudocode
- Alternatives considered
- Implementation checklist
- API reference quick-look tables
