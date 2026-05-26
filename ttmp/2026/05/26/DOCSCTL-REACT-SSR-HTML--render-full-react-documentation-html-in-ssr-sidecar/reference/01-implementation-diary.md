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
    - Path: web/server.mjs
      Note: Current sidecar fetch/serialization logic to connect to Redux state
    - Path: web/src/entry-client.tsx
      Note: Current implementation deletes preloaded state instead of hydrating from it
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
