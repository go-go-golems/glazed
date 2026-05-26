# Tasks

## Phase 0: Baseline evidence and failing acceptance checks

- [ ] Capture current `curl` output for `/glazed/v1.2.15` showing shell/JSON rather than full visible React tree/article HTML.
- [ ] Capture current `curl` output for `/glazed/v1.2.15/sections/exposing-a-simple-sql-table` showing the same limitation.
- [ ] Add or document a failing acceptance check that expects visible article body text in initial SSR HTML.
- [ ] Confirm current production pod/images before changing anything.

## Phase 1: Store factory and shared routes

- [ ] Add a reusable `makeStore(preloadedState?)` factory in `web/src/store.ts`.
- [ ] Preserve or adapt the existing singleton `store` export for development code that still imports it.
- [ ] Create a shared `web/src/AppRoutes.tsx` route component.
- [ ] Update `web/src/main.tsx`, `web/src/entry-client.tsx`, and `web/src/entry-server.tsx` to use `AppRoutes`.
- [ ] Run `cd web && pnpm test`.

## Phase 2: Server-side RTK Query cache preloading

- [ ] Update `web/src/entry-server.tsx` so `renderApp(url, data)` uses the prefetched `packages`, `sections`, and `section` data.
- [ ] Insert cache entries with `helpApi.util.upsertQueryData` for `listPackages`, `listSections`, and `getSection`.
- [ ] Ensure cache keys exactly match the arguments used by `App.tsx` hooks.
- [ ] Return both `html` and `preloadedState: store.getState()` from `renderApp`.
- [ ] Add unit tests for server rendering package and section pages with preloaded data.

## Phase 3: Browser hydration from preloaded state

- [ ] Update `web/src/entry-client.tsx` to read `window.__PRELOADED_STATE__` before deleting it.
- [ ] Initialize the browser store with `makeStore(preloadedState)`.
- [ ] Verify hydration does not immediately fall back to `Loading…` for server-provided data.
- [ ] Check browser console for hydration mismatch warnings.

## Phase 4: Safe state serialization in the Node sidecar

- [ ] Update `web/server.mjs` to serialize `renderApp`'s returned Redux state, not a parallel hand-built state object.
- [ ] Add a dedicated safe inline-script JSON serializer.
- [ ] Preserve existing canonical, alternate markdown, JSON-LD, Link header, and hidden/noscript metadata behavior.
- [ ] Improve package/version page metadata if practical.

## Phase 5: Local validation

- [ ] Run `go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl`.
- [ ] Run `cd web && pnpm test && pnpm build && pnpm build:ssr`.
- [ ] Build the SSR image locally with `docker build --target ssr`.
- [ ] Run the SSR container locally against `https://docs.yolo.scapegoat.dev/api`.
- [ ] Verify local curl output contains visible package index and section article body text before JavaScript.

## Phase 6: Production rollout

- [ ] Push Glazed changes and wait for runtime + SSR image builds.
- [ ] Update `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml` to the new image tags.
- [ ] Push the k3s GitOps commit.
- [ ] Force Argo CD refresh if it has not picked up the new revision.
- [ ] Watch `kubectl -n docs-yolo rollout status deploy/docs-yolo --timeout=300s`.

## Phase 7: Production acceptance

- [ ] `curl https://docs.yolo.scapegoat.dev/glazed/v1.2.15` contains visible package index text and known section titles.
- [ ] `curl https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table` contains visible article title and body text.
- [ ] Production HTML contains `window.__PRELOADED_STATE__` with Redux/RTK Query state.
- [ ] Root and nested static asset routes still return JavaScript/CSS MIME types.
- [ ] Section `.md` mirrors still return Markdown.
- [ ] Playwright direct-navigation tests hydrate without module MIME errors or hydration mismatch warnings.

## Phase 8: Documentation and delivery

- [x] Create this continuation ticket.
- [x] Write the analysis and implementation guide.
- [x] Add task plan.
- [ ] Update the implementation diary after implementation begins.
- [ ] Record final production validation evidence.
- [ ] Re-upload the completed guide bundle to reMarkable after implementation.
