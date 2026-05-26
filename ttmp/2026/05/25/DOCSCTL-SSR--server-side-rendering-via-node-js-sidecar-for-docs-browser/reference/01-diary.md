---
Title: Diary
Ticket: DOCSCTL-SSR
Status: active
Topics:
    - docsctl
    - ssr
    - react
    - node
    - kubernetes
    - sidecar
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: .devctl.yaml
      Note: Devctl plugin wiring config
    - Path: pkg/help/server/serve.go
    - Path: plugins/glazed.py
      Note: Devctl plugin for starting docs server + SSR sidecar
    - Path: web/server.mjs
      Note: Express SSR server with window mock and dynamic import
    - Path: web/src/entry-client.tsx
    - Path: web/src/entry-server.tsx
    - Path: web/vite.config.ts
      Note: Updated ssr.noExternal list with react and use-sync-external-store
ExternalSources: []
Summary: 'SSR sidecar implementation: entry points, Express server, Go reverse proxy, devctl plugin'
LastUpdated: 2026-05-25T21:52:00-04:00
WhatFor: ""
WhenToUse: ""
---





# Diary

## Goal

Implement server-side rendering via a Node.js sidecar for the Glazed docs browser, improving SEO, a14y score, and initial page load by serving pre-rendered HTML instead of an empty SPA shell.

## Step 1: Create SSR frontend entry points and Express server

Created the SSR build pipeline for the React app: entry-server.tsx (StaticRouter + renderToString), entry-client.tsx (hydrateRoot), and server.mjs (Express SSR server). Updated main.tsx to keep createRoot for dev mode, added SSR build config to vite.config.ts, and added build:ssr/build:all/ssr scripts to package.json. Added express + @types/express as dev dependencies.

### Prompt Context

**User prompt (verbatim):** (see session summary)

**Assistant interpretation:** Implement the SSR sidecar approach (Option A): Node SSR sidecar in same k3s pod, Go server proxies page requests to it, falls back to SPA when sidecar unavailable.

**Inferred user intent:** Enable SSR for the docs browser to improve a14y score from 42/100 and provide meaningful HTML to agents/crawlers.

**Commit (code):** 7ea2f9d — "DOCSCTL-SSR: add SSR sidecar — entry points, Express server, build config"

### What I did
- Created `web/src/entry-server.tsx` — renders React app to HTML string using StaticRouter
- Created `web/src/entry-client.tsx` — hydrates server-rendered HTML using hydrateRoot
- Created `web/server.mjs` — Express SSR server with health endpoint, API pre-fetching, and HTML injection
- Updated `web/src/main.tsx` — kept createRoot for dev mode (no SSR)
- Updated `web/vite.config.ts` — added ssr.noExternal config for key packages
- Updated `web/package.json` — added build:ssr, build:all, ssr scripts

### Why
SPA returns empty `<div id="root"></div>` for all routes, making it invisible to agents/crawlers. SSR pre-renders React components to HTML so the page has meaningful content on first load.

### What worked
- Both client and SSR builds pass without errors
- All 14 tests pass
- entry-server.tsx creates a fresh Redux store per request (no shared mutable state)

### What didn't work
- (none in this step)

### What I learned
- Vite SSR build with noExternal: certain packages (react-dom, react-router-dom, @reduxjs/toolkit, react-redux) must be bundled into the SSR output, not externalized
- ESM imports are hoisted — runtime code can't run before static imports. Must use dynamic import() for modules that need globals set up first.

### What was tricky to build
- Deciding what to put in noExternal vs external: React ecosystem packages need to be bundled because they use internal React context that must be the same instance

### What warrants a second pair of eyes
- The SSR entry-server.tsx doesn't pre-populate the RTK Query cache — it renders "loading" state. This is intentional for now but limits SSR SEO benefit.

### What should be done in the future
- Pre-populate RTK Query cache in entry-server.tsx for full SSR content rendering
- Consider using renderToPipeableStream instead of renderToString for streaming SSR

### Code review instructions
- Start with `web/src/entry-server.tsx` — the renderApp() function
- Verify `web/src/entry-client.tsx` uses hydrateRoot (not createRoot)
- Check `web/vite.config.ts` ssr.noExternal list
- Run: `cd web && pnpm build:all && pnpm test`

---

## Step 2: Add --ssr-url flag and reverse proxy to Go server

Added SSRURL field to ServeSettings, --ssr-url flag, ServeOption pattern with WithSSRURL(), and newSSRProxy function that reverse-proxies page requests to the SSR sidecar with SPA fallback on error or 5xx.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Commit (code):** 9d99c44 — "DOCSCTL-SSR: add SSR sidecar — entry points, Express server, build config" (amended to include Go changes)

### What I did
- Added `SSRURL` field to `ServeSettings` struct
- Added `--ssr-url` flag to the serve command
- Created `ServeOption` type and `WithSSRURL()` option function
- Updated `NewServeHandler` signature to accept `...ServeOption`
- Implemented `newSSRProxy()` — reverse proxy that:
  - Copies request headers/body to the SSR sidecar
  - Falls back to SPA on connection errors
  - Falls back to SPA on 5xx responses from the sidecar
  - Passes through 2xx/3xx/4xx responses directly
- Updated serve command help text

### Why
The Go server needs to proxy page requests to the SSR sidecar so that external clients (browsers, agents, crawlers) get server-rendered HTML instead of an empty SPA shell.

### What worked
- Go build and tests pass
- gosec G704 false positive resolved with `#nosec` comment (SSR URL comes from CLI flag, not user input)

### What didn't work
- Initial `//nosec G704` (without `#`) wasn't recognized by gosec — must use `// #nosec G704` format

### What I learned
- gosec `#nosec` format: must use `#` prefix and include rationale after `--`
- httputil.ReverseProxy with custom Director + ModifyResponse for selective fallback

### What was tricky to build
- The fallback logic: on error or 5xx, serve the SPA fallback instead of returning an error page. This required reading the response body, checking the status, and re-reading the SPA file.

### What warrants a second pair of eyes
- The reverse proxy error handling: is the SPA fallback correct for all error modes?
- The `#nosec G704` justification — ssrURL is admin-controlled via CLI flag

### What should be done in the future
- Add SSR proxy tests (unit tests for newSSRProxy with httptest.Server)

### Code review instructions
- Start with `pkg/help/server/serve.go` — look for `newSSRProxy` function
- Check ServeOption pattern and WithSSRURL()
- Run: `go test ./pkg/help/server/... -count=1`

---

## Step 3: Fix SSR runtime issues and create devctl plugin

Fixed three SSR runtime bugs and created a devctl plugin for local development.

### Prompt Context

**User prompt (verbatim):** "Add devctl support to start the docs server + the SSR sidecar for testing, so that after writing the code, things can be tested properly."

**Assistant interpretation:** Create a devctl plugin that starts both the Go docs server and Node SSR sidecar together, wiring the --ssr-url flag so the Go server proxies to the sidecar.

**Inferred user intent:** Enable easy local testing of the full SSR stack without manually starting multiple processes.

**Commit (code):** dab2fc9 — "DOCSCTL-SSR: add devctl plugin for docs server + SSR sidecar, fix SSR runtime issues"

### What I did
- Created `plugins/glazed.py` — devctl plugin implementing config.mutate, validate.run, and launch.plan
- Created `.devctl.yaml` — plugin wiring config
- Added `.devctl/` to `.gitignore`
- Fixed server.mjs: added `window` mock before dynamic import of SSR bundle (site-config.js and api.ts reference `window` at module-load time)
- Fixed server.mjs: Express 5 wildcard route syntax `{*path}` instead of `*`
- Fixed vite.config.ts: added `react` and `use-sync-external-store` to ssr.noExternal to prevent dual-React-instance hook errors

### Why
Without the devctl plugin, testing SSR requires manually starting 3 processes (build, Go server, Node sidecar) and wiring them together. The plugin automates this with health checks and process supervision.

### What worked
- `devctl up` successfully starts both services with health checks
- Go server proxies page requests to SSR sidecar → returns pre-rendered HTML
- `devctl plan` shows correct config (ports, URLs)
- `devctl validate` passes (checks go, node, pnpm, node_modules)

### What didn't work
- First attempt: SSR server crashed with `ReferenceError: window is not defined` because ESM static imports are hoisted — the window mock was set up AFTER the import already executed
- First attempt: Express 5 route `app.get('*')` throws `PathError: Missing parameter name` — Express 5 requires `{*name}` syntax
- First attempt: `react` missing from Vite `ssr.noExternal` caused "Invalid hook call" — the SSR bundle used React from global node_modules instead of the bundled copy, creating two React instances

### What I learned
- ESM import hoisting: static imports execute before any runtime code. Use dynamic `await import()` to set up globals first.
- Express 5 breaking change: `*` wildcard replaced by `{*param}` named wildcard
- Vite SSR noExternal: when you mark some packages as noExternal, ALL unlisted packages get externalized. Must include `react` explicitly.
- `use-sync-external-store` is a transitive dependency of `@reduxjs/toolkit` that also needs noExternal

### What was tricky to build
- The window mock ordering: I initially tried to set up the mock before a static `import` statement, but ESM hoists imports. The solution was to use `await import()` (dynamic import) in server.mjs.
- The dual React instance: the error "Invalid hook call" is notoriously hard to debug. The root cause was that `react` was NOT in the noExternal list, so Vite externalized it, and Node resolved it from the global `node_modules` instead of bundling it.

### What warrants a second pair of eyes
- The `globalThis.window` mock in server.mjs — is it safe to set this before loading the SSR bundle? Could there be side effects?
- The devctl plugin's launch.plan: the SSR sidecar uses `exec node server.mjs` to replace the bash shell. Does this interact correctly with devctl's PID tracking?

### What should be done in the future
- Pre-populate RTK Query cache in entry-server.tsx for full SSR content (currently renders "loading" skeleton)
- Add SSR proxy unit tests
- Create SSR Dockerfile for k3s deployment
- Run a14y audit against SSR-enabled server to verify score improvement

### Code review instructions
- Start with `plugins/glazed.py` — the devctl plugin
- Check `web/server.mjs` — the window mock and dynamic import
- Check `web/vite.config.ts` — the updated noExternal list
- Test: `devctl up && curl http://127.0.0.1:8088/ | grep 'div id="root"'`
