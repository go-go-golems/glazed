---
ticket: DOCSCTL-URL-NAV
doc_type: reference
title: "Investigation diary"
status: active
intent: long-term
topics: [docsctl, urls, routing, spa, go, typescript]
created: 2026-05-25
---

# Investigation Diary

## Goal

Capture the investigation and design process for DOCSCTL-URL-NAV: proper semantic URL routing in the docsctl SPA.

## Step 1: Architecture Discovery and Gap Analysis

I examined the full stack of the docs.yolo.scapegoat.dev documentation system, from the Go HTTP server through the React SPA to the SQLite store and publish pipeline. The goal was to understand every component that touches URLs and identify what prevents deep linking.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to properly handle docsctl URL to have package + version + page + section in the URL and properly navigate to that page. (see https://docs.yolo.scapegoat.dev which doesnt' really handle that)."

**Assistant interpretation:** Design a system where docs.yolo.scapegoat.dev URLs encode package, version, page slug, and section heading, and navigating to such a URL loads the correct state.

**Inferred user intent:** The user wants shareable, bookmarkable URLs that fully encode the documentation navigation state, so that anyone can open a URL and land on the right page.

### What I did

- Read every Go source file in `pkg/help/server/`, `pkg/help/store/`, `pkg/help/model/`, `pkg/help/loader/`, `pkg/help/publish/`, and `cmd/docsctl/`
- Read every TypeScript source file in `web/src/`
- Read the SPA handler in `pinocchio/pkg/spa/spa.go`
- Visited https://docs.yolo.scapegoat.dev and observed the URL behavior (hash-based routing, no package/version in URL)
- Tested direct navigation to a semantic URL (`/pinocchio/v1.2.15/sections/slug`) which resulted in a blank page
- Identified 6 specific gaps (G1–G6)

### Why

The investigation was necessary to write an evidence-based design document. Without understanding how the Go server routes, how the SPA resolves API URLs, and how the data model identifies sections, any proposed solution would be speculative.

### What worked

- The browser testing immediately confirmed the core problem: `HashRouter` prevents server-side routing and swallows heading fragments
- The codebase is well-structured and the server already does SPA fallback correctly, which means the server-side changes are minimal
- The `site-config.js` runtime config mechanism already supports `apiBaseUrl`, making the API resolution fix trivial

### What didn't work

- Nothing failed during investigation; all source files were accessible and readable

### What I learned

- The `resolveApiBaseUrl` function in `api.ts` is the most dangerous piece of the migration: it derives the API URL from `window.location.pathname`, which changes completely with `BrowserRouter`
- The section identity is a composite key `(package_name, package_version, slug)`, not just `slug` — this is why the URL must include package and version
- The `SQLiteDirLoader` convention (`X/Y/X.db`) and the `DirectoryPackageStore` convention (`{pkg}/{ver}/{pkg}.db`) already establish a clear package/version hierarchy that the URL scheme should mirror

### What was tricky to build

- Understanding that `HashRouter` consumes the `#` fragment for routing, making heading anchors fundamentally broken — this is not a fixable bug but an architectural constraint of hash routing
- The `resolveApiBaseUrl` function has two code paths (static mode vs. server mode) and the pathname-based resolution only applies in server mode, but it's the default and the one that breaks

### What warrants a second pair of eyes

- The decision to use `_` as the URL segment for "no version" — is this convention clear enough?
- Whether the existing `HashRouter` bookmarks should be supported with a redirect script in `index.html`
- Whether the API should also support semantic URL paths (not just query params)

### What should be done in the future

- Add a `<script>` in `index.html` to redirect legacy `/#/sections/slug` URLs to the new format
- Consider adding `Accept: application/json` routing in the Go server for semantic URLs
- Add E2E tests with Playwright for the full URL navigation flow

### Code review instructions

- Start with `web/src/main.tsx` (the router change) and `web/src/App.tsx` (the URL state derivation)
- Verify that all `navigate()` calls include `/${packageName}/${version}` prefix
- Check that `resolveRuntimeBaseUrl` in `api.ts` returns `/api` explicitly
- Test with `glaze serve --from-sqlite-dir <dir>` and navigate to a semantic URL

### Technical details

Key command for testing locally:
```bash
cd glazed && go run ./cmd/glaze serve --from-sqlite-dir /path/to/help-dbs --address :8088
```

The production deployment uses:
```bash
glaze serve --from-sqlite-dir <dir> --reload-interval 30s
```

---

## Step 2: Implementation — BrowserRouter migration and semantic URLs

### Prompt Context

**User prompt (verbatim):** "alright, add tasks, then do the ticket. Commit at appropriate intervals, keep a diary as you go."

**Assistant interpretation:** Implement the design from Step 1. Create granular tasks, make the code changes, commit at logical boundaries, and update the diary.

**Inferred user intent:** Execute the full implementation, not just the design.

### What I did

**Commit 1 (d4c037a): Fixed API base URL resolution**
- Added `apiBaseUrl: '/api'` to `web/public/site-config.js`
- Updated `api.ts` module comment explaining why explicit apiBaseUrl is essential for BrowserRouter
- This ensures API calls always go to `/api` regardless of the browser pathname

**Commit 2 (f9aec88): Switched from HashRouter to BrowserRouter**
- `main.tsx`: Replaced `HashRouter` with `BrowserRouter` containing three routes: `/:package/:version/sections/:slug`, `/:package/:version`, and `*` (catch-all)
- `App.tsx`: Complete rewrite to derive package/version/slug from `useParams()` instead of `matchPath`. All `navigate()` calls now include the `/${package}/${version}` prefix. Added `versionFromUrl`/`versionToUrl` helpers that use `_` as the placeholder for unversioned packages. Added auto-redirect from `/` to default package/version. Added auto-redirect from nonexistent package to default.
- `App.test.tsx`: Switched from `HashRouter` to `MemoryRouter` with the same route structure. Updated assertions to match new URL patterns. Added test for empty state when no slug is in URL.
- `api.test.ts`: Added test for explicit `apiBaseUrl` taking precedence over pathname-derived URL.
- `vite.config.ts`: Updated comment from "HashRouter" to "BrowserRouter".
- `index.html`: Added legacy hash URL redirect script that converts `#/sections/slug` to `/default/_/sections/slug`.

**Commit 3 (c33a292): Fixed stale section data bug**
- When navigating from `/{pkg}/{ver}/sections/slug` to `/{pkg}/{ver}` (no slug), the `useGetSectionQuery` with `skip: true` returns the last cached result. The section was still showing even though the URL had no slug.
- Fix: `const section = activeSlug ? sectionData : undefined;`

### Why

The HashRouter prevented deep linking. The BrowserRouter migration enables semantic URLs. The API base URL fix prevents a critical bug where API calls would be routed to the wrong path. The stale section fix prevents confusing UX when switching packages.

### What worked

- The server already did SPA fallback correctly (serving `index.html` for all non-`/api` paths), so no Go code changes were needed
- The `site-config.js` mechanism already supported `apiBaseUrl`, making the API fix trivial
- The `useParams()` approach is much cleaner than `matchPath` for extracting route parameters
- All 14 tests pass, TypeScript compiles, production build succeeds
- Live browser testing confirmed: root redirect, section navigation, package switching, direct URL navigation, heading anchors, and browser back/forward all work

### What didn't work

- The initial implementation forgot that `useGetSectionQuery` with `skip: true` returns stale cached data. This caused the old section to still render after navigating to a URL without a slug. Fixed in commit 3.

### What I learned

- RTK Query's `skip` option prevents new fetches but returns the last cached result. This is documented but easy to miss. The fix pattern is to null out the data yourself when the query should not be active.
- The `_` convention for "no version" URLs works well in practice. The `versionFromUrl`/`versionToUrl` helpers make the conversion clean.
- The `ListPackages` API response already provides `defaultPackage` and `defaultVersion` which makes the root redirect straightforward.

### What was tricky to build

- The stale RTK Query cache was the trickiest issue. The symptom was subtle: switching from glazed's "Build Your First Glazed Command" section to pinocchio showed the pinocchio tree but the glazed section title in the content area. The fix was a single line but understanding the root cause took careful thought.
- The `handlePackageChange` function needed to navigate to a URL without a slug (just `/{pkg}/{ver}`), which naturally clears the section view after the stale-data fix.

### What warrants a second pair of eyes

- The legacy hash URL redirect in `index.html` defaults the package to "default" since old hash URLs didn't encode package/version. This might not match the actual default package on all deployments.
- The route `/:package/:version` could conflict with other two-segment paths if the app is extended in the future. Consider a prefix like `/docs/:package/:version` for future-proofing.

### What should be done in the future

- Rebuild the SPA assets embedded in the Go binary (`make fetch-spa` or equivalent) and deploy to docs.yolo.scapegoat.dev
- Add E2E tests with Playwright for the full URL navigation flow
- Consider adding a "Copy link" button next to section titles for easy URL sharing
- Consider adding `Accept: application/json` routing in the Go server for semantic URL paths

### Code review instructions

- Three commits: d4c037a (API base URL), f9aec88 (BrowserRouter), c33a292 (stale section fix)
- Start with `web/src/main.tsx` for the router setup
- Then `web/src/App.tsx` for the URL state derivation and navigation
- Then `web/src/App.test.tsx` for the test updates
- Verify `web/public/site-config.js` has `apiBaseUrl: '/api'`
- Run `pnpm test` and `npx tsc --noEmit` in `web/`
- Run `go test ./pkg/help/server/...` for the Go server

### Technical details

Testing commands:
```bash
# Build the Go binary
cd glazed && go build -o /tmp/glaze-test ./cmd/glaze

# Export help databases
/tmp/glaze-test help export --format sqlite --output-path /tmp/help-dbs/glazed/glazed.db
pinocchio help export --format sqlite --output-path /tmp/help-dbs/pinocchio/v1.2.15/pinocchio.db

# Start the server
/tmp/glaze-test serve --from-sqlite-dir /tmp/help-dbs --address :8088

# For frontend dev, use the Vite dev server alongside the Go server:
cd web && pnpm dev
```
