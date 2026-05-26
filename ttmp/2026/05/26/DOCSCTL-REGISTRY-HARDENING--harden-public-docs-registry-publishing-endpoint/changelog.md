# Changelog

## 2026-05-26

- Initial workspace created


## 2026-05-26

Created hardening ticket and intern-oriented registry hardening guide

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md — Primary design and implementation guide
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/reference/01-diary.md — Diary for ticket setup


## 2026-05-26

Uploaded hardening guide bundle to reMarkable

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/reference/01-diary.md — Records successful reMarkable upload


## 2026-05-26

Phase 2: added request IDs, access logs, rate limits, concurrency limits, and handler tests (commit f68238b)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go — New hardening CLI flags
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go — Middleware and publish semaphore wiring
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry_middleware.go — Request ID/access log/rate limiter implementation
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry_test.go — Hardening tests

