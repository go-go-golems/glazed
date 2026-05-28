---
title: Typography Debug Palette — Analysis, Design & Implementation Guide
doc-type: design
ticket: GL-012
topics: typography, css, frontend, debug-tooling
status: active
owners: [manuel]
intent: long-term
created: 2026-05-28
---

# Typography Debug Palette — Analysis, Design & Implementation Guide

**Ticket:** GL-012
**Date:** 2026-05-28
**Audience:** New intern joining the Glazed team — you should be able to read this document from start to finish, understand the entire typography system, and implement the debug palette without needing to ask questions.

---

## 1. What This Document Is

This is an intern-ready guide. It explains every moving part of the Glazed docs site typography, describes a new **Typography Debug Palette** feature, and provides a step-by-step implementation plan with pseudocode, ASCII wireframes, file references, and API details.

By the end you will know:

- How the current typography system works (CSS variables, component styles, font loading)
- What every relevant file does and where it lives
- What the debug palette should look like and how it behaves
- How to implement it, test it, and ship it

---

## 2. The Glazed Docs Site — Architecture Overview

The Glazed help browser is a **single-page React application** served by a **Go HTTP server**. The Go binary embeds the compiled frontend assets via `go:embed`. The system has two main halves:

### 2.1 Go Backend (the help server)

The Go server lives in `pkg/help/server/`. It:

- Reads Markdown help sections from SQLite databases, JSON files, or the filesystem
- Serves a REST API under `/api/*`:
  - `GET /api/health` — health check
  - `GET /api/packages` — list available help packages and versions
  - `GET /api/sections` — list/search sections (with filters for type, topic, package, version)
  - `GET /api/sections/:slug` — full section detail (title, content, headings, flags, commands)
- Serves the SPA for all non-API routes (fallback to `index.html`)
- Optionally proxies to an SSR sidecar for server-side rendering

**Key files:**

| File | Purpose |
|------|---------|
| `pkg/help/server/serve.go` | Main server setup, `NewServeHandler()`, route dispatch |
| `pkg/help/server/handlers.go` | HTTP handler functions for each API endpoint |
| `pkg/help/server/types.go` | Go structs for API request/response shapes |
| `pkg/help/server/middleware.go` | CORS middleware |
| `cmd/glaze/cmds/docs.go` | Cobra command that wires the serve command |

The backend is **not changed** by this feature. The debug palette is purely a frontend concern.

### 2.2 React Frontend (the web app)

The frontend lives in `web/`. It is a React 18 + TypeScript + Vite application:

- **State management:** Redux Toolkit + RTK Query (`web/src/store.ts`, `web/src/services/api.ts`)
- **Routing:** React Router v6 with BrowserRouter (`web/src/AppRoutes.tsx`)
- **Styling:** CSS files per component, using `data-part` selectors (no CSS modules, no Tailwind)
- **Build:** Vite builds to `web/dist/`, which is embedded into the Go binary by `cmd/build-web/main.go`

**Key directories:**

```
web/
├── public/
│   ├── fonts/
│   │   ├── ChicagoFLF.woff2     ← The only custom font file
│   │   └── NOTICE.md            ← Font licensing info (Apache 2.0)
│   └── site-config.js           ← Runtime config (API base URL, mode)
├── src/
│   ├── styles/global.css        ← ROOT CSS VARIABLES AND RESETS (critical file)
│   ├── App.tsx                  ← Main app component, wires everything
│   ├── AppRoutes.tsx            ← Route definitions
│   ├── store.ts                 ← Redux store (RTK Query only)
│   ├── types/index.ts           ← TypeScript interfaces (mirrors Go types.go)
│   ├── services/api.ts          ← RTK Query API slice
│   └── components/
│       ├── AppLayout/           ← Two-pane sidebar + content layout
│       ├── TitleBar/            ← Classic Mac title bar
│       ├── MenuBar/             ← Menu bar (Apple icon + menu items)
│       ├── SearchBar/           ← Search input
│       ├── PackageSelector/     ← Package/version dropdowns
│       ├── NavigationModeToggle/← Tree vs Search toggle
│       ├── TypeFilter/          ← Section type filter buttons
│       ├── DocumentationTree/   ← Tree navigation (expandable)
│       ├── SectionList/         ← Card-list navigation
│       ├── SectionView/         ← Section header + markdown body
│       ├── SectionHeader/       ← Title, slug, tags for a section
│       ├── Markdown/            ← ReactMarkdown renderer
│       ├── Badge/               ← Tag/type badge chip
│       ├── StatusBar/           ← Footer status line
│       ├── EmptyState/          ← Empty screen placeholder
│       └── PackageIndex/        ← Landing page with section cards
├── .storybook/                  ← Storybook config
├── package.json
├── vite.config.ts
└── tsconfig.json
```

### 2.3 The Component Convention

Every component follows a strict pattern:

1. **`ComponentName.tsx`** — React component with TypeScript props
2. **`parts.ts`** — Exports a `Parts` constant object for `data-part` selectors:
   ```typescript
   export const AppLayoutParts = {
     root:    'app-layout',
     sidebar: 'app-layout-sidebar',
     content: 'app-layout-content',
   } as const;
   ```
3. **`styles/component-name.css`** — Styles targeting `[data-part='...']` selectors

This convention means **CSS never uses class names** — everything is keyed by `data-part` attributes. This is how the debug palette will target and override styles.

---

## 3. Current Typography System — Deep Dive

### 3.1 The Root CSS Variables

The file `web/src/styles/global.css` defines **all** typography variables in a single `:root` block:

```css
:root {
  /* Layout */
  --layout-sidebar-width: 280px;

  /* Typography — classic Mac stack */
  --font-ui: 'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif;
  --font-mono: 'Monaco', 'Courier New', monospace;
  --font-size-sm: 11px;
  --font-size-base: 13px;
  --font-size-lg: 15px;
  --font-size-xl: 18px;

  /* Colors — classic Mac palette */
  --color-bg: #ffffff;
  --color-fg: #000000;
  --color-sidebar-bg: #ffffff;
  --color-sidebar-fg: #000000;
  --color-accent: #000000;
  --color-border: #000000;
  --color-selection-bg: #000000;
  --color-selection-fg: #ffffff;
}
```

**What this means:**

- There are exactly **2 font family stacks**: UI font (Chicago_) and mono font (Monaco)
- There are **4 size tokens**: sm (11px), base (13px), lg (15px), xl (18px)
- Colors are **monochrome** — the only colors are black, white, and grays
- The `--color-accent` is currently `#000000` (black), which means links and highlights are black

### 3.2 The Font File

The site uses a **single custom font file**: `web/public/fonts/ChicagoFLF.woff2`.

- **Source:** Vendored from the [infinite-mac](https://github.com/mihaip/infinite-mac) project
- **License:** Apache 2.0
- **Loading:** Via `@font-face` in `global.css`:
  ```css
  @font-face {
    font-family: 'Chicago_';
    src: url('/fonts/ChicagoFLF.woff2') format('woff2');
    font-display: swap;
  }
  ```
- The `font-display: swap` means text renders immediately with fallback fonts, then swaps in Chicago_ when loaded

### 3.3 How Typography Flows Through Components

Typography settings propagate through two channels:

**Channel 1: CSS variable inheritance** — Components reference `var(--font-ui)`, `var(--font-mono)`, `var(--font-size-base)`, etc. Changing the `:root` variable changes every component that uses it.

**Channel 2: Hardcoded values** — Many components have font sizes, weights, and colors hardcoded directly in their CSS files. These do NOT use variables and must be overridden individually.

Here is the **complete typography audit** — every font-related property in every component CSS file:

#### Typography Elements and Their Current Values

| Element | CSS File | font-family | font-size | font-weight | color |
|---------|----------|-------------|-----------|-------------|-------|
| `html, body` | `global.css` | `var(--font-ui)` | `var(--font-size-base)` | inherited | `var(--color-fg)` |
| `.app-root` | `global.css` | `var(--font-ui)` | `13px` | inherited | `#000` |
| Title bar title | `titlebar.css` | `var(--font-ui)` | `12px` | `700` | `#000` |
| Menu bar items | `menubar.css` | `var(--font-ui)` | `12px` | `700` | `#000` |
| Menu bar apple | `menubar.css` | — | `14px` | — | — |
| Menu bar app name | `menubar.css` | — | `11px` | `400` | — |
| Search input | `searchbar.css` | `var(--font-ui)` | `12px` | inherited | `var(--color-fg)` |
| Package selector | `package-selector.css` | `inherited` | `13px` | inherited | `#000` |
| Nav mode toggle label | `navigation-mode-toggle.css` | inherited | `13px` | inherited | `#000` |
| Nav mode toggle button | `navigation-mode-toggle.css` | `inherited` | inherited | inherited | `#000`/`#fff` |
| Type filter button | `typefilter.css` | `var(--font-ui)` | `10px` | `400`/`700` | `#000`/`#fff` |
| Doc tree row | `documentation-tree.css` | inherited | `12px` | `700` (groups) | `#111` |
| Doc tree heading | `documentation-tree.css` | inherited | `11px` | inherited | `#3f4b5a` |
| Section list item | `section-list.css` | `var(--font-ui)` | inherited | inherited | `#000`/`#fff` |
| Section card title | `section-list.css` | inherited | `12px` | `700` | `#000` |
| Section card short | `section-list.css` | inherited | `10px` | inherited | `#777`/`#aaa` |
| Section card top badge | `section-list.css` | inherited | `9px` | inherited | `#999`/`#888` |
| Badge | `badge.css` | `var(--font-ui)` | `10px` | `var(--badge-weight)` | `var(--badge-color)` |
| Status bar | `statusbar.css` | inherited | `10px` | inherited | `#777` |
| Empty state | `empty-state.css` | inherited | `13px` | inherited | `#999` |
| Section header slug | `section-view.css` | `var(--font-mono)` | `10px` | inherited | `#999` |
| Section header heading | `section-view.css` | inherited | `24px` | `700` | `#000` |
| Section header subtitle | `section-view.css` | inherited | `12px` | inherited | `#555` |
| Section view body | `section-view.css` | `var(--font-ui)` | inherited | inherited | `#000` |
| Markdown content root | `markdown.css` | inherited | `13px` | inherited | `#000` |
| Markdown h1 | `markdown.css` | inherited | `1.6em` (20.8px) | `700` | `#000` |
| Markdown h2 | `markdown.css` | inherited | `1.3em` (16.9px) | `700` | `#000` |
| Markdown h3 | `markdown.css` | inherited | `1.1em` (14.3px) | `700` | `#000` |
| Markdown p | `markdown.css` | inherited | inherited | inherited | `#000` |
| Markdown a | `markdown.css` | inherited | inherited | inherited | `var(--color-accent)` |
| Markdown inline code | `markdown.css` | `var(--font-mono)` | `0.9em` | inherited | `#000` |
| Markdown pre | `markdown.css` | `var(--font-mono)` | `12px` | inherited | `#000` |
| Markdown blockquote | `markdown.css` | inherited | inherited | inherited | `#555` |
| Markdown th | `markdown.css` | inherited | inherited | `700` | `#000` |
| Package index heading | `package-index.css` | inherited | `1.6em` | `700` | `#000` |
| Package index group | `package-index.css` | inherited | `1.2em` | `700` | `#000` |
| Package index count | `package-index.css` | inherited | `12px` | inherited | `#666` |
| Package index short | `package-index.css` | inherited | `12px` | inherited | `#555` |

### 3.4 The Problem

Currently, if you want to try a different font, or make all headings 2px bigger, or shift the body text to a lighter weight, you must:

1. Edit the `:root` variables in `global.css` (for variables-based values)
2. Edit **every individual component CSS file** for hardcoded values
3. Rebuild the Vite frontend (`pnpm build`)
4. Re-embed into the Go binary (`go generate`)
5. Restart the server
6. Refresh the browser

This is a **slow, multi-step feedback loop**. There is no way to experiment with typography in real-time. The debug palette eliminates this friction.

---

## 4. The Typography Debug Palette — Feature Specification

### 4.1 What It Is

The Typography Debug Palette is a **floating overlay panel** that appears in the docs browser when activated. It lets you modify **every typography property** for every element in real-time — no rebuild, no restart, just instant visual feedback.

### 4.2 What It Is Not

- It is NOT a theme system or a persistence layer. Changes are **ephemeral** — they reset on page refresh.
- It is NOT a production feature. It is a **developer/debug tool** activated by a keyboard shortcut or hidden button.
- It does NOT modify the Go server. It is 100% client-side.
- It does NOT add color themes beyond monochrome. The scope is: font family, font size, font weight, and monochrome color (gray shades) only.

### 4.3 Activation

- **Keyboard shortcut:** `Ctrl+Shift+T` (T for Typography) or `Cmd+Shift+T` on macOS
- **Hidden button:** A small `𝒜a` icon in the bottom-right corner of the status bar, visible only in dev mode (`import.meta.env.DEV`)
- When activated, the palette slides in from the right side of the screen

### 4.4 Layout

The palette is a floating panel approximately 320px wide, overlaid on the right side of the content pane. It does NOT push or resize any content — it floats above.

### 4.5 Element Groups

The palette organizes typography controls into **element groups**. Each group represents a category of UI elements. Within each group, you can adjust:

| Property | Control Type | Range |
|----------|-------------|-------|
| Font family | Dropdown select | List of available fonts |
| Font size | Stepper (−1 / value / +1) | 8–48px in 1px steps |
| Font weight | Dropdown select | 100, 200, 300, 400, 500, 600, 700, 800, 900 |
| Color | Stepper (shade) | `#000` → `#111` → `#222` → ... → `#999` → `#aaa` → `#bbb` → `#ccc` → `#ddd` → `#eee` → `#fff` |

The element groups are:

1. **Root / Body** — base font and size for the entire page
2. **Title Bar** — title text in both sidebar and content title bars
3. **Menu Bar** — menu items and app name
4. **Sidebar Controls** — search input, package selector, nav mode toggle, type filter
5. **Sidebar Tree** — document tree items and headings
6. **Sidebar Cards** — section list card title and short description
7. **Section Header** — slug label, main heading, subtitle, tags
8. **Markdown Prose** — paragraph text, line height
9. **Markdown Headings** — h1, h2, h3 (each individually adjustable)
10. **Markdown Code** — inline code, code blocks
11. **Markdown Extras** — blockquote, table header, link
12. **Status Bar** — status text
13. **Badges** — tag/type badges

### 4.6 Presets

The palette includes a **preset selector** at the top that loads pre-configured typography settings:

- **Classic Mac (default)** — Current values as they exist today
- **Clean Modern** — System font stack, 16px base, 400/600 weight split, softer grays
- **Dense Terminal** — Monospace everything, 12px base, minimal spacing
- **Large Print** — 18px base, larger headings, high contrast
- **Reset All** — Clears all overrides back to CSS defaults

Presets are defined as a TypeScript constant. Adding a new preset is adding one object to an array.

---

## 5. ASCII Wireframes

### 5.1 The Palette Panel (Default View)

```
┌─────────────────────────────────────────┐
│ 𝒜a Typography Palette          [×]      │
├─────────────────────────────────────────┤
│ Preset: [Classic Mac           ▾]       │
│                                         │
│ ▸ Root / Body                           │
│ ▸ Title Bar                             │
│ ▸ Menu Bar                              │
│ ▸ Sidebar Controls                      │
│ ▸ Sidebar Tree                          │
│ ▸ Sidebar Cards                         │
│ ▶ Section Header                        │
│ ▸ Markdown Prose                        │
│ ▸ Markdown Headings                     │
│ ▸ Markdown Code                         │
│ ▸ Markdown Extras                       │
│ ▸ Status Bar                            │
│ ▸ Badges                                │
│                                         │
│          [ Reset All ]                  │
└─────────────────────────────────────────┘
```

Each group is collapsible (accordion). Clicking the group name expands it to show controls.

### 5.2 Expanded Group — Section Header

```
┌─────────────────────────────────────────┐
│ 𝒜a Typography Palette          [×]      │
├─────────────────────────────────────────┤
│ Preset: [Classic Mac           ▾]       │
│                                         │
│ ▾ Section Header                        │
│   ┌─────────────────────────────────┐   │
│   │ Slug Label                       │   │
│   │ Font: [Chicago_         ▾]      │   │
│   │ Size: [ − ] 10px [ + ]          │   │
│   │ Weight: [400          ▾]        │   │
│   │ Color: [ − ] #999 [ + ]         │   │
│   └─────────────────────────────────┘   │
│   ┌─────────────────────────────────┐   │
│   │ Heading                          │   │
│   │ Font: [Chicago_         ▾]      │   │
│   │ Size: [ − ] 24px [ + ]         │   │
│   │ Weight: [700          ▾]        │   │
│   │ Color: [ − ] #000 [ + ]         │   │
│   └─────────────────────────────────┘   │
│   ┌─────────────────────────────────┐   │
│   │ Subtitle                         │   │
│   │ Font: [Chicago_         ▾]      │   │
│   │ Size: [ − ] 12px [ + ]          │   │
│   │ Weight: [400          ▾]        │   │
│   │ Color: [ − ] #555 [ + ]         │   │
│   └─────────────────────────────────┘   │
│                                         │
│ ▸ Markdown Prose                        │
│ ...                                     │
└─────────────────────────────────────────┘
```

### 5.3 Expanded Group — Markdown Headings

This group has individual sub-controls for h1, h2, h3:

```
┌─────────────────────────────────────────┐
│ ▾ Markdown Headings                     │
│   ┌─────────────────────────────────┐   │
│   │ H1                               │   │
│   │ Size: [ − ] 1.6em [ + ]        │   │
│   │ Weight: [700          ▾]        │   │
│   │ Color: [ − ] #000 [ + ]         │   │
│   └─────────────────────────────────┘   │
│   ┌─────────────────────────────────┐   │
│   │ H2                               │   │
│   │ Size: [ − ] 1.3em [ + ]        │   │
│   │ Weight: [700          ▾]        │   │
│   │ Color: [ − ] #000 [ + ]         │   │
│   └─────────────────────────────────┘   │
│   ┌─────────────────────────────────┐   │
│   │ H3                               │   │
│   │ Size: [ − ] 1.1em [ + ]        │   │
│   │ Weight: [700          ▾]        │   │
│   │ Color: [ − ] #000 [ + ]         │   │
│   └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

Note: Markdown headings use **relative em units** relative to the markdown content root font-size, not absolute px. The palette shows the effective computed value (e.g., "1.6em ≈ 20.8px") and adjusts the em value.

### 5.4 Expanded Group — Markdown Prose

```
┌─────────────────────────────────────────┐
│ ▾ Markdown Prose                        │
│   ┌─────────────────────────────────┐   │
│   │ Body Text                        │   │
│   │ Font: [Chicago_         ▾]      │   │
│   │ Size: [ − ] 13px [ + ]         │   │
│   │ Weight: [400          ▾]        │   │
│   │ Color: [ − ] #000 [ + ]         │   │
│   │ Line Height: [ − ] 1.6 [ + ]   │   │
│   └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### 5.5 Preset Selector Dropdown

```
┌─────────────────────────────────────────┐
│ Preset: [▾]                             │
│ ┌─────────────────────────────────────┐ │
│ │ ● Classic Mac (default)             │ │
│ │ ○ Clean Modern                     │ │
│ │ ○ Dense Terminal                   │ │
│ │ ○ Large Print                      │ │
│ │ ─────────────────────              │ │
│ │ ○ Reset All                        │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### 5.6 The Palette in Context

Here is how the palette looks overlaid on the actual docs browser:

```
┌──────────────────────────────────────────────────────────────────────────┐
│ ☐ 📖 Documentation                  ┃  ☐ 📄 Documentation              │
├──────────────────────────────────────┨───────────────────────────────────┤
│ Package: [glazed ▾]  Ver: [v1.3 ▾]  ┃                                   │
│ [🔍 Search documentation…       ]   ┃  slug: how-to-write-help-entries  │
│ [🌳 Tree] [🔍 Search]               ┃                                   │
│                                      ┃  How to Write Help Entries       ┃
│ ▸ General Topics                     ┃  A guide to authoring Glazed help┃
│   📄 getting-started                 ┃ ───────────────────────────────── ┃
│   📄 configuration                   ┃                                   ┃
│ ▸ Examples                           ┃  Writing good help entries is     ┃
│   📄 csv-to-table                    ┃  important for discoverability.  ┃ ┌──────────────────────────┐
│   📄 json-to-table                   ┃                                   ┃ │ 𝒜a Typography Palette [×]│
│ ▸ Tutorials                          ┃  The help system supports four   ┃ ├──────────────────────────┤
│   📄 how-to-write-help-entries       ┃  section types:                  ┃ │ Preset: [Classic Mac ▾]  │
│     # Overview                        ┃                                   ┃ │                          │
│     # Section Types                   ┃  - **GeneralTopic**: Reference   ┃ │ ▾ Markdown Prose         │
│     # Frontmatter Schema              ┃  - **Example**: Worked examples  ┃ │   Body Text              │
│                                      ┃  - **Application**: App docs     ┃ │   Font: [Chicago_ ▾]    │
│                                      ┃  - **Tutorial**: Step guides     ┃ │   Size: [−] 13px [+]    │
│                                      ┃                                   ┃ │   Weight: [400 ▾]       │
│                                      ┃  Each entry has frontmatter...   ┃ │   Color: [−] #000 [+]   │
│                                      ┃                                   ┃ │   Line Height: 1.6      │
│                                      ┃                                   ┃ │                          │
│                                      ┃                                   ┃ │ ▸ Markdown Headings     │
│                                      ┃                                   ┃ │ ▸ Markdown Code         │
│──────────────────────────────────────┨───────────────────────────────────┃ │          [ Reset All ]   │
│ 12 sections · glazed v1.3.4    𝒜a  ┃                      glazed v1.3.4 ┃ └──────────────────────────┘
└──────────────────────────────────────────────────────────────────────────┘
```

The small `𝒜a` in the status bar is the toggle button (dev mode only).

---

## 6. Technical Design

### 6.1 Architecture Decision: CSS Custom Properties Overlays

The palette will work by **injecting CSS custom property overrides** into the document root. When you change a value in the palette, it:

1. Sets a CSS variable on `document.documentElement` (the `<html>` element)
2. Component CSS that already uses `var(--font-ui)` picks up the change automatically
3. For hardcoded values, the palette injects **additional override rules** via a dynamically-created `<style>` element

This approach has several advantages:

- **Zero rebuild required** — changes are instant
- **Non-destructive** — the original CSS files are never modified
- **Easy reset** — removing the injected styles and variables restores defaults
- **Inspectable** — you can see exactly what changed in browser DevTools

### 6.2 State Management

The palette state is managed by a **new Redux slice** (not RTK Query, since there is no backend). The slice holds:

- `isOpen: boolean` — whether the palette is visible
- `activeGroup: string | null` — which accordion group is expanded
- `overrides: TypographyOverrides` — the current set of property overrides
- `activePreset: string | null` — which preset is applied (or null for manual)

#### Type Definitions

```typescript
// types/typography-palette.ts

/** A single grayscale color value. */
export type GrayColor = 
  | '#000' | '#111' | '#222' | '#333' | '#444' 
  | '#555' | '#666' | '#777' | '#888' | '#999'
  | '#aaa' | '#bbb' | '#ccc' | '#ddd' | '#eee' | '#fff';

/** Available font families for the dropdown. */
export type FontFamily = 'ui' | 'mono';
export const FONT_FAMILY_LABELS: Record<FontFamily, string> = {
  ui: 'Chicago_',
  mono: 'Monaco',
};

/** Standard font weight values. */
export type FontWeight = 100 | 200 | 300 | 400 | 500 | 600 | 700 | 800 | 900;

/** Typography properties that can be overridden for any element. */
export interface TypographyProperties {
  fontFamily?: FontFamily;
  fontSize?: number;       // in px (or em for relative elements)
  fontSizeUnit?: 'px' | 'em';
  fontWeight?: FontWeight;
  color?: GrayColor;
  lineHeight?: number;    // unitless multiplier (e.g., 1.6)
}

/** Map of element ID → override properties. */
export type TypographyOverrides = Record<string, TypographyProperties>;

/** A preset is a named collection of overrides. */
export interface TypographyPreset {
  id: string;
  label: string;
  overrides: TypographyOverrides;
}

/** Accordion group definition. */
export interface TypographyGroup {
  id: string;
  label: string;
  /** Sub-elements within this group. */
  elements: TypographyElement[];
}

/** A single adjustable element within a group. */
export interface TypographyElement {
  id: string;
  label: string;
  /** Which properties are adjustable for this element. */
  adjustable: ('fontFamily' | 'fontSize' | 'fontWeight' | 'color' | 'lineHeight')[];
  /** Default values (from the current CSS). */
  defaults: TypographyProperties;
  /** CSS selector to target this element. */
  selector: string;
  /** CSS properties to set (maps abstract property → CSS property name). */
  cssPropertyMap: Record<string, string>;
}
```

### 6.3 Element Registry

The palette needs to know about every adjustable element. This is defined as a **constant array** of `TypographyGroup` objects. Here is the pseudocode for the full registry:

```typescript
// components/TypographyPalette/element-registry.ts

import type { TypographyGroup, GrayColor, FontWeight, FontFamily } from '../../types/typography-palette';

const GRAY_COLORS: GrayColor[] = [
  '#000', '#111', '#222', '#333', '#444',
  '#555', '#666', '#777', '#888', '#999',
  '#aaa', '#bbb', '#ccc', '#ddd', '#eee', '#fff',
];

const FONT_WEIGHTS: FontWeight[] = [100, 200, 300, 400, 500, 600, 700, 800, 900];

export const TYPOGRAPHY_GROUPS: TypographyGroup[] = [
  {
    id: 'root',
    label: 'Root / Body',
    elements: [
      {
        id: 'root.body',
        label: 'Body Text',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: '.app-root',
        cssPropertyMap: {
          fontFamily: 'font-family',
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'titlebar',
    label: 'Title Bar',
    elements: [
      {
        id: 'titlebar.title',
        label: 'Title Text',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='titlebar-title']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'menubar',
    label: 'Menu Bar',
    elements: [
      {
        id: 'menubar.items',
        label: 'Menu Items',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='menubar']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'menubar.appname',
        label: 'App Name',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='menubar-title']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'sidebar-controls',
    label: 'Sidebar Controls',
    elements: [
      {
        id: 'sidebar.search',
        label: 'Search Input',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='searchbar-input']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
      {
        id: 'sidebar.packageselector',
        label: 'Package Selector',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='package-selector-root']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
      {
        id: 'sidebar.navtoggle',
        label: 'Nav Mode Toggle',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='navigation-mode-toggle-root']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
      {
        id: 'sidebar.typefilter',
        label: 'Type Filter',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='typefilter-button']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'sidebar-tree',
    label: 'Sidebar Tree',
    elements: [
      {
        id: 'tree.row',
        label: 'Document Row',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', color: '#111' },
        selector: "[data-part='documentation-tree-row']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'tree.heading',
        label: 'Heading Row',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', color: '#3f4b5a' },
        selector: "[data-part='documentation-tree-row'][data-kind='heading']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
    ],
  },
  {
    id: 'sidebar-cards',
    label: 'Sidebar Cards',
    elements: [
      {
        id: 'cards.title',
        label: 'Card Title',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part~='section-card-title']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'cards.short',
        label: 'Card Description',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', color: '#777' },
        selector: "[data-part~='section-card-short']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
    ],
  },
  {
    id: 'section-header',
    label: 'Section Header',
    elements: [
      {
        id: 'header.slug',
        label: 'Slug Label',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#999' },
        selector: "[data-part='section-header-slug']",
        cssPropertyMap: {
          fontFamily: 'font-family',
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'header.heading',
        label: 'Heading',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 24, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='section-header-heading']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'header.subtitle',
        label: 'Subtitle',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#555' },
        selector: "[data-part='section-header-subtitle']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'markdown-prose',
    label: 'Markdown Prose',
    elements: [
      {
        id: 'prose.body',
        label: 'Body Text',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'lineHeight'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000', lineHeight: 1.6 },
        selector: "[data-part='markdown-content']",
        cssPropertyMap: {
          fontFamily: 'font-family',
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
          lineHeight: 'line-height',
        },
      },
    ],
  },
  {
    id: 'markdown-headings',
    label: 'Markdown Headings',
    elements: [
      {
        id: 'headings.h1',
        label: 'H1',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.6, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h1",
        cssPropertyMap: { fontSize: 'font-size', fontWeight: 'font-weight', color: 'color' },
      },
      {
        id: 'headings.h2',
        label: 'H2',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.3, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h2",
        cssPropertyMap: { fontSize: 'font-size', fontWeight: 'font-weight', color: 'color' },
      },
      {
        id: 'headings.h3',
        label: 'H3',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.1, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h3",
        cssPropertyMap: { fontSize: 'font-size', fontWeight: 'font-weight', color: 'color' },
      },
    ],
  },
  {
    id: 'markdown-code',
    label: 'Markdown Code',
    elements: [
      {
        id: 'code.inline',
        label: 'Inline Code',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 0.9, fontSizeUnit: 'em', fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] code",
        cssPropertyMap: {
          fontFamily: 'font-family',
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
      {
        id: 'code.block',
        label: 'Code Block',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] pre",
        cssPropertyMap: {
          fontFamily: 'font-family',
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
  {
    id: 'markdown-extras',
    label: 'Markdown Extras',
    elements: [
      {
        id: 'extras.blockquote',
        label: 'Blockquote',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#555' },
        selector: "[data-part='markdown-content'] blockquote",
        cssPropertyMap: { fontSize: 'font-size', fontWeight: 'font-weight', color: 'color' },
      },
      {
        id: 'extras.link',
        label: 'Link',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] a",
        cssPropertyMap: { fontWeight: 'font-weight', color: 'color' },
      },
      {
        id: 'extras.table-header',
        label: 'Table Header',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] th",
        cssPropertyMap: { fontWeight: 'font-weight', color: 'color' },
      },
    ],
  },
  {
    id: 'statusbar',
    label: 'Status Bar',
    elements: [
      {
        id: 'statusbar.text',
        label: 'Status Text',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', color: '#777' },
        selector: "[data-part='statusbar']",
        cssPropertyMap: { fontSize: 'font-size', color: 'color' },
      },
    ],
  },
  {
    id: 'badges',
    label: 'Badges',
    elements: [
      {
        id: 'badges.badge',
        label: 'Badge',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='badge']",
        cssPropertyMap: {
          fontSize: 'font-size',
          fontWeight: 'font-weight',
          color: 'color',
        },
      },
    ],
  },
];
```

### 6.4 Presets

```typescript
// components/TypographyPalette/presets.ts

import type { TypographyPreset, TypographyOverrides } from '../../types/typography-palette';

const CLASSIC_MAC_DEFAULTS: TypographyOverrides = {};  // empty = no overrides = use CSS defaults

const CLEAN_MODERN: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222' },
  'titlebar.title':      { fontSize: 13, fontWeight: 600, color: '#111' },
  'header.heading':      { fontSize: 28, fontWeight: 700, color: '#111' },
  'header.subtitle':     { fontSize: 14, fontWeight: 400, color: '#666' },
  'prose.body':          { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222', lineHeight: 1.7 },
  'headings.h1':         { fontSize: 2.0, fontWeight: 700, color: '#111' },
  'headings.h2':         { fontSize: 1.5, fontWeight: 600, color: '#222' },
  'headings.h3':         { fontSize: 1.25, fontWeight: 600, color: '#333' },
  'code.inline':         { fontSize: 0.85, color: '#333' },
  'code.block':          { fontSize: 14, color: '#222' },
  'extras.link':         { color: '#333' },
  'extras.blockquote':  { color: '#666' },
  'statusbar.text':      { fontSize: 11, color: '#888' },
  'badges.badge':        { fontSize: 11, fontWeight: 500, color: '#333' },
};

const DENSE_TERMINAL: TypographyOverrides = {
  'root.body':           { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111' },
  'titlebar.title':      { fontSize: 11, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 16, fontWeight: 700, color: '#000' },
  'header.subtitle':     { fontSize: 11, fontWeight: 400, color: '#888' },
  'prose.body':          { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111', lineHeight: 1.4 },
  'headings.h1':         { fontSize: 1.4, fontWeight: 700, color: '#000' },
  'headings.h2':         { fontSize: 1.2, fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.0, fontWeight: 700, color: '#222' },
  'code.inline':         { fontFamily: 'mono', fontSize: 1.0, color: '#111' },
  'code.block':          { fontFamily: 'mono', fontSize: 11, color: '#111' },
  'statusbar.text':      { fontSize: 9, color: '#999' },
  'badges.badge':        { fontSize: 9, fontWeight: 700, color: '#000' },
};

const LARGE_PRINT: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000' },
  'titlebar.title':      { fontSize: 15, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 32, fontWeight: 700, color: '#000' },
  'header.subtitle':     { fontSize: 16, fontWeight: 400, color: '#333' },
  'prose.body':          { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000', lineHeight: 1.8 },
  'headings.h1':         { fontSize: 2.0, fontWeight: 700, color: '#000' },
  'headings.h2':         { fontSize: 1.5, fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.3, fontWeight: 700, color: '#111' },
  'code.inline':         { fontSize: 0.85, color: '#000' },
  'code.block':          { fontSize: 16, color: '#000' },
  'statusbar.text':      { fontSize: 12, color: '#555' },
  'badges.badge':        { fontSize: 12, fontWeight: 700, color: '#000' },
};

export const TYPOGRAPHY_PRESETS: TypographyPreset[] = [
  { id: 'classic-mac',    label: 'Classic Mac (default)', overrides: CLASSIC_MAC_DEFAULTS },
  { id: 'clean-modern',   label: 'Clean Modern',          overrides: CLEAN_MODERN },
  { id: 'dense-terminal', label: 'Dense Terminal',       overrides: DENSE_TERMINAL },
  { id: 'large-print',    label: 'Large Print',           overrides: LARGE_PRINT },
];
```

### 6.5 CSS Override Engine

When an override is applied, the palette generates CSS rules and injects them into a `<style>` element. The engine is a standalone utility:

```typescript
// components/TypographyPalette/css-override-engine.ts

import type { TypographyOverrides, TypographyElement } from '../../types/typography-palette';
import { TYPOGRAPHY_GROUPS } from './element-registry';

const STYLE_ID = 'typography-palette-overrides';

/** Font family stacks matching the FontFamily type. */
const FONT_STACKS: Record<string, string> = {
  ui: "'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif",
  mono: "'Monaco', 'Courier New', monospace",
};

/** Build a flat map of element ID → element definition. */
function buildElementMap(): Map<string, TypographyElement> {
  const map = new Map<string, TypographyElement>();
  for (const group of TYPOGRAPHY_GROUPS) {
    for (const elem of group.elements) {
      map.set(elem.id, elem);
    }
  }
  return map;
}

/** Generate a CSS rule string for a single element override. */
function generateRule(
  elementId: string,
  overrides: TypographyOverrides,
  elementMap: Map<string, TypographyElement>,
): string | null {
  const elem = elementMap.get(elementId);
  if (!elem) return null;

  const props = overrides[elementId];
  if (!props) return null;

  const declarations: string[] = [];

  if (props.fontFamily !== undefined) {
    const stack = FONT_STACKS[props.fontFamily] || FONT_STACKS['ui'];
    declarations.push(`  font-family: ${stack};`);
  }
  if (props.fontSize !== undefined) {
    const unit = props.fontSizeUnit || 'px';
    declarations.push(`  font-size: ${props.fontSize}${unit};`);
  }
  if (props.fontWeight !== undefined) {
    declarations.push(`  font-weight: ${props.fontWeight};`);
  }
  if (props.color !== undefined) {
    declarations.push(`  color: ${props.color};`);
  }
  if (props.lineHeight !== undefined) {
    declarations.push(`  line-height: ${props.lineHeight};`);
  }

  if (declarations.length === 0) return null;

  return `${elem.selector} {\n${declarations.join('\n')}\n}`;
}

/** Apply all overrides to the DOM. */
export function applyOverrides(overrides: TypographyOverrides): void {
  const elementMap = buildElementMap();
  const rules: string[] = [];

  for (const elementId of Object.keys(overrides)) {
    const rule = generateRule(elementId, overrides, elementMap);
    if (rule) rules.push(rule);
  }

  let styleEl = document.getElementById(STYLE_ID) as HTMLStyleElement | null;
  if (!styleEl) {
    styleEl = document.createElement('style');
    styleEl.id = STYLE_ID;
    document.head.appendChild(styleEl);
  }

  styleEl.textContent = rules.length > 0 ? rules.join('\n\n') : '';
}

/** Remove all overrides from the DOM. */
export function clearOverrides(): void {
  const styleEl = document.getElementById(STYLE_ID);
  if (styleEl) {
    styleEl.textContent = '';
  }
}
```

### 6.6 Redux Slice

```typescript
// store/typographyPaletteSlice.ts

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { TypographyOverrides, TypographyProperties } from '../types/typography-palette';

interface TypographyPaletteState {
  isOpen: boolean;
  activeGroup: string | null;
  activePreset: string | null;
  overrides: TypographyOverrides;
}

const initialState: TypographyPaletteState = {
  isOpen: false,
  activeGroup: null,
  activePreset: null,
  overrides: {},
};

const typographyPaletteSlice = createSlice({
  name: 'typographyPalette',
  initialState,
  reducers: {
    togglePalette(state) {
      state.isOpen = !state.isOpen;
    },
    openPalette(state) {
      state.isOpen = true;
    },
    closePalette(state) {
      state.isOpen = false;
    },
    setActiveGroup(state, action: PayloadAction<string | null>) {
      state.activeGroup = action.payload;
    },
    setPreset(state, action: PayloadAction<{ presetId: string; overrides: TypographyOverrides }>) {
      state.activePreset = action.payload.presetId;
      state.overrides = { ...action.payload.overrides };
    },
    setOverride(state, action: PayloadAction<{ elementId: string; properties: TypographyProperties }>) {
      const { elementId, properties } = action.payload;
      state.overrides[elementId] = {
        ...state.overrides[elementId],
        ...properties,
      };
      state.activePreset = null; // manual edit clears preset indicator
    },
    removeOverride(state, action: PayloadAction<string>) {
      delete state.overrides[action.payload];
      if (Object.keys(state.overrides).length === 0) {
        state.activePreset = null;
      }
    },
    resetAllOverrides(state) {
      state.overrides = {};
      state.activePreset = null;
      state.activeGroup = null;
    },
  },
});

export const {
  togglePalette,
  openPalette,
  closePalette,
  setActiveGroup,
  setPreset,
  setOverride,
  removeOverride,
  resetAllOverrides,
} = typographyPaletteSlice.actions;

export default typographyPaletteSlice.reducer;
```

### 6.7 Store Integration

The new slice is added to `store.ts` alongside the existing RTK Query reducer:

```typescript
// store.ts — updated

import { configureStore } from '@reduxjs/toolkit';
import { helpApi } from './services/api';
import typographyPaletteReducer from './store/typographyPaletteSlice';

const reducer = {
  [helpApi.reducerPath]: helpApi.reducer,
  typographyPalette: typographyPaletteReducer,  // NEW
};

export function makeStore(preloadedState?: unknown) {
  const config = {
    reducer,
    middleware: (getDefaultMiddleware: any) =>
      getDefaultMiddleware().concat(helpApi.middleware),
    ...(preloadedState ? { preloadedState } : {}),
  };
  return configureStore(config as any);
}

export const store = makeStore();
export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore['getState']>;
export type AppDispatch = AppStore['dispatch']>;
```

### 6.8 React Effect — Bridge Between Redux and DOM

A React effect watches the `overrides` state and applies CSS when it changes:

```typescript
// components/TypographyPalette/useTypographyOverrides.ts

import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { applyOverrides, clearOverrides } from './css-override-engine';

/**
 * Hook that syncs Redux typography overrides to the DOM.
 * Call once in the root App or in the palette component.
 */
export function useTypographyOverrides(): void {
  const overrides = useSelector((state: RootState) => state.typographyPalette.overrides);

  useEffect(() => {
    if (Object.keys(overrides).length === 0) {
      clearOverrides();
    } else {
      applyOverrides(overrides);
    }
  }, [overrides]);
}
```

### 6.9 Keyboard Shortcut Hook

```typescript
// components/TypographyPalette/usePaletteShortcut.ts

import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { togglePalette } from '../../store/typographyPaletteSlice';

/**
 * Hook that registers Ctrl+Shift+T (or Cmd+Shift+T on macOS) to toggle the palette.
 * Only active in dev mode.
 */
export function usePaletteShortcut(): void {
  const dispatch = useDispatch();

  useEffect(() => {
    if (!import.meta.env.DEV) return;

    const handler = (e: KeyboardEvent) => {
      if (e.key === 'T' && (e.ctrlKey || e.metaKey) && e.shiftKey) {
        e.preventDefault();
        dispatch(togglePalette());
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [dispatch]);
}
```

---

## 7. Component Structure

The palette is a new component with the following file structure:

```
web/src/components/TypographyPalette/
├── TypographyPalette.tsx           ← Main panel component
├── TypographyPalette.stories.tsx  ← Storybook story
├── parts.ts                       ← data-part constants
├── element-registry.ts            ← Element group definitions
├── presets.ts                     ← Preset definitions
├── css-override-engine.ts         ← CSS generation and injection
├── useTypographyOverrides.ts      ← Redux→DOM sync hook
├── usePaletteShortcut.ts          ← Keyboard shortcut hook
├── TypographyPaletteGroup.tsx     ← Accordion group component
├── TypographyPaletteElement.tsx   ← Per-element control row
├── FontFamilySelect.tsx           ← Font family dropdown
├── FontSizeStepper.tsx            ← Size +/- stepper
├── FontWeightSelect.tsx           ← Weight dropdown
├── ColorStepper.tsx               ← Gray shade +/- stepper
├── PresetSelector.tsx             ← Preset dropdown
└── styles/
    └── typography-palette.css     ← Palette styles
```

### 7.1 Main Component — TypographyPalette.tsx

```typescript
// Pseudocode for the main panel component

export function TypographyPalette() {
  const dispatch = useDispatch();
  const isOpen = useSelector((s: RootState) => s.typographyPalette.isOpen);
  const activeGroup = useSelector((s: RootState) => s.typographyPalette.activeGroup);
  const activePreset = useSelector((s: RootState) => s.typographyPalette.activePreset);
  const overrides = useSelector((s: RootState) => s.typographyPalette.overrides);

  useTypographyOverrides(); // keep DOM in sync

  if (!isOpen) return null;

  return (
    <div data-part={TypographyPaletteParts.root}>
      {/* Header */}
      <div data-part={TypographyPaletteParts.header}>
        <span data-part={TypographyPaletteParts.title}>𝒜a Typography Palette</span>
        <button data-part={TypographyPaletteParts.closeBtn}
                onClick={() => dispatch(closePalette())}>×</button>
      </div>

      {/* Preset Selector */}
      <PresetSelector activePreset={activePreset} overrides={overrides} />

      {/* Accordion Groups */}
      <div data-part={TypographyPaletteParts.groups}>
        {TYPOGRAPHY_GROUPS.map(group => (
          <TypographyPaletteGroup
            key={group.id}
            group={group}
            isExpanded={activeGroup === group.id}
            overrides={overrides}
            onToggle={() => dispatch(setActiveGroup(
              activeGroup === group.id ? null : group.id
            ))}
            onChange={(elementId, properties) =>
              dispatch(setOverride({ elementId, properties }))
            }
          />
        ))}
      </div>

      {/* Reset Button */}
      <div data-part={TypographyPaletteParts.footer}>
        <button data-part={TypographyPaletteParts.resetBtn}
                onClick={() => dispatch(resetAllOverrides())}>
          Reset All
        </button>
      </div>
    </div>
  );
}
```

### 7.2 Element Control Row — TypographyPaletteElement.tsx

Each adjustable property gets a row with a label and control:

```typescript
// Pseudocode for a single element's control panel

interface TypographyPaletteElementProps {
  element: TypographyElement;
  currentOverrides: TypographyProperties | undefined;
  onChange: (properties: TypographyProperties) => void;
}

export function TypographyPaletteElement({
  element,
  currentOverrides,
  onChange,
}: TypographyPaletteElementProps) {
  const effective = { ...element.defaults, ...currentOverrides };

  return (
    <div data-part={TypographyPaletteParts.element}>
      <span data-part={TypographyPaletteParts.elementLabel}>
        {element.label}
      </span>

      {element.adjustable.includes('fontFamily') && (
        <FontFamilySelect
          value={effective.fontFamily!}
          onChange={(v) => onChange({ fontFamily: v })}
        />
      )}

      {element.adjustable.includes('fontSize') && (
        <FontSizeStepper
          value={effective.fontSize!}
          unit={effective.fontSizeUnit || 'px'}
          onChange={(v) => onChange({ fontSize: v })}
        />
      )}

      {element.adjustable.includes('fontWeight') && (
        <FontWeightSelect
          value={effective.fontWeight!}
          onChange={(v) => onChange({ fontWeight: v })}
        />
      )}

      {element.adjustable.includes('color') && (
        <ColorStepper
          value={effective.color!}
          onChange={(v) => onChange({ color: v })}
        />
      )}

      {element.adjustable.includes('lineHeight') && (
        <FontSizeStepper  // reuse stepper with 0.1 step
          value={effective.lineHeight!}
          unit=""
          step={0.1}
          onChange={(v) => onChange({ lineHeight: v })}
        />
      )}
    </div>
  );
}
```

---

## 8. Palette CSS

The palette should look like it belongs in the Classic Mac aesthetic — black borders, white background, the same Chicago_ font:

```css
/* typography-palette.css */

[data-part='typography-palette'] {
  position: fixed;
  top: 8px;
  right: 8px;
  width: 320px;
  max-height: calc(100vh - 16px);
  background: #fff;
  border: 2px solid #000;
  box-shadow: 3px 3px 0 #000;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  font-family: var(--font-ui);
  font-size: 11px;
  overflow: hidden;
}

[data-part='typography-palette-header'] {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  background: #000;
  color: #fff;
  font-weight: 700;
  font-size: 12px;
}

[data-part='typography-palette-close-btn'] {
  background: none;
  border: none;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  padding: 0 4px;
}

[data-part='typography-palette-groups'] {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

/* Accordion group */
[data-part='typography-palette-group'] {
  border-bottom: 1px solid #ddd;
}

[data-part='typography-palette-group-header'] {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  cursor: pointer;
  font-weight: 700;
  font-size: 11px;
  background: #f5f5f5;
  border: none;
  width: 100%;
  text-align: left;
}

[data-part='typography-palette-group-header']:hover {
  background: #eee;
}

/* Element block within expanded group */
[data-part='typography-palette-element'] {
  padding: 6px 10px 8px;
  border-bottom: 1px solid #eee;
}

[data-part='typography-palette-element-label'] {
  font-weight: 700;
  font-size: 10px;
  display: block;
  margin-bottom: 4px;
  color: #555;
}

/* Control rows */
[data-part='typography-palette-control-row'] {
  display: grid;
  grid-template-columns: 52px 1fr;
  align-items: center;
  gap: 6px;
  margin-bottom: 3px;
}

[data-part='typography-palette-control-label'] {
  font-size: 10px;
  color: #777;
}

/* Stepper (size, color) */
[data-part='typography-palette-stepper'] {
  display: inline-flex;
  align-items: center;
  border: 1px solid #999;
  background: #fff;
}

[data-part='typography-palette-stepper-btn'] {
  width: 20px;
  height: 20px;
  border: none;
  background: #eee;
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;
  padding: 0;
}

[data-part='typography-palette-stepper-btn']:hover {
  background: #ddd;
}

[data-part='typography-palette-stepper-value'] {
  padding: 0 6px;
  font-family: var(--font-mono);
  font-size: 10px;
  min-width: 50px;
  text-align: center;
}

/* Dropdowns */
[data-part='typography-palette-select'] {
  width: 100%;
  border: 1px solid #999;
  background: #fff;
  font: inherit;
  font-size: 10px;
  padding: 2px 4px;
}

/* Footer */
[data-part='typography-palette-footer'] {
  padding: 6px 10px;
  border-top: 2px solid #000;
  text-align: center;
}

[data-part='typography-palette-reset-btn'] {
  padding: 4px 16px;
  border: 2px solid #000;
  background: #fff;
  font: inherit;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
}

[data-part='typography-palette-reset-btn']:hover {
  background: #000;
  color: #fff;
}

/* Preset selector */
[data-part='typography-palette-preset-row'] {
  padding: 6px 10px;
  border-bottom: 1px solid #ddd;
  display: grid;
  grid-template-columns: 40px 1fr;
  align-items: center;
  gap: 6px;
}

[data-part='typography-palette-preset-label'] {
  font-size: 10px;
  font-weight: 700;
}

[data-part='typography-palette-preset-select'] {
  border: 1px solid #999;
  background: #fff;
  font: inherit;
  font-size: 10px;
  padding: 2px 4px;
}
```

---

## 9. Wiring Into the App

### 9.1 App.tsx Changes

The `App` component needs two additions:

1. Render the `<TypographyPalette />` component
2. Call `usePaletteShortcut()` to register the keyboard shortcut

```typescript
// In App.tsx, add imports:
import { TypographyPalette } from './components/TypographyPalette/TypographyPalette';
import { usePaletteShortcut } from './components/TypographyPalette/usePaletteShortcut';

// Inside the App() function, add:
usePaletteShortcut();

// In the JSX, add the palette as a sibling to the app-root div:
return (
  <>
    <div className="app-root">
      <AppLayout ... />
    </div>
    <TypographyPalette />
  </>
);
```

### 9.2 StatusBar Dev Button

In `StatusBar.tsx`, add a small `𝒜a` button in dev mode:

```typescript
// In StatusBar.tsx:
import { usePaletteShortcut } from '../TypographyPalette/usePaletteShortcut';

// Inside StatusBar component:
const dispatch = useDispatch();

// In the JSX, add before the version text:
{import.meta.env.DEV && (
  <button
    data-part="statusbar-typography-btn"
    onClick={() => dispatch(togglePalette())}
    title="Typography Debug Palette (Ctrl+Shift+T)"
    style={{
      background: 'none',
      border: '1px solid #999',
      padding: '0 4px',
      cursor: 'pointer',
      fontSize: 10,
      fontFamily: 'serif',
    }}
  >
    𝒜a
  </button>
)}
```

---

## 10. Step-by-Step Implementation Plan

### Phase 1: Foundation (estimated: 2–3 hours)

| Step | What | Files |
|------|------|-------|
| 1.1 | Create `types/typography-palette.ts` with all type definitions | `web/src/types/typography-palette.ts` |
| 1.2 | Create `store/typographyPaletteSlice.ts` with the Redux slice | `web/src/store/typographyPaletteSlice.ts` |
| 1.3 | Wire the slice into `store.ts` | `web/src/store.ts` |
| 1.4 | Create `css-override-engine.ts` with `applyOverrides()` and `clearOverrides()` | `web/src/components/TypographyPalette/css-override-engine.ts` |
| 1.5 | Create `useTypographyOverrides.ts` hook | `web/src/components/TypographyPalette/useTypographyOverrides.ts` |
| 1.6 | Create `usePaletteShortcut.ts` hook | `web/src/components/TypographyPalette/usePaletteShortcut.ts` |
| 1.7 | Test: verify that manually dispatching `setOverride` from browser console applies CSS | Console test |

### Phase 2: Element Registry and Presets (estimated: 1–2 hours)

| Step | What | Files |
|------|------|-------|
| 2.1 | Create `element-registry.ts` with all groups and elements | `web/src/components/TypographyPalette/element-registry.ts` |
| 2.2 | Create `presets.ts` with all preset definitions | `web/src/components/TypographyPalette/presets.ts` |
| 2.3 | Test: verify that applying a preset via console dispatch changes typography | Console test |

### Phase 3: UI Components (estimated: 3–4 hours)

| Step | What | Files |
|------|------|-------|
| 3.1 | Create `parts.ts` for the palette | `web/src/components/TypographyPalette/parts.ts` |
| 3.2 | Create `typography-palette.css` | `web/src/components/TypographyPalette/styles/typography-palette.css` |
| 3.3 | Create `FontFamilySelect.tsx` | `web/src/components/TypographyPalette/FontFamilySelect.tsx` |
| 3.4 | Create `FontSizeStepper.tsx` | `web/src/components/TypographyPalette/FontSizeStepper.tsx` |
| 3.5 | Create `FontWeightSelect.tsx` | `web/src/components/TypographyPalette/FontWeightSelect.tsx` |
| 3.6 | Create `ColorStepper.tsx` | `web/src/components/TypographyPalette/ColorStepper.tsx` |
| 3.7 | Create `PresetSelector.tsx` | `web/src/components/TypographyPalette/PresetSelector.tsx` |
| 3.8 | Create `TypographyPaletteElement.tsx` | `web/src/components/TypographyPalette/TypographyPaletteElement.tsx` |
| 3.9 | Create `TypographyPaletteGroup.tsx` (accordion) | `web/src/components/TypographyPalette/TypographyPaletteGroup.tsx` |
| 3.10 | Create `TypographyPalette.tsx` (main panel) | `web/src/components/TypographyPalette/TypographyPalette.tsx` |
| 3.11 | Add Storybook story | `web/src/components/TypographyPalette/TypographyPalette.stories.tsx` |

### Phase 4: Integration (estimated: 1–2 hours)

| Step | What | Files |
|------|------|-------|
| 4.1 | Add `<TypographyPalette />` to `App.tsx` | `web/src/App.tsx` |
| 4.2 | Add `usePaletteShortcut()` call in `App.tsx` | `web/src/App.tsx` |
| 4.3 | Add dev-mode `𝒜a` button to `StatusBar.tsx` | `web/src/components/StatusBar/StatusBar.tsx` |
| 4.4 | Manual end-to-end test with the running dev server | `pnpm dev` |

### Phase 5: Polish (estimated: 1–2 hours)

| Step | What | Files |
|------|------|-------|
| 5.1 | Verify all presets look good with real content | Manual testing |
| 5.2 | Verify em-unit headings compute correctly when base size changes | Manual testing |
| 5.3 | Verify palette doesn't interfere with Storybook | Storybook |
| 5.4 | Verify palette is hidden in production builds | Build test |
| 5.5 | Add "Copy CSS" button to export overrides as CSS | Optional |

---

## 11. File Reference Summary

### Files That Already Exist (read-only for this feature)

| Path | Role |
|------|------|
| `web/src/styles/global.css` | Root `:root` CSS variables — the source of truth for default typography values |
| `web/src/components/Markdown/styles/markdown.css` | Prose typography (h1–h3, p, code, pre, blockquote, tables, links) |
| `web/src/components/SectionView/styles/section-view.css` | Section header and body typography |
| `web/src/components/TitleBar/styles/titlebar.css` | Title bar typography |
| `web/src/components/MenuBar/styles/menubar.css` | Menu bar typography |
| `web/src/components/SearchBar/styles/searchbar.css` | Search input typography |
| `web/src/components/PackageSelector/styles/package-selector.css` | Package selector typography |
| `web/src/components/NavigationModeToggle/styles/navigation-mode-toggle.css` | Nav toggle typography |
| `web/src/components/TypeFilter/styles/typefilter.css` | Type filter button typography |
| `web/src/components/DocumentationTree/styles/documentation-tree.css` | Tree row and heading typography |
| `web/src/components/SectionList/styles/section-list.css` | Section card typography |
| `web/src/components/Badge/styles/badge.css` | Badge typography |
| `web/src/components/StatusBar/styles/statusbar.css` | Status bar typography |
| `web/src/components/EmptyState/styles/empty-state.css` | Empty state typography |
| `web/src/components/PackageIndex/styles/package-index.css` | Package index typography |
| `web/src/App.tsx` | Root app component — where palette is rendered |
| `web/src/store.ts` | Redux store — where palette slice is added |
| `web/src/services/api.ts` | RTK Query API — existing state management pattern |
| `web/src/types/index.ts` | TypeScript interfaces — existing pattern for types |
| `web/public/fonts/ChicagoFLF.woff2` | The custom Chicago_ font file |
| `web/public/site-config.js` | Runtime config |
| `web/vite.config.ts` | Vite build config |
| `web/package.json` | Dependencies and scripts |
| `pkg/help/server/serve.go` | Go server (not modified) |
| `pkg/help/server/handlers.go` | API handlers (not modified) |
| `pkg/help/server/types.go` | Go response types (not modified) |

### New Files to Create

| Path | Purpose |
|------|---------|
| `web/src/types/typography-palette.ts` | TypeScript type definitions |
| `web/src/store/typographyPaletteSlice.ts` | Redux slice for palette state |
| `web/src/components/TypographyPalette/TypographyPalette.tsx` | Main panel component |
| `web/src/components/TypographyPalette/TypographyPalette.stories.tsx` | Storybook story |
| `web/src/components/TypographyPalette/parts.ts` | data-part constants |
| `web/src/components/TypographyPalette/element-registry.ts` | Element group and selector definitions |
| `web/src/components/TypographyPalette/presets.ts` | Preset definitions |
| `web/src/components/TypographyPalette/css-override-engine.ts` | CSS generation and injection utility |
| `web/src/components/TypographyPalette/useTypographyOverrides.ts` | Redux→DOM sync hook |
| `web/src/components/TypographyPalette/usePaletteShortcut.ts` | Keyboard shortcut hook |
| `web/src/components/TypographyPalette/TypographyPaletteGroup.tsx` | Accordion group component |
| `web/src/components/TypographyPalette/TypographyPaletteElement.tsx` | Per-element control row |
| `web/src/components/TypographyPalette/FontFamilySelect.tsx` | Font family dropdown |
| `web/src/components/TypographyPalette/FontSizeStepper.tsx` | Size stepper |
| `web/src/components/TypographyPalette/FontWeightSelect.tsx` | Weight dropdown |
| `web/src/components/TypographyPalette/ColorStepper.tsx` | Gray shade stepper |
| `web/src/components/TypographyPalette/PresetSelector.tsx` | Preset dropdown |
| `web/src/components/TypographyPalette/styles/typography-palette.css` | Palette styles |

### Files to Modify

| Path | Change |
|------|--------|
| `web/src/store.ts` | Add `typographyPalette` reducer |
| `web/src/App.tsx` | Render `<TypographyPalette />` + call `usePaletteShortcut()` |
| `web/src/components/StatusBar/StatusBar.tsx` | Add dev-mode `𝒜a` toggle button |

---

## 12. API References

### CSS Custom Properties (current)

| Variable | Current Value | Used By |
|----------|--------------|---------|
| `--font-ui` | `'Chicago_', 'Geneva', ...` | Most UI components |
| `--font-mono` | `'Monaco', 'Courier New', ...` | Code, slug labels |
| `--font-size-sm` | `11px` | Not widely used |
| `--font-size-base` | `13px` | `html, body` |
| `--font-size-lg` | `15px` | Not widely used |
| `--font-size-xl` | `18px` | Not widely used |
| `--color-fg` | `#000000` | `html, body`, search input |
| `--color-accent` | `#000000` | Links in markdown |
| `--color-selection-bg` | `#000000` | Selection highlight |
| `--color-selection-fg` | `#ffffff` | Selection text |

### Redux Actions (new)

| Action | Payload | Effect |
|--------|---------|--------|
| `typographyPalette/togglePalette` | none | Toggle `isOpen` |
| `typographyPalette/openPalette` | none | Set `isOpen = true` |
| `typographyPalette/closePalette` | none | Set `isOpen = false` |
| `typographyPalette/setActiveGroup` | `string \| null` | Expand/collapse accordion group |
| `typographyPalette/setPreset` | `{ presetId, overrides }` | Load preset overrides |
| `typographyPalette/setOverride` | `{ elementId, properties }` | Set one element's properties |
| `typographyPalette/removeOverride` | `string` (elementId) | Remove one element's overrides |
| `typographyPalette/resetAllOverrides` | none | Clear all overrides |

### Redux State Selectors (new)

| Selector | Returns | Description |
|----------|---------|-------------|
| `state.typographyPalette.isOpen` | `boolean` | Palette visibility |
| `state.typographyPalette.activeGroup` | `string \| null` | Expanded accordion group |
| `state.typographyPalette.activePreset` | `string \| null` | Current preset ID |
| `state.typographyPalette.overrides` | `TypographyOverrides` | Current property overrides |

### DOM Side Effects (new)

| ID | Element | Purpose |
|----|---------|---------|
| `typography-palette-overrides` | `<style>` in `<head>` | Injected CSS rules for overrides |

---

## 13. Edge Cases and Design Decisions

### 13.1 em vs px for Markdown Headings

Markdown headings (`h1`, `h2`, `h3`) use relative `em` units based on the markdown content root's `font-size`. When the palette changes the prose body size from 13px to 16px, an `h1` at `1.6em` goes from 20.8px to 25.6px. The palette's stepper for heading sizes shows both the em value and the approximate computed px value:

```
Size: [ − ] 1.6em ≈ 20.8px [ + ]
```

The step size for em values is 0.1em.

### 13.2 Font Family Override Scope

When the user changes the font family for "Markdown Prose → Body Text" from `ui` to `mono`, the entire markdown content switches to Monaco. However, `code` and `pre` elements within markdown already have their own font-family set to `var(--font-mono)`. The CSS specificity of `[data-part='markdown-content'] code` is higher than the generic `[data-part='markdown-content']` override, so code stays in Monaco even when prose switches.

But if the user explicitly changes "Markdown Code → Inline Code" font to `ui`, that override is more specific (targets the code element directly) and will win.

**Rule:** More-specific selectors in the element registry always override less-specific ones, matching normal CSS cascade behavior.

### 13.3 Dev-Only Guard

The palette should be entirely hidden in production. Three guard points:

1. `usePaletteShortcut` returns early if `!import.meta.env.DEV`
2. The `𝒜a` button in StatusBar only renders when `import.meta.env.DEV`
3. The `TypographyPalette` component still renders if `isOpen` is true (in case someone opens it in dev and the state persists across HMR), but the toggle button and shortcut are dev-only

Vite's `import.meta.env.DEV` is `true` in dev mode and `false` in production builds. Dead code elimination removes the dev-only code from the production bundle.

### 13.4 Palette Z-Index

The palette uses `z-index: 9999` to float above all content. This is safe because no other component in the docs browser uses z-index (the app does not have modals or dropdowns that need high z-index).

### 13.5 Performance

The `applyOverrides` function regenerates all CSS rules on every change. This is fine because:

- The number of rules is small (≤ 30 elements × 5 properties = 150 declarations max)
- React batches state updates, so multiple stepper clicks result in one DOM write
- Setting `textContent` on a `<style>` element triggers a single style recalculation

### 13.6 No Persistence

Overrides are intentionally ephemeral. Refreshing the page resets everything. This is a design choice, not a limitation. The goal is quick experimentation, not saved themes. If persistence is desired in the future, `localStorage` can be added in a follow-up.

---

## 14. Testing Strategy

### 14.1 Unit Tests

| Test | File | What It Verifies |
|------|------|------------------|
| CSS override generation | `css-override-engine.test.ts` | Given overrides, generates correct CSS rules |
| Empty overrides | `css-override-engine.test.ts` | No rules when overrides is `{}` |
| Redux slice | `typographyPaletteSlice.test.ts` | Actions update state correctly |
| Preset application | `presets.test.ts` | Each preset has valid overrides |

### 14.2 Integration Tests (Manual)

| Test | Steps | Expected |
|------|-------|----------|
| Toggle palette | Press `Ctrl+Shift+T` | Palette appears/disappears |
| Status bar button | Click `𝒜a` | Palette opens |
| Change font size | Expand "Markdown Prose", click + on size | Body text gets bigger |
| Change font weight | Expand "Title Bar", change weight to 400 | Title bar text becomes light |
| Apply preset | Select "Dense Terminal" | Everything switches to monospace 12px |
| Reset all | Click "Reset All" | All changes revert to defaults |
| Code block isolation | Change prose font to mono, check code blocks | Code blocks stay in Monaco |
| Dev guard | Build production, open page | No `𝒜a` button, no palette |

### 14.3 Storybook

The palette gets its own Storybook story showing it in isolation (with a mock content area behind it).

---

## 15. Future Enhancements (Out of Scope for V1)

These are explicitly **not** part of the initial implementation but are documented here so they can be picked up later:

- **Export CSS:** A "Copy CSS" button that copies all overrides as a `:root { ... }` block you can paste into `global.css`
- **Persistence:** Save overrides to `localStorage` and reload on page refresh
- **Custom font loading:** Upload a `.woff2` file and preview it before adding to the project
- **Spacing controls:** Line height, margins, padding for prose elements
- **Color accent:** Extend beyond monochrome to support a single accent color (for links, badges)
- **Multiple font stacks:** Add more font options beyond Chicago_ and Monaco
- **Palette docking:** Let the palette dock to the left or right instead of floating

---

## 16. Glossary

| Term | Definition |
|------|-----------|
| **data-part** | A custom HTML attribute used instead of CSS class names for styling. Example: `data-part='titlebar-title'` |
| **Chicago_** | The custom bitmap-style font used for the Classic Mac aesthetic. File: `web/public/fonts/ChicagoFLF.woff2` |
| **Monaco** | The monospace font in the font stack. Native macOS font, no web font needed. |
| **em unit** | Relative CSS unit. `1.6em` means 1.6× the parent's font-size. |
| **RTK Query** | Redux Toolkit Query — data fetching and caching layer used by the app |
| **Vite** | Frontend build tool. Handles dev server, HMR, and production builds. |
| **go:embed** | Go compiler directive to embed files into the binary at build time |
| **SSR** | Server-Side Rendering — the Go server can proxy to a Node.js sidecar for SSR |
| **Storybook** | Component development environment. Each component has `.stories.tsx` files. |
| **Classic Mac aesthetic** | The visual design language of this app: black borders, 2px solid lines, Chicago_ font, checkerboard background, retro scrollbar |
