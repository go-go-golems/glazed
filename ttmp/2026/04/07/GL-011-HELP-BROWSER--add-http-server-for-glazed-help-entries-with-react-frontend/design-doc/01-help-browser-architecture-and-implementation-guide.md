---
Title: Help Browser Architecture and Implementation Guide
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
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/help.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/model/section.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/store.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/query.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/cmd/cobra.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/cmd/glaze/main.go
    - /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/ttmp/2026/04/07/GL-011-HELP-BROWSER--add-http-server-for-glazed-help-entries-with-react-frontend/sources/local/glazed-docs-browser(2).jsx
ExternalSources: []
Summary: "Comprehensive design and implementation guide for adding an HTTP server and React frontend to browse Glazed help entries, served as a single Go binary."
LastUpdated: 2026-04-07T20:26:06.134180502-04:00
WhatFor: "Guide an intern through building the glaze help browser end-to-end"
WhenToUse: "Reference when implementing the help browser server, React frontend, or Dagger build pipeline"
---

# Help Browser Architecture and Implementation Guide

## Executive Summary

This document describes how to add an HTTP server and a modern React frontend to the
Glazed CLI tool so that users can browse help documentation in a web browser. Today,
Glazed stores its help entries as Markdown files with YAML frontmatter, loads them into
an SQLite-backed store at startup, and renders them in the terminal using Charm
(bubbletea/glamour). The goal is to keep all of that working, but also let the user run
`glaze serve file1 file2 dir1 dir2...` to discover Glazed help files from the given
paths, serve them over HTTP, and present them in a beautiful, searchable, modular React
SPA.

A monolithic JSX prototype (`glazed-docs-browser(2).jsx`) already exists and demonstrates
the desired look-and-feel. This guide decomposes that prototype into a proper
**Vite + React + RTK Query + Storybook** frontend in a `web/` directory, with
theming support (CSS variables, `data-part` selectors), and wires a **Dagger-powered
`go:generate` builder** so that `go generate ./cmd/help-browser` produces the static
assets that get embedded into the Go binary via `go:embed`. The result is a single
`glaze` binary that can serve its own documentation.

The document is written so that a new intern can understand every layer of the system --
from the Go help-system internals to the React component tree to the Dagger build
pipeline -- and implement each piece step by step.

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Glossary](#glossary)
3. [System Overview](#system-overview)
4. [Existing Architecture](#existing-architecture)
5. [Proposed Architecture](#proposed-architecture)
6. [Go Backend Design](#go-backend-design)
7. [REST API Specification](#rest-api-specification)
8. [React Frontend Design](#react-frontend-design)
9. [React Component Decomposition](#react-component-decomposition)
10. [RTK Query Integration](#rtk-query-integration)
11. [Theming System](#theming-system)
12. [Build Pipeline: Dagger + go:generate + go:embed](#build-pipeline)
13. [Storybook Integration](#storybook-integration)
14. [File Layout](#file-layout)
15. [Implementation Plan (Phased)](#implementation-plan)
16. [Testing Strategy](#testing-strategy)
17. [Risks and Open Questions](#risks-and-open-questions)
18. [References](#references)

---

## Problem Statement

Glazed already has a rich help system backed by SQLite, with Markdown files that have
YAML frontmatter (title, slug, section type, topics, commands, flags). Users interact
with it only through the terminal -- either via `glaze help <slug>` or through the
Charm TUI (`glaze help-ui`). There is no way to browse documentation in a web browser,
which limits:

- **Discoverability**: Users must know the exact slug or use the CLI query DSL.
- **Presentation**: Terminal output cannot render tables, syntax-highlighted code blocks,
  or interactive navigation as richly as a browser can.
- **Sharing**: You cannot send a colleague a link to a help page.
- **Onboarding**: New users exploring Glazed for the first time have no friendly graphical
  interface to click through.

The goal is to add a `glaze serve` subcommand that starts an HTTP server, exposes a
REST API over the existing `HelpSystem` / `Store`, and serves a React SPA that lets
users search, filter, and read help entries interactively -- all compiled into a single
Go binary.

---

## Glossary

This section defines every term and concept you will encounter in this document. If you
are new to the project, read this section first.

| Term | Definition |
|------|------------|
| **Glazed** | A Go CLI framework for formatting structured data. The main binary is `glaze`. |
| **Help System** | The Go package `pkg/help/` that loads, stores, queries, and renders documentation sections. |
| **Section** | A single help document with metadata (slug, type, topics, commands, flags) and Markdown content. |
| **Section Type** | One of: `GeneralTopic`, `Example`, `Application`, `Tutorial`. |
| **Slug** | A unique, URL-friendly identifier for a section (e.g., `help-system`). |
| **Store** | The SQLite-backed persistence layer (`pkg/help/store/`) that CRUDs sections. |
| **Query DSL** | A boolean query language for filtering sections (e.g., `type:example AND topic:database`). |
| **HelpSystem** | The top-level Go struct that owns a `Store` and provides high-level operations like `LoadSectionsFromFS()`, `AddSection()`, `GetSectionWithSlug()`. |
| **Frontmatter** | YAML metadata at the top of a Markdown file, delimited by `---`. |
| **go:embed** | A Go compiler directive that embeds files from the source tree into the compiled binary at build time. |
| **go:generate** | A Go compiler directive that runs a command before compilation (e.g., to build frontend assets). |
| **Dagger** | A programmable CI/CD toolkit that runs build steps in containers. We use it so that `go generate` can build the React app without requiring Node.js on the host. |
| **Vite** | A fast JavaScript/TypeScript build tool and dev server for modern web projects. |
| **React** | A JavaScript library for building user interfaces with components. |
| **RTK Query** | A data-fetching and caching library, part of Redux Toolkit. Auto-generates React hooks for API endpoints. |
| **Redux Toolkit (RTK)** | The official, opinionated toolset for efficient Redux development. |
| **Storybook** | A tool for developing UI components in isolation with hot-reload and interactive controls. |
| **SPA** | Single Page Application -- the browser loads one HTML file, and JavaScript handles all navigation. |
| **CSS Variable / Token** | A named value (e.g., `--color-bg: #fff`) that can be overridden to change the theme. |
| **data-part** | An HTML attribute convention (e.g., `data-part="sidebar"`) used as a stable styling hook instead of CSS classes. |
| **pnpm** | A fast, disk-efficient JavaScript package manager (alternative to npm/yarn). |
| **Cobra** | A Go library for building CLI applications with commands, flags, and help. Glazed uses it extensively. |

---

## System Overview

The help browser has three main layers. The diagram below shows how data flows from Markdown files on disk all the way to the user's browser:

```
|---------------------------|---- BROWSER ----|---------------------------|
| SearchBar | SectionList | SectionView | TypeFilter |          |
|          | (sidebar)  | (main)     | (badges)   |          |
|-------------------------------| RTK Query |-------------------------------|
                                    | HTTP (JSON)
                                    v
|---------------------------| GO HTTP SERVER (:8088) |-------------------|
| GET /api/sections | GET /api/sections/:slug | GET /api/sections/search |
|                          HelpSystem (pkg/help/help.go)                  |
| - LoadSectionsFromFS() - reads .md files                              |
| - Store - SQLite-backed CRUD + full-text search                     |
| - SectionQuery - boolean query DSL                                  |
|-------------------| go:embed static |  Markdown files on disk |-------|
|                    (built by Dagger/Vite)  (discovered from args) |
|-----------------------------------------------------------------------|
```

The user invokes `glaze serve file1 file2 dir1 dir2...`. The Go binary:

1. **Discovers** all `.md` files from the given paths (recursing into directories).
2. **Parses** YAML frontmatter + Markdown body into `Section` structs.
3. **Loads** them into the in-memory SQLite `Store`.
4. **Starts** an HTTP server on `:8088` (configurable).
5. **Serves** the React SPA for `/` (from `go:embed`).
6. **Serves** REST JSON endpoints under `/api/*`.

---

## Existing Architecture

Before we build anything new, you need to understand what already exists. This section
walks through every relevant package and file.

### The Help System (`pkg/help/`)

The help system is the heart of Glazed's documentation infrastructure. It lives in
`pkg/help/` and is organized into several sub-packages:

#### `pkg/help/help.go` -- Top-level API

This file defines two key types:

- **`Section`** -- Wraps `model.Section` and adds a back-reference to the `HelpSystem`. Provides convenience methods like `IsForCommand()`, `IsForTopic()`, `DefaultGeneralTopic()`, `DefaultExamples()`.
- **`HelpSystem`** -- The top-level orchestrator. Owns a `Store`. Key methods:
  - `NewHelpSystem()` -- creates an in-memory SQLite store.
  - `LoadSectionsFromFS(f fs.FS, dir string)` -- recursively reads `.md` files from an `embed.FS` or any `fs.FS`, parses frontmatter, and upserts into the store.
  - `AddSection(section *Section)` -- upserts a section into the store.
  - `GetSectionWithSlug(slug string)` -- retrieves a section by its unique slug.

#### `pkg/help/model/section.go` -- Data model

The `Section` struct is the core data type. It has JSON tags, which means it serializes
directly to the API response shape we want:

```go
type Section struct {
    ID          int64       `json:"id,omitempty"`
    Slug        string      `json:"slug"`
    SectionType SectionType `json:"section_type"`
    Title       string      `json:"title"`
    SubTitle    string      `json:"sub_title"`
    Short       string      `json:"short"`
    Content     string      `json:"content"`
    Topics      []string    `json:"topics"`
    Flags       []string    `json:"flags"`
    Commands    []string    `json:"commands"`
    IsTopLevel     bool `json:"is_top_level"`
    IsTemplate     bool `json:"is_template"`
    ShowPerDefault bool `json:"show_per_default"`
    Order          int  `json:"order"`
}
```

**`SectionType`** is an enum: `GeneralTopic=0`, `Example=1`, `Application=2`, `Tutorial=3`.

#### `pkg/help/store/` -- SQLite persistence

The `Store` struct wraps `database/sql` with SQLite (`github.com/mattn/go-sqlite3`). Key methods:

- `New(dbPath string)` / `NewInMemory()` -- constructor.
- `Insert(ctx, section)` / `Update(ctx, section)` / `Upsert(ctx, section)` / `Delete(ctx, id)` -- CRUD.
- `GetBySlug(ctx, slug)` -- fetch by unique slug.
- `GetByID(ctx, id)` -- fetch by primary key.
- `List(ctx, orderBy)` -- list all sections.
- `Count(ctx)` -- total sections.
- `Query(ctx, predicates...)` -- advanced filtering via `QueryCompiler`.

The store also supports **full-text search** via SQLite FTS5 (with a `nofts` fallback).

#### `pkg/help/query.go` -- Query builder

`SectionQuery` is a fluent builder for filtering sections:

```go
results, err := query.FindSections(ctx, helpSystem.Store)
```

#### `pkg/help/dsl/` -- Query DSL parser

A boolean query DSL for advanced search:

```
type:example AND topic:database
type:example OR type:tutorial
"full text search phrase"
```

This is exposed to users via `glaze help --query "..."` and will be available via the
REST API's search endpoint.

#### `pkg/help/cmd/cobra.go` -- CLI integration

`SetupCobraRootCommand(helpSystem, rootCmd)` wires the help system into Cobra's
help system. It creates a `help` subcommand and overrides Cobra's default help
function to display Glazed sections alongside standard Cobra help.

#### `pkg/help/ui/` -- Terminal TUI

A Charm bubbletea-based TUI for browsing help entries in the terminal. This is the
closest existing analogue to what we're building for the web -- it has the same
list-detail navigation pattern.

### Main Entry Point (`cmd/glaze/main.go`)

The `main()` function:

1. Creates a `HelpSystem`.
2. Loads embedded docs via `doc.AddDocToHelpSystem(helpSystem)`.
3. Wires Cobra commands.
4. Executes the root command.

We will add a new `serve` subcommand here.

---

## Proposed Architecture

The proposed architecture adds two new subsystems to the existing Glazed codebase:

1. **A Go HTTP server** that exposes the existing `HelpSystem` via REST endpoints and serves
   the React SPA's static files from an embedded filesystem.
2. **A React frontend** in `web/` that provides a browser-based documentation browser with
   search, filtering, and rich Markdown rendering.
3. **A Dagger build pipeline** that compiles the React frontend inside a container and
   outputs `dist/` for `go:embed` consumption.

### Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP framework | Go `net/http` + `http.ServeMux` | Zero dependencies, sufficient for our needs, matches the reference guide pattern |
| Frontend framework | React 18 + TypeScript | Required by spec, strong ecosystem |
| State management | Redux Toolkit + RTK Query | Auto-caching, auto-refetching, generated hooks, standard pattern |
| Build tool | Vite 5 | Fast HMR, simple config, standard for React projects |
| CSS strategy | CSS variables + data-part selectors | Stable theming API, no class-name coupling |
| Component dev | Storybook 8 | Isolated development, visual regression baseline |
| Frontend builder | Dagger (Go SDK) | Runs in container, no Node.js required on host |
| Static embedding | go:embed | Single binary deployment, zero external dependencies |
| CLI integration | Cobra subcommand `glaze serve` | Consistent with existing Glazed CLI patterns |
| Markdown rendering | react-markdown + remark-gfm | Rich rendering (tables, code blocks, etc.) in the browser |

### Alternatives Considered

- **Gin / Echo / Chi for HTTP**: These frameworks add complexity for little benefit since
  our API surface is small (3-4 endpoints). Standard `net/http` is sufficient and keeps
  the dependency tree small.
- **Next.js / SSR**: Overkill for a documentation browser that reads from a local Go API.
  A pure SPA is simpler and can be fully embedded.
- **Webpack instead of Vite**: Vite is faster and has simpler configuration. The
  reference guide (`how-to-create-a-web-app-with-react-rtk-vite-dagger-gen.md`) already
  uses Vite, so we follow that pattern.
- **CSS Modules / Styled Components**: CSS variables + `data-part` selectors provide a
  more stable, themeable API that doesn't couple consumers to class names or JS-in-CSS.
- **Makefile for building frontend**: Dagger is preferred because it runs in a container
  and doesn't require Node.js/pnpm to be installed on the host machine.

---

## Go Backend Design

The backend is a straightforward HTTP server that wraps the existing `HelpSystem`.
It discovers help files from user-supplied paths, loads them, and exposes REST endpoints.

### File discovery from arguments

The `serve` command accepts positional arguments that are files or directories.
For each argument:

- If it's a file ending in `.md`, parse and load it.
- If it's a directory, walk it recursively and load all `.md` files.
- If no arguments are given, fall back to the embedded docs (the existing `doc.AddDocToHelpSystem` behavior).

Pseudocode for the discovery logic:

```go
func loadSectionsFromArgs(hs *help.HelpSystem, args []string) error {
    if len(args) == 0 {
        // Use embedded docs
        return doc.AddDocToHelpSystem(hs)
    }
    for _, arg := range args {
        info, err := os.Stat(arg)
        if err != nil {
            return fmt.Errorf("cannot access %s: %w", arg, err)
        }
        if info.IsDir() {
            // Walk directory, load .md files
            err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
                if err != nil { return err }
                if d.IsDir() || !strings.HasSuffix(path, ".md") { return nil }
                return loadSingleFile(hs, path)
            })
        } else {
            err = loadSingleFile(hs, arg)
        }
        if err != nil {
            return err
        }
    }
    return nil
}

func loadSingleFile(hs *help.HelpSystem, path string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    section, err := help.LoadSectionFromMarkdown(b)
    if err != nil { return fmt.Errorf("parse %s: %w", path, err) }
    hs.AddSection(section)
    return nil
}
```

### Server structure

The server is defined in a new file `cmd/help-browser/main.go` (or wired as a Cobra
subcommand in `cmd/glaze/main.go` -- see the Implementation Plan for the decision).

```go
package main

import (
    "embed"
    "io/fs"
    "log"
    "net/http"

    "github.com/go-go-golems/glazed/pkg/help"
)

//go:embed dist
var staticFS embed.FS

type Server struct {
    helpSystem *help.HelpSystem
    addr       string
    mux        *http.ServeMux
}

func NewServer(hs *help.HelpSystem, addr string) *Server {
    s := &Server{
        helpSystem: hs,
        addr:       addr,
        mux:        http.NewServeMux(),
    }
    s.registerRoutes()
    return s
}

func (s *Server) registerRoutes() {
    // API routes
    s.mux.HandleFunc("/api/sections", s.handleListSections)
    s.mux.HandleFunc("/api/sections/", s.handleGetSection)   // /api/sections/{slug}
    s.mux.HandleFunc("/api/sections/search", s.handleSearch)  // /api/sections/search?q=...
    s.mux.HandleFunc("/api/health", s.handleHealth)

    // Static files (SPA)
    distFS, _ := fs.Sub(staticFS, "dist")
    s.mux.Handle("/", http.FileServer(http.FS(distFS)))
}

func (s *Server) ListenAndServe() error {
    log.Printf("Serving help browser on %s", s.addr)
    return http.ListenAndServe(s.addr, s.mux)
}
```

### SPA fallback handling

Since this is a Single Page Application, all non-API, non-static-asset routes must
fall through to `index.html` so that client-side routing works. The standard
`http.FileServer` handles this if we use `react-router` with `HashRouter`, or we can
add a custom middleware that serves `index.html` for any path that doesn't match
a static file:

```go
// spaHandler implements http.Handler for serving a React SPA
type spaHandler struct {
    staticFS   fs.FS
    indexHTML  []byte
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Try to serve the static file
    path := r.URL.Path
    if path == "/" { path = "/index.html" }
    
    // If the file exists, serve it
    f, err := h.staticFS.Open(path[1:]) // strip leading /
    if err == nil {
        f.Close()
        http.FileServer(http.FS(h.staticFS)).ServeHTTP(w, r)
        return
    }
    
    // Otherwise, serve index.html (SPA fallback)
    w.Header().Set("Content-Type", "text/html")
    w.Write(h.indexHTML)
}
```

---

## REST API Specification

The API is intentionally small. Four endpoints cover all the operations the React
frontend needs.

### `GET /api/health`

Health check endpoint for monitoring.

**Response:**

```json
{
  "ok": true,
  "sections": 42
}
```

**Status codes:** `200` on success.

---

### `GET /api/sections`

List all sections (summary view, no content). Returns enough data for the sidebar
list, type filter, and search.

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `type` | string | Filter by section type: `GeneralTopic`, `Example`, `Application`, `Tutorial` |
| `topic` | string | Filter by topic tag |
| `command` | string | Filter by associated command |
| `toplevel` | bool | If `true`, only return top-level sections |
| `q` | string | Full-text search across title, short, topics, and slug |

**Response:**

```json
{
  "sections": [
    {
      "slug": "help-system",
      "section_type": "GeneralTopic",
      "title": "Help System",
      "short": "Glazed provides a powerful, queryable help system...",
      "topics": ["help", "documentation", "cli"],
      "commands": ["help"],
      "flags": ["flag", "topic", "command"],
      "is_top_level": true,
      "show_per_default": true,
      "order": 0
    }
  ],
  "total": 42
}
```

**Implementation notes:**

- The `section_type` field uses the **string** representation (`"GeneralTopic"`), not the
  integer, for readability. The Go handler converts using `model.SectionType.String()`.
- The `content` field is **omitted** in list responses to keep the payload small.
- The `q` parameter performs a case-insensitive search across `title`, `short`, `topics`,
  and `slug` fields. If FTS5 is available, it uses full-text search; otherwise it falls
  back to `LIKE` queries.

**Handler pseudocode:**

```go
func (s *Server) handleListSections(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    query := r.URL.Query()

    // Build filter from query params
    sectionType := query.Get("type")
    topic := query.Get("topic")
    searchQuery := query.Get("q")

    // Use the store's query system
    sq := help.NewSectionQuery().ReturnAllTypes()
    
    if sectionType != "" {
        st, _ := model.SectionTypeFromString(sectionType)
        sq.Types[st] = true
    }
    if topic != "" {
        sq.Topics = []string{topic}
    }

    // Get all matching sections
    all, err := s.helpSystem.Store.List(ctx, "title ASC")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Filter in memory for simple cases, or use QueryCompiler for complex ones
    sections := filterSections(all, sectionType, topic, searchQuery)

    // Serialize without content
    summaries := make([]SectionSummary, len(sections))
    for i, sec := range sections {
        summaries[i] = toSummary(sec)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ListResponse{
        Sections: summaries,
        Total:    len(summaries),
    })
}
```

---

### `GET /api/sections/{slug}`

Get a single section by slug, including its full Markdown content.

**Response:**

```json
{
  "slug": "help-system",
  "section_type": "GeneralTopic",
  "title": "Help System",
  "short": "Glazed provides a powerful, queryable help system...",
  "content": "## Overview

The Glazed help system provides...",
  "topics": ["help", "documentation", "cli"],
  "commands": ["help"],
  "flags": ["flag", "topic", "command"],
  "is_top_level": true,
  "show_per_default": true,
  "related": [
    {
      "slug": "writing-help-entries",
      "title": "Writing Help Entries",
      "section_type": "GeneralTopic"
    }
  ]
}
```

**Status codes:** `200` on success, `404` if slug not found.

**Implementation notes:**

- The `related` field is computed by looking at sections that share topics or commands
  with the current section. This is the web equivalent of the `See Also` links.
- The `content` field is the raw Markdown body. The React frontend renders it to HTML
  using `react-markdown`.

---

### `GET /api/sections/search?q={query}`

Advanced search using the Glazed query DSL.

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Query DSL expression (e.g., `type:example AND topic:database`) |

**Response:** Same shape as `GET /api/sections` (list of summaries).

**Implementation:** Uses the existing DSL parser and `QueryCompiler`:

```go
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    if q == "" {
        // Redirect to list
        s.handleListSections(w, r)
        return
    }
    
    // Use the DSL parser
    results, err := s.helpSystem.QuerySections(q)
    if err != nil {
        http.Error(w, fmt.Sprintf("invalid query: %v", err), 400)
        return
    }
    
    summaries := make([]SectionSummary, len(results))
    for i, sec := range results {
        summaries[i] = toSummary(sec.Section)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ListResponse{
        Sections: summaries,
        Total:    len(summaries),
    })
}
```

---

### CORS configuration

During development, the Vite dev server runs on `:5173` and proxies API calls to the Go
server on `:8088`. In production, everything is same-origin. For development, we enable
CORS on the Go server:

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" {
            w.WriteHeader(204)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## React Frontend Design

The React frontend is a Single Page Application built with Vite and TypeScript. It
communicates with the Go backend exclusively through the REST API, using RTK Query
for data fetching and caching. The UI is a two-pane layout: a sidebar with a searchable,
filterable list of sections, and a main content area that renders the selected section's
Markdown content.

### Why React + RTK Query + Vite?

These three technologies work together as follows:

- **React** renders the UI as a tree of components. Each component is a function that
takes props (inputs) and returns JSX (HTML-like syntax). React efficiently updates the
DOM when data changes.

- **RTK Query** (part of Redux Toolkit) manages all server communication. You define
your API endpoints once, and RTK Query auto-generates React hooks like `useListSectionsQuery()`
and `useGetSectionQuery(slug)`. It handles caching, loading states, error states, and
automatic refetching.

- **Vite** is the build tool. It provides a fast dev server with Hot Module Replacement
(HMR) and produces an optimized production build that gets embedded into the Go binary.

### How the frontend communicates with the backend

```
React component
    | calls useListSectionsQuery({ q: "database" })
    v
RTK Query (api slice)
    | GET /api/sections?q=database
    v
Go HTTP Server (:8088)
    | queries SQLite Store
    v
JSON response --> RTK Query caches it --> component re-renders
```

### Vite proxy configuration for development

```ts
// web/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8088', changeOrigin: true },
    },
  },
  build: { outDir: 'dist', sourcemap: false },
})
```

---

## React Component Decomposition

This section maps every visual element in the prototype JSX to a modular React component.

### Component tree

```
<App>
  +-- <MenuBar />            -- Top menu bar
  +-- <AppLayout>            -- Two-pane container
  |   +-- <Sidebar>           -- Left panel
  |   |   +-- <TitleBar />     -- "Sections" header
  |   |   +-- <SearchBar />    -- Search input
  |   |   +-- <TypeFilter />   -- Filter buttons (All, Topic, Example, App, Tutorial)
  |   |   +-- <SectionList>    -- Scrollable list
  |   |   |   +-- <SectionCard />  -- Individual entry
  |   |   +-- <StatusBar />    -- Bottom bar ("N sections")
  |   +-- <ContentPanel>      -- Right panel
  |       +-- <TitleBar />     -- Section title header
  |       +-- <EmptyState />   -- No section selected
  |       +-- <SectionView>    -- Full section rendering
  |           +-- <SectionHeader /> -- Badge, slug, title, short, tags
  |           +-- <MarkdownContent /> -- Rendered Markdown body
```

### Mapping from prototype to components

| Prototype code region | New component | Module path |
|-----------------------|---------------|-------------|
| `TitleBar` function | `<TitleBar />` | `src/components/TitleBar/TitleBar.tsx` |
| `Badge` function | `<Badge />` | `src/components/Badge/Badge.tsx` |
| `inlineFormat` function | `<InlineFormat />` | `src/components/Markdown/InlineFormat.tsx` |
| `renderMarkdown` function | `<MarkdownContent />` | `src/components/Markdown/MarkdownContent.tsx` |
| Menu bar div | `<MenuBar />` | `src/components/MenuBar/MenuBar.tsx` |
| Search input | `<SearchBar />` | `src/components/SearchBar/SearchBar.tsx` |
| Type filter buttons | `<TypeFilter />` | `src/components/TypeFilter/TypeFilter.tsx` |
| Section list loop | `<SectionList />` + `<SectionCard />` | `src/components/SectionList/` |
| Status bar | `<StatusBar />` | `src/components/StatusBar/StatusBar.tsx` |
| Section header area | `<SectionHeader />` | `src/components/SectionView/SectionHeader.tsx` |
| Content rendering area | `<SectionView />` | `src/components/SectionView/SectionView.tsx` |
| Empty state | `<EmptyState />` | `src/components/EmptyState/EmptyState.tsx` |
| Root App | `<App />` | `src/App.tsx` |

### Component module structure

Each component follows this layout:

```
src/components/Badge/
  Badge.tsx          -- Component implementation
  Badge.stories.tsx   -- Storybook stories
  parts.ts           -- data-part names for this component
  index.ts           -- Public re-exports
  styles/
    badge.css        -- Structural styles using tokens
    theme-default.css -- Default token values
```

### Key component specs

**`<Badge />`** -- Props: `{ text: string; variant: 'topic' | 'type' | 'command' | 'flag' }`. Color varies by variant and section type. `data-part="root"`.

**`<SearchBar />`** -- Props: `{ value, onChange, placeholder? }`. Search input with retro inset border. `data-part`: `root`, `icon`, `input`.

**`<TypeFilter />`** -- Props: `{ activeFilter, onFilterChange }`. Row of toggle buttons. `data-part`: `root`, `button`. `data-state`: `active`, `inactive`.

**`<SectionCard />`** -- Props: `{ section, isActive, onSelect }`. Clickable card with type badge, title, short description. `data-part`: `root`, `badge-row`, `title`, `short`. `data-state`: `active`, `inactive`.

**`<SectionView />`** -- Props: `{ section: SectionDetail }`. Full content view with header and rendered Markdown. `data-part`: `root`, `header`, `divider`, `content`.

**`<MarkdownContent />`** -- Props: `{ content: string }`. Renders Markdown using `react-markdown` + `remark-gfm`. Handles headings, code blocks, tables, lists, inline formatting.

**`<EmptyState />`** -- No props. Centered icon with "Select a section" text.

---

## RTK Query Integration

RTK Query auto-generates React hooks from an API slice definition. Here is the complete setup:

### API slice (`web/src/services/api.ts`)

```ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export interface SectionSummary {
  slug: string
  section_type: string
  title: string
  short: string
  topics: string[]
  commands: string[]
  flags: string[]
  is_top_level: boolean
  show_per_default: boolean
}

export interface SectionDetail extends SectionSummary {
  content: string
  related: SectionSummary[]
}

export interface ListSectionsResponse {
  sections: SectionSummary[]
  total: number
}

export const helpApi = createApi({
  reducerPath: 'helpApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Section'],
  endpoints: (builder) => ({
    listSections: builder.query<ListSectionsResponse, {
      q?: string; type?: string; topic?: string; toplevel?: boolean
    }>({
      query: (params) => ({ url: '/sections', params }),
      providesTags: ['Section'],
    }),
    getSection: builder.query<SectionDetail, string>({
      query: (slug) => `/sections/${slug}`,
      providesTags: (_r, _e, slug) => [{ type: 'Section', id: slug }],
    }),
    searchSections: builder.query<ListSectionsResponse, string>({
      query: (q) => ({ url: '/sections/search', params: { q } }),
      providesTags: ['Section'],
    }),
    healthCheck: builder.query<{ ok: boolean; sections: number }, void>({
      query: () => '/health',
    }),
  }),
})

export const {
  useListSectionsQuery,
  useGetSectionQuery,
  useSearchSectionsQuery,
  useHealthCheckQuery,
} = helpApi
```

### Redux store (`web/src/store.ts`)

```ts
import { configureStore } from '@reduxjs/toolkit'
import { helpApi } from './services/api'

export const store = configureStore({
  reducer: { [helpApi.reducerPath]: helpApi.reducer },
  middleware: (gDM) => gDM().concat(helpApi.middleware),
})
```

### Usage in components

```tsx
function App() {
  const [activeSlug, setActiveSlug] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState('All')

  const { data: listData } = useListSectionsQuery({
    q: searchQuery || undefined,
    type: typeFilter !== 'All' ? typeFilter : undefined,
  })
  const { data: section } = useGetSectionQuery(activeSlug!, {
    skip: !activeSlug,
  })
  // ... render
}
```

RTK Query handles all caching, loading states, and automatic refetching when query
parameters change. Components never call `fetch()` directly.

---

## Theming System

The help browser uses a CSS-variable-based theming system with `data-part` selectors for
stable styling hooks. This section explains how the theming works and how to customize it.

### Why CSS variables + data-part?

Traditional CSS approaches use class names (like `.sidebar`, `.badge`), but class names
are implementation details that can change during refactoring. Instead:

- **CSS variables** (e.g., `--color-bg: #fff`) provide named, overridable values for
colors, fonts, spacing, etc. A user can override these to create a completely different
look without touching any component code.

- **`data-part` attributes** (e.g., `data-part="title"`) provide stable selectors that
don't change when class names are refactored. Consumers style against `data-part`, not
class names.

This means you can swap the entire visual theme by overriding CSS variables, and you
can target specific elements for custom styling using `data-part` selectors that are
guaranteed to remain stable across refactors.

### Token categories

The theme defines CSS variables in these groups:

```css
:root {
  /* Colors */
  --color-bg: #a8a8a8;
  --color-surface: #ffffff;
  --color-text: #000000;
  --color-muted: #777777;
  --color-accent: #4a7c59;
  --color-border: #000000;
  --color-code-bg: #1a1a1a;
  --color-code-text: #c0c0c0;

  /* Typography */
  --font-family: 'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', sans-serif;
  --font-mono: 'Monaco', 'Courier New', monospace;
  --font-size-base: 13px;
  --font-size-sm: 10px;
  --font-size-lg: 16px;
  --font-size-xl: 24px;

  /* Spacing */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 24px;
  --space-6: 32px;

  /* Borders */
  --border-width: 2px;
  --border-color: #000000;
  --radius: 2px;

  /* Shadows */
  --shadow-window: 2px 2px 0 #000;

  /* Layout */
  --sidebar-width: 280px;
  --content-max-width: 680px;
  --menubar-height: 22px;
  --titlebar-height: 22px;
}
```

### How theming layers work

There are three layers, applied in order:

1. **Base structural CSS** (`widget.css`): Uses `data-part` selectors and references
tokens. Defines layout, borders, and structure. Never hardcodes colors or sizes.

2. **Default theme** (`theme-default.css`): Sets all CSS variable values. This is the
\"classic Mac\" retro look from the prototype.

3. **Consumer overrides**: A user can override any variable at the `:root` level or
scope overrides to a specific `data-widget` attribute.

Example of base structural CSS:

```css
/* Base structure -- no hardcoded colors */
:where([data-widget="help-browser"]) [data-part="badge"] {
  display: inline-block;
  padding: var(--space-1) 7px;
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius);
  font-size: var(--font-size-sm);
  font-family: var(--font-family);
}
```

Example of a theme override:

```css
/* Dark theme override */
[data-widget="help-browser"].theme-dark {
  --color-bg: #1a1a2e;
  --color-surface: #16213e;
  --color-text: #e0e0e0;
  --color-muted: #888888;
  --color-border: #333333;
}
```

### Unstyled mode

An `unstyled` prop on the root `<App>` component skips importing the default theme CSS.
The `data-part` attributes are still rendered so consumers can apply their own styles.
This is useful for embedding the help browser inside another application with its own
design system.

---

## RTK Query Integration

This section explains how RTK Query connects the React frontend to the Go backend API.

### What is RTK Query?

RTK Query is a data-fetching and caching library built into Redux Toolkit. Instead of
writing `useEffect` + `fetch` + state management for every API call, you:

1. **Define an API slice** that describes your endpoints (URL, method, response type).
2. **RTK Query generates React hooks** automatically (e.g., `useListSectionsQuery()`).
3. **Components call these hooks** and get `{ data, isLoading, error }` back.
4. **RTK Query handles caching**, deduplication, background refetching, and optimistic updates.

### API slice definition

The API slice is the single source of truth for all backend communication:

```ts
// web/src/services/api.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export interface SectionSummary {
  slug: string
  section_type: string
  title: string
  short: string
  topics: string[]
  commands: string[]
  flags: string[]
  is_top_level: boolean
  show_per_default: boolean
}

export interface SectionDetail extends SectionSummary {
  content: string
  related: SectionSummary[]
}

export interface ListSectionsResponse {
  sections: SectionSummary[]
  total: number
}

export const helpApi = createApi({
  reducerPath: 'helpApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Section'],
  endpoints: (builder) => ({
    listSections: builder.query<ListSectionsResponse, {
      q?: string
      type?: string
      topic?: string
      toplevel?: boolean
    }>({
      query: (params) => ({
        url: '/sections',
        params,
      }),
      providesTags: ['Section'],
    }),

    getSection: builder.query<SectionDetail, string>({
      query: (slug) => `/sections/${slug}`,
      providesTags: (_r, _e, slug) => [{ type: 'Section', id: slug }],
    }),

    searchSections: builder.query<ListSectionsResponse, string>({
      query: (q) => ({
        url: '/sections/search',
        params: { q },
      }),
      providesTags: ['Section'],
    }),

    healthCheck: builder.query<{ ok: boolean; sections: number }, void>({
      query: () => '/health',
    }),
  }),
})

// Auto-generated hooks
export const {
  useListSectionsQuery,
  useGetSectionQuery,
  useSearchSectionsQuery,
  useHealthCheckQuery,
} = helpApi
```

### Redux store setup

```ts
// web/src/store.ts
import { configureStore } from '@reduxjs/toolkit'
import { helpApi } from './services/api'

export const store = configureStore({
  reducer: {
    [helpApi.reducerPath]: helpApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(helpApi.middleware),
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
```

### How components use the hooks

Example -- the sidebar filtering logic:

```tsx
// In App.tsx
import { useListSectionsQuery, useGetSectionQuery } from './services/api'

function App() {
  const [activeSlug, setActiveSlug] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('All')

  // RTK Query automatically caches and refetches when args change
  const { data: listData, isLoading } = useListSectionsQuery({
    q: searchQuery || undefined,
    type: typeFilter !== 'All' ? typeFilter : undefined,
  })

  const { data: section } = useGetSectionQuery(activeSlug!, {
    skip: !activeSlug, // Don't fetch if no section selected
  })

  return (
    <AppLayout>
      <Sidebar
        sections={listData?.sections ?? []}
        activeSlug={activeSlug}
        onSelect={setActiveSlug}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        typeFilter={typeFilter}
        onTypeFilterChange={setTypeFilter}
        totalCount={listData?.total ?? 0}
      />
      <ContentPanel section={section} />
    </AppLayout>
  )
}
```

Notice how the component doesn't manage any loading states, caching, or fetch logic.
RTK Query handles all of that. The component just describes *what* data it needs and
RTK Query provides it.

---

## Build Pipeline: Dagger + go:generate + go:embed {#build-pipeline}

This section explains how the React frontend gets built and embedded into the Go binary.
The key insight is that **the developer never needs to install Node.js or pnpm** -- the
Dagger builder runs the entire frontend build inside a container.

### The three-step pipeline

```
Step 1: go generate          Step 2: Dagger container        Step 3: go:embed
+------------------+      +-------------------------+      +------------------+
| //go:generate    |      | node:22 container        |      | //go:embed dist  |
| go run ../build- |----->|                          |----->| var staticFS     |
| web              |      | 1. corepack enable       |      | embed.FS         |
|                  |      | 2. pnpm install           |      |                  |
| (runs Dagger)    |      | 3. pnpm build             |      | (in Go binary)   |
+------------------+      | -> outputs dist/          |      +------------------+
                          +-------------------------+
```

### How each piece works

#### `go:generate` -- The trigger

In `cmd/help-browser/gen.go`:

```go
//go:generate go run ../build-web
package main
```

When you run `go generate ./cmd/help-browser`, Go executes the `build-web` program.
This is a standard Go build directive -- nothing fancy.

#### Dagger builder -- The containerized build

In `cmd/build-web/main.go`. This is a Go program that uses the Dagger SDK to:

1. Start a `node:22` container.
2. Mount the `web/` directory into the container.
3. Enable Corepack (which provides pnpm).
4. Run `pnpm install` to install dependencies.
5. Run `pnpm build` to produce `dist/`.
6. Export `dist/` to the host filesystem at `cmd/help-browser/dist/`.

```go
// cmd/build-web/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "dagger.io/dagger"
)

func main() {
    ctx := context.Background()
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil { log.Fatalf("connect: %v", err) }
    defer client.Close()

    wd, _ := os.Getwd()
    repoRoot := filepath.Dir(filepath.Dir(wd))  // cmd/build-web -> repo/
    webPath := filepath.Join(repoRoot, "web")
    outPath := filepath.Join(filepath.Dir(wd), "help-browser", "dist")

    base := client.Container().From("node:22")
    webDir := client.Host().Directory(webPath)

    ctr := base.
        WithWorkdir("/src").
        WithMountedDirectory("/src", webDir).
        WithExec([]string{"sh", "-lc",
            "corepack enable && corepack prepare pnpm@10.15.0 --activate"}).
        WithExec([]string{"sh", "-lc", "pnpm install --reporter=append-only"}).
        WithExec([]string{"sh", "-lc", "pnpm build"})

    dist := ctr.Directory("/src/dist")
    if _, err := dist.Export(ctx, outPath); err != nil {
        log.Fatalf("export: %v", err)
    }
    log.Printf("Built web -> %s", outPath)
}
```

**Key points:**
- `dagger.Connect()` connects to the Dagger engine (which uses Docker).
- `client.Container().From("node:22")` creates a container with Node.js.
- `client.Host().Directory(webPath)` mounts the `web/` directory from the host.
- Each `WithExec()` adds a command to the container's execution pipeline.
- `ctr.Directory("/src/dist").Export()` copies the build output back to the host.

#### `go:embed` -- The embedding

In `cmd/help-browser/main.go`:

```go
//go:embed dist
var staticFS embed.FS
```

This tells the Go compiler to embed the entire `dist/` directory into the binary.
At runtime, `staticFS` behaves like a read-only filesystem. The HTTP server serves
these files for all non-API routes.

### The complete build workflow

```bash
# 1. Build the frontend (runs in Dagger container, no Node.js needed on host)
go generate ./cmd/help-browser

# 2. Build the Go binary (embeds dist/)
go build -o glaze ./cmd/glaze

# 3. Run the server
./glaze serve --addr :8088 docs/ pkg/doc/topics/

# 4. Open browser
open http://localhost:8088
```

### Development workflow (with live reload)

During development, you run two processes:

```bash
# Terminal 1: Go server
go run ./cmd/glaze serve --addr :8088 docs/

# Terminal 2: Vite dev server (with HMR)
cd web && pnpm dev
```

The Vite dev server runs on `:5173` and proxies `/api/*` requests to the Go server
on `:8088`. Changes to React components appear instantly in the browser via HMR.
Changes to Go code require restarting the Go server.

---

## Storybook Integration

Storybook is a development environment for UI components that runs outside your main
application. It lets you browse a component library, view different states of each
component, and interactively test them.

### Why Storybook?

- **Isolated development**: Build and test components without running the full app.
- **Visual documentation**: Each story demonstrates a component's props and states.
- **Design review**: Share a URL with teammates to review UI changes.
- **Regression baseline**: Screenshots of stories serve as a visual regression baseline.

### Storybook setup

```bash
cd web
pnpm add -D @storybook/react-vite @storybook/addon-essentials
```

Configuration in `web/.storybook/main.ts`:

```ts
import type { StorybookConfig } from '@storybook/react-vite'

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: ['@storybook/addon-essentials'],
  framework: {
    name: '@storybook/react-vite',
    options: {},
  },
}
export default config
```

### Example stories

For each component, we create stories that demonstrate its key states:

```tsx
// web/src/components/Badge/Badge.stories.tsx
import type { Meta, StoryObj } from '@storybook/react'
import { Badge } from './Badge'

const meta: Meta<typeof Badge> = {
  title: 'Components/Badge',
  component: Badge,
  argTypes: {
    variant: {
      control: 'select',
      options: ['topic', 'type', 'command', 'flag'],
    },
  },
}
export default meta

type Story = StoryObj<typeof Badge>

export const TopicBadge: Story = {
  args: { text: 'documentation', variant: 'topic' },
}

export const TypeBadgeTopic: Story = {
  args: { text: 'GeneralTopic', variant: 'type' },
}

export const TypeBadgeExample: Story = {
  args: { text: 'Example', variant: 'type' },
}

export const CommandBadge: Story = {
  args: { text: 'help', variant: 'command' },
}

export const FlagBadge: Story = {
  args: { text: '--output', variant: 'flag' },
}
```

### Required stories per component

| Component | Required stories |
|-----------|-----------------|
| `<Badge />` | Each variant (topic, type, command, flag) |
| `<TitleBar />` | Default |
| `<SearchBar />` | Empty, with text |
| `<TypeFilter />` | Each filter active |
| `<SectionCard />` | Active, inactive, with top indicator |
| `<SectionList />` | Empty, with items, filtered |
| `<SectionView />` | With sample section data |
| `<MarkdownContent />` | Headings, code blocks, tables, lists, inline formatting |
| `<EmptyState />` | Default |
| `<MenuBar />` | Default |
| `<StatusBar />` | Default |

---

## File Layout

The complete file layout for the new code:

```
glazed/                             ← Repository root
+-- cmd/
|   +-- glaze/
|   |   +-- main.go                 ← ADD: register `serve` subcommand
|   +-- build-web/
|   |   +-- main.go                 ← NEW: Dagger builder
|   +-- help-browser/
|       +-- main.go                 ← NEW: Server + Cobra command
|       +-- gen.go                  ← NEW: go:generate directive
|       +-- dist/                   ← GENERATED: Vite build output
|           +-- index.html
|           +-- assets/
+-- pkg/
|   +-- help/
|       +-- server/                 ← NEW: HTTP handler package
|       |   +-- server.go           ← Server struct, route registration
|       |   +-- handlers.go         ← API endpoint handlers
|       |   +-- spa.go              ← SPA fallback handler
|       |   +-- middleware.go        ← CORS middleware
|       |   +-- types.go            ← Request/response types
|       |   +-- server_test.go      ← Handler tests
|       +-- help.go                 ← EXISTING: HelpSystem
|       +-- model/                  ← EXISTING: Section model
|       +-- store/                  ← EXISTING: SQLite store
+-- web/                            ← NEW: React frontend
|   +-- index.html
|   +-- package.json
|   +-- pnpm-lock.yaml
|   +-- tsconfig.json
|   +-- vite.config.ts
|   +-- .storybook/
|   |   +-- main.ts
|   |   +-- preview.ts
|   +-- src/
|       +-- main.tsx                ← Entry point
|       +-- App.tsx                 ← Root component
|       +-- store.ts                ← Redux store config
|       +-- services/
|       |   +-- api.ts              ← RTK Query API slice
|       +-- types/
|       |   +-- index.ts            ← TypeScript interfaces
|       +-- hooks/
|       |   +-- useSectionNavigation.ts  ← URL/slug state hook
|       +-- components/
|       |   +-- AppLayout/
|       |   |   +-- AppLayout.tsx
|       |   |   +-- AppLayout.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   |       +-- app-layout.css
|       |   |       +-- theme-default.css
|       |   +-- MenuBar/
|       |   |   +-- MenuBar.tsx
|       |   |   +-- MenuBar.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- TitleBar/
|       |   |   +-- TitleBar.tsx
|       |   |   +-- TitleBar.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- Badge/
|       |   |   +-- Badge.tsx
|       |   |   +-- Badge.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- SearchBar/
|       |   |   +-- SearchBar.tsx
|       |   |   +-- SearchBar.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- TypeFilter/
|       |   |   +-- TypeFilter.tsx
|       |   |   +-- TypeFilter.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- SectionList/
|       |   |   +-- SectionList.tsx
|       |   |   +-- SectionCard.tsx
|       |   |   +-- SectionList.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- SectionView/
|       |   |   +-- SectionView.tsx
|       |   |   +-- SectionHeader.tsx
|       |   |   +-- SectionView.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- Markdown/
|       |   |   +-- MarkdownContent.tsx
|       |   |   +-- InlineFormat.tsx
|       |   |   +-- MarkdownContent.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- EmptyState/
|       |   |   +-- EmptyState.tsx
|       |   |   +-- EmptyState.stories.tsx
|       |   |   +-- parts.ts
|       |   |   +-- index.ts
|       |   |   +-- styles/
|       |   +-- StatusBar/
|       |       +-- StatusBar.tsx
|       |       +-- StatusBar.stories.tsx
|       |       +-- parts.ts
|       |       +-- index.ts
|       |       +-- styles/
|       +-- styles/
|           +-- global.css           ← Root CSS variables, resets
|           +-- theme-default.css    ← Default theme token values
+-- ttmp/                           ← EXISTING: Ticket docs
+-- docs/                           ← EXISTING: Help Markdown files
```

---

## Implementation Plan (Phased)

This plan breaks the work into phases. Each phase produces a working, testable increment.
You should implement them in order.

### Phase 1: Go HTTP Server Scaffold (Tasks 1-2)

**Goal:** Get a working HTTP server that serves JSON from the existing HelpSystem.

**Files to create:**
- `pkg/help/server/server.go` -- Server struct with `NewServer()`, `ListenAndServe()`
- `pkg/help/server/handlers.go` -- `handleListSections`, `handleGetSection`, `handleHealth`
- `pkg/help/server/types.go` -- Request/response Go structs
- `pkg/help/server/middleware.go` -- CORS middleware
- `cmd/help-browser/main.go` -- Cobra command that wires everything together

**Steps:**

1. Create `pkg/help/server/types.go` with response structs:
   ```go
   type SectionSummary struct {
       Slug        string   `json:"slug\"`
       SectionType string   `json:\"section_type\"`
       Title       string   `json:\"title\"`
       Short       string   `json:\"short\"`
       Topics      []string `json:\"topics\"`
       Commands    []string `json:\"commands\"`
       Flags       []string `json:\"flags\"`
       IsTopLevel  bool     `json:\"is_top_level\"`
   }
   type ListResponse struct {
       Sections []SectionSummary `json:\"sections\"`
       Total    int              `json:\"total\"`
   }
   ```

2. Create `pkg/help/server/handlers.go` with handler functions.

3. Create `pkg/help/server/server.go` with route registration.

4. Add a Cobra `serve` subcommand in `cmd/help-browser/main.go` that:
   - Creates a `HelpSystem`
   - Loads sections from positional args (files/directories)
   - Creates a `Server` and calls `ListenAndServe()`

5. Write tests in `pkg/help/server/server_test.go` using `httptest.NewServer`.

6. Verify: `go test ./pkg/help/server/...` and manually `curl localhost:8088/api/sections`.

**Validation:**
```bash
go run ./cmd/help-browser serve docs/
curl -s localhost:8088/api/health | jq
curl -s localhost:8088/api/sections | jq '.total'
curl -s localhost:8088/api/sections/help-system | jq '.title'
```

---

### Phase 2: Scaffold React Frontend (Task 3)

**Goal:** Set up the `web/` directory with Vite, React, TypeScript, RTK Query.

**Files to create:**
- `web/package.json`
- `web/vite.config.ts`
- `web/tsconfig.json`
- `web/index.html`
- `web/src/main.tsx`
- `web/src/App.tsx` (placeholder)
- `web/src/store.ts`
- `web/src/services/api.ts`
- `web/src/types/index.ts`

**Steps:**

1. Create `web/` directory and `package.json` with dependencies:
   ```json
   {
     "name": "glazed-help-browser",
     "private": true,
     "version": "0.1.0",
     "type": "module",
     "packageManager": "pnpm@10.15.0",
     "scripts": {
       "dev": "vite",
       "build": "tsc && vite build",
       "preview": "vite preview"
     },
     "dependencies": {
       "@reduxjs/toolkit": "^2.2.3",
       "react": "^18.3.1",
       "react-dom": "^18.3.1",
       "react-redux": "^9.0.0",
       "react-markdown": "^9.0.1",
       "remark-gfm": "^4.0.0"
     },
     "devDependencies": {
       "@types/react": "^18.2.66",
       "@types/react-dom": "^18.2.22",
       "@vitejs/plugin-react": "^4.3.1",
       "typescript": "^5.5.3",
       "vite": "^5.4.0"
     }
   }
   ```

2. Create `web/vite.config.ts` with proxy to Go server.

3. Create TypeScript types in `web/src/types/index.ts` matching the API response shapes.

4. Create the RTK Query API slice in `web/src/services/api.ts`.

5. Create the Redux store in `web/src/store.ts`.

6. Create `web/src/main.tsx` that renders `<App />` inside `<Provider>`.

7. Create `web/src/App.tsx` with a placeholder that calls `useListSectionsQuery()` and
   displays the count.

**Validation:**
```bash
# Terminal 1
go run ./cmd/help-browser serve docs/
# Terminal 2
cd web && pnpm install && pnpm dev
# Open http://localhost:5173 -- should see section count
```

---

### Phase 3: Decompose Prototype into Components (Tasks 4-5)

**Goal:** Port the visual design from `glazed-docs-browser(2).jsx` into modular React components.

**Steps:**

1. **Extract `<MenuBar />`**: Copy the menu bar HTML structure into
   `web/src/components/MenuBar/MenuBar.tsx`. Convert inline styles to CSS classes using
   CSS variables.

2. **Extract `<TitleBar />`**: The `TitleBar` function from the prototype becomes a
   standalone component with `data-part` attributes.

3. **Extract `<Badge />`**: The `Badge` function becomes a component with variant-based
   coloring.

4. **Extract `<SearchBar />`**: The search input with its container and icon.

5. **Extract `<TypeFilter />`**: The row of filter buttons.

6. **Extract `<SectionCard />`**: A single section entry in the sidebar.

7. **Extract `<SectionList />`**: The scrollable list that renders `<SectionCard />` for each section.

8. **Extract `<SectionHeader />`**: The header area of the content panel (badges, title, short, tags).

9. **Extract `<MarkdownContent />`**: Replace the hand-written `renderMarkdown` function with
   `react-markdown`. Configure custom renderers to match the prototype's visual style:
   - Code blocks with language labels
   - Tables with borders
   - Headings with bottom borders

10. **Extract `<SectionView />`**: Composes `<SectionHeader />` + `<MarkdownContent />`.

11. **Extract `<EmptyState />`**: The placeholder when no section is selected.

12. **Extract `<StatusBar />`**: The bottom bar.

13. **Wire `<App />`**: Compose all components, connecting them with RTK Query hooks.

**For each component**, follow this pattern:

a. Create the directory: `web/src/components/ComponentName/`
b. Create `ComponentName.tsx` with the component implementation.
c. Create `parts.ts` with `data-part` names:
   ```ts
   export const PARTS = {
     root: 'component-name',
     // ... sub-parts
   } as const
   ```
d. Create `styles/component-name.css` using `data-part` selectors and CSS variables.
e. Create `styles/theme-default.css` with default token values for this component.
f. Create `index.ts` re-exporting the public API.
g. Verify in browser that it matches the prototype.

**Validation:**
- The web UI matches the prototype's visual appearance.
- Search filtering works.
- Type filtering works.
- Section selection and content rendering works.
- Markdown rendering handles code blocks, tables, lists correctly.

---

### Phase 4: Theming System (Task 5)

**Goal:** Complete the CSS variable + data-part theming system.

**Steps:**

1. Create `web/src/styles/global.css` with all CSS variables at `:root`.
2. Create `web/src/styles/theme-default.css` with the classic Mac retro theme values.
3. Convert all component styles to use CSS variables instead of hardcoded values.
4. Ensure all elements have `data-part` attributes.
5. Test theme override by adding a CSS class that overrides variables.
6. Implement `unstyled` prop on `<App />`.

**Validation:**
- Override `--color-bg` in browser DevTools and see the background change.
- Target `[data-part="badge"]` in DevTools and see styles apply.
- Toggle `unstyled` mode and verify no base CSS is applied.

---

### Phase 5: Storybook Stories (Task 6)

**Goal:** Add Storybook stories for all components.

**Steps:**

1. Install Storybook: `pnpm add -D @storybook/react-vite @storybook/addon-essentials`
2. Configure `.storybook/main.ts` and `.storybook/preview.ts`.
3. Create stories for each component (see the [Storybook Integration](#storybook-integration) section).
4. Run `pnpm storybook` and verify all stories render correctly.

**Validation:**
```bash
cd web && pnpm storybook
# Open http://localhost:6006 -- all components visible
```

---

### Phase 6: Dagger Build Pipeline (Tasks 7-8)

**Goal:** Create the Dagger builder and wire go:generate + go:embed.

**Steps:**

1. Create `cmd/build-web/main.go` (Dagger builder -- see [Build Pipeline](#build-pipeline)).
2. Create `cmd/help-browser/gen.go` with `//go:generate go run ../build-web`.
3. Add `//go:embed dist` to `cmd/help-browser/main.go`.
4. Create the SPA fallback handler.
5. Run `go generate ./cmd/help-browser` and verify `dist/` is produced.
6. Build and run the single binary.

**Validation:**
```bash
go generate ./cmd/help-browser
ls cmd/help-browser/dist/index.html  # Should exist
go build -o glaze ./cmd/glaze
./glaze serve docs/
curl localhost:8088/                  # Should serve index.html
curl localhost:8088/api/health        # Should return JSON
```

---

### Phase 7: Cobra Integration (Task 9)

**Goal:** Wire the `serve` command as a subcommand of the main `glaze` binary.

**Steps:**

1. Move or refactor the server setup from `cmd/help-browser/` into a reusable function
   in `pkg/help/server/`.
2. In `cmd/glaze/main.go`, add:
   ```go
   serveCmd, err := server.NewServeCommand(helpSystem)
   cobra.CheckErr(err)
   rootCmd.AddCommand(serveCmd)
   ```
3. Test that `glaze serve --help` works and `glaze serve docs/` starts the server.

**Validation:**
```bash
go run ./cmd/glaze serve --help
go run ./cmd/glaze serve docs/
```

---

### Phase 8: Integration Testing (Task 10)

**Goal:** End-to-end tests.

**Steps:**

1. Add Go integration tests that start the server with `httptest.NewServer`, make API calls,
   and verify responses.
2. Add a test that loads a fixture `.md` file and verifies it appears in the API.
3. Add a test that verifies the SPA fallback serves `index.html` for unknown paths.
4. Optionally add Playwright or Cypress tests for the frontend.

---

## Testing Strategy

### Backend tests (Go)

| Test type | Tool | What it covers |
|-----------|------|---------------|
| Unit tests | `go test` | Handler functions, file discovery, JSON serialization |
| Integration tests | `httptest.NewServer` | Full request/response cycles with real HelpSystem |
| Edge cases | `go test` | Missing slug (404), empty search, invalid DSL query |

Example test:

```go
func TestListSections(t *testing.T) {
    hs := help.NewHelpSystem()
    hs.AddSection(&help.Section{
        Section: &model.Section{
            Slug:        \"test-section\",
            SectionType: model.SectionGeneralTopic,
            Title:       \"Test Section\",
            Short:       \"A test\",
            Topics:      []string{\"test\"},
        },
    })

    srv := server.NewServer(hs, \"\")
    ts := httptest.NewServer(srv.Handler())
    defer ts.Close()

    resp, err := http.Get(ts.URL + \"/api/sections\")
    require.NoError(t, err)
    defer resp.Body.Close()

    var result server.ListResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    assert.Equal(t, 1, result.Total)
    assert.Equal(t, \"test-section\", result.Sections[0].Slug)
}
```

### Frontend tests (TypeScript)

| Test type | Tool | What it covers |
|-----------|------|---------------|
| Component tests | Vitest + Testing Library | Component rendering, user interactions |
| API integration | MSW (Mock Service Worker) | Mocked API responses |
| Visual tests | Storybook snapshots | Component visual states |

---

## Risks and Open Questions

### Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Dagger engine not available in CI | Build fails | Fall back to Makefile + local Node.js |
| Large help file sets (>1000 sections) | Slow API responses | Add pagination to list endpoint |
| Markdown rendering inconsistencies | UI bugs | Use `react-markdown` which handles GFM; test with real help files |
| `go:embed` binary size | Large binary | Compress assets; consider lazy-loading |
| Browser compatibility | CSS issues in older browsers | Target modern browsers (Chrome, Firefox, Safari latest 2 versions) |

### Open Questions

1. **URL routing**: Should we use `react-router` with `BrowserRouter` (requires SPA fallback)
   or `HashRouter` (simpler, no server fallback needed)? Recommendation: `HashRouter` for v1
   to avoid the SPA fallback complexity.

2. **Pagination**: The initial `GET /api/sections` returns all sections. If the help file set
   grows large, we should add `?page=1&limit=50`. Not needed for v1.

3. **WebSocket updates**: If the user adds files while the server is running, should the
   browser auto-update? Not for v1 -- the user can refresh.

4. **Authentication**: Should the serve endpoint support basic auth? Not for v1 -- it's a
   local development tool.

---

## References

### Project files

- `pkg/help/help.go` -- HelpSystem struct and API
- `pkg/help/model/section.go` -- Section data model with JSON tags
- `pkg/help/store/store.go` -- SQLite store CRUD operations
- `pkg/help/store/query.go` -- QueryCompiler for advanced filtering
- `pkg/help/dsl/` -- Query DSL parser (lexer, parser, compiler)
- `pkg/help/cmd/cobra.go` -- Existing Cobra help command integration
- `pkg/help/ui/model.go` -- Existing Charm TUI (reference for navigation patterns)
- `cmd/glaze/main.go` -- Main entry point where `serve` will be registered

### Imported prototype

- `ttmp/.../sources/local/glazed-docs-browser(2).jsx` -- The monolithic React prototype

### External documentation

- **Vite**: https://vitejs.dev/guide/
- **React**: https://react.dev/learn
- **RTK Query**: https://redux-toolkit.js.org/rtk-query/overview
- **Dagger Go SDK**: https://docs.dagger.io/sdk/go
- **go:embed**: https://pkg.go.dev/embed
- **Storybook for React**: https://storybook.js.org/docs/get-started/frameworks/react-vite
- **react-markdown**: https://github.com/remarkjs/react-markdown
- **Glazed help system reference**: `glaze help help-system`
- **Dagger+Vite+React guide**: `remarquee/pkg/doc/topics/how-to-create-a-web-app-with-react-rtk-vite-dagger-gen.md`

### Related tickets

- This ticket: GL-011-HELP-BROWSER
- Diary: `ttmp/.../reference/01-diary.md`
