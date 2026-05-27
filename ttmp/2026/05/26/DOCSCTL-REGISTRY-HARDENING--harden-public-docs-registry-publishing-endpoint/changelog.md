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


## 2026-05-26

Phase 4: enriched PublisherIdentity with JWT provenance claims and added publish-specific audit events without token logging (commit 889dffe)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/audit.go — Publish audit event fields and structured logging
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go — PublisherIdentity provenance fields
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/jwt_auth.go — JWT claim propagation into PublisherIdentity
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go — Publish outcome audit emission and error-code mapping


## 2026-05-26

Phase 5: exposed Prometheus text metrics for registry requests and publish outcomes, and documented alert sketches/log fallbacks (commit ee4ffe6)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/metrics.go — Metrics collector and /metrics renderer
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go — /metrics route and publish outcome metric recording
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry_middleware.go — HTTP request metric recording
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md — Phase 5 metric and alert guidance


## 2026-05-26

Phase 6: added HTTP-level negative response assertions, a negative proof plan, and a secret-free production-safe probe script (commit 1e14425)

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry_test.go — Stable negative HTTP response assertions
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/scripts/01-production-safe-negative-probes.sh — Production-safe unauthenticated probe script
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/sources/03-negative-proof-plan.md — Negative proof matrix and evidence rules


## 2026-05-26

Phase 7: deployed registry audit/metrics/negative-proof hardening to production, validated health, metrics, safe negative probe, and audit logs (Glazed 312fa79, k3s 05deb9c)

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml — Production images updated to sha-312fa79
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/sources/04-production-rollout-evidence.md — Production validation evidence


## 2026-05-26

Closed after Phase 7 production rollout: audit events, metrics, negative proof probes, k3s deployment, and production validation evidence are complete.

