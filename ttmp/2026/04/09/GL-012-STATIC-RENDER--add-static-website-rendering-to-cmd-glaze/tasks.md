# Tasks

## Implementation Phases

### Phase 1: Command Surface and Shared Loading

- [x] Add a dedicated static-site command to `cmd/glaze/main.go` and keep `serve` unchanged
- [x] Decide and document the final CLI name and flags (`render-site` is the proposed default)
- [x] Reuse the same path-loading semantics as `glaze serve`: embedded docs by default, explicit paths replace preloaded docs
- [x] Extract shared markdown file and directory loading from `pkg/help/server/serve.go` into reusable helpers instead of duplicating walker logic
- [x] Add a deterministic settings struct for output path, site title, base path, and overwrite behavior

### Phase 2: Static Snapshot Builder

- [x] Introduce a package responsible for building a static-site snapshot from the canonical `HelpSystem` / `Store`
- [x] Produce a stable ordered section list derived from `store.OrderByOrder()` and fallback ordering rules
- [x] Emit per-section JSON payloads from `model.Section` / `server.DetailFromModel(...)`
- [x] Emit list and index payloads for topics, commands, flags, top-level sections, and default sections
- [x] Emit a site manifest with build metadata and relative asset/data paths

### Phase 3: Frontend Static Mode

- [x] Add a small runtime config contract so the frontend can run in `server` mode or `static` mode
- [x] Keep the existing SPA as the primary viewer instead of introducing a second HTML renderer
- [x] Teach the frontend to load pre-generated JSON files instead of calling live `/api/...` endpoints in static mode
- [x] Make section selection URL-addressable so exported sites support bookmarkable links
- [x] Verify that `HashRouter`-based navigation remains compatible with simple static hosting

### Phase 4: Site Export Writer

- [x] Copy the built frontend assets into the output directory
- [x] Write a runtime config file and static data tree into the output directory
- [x] Generate a landing `index.html` and any host-compatibility helper files needed for static hosting
- [x] Support repeated exports to the same directory with explicit overwrite semantics
- [x] Document the output directory structure and hosting assumptions

### Phase 5: Validation and Documentation

- [x] Add unit tests for shared path loading, snapshot generation, and deterministic output ordering
- [x] Add integration tests that render a small fixture site and assert on the emitted files
- [x] Add frontend tests for static-mode data loading and route selection
- [ ] Add a user-facing help page describing how to build and host a static exported site
- [ ] Add a developer playbook for validating the exported site locally

### Phase 6: Optional Follow-Ons

- [ ] Decide whether `glaze serve` should later gain an `--export` shortcut that delegates to the static renderer
- [ ] Decide whether to emit SEO-oriented per-section prerendered HTML in addition to the SPA shell
- [ ] Decide whether to expose the static snapshot package as a reusable library for other Go applications
