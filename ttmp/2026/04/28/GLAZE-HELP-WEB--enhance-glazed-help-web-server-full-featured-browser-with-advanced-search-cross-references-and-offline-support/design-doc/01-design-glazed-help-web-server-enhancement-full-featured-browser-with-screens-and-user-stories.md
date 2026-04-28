---
Title: 'Design: Glazed Help Web Server Enhancement — Full-Featured Browser with Screens and User Stories'
Ticket: GLAZE-HELP-WEB
Status: active
Topics:
    - glazed
    - help-system
    - web
    - react
    - server
    - ui
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/server/handlers.go
      Note: API route handlers
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go
      Note: HTTP server command and handler composition
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/server/types.go
      Note: Request/response types
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/ui/model.go
      Note: TUI model for feature parity benchmark
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/web/static.go
      Note: SPA static file handler
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/web/src/App.tsx
      Note: Root React component
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/web/src/services/api.ts
      Note: RTK Query API slice
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-28T08:00:00-04:00
WhatFor: ""
WhenToUse: ""
---








# Design: Glazed Help Web Server Enhancement

## Executive Summary

This document describes the design for enhancing the Glazed help web server from its current minimal state into a **full-featured help browser**. The existing `glaze serve-help` command already starts an HTTP server and serves a React SPA, but the frontend is basic: a searchable sidebar list and a markdown content viewer. This design adds advanced search, cross-references, topic browsing, command coverage maps, and offline support — turning the help browser into a genuine documentation platform.

The enhancements are organized around **user stories** derived from real use cases, then translated into **screen designs** with ASCII mockups, **API contracts**, and an **implementation plan**. The document is written for a new intern and explains every part of the system from first principles.

---

## Problem Statement and Scope

### What exists today?

The Glazed help system provides three ways to consume documentation:

1. **Terminal:** `glaze help <topic>`, `glaze help --ui` (Bubble Tea TUI)
2. **HTTP server:** `glaze serve-help` (React SPA at `localhost:8088`)
3. **Static export:** `glaze render-site` (static SPA files for hosting)

The web frontend (React) currently supports:
- Sidebar list of all sections with search filtering
- Type filter buttons (All, Topic, Example, App, Tutorial)
- Click-to-view markdown content rendering
- Client-side routing (`#/sections/:slug`)
- Server API: `GET /api/health`, `GET /api/sections`, `GET /api/sections/:slug`

### What is missing?

While functional, the current web browser is a **read-only list viewer**. It lacks the affordances users expect from modern documentation platforms:

- **No full-text search.** The search box only filters the already-loaded list by title, short description, topics, and slug. It does not search the body content.
- **No cross-references.** When reading about `json` command help, there is no "See Also" list linking to related topics, examples, or tutorials.
- **No topic browsing.** Users cannot explore all sections tagged with `database` or see a tag cloud.
- **No command coverage map.** There is no way to see which commands have documentation and which do not.
- **No DSL query support.** The terminal TUI supports the Glazed query DSL (`type:example AND topic:database`), but the web UI does not.
- **No deep linking to anchors.** URLs only go to sections, not to headings within sections.
- **No offline support.** The SPA requires a live server (or pre-built static JSON). There is no service worker caching.
- **No print/export from the browser.** Users cannot download a section as PDF or print a formatted page.
- **No dark mode or accessibility features.** The retro styling is fun but lacks contrast modes and keyboard navigation.

### Scope of this ticket

This ticket covers the **design and phased implementation** of the following enhancements:

1. **Enhanced Search** — Full-text search over content, with highlighting and query DSL support.
2. **Cross-Reference Panel** — "See Also" sidebar showing related sections by topic, command, and flag.
3. **Topic Browser** — Dedicated topic/tag exploration screen with a tag cloud.
4. **Command Coverage Map** — Dashboard showing which CLI commands have help coverage.
5. **Deep Linking** — Hash-based anchor links to headings within sections.
6. **Offline Support** — Service worker with cache-first strategy for static assets and JSON data.
7. **Print/Export** — Browser-native print styles and "Download as Markdown" button.
8. **Accessibility & Theming** — Keyboard shortcuts, focus management, and dark mode toggle.

### Out of scope

- Rewriting the Go backend storage layer (the existing `store.Store` is sufficient).
- Replacing the TUI (`glaze help --ui`) — it remains a separate interface.
- Multi-language i18n support.
- User authentication or edit-in-browser.

---

## Current-State Architecture

Before designing enhancements, we must understand the existing stack. This section explains every layer.

### Layer 1: The Go HTTP Server

The `glaze serve-help` command starts an HTTP server with two routes:

```
GET /api/*   → REST API (JSON)
GET /*       → React SPA (index.html + static assets)
```

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go`

The server is created by `NewServeCommand`, which implements `cmds.BareCommand`. It:
1. Loads help sections into the `HelpSystem` store (embedded docs by default, or from explicit paths).
2. Creates an `http.Handler` that multiplexes `/api` to the API handler and everything else to the SPA handler.
3. Listens on `:8088` (configurable via `--address`).

**Key code:**
```go
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    apiHandler := NewHandler(deps)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cleanPath := stdpath.Clean("/" + r.URL.Path)
        if strings.HasPrefix(cleanPath, "/api/") {
            apiHandler.ServeHTTP(w, r)
            return
        }
        spaHandler.ServeHTTP(w, r)
    })
}
```

### Layer 2: The REST API

The API is defined in `pkg/help/server/handlers.go` and `types.go`.

**Endpoints:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check + section count |
| `/api/sections` | GET | List all sections (summary shape) |
| `/api/sections/search` | GET | Search sections (same as `/api/sections` with `?q=...`) |
| `/api/sections/:slug` | GET | Get one section (full detail shape) |

**Response types:**

```go
// SectionSummary — list results
type SectionSummary struct {
    ID         int64    `json:"id"`
    Slug       string   `json:"slug"`
    Type       string   `json:"type"`       // "GeneralTopic" | "Example" | ...
    Title      string   `json:"title"`
    Short      string   `json:"short"`
    Topics     []string `json:"topics"`
    IsTopLevel bool     `json:"isTopLevel"`
}

// SectionDetail — full content
type SectionDetail struct {
    ID         int64    `json:"id"`
    Slug       string   `json:"slug"`
    Type       string   `json:"type"`
    Title      string   `json:"title"`
    Short      string   `json:"short"`
    Topics     []string `json:"topics"`
    Flags      []string `json:"flags"`
    Commands   []string `json:"commands"`
    IsTopLevel bool     `json:"isTopLevel"`
    Content    string   `json:"content"`    // Full markdown body
}
```

**File references:**
- `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/handlers.go`
- `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/types.go`

### Layer 3: The React SPA

The frontend is a Vite + React + TypeScript application located at `/home/manuel/code/wesen/corporate-headquarters/glazed/web/`.

**Tech stack:**
- **Build tool:** Vite (with `@vitejs/plugin-react`)
- **Router:** `react-router-dom` (HashRouter — `HashRouter` means routes use `#/sections/slug` instead of `/sections/slug`, which avoids needing server-side route fallback)
- **State management:** Redux Toolkit + RTK Query (for API caching)
- **Styling:** CSS with data-part attributes (BEM-like naming without classes)
- **Markdown rendering:** Custom `MarkdownContent` component

**File references:**
- Entry: `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/main.tsx`
- Root component: `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/App.tsx`
- API layer: `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/services/api.ts`
- Store: `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/store.ts`
- Types: `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/types/index.ts`

**Component inventory:**

| Component | File | Purpose |
|-----------|------|---------|
| `AppLayout` | `components/AppLayout/AppLayout.tsx` | Two-pane layout (sidebar + content) |
| `TitleBar` | `components/TitleBar/TitleBar.tsx` | Header bar with title |
| `MenuBar` | `components/MenuBar/MenuBar.tsx` | Retro menu bar (File, Edit, View, Help) |
| `SearchBar` | `components/SearchBar/SearchBar.tsx` | Text input with magnifying glass icon |
| `TypeFilter` | `components/TypeFilter/TypeFilter.tsx` | Row of filter buttons |
| `SectionList` | `components/SectionList/SectionList.tsx` | Scrollable list of `SectionCard` |
| `SectionCard` | `components/SectionList/SectionCard.tsx` | Clickable summary with badge and title |
| `SectionView` | `components/SectionView/SectionView.tsx` | Full section rendering |
| `SectionHeader` | `components/SectionView/SectionHeader.tsx` | Title, slug, tags |
| `MarkdownContent` | `components/Markdown/MarkdownContent.tsx` | Markdown body renderer |
| `EmptyState` | `components/EmptyState/EmptyState.tsx` | Placeholder when no section selected |
| `StatusBar` | `components/StatusBar/StatusBar.tsx` | Footer with result count |
| `Badge` | `components/Badge/Badge.tsx` | Small pill labels for type/topic/command/flag |

### Layer 4: The TUI (Terminal UI)

For comparison, the terminal UI (`glaze help --ui`) provides:
- Real-time search with DSL support
- Glamour-rendered markdown
- Clipboard copy (`y`)
- Keyboard shortcuts (`/` search, `?` help, `ctrl+h` DSL cheatsheet)

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/ui/model.go`

The TUI is richer in some ways (DSL search, clipboard) but lacks the spatial layout and hyperlinking of a web browser.

### Architecture Diagram (Current State)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         User Interaction                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────────┐ │
│  │ Terminal    │  │ Browser     │  │ Static Site (render-site)       │ │
│  │ glaze help  │  │ localhost   │  │ HTML + JSON files               │ │
│  │ --ui        │  │ :8088       │  │ (no live server)                │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────────────────────────┘ │
│         │                │                                             │
│         ▼                ▼                                             │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                        help.HelpSystem                            │  │
│  │  ┌─────────────────────────────────────────────────────────────┐ │  │
│  │  │                      store.Store                             │ │  │
│  │  │  (SQLite — sections table with full-text search optional)    │ │  │
│  │  └─────────────────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│         │                           │                                   │
│         ▼                           ▼                                   │
│  ┌──────────────┐          ┌──────────────────┐                        │
│  │ Bubble Tea   │          │ Go HTTP Server   │                        │
│  │ TUI Model    │          │ (pkg/help/server)│                        │
│  │ (pkg/help/ui)│          │                  │                        │
│  └──────────────┘          │  /api/health     │                        │
│                            │  /api/sections   │                        │
│                            │  /api/sections/  │                        │
│                            │    :slug         │                        │
│                            └────────┬─────────┘                        │
│                                     │                                  │
│                                     ▼                                  │
│                            ┌──────────────────┐                        │
│                            │ React SPA        │                        │
│                            │ (web/src/App.tsx)│                        │
│                            │                  │                        │
│                            │  Sidebar (list)  │                        │
│                            │  Content (view)  │                        │
│                            └──────────────────┘                        │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Appendix A: Current React Component Hierarchy (Designer Reference)

This appendix is a **complete inventory of the current React component tree** for the Glazed Help Browser web app. It is intended as a starting point for web designers: every component is listed with its props, state dependencies (hooks / Redux), CSS `data-part` attributes, and file location. Use this to understand the layout before proposing redesigns.

### High-Level Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  MenuBar (decorative)                                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Sidebar (280px) ──────────────────────┐  ┌─ Content ──────────────────┐  │
│  │                                         │  │                           │  │
│  │  TitleBar("📁 Sections")                │  │  TitleBar("📄 Doc...")    │  │
│  │  ┌─ SearchBar ──────────────────────┐   │  │                           │  │
│  │  │ 🔍  [Search input]               │   │  │  Loading / Error /        │  │
│  │  └─────────────────────────────────┘   │  │  SectionView / EmptyState   │  │
│  │  [All][Topic][Example][App][Tut]       │  │                           │  │
│  │                                         │  │                           │  │
│  │  SectionList                            │  │                           │  │
│  │  ├─ SectionCard (active)                │  │                           │  │
│  │  ├─ SectionCard                         │  │                           │  │
│  │  └─ ...                                 │  │                           │  │
│  │                                         │  │                           │  │
│  │  StatusBar("24 sections")               │  │                           │  │
│  │                                         │  │                           │  │
│  └─────────────────────────────────────────┘  └───────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Tree (Hierarchical)

```
App (root)
├── MenuBar (not currently rendered in App.tsx, but exists)
└── AppLayout
    ├── sidebar (prop: ReactNode)
    │   ├── TitleBar
    │   ├── <div> — search + filter wrapper
    │   │   ├── SearchBar
    │   │   └── TypeFilter
    │   ├── SectionList
    │   │   └── SectionCard[]
    │   │       └── Badge (inside each card)
    │   └── StatusBar
    └── content (prop: ReactNode)
        ├── TitleBar
        ├── Loading placeholder (inline)
        ├── Error placeholder (inline)
        ├── SectionView (conditionally rendered)
        │   ├── SectionHeader
        │   │   └── Badge[]
        │   └── MarkdownContent
        └── EmptyState (conditionally rendered)
```

---

### Component Inventory

#### `App` — Root Component

**File:** `web/src/App.tsx`

**Local State (useState):**
| State | Type | Initial | Purpose |
|-------|------|---------|---------|
| `search` | `string` | `''` | Search query text |
| `filter` | `FilterValue` | `'All'` | Active type filter |

**Router Hooks:**
- `useLocation()` — to read current URL and extract active slug
- `useNavigate()` — to push `#/sections/:slug` on selection
- `matchPath('/sections/:slug', location.pathname)` — parses active slug

**RTK Query Hooks:**
| Hook | Data | Purpose |
|------|------|---------|
| `useListSectionsQuery()` | `ListSectionsResponse` | Load all section summaries |
| `useGetSectionQuery(slug, { skip })` | `SectionDetail` | Load full content for active section |

**Computed:**
- `activeSlug: string | null` — extracted from URL
- `filtered: SectionSummary[]` — client-side filter by `filter` + `search`

**Callback:**
- `handleSelect(slug: string)` — navigates to `#/sections/${slug}`

**Layout:** Renders `AppLayout` with `sidebar` and `content` props.

---

#### `AppLayout` — Two-Pane Layout Container

**File:** `web/src/components/AppLayout/AppLayout.tsx`

**Props:**
| Prop | Type | Description |
|------|------|-------------|
| `sidebar` | `ReactNode` | Left pane content (sidebar) |
| `content` | `ReactNode` | Right pane content (main view) |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `app-layout` | `[data-part='app-layout']` | `display: flex`, `flex: 1`, `padding: 8px`, `gap: 1px` |
| `app-layout-sidebar` | `[data-part='app-layout-sidebar']` | `width: 280px`, `border: 2px solid #000`, `box-shadow: 2px 2px 0 #000` |
| `app-layout-content` | `[data-part='app-layout-content']` | `flex: 1`, `border: 2px solid #000`, `box-shadow: 2px 2px 0 #000` |

**CSS File:** `web/src/components/AppLayout/styles/app-layout.css`

---

#### `MenuBar` — Retro Menu Bar (Decorative)

**File:** `web/src/components/MenuBar/MenuBar.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `title` | `string` | `'Glazed Help Browser'` | App title shown at right |

**Static Items:** `['File', 'Edit', 'View', 'Help']` + Apple logo (``)

**Note:** Currently **not rendered** in `App.tsx`. It exists as a component but is unused. The design proposes making it functional (View → Dark Mode, etc.).

**CSS `data-part` attributes:**
| Part | Selector |
|------|----------|
| `menubar` | `[data-part='menubar']` |
| `menubar-item` | `[data-part='menubar-item']` |
| `menubar-apple` | `[data-part='menubar-apple']` |
| `menubar-title` | `[data-part='menubar-title']` |

**CSS File:** `web/src/components/MenuBar/styles/menubar.css`

---

#### `TitleBar` — Window Title Bar

**File:** `web/src/components/TitleBar/TitleBar.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `title` | `string` | (required) | Centered title text |
| `icon` | `React.ReactNode` | `undefined` | Optional icon in left box |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `titlebar` | `[data-part='titlebar']` | `height: 22px`, `border-bottom: 2px solid #000` |
| `titlebar-icon` | `[data-part='titlebar-icon']` | `width: 20px`, `border-right: 2px solid #000` |
| `titlebar-icon-box` | `[data-part='titlebar-icon-box']` | `11px × 11px` square box |
| `titlebar-ruler` | `[data-part='titlebar-ruler']` | Flex container with stripe lines |
| `titlebar-stripe` | `[data-part='titlebar-stripe']` | 1px black horizontal stripe pattern |
| `titlebar-title` | `[data-part='titlebar-title']` | `font-size: 12px`, `font-weight: 700` |

**CSS File:** `web/src/components/TitleBar/styles/titlebar.css`

---

#### `SearchBar` — Search Input

**File:** `web/src/components/SearchBar/SearchBar.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `string` | (required) | Current search text |
| `onChange` | `(value: string) => void` | (required) | Called on every keystroke |
| `placeholder` | `string` | `'Search…'` | Input placeholder |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `searchbar` | `[data-part='searchbar']` | `border: 2px inset #999` |
| `searchbar-icon` | `[data-part='searchbar-icon']` | Magnifying glass emoji, `padding: 0 6px` |
| `searchbar-input` | `[data-part='searchbar-input']` | `border: none`, `outline: none`, `font-size: 12px` |

**CSS File:** `web/src/components/SearchBar/styles/searchbar.css`

---

#### `TypeFilter` — Section Type Filter Buttons

**File:** `web/src/components/TypeFilter/TypeFilter.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `FilterValue` | (required) | Currently selected filter |
| `onChange` | `(value: FilterValue) => void` | (required) | Called when a button is clicked |

**Filter Values:** `'All' | 'GeneralTopic' | 'Example' | 'Application' | 'Tutorial'`

**Button Labels:**
| Value | Label |
|-------|-------|
| `All` | All |
| `GeneralTopic` | Topic |
| `Example` | Example |
| `Application` | App |
| `Tutorial` | Tutorial |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `typefilter` | `[data-part='typefilter']` | `display: flex`, `gap: 4px` |
| `typefilter-button` | `[data-part='typefilter-button']` | `padding: 2px 8px`, `font-size: 10px`, `border: 1px solid #999` |

**Active State:** `[aria-pressed='true']` → `font-weight: 700`, `border: 2px solid #000`, `background: #000`, `color: #fff`

**CSS File:** `web/src/components/TypeFilter/styles/typefilter.css`

---

#### `SectionList` — Scrollable List Container

**File:** `web/src/components/SectionList/SectionList.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `sections` | `SectionSummary[]` | (required) | Array of sections to display |
| `activeSlug` | `string \| null` | (required) | Currently selected slug (for highlighting) |
| `onSelect` | `(slug: string) => void` | (required) | Called when a card is clicked |

**Renders:** `SectionCard` for each section in `sections`.

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `section-list` | `[data-part~='section-list']` | `flex: 1`, `overflow-y: auto` |
| `section-list-item` | `[data-part~='section-list-item']` | Individual card wrapper |

**CSS File:** `web/src/components/SectionList/styles/section-list.css`

---

#### `SectionCard` — Individual Section Summary Card

**File:** `web/src/components/SectionList/SectionCard.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `section` | `SectionSummary` | (required) | Section data to display |
| `isActive` | `boolean` | (required) | Whether this card is selected |
| `onClick` | `() => void` | (required) | Click handler |

**Displays:**
- Type `Badge` (top-left)
- `◆ TOP` indicator (if `section.isTopLevel`)
- Title (`stripMarkdown(section.title)`)
- Short description (`stripMarkdown(section.short)`, clamped to 2 lines)

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `section-card` | `[data-part~='section-card']` | Card root (combined with `section-list-item`) |
| `section-card-meta` | `[data-part~='section-card-meta']` | Badge row, `display: flex`, `gap: 6px` |
| `section-card-top-badge` | `[data-part~='section-card-top-badge']` | `◆ TOP` text, `font-size: 9px`, `color: #999` |
| `section-card-title` | `[data-part~='section-card-title']` | `font-weight: 700`, `font-size: 12px` |
| `section-card-short` | `[data-part~='section-card-short']` | `font-size: 10px`, `color: #777`, 2-line clamp |

**Active State:** `[aria-selected='true']` → `background: #000`, `color: #fff`

**Odd rows:** `:nth-child(odd)` → `background: #f5f5f5`

**CSS File:** `web/src/components/SectionList/styles/section-list.css`

---

#### `StatusBar` — Footer Count Bar

**File:** `web/src/components/StatusBar/StatusBar.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `count` | `number` | (required) | Number of sections shown |
| `version` | `string` | `'v0.1'` | Version string (right side) |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `statusbar` | `[data-part='statusbar']` | `border-top: 2px solid #000`, `padding: 5px 10px`, `display: flex`, `justify-content: space-between` |
| `statusbar-count` | `[data-part='statusbar-count']` | `font-size: 10px` |
| `statusbar-version` | `[data-part='statusbar-version']` | `font-size: 10px` |

**CSS File:** `web/src/components/StatusBar/styles/statusbar.css`

---

#### `SectionView` — Full Section Content Viewer

**File:** `web/src/components/SectionView/SectionView.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `section` | `SectionDetail` | (required) | Full section data (including `content`) |

**Renders:**
- `SectionHeader` (title, slug, tags)
- `MarkdownContent` (rendered markdown body)

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `section-view` | `[data-part='section-view']` | `max-width: 680px`, `margin: 0 auto`, `padding: 28px 32px 60px` |
| `section-view-body` | `[data-part='section-view-body']` | `border-top: 2px solid #000`, `padding-top: 20px` |

**CSS File:** `web/src/components/SectionView/styles/section-view.css`

---

#### `SectionHeader` — Title, Slug, and Tag Badges

**File:** `web/src/components/SectionView/SectionHeader.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `section` | `SectionDetail` | (required) | Full section data |

**Displays:**
- Slug (monospace pill: `section.slug`)
- Title (`h1`: `section.title`)
- Short description (`p`: `section.short`)
- Tags: type `Badge`, topic `Badge`s, command `Badge`s, flag `Badge`s
- Flags are capped at 4 visible + `+N` overflow indicator

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `section-header` | `[data-part='section-header']` | `margin-bottom: 24px` |
| `section-header-slug` | `[data-part='section-header-slug']` | `font-size: 10px`, `font-family: var(--font-mono)`, `background: #f0f0f0`, `border: 1px solid #ccc` |
| `section-header-heading` | `[data-part='section-header-heading']` | `font-size: 24px`, `font-weight: 700` |
| `section-header-subtitle` | `[data-part='section-header-subtitle']` | `font-size: 12px`, `color: #555` |
| `section-header-tags` | `[data-part='section-header-tags']` | `display: flex`, `gap: 5px`, `flex-wrap: wrap` |

**CSS File:** `web/src/components/SectionView/styles/section-view.css`

---

#### `MarkdownContent` — Markdown Renderer

**File:** `web/src/components/Markdown/MarkdownContent.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `content` | `string` | (required) | Raw markdown string |

**Renderer:** `react-markdown` with `remark-gfm` plugin (GitHub Flavored Markdown).

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `markdown-content` | `[data-part='markdown-content']` | None (inherits from `section-view`) |

**CSS File:** `web/src/components/Markdown/styles/markdown.css`

---

#### `EmptyState` — Placeholder When No Section Selected

**File:** `web/src/components/EmptyState/EmptyState.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `label` | `string` | `'Select a section from the list.'` | Placeholder text |

**Displays:** Open book emoji (`📖`) + label text.

**CSS `data-part` attributes:**
| Part | Selector |
|------|----------|
| `empty-state` | `[data-part='empty-state']` |
| `empty-state-icon` | `[data-part='empty-state-icon']` |
| `empty-state-label` | `[data-part='empty-state-label']` |

**CSS File:** `web/src/components/EmptyState/styles/empty-state.css`

---

#### `Badge` — Coloured Tag Pill

**File:** `web/src/components/Badge/Badge.tsx`

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `text` | `string` | (required) | Label text |
| `variant` | `BadgeVariant` | `'topic'` | `'type' \| 'topic' \| 'command' \| 'flag'` |

**Type Badge Colours:**
| Section Type | Label | Color |
|--------------|-------|-------|
| `GeneralTopic` | Topic | `#4a7c59` |
| `Example` | Example | `#b8860b` |
| `Application` | App | `#4a6a8c` |
| `Tutorial` | Tutorial | `#8c4a6a` |

**Other Variant Colours:**
| Variant | Color |
|---------|-------|
| `topic` | `#000` |
| `command` | `#4a7c59` |
| `flag` | `#8c4a6a` |

**CSS `data-part` attributes:**
| Part | Selector | Styles |
|------|----------|--------|
| `badge` | `[data-part='badge']` | `display: inline-block`, `padding: 1px 7px`, `border: 1.5px solid var(--badge-color)`, `border-radius: 2px`, `font-size: 10px` |

**CSS Variables:** `--badge-color`, `--badge-weight`

**CSS File:** `web/src/components/Badge/styles/badge.css`

---

### State Management (Redux / RTK Query)

**Store File:** `web/src/store.ts`

**Current Slices:**
| Slice | Reducer Path | Purpose |
|-------|-------------|---------|
| `helpApi` | `'helpApi'` | RTK Query auto-generated reducer for API caching |

**No custom slices exist yet.** The design proposes adding:
- `uiSlice` — dark mode, sidebar visibility, active screen
- `searchSlice` — query string, search history, active filters
- `offlineSlice` — cached section IDs, sync status

**API Endpoints (RTK Query):**
| Endpoint | Hook | Returns | Cache Tags |
|----------|------|---------|------------|
| `GET /api/health` | `useHealthCheckQuery` | `HealthResponse` | — |
| `GET /api/sections` | `useListSectionsQuery` | `ListSectionsResponse` | `Section` list + per-slug |
| `GET /api/sections/:slug` | `useGetSectionQuery` | `SectionDetail` | `Section:{slug}` |

**API File:** `web/src/services/api.ts`

---

### TypeScript Data Types

**File:** `web/src/types/index.ts`

```typescript
interface SectionSummary {
  id: number;
  slug: string;
  type: string;        // "GeneralTopic" | "Example" | "Application" | "Tutorial"
  title: string;
  short: string;
  topics: string[];
  isTopLevel: boolean;
}

interface SectionDetail extends SectionSummary {
  flags: string[];
  commands: string[];
  content: string;     // Full markdown body
}

interface ListSectionsResponse {
  sections: SectionSummary[];
  total: number;
  limit: number;
  offset: number;
}

interface HealthResponse {
  ok: boolean;
  sections: number;
}

type SectionType = 'GeneralTopic' | 'Example' | 'Application' | 'Tutorial';
type FilterValue = 'All' | SectionType;
```

---

### Global CSS Variables and Theme

**File:** `web/src/styles/global.css`

**Layout Variables:**
| Variable | Value | Purpose |
|----------|-------|---------|
| `--layout-sidebar-width` | `280px` | Sidebar fixed width |

**Typography Variables:**
| Variable | Value |
|----------|-------|
| `--font-ui` | `'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif` |
| `--font-mono` | `'Monaco', 'Courier New', monospace` |
| `--font-size-sm` | `11px` |
| `--font-size-base` | `13px` |
| `--font-size-lg` | `15px` |
| `--font-size-xl` | `18px` |

**Color Variables:**
| Variable | Value | Purpose |
|----------|-------|---------|
| `--color-bg` | `#ffffff` | Page background |
| `--color-fg` | `#000000` | Text color |
| `--color-sidebar-bg` | `#ffffff` | Sidebar background |
| `--color-sidebar-fg` | `#000000` | Sidebar text |
| `--color-accent` | `#000000` | Accent / selection |
| `--color-border` | `#000000` | Borders |
| `--color-selection-bg` | `#000000` | Selection background |
| `--color-selection-fg` | `#ffffff` | Selection text |

**Aesthetic:** Classic Mac / System 7. Background is a grey dot pattern (`#a8a8a8` with 2px black dots). Windows have thick black borders (`2px solid #000`) and hard drop shadows (`box-shadow: 2px 2px 0 #000`).

---

### File Tree (Web Frontend)

```
web/src/
├── main.tsx                 # Entry point: ReactDOM.createRoot + HashRouter + Provider
├── App.tsx                  # Root component: state, filtering, layout wiring
├── store.ts                 # Redux store: RTK Query reducer only
├── styles/
│   └── global.css           # CSS variables, resets, scrollbar, selection styles
├── types/
│   └── index.ts             # TypeScript interfaces mirroring Go types
├── services/
│   └── api.ts               # RTK Query API slice (3 endpoints)
├── components/
│   ├── AppLayout/
│   │   ├── AppLayout.tsx    # Two-pane layout container
│   │   ├── parts.ts         # data-part constants
│   │   └── styles/
│   │       └── app-layout.css
│   ├── TitleBar/
│   │   ├── TitleBar.tsx     # Window title bar with stripes
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── titlebar.css
│   ├── MenuBar/
│   │   ├── MenuBar.tsx      # Retro menu bar (decorative, unused)
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── menubar.css
│   ├── SearchBar/
│   │   ├── SearchBar.tsx    # Search input with magnifying glass
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── searchbar.css
│   ├── TypeFilter/
│   │   ├── TypeFilter.tsx   # Filter button row
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── typefilter.css
│   ├── SectionList/
│   │   ├── SectionList.tsx  # Scrollable list container
│   │   ├── SectionCard.tsx  # Individual card
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── section-list.css
│   ├── SectionView/
│   │   ├── SectionView.tsx  # Full content viewer
│   │   ├── SectionHeader.tsx # Title, slug, tags
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── section-view.css
│   ├── Markdown/
│   │   ├── MarkdownContent.tsx # react-markdown wrapper
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── markdown.css
│   ├── EmptyState/
│   │   ├── EmptyState.tsx   # Placeholder when nothing selected
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── empty-state.css
│   ├── StatusBar/
│   │   ├── StatusBar.tsx    # Bottom count bar
│   │   ├── parts.ts
│   │   └── styles/
│   │       └── statusbar.css
│   └── Badge/
│       ├── Badge.tsx        # Coloured tag pill
│       ├── parts.ts
│       └── styles/
│           └── badge.css
```

---

## Needs and Affordances Analysis

An **affordance** is what a user interface "offers" — what actions it makes possible. This section lists every need we want the help browser to serve, then maps each need to concrete affordances.

### Need 1: Find Information Quickly

**User goal:** "I need to learn about the `json` command's output flags."

**Current affordances:**
- Terminal: `glaze help --query "json AND flag:output"` or `glaze help json`
- Web: Type "json" in search box, scroll list, click section

**Missing affordances:**
- Full-text search inside markdown content (not just titles/topics)
- Search result snippets showing context around matches
- Keyboard shortcut (`/`) to focus search from anywhere
- Search history / recent queries

### Need 2: Explore Related Content

**User goal:** "I'm reading about `json` command help. What examples exist? What tutorials?"

**Current affordances:**
- Terminal: None (must run new queries)
- Web: None (must manually search)

**Missing affordances:**
- "See Also" panel showing sections with overlapping topics/commands/flags
- Bidirectional cross-references (if A links to B, B knows about A)
- Topic cluster visualization

### Need 3: Browse by Category

**User goal:** "Show me all tutorials." or "Show me everything about `database`."

**Current affordances:**
- Terminal: `glaze help --tutorials` or `glaze help --topic database`
- Web: Type filter buttons (Topic, Example, App, Tutorial) — but only one at a time, and no topic browsing

**Missing affordances:**
- Topic index page with tag cloud
- Multi-select type filtering (e.g., "Examples AND Tutorials")
- Hierarchical topic browsing

### Need 4: Assess Documentation Coverage

**User goal:** "As a maintainer, I want to see which commands lack help sections."

**Current affordances:**
- None

**Missing affordances:**
- Coverage dashboard showing commands vs. documented commands
- Coverage percentage by section type
- "Undocumented commands" list

### Need 5: Read Comfortably

**User goal:** "I'm reading a long tutorial on my laptop at night."

**Current affordances:**
- Web: Basic markdown rendering

**Missing affordances:**
- Dark mode toggle
- Font size adjustment
- Table of contents for long sections
- Anchor links to headings
- Print-friendly layout

### Need 6: Access Documentation Offline

**User goal:** "I'm on a plane and need to refer to the glazed docs."

**Current affordances:**
- Static export (`render-site`) works offline if files are saved locally
- Live server (`serve-help`) requires network

**Missing affordances:**
- Service worker caching all sections after first visit
- "Save for offline" per-section toggle
- Offline indicator when disconnected

### Need 7: Share and Reference

**User goal:** "I want to send a link to a specific paragraph in the docs."

**Current affordances:**
- Web: Links to sections (`#/sections/help-system`)

**Missing affordances:**
- Anchor links to headings within sections (`#/sections/help-system#query-dsl`)
- Copy-link-to-heading button on hover
- Social sharing metadata (OpenGraph tags)

### Need 8: Export from the Browser

**User goal:** "I want to download this section as a markdown file."

**Current affordances:**
- Terminal: Not applicable
- Web: None (user would copy-paste)

**Missing affordances:**
- "Download as Markdown" button
- "Print this page" with custom print styles

---

## User Stories

Based on the needs analysis, here are detailed user stories organized by persona.

### Persona: New User (Developer)

> **Story 1.1:** As a new developer using `glaze`, I want to search for "json output format" and see results from section titles, short descriptions, topics, and full content — so that I can find relevant documentation even if I don't know the exact command name.

> **Story 1.2:** As a new developer, I want search results to show a snippet of text around the matching term — so that I can decide which result is relevant without clicking each one.

> **Story 1.3:** As a new developer, I want to press `/` anywhere in the web app to focus the search box — so that I don't have to reach for my mouse.

### Persona: Power User (Maintainer)

> **Story 2.1:** As a maintainer, I want to see a "See Also" panel when viewing a section — so that I can discover related examples, tutorials, and topics without running separate searches.

> **Story 2.2:** As a maintainer, I want a topic browser that shows all topics as a tag cloud — so that I can explore the documentation graphically and find under-documented areas.

> **Story 2.3:** As a maintainer, I want a coverage dashboard showing which CLI commands have zero help sections — so that I can prioritize documentation work.

> **Story 2.4:** As a maintainer, I want to use the query DSL in the web UI (`type:example AND topic:database`) — so that I can perform complex filtered searches the same way I do in the terminal.

### Persona: End User (Operator)

> **Story 3.1:** As an operator reading docs at night, I want to toggle dark mode — so that the bright white background does not strain my eyes.

> **Story 3.2:** As an operator reading a long tutorial, I want a table of contents sidebar showing all headings — so that I can jump to specific sections without scrolling.

> **Story 3.3:** As an operator, I want to click a link icon that appears when hovering over a heading — so that I can copy a deep link to that exact paragraph.

> **Story 3.4:** As an operator, I want to download the current section as a markdown file — so that I can save it to my notes or include it in a runbook.

### Persona: Offline User

> **Story 4.1:** As a user who has previously visited the help site, I want the site to work when I disconnect from the network — so that I can continue reading cached documentation.

> **Story 4.2:** As an offline user, I want a visual indicator showing which content is available offline — so that I know whether I can trust the page to load.

---

## Screen Designs with ASCII Mockups

This section describes every screen in the enhanced help browser, with ASCII art showing layout and widgets.

### Screen 1: Home / Browse View (Enhanced)

The current home view is a two-pane layout. The enhanced version adds a **command palette** (`Cmd+K` or `Ctrl+K`), keyboard shortcuts, and a **coverage indicator** in the status bar.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │  ← MenuBar
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Sidebar (320px) ──────────────────────┐  ┌─ Content ──────────────────┐  │
│  │                                         │  │                           │  │
│  │  📁 Sections                    [?]    │  │                           │  │
│  │  ┌─────────────────────────────────┐   │  │    Select a section       │  │
│  │  │ 🔍 Search...            ⌘K     │   │  │    from the sidebar       │  │
│  │  └─────────────────────────────────┘   │  │    to view its content.   │  │
│  │                                         │  │                           │  │
│  │  [All] [Topic] [Example] [App] [Tut]   │  │    Press ⌘K to search     │  │
│  │                                         │  │    across all sections.   │  │
│  │  ┌─ SectionCard (active) ──────────┐   │  │                           │  │
│  │  │ [Topic]  ◆ TOP                   │   │  │                           │  │
│  │  │ Help System                      │   │  │                           │  │
│  │  │ Overview of the glazed help sys… │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │  ┌─ SectionCard ───────────────────┐   │  │                           │  │
│  │  │ [Example]                        │   │  │                           │  │
│  │  │ help-example-1                   │   │  │                           │  │
│  │  │ Show the list of all toplevel... │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │  ┌─ SectionCard ───────────────────┐   │  │                           │  │
│  │  │ [GeneralTopic]                   │   │  │                           │  │
│  │  │ Markdown Style                   │   │  │                           │  │
│  │  │ Guide to markdown formatting...  │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │                                         │  │                           │  │
│  │  ─── 24 sections ───                    │  │                           │  │
│  │                                         │  │                           │  │
│  └─────────────────────────────────────────┘  └───────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **MenuBar:** Retro Macintosh-style menu. Currently decorative. Enhancement: make "View" functional (toggle dark mode, toggle sidebar, zoom).
- **SearchBar:** Text input with `⌘K` shortcut hint. Enhancement: pressing `⌘K` or `/` focuses it; supports DSL queries.
- **TypeFilter:** Toggle buttons. Enhancement: allow multi-select (e.g., Example + Tutorial).
- **SectionList + SectionCard:** Scrollable list. Enhancement: virtual scrolling for large lists (>500 sections).
- **StatusBar:** Shows section count. Enhancement: add coverage percentage and offline indicator.

### Screen 2: Section View with Cross-References

When a section is selected, the content pane shows the full markdown. The enhanced version adds a **right sidebar** for cross-references and a **table of contents**.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Sidebar (280px) ──────────────────┐  ┌─ Content ─────────────────────┐  │
│  │                                     │  │  📄 Help System — glaze help  │  │
│  │  🔍 json                            │  │     help-system               │  │
│  │                                     │  │                               │  │
│  │  [All] [Topic] [Example] [App]      │  │  Type: GeneralTopic           │  │
│  │                                     │  │  Topics: help, documentation  │  │
│  │  ┌─ SectionCard ─────────────────┐  │  │  Commands: help               │  │
│  │  │ [Example]  ◆ TOP              │  │  │                               │  │
│  │  │ JSON Output Example            │  │  │  ┌─ Table of Contents ─────┐  │  │
│  │  │ How to format JSON output...   │  │  │  │ 1. Overview             │  │  │
│  │  └────────────────────────────────┘  │  │  │ 2. Query DSL            │  │  │
│  │  ┌─ SectionCard (active) ───────┐  │  │  │ 3. See Also             │  │  │
│  │  │ [GeneralTopic]                │  │  │  └─────────────────────────┘  │  │
│  │  │ Help System  ← active         │  │  │                               │  │
│  │  │ Overview of the glazed...     │  │  │  The help system allows...    │  │
│  │  └────────────────────────────────┘  │  │                               │  │
│  │                                     │  │  ## Overview                  │  │
│  │                                     │  │                               │  │
│  │                                     │  │  The Glazed help system...    │  │
│  │                                     │  │                               │  │
│  └─────────────────────────────────────┘  │                               │  │
│                                           │  ## Query DSL                 │  │
│  ┌─ Right Sidebar (240px) ─────────────┐  │                               │  │
│  │                                      │  │  The query DSL supports...    │  │
│  │  🔗 See Also                         │  │                               │  │
│  │  ──────────────────────────────────  │  │  ## See Also                  │  │
│  │  By Topic:                           │  │                               │  │
│  │  • markdown-style (Topic)            │  │  • markdown-style             │  │
│  │  • writing-help-entries (Tutorial)   │  │  • writing-help-entries       │  │
│  │                                      │  │  • serve-help-over-http       │  │
│  │  By Command:                         │  │                               │  │
│  │  • help --ui (TUI)                   │  │                               │  │
│  │  • help --query (DSL)                │  │                               │  │
│  │                                      │  │                               │  │
│  │  By Flag:                            │  │                               │  │
│  │  • --long-help                       │  │                               │  │
│  │                                      │  │                               │  │
│  │  [📥 Download Markdown]              │  │                               │  │
│  │  [🖨️ Print]                          │  │                               │  │
│  └──────────────────────────────────────┘  └───────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **SectionHeader:** Title, slug, type badge, topic/command/flag tags. Enhancement: add "Copy link" button.
- **TableOfContents:** Extracted from markdown headings (`##`, `###`). Clicking scrolls to heading. Updates URL hash.
- **MarkdownContent:** Rendered markdown. Enhancement: heading hover shows anchor link icon; code blocks have copy button.
- **CrossReferencePanel (right sidebar):**
  - "By Topic" — sections sharing any topic with current section
  - "By Command" — sections documenting the same commands
  - "By Flag" — sections documenting the same flags
  - "Download Markdown" — reconstructs and downloads the `.md` file
  - "Print" — triggers browser print with custom styles

### Screen 3: Search Results with Snippets

When the user types a query, the sidebar transforms into a search results list with highlighted snippets.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Sidebar (320px) ──────────────────────┐  ┌─ Content ──────────────────┐  │
│  │                                         │  │                           │  │
│  │  🔍 "json output"               [×]    │  │    Select a result        │  │
│  │  ┌─────────────────────────────────┐   │  │    to view details.       │  │
│  │  │ 3 results in 24 sections       │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │                                         │  │                           │  │
│  │  ┌─ SearchResult ──────────────────┐   │  │                           │  │
│  │  │ [Example]                        │   │  │                           │  │
│  │  │ JSON Output Example              │   │  │                           │  │
│  │  │ ...how to use the <mark>json</mark> │   │  │                           │  │
│  │  │ command with <mark>output</mark>... │   │  │                           │  │
│  │  │                                  │   │  │                           │  │
│  │  │ Score: 92%  •  Topics: json, ex… │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │                                         │  │                           │  │
│  │  ┌─ SearchResult ──────────────────┐   │  │                           │  │
│  │  │ [GeneralTopic]                   │   │  │                           │  │
│  │  │ Help System                      │   │  │                           │  │
│  │  │ ...the <mark>json</mark> formatter│   │  │                           │  │
│  │  │ supports custom <mark>output</mark>│   │  │                           │  │
│  │  │ templates...                     │   │  │                           │  │
│  │  │                                  │   │  │                           │  │
│  │  │ Score: 67%  •  Topics: help      │   │  │                           │  │
│  │  └─────────────────────────────────┘   │  │                           │  │
│  │                                         │  │                           │  │
│  └─────────────────────────────────────────┘  └───────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **SearchBar:** Shows current query with clear button (`×`).
- **ResultCount:** "N results in M sections."
- **SearchResult:** Enhanced `SectionCard` with:
  - Highlighted snippet showing matching text with `<mark>` tags
  - Relevance score (if FTS ranking is available)
  - Topic/command/flag metadata chips

### Screen 4: Topic Browser

A dedicated screen accessible via the menu or a "Browse Topics" button. Shows all topics as a tag cloud and lets users drill into topic clusters.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Content (full width) ─────────────────────────────────────────────────┐  │
│  │                                                                         │  │
│  │  🏷️ Topic Browser                                                       │  │
│  │                                                                         │  │
│  │  ┌─ Tag Cloud ───────────────────────────────────────────────────────┐  │  │
│  │  │                                                                    │  │  │
│  │  │   database    json      csv      templates    markdown            │  │  │
│  │  │      help    output    flags      commands      yaml              │  │  │
│  │  │   sorting    tables    http       server        ui                │  │  │
│  │  │                                                                    │  │  │
│  │  │  (font size = number of sections tagged)                          │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                         │  │
│  │  ┌─ Selected Topic: "json" ──────────────────────────────────────────┐  │  │
│  │  │                                                                    │  │  │
│  │  │  5 sections tagged with "json":                                    │  │  │
│  │  │                                                                    │  │  │
│  │  │  ┌─ SectionCard ───────────────────────────────────────────────┐  │  │  │
│  │  │  │ [Example]  JSON Output Example                                │  │  │  │
│  │  │  │ [GeneralTopic]  JSON Command Reference                        │  │  │  │
│  │  │  │ [Tutorial]  Working with JSON Data                            │  │  │  │
│  │  │  │ [Application]  JSON Pipeline Tutorial                         │  │  │  │
│  │  │  │ [Example]  JSON to CSV Conversion                             │  │  │  │
│  │  │  └──────────────────────────────────────────────────────────────┘  │  │  │
│  │  │                                                                    │  │  │
│  │  │  Related topics: output, csv, yaml, templates                      │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                         │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **TagCloud:** Each topic is a clickable tag. Font size proportional to section count. Color indicates density.
- **TopicDetailPanel:** Appears when a topic is clicked. Shows all sections with that topic, grouped by type.
- **RelatedTopics:** Shows topics that co-occur with the selected topic (Jaccard similarity or simple co-occurrence count).

### Screen 5: Coverage Dashboard

A maintainer-focused screen showing documentation coverage metrics.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Content (full width) ─────────────────────────────────────────────────┐  │
│  │                                                                         │  │
│  │  📊 Documentation Coverage Dashboard                                   │  │
│  │                                                                         │  │
│  │  ┌─ Metrics ─────────────────────────────────────────────────────────┐  │  │
│  │  │  Total sections:        24                                         │  │  │
│  │  │  General Topics:        8  (████░░░░░░  33%)                       │  │  │
│  │  │  Examples:              6  (██████░░░░  25%)                       │  │  │
│  │  │  Applications:          4  (████░░░░░░  17%)                       │  │  │
│  │  │  Tutorials:             6  (██████░░░░  25%)                       │  │  │
│  │  │                                                                      │  │  │
│  │  │  Commands documented:   12 / 18  (████████████░░░░░░  67%)         │  │  │
│  │  │  Flags documented:      24 / 45  (████████████░░░░░░░░░░░░  53%)   │  │  │
│  │  │  Top-level sections:    10 / 24  (████████████░░░░░░  42%)         │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                         │  │
│  │  ┌─ Undocumented Commands ────────────────────────────────────────────┐  │  │
│  │  │  These commands have no associated help sections:                    │  │  │
│  │  │                                                                      │  │  │
│  │  │  • yaml     • csv      • html      • markdown                       │  │  │
│  │  │                                                                      │  │  │
│  │  │  [Create help section for yaml]  [Create help section for csv]      │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                         │  │
│  │  ┌─ Topic Coverage Heatmap ───────────────────────────────────────────┐  │  │
│  │  │                                                                      │  │  │
│  │  │  database ████████░░  json ██████████  csv ████░░░░░░               │  │  │
│  │  │  templates ██████░░░░  http █████░░░░░░  ui ██████████              │  │  │
│  │  │                                                                      │  │  │
│  │  └────────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                         │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **CoverageMetrics:** Bar charts showing section type distribution and coverage ratios.
- **UndocumentedCommandsList:** List of commands with zero associated help sections. Each has a "Create" button that generates a template markdown file.
- **TopicHeatmap:** Grid showing topic density. Darker = more sections.

### Screen 6: Command Palette (Modal)

Triggered by `Cmd+K` or `Ctrl+K`. A fuzzy-search modal that overlays any screen.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ... rest of app dimmed ...                                                 │
│                                                                             │
│         ┌─ Command Palette ──────────────────────────────────────┐          │
│         │                                                        │          │
│         │  🔍  Type a command or search...                       │          │
│         │  ────────────────────────────────────────────────────  │          │
│         │  Recent                                               │          │
│         │  → JSON Output Example                                │          │
│         │  → Help System                                        │          │
│         │  → Markdown Style                                     │          │
│         │                                                        │          │
│         │  Commands                                             │          │
│         │  → Go to Topic Browser      ⌘T                        │          │
│         │  → Go to Coverage Dashboard ⌘D                        │          │
│         │  → Toggle Dark Mode         ⌘⇧D                       │          │
│         │  → Toggle Sidebar           ⌘B                        │          │
│         │                                                        │          │
│         │  Search Results                                         │          │
│         │  → JSON Output Example      Example  •  json, output   │          │
│         │  → JSON Command Reference   Topic    •  json, help     │          │
│         │                                                        │          │
│         │  ─── 2 results ───                                    │          │
│         │                                                        │          │
│         └────────────────────────────────────────────────────────┘          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Widgets on this screen:**
- **CommandPaletteInput:** Fuzzy-search input. Supports `>` prefix for commands (e.g., `>dark` for "Toggle Dark Mode").
- **RecentList:** Last 5 visited sections.
- **CommandList:** App-level commands (navigate to screens, toggle features).
- **ResultList:** Fuzzy-matched sections from the full list.

### Screen 7: Dark Mode

The entire app supports a dark color scheme toggled via View menu or `Cmd+Shift+D`.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│    File  Edit  View  Help                              Glazed Help Browser │
│  (dark blue background, light text)                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─ Sidebar ──────────────────────────────┐  ┌─ Content ──────────────────┐  │
│  │ (dark gray bg, white text)             │  │ (dark gray bg, white text)│  │
│  │                                         │  │                           │  │
│  │  📁 Sections                    [?]    │  │  📄 Help System            │  │
│  │  ┌─────────────────────────────────┐   │  │  (light headings, dim     │  │
│  │  │ 🔍 Search...            ⌘K     │   │  │   body text, code blocks  │  │
│  │  └─────────────────────────────────┘   │  │   with dark syntax hl)    │  │
│  │  ...                                   │  │                           │  │
│  └─────────────────────────────────────────┘  └───────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Proposed API Enhancements

The current API is minimal. The following endpoints and query parameters are proposed.

### New Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /api/sections/search` | GET | Full-text search with `?q=` query param. Returns results with `snippet` field containing highlighted excerpt. |
| `GET /api/topics` | GET | List all unique topics with section counts. |
| `GET /api/commands` | GET | List all unique commands with associated section counts. |
| `GET /api/coverage` | GET | Coverage metrics: total sections, commands documented/undocumented, flags documented/undocumented. |

### Enhanced Existing Endpoints

**`GET /api/sections/:slug`** — Add `?related=true` parameter to include a `related` array in the response:

```json
{
  "id": 1,
  "slug": "help-system",
  "title": "Help System",
  "content": "...",
  "related": {
    "by_topic": [
      {"slug": "markdown-style", "title": "Markdown Style", "type": "GeneralTopic"}
    ],
    "by_command": [
      {"slug": "help-example-1", "title": "Show the list of all toplevel topics", "type": "Example"}
    ],
    "by_flag": []
  }
}
```

### Pseudocode for New Handlers

```go
// pkg/help/server/handlers.go additions

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    q := r.URL.Query().Get("q")
    if q == "" {
        writeError(w, http.StatusBadRequest, "bad_request", "missing q parameter")
        return
    }

    // Use FTS if available, fallback to LIKE search
    var sections []*model.Section
    var err error
    if h.deps.Store.HasFTS() {
        sections, err = h.deps.Store.SearchFTS(ctx, q)
    } else {
        sections, err = h.deps.Store.Find(ctx, store.Or(
            store.TitleContains(q),
            store.ContentContains(q),
        ))
    }

    if err != nil {
        writeError(w, http.StatusInternalServerError, "internal_error", "search failed")
        return
    }

    // Generate snippets (simplified)
    results := make([]SearchResult, len(sections))
    for i, s := range sections {
        results[i] = SearchResult{
            SectionSummary: SummaryFromModel(s),
            Snippet: generateSnippet(s.Content, q),
        }
    }

    writeJSON(w, http.StatusOK, SearchResponse{
        Results: results,
        Total:   len(results),
    })
}

func (h *Handler) handleTopics(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    topics, err := h.deps.Store.ListTopics(ctx)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "internal_error", "failed to list topics")
        return
    }
    writeJSON(w, http.StatusOK, topics)
}

func (h *Handler) handleCoverage(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    total, _ := h.deps.Store.Count(ctx)
    commands, _ := h.deps.Store.ListCommands(ctx)
    flags, _ := h.deps.Store.ListFlags(ctx)

    // This requires the store to know about "all commands" — which comes from
    // cobra command introspection, not the help store. For v1, we can only
    // report "commands that have at least one section".
    writeJSON(w, http.StatusOK, CoverageResponse{
        TotalSections:      int(total),
        CommandsWithDocs:   len(commands),
        // FlagsWithDocs, etc.
    })
}
```

---

## Frontend Architecture Plan

### State Management Expansion

The current Redux store only has the RTK Query API slice. We need additional slices for:

1. **UI slice:** dark mode, sidebar visibility, active screen (browse, topic, coverage)
2. **Search slice:** query string, search history, active filters
3. **Offline slice:** cached section IDs, sync status, last fetch timestamp

```typescript
// web/src/slices/uiSlice.ts
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface UIState {
  darkMode: boolean;
  sidebarVisible: boolean;
  activeScreen: 'browse' | 'topics' | 'coverage';
}

const initialState: UIState = {
  darkMode: false,
  sidebarVisible: true,
  activeScreen: 'browse',
};

export const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleDarkMode: (state) => { state.darkMode = !state.darkMode; },
    toggleSidebar: (state) => { state.sidebarVisible = !state.sidebarVisible; },
    setActiveScreen: (state, action: PayloadAction<UIState['activeScreen']>) => {
      state.activeScreen = action.payload;
    },
  },
});
```

### Component Additions

| Component | Purpose | Screen |
|-----------|---------|--------|
| `CommandPalette` | Modal fuzzy search + commands | Global |
| `TableOfContents` | Heading list with anchor links | Section View |
| `CrossReferencePanel` | Related sections sidebar | Section View |
| `TagCloud` | Topic browser visual | Topic Browser |
| `TopicDetailPanel` | Sections for selected topic | Topic Browser |
| `CoverageDashboard` | Metrics and charts | Coverage |
| `UndocumentedList` | Commands needing docs | Coverage |
| `OfflineIndicator` | Network status badge | Global |
| `DarkModeToggle` | Theme switcher | Global |
| `HeadingAnchor` | Hover link icon on headings | Markdown Content |

### Routing Expansion

Current routes (HashRouter):
- `/#/` — Home (browse)
- `/#/sections/:slug` — Section view

Proposed routes:
- `/#/` — Home (browse)
- `/#/sections/:slug` — Section view
- `/#/sections/:slug#:anchor` — Section view scrolled to anchor
- `/#/topics` — Topic browser
- `/#/topics/:topic` — Topic detail
- `/#/coverage` — Coverage dashboard
- `/#/search?q=...` — Search results (shareable URL)

### Service Worker (Offline Support)

Use Vite PWA plugin or a custom service worker:

```javascript
// sw.js — Cache-first strategy for API responses
const CACHE_NAME = 'glazed-help-v1';

self.addEventListener('fetch', (event) => {
  if (event.request.url.includes('/api/')) {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        // Return cached if available, otherwise fetch and cache
        if (cached) return cached;
        return fetch(event.request).then((response) => {
          const clone = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, clone));
          return response;
        });
      })
    );
  }
});
```

---

## Implementation Phases

### Phase 1: Enhanced Search

**Backend:**
- Add `GET /api/sections/search?q=...` endpoint with FTS support (if build tag enabled).
- Add `snippet` generation for search results.

**Frontend:**
- Enhance `SearchBar` to support DSL queries (`type:example AND topic:database`).
- Create `SearchResult` component with snippet highlighting.
- Add `Cmd+K` shortcut and `CommandPalette` modal.

**Validation:**
- Search "json" and see results from content, not just titles.
- Search `type:example` and see only examples.

### Phase 2: Cross-References and Deep Linking

**Backend:**
- Add `?related=true` to `GET /api/sections/:slug`.
- Implement topic/command/flag co-occurrence queries in the store.

**Frontend:**
- Create `CrossReferencePanel` right sidebar.
- Create `TableOfContents` from markdown headings.
- Add anchor links to headings (`HeadingAnchor`).
- Update routing to support `#anchor` hashes.

**Validation:**
- View "Help System" section and see "See Also" panel populated.
- Click a heading anchor link and verify URL updates and scrolls.

### Phase 3: Topic Browser and Coverage Dashboard

**Backend:**
- Add `GET /api/topics` endpoint.
- Add `GET /api/coverage` endpoint.

**Frontend:**
- Create `TagCloud`, `TopicDetailPanel`, `CoverageDashboard`, `UndocumentedList`.
- Add navigation items to MenuBar (View → Topic Browser, View → Coverage).
- Add route handlers for `/#/topics` and `/#/coverage`.

**Validation:**
- Click "View → Topic Browser" and see tag cloud.
- Click a topic and see related sections.
- Visit Coverage dashboard and see bar charts.

### Phase 4: Dark Mode and Accessibility

**Frontend:**
- Add `uiSlice` with dark mode toggle.
- Implement CSS custom properties for theming (`--bg-primary`, `--text-primary`, etc.).
- Add `DarkModeToggle` to MenuBar or Command Palette.
- Add keyboard shortcuts (`Cmd+Shift+D`, `Cmd+B` for sidebar).
- Add `aria-label`, `role`, and focus management.

**Validation:**
- Toggle dark mode; verify all screens invert correctly.
- Tab through interface; verify focus rings visible.

### Phase 5: Offline Support

**Frontend:**
- Add service worker with cache-first strategy.
- Add `OfflineIndicator` component showing sync status.
- Add `offline` slice to Redux tracking cached sections.

**Validation:**
- Load site, then disconnect network.
- Verify already-visited sections still load.
- Verify offline indicator appears.

### Phase 6: Print and Export from Browser

**Frontend:**
- Add `@media print` CSS styles hiding sidebar and showing clean content.
- Add "Download Markdown" button to `CrossReferencePanel`.
- Add "Print" button triggering `window.print()`.

**Validation:**
- Click "Print" and verify browser print preview shows clean layout.
- Click "Download Markdown" and verify `.md` file downloads with correct frontmatter.

---

## Testing Strategy

### Backend Tests

- `TestHandleSearch` — assert search returns results with snippets.
- `TestHandleTopics` — assert topic list is deduplicated and counted.
- `TestHandleCoverage` — assert metrics match seeded data.
- `TestGetSectionWithRelated` — assert `related` object contains expected sections.

### Frontend Tests

- **Component tests** (Vitest + React Testing Library):
  - `SearchBar` — typing triggers search, `/` focuses input.
  - `CommandPalette` — `Cmd+K` opens, fuzzy search works, Enter selects.
  - `CrossReferencePanel` — renders related sections grouped by type.
  - `TagCloud` — clicking tag calls onChange with correct topic.
  - `DarkModeToggle` — toggles class on document body.

- **Integration tests**:
  - Search flow: type query → see results → click result → view section.
  - Topic flow: navigate to topics → click tag → see sections → click section.
  - Offline flow: load section → go offline → reload section → content still visible.

### End-to-End Tests

- Start `glaze serve-help`.
- Open browser at `http://localhost:8088`.
- Execute user stories manually and verify each acceptance criterion.

---

## Risks, Alternatives, and Open Questions

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| FTS5 not available in all builds | Medium | Medium | Fallback to `LIKE` search; document limitation |
| Large help systems slow down browser | Medium | High | Virtual scrolling, pagination, lazy loading |
| Service worker cache grows unbounded | Low | Medium | Implement cache eviction (LRU, max age) |
| Dark mode CSS is hard to maintain | Low | Low | Use CSS custom properties; test with Storybook |

### Alternatives Considered

1. **Use Next.js or Astro instead of Vite+React.**
   - Rejected: The current Vite setup is simple and the SPA is small. SSR would complicate the Go embedding story.

2. **Use a third-party documentation framework (Docusaurus, Nextra).**
   - Rejected: We need tight integration with the Glazed help section model and the ability to serve from a Go binary. Third-party frameworks are designed for file-based content.

3. **Implement search entirely client-side.**
   - Rejected: For large help systems (>1000 sections), client-side search is slow and memory-intensive. Server-side FTS is more scalable.

### Open Questions

1. **Should the coverage dashboard require cobra command introspection?**
   - The help store knows which commands *have* sections, but not which commands *exist* in the binary. We may need to pass the `cobra.Command` tree to the server for true coverage metrics.
   - **Decision:** Phase 3 v1 will show "commands with docs" only. Full coverage (commands without docs) is Phase 3 v2.

2. **Should search snippets use server-side or client-side highlighting?**
   - Server-side: more accurate, can use FTS offsets. Client-side: simpler, works with fallback search.
   - **Decision:** Server-side for FTS builds, client-side fallback for non-FTS builds.

3. **Should the service worker cache the entire help system or just visited pages?**
   - Full cache: better offline experience but large initial download. Visit-based: smaller but gaps in offline coverage.
   - **Decision:** Start with visit-based (cache sections as user visits them). Add "Download all for offline" button later.

---

## File References

### Go Backend

| File | Role |
|------|------|
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go` | HTTP server command and handler composition |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/handlers.go` | API route handlers |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/types.go` | Request/response types |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/middleware.go` | CORS middleware |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go` | SQLite store |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/query.go` | Predicate query system |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/fts5.go` | FTS5 full-text search (build tag gated) |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/web/static.go` | SPA static file handler |

### React Frontend

| File | Role |
|------|------|
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/main.tsx` | Entry point with HashRouter |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/App.tsx` | Root component |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/store.ts` | Redux store configuration |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/services/api.ts` | RTK Query API slice |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/types/index.ts` | TypeScript interfaces |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/components/AppLayout/AppLayout.tsx` | Two-pane layout |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/components/SearchBar/SearchBar.tsx` | Search input |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/components/SectionList/SectionList.tsx` | Section list |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/components/SectionView/SectionView.tsx` | Section content viewer |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/web/src/components/Markdown/MarkdownContent.tsx` | Markdown renderer |

---

## Conclusion

This design transforms the Glazed help web server from a simple list viewer into a **full-featured documentation browser**. The enhancements are grounded in concrete user stories and organized into six implementation phases. Each phase adds measurable value while building on the existing architecture:

- **Phase 1** (Search) makes finding information fast.
- **Phase 2** (Cross-references + Deep Linking) makes reading contextual.
- **Phase 3** (Topics + Coverage) makes exploration and maintenance easy.
- **Phase 4** (Dark Mode + Accessibility) makes reading comfortable for everyone.
- **Phase 5** (Offline) makes the docs available anywhere.
- **Phase 6** (Print/Export) makes the docs portable.

The ASCII mockups provide a shared visual language for designers and developers. The API contracts and pseudocode give implementers a clear path forward. The file references anchor every decision to the existing codebase.
