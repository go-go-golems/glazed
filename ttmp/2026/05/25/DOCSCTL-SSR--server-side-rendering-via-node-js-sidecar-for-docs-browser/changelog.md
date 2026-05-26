# Changelog

## 2026-05-25

- Initial workspace created
- Created SSR entry points (entry-server.tsx, entry-client.tsx) and Express server (server.mjs) (commit 7ea2f9d)
- Added SSR build config to vite.config.ts and package.json (commit 7ea2f9d)
- Added --ssr-url flag and reverse proxy to Go server (commit 9d99c44)
- Created devctl plugin (plugins/glazed.py) for docs server + SSR sidecar (commit dab2fc9)
- Fixed SSR runtime issues: window mock for Node, Express 5 wildcard syntax, dual React instance (commit dab2fc9)
- Verified full SSR stack works: devctl up → Go server proxies to SSR sidecar → pre-rendered HTML returned (commit dab2fc9)

## 2026-05-25

Step 3: Fix SSR runtime issues (window mock, Express 5 syntax, dual React) + create devctl plugin (commit dab2fc9)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/plugins/glazed.py — Devctl plugin
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/server.mjs — Window mock + dynamic import + Express 5 fix
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/vite.config.ts — Updated noExternal list

