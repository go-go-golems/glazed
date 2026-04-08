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

- [ ] Create `web/package.json` with dependencies (React 18, RTK, react-markdown, remark-gfm, Vite, TypeScript)
- [ ] Create `web/vite.config.ts` with API proxy to Go server on `:8088`
- [ ] Create `web/tsconfig.json` and `web/index.html`
- [ ] Create `web/src/types/index.ts` with TypeScript interfaces matching API response shapes
- [ ] Create `web/src/services/api.ts` with RTK Query `createApi` (4 endpoints: health, listSections, getSection, searchSections)
- [ ] Create `web/src/store.ts` with Redux store configuration
- [ ] Create `web/src/main.tsx` entry point with Provider and BrowserRouter
- [ ] Create `web/src/App.tsx` placeholder that calls `useListSectionsQuery()` and displays section count

### Phase 3: Component Decomposition (Port from JSX Prototype)

- [ ] Extract `<MenuBar />` into `web/src/components/MenuBar/` (menu bar with File/Edit/View/Help items)
- [ ] Extract `<TitleBar />` into `web/src/components/TitleBar/` (retro title bar with icon and centered title)
- [ ] Extract `<Badge />` into `web/src/components/Badge/` (colored tags: type, command, flag, topic variants)
- [ ] Extract `<SearchBar />` into `web/src/components/SearchBar/` (search input with icon)
- [ ] Extract `<TypeFilter />` into `web/src/components/TypeFilter/` (filter buttons: All, Topic, Example, App, Tutorial)
- [ ] Extract `<SectionCard />` into `web/src/components/SectionList/SectionCard.tsx`
- [ ] Extract `<SectionList />` into `web/src/components/SectionList/` (scrollable list with alternating backgrounds)
- [ ] Extract `<SectionHeader />` into `web/src/components/SectionView/SectionHeader.tsx`
- [ ] Extract `<MarkdownContent />` into `web/src/components/Markdown/` (react-markdown with GFM: code blocks, tables, headings)
- [ ] Extract `<SectionView />` into `web/src/components/SectionView/` (composes header + content)
- [ ] Extract `<EmptyState />` into `web/src/components/EmptyState/` (book icon placeholder)
- [ ] Extract `<StatusBar />` into `web/src/components/StatusBar/` (section count + version)
- [ ] Wire `<App.tsx>`: compose all components, connect RTK Query hooks, add sidebar filtering and section selection

### Phase 4: Theming System

- [ ] Create `web/src/styles/global.css` with all CSS variables at `:root` (colors, fonts, spacing, borders, shadows, layout)
- [ ] Create `web/src/styles/theme-default.css` with "classic Mac" retro theme defaults
- [ ] Convert all component `.css` files to use `data-part` selectors and CSS variables (no hardcoded values)
- [ ] Create `parts.ts` for each component with stable `data-part` name constants
- [ ] Implement `unstyled` prop on `<App />` that skips importing base CSS
- [ ] Verify theme override works in browser DevTools (override `--color-bg` and observe change)

### Phase 5: Storybook Stories

- [ ] Install Storybook: `pnpm add -D @storybook/react-vite @storybook/addon-essentials`
- [ ] Configure `.storybook/main.ts` and `.storybook/preview.ts`
- [ ] Add stories for `<Badge />` (5 variants: topic, GeneralTopic, Example, command, flag)
- [ ] Add stories for `<TitleBar />` (default)
- [ ] Add stories for `<SearchBar />` (empty, with text)
- [ ] Add stories for `<TypeFilter />` (each filter active)
- [ ] Add stories for `<SectionCard />` (active, inactive, top indicator)
- [ ] Add stories for `<SectionList />` (empty, with items, filtered)
- [ ] Add stories for `<SectionView />` (with sample section data)
- [ ] Add stories for `<MarkdownContent />` (headings, code blocks, tables, lists, inline formatting)
- [ ] Add stories for `<EmptyState />` (default)
- [ ] Add stories for `<MenuBar />` (default)
- [ ] Add stories for `<StatusBar />` (default)
- [ ] Run `pnpm storybook` and verify all stories render correctly

### Phase 6: Dagger Build Pipeline

- [ ] Create `cmd/build-web/main.go` with Dagger Go SDK (node:22 container, corepack, pnpm install, pnpm build, export dist/)
- [ ] Create `cmd/help-browser/gen.go` with `//go:generate go run ../build-web`
- [ ] Add `//go:embed dist` to `cmd/help-browser/main.go`
- [ ] Run `go generate ./cmd/help-browser` and verify `cmd/help-browser/dist/` contains `index.html` + assets/
- [ ] Build binary: `go build -o glaze ./cmd/glaze`
- [ ] Verify single binary serves both SPA and API: `./glaze serve docs/`

### Phase 7: Cobra Integration

- [ ] Refactor server setup into reusable function in `pkg/help/server/` (e.g., `NewServeCommand(helpSystem)`)
- [ ] Add `serve` subcommand in `cmd/glaze/main.go`
- [ ] Verify `glaze serve --help` works
- [ ] Verify `glaze serve docs/` starts the server correctly

### Phase 8: Integration Testing

- [ ] Add Go integration tests with `httptest.NewServer` (full request/response cycles)
- [ ] Add Go test that loads a fixture `.md` file and verifies it appears in the API
- [ ] Add Go test for 404 when slug not found
- [ ] Add Go test for SPA fallback (serves `index.html` for unknown paths)
- [ ] Add Go test for CORS headers
- [ ] Optionally: add Playwright or Cypress tests for the frontend
