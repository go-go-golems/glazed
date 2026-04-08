# Tasks

## Implementation Phases

### Phase 1: Go HTTP Server Scaffold

- [x] Create `pkg/help/server/types.go` with `SectionSummary`, `SectionDetail`, `ListResponse`, `HealthResponse` structs
- [x] Create `pkg/help/server/handlers.go` with `handleHealth`, `handleListSections`, `handleGetSection`, `handleSearch` handlers
- [x] Create `pkg/help/server/middleware.go` with CORS middleware
- [x] Create `pkg/help/server/spa.go` with SPA fallback handler
- [x] Create `pkg/help/server/server.go` with `NewServer()`, `ServeMux` registration, and `ListenAndServe()`
- [x] Write `pkg/help/server/server_test.go` with `httptest.NewServer` integration tests
- [x] Create `cmd/help-browser/main.go` with file discovery from CLI args and `HelpSystem` initialization

### Phase 2: React Frontend Scaffold

- [x] Create `web/package.json` with dependencies (React 18, RTK, react-markdown, remark-gfm, Vite, TypeScript)
- [x] Create `web/vite.config.ts` with API proxy to Go server on `:8088`
- [x] Create `web/tsconfig.json` and `web/index.html`
- [x] Create `web/src/types/index.ts` with TypeScript interfaces matching API response shapes
- [x] Create `web/src/services/api.ts` with RTK Query `createApi` (4 endpoints: health, listSections, getSection, searchSections)
- [x] Create `web/src/store.ts` with Redux store configuration
- [x] Create `web/src/main.tsx` entry point with Provider and BrowserRouter
- [x] Create `web/src/App.tsx` placeholder that calls `useListSectionsQuery()` and displays section count

### Phase 3: Component Decomposition (Port from JSX Prototype)

- [x] Extract `<MenuBar />` into `web/src/components/MenuBar/` (menu bar with File/Edit/View/Help items)
- [x] Extract `<TitleBar />` into `web/src/components/TitleBar/` (retro title bar with icon and centered title)
- [x] Extract `<Badge />` into `web/src/components/Badge/` (colored tags: type, command, flag, topic variants)
- [x] Extract `<SearchBar />` into `web/src/components/SearchBar/` (search input with icon)
- [x] Extract `<TypeFilter />` into `web/src/components/TypeFilter/` (filter buttons: All, Topic, Example, App, Tutorial)
- [x] Extract `<SectionCard />` into `web/src/components/SectionList/SectionCard.tsx`
- [x] Extract `<SectionList />` into `web/src/components/SectionList/` (scrollable list with alternating backgrounds)
- [x] Extract `<SectionHeader />` into `web/src/components/SectionView/SectionHeader.tsx`
- [x] Extract `<MarkdownContent />` into `web/src/components/Markdown/` (react-markdown with GFM: code blocks, tables, headings)
- [x] Extract `<SectionView />` into `web/src/components/SectionView/` (composes header + content)
- [x] Extract `<EmptyState />` into `web/src/components/EmptyState/` (book icon placeholder)
- [x] Extract `<StatusBar />` into `web/src/components/StatusBar/` (section count + version)
- [x] Wire `<App.tsx>`: compose all components, connect RTK Query hooks, add sidebar filtering and section selection

### Phase 4: Theming System

- [x] Create `web/src/styles/global.css` with all CSS variables at `:root` (colors, fonts, spacing, borders, shadows, layout)
- [x] Create `web/src/styles/theme-default.css` with "classic Mac" retro theme defaults
- [x] Convert all component `.css` files to use `data-part` selectors and CSS variables (no hardcoded values)
- [x] Create `parts.ts` for each component with stable `data-part` name constants
- [x] Implement `unstyled` prop on `<App />` that skips importing base CSS
- [x] Verify theme override works in browser DevTools (override `--color-bg` and observe change)

### Phase 5: Storybook Stories

- [x] Install Storybook: `pnpm add -D @storybook/react-vite @storybook/addon-essentials`
- [x] Configure `.storybook/main.ts` and `.storybook/preview.ts`
- [x] Add stories for `<Badge />` (5 variants: topic, GeneralTopic, Example, command, flag)
- [x] Add stories for `<TitleBar />` (default)
- [x] Add stories for `<SearchBar />` (empty, with text)
- [x] Add stories for `<TypeFilter />` (each filter active)
- [x] Add stories for `<SectionCard />` (active, inactive, top indicator)
- [x] Add stories for `<SectionList />` (empty, with items, filtered)
- [x] Add stories for `<SectionView />` (with sample section data)
- [x] Add stories for `<MarkdownContent />` (headings, code blocks, tables, lists, inline formatting)
- [x] Add stories for `<EmptyState />` (default)
- [x] Add stories for `<MenuBar />` (default)
- [x] Add stories for `<StatusBar />` (default)
- [x] Run `pnpm storybook` and verify all stories render correctly

### Phase 6: Dagger Build Pipeline

- [x] Create `cmd/build-web/main.go` with Dagger Go SDK (node:22 container, corepack, pnpm install, pnpm build, export dist/)
- [x] Create `cmd/help-browser/gen.go` with `//go:generate go run ../build-web`
- [x] Add `//go:embed dist` to `cmd/help-browser/main.go` (now in embed.go)
- [x] Run `go generate ./cmd/help-browser` and verify `cmd/help-browser/dist/` contains `index.html` + assets/
- [x] Build binary: `go build -o glaze ./cmd/glaze` (uses cmd/help-browser)
- [ ] Verify single binary serves both SPA and API: `./glaze serve docs/`

### Phase 7: Cobra Integration

- [ ] Refactor server setup into reusable function in `pkg/help/server/` (e.g., `NewServeCommand(helpSystem)`)
- [ ] Add `serve` subcommand in `cmd/glaze/main.go`
- [ ] Verify `glaze serve --help` works
- [ ] Verify `glaze serve docs/` starts the server correctly

### Phase 8: Integration Testing

- [x] Add Go integration tests with `httptest.NewServer` (full request/response cycles)
- [x] Add Go test that loads a fixture `.md` file and verifies it appears in the API
- [x] Add Go test for 404 when slug not found
- [x] Add Go test for SPA fallback (serves `index.html` for unknown paths)
- [x] Add Go test for CORS headers
- [ ] Optionally: add Playwright or Cypress tests for the frontend
