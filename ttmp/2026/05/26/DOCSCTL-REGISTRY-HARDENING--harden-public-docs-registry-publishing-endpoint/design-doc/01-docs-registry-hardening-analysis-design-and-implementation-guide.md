---
Title: docs-registry Hardening Analysis Design and Implementation Guide
Ticket: DOCSCTL-REGISTRY-HARDENING
Status: active
Topics:
    - docs-yolo
    - cicd
    - oidc
    - backend
    - security
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml
      Note: Production docs-yolo pod topology and docs-registry runtime arguments
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml
      Note: Public TLS ingress for docs browser and registry hosts
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml
      Note: Reusable GitHub workflow that mints publish JWTs and calls docsctl publish
    - Path: ../../../../../../../../../../code/wesen/terraform/vault/github-actions/envs/k3s/main.tf
      Note: Vault GitHub OIDC roles and package-scoped Identity/OIDC publish token roles
    - Path: cmd/docs-registry/main.go
      Note: Registry CLI flags and HTTP server construction for auth mode
    - Path: pkg/help/publish/auth.go
      Note: PublisherAuth interface
    - Path: pkg/help/publish/directory_store.go
      Note: Filesystem publication
    - Path: pkg/help/publish/jwt_auth.go
      Note: Vault Identity/OIDC JWT verification and package claim authorization
    - Path: pkg/help/publish/registry.go
      Note: Upload API handler
    - Path: pkg/help/publish/sqlite_validator.go
      Note: Read-only SQLite validation gate before package publication
ExternalSources: []
Summary: Design and implementation guide for hardening the public docs-registry upload service before broader package onboarding.
LastUpdated: 2026-05-26T19:20:00-04:00
WhatFor: 'Use this when implementing or reviewing docs-registry production hardening: limits, quotas, audit logs, immutability, alerts, rollback, and negative auth proofs.'
WhenToUse: Before onboarding additional packages such as pinocchio, sqleton, and remarquee to automated docs publishing.
---











# docs-registry Hardening Analysis Design and Implementation Guide

## Executive Summary

`docs-registry` is the write-side API for `docs.yolo.scapegoat.dev`. GitHub release workflows export a Glazed help SQLite database, obtain a short-lived Vault Identity/OIDC publish JWT, and upload that SQLite database to `https://docs-registry.yolo.scapegoat.dev`. The public reader service, `docs-browser`, reloads the published package/version files from shared storage and serves the documentation site.

The security posture is already much better than a static shared token: uploads require a Vault-signed JWT with `aud = docs-registry`, `token_use = docsctl-publish`, and `package = <requested package>`. Vault only mints those JWTs for specific GitHub repositories, release-tag pushes, and workflow references. However, the registry endpoint is now public. Public exposure means the next phase must harden operational behavior, not just cryptographic authentication.

This ticket defines the hardening work required before broad onboarding of `pinocchio`, `sqleton`, and `remarquee` docs publishing. The design focuses on concrete controls: request limits, body-size enforcement, storage quotas, immutable release versions, structured audit logging, metrics/alerts, negative authentication proof cases, and a documented rollback path.

## System Map

The docs publishing system has four major zones:

- **Package repository**: for example `glazed`, later `pinocchio`, `sqleton`, or `remarquee`.
- **Reusable CI workflow**: `infra-tooling/.github/workflows/publish-docsctl.yml`.
- **Vault**: validates GitHub OIDC and mints package-scoped publish JWTs.
- **k3s docs-yolo deployment**: hosts `docs-browser`, `docs-registry`, and `docs-ssr` in one pod with shared package storage.

```text
GitHub release tag push
        |
        v
Package release workflow
        |
        | uses reusable workflow_call
        v
infra-tooling publish-docsctl.yml
        |
        | 1. export help sqlite
        | 2. login to Vault via GitHub OIDC
        | 3. mint Vault Identity/OIDC publish JWT
        | 4. docsctl publish PUT request
        v
https://docs-registry.yolo.scapegoat.dev
        |
        | validate JWT, validate SQLite, write package/version DB
        v
/var/lib/glazed-docs/packages
        |
        | docs-browser reload interval
        v
https://docs.yolo.scapegoat.dev
```

The k3s pod currently contains:

```text
Deployment docs-yolo
  container docs-browser  : serves public docs UI/API on :8088
  container docs-registry : accepts authenticated uploads on :8090
  container docs-ssr      : renders SSR HTML on :8089, localhost-only inside pod
  PVC package-root        : /var/lib/glazed-docs/packages
```

Important deployment file:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`

Important ingress file:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml`

The public hosts are:

- `https://docs.yolo.scapegoat.dev` for read traffic.
- `https://docs-registry.yolo.scapegoat.dev` for publish traffic.

## Current Upload API

The write API is intentionally small.

### Health

```http
GET /healthz HTTP/1.1
Host: docs-registry.yolo.scapegoat.dev
```

Response:

```json
{"ok":true}
```

### List Published Packages

```http
GET /v1/packages HTTP/1.1
Host: docs-registry.yolo.scapegoat.dev
```

Response shape:

```json
{
  "packages": [
    {
      "packageName": "glazed",
      "version": "v1.3.4",
      "sectionCount": 333,
      "slugCount": 333,
      "path": "glazed/v1.3.4/glazed.db",
      "sha256": "...",
      "publishedBy": "go-go-golems/glazed",
      "publishedAt": "2026-05-26T...Z"
    }
  ]
}
```

### Publish SQLite Database

```http
PUT /v1/packages/{package}/versions/{version}/sqlite HTTP/1.1
Host: docs-registry.yolo.scapegoat.dev
Authorization: Bearer <vault-identity-oidc-publish-jwt>
Content-Type: application/octet-stream

<sqlite bytes>
```

Successful response shape:

```json
{
  "ok": true,
  "package": {
    "packageName": "glazed",
    "version": "v1.3.4",
    "sectionCount": 333,
    "slugCount": 333,
    "path": "glazed/v1.3.4/glazed.db",
    "sha256": "...",
    "publishedBy": "go-go-golems/glazed",
    "publishedAt": "2026-05-26T...Z"
  },
  "validation": {
    "packageName": "glazed",
    "version": "v1.3.4",
    "sectionCount": 333,
    "slugCount": 333
  },
  "actor": {
    "subject": "go-go-golems/glazed",
    "packageName": "glazed",
    "method": "vault-oidc-jwt"
  }
}
```

Current implementation entry points:

- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/jwt_auth.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/sqlite_validator.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go`

## Current Security Baseline

The current baseline already includes several important controls.

### JWT Authentication

`docs-registry --auth-mode vault-oidc-jwt` uses OIDC discovery and JWKS verification through `github.com/coreos/go-oidc/v3/oidc`.

Current production command:

```text
/usr/local/bin/docs-registry
  --address :8090
  --package-root /var/lib/glazed-docs/packages
  --auth-mode vault-oidc-jwt
  --jwt-issuer https://vault.yolo.scapegoat.dev/v1/identity/oidc
  --jwt-client-id docs-registry
```

The registry validates:

- The token is present as a Bearer token.
- The token signature validates against Vault Identity/OIDC JWKS.
- The token issuer matches `https://vault.yolo.scapegoat.dev/v1/identity/oidc`.
- The token audience matches `docs-registry`.
- The token is time-valid according to OIDC verification.
- The custom claim `token_use` equals `docsctl-publish`.
- The custom claim `package` matches the requested path package.

### Vault Claim Binding

Terraform binds GitHub OIDC login roles to specific repositories and release workflows.

Important file:

- `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`

For each docs publisher, Vault requires claims like:

```hcl
bound_claims = {
  repository_owner = "go-go-golems"
  repository       = each.value.repository
  repository_id    = each.value.repository_id
  ref_type         = "tag"
  ref              = "refs/tags/v*"
  event_name       = "push"
  workflow_ref     = each.value.workflow_ref
  job_workflow_ref = var.docsctl_publish_job_workflow_ref
}
```

This means the registry does not need to understand every GitHub OIDC detail. Vault enforces those details before it will mint a publish JWT. The registry then enforces the claims that matter at upload time.

### Upload Body Limit

`docs-registry` already has `--max-upload-bytes`, defaulting to `64 MiB`. `RegistryHandler.receiveUpload` rejects uploads whose `Content-Length` is too large and also wraps the request body in `io.LimitReader(maxBytes + 1)` so clients without `Content-Length` cannot stream unbounded data.

This is good baseline protection, but it needs operational configuration, test coverage, and possibly separate ingress-level limits.

### SQLite Validation

The uploaded file is opened read-only and inspected before publication. The validator currently checks that:

- The file opens as SQLite in read-only query-only mode.
- The `sections` table exists.
- Required columns such as `slug` and `title` exist.
- The database contains at least one section.
- Slugs are non-empty.
- Slugs are unique.
- Package/version metadata mismatches are reported as warnings when metadata columns exist.

### Filesystem Safety

Package names and versions are validated as safe path segments before writing to disk. `DirectoryPackageStore` computes the destination under the package root and checks that computed paths do not escape the configured root.

Current materialized path shape:

```text
/var/lib/glazed-docs/packages/{package}/{version}/{package}.db
/var/lib/glazed-docs/packages/catalog.json
```

## Problem Statement

The current registry is authenticated, but public endpoints fail in more ways than private endpoints. Hardening must assume that anyone on the Internet can send requests to `docs-registry.yolo.scapegoat.dev`, including malformed bodies, very large bodies, high-rate requests, unauthenticated requests, replayed old tokens, and repeated writes to existing versions.

The main risks are:

- **Resource exhaustion**: too many requests, too many uploads, large request bodies, or repeated validation work can consume CPU, memory, disk I/O, temporary storage, or persistent storage.
- **Disk exhaustion**: even valid authenticated publishers can accidentally create many versions or large files.
- **Mutable release history**: overwriting a previously published release tag can surprise readers and complicate incident response.
- **Insufficient audit trail**: if something bad is published, operators need to know who published it, from which repository/run, when, how large it was, and what checks passed or failed.
- **Weak operational visibility**: without metrics and alerts, the first sign of trouble may be broken docs or a full disk.
- **Unproven negative cases**: successful publish proof is not enough; we need evidence that invalid tokens and invalid requests fail as expected.
- **Rollback uncertainty**: if Vault OIDC, JWKS discovery, or registry JWT validation fails, operators need a clear and tested rollback path.

The hardening goal is not to make the registry complex. The goal is to keep the API small while making failure behavior explicit, bounded, observable, and easy to operate.

## Proposed Solution

Implement hardening as a set of narrow layers around the current upload pipeline.

```text
HTTP request
  |
  v
Ingress/TLS
  |
  v
registry middleware
  - request ID
  - structured access log
  - method/path allowlist via ServeMux
  - rate limit
  - body cap
  - timeouts
  |
  v
PublisherAuth
  - OIDC signature/audience/issuer/time
  - token_use claim
  - package claim
  |
  v
Upload receive temp file
  - max bytes
  - temp dir permissions
  - cleanup
  |
  v
SQLite validation
  - schema
  - slugs
  - package/version warnings
  |
  v
Publish policy
  - validate package/version
  - immutable release version check
  - per-package/version quotas
  |
  v
DirectoryPackageStore
  - atomic write
  - catalog update
  |
  v
metrics + audit event + response
```

The implementation should keep each concern separate:

- `RegistryHandler` owns HTTP request handling.
- `PublisherAuth` owns authentication/authorization.
- `PackageStore` owns persistence.
- A new policy layer should own immutability and quota decisions.
- Middleware should own request ID, rate limits, and access logs.
- Metrics should be emitted at boundaries where outcomes are known.

## Hardening Requirement 1: Rate Limits

### Goal

Prevent unauthenticated and authenticated request bursts from consuming too much CPU, network, or disk I/O.

### Recommended Behavior

Use separate limits for low-cost read endpoints and high-cost write endpoints.

Suggested initial values:

- `GET /healthz`: generous, for probes and uptime checks.
- `GET /v1/packages`: moderate, because it reads catalog data.
- `PUT /v1/packages/{package}/versions/{version}/sqlite`: strict, because it performs auth, disk writes, SQLite open, validation, and catalog update.

A practical first pass can be in-process token buckets:

```text
per source IP:
  publish requests: 10 per minute, burst 3
  list requests:    60 per minute, burst 20

global:
  concurrent publish requests: 1 or 2
```

The registry currently runs as one replica because it writes to one PVC. In-process limits are acceptable for one replica. If the service scales horizontally later, move rate limits to Traefik middleware, Redis, or another shared limiter.

### Pseudocode

```go
type RateLimiter interface {
    Allow(key string, routeClass string) bool
}

func rateLimitMiddleware(next http.Handler, limiter RateLimiter) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := clientIP(r)
        class := classifyRoute(r.Method, r.URL.Path)
        if !limiter.Allow(key, class) {
            writeRegistryError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### File Targets

- Add middleware near `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go` or a new `registry_middleware.go`.
- Add CLI flags in `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go`.
- Add tests in `pkg/help/publish/*_test.go`.

## Hardening Requirement 2: Body-Size Limits

### Goal

Reject oversized uploads before the registry spends unnecessary time or disk space on them.

### Current State

`RegistryHandler.receiveUpload` already enforces `MaxUploadBytes` with both `ContentLength` and `io.LimitReader` checks. The default is `64 MiB`.

### Required Additions

- Set an explicit production value in the k3s deployment, not only the Go default.
- Add tests for:
  - body exactly at limit,
  - body over limit with `Content-Length`,
  - body over limit without `Content-Length`,
  - empty body.
- Add structured audit log fields for rejected size.
- Consider ingress-level body limits so Traefik rejects huge bodies before the pod sees them.

### Recommended Production Flag

```yaml
- --max-upload-bytes
- "67108864"
```

This makes the production limit visible in GitOps rather than hidden in source defaults.

### Pseudocode

```go
func (h *RegistryHandler) receiveUpload(r *http.Request) (string, error) {
    maxBytes := configuredOrDefault(h.MaxUploadBytes)

    if r.ContentLength > maxBytes {
        return "", MaxUploadError{Max: maxBytes, Observed: r.ContentLength}
    }

    reader := io.LimitReader(r.Body, maxBytes+1)
    written, err := io.Copy(tempFile, reader)
    if err != nil { return "", err }

    if written > maxBytes {
        return "", MaxUploadError{Max: maxBytes, Observed: written}
    }

    if written == 0 {
        return "", EmptyUploadError{}
    }

    return tempPath, nil
}
```

## Hardening Requirement 3: Storage Quotas

### Goal

Protect the shared package PVC from accidental or malicious disk exhaustion.

### Quota Types

Implement two quota layers:

- **Per-upload quota**: already covered by max upload bytes.
- **Per-package persistent quota**: total bytes under `/var/lib/glazed-docs/packages/{package}`.
- **Global persistent quota warning**: alert when the PVC or package root approaches capacity.

Suggested initial values:

```text
max upload bytes:          64 MiB
max bytes per package:     512 MiB
max versions per package:  configurable, e.g. 25 retained versions
PVC warning threshold:     70 percent
PVC critical threshold:    85 percent
```

The simplest implementation is a pre-publish quota check that walks the package directory and sums file sizes. Because publish volume is low, a filesystem walk is acceptable.

### Pseudocode

```go
type QuotaPolicy struct {
    MaxPackageBytes int64
    MaxVersionsPerPackage int
}

func (p QuotaPolicy) CheckBeforePublish(root, packageName, version string, newFileSize int64) error {
    usage := scanPackageUsage(root, packageName)

    if usage.HasVersion(version) {
        // Immutability policy decides this before quota.
        return nil
    }

    projected := usage.TotalBytes + newFileSize
    if p.MaxPackageBytes > 0 && projected > p.MaxPackageBytes {
        return ErrPackageQuotaExceeded
    }

    if p.MaxVersionsPerPackage > 0 && usage.VersionCount+1 > p.MaxVersionsPerPackage {
        return ErrPackageVersionQuotaExceeded
    }

    return nil
}
```

### API Error

```json
{
  "error": "quota_exceeded",
  "message": "package glazed would exceed configured storage quota"
}
```

Recommended HTTP status: `507 Insufficient Storage` or `409 Conflict`. Prefer `507` for byte quota and `409` for version-count policy.

## Hardening Requirement 4: Immutable Release Versions

### Goal

Make release documentation write-once by default. A tag such as `glazed@v1.3.4` should not silently change after publication.

### Rationale

The publishing model intentionally uses GitHub release tags as docs versions. Release tags are operationally expected to be stable. If a re-run can overwrite a version silently, then readers and operators cannot distinguish:

- an initial publish,
- a CI retry with identical output,
- a corrected publish,
- a malicious overwrite,
- or an accidental upload from the wrong commit.

### Recommended Behavior

Default policy:

- If `{package}/{version}/{package}.db` does not exist, publish succeeds.
- If it exists and the new SHA-256 matches the existing file, return success with `already_published: true`.
- If it exists and SHA-256 differs, reject with `409 Conflict`.
- Allow forced overwrite only through an explicit admin-only path or a temporary CLI flag used during controlled rollback.

### Pseudocode

```go
func publishWithImmutability(req PublishRequest, tmpPath string) (*PublishedPackage, error) {
    target := targetPath(req.PackageName, req.Version)
    newSHA := sha256File(tmpPath)

    if exists(target) {
        oldSHA := sha256File(target)
        if oldSHA == newSHA {
            return existingPackageMetadata(target), nil // idempotent retry
        }
        return nil, ErrVersionAlreadyExists
    }

    return atomicPublish(tmpPath, target)
}
```

### API Error

```json
{
  "error": "version_already_exists",
  "message": "glazed@v1.3.4 is already published with different content"
}
```

Recommended HTTP status: `409 Conflict`.

### File Target

This belongs in or near:

- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go`

Possible API change:

```go
type DirectoryPackageStore struct {
    Root string
    Now func() time.Time
    AllowOverwrite bool
    IdempotentOverwrite bool
}
```

Default should be safe: `AllowOverwrite = false`, `IdempotentOverwrite = true`.

## Hardening Requirement 5: Structured Audit Logs

### Goal

Every publish attempt should produce an operator-readable audit event. The event must be useful for debugging and incident response without leaking secrets.

### Required Fields

For every request:

- `request_id`
- `method`
- `path`
- `remote_addr`
- `client_ip`
- `user_agent`
- `status`
- `duration_ms`

For publish requests:

- `package`
- `version`
- `auth_method`
- `subject`
- `repository`
- `repository_id`
- `workflow_ref`
- `job_workflow_ref`
- `run_id`
- `content_length`
- `stored_bytes`
- `sha256`
- `section_count`
- `slug_count`
- `outcome`
- `error_code`

Never log:

- raw JWTs,
- Vault client tokens,
- GitHub tokens,
- static publisher tokens,
- full Authorization headers.

### Pseudocode

```go
func auditPublish(r *http.Request, event PublishAuditEvent) {
    slog.Info("docs_registry_publish",
        "request_id", event.RequestID,
        "package", event.PackageName,
        "version", event.Version,
        "subject", event.Subject,
        "repository_id", event.RepositoryID,
        "run_id", event.RunID,
        "status", event.Status,
        "outcome", event.Outcome,
        "bytes", event.StoredBytes,
        "sha256", event.SHA256,
        "duration_ms", event.DurationMS,
    )
}
```

### Needed Identity Expansion

`PublisherIdentity` currently exposes only:

```go
type PublisherIdentity struct {
    Subject     string
    PackageName string
    Method      string
}
```

For audit logging, extend it with non-sensitive claims:

```go
type PublisherIdentity struct {
    Subject        string
    PackageName    string
    Method         string
    Repository     string
    RepositoryID   string
    WorkflowRef    string
    JobWorkflowRef string
    RunID          string
}
```

`JWTPublisherAuth.AuthorizePublish` already decodes these claims; it should copy them into `PublisherIdentity`.

## Hardening Requirement 6: Metrics and Alerts

### Goal

Operators should know when the registry is unhealthy before users report broken docs.

### Recommended Metrics

Expose Prometheus-style metrics on a private or same-port endpoint such as `/metrics`.

Counters:

```text
docs_registry_requests_total{method,path,status}
docs_registry_publish_attempts_total{package,outcome}
docs_registry_auth_failures_total{reason}
docs_registry_upload_rejections_total{reason}
docs_registry_quota_rejections_total{package,reason}
```

Histograms:

```text
docs_registry_request_duration_seconds{method,path}
docs_registry_upload_bytes{package}
docs_registry_sqlite_validation_duration_seconds{package}
```

Gauges:

```text
docs_registry_package_root_bytes
docs_registry_package_bytes{package}
docs_registry_package_versions{package}
```

### Initial Alerts

- Auth failures spike over baseline.
- HTTP 5xx responses are non-zero for more than 5 minutes.
- Upload too large rejections spike.
- Package root disk usage exceeds 70 percent warning or 85 percent critical.
- Publish attempts for unknown package names occur.
- Version overwrite conflicts occur.

### Implementation Note

If Prometheus is not already scraping this namespace, start with structured logs and Kubernetes events, then add metrics. The design should not block body limits or immutability on metrics availability.

## Hardening Requirement 7: Negative Auth Proofs

### Goal

Create repeatable evidence that invalid publish paths fail. Successful publishing proves only the happy path. Hardening requires proof that bad paths are rejected.

### Cases to Prove

- Missing Authorization header returns `401 unauthorized`.
- Malformed Bearer token returns `401 unauthorized`.
- Tampered JWT signature returns `401 unauthorized`.
- Expired JWT returns `401 unauthorized`.
- JWT with wrong audience returns `401 unauthorized`.
- JWT with wrong `token_use` returns `403 forbidden`.
- JWT with wrong `package` claim returns `403 forbidden`.
- Valid JWT for `glazed` cannot publish `pinocchio`.
- Upload over `--max-upload-bytes` returns `413 upload_too_large`.
- Invalid SQLite returns `400 invalid_help_db`.
- Duplicate release version with different bytes returns `409 version_already_exists` once immutability is implemented.

### Test Layers

Use three layers of tests:

1. **Unit tests** for auth and handlers using `httptest`.
2. **Integration tests** using a temporary package root and generated SQLite fixtures.
3. **Live proof scripts** for production-safe negative cases that do not mutate storage.

Production-safe live tests should avoid publishing real versions unless using a dedicated non-public test package and a dedicated Vault role. Do not brute-force tokens or generate noisy traffic against production.

### Pseudocode for Handler Test

```go
func TestRegistryRejectsWrongPackageJWT(t *testing.T) {
    auth := fakeAuthReturningIdentityForPackage("glazed")
    store := fakeStore{}
    handler := NewRegistryHandler(auth, store)

    req := httptest.NewRequest(
        "PUT",
        "/v1/packages/pinocchio/versions/v1.0.0/sqlite",
        validSQLiteBody(),
    )
    req.Header.Set("Authorization", "Bearer fake")

    rr := httptest.NewRecorder()
    handler.Handler().ServeHTTP(rr, req)

    require.Equal(t, http.StatusForbidden, rr.Code)
}
```

## Hardening Requirement 8: Rollback Path

### Goal

Operators need a clear way to recover if Vault OIDC publish tokens stop working.

### Current Rollback Capability

`docs-registry` still supports:

```text
--auth-mode static-catalog
--publisher-catalog /path/to/publishers.json
```

This is useful as an emergency rollback, but it must be documented and kept secure.

### Rollback Runbook

1. Confirm failure scope:
   - Is Vault down?
   - Is OIDC discovery down?
   - Is JWKS unavailable?
   - Is only one package failing due to role claims?
2. If read-side docs still work, avoid touching `docs-browser`.
3. Prepare static token catalog as a Kubernetes Secret or sealed secret.
4. Change `docs-registry` args in GitOps:
   - `--auth-mode static-catalog`
   - `--publisher-catalog /etc/docs-yolo/publishers.json`
5. Mount the secret read-only into the registry container.
6. Push GitOps commit.
7. Force Argo CD hard refresh.
8. Watch rollout.
9. Run one controlled publish or health check.
10. Revert to Vault OIDC as soon as Vault is healthy.
11. Rotate any static token used during rollback.

### Important Warning

Static tokens are long-lived credentials. They should not become the normal path again. The rollback path exists to preserve operational continuity, not to weaken the architecture permanently.

## Design Decisions

### Decision 1: Keep Vault OIDC as the primary auth model

Vault OIDC is the right primary model because publish credentials are short-lived, package-scoped, and minted only after GitHub OIDC claim validation. The registry receives only a purpose-built publish JWT, not a Vault client token.

### Decision 2: Add hardening around the existing API instead of redesigning the API

The current API is small and easy to reason about. Hardening should preserve:

```text
PUT /v1/packages/{package}/versions/{version}/sqlite
Authorization: Bearer <token>
```

This avoids changing `docsctl publish`, the reusable workflow, and package release workflows unless new optional flags are needed.

### Decision 3: Make release versions immutable by default

Release-tag documentation should be stable. Idempotent retries with identical content should be safe, but different content for an existing version should require a deliberate operator decision.

### Decision 4: Start with in-process limits, but document when to move limits to infrastructure

The deployment currently runs one registry replica against one PVC. In-process rate and concurrency limits are enough for the immediate architecture. If replicas increase, rate limiting must move to Traefik, Redis, or another shared limiter.

### Decision 5: Prefer structured logs before a full metrics stack if needed

Metrics are valuable, but structured logs can be implemented immediately and consumed by existing Kubernetes logging. The first implementation should not block on observability platform changes.

## Alternatives Considered

### Keep current behavior and only onboard more packages

Rejected. Authentication would still be strong, but public operational failure modes would remain under-controlled. More packages increase the blast radius of mistakes and make auditability more important.

### Hide docs-registry behind a private network only

Not sufficient for GitHub-hosted runners. The release workflow needs to reach the registry from GitHub Actions. A VPN/private runner model could work later, but it is a larger operational change.

### Use long-lived GitHub or registry tokens again

Rejected. Static tokens are harder to scope, rotate, and audit. They should remain a rollback mechanism, not the primary path.

### Push directly to object storage from CI

Rejected for now. Direct object storage writes would move policy enforcement into bucket IAM and CI scripts. The registry centralizes validation, catalog updates, package/version policy, and audit logs.

### Accept overwrites forever

Rejected. It makes release documentation mutable and weakens incident response. Idempotent retries should be supported; silent different-content overwrites should not.

## Implementation Plan

### Phase 1: Documentation and Ticket Setup

Status: this document.

Tasks:

- Create ticket `DOCSCTL-REGISTRY-HARDENING`.
- Record system map, current code references, hardening requirements, and implementation order.
- Upload this guide to reMarkable for review.

### Phase 2: Handler-Level Limits and Audit Foundation

Implement:

- request ID middleware,
- structured access logs,
- explicit production `--max-upload-bytes`,
- publish concurrency limit,
- basic per-IP rate limiter,
- tests for rate limiting and body limits.

Validation commands:

```bash
go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl
```

### Phase 3: Immutability and Quotas

Implement:

- idempotent same-SHA re-publish behavior,
- reject different-content overwrite with `409`,
- per-package byte quota,
- optional max versions per package,
- tests using temporary package roots.

Validation commands:

```bash
go test ./pkg/help/publish
```

### Phase 4: Identity Enrichment and Audit Events

Implement:

- copy JWT non-sensitive claims into `PublisherIdentity`,
- log publish attempt outcomes,
- ensure raw tokens are never logged,
- tests that audit events contain expected fields and exclude Authorization values.

### Phase 5: Metrics and Alerting

Implement or configure:

- Prometheus metrics endpoint if available,
- Kubernetes/Prometheus alerts if available,
- log-based alert alternatives if metrics are not ready.

### Phase 6: Negative Proof Suite

Implement:

- local test suite for invalid auth cases,
- optional production-safe script under the ticket `scripts/` directory,
- evidence document in `sources/` with exact commands and redacted output.

### Phase 7: Rollout and Production Verification

Rollout sequence:

```bash
# in glazed repo
make test
cd web && pnpm test && pnpm build && pnpm build:ssr

git push origin HEAD:main
# wait for container images, e.g. ghcr.io/go-go-golems/glazed:sha-<commit>

# in k3s repo
# update docs-yolo deployment image tags and args
git push origin main
source /home/manuel/code/wesen/2026-03-27--hetzner-k3s/.envrc
kubectl -n argocd annotate application docs-yolo argocd.argoproj.io/refresh=hard --overwrite
kubectl -n docs-yolo rollout status deploy/docs-yolo --timeout=300s
```

Production checks:

```bash
curl -sS https://docs-registry.yolo.scapegoat.dev/healthz
curl -sS https://docs.yolo.scapegoat.dev/api/health
curl -sS https://docs.yolo.scapegoat.dev/api/packages | jq .
```

## Intern Implementation Notes

Start by reading the code in this order:

1. `cmd/docs-registry/main.go` to understand flags and server construction.
2. `pkg/help/publish/registry.go` to understand HTTP routing and upload flow.
3. `pkg/help/publish/jwt_auth.go` to understand JWT claim validation.
4. `pkg/help/publish/sqlite_validator.go` to understand SQLite safety checks.
5. `pkg/help/publish/directory_store.go` to understand storage and catalog updates.
6. `infra-tooling/.github/workflows/publish-docsctl.yml` to understand the CI caller contract.
7. k3s `docs-yolo` deployment and ingress files to understand runtime topology.
8. Terraform Vault GitHub Actions file to understand why only release-tag workflows can mint publish JWTs.

When implementing, keep these invariants:

- Never log raw tokens.
- Never trust package or version path segments without `ValidatePackageVersion`.
- Never write outside `package-root`.
- Never replace an existing release DB with different bytes unless an explicit admin override exists.
- Always clean up temporary upload files.
- Keep the public API stable unless the ticket explicitly changes it.
- Add tests before production rollout.

## Suggested Code Organization

A clean implementation can add these files:

```text
pkg/help/publish/registry_middleware.go
pkg/help/publish/rate_limit.go
pkg/help/publish/quota.go
pkg/help/publish/audit.go
pkg/help/publish/registry_hardening_test.go
pkg/help/publish/quota_test.go
pkg/help/publish/immutability_test.go
```

Possible struct shape:

```go
type RegistryOptions struct {
    MaxUploadBytes       int64
    TempDir              string
    MaxConcurrentUploads int
    RateLimit            RateLimitOptions
    Quota                QuotaOptions
    Audit                AuditOptions
}

type QuotaOptions struct {
    MaxPackageBytes      int64
    MaxVersionsPerPackage int
    AllowOverwrite       bool
    IdempotentOverwrite  bool
}
```

Avoid a giant `RegistryHandler` if possible. Keep the handler readable and move policy code to small testable helpers.

## Review Checklist

Before merging implementation work, reviewers should check:

- Are all new defaults safe?
- Are production flags explicit in GitOps?
- Can an unauthenticated client consume large disk space?
- Can a valid publisher overwrite a released version with different content?
- Are JWTs and Authorization headers excluded from logs?
- Do errors return stable JSON codes?
- Are tests covering both success and failure paths?
- Does rollback to static catalog remain possible but not default?
- Does the rollout preserve existing published docs?

## Open Questions

- Should per-package quota be enforced by byte total only, version count only, or both?
- Should forced overwrite be implemented at all, or should manual filesystem intervention be the only override?
- Where should long-term registry metrics be scraped from in the current k3s observability stack?
- Should `GET /v1/packages` remain public on the registry host, or should public listing happen only through `docs.yolo.scapegoat.dev/api/packages`?
- Should production negative tests use a dedicated `docsctl-test` package and Vault role?

## References

Code and configuration references:

- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/jwt_auth.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/sqlite_validator.go`
- `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml`
- `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml`

Related tickets:

- `DOCSCTL-CICD-DEPLOY`: reusable GitHub CI/CD action for docs publishing.
- `DOCSCTL-VAULT-OIDC-JWT`: Vault Identity/OIDC publish JWT design and implementation.
- `DOCSCTL-SSR-K3S`: SSR sidecar deployment on k3s.
- `DOCSCTL-REACT-SSR-HTML`: full React-rendered SSR HTML and metadata/font polish.
