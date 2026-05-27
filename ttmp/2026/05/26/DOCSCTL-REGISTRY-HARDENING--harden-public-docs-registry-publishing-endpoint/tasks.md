# Tasks

## TODO

- [x] Phase 2: Add request ID, structured access logs, explicit production body limit, publish concurrency limit, and basic rate limiting.
- [x] Phase 3: Add immutable release-version policy with idempotent same-SHA retries and quota checks.
- [x] Phase 4: Enrich `PublisherIdentity` with non-sensitive JWT claims and emit publish audit events.
- [x] Phase 5: Add metrics and alerting, or document log-based alerts if metrics are not available yet.
- [x] Phase 6: Add local and production-safe negative auth proof cases.
- [x] Phase 7: Roll out to production and capture validation evidence.

## DONE

- [x] Create `DOCSCTL-REGISTRY-HARDENING` ticket.
- [x] Write intern-oriented analysis/design/implementation guide.
- [x] Relate core source, workflow, Terraform, and k3s files to the design doc.
- [x] Upload guide to reMarkable.
- [x] Phase 2.1: Add request IDs and structured access logging around registry handlers
- [x] Phase 2.2: Add publish concurrency limiting and basic per-IP route-class rate limiting
- [x] Phase 2.3: Add handler tests for rate limiting, concurrency, and explicit body-limit edge cases
- [x] Phase 2.4: Make production docs-registry upload/body/rate/concurrency settings explicit in k3s GitOps
- [x] Phase 3.1: Enforce immutable package versions with idempotent same-SHA retries
- [x] Phase 3.2: Add per-package byte and version-count quotas with stable API errors
- [x] Phase 3.3: Make production quota and overwrite settings explicit in k3s GitOps
- [x] Phase 4.1: Copy non-sensitive Vault publish JWT claims into PublisherIdentity
- [x] Phase 4.2: Emit publish-specific audit events for auth, upload, validation, policy, and success outcomes
- [x] Phase 4.3: Add tests for identity claim propagation and audit/error response behavior
- [x] Phase 5.1: Expose in-process Prometheus metrics for HTTP requests and publish outcomes
- [x] Phase 5.2: Add tests for metrics counters and /metrics output
- [x] Phase 5.3: Document alert queries for auth failures, conflicts, quotas, 5xxs, and rate/concurrency rejections
- [x] Phase 6.1: Add HTTP-level tests for unauthenticated, forbidden, invalid DB, duplicate-version, and quota negative responses
- [x] Phase 6.2: Document JWT and GitHub/Vault negative proof matrix
- [x] Phase 6.3: Add production-safe negative probe script that does not require secrets
- [x] Phase 7.1: Push Phase 4-6 Glazed commits and build GHCR images
- [x] Phase 7.2: Deploy Phase 4-6 images through k3s GitOps
- [x] Phase 7.3: Validate health, metrics, safe negative probe, and audit logs in production
