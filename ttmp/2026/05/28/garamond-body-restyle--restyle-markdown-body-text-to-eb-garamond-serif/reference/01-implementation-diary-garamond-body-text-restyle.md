---
title: "Implementation Diary: Garamond Body Text Restyle"
doc-type: reference
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

# Implementation Diary: Garamond Body Text Restyle

## Goal

Record every step taken, what worked, what failed, and what was learned during the effort to switch markdown body text to EB Garamond while keeping all UI chrome in Chicago bitmap font.

---

## Step 1: Initial CSS change (WRONG APPROACH)

**What I did:** Set `font-family: var(--font-serif)` on the root `[data-part='markdown-content']` selector.

**What happened:** Garamond cascaded to ALL children — headers, links, list items, everything inside the markdown container became serif.

**Why it failed:** CSS inheritance. Setting the font on the container root means every child element inherits it unless it explicitly overrides. The headers had no `font-family` override, so they inherited Garamond.

**Lesson:** Don't set the font on a container. Set it on the specific elements that need it.

---

## Step 2: Narrowed to `p` only, added header overrides

**What I did:**
- Moved `font-family: var(--font-serif)` from the root to `[data-part='markdown-content'] p` only
- Added explicit `font-family: var(--font-ui)` on `h1, h2, h3` inside markdown-content
- Removed serif from `li, td, th` (they should inherit Chicago from the container)

**What happened:** The index page still showed Garamond everywhere — on titles, links, headings, sidebar.

**Why it failed:** I was only looking at the Markdown component's CSS. The index page uses completely different components (PackageIndex, SectionList) with their own `data-part` selectors. The `markdown-content` CSS only applies when viewing an actual article.

**Lesson:** The app has multiple "content" views (index, article, section header). Each needs its own serif treatment.

---

## Step 3: Added serif to PackageIndex, SectionView, SectionList descriptions

**What I did:**
- `[data-part='package-index-section-short']` → `var(--font-serif)`
- `[data-part='section-header-subtitle']` → `var(--font-serif)`
- `[data-part~='section-card-short']` → `var(--font-serif)`

**What happened:** Index page descriptions showed Garamond, but the sidebar tree was still rendering in Garamond too. User reported sidebar and titles should be Chicago.

**Why it failed:** The browser extension was overriding `.app-root` with EB Garamond, making everything serif, then patching specific elements back. Our CSS was fighting the extension.

**Lesson:** Always check for external style injection when CSS doesn't behave as expected.

---

## Step 4: Added defensive `font-family: var(--font-ui)` on documentation-tree-root

**What I did:** Added `font-family: var(--font-ui)` to `[data-part='documentation-tree-root']`.

**What happened:** Still Garamond in the tree.

**Why it failed:** The extension's `.app-root { font-family: serif }` was winning over our CSS because of source order.

---

## Step 5: Deep investigation with Playwright

**What I did:** Used Playwright `evaluate` to:
1. Check computed `fontFamily` on every `data-part` element
2. Walk the DOM chain from tree root to body
3. Enumerate ALL matching CSS rules from ALL stylesheets
4. Identify the `<style id="typography-palette-overrides">` element

**What I found:**
- The browser had a `<style id="typography-palette-overrides">` tag injected by a browser extension
- It contained 26 rules that set `.app-root` to EB Garamond and patched individual parts back to Chicago
- It appeared after all Vite stylesheets, winning by source order
- It also set `documentation-tree-row` to serif explicitly

**This was the root cause of all the confusing behavior.**

---

## Step 6: Removed `font-family` from section-view root

**What I did:** Removed `font-family: var(--font-ui)` from `[data-part='section-view']`.

**Why:** Previously, `section-view` forced Chicago on all children. This meant `markdown-content p` couldn't inherit Garamond from its own CSS without an explicit override. The paragraph's own `font-family: var(--font-serif)` would win by specificity, but the forced parent font made the cascade confusing.

After removal, `section-view` inherits from `.app-root` (Chicago), and `markdown-content p` overrides to Garamond cleanly.

---

## Step 7: Added explicit Chicago to section-header

**What I did:** Added `font-family: var(--font-ui)` to `[data-part='section-header']`.

**Why:** Since `section-view` no longer forces Chicago, the header needs its own declaration. The header contains title, slug, subtitle — we want all of those to be Chicago by default, with only subtitle overriding to serif.

---

## Step 8: Merged duplicate `p` rules in markdown.css

**What I did:** Combined two `[data-part='markdown-content'] p` rules into one:

```css
[data-part='markdown-content'] p {
  font-family: var(--font-serif);
  font-size: 15px;
  line-height: 1.7;
  margin: 0.8em 0;
}
```

**Why:** The second rule (with only `margin`) was silently overriding the first rule's `font-family`, `font-size`, and `line-height` because when you redeclare a selector, properties not listed in the second rule don't carry over — they just don't get overridden. Wait, actually CSS doesn't work that way — the second rule's properties override the first, but unmentioned properties from the first are kept. The real issue was that the second rule was adding `margin` that we wanted. The merge was just for clarity.

---

## Current State

The CSS changes are correct in the source files. The remaining issue is the **browser extension** which overrides `.app-root` with Garamond. Once the extension is disabled or its rules aligned with our approach, the fonts should render correctly:

- **Chicago** on all UI chrome (sidebar, headers, buttons, titles, links)
- **EB Garamond** on paragraph-level text only (descriptions, article body, blockquotes)
- **Monaco** on code

## Files Changed

| File | Change |
|---|---|
| `web/public/fonts/EBGaramond12-Regular.woff` | Added |
| `web/public/fonts/EBGaramond12-Italic.woff` | Added |
| `web/public/fonts/EBGaramond12-Bold.woff` | Added |
| `web/src/styles/global.css` | Added `@font-face` × 3, added `--font-serif` |
| `web/src/components/Markdown/styles/markdown.css` | Serif on `p`, `blockquote`; Chicago on `h1-h3` |
| `web/src/components/PackageIndex/styles/package-index.css` | Serif on `section-short` |
| `web/src/components/SectionView/styles/section-view.css` | Serif on `subtitle`; removed forced Chicago from `section-view`; added Chicago to `section-header` |
| `web/src/components/SectionList/styles/section-list.css` | Serif on `section-card-short` |
| `web/src/components/DocumentationTree/styles/documentation-tree.css` | Defensive Chicago on tree root |
