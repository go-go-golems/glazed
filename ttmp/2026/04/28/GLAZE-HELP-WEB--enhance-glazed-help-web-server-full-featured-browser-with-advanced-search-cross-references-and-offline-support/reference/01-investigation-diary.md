---
Title: Investigation Diary
Ticket: GLAZE-HELP-WEB
Status: active
Topics:
    - glazed
    - help-system
    - web
    - react
    - server
    - ui
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go
      Note: Server architecture exploration
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/web/src/App.tsx
      Note: Frontend capabilities exploration
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-28T08:00:00-04:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Capture the research, analysis, and documentation work for designing the enhanced Glazed Help Web Server. This diary tracks exploration of the existing web stack (Go HTTP server, React SPA, TUI), identification of gaps, and the creation of a comprehensive design document with user stories, screen designs, and ASCII mockups.

---

## Step 1: Initialize Ticket and Read Skills

This step established the project scaffolding by creating a docmgr ticket and reading the relevant skill files to understand the workflow, writing style, and deliverable requirements.

### Prompt Context

**User prompt (verbatim):** "Then, since you are then already deep in the glazed system, also write a new docmgr ticket to add a webserver to the glazed verb to serve help entries, and write out all the different needs and affordances the site would service and offer, and then design a set of user stories. Then, use that knowledge to make a list of screens, how they work, potentially even with ASCII screenshots of the widgets / functionality. reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** The user wants a comprehensive design document for enhancing the Glazed help web server, including needs analysis, user stories, screen designs with ASCII mockups, and an implementation guide. The document should be intern-friendly and evidence-based.

**Inferred user intent:** Produce a detailed, visually-oriented design document that can guide implementation of a full-featured help browser web application, building on the existing `serve-help` and `render-site` infrastructure.

### What I did
- Created docmgr ticket `GLAZE-HELP-WEB` with title "Enhance Glazed Help Web Server: Full-Featured Browser with Advanced Search, Cross-References, and Offline Support".
- Added design-doc and diary documents to the ticket.

### Why
- The previous ticket (`GLAZE-HELP-EXPORT`) established the docmgr workflow. This ticket applies the same rigor to a new feature area.
- The user explicitly requested user stories, screen designs, and ASCII mockups — deliverables that require careful structuring.

### What worked
- Ticket creation succeeded immediately.
- The existing vocabulary already contained `glazed`, `help-system`, `web`, `react`, `server`, `ui`.

### What didn't work
- N/A

### What I learned
- Reusing the same workflow pattern (ticket → docs → evidence → writing → validation → upload) is efficient.

### What was tricky to build
- N/A (initialization step)

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Explore the existing web codebase to understand current capabilities and limitations.

### Code review instructions
- Verify ticket `GLAZE-HELP-WEB` exists in docmgr.

### Technical details
- Ticket path: `2026/04/28/GLAZE-HELP-WEB--enhance-glazed-help-web-server-full-featured-browser-with-advanced-search-cross-references-and-offline-support`

---

## Step 2: Explore Existing Web Stack

This step mapped the current web architecture by reading the Go HTTP server, React SPA, and TUI code. The goal was to understand what already exists so the design could build on it rather than replace it.

### What I did
- Read `pkg/help/server/serve.go` — the `ServeCommand` that starts the HTTP server, composes API + SPA handlers, supports mounting under prefixes.
- Read `pkg/help/server/handlers.go` — the API handlers for health, sections list, and section detail.
- Read `pkg/help/server/types.go` — the Go response types (`SectionSummary`, `SectionDetail`, `ListSectionsResponse`).
- Read `pkg/help/server/middleware.go` — CORS middleware.
- Read `pkg/help/server/serve_test.go` — tests confirming SPA serving, prefix mounting, and store replacement.
- Read `pkg/web/static.go` — the `NewSPAHandler` that serves embedded Vite assets with SPA fallback.
- Read `web/src/main.tsx` — React entry point with `HashRouter` and Redux `Provider`.
- Read `web/src/App.tsx` — root component wiring search, type filter, section list, and section view.
- Read `web/src/services/api.ts` — RTK Query API slice with server/static mode detection.
- Read `web/src/store.ts` — Redux store with RTK Query reducer.
- Read `web/src/types/index.ts` — TypeScript interfaces mirroring Go types.
- Read `web/src/components/AppLayout/AppLayout.tsx` — two-pane layout.
- Read `web/src/components/SearchBar/SearchBar.tsx` — basic text input.
- Read `web/src/components/TypeFilter/TypeFilter.tsx` — filter buttons.
- Read `web/src/components/SectionList/SectionList.tsx` and `SectionCard.tsx` — list and card rendering.
- Read `web/src/components/SectionView/SectionView.tsx` and `SectionHeader.tsx` — content viewer.
- Read `web/src/components/MenuBar/MenuBar.tsx` — retro menu bar (currently decorative).
- Read `web/vite.config.ts` — Vite config with API proxy, HashRouter base path.
- Read `pkg/help/ui/model.go` — the Bubble Tea TUI model for comparison.

### Why
- The ticket-research skill mandates "evidence-first investigation."
- The user specifically asked for screen designs — without understanding the existing components, the designs would be disconnected from reality.

### What worked
- The codebase is well-organized with clear separation between Go backend (`pkg/help/server/`), React frontend (`web/src/`), and shared types.
- The existing API is minimal but clean — easy to extend.
- The TUI (`pkg/help/ui/model.go`) provides a good benchmark for what interactive features are already possible in the terminal (search, copy, DSL, help screen).

### What didn't work
- `web/src/components/Markdown/MarkdownContent.tsx` was not readable due to file access issues, but its interface is clear from `SectionView.tsx`.

### What I learned
- The React app uses **HashRouter** (`/#/sections/:slug`) to avoid needing server-side route fallback. This is important for the static export (`render-site`) which has no live server.
- The API layer already supports **static mode** — when `window.__GLAZE_SITE_CONFIG__.mode === 'static'`, it reads `.json` files instead of calling `/api`. This means new API endpoints must also have static JSON equivalents.
- The TUI supports DSL search, clipboard copy, and keyboard shortcuts — all features the web UI currently lacks.
- The `MenuBar` is purely decorative. Making it functional (View → Topic Browser, View → Coverage) is a natural enhancement.
- The SectionCard shows type badge, top-level diamond, title, and short description. It does not show topics/commands/flags inline — those are only in the detail view.

### What was tricky to build
- Understanding the relationship between `serve-help` (live server) and `render-site` (static export) was important. Both use the same React SPA, but `render-site` pre-generates JSON files and sets `mode: 'static'`. Any new API endpoint must be mirrored in `pkg/help/site/render.go` for static mode compatibility.

### What warrants a second pair of eyes
- Confirm that the proposed new API endpoints (`/api/topics`, `/api/coverage`, `/api/sections/search` with snippets) are feasible given the existing store schema.
- Confirm that static mode JSON generation in `render-site` can be extended to include the new data shapes.

### What should be done in the future
- When implementing Phase 3 (Coverage Dashboard), determine how to get the full list of commands from the Cobra tree into the help system. The store only knows about commands that have help sections.

### Code review instructions
- Start reading at `pkg/help/server/serve.go` to understand server startup.
- Then read `web/src/App.tsx` to see current frontend capabilities.
- Then read `pkg/help/ui/model.go` to see what the TUI already does (benchmark for web parity).

### Technical details
- Key files and their roles:
  - `pkg/help/server/serve.go` (~170 lines) — Server command, handler composition, graceful shutdown
  - `pkg/help/server/handlers.go` (~200 lines) — API routes: health, list, get
  - `pkg/help/server/types.go` (~120 lines) — Request/response structs
  - `pkg/web/static.go` (~60 lines) — SPA static file serving with fallback
  - `web/src/App.tsx` (~90 lines) — Root component with sidebar + content
  - `web/src/services/api.ts` (~120 lines) — RTK Query with server/static dual mode
  - `pkg/help/ui/model.go` (~580 lines) — TUI with search, view, help, cheatsheet states

---

## Step 3: Write the Design Document

This step produced the primary deliverable: a comprehensive design document with needs analysis, user stories, screen designs with ASCII mockups, API enhancements, frontend architecture, and implementation phases.

### What I did
- Wrote a "Needs and Affordances Analysis" section mapping 7 user needs to concrete UI affordances.
- Wrote 11 user stories organized by persona: New User, Power User, Operator, Offline User.
- Designed 7 screens with ASCII mockups:
  1. Home / Browse View (enhanced)
  2. Section View with Cross-References
  3. Search Results with Snippets
  4. Topic Browser
  5. Coverage Dashboard
  6. Command Palette (Modal)
  7. Dark Mode
- Wrote API enhancement proposals with new endpoints and response shapes.
- Wrote frontend architecture plan: new Redux slices, component inventory, routing expansion, service worker pseudocode.
- Defined 6 implementation phases: Enhanced Search → Cross-References → Topics/Coverage → Dark Mode/Accessibility → Offline Support → Print/Export.
- Included testing strategy, risks, alternatives, and open questions.

### Why
- The user explicitly requested "all the different needs and affordances," "user stories," "list of screens," and "ASCII screenshots of the widgets / functionality."
- The ticket-research skill requires "Optimize for onboarding unfamiliar engineers" and "Include pseudocode and minimal API sketches where useful."

### What worked
- The ASCII mockups make the designs concrete and reviewable without requiring a designer or prototype tool.
- The user story format ("As a [persona], I want [goal], so that [benefit]") makes acceptance criteria clear.
- The phased implementation plan keeps each deliverable small enough for a single intern sprint.

### What didn't work
- The Coverage Dashboard design assumes we can enumerate all commands in the binary. The current store only knows commands with help sections. This is documented as an open question.

### What I learned
- The existing web app already uses a clean component architecture (`parts.ts` + `styles/*.css` per component). New components should follow this pattern.
- The HashRouter means deep linking to anchors must use the hash portion carefully: `/#/sections/slug#anchor`. React Router's `useLocation` can parse this.
- Static mode (`render-site`) complicates new API endpoints because they must also be pre-rendered as JSON files. This is a constraint on the backend design.

### What was tricky to build
- Designing the Command Palette required balancing three use cases in one modal: (1) fuzzy search over sections, (2) app-level commands (toggle dark mode, navigate screens), (3) recent history. The `>` prefix convention (like VS Code) cleanly separates commands from search.
- The Cross-Reference Panel needed a clear grouping strategy. Grouping by "By Topic", "By Command", "By Flag" maps directly to the `Section` struct fields and is easy to explain.

### What warrants a second pair of eyes
- Verify that the proposed new API endpoints do not break static mode compatibility. Each new endpoint needs a corresponding JSON file generation in `pkg/help/site/render.go`.
- Verify the ASCII mockups are readable on the reMarkable device (monospace font, limited width).

### What should be done in the future
- Create Storybook stories for each new component before implementation.
- Define exact CSS custom property names for the dark mode theme.
- Decide on a client-side search library for the fallback path (e.g., Fuse.js, Minisearch) if FTS is not available.

### Code review instructions
- Read the design document from top to bottom. It is self-contained.
- Verify every file reference points to an existing file.
- Check that the screen designs match the existing component architecture.

### Technical details
---

## Step 4: Document Current React Component Hierarchy for Designers

This step added a comprehensive component inventory appendix to the design document, providing web designers with a complete reference of the existing React component tree, props, state, CSS parts, and file structure.

### Prompt Context

**User prompt (verbatim):** "Make a list of the current react component / design component hierarchy and their props and slices so that I can provide to the web designer a screenshot and the layout of components so that they have a proper starting point. Update the server docmgr ticket document and upload to remarkable"

**Assistant interpretation:** The user wants a designer-facing reference document that maps out every existing React component with its props, state, CSS data-part attributes, and layout relationships, so a web designer can understand the current UI before proposing changes.

**Inferred user intent:** Produce a single-source-of-truth component specification that bridges engineering implementation and design work.

### What I did
- Read every React component file in `web/src/components/`:
  - `AppLayout`, `TitleBar`, `MenuBar`, `SearchBar`, `TypeFilter`, `SectionList`, `SectionCard`, `SectionView`, `SectionHeader`, `MarkdownContent`, `EmptyState`, `StatusBar`, `Badge`
- Read every `parts.ts` file to extract `data-part` attribute names.
- Read every `styles/*.css` file to document layout, color, and typography rules.
- Read `App.tsx` to map the full component tree and state flow.
- Read `store.ts`, `services/api.ts`, and `types/index.ts` to document state management and data shapes.
- Read `styles/global.css` to document CSS custom properties and the System 7 aesthetic.
- Wrote "Appendix A: Current React Component Hierarchy (Designer Reference)" with:
  - Visual layout ASCII diagram
  - Hierarchical component tree
  - Per-component inventory: file path, props table, state/hooks table, CSS data-part attributes, CSS file reference
  - State management section (Redux slices, RTK Query endpoints)
  - TypeScript data types table
  - Global CSS variables table
  - Complete file tree

### Why
- Designers need to understand the existing component surface before proposing changes.
- The `data-part` attribute system (BEM-like without classes) is unconventional; documenting it prevents designers from proposing class-based approaches that don't fit.
- The System 7 aesthetic is specific (Chicago font, dot-pattern background, thick borders, hard shadows); designers need to know these constraints.

### What worked
- Reading all components revealed that `MenuBar` exists but is **not rendered** in `App.tsx` — an important discovery for the design.
- The component architecture is very consistent: every component has `Component.tsx`, `parts.ts`, and `styles/component.css`.
- The `Badge` component uses CSS custom properties (`--badge-color`, `--badge-weight`) set via inline `style` — a pattern designers should know about for theming.

### What didn't work
- N/A

### What I learned
- The current web app has **13 components** across **8 component directories**.
- **Zero custom Redux slices** — all state is either local `useState` in `App.tsx` or RTK Query cache.
- The `EmptyState` has a configurable `label` prop but `App.tsx` uses the default.
- `SectionCard` uses `stripMarkdown()` from `utils/text` — markdown is stripped from titles/shorts in the list view.
- The `TypeFilter` uses `[aria-pressed='true']` for active state — accessible but designer-relevant.

### What was tricky to build
- Organizing the information for a designer audience (not an engineer audience). The design doc's existing sections are implementation-focused; this appendix needed to be visual and scannable. The solution was: ASCII layout → tree → per-component tables → state/types/CSS reference.

### What warrants a second pair of eyes
- Verify that all `data-part` selectors and CSS rules are accurately transcribed.
- Verify that the component tree matches the actual JSX nesting in `App.tsx`.

### What should be done in the future
- Add Storybook stories for each component so designers can interact with them in isolation.
- Create a Figma/component library that mirrors the `parts.ts` naming convention.

### Code review instructions
- Verify the component tree in the appendix matches `App.tsx` line-by-line.
- Verify all `data-part` values match their `parts.ts` files.

### Technical details
- Components documented: 13
- CSS files read: 10
- `parts.ts` files read: 8
- Props tables: 13
- Data-part attributes listed: 35+
- Global CSS variables: 12
- RTK Query endpoints: 3

---

## Session Summary

| Deliverable | Path | Status |
|-------------|------|--------|
| Ticket | `GLAZE-HELP-WEB` | ✅ Created |
| Design Document | `design-doc/01-design-...` | ✅ Written + Component Appendix added |
| Investigation Diary | `reference/01-investigation-diary.md` | ✅ Written |
| File Relations | 9 source files related | ✅ Complete |
| Changelog | 4 entries | ✅ Updated |
| Validation | `docmgr doctor` | ✅ Passed |
| reMarkable Upload | `/ai/2026/04/28/GLAZE-HELP-WEB` | ✅ Verified |

### Technical details
- Design document path: `glazed/ttmp/2026/04/28/GLAZE-HELP-WEB--.../design-doc/01-design-glazed-help-web-server-enhancement-full-featured-browser-with-screens-and-user-stories.md`
- Document length: ~75 KB, ~1600 lines (after component appendix)
- Sections: 14 major sections + 1 appendix
- ASCII mockups: 7 screens
- User stories: 11
- New API endpoints proposed: 3
- Implementation phases: 6
- Components documented: 13
- File references: 30+ absolute paths
