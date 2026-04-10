# Changelog

## 2026-04-09

- Initial workspace created


## 2026-04-09 - Initial research and design package

Created ticket GL-012-STATIC-RENDER for static help-site rendering, replaced placeholder docs with a detailed intern-oriented design guide and diary, added a phased task breakdown, and grounded the proposal in the existing serve/help/web architecture plus the prior help-browser review.

### Related Files

- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/cmd/glaze/main.go — Existing serve command registration informed the proposed static export command wiring
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/help/server/serve.go — Existing loader and serve semantics shaped the recommended shared-loading refactor
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/web/static.go — Existing shared SPA boundary shaped the recommendation to reuse the frontend for static export
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/ttmp/2026/04/09/GL-012-STATIC-RENDER--add-static-website-rendering-to-cmd-glaze/design-doc/01-static-help-website-rendering-architecture-and-implementation-guide.md — Primary architecture and implementation guide created in this step
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/ttmp/2026/04/09/GL-012-STATIC-RENDER--add-static-website-rendering-to-cmd-glaze/reference/01-diary.md — Chronological diary entry for ticket setup and architectural investigation
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/ttmp/2026/04/09/GL-012-STATIC-RENDER--add-static-website-rendering-to-cmd-glaze/tasks.md — Phased task list for implementation planning


## 2026-04-09 - Static export command shipped

Implemented the first working `glaze render-site` slice, including shared path-loading helpers, static snapshot emission, SPA static-mode transport, hash-route selection, generated frontend assets, and end-to-end export validation. This work landed in commit `8181daa`.

### Related Files

- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/cmd/glaze/main.go — Added the new `render-site` command wiring
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/help/loader/paths.go — New shared local markdown path loader extracted from the serve path
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/help/server/serve.go — Updated to reuse the shared loader
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/help/site/render.go — New static export implementation
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/help/site/render_test.go — Static export tests
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/App.tsx — Hash-route based section selection
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/services/api.ts — Static-mode JSON loading support
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/web/dist/index.html — Updated embedded SPA shell for exported sites


## 2026-04-09 - Frontend coverage for static routing and transport

Added the first frontend test harness for the help SPA, covering static runtime-path selection and hash-route based section selection. This work landed in commit `1828776`.

### Related Files

- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/package.json — Added the `vitest` test script and test dependencies
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/pnpm-lock.yaml — Locked the new frontend test dependencies
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/vite.config.ts — Added the shared `jsdom` test environment configuration
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/services/api.ts — Extracted runtime-config helpers used by both production code and tests
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/services/api.test.ts — Added static/server transport selection tests
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/App.test.tsx — Added hash-route selection tests
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/web/src/test/setup.ts — Added shared test cleanup/reset behavior
- /home/manuel/workspaces/2026-04-09/glaze-render-static/glazed/pkg/web/dist/index.html — Rebuilt embedded frontend assets after the runtime helper changes
