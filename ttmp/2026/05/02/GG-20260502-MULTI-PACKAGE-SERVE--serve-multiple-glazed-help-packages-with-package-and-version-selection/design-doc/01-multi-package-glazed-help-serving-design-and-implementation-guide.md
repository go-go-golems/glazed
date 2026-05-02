---
Title: Multi-package Glazed help serving design and implementation guide
Ticket: GG-20260502-MULTI-PACKAGE-SERVE
Status: active
Topics:
    - glazed
    - help
    - server
    - frontend
    - sqlite
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/help/loader/sources.go
      Note: Existing JSON
    - Path: glazed/pkg/help/server/handlers.go
      Note: HTTP API routes need package listing and package/version filters.
    - Path: glazed/pkg/help/server/serve.go
      Note: Serve command flags
    - Path: glazed/pkg/help/store/store.go
      Note: SQLite schema currently has globally unique slug and needs package/version identity.
    - Path: glazed/web/src/App.tsx
      Note: Sidebar UI needs Package and conditional Version selectors.
    - Path: glazed/web/src/services/api.ts
      Note: RTK Query API needs packages endpoint and package-aware section queries.
ExternalSources:
    - /tmp/pi-clipboard-a7fb96dd-2edd-449d-8be1-753df53c7e3a.png
Summary: Design for serving multiple Glazed help packages from embedded docs, commands, and recursively discovered SQLite exports.
LastUpdated: 2026-05-02T10:44:00-04:00
WhatFor: Use when implementing --from-sqlite-dir and the package/version selectors in the Glazed help web app.
WhenToUse: Before modifying pkg/help/server, pkg/help/loader, pkg/help/store, or web/src for multi-package help browsing.
---


# Multi-package Glazed help serving design and implementation guide

## Executive summary

The current `glaze serve` command can serve one logical help corpus assembled from embedded Glazed documentation plus optional Markdown paths, JSON exports, SQLite exports, and command-driven exports. That is useful for a single package, but it does not preserve where each section came from. When several packages are loaded at once, sections are merged into one flat namespace keyed by `slug`, collisions overwrite earlier sections, and the web UI has no package or version selector.

This ticket proposes a multi-package help serving model with three main changes:

1. Add a `--from-sqlite-dir` flag to `glaze serve`.
   - The flag points at one or more root directories.
   - Each root is scanned recursively for SQLite help export files.
   - Directory layout determines package/version metadata:
     - `X/Y/X.db` means package `X`, version `Y`.
     - `X/X.db` means package `X`, no version.
     - `X.db` means package `X`, no version.
2. Preserve package/version metadata for every loaded section, whether it comes from:
   - embedded Glazed docs,
   - `--from-glazed-cmd`,
   - `--from-sqlite`,
   - `--from-sqlite-dir`,
   - JSON exports, or
   - Markdown paths.
3. Update the HTTP API and React web app so users can select a package and, only when versions exist for that package, select a version.

The target UI is the screenshot attached at `/tmp/pi-clipboard-a7fb96dd-2edd-449d-8be1-753df53c7e3a.png`. It keeps the existing Classic Mac dual-pane style but adds two selectors above the type filter:

- `Package` dropdown, for example `Gepetto`.
- `Version` dropdown, for example `v5.1.0`; this is hidden or disabled when the selected package has no versions.

The highest-risk implementation detail is that the current store schema has no package/version fields and enforces uniqueness on `sections.slug`. Multi-package serving requires either extending the store schema to include `package_name` and `package_version` and changing uniqueness to `(package_name, package_version, slug)`, or introducing a separate serving index that wraps sections with package metadata. This guide recommends extending the store and API because it gives one source of truth, keeps filters server-side, and avoids in-memory routing hacks.

## Problem statement and scope

### Current user problem

A user wants to browse help documentation for several Go/Glazed packages at once. Typical packages in this workspace include `glazed`, `pinocchio`, and tools such as `codebase-browser`. Today a developer can try commands like:

```bash
glaze serve --with-embedded --from-glazed-cmd pinocchio
```

or:

```bash
pinocchio help export --format sqlite --output-path /tmp/help/pinocchio.db
glaze serve --from-sqlite /tmp/help/pinocchio.db
```

However, once multiple sources are loaded, the server presents one flat section list. The UI does not tell the user which package a section belongs to. If different packages use the same slug, the last loaded section replaces the previous one. There is also no support for serving a release archive such as:

```text
help-db-root/
  geppetto/
    v5.1.0/
      geppetto.db
    v5.0.0/
      geppetto.db
  pinocchio/
    pinocchio.db
  glazed.db
```

The requested feature solves that by making package/version metadata first-class in loading, storage, API responses, URLs, and UI state.

### In scope

This design covers:

- `glaze serve` CLI flag design.
- Recursive SQLite discovery rules for `--from-sqlite-dir`.
- Source metadata inference for embedded docs, commands, explicit SQLite files, JSON, and Markdown paths.
- Store schema changes needed to distinguish duplicate slugs across packages/versions.
- HTTP API changes for package/version discovery and filtered section retrieval.
- React/RTK Query frontend changes for the package and version selectors shown in the screenshot.
- Testing strategy for loader, store, handlers, and frontend behavior.
- A phased implementation plan for a new intern.

### Out of scope

This design does not require:

- A full static-site multi-package build, though the API shape should not make that impossible later.
- Editing the content format of Glazed Markdown docs.
- Adding package/version metadata to every Markdown frontmatter file immediately.
- Implementing authentication, uploads, or remote package registry features.

## Current-state architecture with file evidence

### Main serving command

The `serve` command lives in `glazed/pkg/help/server/serve.go`. The command settings currently include address, Markdown paths, JSON sources, SQLite sources, command sources, and an embedded-doc merge flag:

- `ServeSettings` is defined at `glazed/pkg/help/server/serve.go:36-44`.
- Existing source flags are registered at `glazed/pkg/help/server/serve.go:81-100`.
- The long help explicitly says sources can be JSON, SQLite, or Glazed-compatible binaries at `glazed/pkg/help/server/serve.go:62-65`.

Current settings:

```go
type ServeSettings struct {
    Address       string   `glazed:"address"`
    Paths         []string `glazed:"paths"`
    FromJSON      []string `glazed:"from-json"`
    FromSQLite    []string `glazed:"from-sqlite"`
    FromGlazedCmd []string `glazed:"from-glazed-cmd"`
    WithEmbedded  bool     `glazed:"with-embedded"`
}
```

The command builds a single list of loaders, clears the preloaded store unless `--with-embedded` is true, then runs each loader into the same `HelpSystem` store. That sequence is at `glazed/pkg/help/server/serve.go:123-151`.

Important implication for the intern: the serve command currently thinks in terms of one `HelpSystem` and one `Store`. There is no package registry object in the runtime path.

### Loader layer

External source loading is in `glazed/pkg/help/loader/sources.go`.

The existing loader interface is small:

```go
type ContentLoader interface {
    Load(ctx context.Context, hs *help.HelpSystem) error
    String() string
}
```

`SQLiteLoader` opens every configured SQLite path as a store, lists all sections, then upserts them into the destination `HelpSystem`. See `glazed/pkg/help/loader/sources.go:104-132`.

`GlazedCommandLoader` runs a binary's help export command and imports JSON. See `glazed/pkg/help/loader/sources.go:138-170`.

Current command export invocation:

```go
cmd := exec.CommandContext(ctx, binary, "help", "export", "--with-content=true", "--output", "json")
```

This matters because `--from-glazed-cmd pinocchio` can load sections, but the loader currently does not mark them as `package=pinocchio`. It only logs the source string.

### Export layer

The export command supports SQLite output in `glazed/pkg/help/cmd/export.go`:

- The command help says `--format sqlite` produces a portable SQLite database at `glazed/pkg/help/cmd/export.go:62-64`.
- The flag is registered at `glazed/pkg/help/cmd/export.go:73-74`.
- SQLite export writes sections into a new store at `glazed/pkg/help/cmd/export.go:270-296`.

This confirms that `--from-sqlite-dir` should target the same store schema produced by `help export --format sqlite`.

Observed command evidence from this investigation:

```bash
pinocchio help export --format sqlite --output-path /tmp/glazed-help-exports/pinocchio.db
sqlite3 /tmp/glazed-help-exports/pinocchio.db "select count(*) from sections;"
# 69
```

The same experiment with `codebase-browser help export --format sqlite ...` failed because the installed `codebase-browser` help command does not currently expose an `export` subcommand. Its `help --help` output lists `--list`, `--query`, `--topics`, and similar flags, but no `export` verb. This is useful compatibility evidence: the multi-package loader should surface clear errors for command sources that lack `help export`, and the ticket should separately track whether `codebase-browser` needs to be rebuilt with a newer Glazed help command.

### Store layer

The help store is SQLite-backed. `store.New` opens a database and creates tables if needed at `glazed/pkg/help/store/store.go:25-37`.

The `sections` table is currently defined at `glazed/pkg/help/store/store.go:53-71`. Key fields are:

```sql
CREATE TABLE IF NOT EXISTS sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    section_type INTEGER NOT NULL,
    title TEXT NOT NULL,
    sub_title TEXT,
    short TEXT,
    content TEXT,
    topics TEXT,
    flags TEXT,
    commands TEXT,
    is_top_level BOOLEAN DEFAULT FALSE,
    is_template BOOLEAN DEFAULT FALSE,
    show_per_default BOOLEAN DEFAULT FALSE,
    order_num INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

The important constraint is `slug TEXT NOT NULL UNIQUE`. This is incompatible with multi-package serving when two packages have a section with the same slug, for example `configuration`, `getting-started`, or `index`.

### Model layer

A help section is represented by `model.Section` in `glazed/pkg/help/model/section.go:55-79`. It includes slug, type, title, content, topics, flags, commands, and display booleans. It does not include package or version metadata.

The intern should understand this as the core domain object. The requested feature needs either:

1. add package/version fields to `model.Section`, or
2. wrap `model.Section` in a serving-only envelope.

This guide recommends option 1 with conservative defaults because API responses, store rows, loaders, and tests already convert directly from `model.Section`.

### HTTP API layer

The server API is in `glazed/pkg/help/server/handlers.go` and `glazed/pkg/help/server/types.go`.

Current routes are registered at `glazed/pkg/help/server/handlers.go:62-65`:

```go
GET /api/health
GET /api/sections/search
GET /api/sections
GET /api/sections/{slug}
```

List behavior is implemented at `glazed/pkg/help/server/handlers.go:112-157`. It accepts filters such as section type, topic, command, flag, and search query, but not package or version.

Response types are in `glazed/pkg/help/server/types.go`:

- `SectionSummary` is defined at `types.go:40-48`.
- `ListSectionsResponse` is defined at `types.go:64-74`.
- `SectionDetail` is defined at `types.go:76-89`.

Neither summary nor detail includes package metadata.

### Frontend layer

The React app root is `glazed/web/src/App.tsx`. It holds local state for search and type filter at `App.tsx:15-18`, fetches all sections at `App.tsx:26`, filters client-side at `App.tsx:35-49`, and renders the sidebar at `App.tsx:53-69`.

Today the sidebar renders:

1. title bar,
2. search bar,
3. type filter,
4. section list,
5. status bar.

The requested screenshot inserts package and version selectors between search and type filter. The current API slice in `glazed/web/src/services/api.ts` only defines `healthCheck`, `listSections`, and `getSection`; see `api.ts:61-104`. It has no package list endpoint.

TypeScript response types are in `glazed/web/src/types/index.ts` and mirror the Go server response types. They also lack package/version fields.

## Target user experience

### Screenshot-derived layout

The attached screenshot shows the same retro two-pane layout, with these sidebar controls:

```text
┌──────────────────────────────┐
│ 📁 Sections                  │
├──────────────────────────────┤
│ 🔍 Search...                 │
│                              │
│ Package  [ Gepetto       v ] │
│ Version  [ v5.1.0        v ] │
│                              │
│ [All] [Topic] [Example] ...  │
├──────────────────────────────┤
│ Topic ◈ TOP                  │
│ Gepetto documentation index  │
│ Task-based index...          │
│                              │
│ Topic ◈ TOP                  │
│ Engine Profiles in Gepetto   │
│ ...                          │
│                              │
│ █ selected section █         │
└──────────────────────────────┘
```

The content title bar shows the selected title plus a command-style hint:

```text
📄 Events, Streaming, and Watermill in Gepetto — glaze help geppetto-events-streaming-watermill
```

### Required selection semantics

The UI should behave as follows:

- On first load, fetch package metadata.
- Select a default package:
  - Prefer `glazed` when embedded docs are present.
  - Otherwise select the first package sorted by display name.
- If the selected package has versions, show the version selector and choose a default version.
  - Prefer the latest semantic version if versions parse as semver.
  - Otherwise prefer lexicographically descending order.
- If the selected package has no versions, hide the version selector entirely, not just disabled, to match the requirement “otherwise don't”.
- Fetch/list sections for the selected package and selected version.
- Search and type filtering can remain client-side initially, but server-side package/version filters must be authoritative.

## Proposed architecture

### Core concept: package identity

Introduce package identity as metadata attached to each section.

Recommended fields:

```go
type Section struct {
    ID          int64       `json:"id,omitempty"`
    Slug        string      `json:"slug"`
    SectionType SectionType `json:"section_type"`

    // New source identity fields.
    PackageName    string `json:"package_name,omitempty"`
    PackageVersion string `json:"package_version,omitempty"`

    Title    string `json:"title"`
    SubTitle string `json:"sub_title"`
    Short    string `json:"short"`
    Content  string `json:"content"`
    // ... existing fields ...
}
```

Terminology:

- `PackageName`: stable package key, for example `glazed`, `pinocchio`, `geppetto`, or `codebase-browser`.
- `PackageVersion`: optional version key, for example `v5.1.0`. Empty string means “unversioned”.
- Display name can initially be derived from `PackageName` by replacing hyphens with spaces or title-casing in the frontend. A future metadata table can add explicit display names.

### Store schema change

Change the `sections` table from a globally unique slug to a package-scoped slug.

Recommended schema after migration:

```sql
CREATE TABLE IF NOT EXISTS sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name TEXT NOT NULL DEFAULT '',
    package_version TEXT NOT NULL DEFAULT '',
    slug TEXT NOT NULL,
    section_type INTEGER NOT NULL,
    title TEXT NOT NULL,
    sub_title TEXT,
    short TEXT,
    content TEXT,
    topics TEXT,
    flags TEXT,
    commands TEXT,
    is_top_level BOOLEAN DEFAULT FALSE,
    is_template BOOLEAN DEFAULT FALSE,
    show_per_default BOOLEAN DEFAULT FALSE,
    order_num INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(package_name, package_version, slug)
);

CREATE INDEX IF NOT EXISTS idx_sections_package ON sections(package_name);
CREATE INDEX IF NOT EXISTS idx_sections_package_version ON sections(package_name, package_version);
CREATE INDEX IF NOT EXISTS idx_sections_slug ON sections(slug);
```

Important migration rule:

- Existing databases without the new columns should still open.
- When `store.New` sees an old database, add columns with defaults:

```sql
ALTER TABLE sections ADD COLUMN package_name TEXT NOT NULL DEFAULT '';
ALTER TABLE sections ADD COLUMN package_version TEXT NOT NULL DEFAULT '';
```

SQLite cannot directly drop the old `UNIQUE(slug)` constraint created inline. For compatibility, the intern should check the current tests and decide whether old exported DBs can be read without modifying their uniqueness. For new destination stores used by `glaze serve`, the table creation should include the new composite unique constraint. For legacy source DBs, loading reads rows and assigns package metadata on import, so the legacy source constraint is not a problem.

### API shape

Add package metadata endpoints and extend existing section endpoints.

Recommended routes:

```text
GET /api/packages
GET /api/sections?package=<name>&version=<version>&type=<type>&q=<query>
GET /api/sections/{slug}?package=<name>&version=<version>
GET /api/health
```

Alternative detail route:

```text
GET /api/packages/{package}/versions/{version}/sections/{slug}
```

The query-parameter version is easier to add because it preserves the current `GET /api/sections/{slug}` route and needs fewer frontend routing changes. If no `package` is supplied and only one package exists, the server may default to that package. If multiple packages exist and no package is supplied, return a clear `400 bad_request` or select the default package consistently. This guide recommends returning all sections only for backward compatibility during Phase 1, then making the frontend always provide package filters.

Recommended response types:

```go
type PackageSummary struct {
    Name        string   `json:"name"`
    DisplayName string  `json:"displayName"`
    Versions    []string `json:"versions"`
    SectionCount int    `json:"sectionCount"`
}

type ListPackagesResponse struct {
    Packages []PackageSummary `json:"packages"`
    DefaultPackage string `json:"defaultPackage,omitempty"`
    DefaultVersion string `json:"defaultVersion,omitempty"`
}
```

Extend section responses:

```go
type SectionSummary struct {
    ID             int64    `json:"id"`
    PackageName    string   `json:"packageName"`
    PackageVersion string   `json:"packageVersion,omitempty"`
    Slug           string   `json:"slug"`
    Type           string   `json:"type"`
    Title          string   `json:"title"`
    Short          string   `json:"short"`
    Topics         []string `json:"topics"`
    IsTopLevel     bool     `json:"isTopLevel"`
}
```

### Loader metadata design

The current `ContentLoader.Load(ctx, hs)` method has no way to pass package metadata except by mutating sections before upsert. That is acceptable.

Introduce a small metadata type:

```go
type PackageRef struct {
    Name    string
    Version string
    Source  string
}
```

Add helper:

```go
func ApplyPackageRef(section *model.Section, ref PackageRef) {
    if section.PackageName == "" {
        section.PackageName = ref.Name
    }
    if section.PackageVersion == "" {
        section.PackageVersion = ref.Version
    }
}
```

Default package inference:

| Source | PackageName | PackageVersion |
| --- | --- | --- |
| embedded Glazed docs | `glazed` | empty unless build-time version is available |
| `--from-glazed-cmd pinocchio` | basename of binary, `pinocchio` | empty |
| `--from-sqlite-dir root/X/Y/X.db` | `X` | `Y` |
| `--from-sqlite-dir root/X/X.db` | `X` | empty |
| `--from-sqlite-dir root/X.db` | filename without `.db` or `.sqlite` | empty |
| `--from-sqlite path` | filename without extension, unless overridden later | empty |
| Markdown paths | last directory name or explicit future flag | empty |
| JSON exports | filename without extension | empty |

A future `--package` or `--source-package` override may be useful, but it is not required for this ticket.

## `--from-sqlite-dir` scanning rules

### Accepted extensions

Accept both `.db` and `.sqlite` because the existing export command permits both. `exportToSQLite` appends `.sqlite` only when the provided output path lacks `.sqlite` or `.db`; see `glazed/pkg/help/cmd/export.go:270-277`.

### Pattern rules

Given a discovered file path relative to a scan root, infer metadata in this order:

1. Versioned package DB: `X/Y/X.db` or `X/Y/X.sqlite`
   - Relative parts length is 3.
   - Filename stem equals first directory name.
   - Package is first directory name.
   - Version is second directory name.
2. Unversioned package DB in directory: `X/X.db` or `X/X.sqlite`
   - Relative parts length is 2.
   - Filename stem equals directory name.
   - Package is directory name.
   - Version is empty.
3. Root package DB: `X.db` or `X.sqlite`
   - Relative parts length is 1.
   - Package is filename stem.
   - Version is empty.

Question: what about `root/X/Y/help.db`? The user asked for `X/Y/X.db`, not arbitrary filenames. The strict rule prevents accidentally importing caches or unrelated SQLite files. If users want arbitrary files, they can use `--from-sqlite` explicitly.

### Pseudocode

```go
func DiscoverSQLitePackages(root string) ([]DiscoveredSQLitePackage, error) {
    var out []DiscoveredSQLitePackage

    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        if !isHelpSQLiteExtension(path) { return nil }

        rel, err := filepath.Rel(root, path)
        if err != nil { return err }
        parts := splitClean(rel)
        stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

        switch len(parts) {
        case 1:
            out = append(out, DiscoveredSQLitePackage{Path: path, Package: stem})
        case 2:
            pkg := parts[0]
            if stem == pkg {
                out = append(out, DiscoveredSQLitePackage{Path: path, Package: pkg})
            }
        case 3:
            pkg := parts[0]
            version := parts[1]
            if stem == pkg {
                out = append(out, DiscoveredSQLitePackage{Path: path, Package: pkg, Version: version})
            }
        }
        return nil
    })
    if err != nil { return nil, err }

    sort.Slice(out, func(i, j int) bool {
        return out[i].Package < out[j].Package ||
            (out[i].Package == out[j].Package && out[i].Version < out[j].Version)
    })
    return out, nil
}
```

### Diagram: ingestion path

```text
CLI flags
  |
  |-- embedded help system (already loaded)
  |       package=glazed, version=""
  |
  |-- --from-glazed-cmd pinocchio
  |       run: pinocchio help export --with-content=true --output json
  |       package=pinocchio, version=""
  |
  |-- --from-sqlite-dir ./helpdbs
  |       discover root/geppetto/v5.1.0/geppetto.db
  |       package=geppetto, version=v5.1.0
  |
  v
loader applies PackageRef to each model.Section
  |
  v
store.Upsert with UNIQUE(package_name, package_version, slug)
  |
  v
HTTP API filters by package/version
  |
  v
React sidebar package/version selectors
```

## Detailed implementation plan

### Phase 1: Add package fields to model and server response types

Files:

- `glazed/pkg/help/model/section.go`
- `glazed/pkg/help/server/types.go`
- `glazed/web/src/types/index.ts`

Steps:

1. Add `PackageName` and `PackageVersion` to `model.Section`.
2. Update `SectionSummary` and `SectionDetail` Go response types.
3. Update `SummaryFromModel` and `DetailFromModel`.
4. Update TypeScript interfaces with `packageName` and `packageVersion`.
5. Keep JSON fields optional enough to avoid breaking static-mode tests.

Suggested Go fields:

```go
PackageName    string `json:"package_name,omitempty"`
PackageVersion string `json:"package_version,omitempty"`
```

Suggested API JSON tags:

```go
PackageName    string `json:"packageName"`
PackageVersion string `json:"packageVersion,omitempty"`
```

### Phase 2: Migrate store schema and query methods

Files:

- `glazed/pkg/help/store/store.go`
- `glazed/pkg/help/store/query.go`
- `glazed/pkg/help/store/query_fts5.go`
- `glazed/pkg/help/store/query_nofts.go`
- `glazed/pkg/help/store/store_test.go`
- `glazed/pkg/help/store/query_test.go`

Steps:

1. Add columns to new table creation.
2. Add an `ensurePackageColumns` migration for existing DBs.
3. Change `Upsert` SQL to insert/update package fields.
4. Change row scans to read package fields.
5. Add predicates:

```go
func InPackage(name string) Predicate
func InPackageVersion(name, version string) Predicate
```

6. Add a store method for package summaries:

```go
func (s *Store) ListPackages(ctx context.Context) ([]PackageInfo, error)
```

7. Preserve backwards compatibility:
   - Sections from old DBs get `PackageName == ""` when read.
   - Loaders should fill package before upserting into the serving store.

Testing:

- Create two sections with same slug and different packages; both must persist.
- Fetch by `(package, version, slug)` and confirm correct content.
- List package summaries and version arrays.

### Phase 3: Add package-aware loaders

Files:

- `glazed/pkg/help/loader/sources.go`
- new optional file `glazed/pkg/help/loader/sqlite_dir.go`
- `glazed/pkg/help/loader/sources_test.go`

Steps:

1. Define `PackageRef` and `DiscoveredSQLitePackage`.
2. Add `SQLiteDirLoader`:

```go
type SQLiteDirLoader struct {
    Roots []string
}
```

3. Implement discovery rules exactly as described above.
4. Refactor `SQLiteLoader` so it can load with a package ref:

```go
func loadSQLitePath(ctx context.Context, hs *help.HelpSystem, path string, ref PackageRef) error
```

5. Update `GlazedCommandLoader` to set package name from `filepath.Base(binary)`.
6. Update JSON and Markdown loaders to infer package name from source path, or leave empty and let serve defaults apply. Prefer setting it to avoid unowned sections.
7. Add collision logging that includes package/version/slug, not only slug.

Testing:

- Use temporary directories with these files:

```text
root/glazed.db
root/pinocchio/pinocchio.db
root/geppetto/v5.1.0/geppetto.db
root/geppetto/v5.0.0/geppetto.db
root/unrelated/cache.db       # ignored unless cache/cache.db pattern matches
```

- Assert discoveries:

```text
glazed, ""
pinocchio, ""
geppetto, v5.1.0
geppetto, v5.0.0
```

### Phase 4: Add `--from-sqlite-dir` to `glaze serve`

Files:

- `glazed/pkg/help/server/serve.go`
- `glazed/pkg/help/server/serve_test.go`

Steps:

1. Add setting:

```go
FromSQLiteDir []string `glazed:"from-sqlite-dir"`
```

2. Register flag:

```go
fields.New(
    "from-sqlite-dir",
    fields.TypeStringList,
    fields.WithHelp("Directories to scan recursively for package/versioned SQLite help exports"),
)
```

3. Update long help with examples:

```bash
glaze serve --from-sqlite-dir ./help-dbs
glaze serve --with-embedded --from-glazed-cmd pinocchio --from-sqlite-dir ./releases
```

4. Add loader in `buildServeLoaders`.
5. Ensure embedded docs get package metadata. Because embedded sections are already loaded before `serve` runs, add a pre-serve normalization step:

```go
if s.WithEmbedded || len(loaders) == 0 {
    ensurePackageForExistingSections(ctx, hs, "glazed", "")
}
```

### Phase 5: Add API package endpoints and filters

Files:

- `glazed/pkg/help/server/handlers.go`
- `glazed/pkg/help/server/types.go`
- `glazed/pkg/help/server/server_test.go`

Steps:

1. Register route:

```go
mux.HandleFunc("GET /api/packages", h.handleListPackages)
```

2. Extend `ListSectionsParams` with:

```go
PackageName string `json:"package,omitempty"`
Version     string `json:"version,omitempty"`
```

3. Parse query params `package` and `version`.
4. Extend `buildPredicate` to add package filters.
5. Update `handleGetSection` to read package/version query params and call a package-aware store getter. If no package is passed, preserve old behavior only when unambiguous.

Suggested ambiguity behavior:

```go
matches, err := h.deps.Store.Find(ctx, store.SlugIs(slug))
if len(matches) == 1 { return matches[0] }
if len(matches) > 1 { return 400 "package is required for duplicate slug" }
return 404
```

### Phase 6: Update the React UI

Files:

- `glazed/web/src/App.tsx`
- `glazed/web/src/services/api.ts`
- `glazed/web/src/types/index.ts`
- new component `glazed/web/src/components/PackageSelector/PackageSelector.tsx`
- new styles `glazed/web/src/components/PackageSelector/styles/package-selector.css`
- tests in `glazed/web/src/App.test.tsx` and API tests

Steps:

1. Add API endpoint:

```ts
listPackages: builder.query<ListPackagesResponse, void>({
  query: () => ({ url: isStaticMode ? '/packages.json' : '/packages' }),
})
```

2. Add local selected package/version state.
3. Initialize state after packages load.
4. Call `useListSectionsQuery({ packageName, version, q })` or serialize params safely.
5. Update `useGetSectionQuery` to include package/version.
6. Render package selector below search:

```tsx
<PackageSelector
  packages={packages}
  selectedPackage={selectedPackage}
  selectedVersion={selectedVersion}
  onPackageChange={setSelectedPackage}
  onVersionChange={setSelectedVersion}
/>
```

7. Hide the version row when the selected package has no versions.
8. Preserve type filter buttons as shown in the screenshot.

### Phase 7: Manual validation with real exports

Commands:

```bash
mkdir -p /tmp/glazed-help-exports/pinocchio/v0
pinocchio help export --format sqlite --output-path /tmp/glazed-help-exports/pinocchio/v0/pinocchio.db

glaze help export --format sqlite --output-path /tmp/glazed-help-exports/glazed.db

go run ./cmd/glaze serve --with-embedded --from-sqlite-dir /tmp/glazed-help-exports
```

Then open `http://localhost:8088` and verify:

- package dropdown contains `glazed` and `pinocchio`.
- selecting `pinocchio` shows only Pinocchio sections.
- selecting `glazed` shows Glazed sections.
- version dropdown appears for `pinocchio` if using `pinocchio/v0/pinocchio.db`.
- version dropdown does not appear for root `glazed.db`.

## Risks and tradeoffs

### Risk: SQLite uniqueness migration complexity

The current inline `UNIQUE` constraint on `slug` is the hardest database compatibility issue. If old databases need in-place writes with duplicate slugs, migration requires table rebuild. However, source export databases are mostly read-only and serving stores are often in-memory or newly created, so the first implementation can keep migration conservative.

### Risk: URL compatibility

Current frontend URLs use `/sections/:slug`. A slug alone is no longer globally unique. The simplest transitional approach is to keep that path but add query state or internal RTK Query params for package/version. Later, browser URLs should include package/version for shareable links:

```text
/packages/pinocchio/sections/autosave
/packages/geppetto/versions/v5.1.0/sections/events-streaming-watermill
```

### Risk: command export compatibility

`pinocchio help export --format sqlite` works in this environment and produced 69 sections. `codebase-browser help export --format sqlite` did not work because the installed command lacks the export subcommand. The loader should keep good error messages for this case, and a separate task should update/rebuild `codebase-browser` if it is expected to be loadable by command.

### Tradeoff: package fields on Section vs wrapper type

Adding fields to `model.Section` touches more code, but it is easier for this codebase because store, server, export, and frontend already use `Section` as the central data shape. A wrapper type would keep the model pure but would force serving-only adapter code everywhere and make store queries more awkward.

## Test checklist

Backend unit tests:

- `SQLiteDirLoader` discovers all three requested patterns.
- `SQLiteDirLoader` ignores arbitrary DBs that do not match requested patterns.
- `GlazedCommandLoader` assigns package name from binary basename.
- Store accepts same slug in different packages.
- Store rejects duplicate `(package, version, slug)` or upserts deterministically.
- `GET /api/packages` returns packages and versions sorted predictably.
- `GET /api/sections?package=X&version=Y` filters correctly.
- `GET /api/sections/{slug}?package=X&version=Y` returns the correct duplicate slug.

Frontend tests:

- Package selector renders when multiple packages exist.
- Version selector renders only for packages with versions.
- Changing package resets version to that package's default.
- Section list updates when selected package changes.
- Existing search and type filters still work.

Manual tests:

- Export `pinocchio` to SQLite and serve from directory.
- Export `glazed` to SQLite and serve alongside embedded docs.
- Try a broken command source and verify the error explains `help export` failed.

## File reference map for the intern

Start here:

- `glazed/pkg/help/server/serve.go` — CLI flag registration, source loader assembly, HTTP server startup.
- `glazed/pkg/help/loader/sources.go` — JSON, SQLite, and command-source import logic.
- `glazed/pkg/help/store/store.go` — SQLite schema and persistence.
- `glazed/pkg/help/model/section.go` — section domain model.
- `glazed/pkg/help/server/handlers.go` — API route registration and list/detail handlers.
- `glazed/pkg/help/server/types.go` — API response contracts.
- `glazed/web/src/App.tsx` — top-level UI wiring and sidebar layout.
- `glazed/web/src/services/api.ts` — RTK Query endpoints.
- `glazed/web/src/types/index.ts` — TypeScript response contracts.
- `glazed/pkg/help/cmd/export.go` — how SQLite help exports are produced.

## Open questions

1. Should package display names be explicit metadata or derived from package keys?
2. Should embedded Glazed docs expose the binary build version as a package version?
3. Should `--from-sqlite` accept package/version annotations, for example `pinocchio:v0:/path/pinocchio.db`?
4. Should frontend routes become package-aware in the first implementation, or is internal package state enough?
5. Should `codebase-browser` be updated to include the newer `help export` verb, or should it be treated only as a future command-source candidate?

## References

- Screenshot: `/tmp/pi-clipboard-a7fb96dd-2edd-449d-8be1-753df53c7e3a.png`
- `glazed/pkg/help/server/serve.go:36-44` — current serve settings.
- `glazed/pkg/help/server/serve.go:81-100` — current serve source flags.
- `glazed/pkg/help/server/serve.go:123-151` — current source loading and server startup.
- `glazed/pkg/help/server/serve.go:154-168` — current loader assembly.
- `glazed/pkg/help/loader/sources.go:104-132` — SQLite loader.
- `glazed/pkg/help/loader/sources.go:138-170` — command loader.
- `glazed/pkg/help/store/store.go:53-71` — current sections table.
- `glazed/pkg/help/model/section.go:55-79` — current section model.
- `glazed/pkg/help/server/handlers.go:62-65` — current API routes.
- `glazed/pkg/help/server/types.go:40-89` — current API response types.
- `glazed/web/src/App.tsx:15-69` — current UI state and sidebar.
- `glazed/web/src/services/api.ts:61-104` — current RTK Query endpoints.
- `glazed/pkg/help/cmd/export.go:62-74` and `270-296` — SQLite export behavior.
