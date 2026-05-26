---
Title: Diary
Ticket: DOCSCTL-REGISTRY-HARDENING
Status: active
Topics:
    - docs-yolo
    - cicd
    - oidc
    - backend
    - security
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/docs-registry/main.go
      Note: Added registry hardening CLI flags for concurrency and per-client rate limits (commit f68238b)
    - Path: pkg/help/publish/registry.go
      Note: Wired request ID/access-log/rate-limit middleware and publish concurrency gate into registry handler (commit f68238b)
    - Path: pkg/help/publish/registry_middleware.go
      Note: New request ID
    - Path: pkg/help/publish/registry_test.go
      Note: Added tests for request IDs
ExternalSources: []
Summary: Chronological diary for docs-registry hardening design and implementation.
LastUpdated: 2026-05-26T19:20:00-04:00
WhatFor: Use this to resume hardening work and understand what was created, validated, and uploaded.
WhenToUse: Before implementing future DOCSCTL-REGISTRY-HARDENING phases.
---


# Diary

## Goal

This diary captures the creation of the `DOCSCTL-REGISTRY-HARDENING` ticket and the first intern-oriented analysis/design/implementation guide for hardening the public `docs-registry` upload endpoint.

## Step 1: Create the hardening ticket and intern implementation guide

The goal of this step was to turn the informal hardening list into a durable ticket workspace with a clear design document. The guide explains the current docs publishing system, the live registry API, the Vault OIDC authentication model, the k3s deployment topology, and the hardening work required before onboarding more packages.

The result is a ticket that can be handed to a new intern as a map of what to read, why each subsystem exists, what risks remain, and how to implement the next phases safely.

### Prompt Context

**User prompt (verbatim):** "go ahead, create a new ticket and Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new docmgr ticket for docs-registry hardening, write a detailed technical design/implementation guide aimed at a new intern, store it in the ticket workspace, and upload it to reMarkable.

**Inferred user intent:** The user wants registry hardening to become a tracked workstream with enough technical context that a new contributor can implement it without rediscovering the existing CI/CD, Vault OIDC, registry, and k3s architecture.

**Commit (code):** N/A — documentation-only step.

### What I did

- Created docmgr ticket `DOCSCTL-REGISTRY-HARDENING` in the Glazed ticket workspace.
- Added design document `design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md`.
- Added this diary document at `reference/01-diary.md`.
- Wrote a detailed guide covering:
  - system topology,
  - registry API references,
  - current security baseline,
  - threat/problem statement,
  - rate limits,
  - body-size limits,
  - storage quotas,
  - immutable release versions,
  - structured audit logs,
  - metrics and alerts,
  - negative auth proofs,
  - rollback path,
  - implementation phases,
  - intern reading order,
  - review checklist.
- Updated `tasks.md` with the planned implementation phases.

### Why

- The registry is now public and authenticated by Vault OIDC, but public exposure still requires operational safeguards.
- Broader onboarding of `pinocchio`, `sqleton`, and `remarquee` should happen after hardening is designed and tracked.
- A detailed guide reduces context loss and makes follow-up implementation easier to review.

### What worked

- `docmgr ticket create-ticket` created the workspace cleanly.
- `docmgr doc add` created both the design doc and diary with frontmatter.
- Existing source files provided enough detail to document the live architecture accurately.

### What didn't work

- N/A for this step. No implementation or validation failures occurred while creating the guide.

### What I learned

- The current registry already has several important baseline controls: OIDC/JWKS validation, package claim checks, default upload byte limits, SQLite validation, and filesystem path safety.
- The main remaining risk is operational hardening: rate, quota, immutability, logs, alerts, and negative proof evidence.

### What was tricky to build

- The guide had to separate what is already implemented from what is only proposed. The current system already enforces a default `64 MiB` upload cap in code, but production should still make that limit explicit in GitOps and add tests/audit fields around it.
- The document also had to avoid treating Vault claim binding and registry claim checks as duplicates. Vault enforces repository/workflow/ref/event constraints before token minting; the registry enforces token validity and requested package matching at upload time.

### What warrants a second pair of eyes

- The proposed overwrite policy should be reviewed: idempotent same-SHA republish is convenient, but different-SHA overwrite rejection changes current `DirectoryPackageStore` behavior.
- Quota defaults should be checked against realistic package sizes and retention expectations.
- Metrics/alerting recommendations should be aligned with the actual k3s observability stack.

### What should be done in the future

- Implement Phase 2 handler-level limits and audit foundation.
- Implement immutability and quotas before onboarding additional packages.
- Add negative auth proof cases and store the evidence in the ticket.

### Code review instructions

- Start with the design doc:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md`
- Compare recommendations against the current implementation:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/jwt_auth.go`
- Validate ticket hygiene with:
  - `docmgr doctor --ticket DOCSCTL-REGISTRY-HARDENING --stale-after 30`

### Technical details

Ticket path:

```text
/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint
```

Upload destination:

```text
/ai/2026/05/26/DOCSCTL-REGISTRY-HARDENING
```

## Step 2: Upload the hardening guide to reMarkable

The goal of this step was to deliver the ticket guide to reMarkable as a readable PDF bundle. I bundled the design guide, task list, and diary so the uploaded artifact includes both the technical plan and the ticket context.

The upload succeeded, which means the guide is now available at the expected dated reMarkable folder for review away from the terminal.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Upload the completed hardening guide and supporting ticket context to reMarkable.

**Inferred user intent:** The user wants the design package available on reMarkable for reading and review.

**Commit (code):** N/A — documentation/upload-only step.

### What I did

- Ran `remarquee upload bundle` with the design guide, `tasks.md`, and diary.
- Used remote directory `/ai/2026/05/26/DOCSCTL-REGISTRY-HARDENING`.
- Used document name `DOCSCTL Registry Hardening Guide`.

### Why

- The user explicitly requested the guide be uploaded to reMarkable.
- Bundling the guide with tasks and diary gives the reviewer both the implementation plan and the work history.

### What worked

- Upload completed successfully with:
  - `OK: uploaded DOCSCTL Registry Hardening Guide.pdf -> /ai/2026/05/26/DOCSCTL-REGISTRY-HARDENING`

### What didn't work

- N/A. The first upload attempt succeeded.

### What I learned

- The Markdown bundle rendered successfully enough for `remarquee` to upload without additional formatting changes.

### What was tricky to build

- N/A for the upload itself. The main operational constraint was to avoid unnecessary reMarkable status/list calls and rely on the upload command success output.

### What warrants a second pair of eyes

- Review the uploaded PDF for readability of long code blocks and diagrams on the reMarkable display.

### What should be done in the future

- Re-upload with `--force` only if the guide is materially revised and the existing PDF should be replaced.

### Code review instructions

- Confirm the source bundle files are correct:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/tasks.md`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/reference/01-diary.md`

### Technical details

Upload command used:

```bash
remarquee upload bundle \
  /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/design-doc/01-docs-registry-hardening-analysis-design-and-implementation-guide.md \
  /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/tasks.md \
  /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/reference/01-diary.md \
  --name "DOCSCTL Registry Hardening Guide" \
  --remote-dir "/ai/2026/05/26/DOCSCTL-REGISTRY-HARDENING" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Add Phase 2 registry request limits and access logs

This step implemented the first executable hardening layer around the public registry API. The registry now assigns or preserves a request ID, emits a structured access log event for each request, can reject request bursts with a simple per-client token bucket, and can reject excess concurrent publish uploads before the expensive authorization/upload/SQLite-validation path runs.

This does not complete all registry hardening. It creates the foundation for bounded request handling and auditability while keeping the existing publish API stable.

### Prompt Context

**User prompt (verbatim):** "go ahead, add tasks, keep a detailed diary as you work, commit at appropriate intervals."

**Assistant interpretation:** Break the hardening work into actionable tasks, start implementation, record detailed diary entries while working, and make focused commits.

**Inferred user intent:** The user wants the hardening ticket to move from design into implementation while preserving reviewable history and documentation.

**Commit (code):** f68238b — "docs-registry: add request limits and access logs"

### What I did

- Added granular Phase 2 tasks for:
  - request IDs and access logs,
  - publish concurrency and rate limiting,
  - handler test coverage,
  - explicit k3s production settings.
- Added CLI flags to `cmd/docs-registry/main.go`:
  - `--max-concurrent-uploads` with default `2`,
  - `--rate-limit-requests-per-minute` with default `60`,
  - `--rate-limit-burst` with default `10`.
- Added middleware in `pkg/help/publish/registry_middleware.go`:
  - request ID propagation via `X-Request-ID`,
  - generated request IDs when absent,
  - structured access logging through `slog.Info`,
  - client IP extraction from `X-Forwarded-For`, `X-Real-IP`, or `RemoteAddr`,
  - route classification for `publish`, `list`, `health`, and `other`,
  - a small in-process token-bucket rate limiter.
- Updated `pkg/help/publish/registry.go`:
  - wraps the registry mux with request ID, access-log, and rate-limit middleware,
  - creates a publish semaphore when `MaxConcurrentUploads > 0`,
  - rejects excess concurrent uploads with `429` and error code `too_many_concurrent_uploads`.
- Added tests in `pkg/help/publish/registry_test.go` for:
  - unknown-length oversized uploads,
  - request ID response header preservation,
  - per-client route-class rate limiting,
  - concurrent publish rejection while a first publish is in progress.

### Why

- A public registry must cheaply reject abusive or accidental bursts instead of allowing every request to reach JWT validation, disk writes, and SQLite validation.
- Request IDs make it possible to correlate client-facing failures with server logs.
- Structured logs are the first audit/observability layer before a full metrics implementation.
- Concurrency limiting protects the single PVC-backed registry from multiple simultaneous expensive publish paths.

### What worked

- Local validation passed:
  - `go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl`
- Pre-commit checks passed during commit:
  - targeted tests,
  - golangci-lint,
  - gosec,
  - govulncheck.
- The new middleware did not require changing the public API or `docsctl publish` client contract.

### What didn't work

- N/A for this step. The first implementation passed the targeted test suite after formatting.

### What I learned

- The existing handler shape made it easy to wrap the mux and keep hardening orthogonal to the core publish flow.
- The concurrency gate belongs at the beginning of `handlePublishSQLite`, before auth and upload reads, because it is intended to protect the whole expensive path.
- A one-replica in-process token bucket is acceptable for the current deployment, but it must be revisited if the registry is scaled horizontally.

### What was tricky to build

- The concurrency test needed a fake store that blocks after the first request reaches `Publish`. That proves the semaphore remains held through validation and storage publication, and a second upload receives `429` while the first is in progress.
- The rate limiter needed to key by both client IP and route class. Without the route class, health/list/publish requests from the same client would consume the same bucket, which is not the behavior described in the design.
- The middleware order matters: request ID wrapping is outermost so the access logger and downstream handler see the same request ID and the response carries the same `X-Request-ID` value.

### What warrants a second pair of eyes

- Default rate limit values (`60/min`, burst `10`) should be reviewed against real GitHub Actions behavior and any health-check traffic that reaches the registry host.
- The current in-process limiter is deliberately simple and not distributed. That is fine for one replica but insufficient for horizontal scaling.
- The concurrency limiter currently returns `429`; reviewers may prefer `503` with `Retry-After` for overload semantics.

### What should be done in the future

- Make production settings explicit in the k3s deployment once a new image containing these flags is built.
- Add richer publish-specific audit events with JWT claims in Phase 4.
- Add metrics/alerts in Phase 5.

### Code review instructions

- Start with `pkg/help/publish/registry.go` to see how the middleware and semaphore are wired.
- Then review `pkg/help/publish/registry_middleware.go` for request ID, access logging, client IP extraction, route classification, and token-bucket logic.
- Review `cmd/docs-registry/main.go` to confirm CLI defaults are safe and documented.
- Review `pkg/help/publish/registry_test.go` for expected status codes and failure behavior.
- Validate with:
  - `go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl`

### Technical details

New error response for concurrent publish overload:

```json
{
  "error": "too_many_concurrent_uploads",
  "message": "too many concurrent uploads"
}
```

New error response for rate limiting:

```json
{
  "error": "rate_limited",
  "message": "too many requests"
}
```
