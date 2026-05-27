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


## 2026-05-26

Phase 2 deployed to docs-yolo with explicit registry body, rate, and concurrency settings (Glazed e50da7e, k3s 99e3f5f)

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml — Production image and registry hardening args


## 2026-05-26

Phase 3: added immutable version policy, idempotent retries, package quotas, and stable publish policy errors (commit 4d519f8)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go — New overwrite/quota CLI flags
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go — Immutability and quota enforcement
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store_test.go — Storage policy tests
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/publish_policy.go — Policy error contracts


## 2026-05-26

Phase 3 deployed immutable registry policy and quotas; fixed boolean flag CrashLoopBackOff with --allow-overwrite=false (Glazed 1e27788, k3s c349919/a702799/5f395e8)

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml — Production immutable/quota registry args and rollout fix

