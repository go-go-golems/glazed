# Tasks

## Phase 0: Baseline evidence and failing acceptance checks

- [x] Capture current `curl` output for `/glazed/v1.2.15` showing shell/JSON rather than full visible React tree/article HTML.
  - Evidence file: `sources/01-baseline-ssr-html-evidence.md`.
  - Record headers, first 500 bytes, `__PRELOADED_STATE__`, `Documentation Index`, known section title, and known body text counts.
- [x] Capture current `curl` output for `/glazed/v1.2.15/sections/exposing-a-simple-sql-table` showing the same limitation.
  - Note: section body text is partially present through sidecar-injected metadata/noscript content, but the visible React tree/article is not rendered from RTK Query cache.
- [x] Add or document a failing acceptance check that expects visible article body text in initial SSR HTML.
  - Target checks are recorded in the design doc and baseline evidence.
- [x] Confirm current production pod/images before changing anything.
  - Current baseline images before this implementation: `ghcr.io/go-go-golems/glazed:sha-a6d688b` and `ghcr.io/go-go-golems/glazed-ssr:sha-a6d688b`.

## Phase 1: Store factory and shared routes

- [x] Add a reusable `makeStore(preloadedState?)` factory in `web/src/store.ts`.
  - Keep request-scoped SSR stores separate from browser/dev singleton store.
  - Accept pragmatic typing at the factory boundary if Redux Toolkit's generics become recursive.
- [x] Preserve or adapt the existing singleton `store` export for development code that still imports it.
  - `main.tsx` can continue to use the singleton.
  - `entry-client.tsx` should create its own store from preloaded state.
- [x] Create a shared `web/src/AppRoutes.tsx` route component.
  - Include `/:package/:version/sections/:slug`.
  - Include `/:package/:version`.
  - Include `*` fallback.
- [x] Update `web/src/main.tsx`, `web/src/entry-client.tsx`, and `web/src/entry-server.tsx` to use `AppRoutes`.
- [x] Run `cd web && pnpm test`.

## Phase 2: Server-side RTK Query cache preloading

- [x] Update `web/src/entry-server.tsx` so `renderApp(url, data)` uses the prefetched `packages`, `sections`, and `section` data.
- [x] Insert cache entries with `helpApi.util.upsertQueryData` for `listPackages`, `listSections`, and `getSection`.
  - Await the upsert dispatches before calling `renderToString()`.
- [x] Ensure cache keys exactly match the arguments used by `App.tsx` hooks.
  - Add `parseDocsRoute()` tests for package/version URLs, section URLs, query/hash stripping, and `_` version normalization.
- [x] Return both `html` and `preloadedState: store.getState()` from `renderApp`.
- [x] Add unit tests for server rendering package and section pages with preloaded data.
  - `web/src/entry-server.test.tsx` checks `Documentation Index`, known section title, article body text, and RTK Query cache state.

## Phase 3: Browser hydration from preloaded state

- [x] Update `web/src/entry-client.tsx` to read `window.__PRELOADED_STATE__` before deleting it.
- [x] Initialize the browser store with `makeStore(preloadedState)`.
- [x] Verify hydration does not immediately fall back to `Loading…` for server-provided data.
  - Covered by SSR render tests and local sidecar curl checks before production rollout.
- [x] Check browser console for hydration mismatch warnings in production after image rollout.

## Phase 4: Safe state serialization in the Node sidecar

- [x] Update `web/server.mjs` to serialize `renderApp`'s returned Redux state, not a parallel hand-built state object.
- [x] Add a dedicated safe inline-script JSON serializer.
  - Escape `<`, `>`, `&`, U+2028, and U+2029.
- [x] Preserve existing canonical, alternate markdown, JSON-LD, Link header, and hidden/noscript metadata behavior.
- [x] Improve package/version page metadata if practical.
  - Optional follow-up if not needed for full React HTML acceptance.

## Phase 5: Local validation

- [x] Run `go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl`.
- [x] Run `cd web && pnpm test && pnpm build && pnpm build:ssr`.
- [x] Build the SSR image locally with `docker build --target ssr`.
- [x] Run the SSR container locally against `https://docs.yolo.scapegoat.dev/api`.
- [x] Verify local curl output contains visible package index and section article body text before JavaScript.
  - Node sidecar local port `8099` checks passed.
  - Docker sidecar local port `8098` checks passed.

## Phase 6: Commit and CI image build

- [x] Commit the React SSR HTML implementation in Glazed.
- [x] Push Glazed changes and wait for runtime + SSR image builds.
- [x] Record successful container workflow run ID.
- [x] Record any CI failures and fixes in the diary.
  - No container workflow failures occurred for this phase; prior local TS/test failures are recorded in the diary.

## Phase 7: Production rollout

- [x] Update `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml` to the new image tags.
- [x] Render manifests with `kubectl kustomize gitops/kustomize/docs-yolo`.
- [x] Push the k3s GitOps commit.
- [x] Force Argo CD refresh if it has not picked up the new revision.
- [x] Watch `kubectl -n docs-yolo rollout status deploy/docs-yolo --timeout=300s`.
- [x] Record final pod images and readiness.

## Phase 8: Production acceptance

- [x] `curl https://docs.yolo.scapegoat.dev/glazed/v1.2.15` contains visible package index text and known section titles.
- [x] `curl https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table` contains visible article title and body text.
- [x] Production HTML contains `window.__PRELOADED_STATE__` with Redux/RTK Query state.
- [x] Root and nested static asset routes still return JavaScript/CSS MIME types.
- [x] Section `.md` mirrors still return Markdown.
- [x] Playwright direct-navigation tests hydrate without module MIME errors or hydration mismatch warnings.
- [x] Record final production evidence in `sources/02-production-ssr-html-evidence.md`.

## Phase 9: Documentation and delivery

- [x] Create this continuation ticket.
- [x] Write the analysis and implementation guide.
- [x] Add task plan.
- [x] Update the implementation diary after implementation begins.
- [x] Record final production validation evidence.
- [x] Update the diary after each implementation/rollout commit.
- [x] Run `docmgr doctor --ticket DOCSCTL-REACT-SSR-HTML --stale-after 30`.
- [x] Re-upload the completed guide bundle to reMarkable after implementation.
