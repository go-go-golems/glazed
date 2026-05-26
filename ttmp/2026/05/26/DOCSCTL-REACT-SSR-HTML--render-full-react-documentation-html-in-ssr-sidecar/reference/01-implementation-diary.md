---
Title: Implementation Diary
Ticket: DOCSCTL-REACT-SSR-HTML
Status: active
Topics:
    - docs-yolo
    - ssr
    - frontend
    - react
    - cicd
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml
      Note: Production rollout to sha-981a6db runtime and SSR images
    - Path: ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/sources/01-baseline-ssr-html-evidence.md
      Note: Baseline curl evidence before data-backed SSR implementation
    - Path: ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/sources/02-production-ssr-html-evidence.md
      Note: Final production curl
    - Path: web/server.mjs
      Note: Current sidecar fetch/serialization logic to connect to Redux state
    - Path: web/src/AppRoutes.tsx
      Note: Shared route table for dev
    - Path: web/src/entry-client.tsx
      Note: Current implementation deletes preloaded state instead of hydrating from it
    - Path: web/src/entry-server.test.tsx
      Note: Regression tests proving React SSR renders package index and article body from prefetched RTK Query data
    - Path: web/src/entry-server.tsx
      Note: Current implementation ignores prefetched data and defines the next implementation seam
    - Path: web/src/store.ts
      Note: Current singleton store setup to refactor into a store factory
ExternalSources: []
Summary: Chronological diary for continuing docs-yolo SSR from sidecar-shell rendering to full React tree/article HTML rendering.
LastUpdated: 2026-05-26T18:35:00-04:00
WhatFor: Use this diary to resume the React SSR HTML work and understand what was planned before implementation.
WhenToUse: Read before editing the React SSR entrypoints, Redux store setup, Node sidecar serialization, or production rollout manifests.
---




# Diary

## Goal

This diary records the continuation ticket for making the docs-yolo SSR sidecar render full React documentation HTML. The previous SSR work deployed the sidecar and fixed direct-link mechanics, but it did not complete data-backed React SSR. This ticket starts from that known gap.

## Step 1: Create the continuation ticket and implementation guide

I created a dedicated continuation ticket after confirming that the currently deployed SSR sidecar is not yet a full static React render. The production behavior is now mechanically correct for direct links and asset loading, but `curl` still shows mostly the Vite shell, scripts, metadata, and serialized JSON rather than the complete visible documentation tree and article body.

The resulting design guide focuses on the missing React data path. The core implementation is to create a request-scoped Redux store on the server, upsert prefetched API responses into the RTK Query cache before `renderToString()`, return the actual Redux state from `renderApp`, and initialize the browser store from that same state during hydration.

### Prompt Context

**User prompt (verbatim):** "commit and push vault. then make a new ticket to continue the SSR work to have proper SSR HTML rendered by the react sidecar. Creaet a analysis + implementation guide and then push to remarkable, add tasks to the ticket."

**Assistant interpretation:** First commit and push the Obsidian Vault article update, then create a new docmgr ticket for completing full React SSR HTML rendering, write a detailed analysis/implementation guide, add tasks, and upload the ticket bundle to reMarkable.

**Inferred user intent:** Preserve the current production/deep-dive documentation, then set up the next focused implementation package for making `curl` return meaningful rendered docs HTML rather than only a hydrated app shell.

### What I did
- Committed and pushed the Obsidian Vault article update in `/home/manuel/code/wesen/go-go-golems/go-go-parc`:
  - commit `36f5811` — `docs: update Vault OIDC docs publishing deep dive`.
- Created docmgr ticket `DOCSCTL-REACT-SSR-HTML` in the Glazed workspace.
- Created design document:
  - `ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/design-doc/01-react-ssr-html-rendering-implementation-guide.md`.
- Created this implementation diary:
  - `ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/reference/01-implementation-diary.md`.
- Replaced the generated task stub with a phased task plan in:
  - `ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/tasks.md`.
- Inspected the current React SSR/client/store/API files:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-server.tsx`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-client.tsx`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/store.ts`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/services/api.ts`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/App.tsx`

### Why
- The previous SSR deployment fixed the infrastructure path but not the core React data rendering path.
- The current `entry-server.tsx` ignores its prefetched `_data` argument, so `renderToString()` cannot produce full tree/article HTML.
- The current `entry-client.tsx` deletes `window.__PRELOADED_STATE__`, so the browser cannot hydrate from the server's data even though the sidecar serializes data.
- A focused ticket prevents this next phase from being mixed with Vault OIDC publishing and k3s sidecar rollout work.

### What worked
- The Obsidian Vault article commit and push succeeded.
- `docmgr ticket create-ticket` and `docmgr doc add` created the new ticket and documents.
- The current code inspection clearly identified the implementation seam: RTK Query cache preloading and client store hydration.
- The design guide now has concrete acceptance criteria based on `curl`, not only Playwright/browser behavior.

### What didn't work
- No implementation work has happened in this ticket yet.
- The existing production SSR still does not satisfy the desired curl-level acceptance criteria.
- The existing frontend architecture duplicates route definitions across entrypoints, which caused an earlier hydration bug and should be cleaned up before deeper SSR work.

### What I learned
- The remaining SSR issue is not primarily a k3s or Docker issue. The sidecar is deployed; the missing part is request-scoped React data hydration.
- `App.tsx` already renders the right UI from RTK Query hooks, so the cleanest path is to make those hooks see cache hits during server rendering rather than building a second HTML renderer in `server.mjs`.
- The cache keys must exactly match the hook arguments in `App.tsx`; otherwise the server can preload data that the hooks do not read.

### What was tricky to build
- The subtle part is defining "SSR done" precisely. A page can be served by an SSR sidecar and still not contain the visible article DOM in initial HTML. This ticket uses curl-level acceptance criteria to avoid overclaiming: the initial response itself must contain package index text and section article body text.
- Another tricky point is avoiding request leakage. The server must not reuse the browser singleton Redux store. Every SSR request needs a fresh store so one package/version/section cannot leak into another request.
- The implementation needs shared routes because route drift already caused a production bug. This is not just refactoring; it is an SSR invariant.

### What warrants a second pair of eyes
- The RTK Query `upsertQueryData` cache keys should be reviewed against the exact hook arguments in `App.tsx`.
- The inline state serializer in `server.mjs` should be reviewed for script-injection safety.
- The store typing in `store.ts` may need pragmatic TypeScript handling to avoid recursive `RootState` type problems.
- The acceptance tests should be reviewed to ensure they check visible server-rendered content, not only hidden/noscript metadata.

### What should be done in the future
- Implement the phases in `tasks.md`.
- Add tests that fail before implementation and pass only when initial HTML contains visible tree/article content.
- Roll out the new SSR image through k3s GitOps and validate production with curl and Playwright.

### Code review instructions
- Start with the design doc's `Proposed Solution` and `Implementation Plan` sections.
- Review `web/src/entry-server.tsx` first; it currently ignores `_data` and is the center of the fix.
- Then review `web/src/store.ts` and `web/src/entry-client.tsx` to ensure server state and client hydration use the same state shape.
- Validate with:
  - `cd web && pnpm test && pnpm build && pnpm build:ssr`
  - `go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl`
  - local SSR container curl checks for package and section URLs.

### Technical details
- Target ticket path:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar`
- Key acceptance examples:

```bash
curl -sS https://docs.yolo.scapegoat.dev/glazed/v1.2.15 \
  | grep 'Documentation Index'

curl -sS https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table \
  | grep 'TODO(manuel, 2022-12-10)'
```

- Expected architecture rule:

```text
server fetches data -> request-scoped store -> RTK Query cache entries -> renderToString -> serialized store -> client makeStore(preloadedState) -> hydrate
```

## Step 2: Implement request-scoped RTK Query SSR rendering

I implemented the first real React SSR pass. The server renderer now creates a fresh Redux store per request, inserts the data that `server.mjs` already fetched from the Go API into the RTK Query cache, renders the shared application route tree with that hydrated cache, and returns the actual Redux state used by the render. The browser hydration entry now reads `window.__PRELOADED_STATE__` and creates its store from that state instead of deleting it before use.

This changes the SSR sidecar from "HTML shell plus parallel JSON" to "React render backed by the same RTK Query cache that the client hydrates." Local curl checks against the built sidecar now show the package index and known article body text in the initial HTML before browser JavaScript executes.

### Prompt Context

**User prompt (verbatim):** "Ok, add detailed tasks, then build it task by task, committing at appropriate intervals, keeping a detailed diary"

**Assistant interpretation:** Expand the ticket tasks into an actionable implementation checklist, then implement the SSR work incrementally with validation, commits, and detailed diary updates.

**Inferred user intent:** Continue from planning into execution while preserving reviewable checkpoints and enough diary detail to understand failures and decisions later.

### What I did
- Captured baseline production SSR evidence in `sources/01-baseline-ssr-html-evidence.md`.
- Expanded `tasks.md` with detailed substeps and marked completed local phases.
- Added `web/src/AppRoutes.tsx` so dev, SSR, and hydration entries share the same route table.
- Refactored `web/src/store.ts` to expose `makeStore(preloadedState?)` while preserving the browser/dev singleton `store` export.
- Updated `web/src/main.tsx` to use `AppRoutes`.
- Updated `web/src/entry-client.tsx` to:
  - read `window.__PRELOADED_STATE__`;
  - delete it only after reading;
  - create the browser store with `makeStore(preloadedState)`;
  - hydrate shared `AppRoutes`.
- Reworked `web/src/entry-server.tsx` to:
  - parse docs routes consistently;
  - create a request-scoped store;
  - await `helpApi.util.upsertQueryData` cache inserts for `listPackages`, `listSections`, and `getSection`;
  - render `AppRoutes` under `StaticRouter`;
  - return both `html` and `preloadedState`.
- Updated `web/server.mjs` to:
  - await `renderApp`;
  - serialize the actual returned Redux state;
  - use a dedicated inline-script JSON serializer that escapes `<`, `>`, `&`, U+2028, and U+2029.
- Added `web/src/entry-server.test.tsx` covering route parsing, package index SSR, and section article SSR.

### Why
- `App.tsx` already renders the correct tree/article UI from RTK Query hooks. The missing piece was making those hooks see cache hits during server rendering.
- A request-scoped store avoids leaking one request's package/version/section data into another SSR request.
- Shared `AppRoutes` prevents route drift between dev, server render, and client hydration. Route drift already caused a previous production bug.
- Serializing the actual Redux state eliminates the parallel-state mismatch where `server.mjs` had one JSON object while React rendered from an empty store.

### What worked
- `cd web && pnpm test` passed with the new SSR tests:
  - 5 test files passed;
  - 19 tests passed.
- `cd web && pnpm build && pnpm build:ssr` passed after a TypeScript store typing fix.
- `go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl` passed.
- Local Node sidecar smoke test against the live API passed:
  - package page contained `Documentation Index`;
  - package page contained `Exposing a simple SQL table using glaze`;
  - section page contained `Exposing a simple SQL table using glaze`;
  - section page contained `TODO(manuel, 2022-12-10)`;
  - both pages contained `window.__PRELOADED_STATE__`.
- Local Docker SSR image build passed:
  - `docker build --target ssr -t glazed-ssr:react-html .`
- Local Docker sidecar smoke test against the live API passed with the same section title/body/preloaded-state checks.

### What didn't work
- The first attempt to preload RTK Query with `store.dispatch(helpApi.util.upsertQueryData(...))` and immediately call `renderToString()` still rendered `Loading…`. The SSR tests failed with HTML showing `No documentation matches.` and `Loading…`.
- The cause was that `upsertQueryData` dispatches asynchronous RTK Query work. The store did not contain fulfilled cache entries at render time until those dispatch results were awaited.
- `pnpm build` initially failed in `web/src/store.ts` because the generic `preloadedState` type confused the `configureStore` overload and middleware tuple inference. The error was a long TS2322/TS2719 overload mismatch around `reducer`, `middleware`, and `preloadedState`.
- The fix was to keep strong exported store/result types but use a pragmatic `config as any` at the factory boundary, with `preloadedState` conditionally included only when present.
- React prints `useLayoutEffect does nothing on the server` warnings during SSR tests. These warnings come from React/React-Redux/server rendering behavior and did not fail the tests. They warrant review if production hydration warnings appear, but they did not block the local SSR acceptance criteria.

### What I learned
- RTK Query cache upserts must be awaited for SSR. Treating them as synchronous produced the exact failure mode this ticket is meant to remove: hooks rendered loading state despite data being available outside React.
- The right state to serialize is `store.getState()`, not a hand-built `{ packages, sections, section }` object. The client needs RTK Query's cache state shape, not the domain data shape alone.
- The existing Node sidecar fetch logic can remain simple. It fetches the data; `entry-server.tsx` is responsible for making that data visible to React through the cache.
- TypeScript's `configureStore` types can become a distraction when making a generic store factory. A narrow escape hatch at the factory boundary is acceptable if tests and exported store types remain useful.

### What was tricky to build
- The hardest part was realizing the cache insert operation was asynchronous. The symptom was a confusing test failure: the preloaded state JSON contained endpoint names after render, but the rendered HTML still showed loading state. Awaiting the dispatch results fixed the ordering.
- It was also tricky to keep route/version normalization exact. `App.tsx` treats `_` as an empty API version; SSR cache keys must do the same or the hooks miss the cache.
- Another sharp edge was preserving the safe serialization and metadata behavior in `server.mjs` while changing the state source. The code now serializes the Redux state but keeps canonical links, alternate markdown links, JSON-LD, and hidden/noscript helpers intact.

### What warrants a second pair of eyes
- Review the `makeStore(preloadedState?)` typing compromise in `web/src/store.ts`.
- Review `parseDocsRoute()` in `entry-server.tsx` against the Go server and `App.tsx` URL semantics.
- Review the inline serializer in `web/server.mjs` for script-injection safety.
- Review whether the server-side `useLayoutEffect` warnings should be addressed with an isomorphic layout effect helper or left as known harmless warnings.

### What should be done in the future
- Commit this local implementation checkpoint.
- Push and wait for GHCR runtime/SSR image builds.
- Roll the new images into k3s and validate production with curl and Playwright.
- Record final production evidence in `sources/02-production-ssr-html-evidence.md`.

### Code review instructions
- Start with `web/src/entry-server.tsx`; verify the cache upsert keys match `App.tsx` hook arguments.
- Then review `web/src/entry-client.tsx` and `web/src/store.ts`; verify the browser uses the serialized state.
- Review `web/server.mjs`; verify it awaits `renderApp` and serializes returned Redux state safely.
- Run:
  - `cd web && pnpm test && pnpm build && pnpm build:ssr`
  - `go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl`
  - `docker build --target ssr -t glazed-ssr:react-html .`

### Technical details
- First failed SSR test symptom:

```text
expected '<div class="app-root"><div data-part=…' to contain 'Documentation Index'
Received HTML included: No documentation matches. ... Loading…
```

- TypeScript build failure summary:

```text
src/store.ts(15,5): error TS2322: Type '{ helpApi: Reducer<CombinedState<...>> }' is not assignable to type 'Reducer<...>'
src/store.ts(16,5): error TS2719: Type '(getDefaultMiddleware: GetDefaultMiddleware<...>) => Tuple<...>' is not assignable ...
```

- Local acceptance commands that passed:

```bash
cd /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web
pnpm test
pnpm build
pnpm build:ssr

cd /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed
go test ./pkg/help/server ./pkg/web ./cmd/docs-registry ./cmd/docsctl
docker build --target ssr -t glazed-ssr:react-html .
```

## Step 3: Roll out the React SSR HTML build to production

I committed the local React SSR implementation, pushed it to Glazed `main`, waited for the container workflow to publish the runtime and SSR sidecar images, then updated the k3s `docs-yolo` deployment to use the new immutable image tags. After forcing an Argo CD refresh, the deployment rolled out successfully with all three containers ready.

Production now satisfies the curl-level acceptance criteria that motivated this ticket. Package pages contain the visible documentation index and section titles in the initial HTML, and direct section pages contain the visible article title and body text before browser JavaScript executes.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementing task by task, commit at checkpoints, and record the rollout and validation details in the diary.

**Inferred user intent:** Complete the production deployment and evidence capture, not just local implementation.

**Commit (code):** `981a6db` — "docs-yolo: render React SSR from preloaded state"

**Commit (GitOps):** `2a27cdb` — "docs-yolo: deploy React SSR HTML rendering"

### What I did
- Committed the Glazed implementation checkpoint:
  - `981a6db` — `docs-yolo: render React SSR from preloaded state`.
- Pushed the commit to `go-go-golems/glazed@main`.
- Watched the container workflow:
  - `https://github.com/go-go-golems/glazed/actions/runs/26479116597`
  - result: success.
- Updated k3s docs-yolo image tags from `sha-a6d688b` to `sha-981a6db` in:
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`.
- Rendered manifests with `kubectl kustomize gitops/kustomize/docs-yolo` and confirmed the new image tags and `docs-ssr` container appeared.
- Committed and pushed the k3s GitOps change:
  - `2a27cdb` — `docs-yolo: deploy React SSR HTML rendering`.
- Forced Argo CD to refresh:
  - `kubectl -n argocd annotate application docs-yolo argocd.argoproj.io/refresh=hard --overwrite`.
- Watched rollout:
  - `kubectl -n docs-yolo rollout status deploy/docs-yolo --timeout=300s`.
- Captured production evidence in:
  - `sources/02-production-ssr-html-evidence.md`.
- Validated with curl and Playwright.

### Why
- The local implementation needed production proof because this ticket's acceptance criteria are about public initial HTML, not only local unit tests.
- The deployment uses immutable `sha-*` image tags, so k3s needed an explicit GitOps image update after GHCR images were published.
- Argo CD had previously reconciled from a cached revision during the SSR sidecar work, so forcing a hard refresh is the known safe rollout sequence.

### What worked
- Glazed container workflow succeeded:
  - runtime image: `ghcr.io/go-go-golems/glazed:sha-981a6db`;
  - SSR image: `ghcr.io/go-go-golems/glazed-ssr:sha-981a6db`.
- k3s rollout succeeded with final images:
  - `docs-browser => ghcr.io/go-go-golems/glazed:sha-981a6db`;
  - `docs-registry => ghcr.io/go-go-golems/glazed:sha-981a6db`;
  - `docs-ssr => ghcr.io/go-go-golems/glazed-ssr:sha-981a6db`.
- Production curl checks passed:
  - `/glazed/v1.2.15` contains `Documentation Index`;
  - `/glazed/v1.2.15` contains `Exposing a simple SQL table using glaze`;
  - `/glazed/v1.2.15/sections/exposing-a-simple-sql-table` contains `TODO(manuel, 2022-12-10)`;
  - production HTML contains `window.__PRELOADED_STATE__` and RTK Query state markers such as `queries`.
- Static/API regression checks passed:
  - root JS asset returned `content-type: text/javascript`;
  - nested JS asset returned `content-type: text/javascript`;
  - CSS asset returned `content-type: text/css`;
  - section markdown mirror returned `content-type: text/markdown`;
  - `/api/health` returned `{"ok":true,"sections":333}`;
  - registry `/healthz` returned `{"ok":true}`.
- Playwright direct navigation to the section page hydrated successfully with visible article content.

### What didn't work
- There were no CI/container workflow failures in this rollout.
- The browser console still reports the known unrelated Chicago font 404 from `cdn.jsdelivr.net`. There were no module MIME errors and no hydration mismatch warnings observed through Playwright.

### What I learned
- The data-backed SSR implementation survives the production sidecar/container boundary; the same checks that passed locally pass through the public Go proxy and Node sidecar.
- The initial HTML now contains visible React-rendered content, not only hidden/noscript helper text and parallel JSON.
- The existing static asset normalization remains important even after full React SSR because hydration still needs the JS and CSS assets.

### What was tricky to build
- The rollout had to wait for the container workflow because the k3s manifests use immutable image tags. Updating k3s before GHCR had the tag would risk image pull failures.
- The validation needed to distinguish content that appears in visible React markup from content that appears in metadata or preloaded JSON. The final evidence uses multiple checks: package index text, known section titles, known article body text, and RTK Query state markers.

### What warrants a second pair of eyes
- Review whether package/version page metadata should now be made more specific. The visible HTML is fixed, but `<title>` may still be generic for package landing pages.
- Review the known external font 404 and decide whether to vendor the font, change the URL, or remove the dependency.
- Review the public registry hardening tasks, which remain outside this SSR ticket.

### What should be done in the future
- Run `docmgr doctor` and re-upload the final ticket bundle to reMarkable.
- Consider adding production-like Playwright SSR acceptance tests to CI.
- Consider eliminating the `useLayoutEffect` SSR warnings observed during local tests.

### Code review instructions
- For the Glazed code commit, start at `web/src/entry-server.tsx`, then `web/src/store.ts`, `web/src/entry-client.tsx`, and `web/server.mjs`.
- For the production rollout, review `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml` commit `2a27cdb`.
- Validate production with:

```bash
curl -sS https://docs.yolo.scapegoat.dev/glazed/v1.2.15 | grep 'Documentation Index'
curl -sS https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table | grep 'TODO(manuel, 2022-12-10)'
```

### Technical details
- Successful container workflow:
  - `https://github.com/go-go-golems/glazed/actions/runs/26479116597`
- Production evidence file:
  - `ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/sources/02-production-ssr-html-evidence.md`

## Step 4: Final ticket validation and reMarkable delivery

I completed the ticket bookkeeping after production validation. The evidence source files needed docmgr frontmatter, so I added frontmatter to both the baseline and production evidence documents, reran `docmgr doctor`, and uploaded the completed bundle to reMarkable with `--force`.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Keep documentation and delivery artifacts current as implementation finishes.

**Inferred user intent:** Leave the ticket in a complete, reviewable state with the final evidence available on reMarkable.

### What I did
- Added YAML frontmatter to:
  - `sources/01-baseline-ssr-html-evidence.md`
  - `sources/02-production-ssr-html-evidence.md`
- Ran `docmgr doctor --ticket DOCSCTL-REACT-SSR-HTML --stale-after 30` successfully.
- Re-uploaded the completed bundle to reMarkable:
  - `DOCSCTL React SSR HTML Guide.pdf`
  - remote path: `/ai/2026/05/26/DOCSCTL-REACT-SSR-HTML`
- Marked final documentation/delivery tasks complete.

### Why
- `docmgr doctor` treats Markdown files under the ticket as managed docs and requires valid frontmatter.
- The reMarkable bundle should include the final implementation diary, task status, changelog, and validation evidence, not only the initial design guide.

### What worked
- `docmgr doctor` passed after adding frontmatter.
- `remarquee upload bundle ... --force --non-interactive` succeeded with:
  - `OK: uploaded DOCSCTL React SSR HTML Guide.pdf -> /ai/2026/05/26/DOCSCTL-REACT-SSR-HTML`

### What didn't work
- The first `docmgr doctor` run failed because the two evidence files were plain Markdown without frontmatter:
  - `frontmatter delimiters '---' not found`
- Adding standard ticket frontmatter fixed the issue.

### What I learned
- Even source/evidence Markdown files in ticket workspaces should include frontmatter if they are meant to be committed as managed ticket documents.

### What was tricky to build
- N/A; this was final documentation hygiene rather than code implementation.

### What warrants a second pair of eyes
- Review whether the evidence files should remain `DocType: reference` or use a dedicated source/evidence doc type if the vocabulary supports it in the future.

### What should be done in the future
- N/A for this ticket.

### Code review instructions
- Confirm the final ticket docs pass `docmgr doctor`.
- Confirm the production evidence source includes the final image tags and curl checks.

### Technical details

```bash
docmgr doctor --ticket DOCSCTL-REACT-SSR-HTML --stale-after 30
remarquee upload bundle ... --name "DOCSCTL React SSR HTML Guide" --remote-dir "/ai/2026/05/26/DOCSCTL-REACT-SSR-HTML" --toc-depth 2 --force --non-interactive
```
