# Changelog

## 2026-05-25

- Initial workspace created


## 2026-05-25

Created ticket and design doc for semantic URL routing. Investigated full stack (Go server, React SPA, SQLite store, publish pipeline). Identified 6 gaps (G1-G6). Proposed 3-phase implementation: BrowserRouter migration, heading anchors, edge cases.


## 2026-05-25

Implemented semantic URL routing. Switched HashRouter to BrowserRouter with /{package}/{version}/sections/{slug} routes. Fixed API base URL resolution (explicit apiBaseUrl in site-config.js). Fixed stale RTK Query cache when navigating away from a slug. Added legacy hash URL redirect in index.html. All 14 tests pass, TypeScript compiles, production build succeeds. Verified with live Go server + Vite dev server: root redirect, section navigation, package switching, direct URL navigation, heading anchors, browser back/forward all work.

