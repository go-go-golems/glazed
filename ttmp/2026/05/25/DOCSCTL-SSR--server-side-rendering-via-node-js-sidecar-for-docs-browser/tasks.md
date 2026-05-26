# Tasks

## TODO

- [x] Pre-populate RTK Query cache in entry-server.tsx for full SSR content rendering (currently renders "loading" skeleton)
- [x] Add SSR proxy unit tests (newSSRProxy with httptest.Server)
- [x] Create SSR Dockerfile for k3s deployment
- [ ] Run a14y audit against SSR-enabled server to verify score improvement

## Done

- [x] Create entry-server.tsx — SSR entry point with StaticRouter (commit 7ea2f9d)
- [x] Create server.mjs — Express SSR HTTP server (commit 7ea2f9d)
- [x] Create entry-client.tsx — hydration entry point, update main.tsx (commit 7ea2f9d)
- [x] Add SSR build to vite.config.ts and package.json scripts (commit 7ea2f9d)
- [x] Add --ssr-url flag to Go server + reverse proxy logic (commit 9d99c44)
- [x] Create devctl plugin for starting docs server + SSR sidecar (commit dab2fc9)
- [x] Fix SSR runtime issues: window mock, Express 5 syntax, dual React (commit dab2fc9)
- [x] Test SSR locally (Go + Node sidecar), verify HTML output (commit dab2fc9)
