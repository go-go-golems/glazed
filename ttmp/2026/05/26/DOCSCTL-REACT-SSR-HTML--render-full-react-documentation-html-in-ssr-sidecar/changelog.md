# Changelog

## 2026-05-26

- Initial workspace created


## 2026-05-26

Created continuation ticket, tasks, diary, and React SSR HTML rendering implementation guide for completing data-backed SSR after the sidecar deployment.

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/server.mjs — Sidecar serialization and HTML injection target
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-client.tsx — Primary implementation target for client store hydration
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-server.tsx — Primary implementation target for RTK Query cache preloading


## 2026-05-26

Implemented local data-backed React SSR: shared routes, request-scoped store factory, RTK Query cache preloading, client hydration from serialized Redux state, safe sidecar serialization, SSR tests, and local Node/Docker validation.

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/server.mjs — Await renderApp and serialize returned Redux state safely
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-client.tsx — Client hydration from preloaded Redux state
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-server.test.tsx — SSR package and section HTML regression tests
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-server.tsx — Request-scoped SSR store and RTK Query cache preloading


## 2026-05-26

Rolled data-backed React SSR to production with Glazed sha-981a6db and k3s commit 2a27cdb; verified curl-level package index/article HTML, preloaded RTK Query state, static asset MIME types, markdown mirrors, API health, registry health, and Playwright hydration.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml — Production image rollout to sha-981a6db (commit 2a27cdb)
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REACT-SSR-HTML--render-full-react-documentation-html-in-ssr-sidecar/sources/02-production-ssr-html-evidence.md — Production acceptance evidence
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/src/entry-server.tsx — Production SSR data-backed render implementation (commit 981a6db)

