---
title: Diary
doc-type: reference
ticket: GL-012
topics: typography, css, frontend, debug-tooling
status: active
owners: [manuel]
intent: long-term
created: 2026-05-28
---

# Diary

## Goal

Capture the implementation journey for GL-012: creating a Typography Debug Palette for the Glazed docs site, including persistence, custom presets, CSS export, and reMarkable upload.

---

## Step 1: Exploration and Understanding

The goal was to deeply understand the Glazed docs site typography system before designing the debug palette. I read every relevant file in the `web/` directory — global.css, all 13 component CSS files, the App.tsx, store.ts, types, API service, Vite config, Storybook config, the font file notice, and the Go server code.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to improve the docs site typography. Mostly, I want to try out different typography settings for the different font styles we have: font, size, weight, color (still monochrome), for the different elements. I could image a little debug mode palette that allows me to setup the different elements and see the results. Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and then upload to remarkable. Start with just the document, add little ascii screenshot of what the palette would look like, different options. Keep a diary as you work. work within this directory, in ./glazed and its docs serving web part."

**Assistant interpretation:** Create a comprehensive intern-ready analysis/design/implementation guide for a Typography Debug Palette feature, store it in a docmgr ticket, keep a diary, and upload to reMarkable.

**Inferred user intent:** Manuel wants a detailed technical document that serves as both a design spec and an onboarding guide, so an intern (or future self) can implement the feature without further questions.

### What I did

- Read all CSS files, App.tsx, store.ts, types, API service, Vite config, Go server code
- Created docmgr ticket GL-012
- Wrote 74KB design document with 16 sections, 6 ASCII wireframes, full typography audit
- Uploaded to reMarkable at `/ai/2026/05/28/GL-012`

### Why

A thorough audit of every typography value is essential for the element registry.

### What worked

- The `data-part` selector pattern makes targeting elements easy
- The complete audit table gives a single reference for every value
- reMarkable upload succeeded on first try

### What didn't work

- N/A

### What I learned

- Only 2 font families but ~30 distinct typography contexts
- Many values are hardcoded in component CSS (not using CSS variables)
- Markdown headings use `em` units, so they cascade from prose body size
- `--color-accent` is currently `#000000` — links are black

### What was tricky to build

Deciding the override architecture: chose hybrid `:root` variables + injected `<style>` rules approach.

### What warrants a second pair of eyes

- CSS selector specificity of injected rules
- em-unit heading computation when prose body size is also overridden

### What should be done in the future

- Extract chosen values back into component CSS files after experimentation
- Add "Copy CSS" export (done in Step 2)
- Convert hardcoded values to CSS variables

---

## Step 2: Full Implementation with Persistence, Custom Presets, and CSS Export

### Prompt Context

**User prompt (verbatim):** "Cook, we do want to keep settings and allow creating new presets and save in local storage. Also, we should be able to export the current settings or a preset to clipboard (as css?) so that we can update the actual files / prompt a coding agent to do the changes for us. Implement this system, run vite in a tmux shell, and keep a diary as you work, committing at appropriate intervals (including ticket files)"

**Assistant interpretation:** Implement the full Typography Debug Palette with persistence (localStorage), custom preset creation, CSS export to clipboard, run dev server, commit at intervals.

**Inferred user intent:** Ship a working implementation, not just a design doc. The persistence and export features are critical — the palette must be a practical tool, not just a toy.

### What I did

1. Started Vite dev server in tmux session `glazed-docs` on port 5173
2. Started Go API server in tmux session `glazed-api` on port 8088
3. Created 18 new files:
   - `web/src/types/typography-palette.ts` — type definitions (GrayColor, FontFamily, FontWeight, TypographyProperties, TypographyOverrides, TypographyPreset, TypographyGroup, TypographyElement, PersistedPaletteState)
   - `web/src/store/typographyPaletteSlice.ts` — Redux slice with persistence
   - `web/src/components/TypographyPalette/element-registry.ts` — 13 groups, 30+ elements
   - `web/src/components/TypographyPalette/presets.ts` — 4 built-in presets
   - `web/src/components/TypographyPalette/css-override-engine.ts` — CSS generation + clipboard export
   - `web/src/components/TypographyPalette/persistence.ts` — localStorage save/load/clear
   - `web/src/components/TypographyPalette/TypographyPalette.tsx` — main panel
   - `web/src/components/TypographyPalette/TypographyPaletteGroup.tsx` — accordion group
   - `web/src/components/TypographyPalette/TypographyPaletteElement.tsx` — per-element controls
   - `web/src/components/TypographyPalette/FontFamilySelect.tsx` — font dropdown
   - `web/src/components/TypographyPalette/FontSizeStepper.tsx` — size +/− stepper
   - `web/src/components/TypographyPalette/FontWeightSelect.tsx` — weight dropdown
   - `web/src/components/TypographyPalette/ColorStepper.tsx` — gray shade stepper
   - `web/src/components/TypographyPalette/parts.ts` — data-part constants
   - `web/src/components/TypographyPalette/styles/typography-palette.css` — palette styles
   - `web/src/components/TypographyPalette/useTypographyOverrides.ts` — Redux→DOM sync hook
   - `web/src/components/TypographyPalette/usePaletteShortcut.ts` — Ctrl+Shift+T hook
4. Modified 3 existing files:
   - `web/src/store.ts` — added typographyPalette reducer
   - `web/src/App.tsx` — render palette + shortcut hook + fragment wrapper
   - `web/src/components/StatusBar/StatusBar.tsx` — added 𝒜a dev-only toggle button
5. Verified in browser:
   - Palette opens/closes with 𝒜a button
   - 13 accordion groups expand/collapse correctly
   - All controls (font, size, weight, color, line-height) work
   - Dense Terminal preset applies monospace 12px instantly
   - CSS override engine injects rules into `<style id="typography-palette-overrides">`
   - Export menu shows "Copy as CSS rules" and "Copy as CSS variables"
   - ★ Save button visible when overrides exist
6. Committed: `117db61 feat(web): implement Typography Debug Palette`
7. Committed ticket docs: `2d6700b docs(GL-012): add typography debug palette design doc and diary`

### Why

The design doc laid out the architecture; this step makes it real.

### What worked

- Vite HMR picked up all changes instantly — no page reload needed
- The `data-part` selectors work perfectly for targeting specific elements
- The CSS override engine generates correct rules (verified by inspecting the injected `<style>` element)
- The Redux slice architecture cleanly separates concerns (state, persistence, DOM effects)
- localStorage persistence works — overrides survive page refresh

### What didn't work

- Initial attempt used `go run ./cmd/glaze help serve --address :8088` which was wrong syntax; correct command is `go run ./cmd/glaze serve --address :8088`
- The SSR hydration mismatch warnings existed before my changes (the SSR sidecar is not running, so the client hydration always fails — this is a pre-existing issue, not introduced by the palette)

### What I learned

- The `import.meta.env.DEV` guard works perfectly with Vite — in production builds, the dead code is eliminated
- The `<style>` element injection approach is very effective — you can see the rules in DevTools and they override the component CSS by cascade order (later in the document)
- The fragment wrapper (`<> ... </>`) in App.tsx is needed because the palette is a fixed-position overlay, not inside the `.app-root` div

### What was tricky to build

- **Export menu positioning:** The export menu appears above the footer as a dropdown. Using `position: absolute` with `bottom: 100%` works but needs a click-outside handler to close. Implemented with a `useRef` + `mousedown` event listener.
- **Save preset form:** The inline form needs to appear/disappear cleanly. Used a state toggle + auto-focus input. Enter to save, Escape to cancel.
- **Preset indicator:** When you manually edit after applying a preset, the active preset should clear (set to "Custom"). The slice handles this by setting `activePreset = null` on any `setOverride` action.

### What warrants a second pair of eyes

- The CSS specificity: injected `<style>` rules use the same selectors as component CSS. This works because the injected stylesheet appears later in the document, winning by cascade order. But if any component CSS uses `!important`, the overrides would break. I didn't find any `!important` in the existing CSS.
- The `TypographyPaletteElement` merge of defaults + overrides: `{ ...element.defaults, ...currentOverrides }` means any undefined property in overrides falls back to defaults. This is correct but means partial overrides (e.g., only changing color) correctly preserve other defaults.

### What should be done in the future

- Verify the palette works correctly in production builds (the dev-only guard should hide it)
- Add Storybook stories for the palette components
- Add unit tests for the CSS override engine
- Consider adding a "Delete all data" button to clear localStorage

### Code review instructions

- Start with `web/src/components/TypographyPalette/TypographyPalette.tsx` — the main panel component
- Check `web/src/store/typographyPaletteSlice.ts` — the Redux slice with persistence
- Verify `web/src/components/TypographyPalette/css-override-engine.ts` — the CSS generation
- Check `web/src/App.tsx` for the palette integration (import, hook, render)
- Run `pnpm dev` and test the palette interactively

### Technical details

- Commit: `117db61` — "feat(web): implement Typography Debug Palette for docs site"
- Commit: `2d6700b` — "docs(GL-012): add typography debug palette design doc and diary"
- Vite dev server: tmux session `glazed-docs`, port 5173
- Go API server: tmux session `glazed-api`, port 8088
- 20 files changed, 1753 insertions
- localStorage key: `glazed-typography-palette`

---

## Step 3: Baseline Design System, Scale Mode, and Spacing Controls

### Prompt Context

**User prompt (verbatim):** "also allow adjusting line spacing (vertical and horizontal). also, in order to minimize visual styles, allow defining baseline parmeters that can be selected, and then a override mode, at least for the styles where it makes sense to have some design system, because thats alot of settings."

**Assistant interpretation:** Add letter-spacing/word-spacing controls, and add a baseline design system with scale ratios so elements can reference steps instead of absolute values. This reduces the number of individual settings users need to tweak.

**Inferred user intent:** The palette has too many individual settings. A design system approach (baseline params + scale steps) lets you set a few global values and have all elements derive their sizes from those. This is how real design systems work (e.g., Tailwind, Material Design type scale).

### What I did

1. Added `BaselineParameters` type and `DEFAULT_BASELINE` constants to `types/typography-palette.ts`
2. Added `ScaleRatioName`, `SCALE_RATIOS`, `SCALE_STEP_LABELS`, `computeScaledValue()` utility
3. Added `ElementSizeMode` ('custom' | 'scale'), `ElementScaleSteps`, `ElementSizeModeMap` types
4. Added `letterSpacing` and `wordSpacing` to `TypographyProperties`
5. Updated `element-registry.ts`: added `supportsScale`, `defaultFontSizeStep`, `defaultLineHeightStep` to 25+ elements; added `letterSpacing` and `wordSpacing` to adjustable properties on prose, headings, and code
6. Created `BaselineParameters.tsx` component: base font size, scale ratio dropdown, line height, letter spacing, word spacing + scale preview
7. Created `ScaleStepSelect.tsx` component: dropdown showing step labels (xs/−3 through 5xl/+6) with computed values
8. Updated `TypographyPaletteElement.tsx`: added Custom/Scale mode toggle per element, ScaleStepSelect for size in scale mode
9. Updated `useTypographyOverrides.ts`: resolve scale-mode elements to concrete CSS using `computeScaledValue()`, merge with custom overrides
10. Updated Redux slice with `setBaseline`, `setElementMode`, `setElementScaleSteps` actions
11. Updated persistence to save/load baseline, elementModes, elementScaleSteps
12. Added "Scale System (1.25)" preset that puts all elements in scale mode with Major Third ratio and base 16px
13. Updated CSS override engine to emit `letter-spacing` and `word-spacing` declarations
14. Verified in browser: Scale System preset correctly computes all sizes from baseline
15. Committed: `8488851 feat(web): add baseline design system, scale mode, and spacing controls`

### Why

The v1 palette had 30+ individual settings. Without a design system, every element is independent — changing the overall scale requires adjusting every element manually. The baseline + scale steps approach means you set 5 global params and every scale-mode element derives its value automatically.

### What worked

- The `computeScaledValue(base, ratio, step)` utility is simple and correct
- Scale preview row in baseline panel shows computed sizes at a glance
- Custom/Scale toggle per element gives fine-grained control
- The "Scale System (1.25)" preset demonstrates the full design system approach
- CSS override engine correctly generates rules for scale-mode elements
- HMR picked up all changes instantly

### What didn't work

- No major failures. The implementation went smoothly.

### What I learned

- Modular scales are powerful: Major Third (1.25) at base 16px gives: 8.19, 10.24, 12.8, 16, 20, 25, 31.25, 39.06 — these are the classic "design system" sizes
- Em-based elements (markdown headings) need special handling in scale mode: the step applies to the em multiplier, not the px value, since the parent's px is already scaled
- Letter spacing at 0.01–0.05em is the sweet spot for readability improvement without looking odd
- The persistence layer needed to expand to store baseline + modes + steps, not just overrides

### What was tricky to build

- **Scale mode for em-based elements:** Markdown headings use `em` units (1.6em, 1.3em, etc.). In scale mode, the step computes a multiplier, not a px value. For `headings.h1` at step +4 with ratio 1.25: `1 × 1.25^4 = 2.44em`. This is close to but not identical to the CSS default of 1.6em. The step values give a different but consistent scale.
- **Resolving scale to CSS:** The `useTypographyOverrides` hook now does a two-pass: first resolve all scale-mode elements to concrete properties, then merge custom overrides on top. This means you can use scale mode for most elements and override specific ones with custom values.
- **Line height in scale mode:** Line height isn't really a "scale step" — it's a multiplier. I implemented it as an offset from the baseline: `baseLineHeight + step × 0.1`. Step 0 = exactly the baseline, step +1 = 0.1 more, etc.

### What warrants a second pair of eyes

- The em-based scale computation: when prose body is 16px and h1 step is +4, `computeScaledValue(1, 1.25, 4) = 2.44em`. The actual rendered size would be 16 × 2.44 = 39px. Is this the right behavior? Alternative: compute px values and convert to em at the end.
- The line height offset approach (step × 0.1) is simple but arbitrary. A better approach might be to scale line height proportionally with font size.

### What should be done in the future

- Consider a "Convert to CSS" action that takes the current scale-mode state and outputs CSS custom properties that can be pasted into global.css
- Add a visual preview of the type scale (all steps rendered with sample text)
- Consider allowing custom scale ratios (not just the 8 named ones)
- The presets should store and restore baseline + mode + steps (partially done)

### Code review instructions

- Start with `web/src/types/typography-palette.ts` — all the new types
- Check `web/src/components/TypographyPalette/useTypographyOverrides.ts` — the scale resolution logic
- Check `web/src/components/TypographyPalette/BaselineParameters.tsx` — the baseline UI
- Check `web/src/components/TypographyPalette/TypographyPaletteElement.tsx` — the Custom/Scale toggle
- Verify the "Scale System (1.25)" preset works by selecting it in the browser

### Technical details

- Commit: `8488851` — "feat(web): add baseline design system, scale mode, and spacing controls"
- 13 files changed, 878 insertions, 96 deletions
- New components: BaselineParameters.tsx, ScaleStepSelect.tsx
- localStorage now stores: overrides + baseline + elementModes + elementScaleSteps

---

## Step 4: Writing the Obsidian Vault Article

### Prompt Context

**User prompt (verbatim):** "Write a detailed project report about the work you did as a deep dive technical report for other developers and designers who might want to implement a similar system, for our obsidian vault, in a textbook writing style (no analogies, see skill)"

**Assistant interpretation:** Write a comprehensive technical deep-dive article in the Obsidian vault using textbook-authoring style (Peter Norvig: foundational first, no analogies, concrete examples, prose paragraphs with bullet rhythm breaks). Target audience: other developers/designers who want to build a similar system.

**Inferred user intent:** Create durable engineering knowledge — an article that outlives the specific project and teaches the pattern to future readers.

### What I did

1. Read the textbook-authoring skill and the obsidian-vault-writing skill
2. Read the existing ARTICLE exemplar (Go Wasm Browser Playbook) to match tone and structure
3. Re-read all 20 implementation files to ground the article in concrete details
4. Wrote a 32KB article with 9 major sections:
   - Why this system exists
   - Architecture overview (with mermaid diagram)
   - The type system (TypographyProperties, BaselineParameters, ScaleSteps)
   - The element registry (full table of 13 groups)
   - The CSS override engine (rule generation, export formats)
   - The resolution layer (scale/custom merge algorithm)
   - The Redux slice (state categories, action design)
   - The element control component (Custom/Scale toggle, spacing controls)
   - Presets, persistence, keyboard shortcut, file inventory
5. Added design decisions section with rationale for each key choice
6. Added common failure modes section
7. Added working rules section
8. Committed to vault: `96f56a4 ARTICLE: Typography Debug Palette`
9. Uploaded to reMarkable at `/ai/2026/05/28/GL-012`

### Why

The article captures the design knowledge in a form that other teams can reuse. The ticket docs are project-specific; the vault article is durable and discoverable.

### What worked

- The textbook-authoring style produced clean, direct prose without analogies
- The mermaid architecture diagram is clear and shows data flow between layers
- The code snippets are grounded in actual implementation
- The failure modes section covers the three most likely issues

### What didn't work

- N/A

### What I learned

- The Peter Norvig style works well for technical reports: explain why before how, use concrete code, break rhythm with tables and diagrams
- Obsidian mermaid diagrams render natively in the vault

### What warrants a second pair of eyes

- The resolution algorithm description — verify it matches the actual `useTypographyOverrides` implementation
- The failure modes — are there others we've hit during development that should be documented?

### What should be done in the future

- Add a section on testing strategy (unit tests for the override engine, integration tests for the palette)
- Consider adding interactive exercises if the article is ever used for onboarding

### Code review instructions

- Read the article at: `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/28/ARTICLE - Typography Debug Palette - Design System, Live Overrides, and Modular Scale.md`
- Verify the architecture diagram matches the actual component structure
- Check that the code snippets are accurate

### Technical details

- Article: 32,669 bytes, 537 lines
- Vault commit: `96f56a4`
- reMarkable: uploaded to `/ai/2026/05/28/GL-012`

---

## Step 5: Add EB Garamond Serif Font and Serif Presets

### Prompt Context

**User prompt (verbatim):** "also make presets with serif body as well (apple garamond? or close)"

**Assistant interpretation:** Add a serif body font (closest free equivalent to Apple Garamond) and create presets that use it.

**Inferred user intent:** Expand the palette's font selection with a serif option for editorial/reading presets, making the tool useful for experimenting with traditional book-like typography.

**Commit (code):** `11b25d9` — "feat(typography-palette): add EB Garamond serif font and two serif presets"

### What I did

- Downloaded EB Garamond woff2 files (latin + latin-ext subsets, weights 400/500/600/700) from @fontsource npm package
- Added 5 `@font-face` declarations in `global.css` (400 normal, 400 italic, 500 normal, 600 normal, 700 normal) with `local()` fallback
- Extended `FontFamily` type from `'ui' | 'mono'` to `'ui' | 'mono' | 'serif'`
- Added `FONT_STACKS.serif`: `'EB Garamond', 'Garamond', 'Georgia', 'Palatino', 'Times New Roman', serif`
- Added `fontFamily` to `adjustable` for headings.h1/h2/h3 and header.heading in element-registry.ts
- Created two new built-in presets:
  - **Serif Editorial**: Garamond body+headings, Chicago_ chrome, generous line height (1.75), slight letter/word spacing
  - **Serif Scale**: Full design system with Garamond, all elements in scale mode, Major Third ratio at 17px base
- Updated `fonts/NOTICE.md` with EB Garamond SIL OFL 1.1 license info
- Verified both presets render correctly in the browser (EB Garamond 400/600/700 all loaded)

### Why

Apple Garamond is proprietary. EB Garamond is the closest open-source equivalent — digitized by Georg Duffner from the original Garamond typeface designs, available under SIL OFL 1.1. It has the same old-style serif characteristics: high stroke contrast, humanist letterforms, comfortable reading rhythm.

### What worked

- Using @fontsource npm package to get properly subset woff2 files — much easier than scraping Google Fonts URLs
- Adding the `local()` fallback in `@font-face` src — users who have EB Garamond installed system-wide will use that instead of downloading
- Serif Scale preset demonstrates that the design system works with any font family, not just the default Chicago_

### What didn't work

- First attempt to download woff2 from Google Fonts gstatic URLs returned 404 HTML pages (1.6KB). The gstatic URLs from the CSS API are TTF, not woff2, and the subset-specific URLs are unstable across versions. @fontsource was the reliable path.

### What I learned

- Google Fonts CSS API serves different formats based on `Accept` header, but the woff2 URLs are not stable — they change with each font version update
- @fontsource packages fonts as individual subset+weight woff2 files, which is exactly what you need for vendoring
- EB Garamond has excellent weight coverage (400–800) including italic, making it a strong candidate for body text presets

### What was tricky to build

- The `@font-face` src stacking: `src: local('EB Garamond'), url('latin-ext') format('woff2'), url('latin') format('woff2')`. The latin-ext subset includes more glyphs but is larger; browsers load the smallest file that covers the needed characters. Placing latin-ext first ensures extended Latin support when needed, while the latin-only file serves as a smaller fallback.

### What warrants a second pair of eyes

- Verify the @font-face `src` ordering: should `local()` come before the first `url()`, or does it need to be in a separate `src` descriptor? Current spec says `local()` can be part of the same `src` list, and browsers try each source in order.
- Verify that the `font-display: swap` on serif faces doesn't cause layout shift when the font loads after initial render.

### What should be done in the future

- Consider adding italic-specific presets (EB Garamond italic is beautiful for blockquotes)
- Consider a "Book" preset with even more generous line height (1.9+) and wider letter spacing for sustained reading
- Test with long-form markdown content that uses many italic passages

### Code review instructions

- Check `global.css` for correct @font-face syntax
- Verify preset values in `presets.ts` produce good visual results at different viewport sizes
- Confirm FONT_STACKS fallback chain is sensible

### Technical details

- Font files: 10 woff2 files, total ~460KB on disk (latin-ext: 57–65KB each, latin: 22–23KB each)
- FontFamily type now has 3 values: 'ui' | 'mono' | 'serif'
- 7 built-in presets total (was 5)

---

## Step 6: Typeface Role System (Display / Body / Code)

### Prompt Context

**User prompt (verbatim):** "allow me to set different typefaces for titles / tree nav etc... For example I would use chicago as title / menu / headers, monaco for code, and EB garamond for body text. I think that's the big three from which the rest derives."

**Assistant interpretation:** Add a typeface role system with three roles — Display, Body, Code — that cascade font family assignments to element groups. This replaces per-element font family picking with a higher-level abstraction.

**Inferred user intent:** The user wants to control font family at the role level, not the element level. Changing "Body" from serif to ui should update all body-assigned elements (prose, tree, cards, blockquote) in one action. The three roles capture the fundamental typographic structure of a documentation browser.

**Commit (code):** `a1c83c0` — "feat(typography-palette): add typeface role system (Display/Body/Code)"

### What I did

1. Added `TypefaceRole` type (`'display' | 'body' | 'code'`), `TypefaceRoleMap`, and `DEFAULT_TYPEFACE_ROLES` to `types/typography-palette.ts`
2. Added `typefaceRole: TypefaceRole` field to every element in `element-registry.ts`:
   - **Display**: titlebar.title, menubar.items/appname, sidebar controls, header.heading, statusbar, badges
   - **Body**: root.body, prose.body, tree.row/heading, cards.title/short, header.subtitle, blockquote, links, table headers
   - **Code**: header.slug, code.inline, code.block
3. Added `typefaceRoles` state + `setTypefaceRole` action to Redux slice
4. Updated resolution layer: for each element, if no explicit fontFamily override exists, inherit from the element's `typefaceRole` mapping
5. Added 🔤 Typeface Roles section to BaselineParameters panel with three dropdowns
6. Simplified presets: serif presets now use `typefaceRoles: { display: 'ui', body: 'serif', code: 'mono' }` instead of per-element `fontFamily` overrides
7. Added Dense Terminal preset role: `{ display: 'mono', body: 'mono', code: 'mono' }`
8. Updated persistence to include `typefaceRoles` with graceful migration from older localStorage state

### Why

Per-element font family selection is correct for fine-tuning but wrong as the primary mechanism. The user's mental model is "I want Garamond for reading, Chicago for chrome, Monaco for code" — three decisions, not thirty. The role system maps this mental model directly. Role assignments cascade to element groups; per-element overrides remain available as escape hatches.

### What worked

- Adding `typefaceRole` to the element registry was the right place to define the default role assignment — it's co-located with the element definition
- The resolution layer naturally fits between scale-mode computation and custom override merging: role → scale → custom
- Verified interactively: changing Body from serif to ui in the dropdown immediately switches all body-assigned elements

### What didn't work

- N/A

### What I learned

- The three-role model (Display/Body/Code) is a standard pattern in design systems. Material Design uses display/headline/body/label/code. The user's model is even simpler — three roles cover everything in a documentation browser.
- Headings (h1/h2/h3) are assigned to 'display' because they function as titles/signposts, not reading text. This matches the user's intent ("Chicago as title / menu / headers").
- Tree navigation is assigned to 'body' because it's reading-oriented — you scan tree labels the same way you scan body text.

### What was tricky to build

- The role resolution must happen before custom override merging, because per-element fontFamily overrides should win over the role assignment. This means the resolution loop iterates over all elements (not just scale-mode elements) to apply role fonts, then merges custom overrides on top.

### What warrants a second pair of eyes

- Verify the role assignments in element-registry.ts match the user's expectation. Specifically: are headings really "display" or should they be "body"? The user said "chicago as title / menu / headers" — headers are display.
- Verify that per-element fontFamily overrides still work when a role is set (the escape hatch).

### What should be done in the future

- Consider adding a fourth role: "Navigation" (for tree, sidebar controls) — but the current three seem sufficient
- Consider showing which role each element belongs to in the accordion group UI

### Code review instructions

- Check `element-registry.ts` for typefaceRole assignments on all 30+ elements
- Verify resolution logic in `useTypographyOverrides.ts`: role fonts → scale computed → custom overrides
- Test: select Serif Editorial preset, then change Body role to ui — all body text should switch to Chicago_

### Technical details

- 8 files changed, 265 insertions, 152 deletions (net +113 lines)
- Serif presets simplified: ~20 fontFamily overrides removed, replaced by 3-line typefaceRoles
- Commit: `a1c83c0`
