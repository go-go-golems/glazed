---
title: "CSS Architecture Analysis: Fonts, Parts, and the Typography Palette Override System"
doc-type: design
status: active
intent: long-term
topics:
  - css
  - fonts
  - typography
  - architecture
owners:
  - manuel
ticket: garamond-body-restyle
created: "2026-05-28"
---

# CSS Architecture Analysis: Fonts, Parts, and the Typography Palette Override System

## Purpose

This document is a complete technical guide for a new intern joining the Glazed Help Browser project. It explains **every layer** of the CSS and font system so you can understand why fonts render the way they do, where to change them, and what traps to avoid.

The motivating problem: we wanted to switch markdown body text to **EB Garamond** (serif) while keeping all UI chrome (headers, sidebar, buttons, status bar) in **Chicago** (bitmap). What seemed like a simple CSS change turned into a multi-hour investigation because of how the system is layered.

---

## 1. What is the Glazed Help Browser?

The Glazed Help Browser is a single-page application (SPA) written in React + TypeScript, built with Vite. It renders documentation stored in a Go backend as a browsable, classic Mac / System 7 styled help viewer.

```
┌─────────────────────────────────────────────────────┐
│  Browser                                            │
│  ┌──────────────┐  ┌─────────────────────────────┐  │
│  │  Sidebar      │  │  Content Pane               │  │
│  │  (Tree/List)  │  │  (Markdown or Index)        │  │
│  │               │  │                              │  │
│  │  Chicago ◄────┤  │  Headers: Chicago            │  │
│  │               │  │  Paragraphs: EB Garamond     │  │
│  │               │  │  Code: Monaco                │  │
│  └──────────────┘  └─────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐│
│  │  Status Bar (Chicago)                            ││
│  └──────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────┘
```

**Key directories:**

| Directory | What lives here |
|---|---|
| `web/src/components/` | React components, each in its own folder |
| `web/src/components/<Name>/styles/` | Component-specific CSS files |
| `web/src/components/<Name>/parts.ts` | `data-part` attribute constants |
| `web/src/styles/global.css` | Root CSS variables, `@font-face`, resets |
| `web/public/fonts/` | Font binary files (`.woff`, `.woff2`) |

**Tech stack:**

- **Vite** — dev server + bundler
- **React 18** with `hydrateRoot` (SSR-compatible)
- **Redux Toolkit** — state management
- **React Router v6** — client-side routing with `BrowserRouter`
- **react-markdown + remark-gfm** — markdown rendering
- **No CSS-in-JS** — plain CSS files imported via Vite


## 2. The `data-part` Convention

### 2.1 Why not className?

This project uses a **`data-part` attribute** convention instead of CSS classes for styling. Every component exports a `parts.ts` file with stable string constants, and the CSS targets those via attribute selectors.

**Example — `Badge` component:**

```typescript
// src/components/Badge/parts.ts
export const BadgeParts = {
  root: 'badge',
} as const;
```

```tsx
// src/components/Badge/Badge.tsx
<span data-part={BadgeParts.root}>
  {label}
</span>
```

```css
/* src/components/Badge/styles/badge.css */
[data-part='badge'] {
  font-family: var(--font-ui);
  font-size: 10px;
  /* ... */
}
```

### 2.2 The complete parts registry

Every component in the system follows this pattern. Here is the full list of `data-part` values:

| Component | Parts | File |
|---|---|---|
| AppLayout | `app-layout`, `app-layout-sidebar`, `app-layout-content` | `AppLayout/parts.ts` |
| TitleBar | `titlebar`, `titlebar-icon`, `titlebar-ruler`, `titlebar-title` | `TitleBar/parts.ts` |
| MenuBar | `menubar`, `menubar-item`, `menubar-label` | `MenuBar/parts.ts` |
| SearchBar | `searchbar`, `searchbar-input` | `SearchBar/parts.ts` |
| PackageSelector | `package-selector-root`, `package-selector-row`, `package-selector-label`, `package-selector-select` | `PackageSelector/parts.ts` |
| NavigationModeToggle | `navigation-mode-toggle-root`, `navigation-mode-toggle-label`, `navigation-mode-toggle-buttons`, `navigation-mode-toggle-button` | `NavigationModeToggle/parts.ts` |
| TypeFilter | `typefilter`, `typefilter-button` | `TypeFilter/parts.ts` |
| DocumentationTree | `documentation-tree-root`, `documentation-tree-group`, `documentation-tree-row`, `documentation-tree-disclosure`, `documentation-tree-icon`, `documentation-tree-label`, `documentation-tree-children`, `documentation-tree-heading`, `documentation-tree-empty` | `DocumentationTree/parts.ts` |
| SectionList | `section-list`, `section-list-item` | `SectionList/parts.ts` |
| SectionCard | `section-card`, `section-card-meta`, `section-card-title`, `section-card-short`, `section-card-top-badge` | `SectionList/parts.ts` |
| PackageIndex | `package-index`, `package-index-heading`, `package-index-count`, `package-index-group`, `package-index-group-heading`, `package-index-section-item`, `package-index-section-link`, `package-index-section-short` | `PackageIndex/parts.ts` |
| SectionView | `section-view`, `section-view-body` | `SectionView/parts.ts` |
| SectionHeader | `section-header`, `section-header-slug`, `section-header-heading`, `section-header-subtitle`, `section-header-tags` | `SectionView/parts.ts` |
| Markdown | `markdown-content` | `Markdown/parts.ts` |
| StatusBar | `statusbar`, `statusbar-count`, `statusbar-version` | `StatusBar/parts.ts` |
| Badge | `badge` | `Badge/parts.ts` |
| EmptyState | `empty-state`, `empty-state-icon`, `empty-state-label` | `EmptyState/parts.ts` |

### 2.3 Space-separated parts (multi-part elements)

Some elements carry **multiple** `data-part` values separated by spaces. For example, `SectionCard` renders:

```tsx
<div data-part={`${SectionListParts.item} ${SectionCardParts.root}`}>
```

Which produces: `data-part="section-list-item section-card"`

CSS must use the **`~=` (word-match) selector** instead of `=`:

```css
/* WRONG — only matches exact "section-card" */
[data-part='section-card-title'] { ... }

/* CORRECT — matches any whitespace-separated token */
[data-part~='section-card-title'] { ... }
```

This is why `section-list.css` uses `[data-part~='section-card-short']` while other files use `[data-part='markdown-content']`.

### 2.4 How to know which selector to use

```
If the element has a SINGLE data-part value → use [data-part='exact-value']
If the element has MULTIPLE data-part values → use [data-part~='word-value']
```

Check the component's JSX: if it uses string concatenation or template literals for `data-part`, use `~=`.


## 3. The Font Stack

### 3.1 Three font families

The system defines three font variables in `global.css`:

```css
:root {
  --font-ui: 'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif;
  --font-mono: 'Monaco', 'Courier New', monospace;
  --font-serif: 'EB Garamond', 'Garamond', 'Palatino', 'Georgia', serif;
}
```

| Variable | Purpose | Source font |
|---|---|---|
| `--font-ui` | All UI chrome: headers, buttons, sidebar, status bar | Chicago (bitmap woff2) |
| `--font-mono` | Code blocks, inline code, slug labels | Monaco (system) |
| `--font-serif` | Paragraph body text only | EB Garamond 12 (woff) |

### 3.2 Font file loading

Fonts are loaded via `@font-face` rules in `global.css` and stored in `web/public/fonts/`:

```
web/public/fonts/
├── ChicagoFLF.woff2        ← --font-ui primary
├── EBGaramond12-Regular.woff  ← --font-serif regular
├── EBGaramond12-Italic.woff   ← --font-serif italic
├── EBGaramond12-Bold.woff     ← --font-serif bold
└── NOTICE.md
```

The `@font-face` declarations:

```css
@font-face {
  font-family: 'Chicago_';
  src: url('/fonts/ChicagoFLF.woff2') format('woff2');
  font-display: swap;
}

@font-face {
  font-family: 'EB Garamond';
  src: url('/fonts/EBGaramond12-Regular.woff') format('woff');
  font-weight: 400;
  font-display: swap;
}
/* ... italic and bold variants */
```

**Important:** these URLs are root-relative (`/fonts/...`). In dev, Vite serves them from `public/fonts/`. In production, they're embedded in the Go binary via `go:embed`.

### 3.3 Which font should go where?

```
┌─────────────────────────────────────────────────────────────┐
│ Element type              │ Font variable   │ Rationale     │
├─────────────────────────────────────────────────────────────┤
│ Title bar, menu bar       │ --font-ui       │ UI chrome     │
│ Sidebar tree labels       │ --font-ui       │ UI chrome     │
│ Section card titles       │ --font-ui       │ UI chrome     │
│ Status bar                │ --font-ui       │ UI chrome     │
│ Buttons, selectors        │ --font-ui       │ UI chrome     │
│ Search input              │ --font-ui       │ UI chrome     │
│ Package index heading     │ --font-ui       │ Header        │
│ Package index group head  │ --font-ui       │ Header        │
│ Package section links     │ --font-ui       │ Title/anchor  │
│ Section header h1         │ --font-ui       │ Header        │
│ Markdown h1, h2, h3       │ --font-ui       │ Headers       │
├─────────────────────────────────────────────────────────────┤
│ Markdown paragraphs <p>   │ --font-serif    │ Body text     │
│ Markdown blockquotes      │ --font-serif    │ Body text     │
│ Package section short     │ --font-serif    │ Description   │
│ Section header subtitle   │ --font-serif    │ Description   │
│ Section card short        │ --font-serif    │ Description   │
├─────────────────────────────────────────────────────────────┤
│ Code blocks <pre>         │ --font-mono     │ Code          │
│ Inline code <code>        │ --font-mono     │ Code          │
│ Section header slug       │ --font-mono     │ Code-like     │
└─────────────────────────────────────────────────────────────┘
```


## 4. CSS Cascade and Inheritance

### 4.1 The cascade chain for font-family

When the browser resolves `font-family` for any element, it walks up the DOM tree until it finds an element with an explicit `font-family`. Understanding this chain is critical.

**DOM hierarchy:**

```
<html>                    ← font-family: Chicago_ (from body)
  <body>                  ← font-family: Chicago_ (global.css)
    <div id="root">       ← font-family: Chicago_ (inherited)
      <div class="app-root">  ← font-family: var(--font-ui) (global.css)
        <div data-part="app-layout">
          <div data-part="app-layout-sidebar">     ← inherits from .app-root
            <div data-part="documentation-tree-root">  ← font-family: var(--font-ui) (CSS)
              ...
            </div>
          </div>
          <div data-part="app-layout-content">     ← inherits from .app-root
            <div data-part="section-view">          ← NO explicit font-family
              <div data-part="section-view-body">   ← NO explicit font-family
                <div data-part="markdown-content">  ← NO explicit font-family on root
                  <p>                               ← font-family: var(--font-serif) (CSS)
                    Text here in Garamond
                  </p>
                  <h2>                             ← font-family: var(--font-ui) (CSS)
                    Header in Chicago
                  </h2>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </body>
</html>
```

### 4.2 The golden rule

**Font-family only needs to be set where it differs from the parent.** The `.app-root` establishes Chicago as the default. Only elements that need Garamond must explicitly override it.

This means:
- **Do NOT** set `font-family: var(--font-ui)` on every single component. Let inheritance do the work.
- **DO** set `font-family: var(--font-serif)` only on the specific elements that need body text.
- **DO** explicitly set `font-family: var(--font-ui)` only when a parent might have overridden it with serif.

### 4.3 The override trap: specificity and source order

CSS rules are resolved by: **(1) specificity**, then **(2) source order** (last one wins). A `<style>` tag injected later in the `<head>` will override an earlier `<style>` tag with the same specificity.

This is exactly what bit us with the browser extension.


## 5. The Browser Extension Problem

### 5.1 Discovery

During this investigation, we found that the browser had a **browser extension** injecting a `<style id="typography-palette-overrides">` tag into the page `<head>`. This tag contained 26 CSS rules that overrode our application styles.

**Injected rules (abridged):**

```css
/* Rule 1 — forces EVERYTHING to Garamond */
.app-root { font-family: "EB Garamond", Garamond, Georgia, serif; }

/* Rules 2-8 — patch specific UI elements back to Chicago */
[data-part="titlebar-title"] { font-family: Chicago_, ... sans-serif; }
[data-part="menubar"] { font-family: Chicago_, ... sans-serif; }
[data-part="searchbar-input"] { font-family: Chicago_, ... sans-serif; }
/* ... etc */

/* Rules 9-10 — forces tree rows to Garamond (!) */
[data-part="documentation-tree-row"] { font-family: "EB Garamond", ... serif; }

/* Rules 11-12 — forces section cards to Garamond */
[data-part~="section-card-title"] { font-family: "EB Garamond", ... serif; }
[data-part~="section-card-short"] { font-family: "EB Garamond", ... serif; }

/* Rules 13-14 — section header: slug=mono, heading=chicago */
[data-part="section-header-slug"] { font-family: Monaco, ... monospace; }
[data-part="section-header-heading"] { font-family: Chicago_, ... sans-serif; }

/* Rules 15-25 — markdown elements */
[data-part="markdown-content"] { font-family: "EB Garamond", ... serif; font-size: 15px; ... }
[data-part="markdown-content"] h1 { font-family: Chicago_, ... sans-serif; }
[data-part="markdown-content"] h2 { font-family: Chicago_, ... sans-serif; }
/* ... code, pre, blockquote, etc */
```

### 5.2 Why this breaks our CSS

The extension's `<style>` tag appears **after** all Vite-injected style tags in the `<head>`. Since `.app-root { font-family: serif }` has the same specificity as our `.app-root { font-family: var(--font-ui) }`, the extension wins by source order.

**Cascade order in the browser:**

```
Sheet index 15 (Vite/global.css):    .app-root { font-family: var(--font-ui); }     ← loses
Sheet index 16 (extension):           .app-root { font-family: "EB Garamond", serif; } ← wins
```

### 5.3 How to detect this

Use Playwright to dump all stylesheets targeting an element:

```javascript
// Pseudocode for a Playwright evaluate script
const el = document.querySelector('.app-root');
const sheets = document.styleSheets;
for (const sheet of sheets) {
  for (const rule of sheet.cssRules) {
    if (rule.selectorText && el.matches(rule.selectorText)) {
      console.log({
        selector: rule.selectorText,
        fontFamily: rule.style.fontFamily,
        id: sheet.ownerNode?.id,  // "typography-palette-overrides" = extension
        href: sheet.href,          // null = inline <style>
      });
    }
  }
}
```

### 5.4 How to fix it

**Option A (recommended): Remove or disable the browser extension.** The app's own CSS should be the single source of truth for fonts.

**Option B: Use `!important` in app CSS.** This is a last resort — it makes the cascade harder to reason about.

**Option C: Use higher specificity.** E.g., `html .app-root { font-family: var(--font-ui); }` would beat `.app-root` even with source order disadvantage. But this pollutes selectors.

### 5.5 Lesson learned

When debugging CSS that looks correct in the source files but doesn't render correctly, **always check for injected stylesheets**. Browser extensions (Stylus, Dark Reader, custom typography extensions) can silently override your styles.


## 6. Component-by-Component CSS Map

This section is a reference for every CSS file in the system, what it styles, and what font it uses.

### 6.1 Global styles (`src/styles/global.css`)

```css
/* Sets the root font for the entire app */
.app-root {
  font-family: var(--font-ui);   /* Chicago */
  font-size: 13px;
}

/* Body inherits Chicago */
body {
  font-family: var(--font-ui);
}
```

**Variables defined here:**

| Variable | Value | Used by |
|---|---|---|
| `--font-ui` | `'Chicago_', 'Geneva', ...` | All UI chrome |
| `--font-mono` | `'Monaco', 'Courier New', monospace` | Code, slugs |
| `--font-serif` | `'EB Garamond', 'Garamond', ...` | Body text |
| `--font-size-sm` | `11px` | Small labels |
| `--font-size-base` | `13px` | Default |
| `--font-size-lg` | `15px` | Serif body |
| `--font-size-xl` | `18px` | Large headings |

### 6.2 AppLayout (`components/AppLayout/styles/app-layout.css`)

No font-family declarations. Inherits from `.app-root` (Chicago). This is correct — the layout shell should not override fonts.

### 6.3 TitleBar (`components/TitleBar/styles/titlebar.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.4 MenuBar (`components/MenuBar/styles/menubar.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.5 SearchBar (`components/SearchBar/styles/searchbar.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.6 PackageSelector (`components/PackageSelector/styles/package-selector.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.7 NavigationModeToggle (`components/NavigationModeToggle/styles/navigation-mode-toggle.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.8 TypeFilter (`components/TypeFilter/styles/typefilter.css`)

No font-family declarations. Inherits Chicago. Correct.

### 6.9 DocumentationTree (`components/DocumentationTree/styles/documentation-tree.css`)

```css
[data-part='documentation-tree-root'] {
  font-family: var(--font-ui);  /* Explicit Chicago — good for safety */
}
```

**Status:** Has explicit `font-family: var(--font-ui)`. This was added during this investigation as a safety net. Before, it relied on inheritance from `.app-root`, which the browser extension was overriding.

### 6.10 SectionList + SectionCard (`components/SectionList/styles/section-list.css`)

```css
/* Section card short description — SERIF */
[data-part~='section-card-short'] {
  font-family: var(--font-serif);  /* Garamond for sidebar descriptions */
  font-size: 11px;
}

/* Section card title — inherits Chicago from parent */
[data-part~='section-card-title'] {
  /* No font-family — inherits Chicago */
}
```

### 6.11 PackageIndex (`components/PackageIndex/styles/package-index.css`)

```css
/* Section short description in index — SERIF */
[data-part='package-index-section-short'] {
  font-family: var(--font-serif);  /* Garamond for descriptions */
  font-size: 14px;
}

/* Everything else inherits Chicago */
```

### 6.12 SectionView + SectionHeader (`components/SectionView/styles/section-view.css`)

```css
/* Section header — explicit Chicago */
[data-part='section-header'] {
  font-family: var(--font-ui);  /* All chrome = Chicago */
}

/* Subtitle/description — SERIF */
[data-part='section-header-subtitle'] {
  font-family: var(--font-serif);  /* Garamond for description */
  font-size: 14px;
}

/* Section view root — NO font override (was removed) */
[data-part='section-view'] {
  /* Previously had font-family: var(--font-ui) here.
     Removed because it overrode markdown paragraph font.
     Now inherits from .app-root = Chicago. */
}

/* Section view body — wraps markdown */
[data-part='section-view-body'] {
  /* No font-family. Inherits Chicago from section-view.
     Markdown paragraphs override to serif via their own CSS. */
}
```

### 6.13 MarkdownContent (`components/Markdown/styles/markdown.css`)

```css
/* Root — no font override (inherits Chicago) */
[data-part='markdown-content'] {
  font-size: 13px;
  line-height: 1.6;
}

/* Paragraphs — SERIF */
[data-part='markdown-content'] p {
  font-family: var(--font-serif);  /* Garamond */
  font-size: 15px;
  line-height: 1.7;
}

/* Headers — explicit CHICAGO */
[data-part='markdown-content'] h1,
[data-part='markdown-content'] h2,
[data-part='markdown-content'] h3 {
  font-family: var(--font-ui);  /* Chicago */
  font-weight: 700;
}

/* Code — MONO */
[data-part='markdown-content'] code,
[data-part='markdown-content'] pre {
  font-family: var(--font-mono);  /* Monaco */
}

/* Blockquotes — SERIF */
[data-part='markdown-content'] blockquote {
  font-family: var(--font-serif);  /* Garamond */
}
```

### 6.14 StatusBar (`components/StatusBar/styles/statusbar.css`)

No font-family declarations. Inherits Chicago from `.app-root`. Correct.

### 6.15 Badge (`components/Badge/styles/badge.css`)

No font-family declarations. Inherits Chicago. Correct.


## 7. Architecture Review: What's Good, What's Problematic

### 7.1 What works well

- **`data-part` convention is solid.** Attribute selectors `[data-part='foo']` are:
  - Globally unique (no naming collisions between components)
  - Searchable (`grep -r "data-part='badge'" src/`)
  - Typed via `parts.ts` constants (IDE autocomplete, rename safety)
  - Decoupled from class names (which might change for behavioral reasons)

- **CSS files co-located with components.** Each component's styles live in `ComponentName/styles/`. Easy to find.

- **CSS variables for fonts.** `--font-ui`, `--font-mono`, `--font-serif` make it clear what each font is for. Changing a font means changing one variable, not 50 files.

- **No CSS-in-JS runtime.** Plain CSS files imported via Vite. Zero runtime overhead. No styled-components, no emotion, no CSS modules hash computation.

### 7.2 What's problematic

#### Problem 1: Inconsistent use of `font-family` declarations

Some components set `font-family: var(--font-ui)` explicitly, others rely on inheritance. This creates confusion:

```
documentation-tree-root  → explicit var(--font-ui)   ✓ defensive
app-layout               → no font declaration       ✓ inherits from .app-root
section-view             → no font declaration        ✓ (removed during investigation)
section-header           → explicit var(--font-ui)    ✓ defensive
package-index            → no font declaration        ✓ inherits
markdown-content root    → no font declaration        ✓ (paragraphs override individually)
```

**Recommendation:** Adopt one of two strategies:

**Strategy A — Explicit everywhere (verbose but safe):**
```css
/* Set font on every component root */
[data-part='documentation-tree-root'] { font-family: var(--font-ui); }
[data-part='package-index'] { font-family: var(--font-ui); }
[data-part='section-view'] { font-family: var(--font-ui); }
```

**Strategy B — Inherit by default (current, minimal):**
Only set `font-family` where it differs from the parent. Trust the cascade. This is what we mostly do now.

Strategy B is cleaner but fragile — any injected stylesheet breaks it. Strategy A is more defensive.

#### Problem 2: The `.app-root` selector is a class, not a data-part

The root app container uses `className="app-root"` instead of `data-part="app-root"`. This breaks the convention:

```css
/* Inconsistent: class selector */
.app-root { font-family: var(--font-ui); }

/* Consistent with the rest: attribute selector */
[data-part='app-root'] { font-family: var(--font-ui); }
```

**Recommendation:** Migrate `.app-root` to `data-part="app-root"` for consistency.

#### Problem 3: No defensive `@layer` usage

CSS `@layer` allows declaring priority tiers:

```css
@layer reset, base, components, overrides;

@layer components {
  [data-part='markdown-content'] p {
    font-family: var(--font-serif);
  }
}
```

Without layers, any injected stylesheet with equal specificity wins by source order. With layers, the application's `components` layer could be declared to win over unlayered styles.

**Recommendation:** Consider adopting `@layer` for defensive styling.

#### Problem 4: `section-view` previously forced `font-family: var(--font-ui)`

Before this investigation, `section-view.css` had:

```css
[data-part='section-view'] {
  font-family: var(--font-ui);  /* ← This forced Chicago on ALL children */
}
```

This made it impossible for `markdown-content p` to inherit Garamond from its own CSS, because `section-view` was forcing Chicago higher up the tree. Even though `[data-part='markdown-content'] p` has higher specificity than `[data-part='section-view']`, both set `font-family` explicitly — so the paragraph's own rule does win. But the confusion arose because of the browser extension simultaneously overriding `.app-root`.

**This was removed.** The section-view now inherits from `.app-root`.

#### Problem 5: Duplicate CSS rules for `p` in markdown.css

During the investigation, we accidentally created two rules for `[data-part='markdown-content'] p`:

```css
/* First declaration */
[data-part='markdown-content'] p {
  font-family: var(--font-serif);
  font-size: 15px;
  line-height: 1.7;
}

/* Second declaration (from original file) — overrides the first! */
[data-part='markdown-content'] p {
  margin: 0.8em 0;   /* But font-family, font-size, line-height are LOST */
}
```

**This was fixed** by merging into a single rule. But it illustrates a risk of the file-based approach: when editing CSS, always check for duplicate selectors in the same file.


## 8. Debugging Playbook: "Why is my font wrong?"

When a font doesn't render as expected, follow this checklist:

### Step 1: Check computed style

```javascript
// Playwright evaluate
const el = document.querySelector('[data-part="documentation-tree-label"]');
console.log(getComputedStyle(el).fontFamily);
```

If it shows `"EB Garamond"` when you expected `"Chicago_"`, proceed to Step 2.

### Step 2: Walk the DOM chain

```javascript
let el = document.querySelector('[data-part="target-part"]');
while (el && el !== document.documentElement) {
  const cs = getComputedStyle(el);
  const part = el.getAttribute('data-part') || el.tagName;
  console.log({ element: part, fontFamily: cs.fontFamily });
  el = el.parentElement;
}
```

Find where the font switches from expected to unexpected.

### Step 3: Find all matching CSS rules

```javascript
const el = document.querySelector('.app-root');  // or your target
const sheets = document.styleSheets;
for (let i = 0; i < sheets.length; i++) {
  const sheet = sheets[i];
  try {
    for (const rule of sheet.cssRules) {
      if (rule.selectorText && el.matches(rule.selectorText) && rule.style.fontFamily) {
        console.log({
          sheetIndex: i,
          selector: rule.selectorText,
          fontFamily: rule.style.fontFamily,
          ownerNodeId: sheet.ownerNode?.id,  // Key: identifies injected styles
          href: sheet.href?.split('/').pop() || 'inline',
        });
      }
    }
  } catch(e) { /* cross-origin sheets */ }
}
```

If you see multiple rules, the **last one by source order** with the highest specificity wins.

### Step 4: Check for browser extensions

```javascript
// Look for <style> tags with IDs that aren't from your app
const headStyles = document.querySelectorAll('head style[id]');
headStyles.forEach(s => console.log(s.id));
```

Common extension-injected IDs:
- `typography-palette-overrides`
- `typography-palette-highlight-style`
- `dark-reader-style`
- `stylus-*`

### Step 5: Check CSS variable resolution

```javascript
const root = document.documentElement;
const cs = getComputedStyle(root);
console.log({
  '--font-ui': cs.getPropertyValue('--font-ui'),
  '--font-serif': cs.getPropertyValue('--font-serif'),
  '--font-mono': cs.getPropertyValue('--font-mono'),
});
```

If a variable is empty or wrong, it may not be defined on `:root` or may be overridden by another rule.


## 9. The Complete Font Assignment: What We Changed

### 9.1 Files modified

| File | Change | Why |
|---|---|---|
| `web/public/fonts/EBGaramond12-Regular.woff` | **Added** | Garamond regular weight |
| `web/public/fonts/EBGaramond12-Italic.woff` | **Added** | Garamond italic |
| `web/public/fonts/EBGaramond12-Bold.woff` | **Added** | Garamond bold |
| `web/src/styles/global.css` | Added `@font-face` for EB Garamond; added `--font-serif` variable | Font loading and variable |
| `web/src/components/Markdown/styles/markdown.css` | `p`, `blockquote` → `var(--font-serif)` 15px; `h1-h3` → `var(--font-ui)` | Serif body, Chicago headers |
| `web/src/components/PackageIndex/styles/package-index.css` | `package-index-section-short` → `var(--font-serif)` 14px | Serif descriptions in index |
| `web/src/components/SectionView/styles/section-view.css` | `section-header-subtitle` → `var(--font-serif)` 14px; removed `font-family: var(--font-ui)` from `section-view`; added it to `section-header` | Serif subtitle, no forced font on body |
| `web/src/components/SectionList/styles/section-list.css` | `section-card-short` → `var(--font-serif)` 11px | Serif descriptions in sidebar |
| `web/src/components/DocumentationTree/styles/documentation-tree.css` | `documentation-tree-root` → `font-family: var(--font-ui)` | Defensive Chicago on tree |

### 9.2 Elements that should be serif (EB Garamond)

- `[data-part='markdown-content'] p` — article paragraphs
- `[data-part='markdown-content'] blockquote` — article quotes
- `[data-part='package-index-section-short']` — index descriptions
- `[data-part='section-header-subtitle']` — article page subtitle
- `[data-part~='section-card-short']` — sidebar card descriptions

### 9.3 Elements that should be Chicago (bitmap)

Everything else, including:
- All headers (h1, h2, h3) everywhere
- Title bar, menu bar, status bar
- Sidebar tree labels and rows
- Package index headings, links, count text
- Search input, package selector, navigation toggle, type filter
- Section card titles
- Buttons, badges, empty states

### 9.4 Elements that should be Monaco (mono)

- Code blocks and inline code in markdown
- Section header slug label


## 10. Open Issues and Recommendations

### 10.1 The browser extension conflict (HIGH PRIORITY)

The `<style id="typography-palette-overrides">` extension overrides `.app-root` with EB Garamond, making the entire app serif by default, then patches individual elements back to Chicago. This is the **opposite** of our approach (Chicago by default, serif only on specific elements).

**Action needed:** Either:
1. Disable/remove the extension for this project, OR
2. Make the extension aware of the `data-part` convention so it uses the same approach

### 10.2 Migrate `.app-root` to `data-part` (MEDIUM)

```diff
- <div className="app-root">
+ <div data-part="app-root">
```

```diff
- .app-root { font-family: var(--font-ui); }
+ [data-part='app-root'] { font-family: var(--font-ui); }
```

This makes the entire system consistent: all styling targets `data-part` attributes.

### 10.3 Consider `@layer` for cascade control (LOW)

```css
@layer reset, tokens, components;

@layer tokens {
  :root {
    --font-ui: 'Chicago_', ...;
    --font-serif: 'EB Garamond', ...;
  }
}

@layer components {
  [data-part='markdown-content'] p {
    font-family: var(--font-serif);
  }
}
```

Unlayered styles (like browser extensions) would then have lower priority than layered component styles.

### 10.4 Add a font debug page (NICE-TO-HAVE)

A hidden route (e.g., `/_debug/fonts`) that renders every `data-part` element with its computed font-family would make future debugging much faster.

---

## 11. Quick Reference: File Locations

```
web/
├── public/
│   ├── fonts/
│   │   ├── ChicagoFLF.woff2           ← --font-ui source
│   │   ├── EBGaramond12-Regular.woff  ← --font-serif regular
│   │   ├── EBGaramond12-Italic.woff   ← --font-serif italic
│   │   └── EBGaramond12-Bold.woff     ← --font-serif bold
│   └── site-config.js
├── src/
│   ├── styles/
│   │   └── global.css                 ← @font-face, :root variables, .app-root
│   ├── components/
│   │   ├── AppLayout/styles/           ← layout (no font decls)
│   │   ├── TitleBar/styles/            ← title bar (no font decls)
│   │   ├── MenuBar/styles/             ← menu bar (no font decls)
│   │   ├── SearchBar/styles/           ← search input (no font decls)
│   │   ├── PackageSelector/styles/     ← dropdown (no font decls)
│   │   ├── NavigationModeToggle/styles/ ← toggle buttons (no font decls)
│   │   ├── TypeFilter/styles/          ← filter buttons (no font decls)
│   │   ├── DocumentationTree/styles/   ← sidebar tree (var(--font-ui))
│   │   ├── SectionList/styles/         ← sidebar cards (var(--font-serif) on short)
│   │   ├── PackageIndex/styles/        ← index page (var(--font-serif) on short)
│   │   ├── SectionView/styles/         ← article page (var(--font-serif) on subtitle)
│   │   ├── Markdown/styles/            ← article body (var(--font-serif) on p)
│   │   ├── StatusBar/styles/           ← bottom bar (no font decls)
│   │   ├── Badge/styles/               ← tag badges (no font decls)
│   │   └── EmptyState/styles/          ← empty placeholder (no font decls)
│   ├── App.tsx                         ← main layout, <div class="app-root">
│   └── entry-client.tsx                ← imports global.css
└── index.html                          ← shell HTML, no inline styles
```
