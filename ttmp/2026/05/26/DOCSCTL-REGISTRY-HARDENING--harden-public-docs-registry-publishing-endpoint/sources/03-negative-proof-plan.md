---
Title: Negative Proof Plan
Type: source
Ticket: DOCSCTL-REGISTRY-HARDENING
Topics:
  - docs-yolo
  - backend
  - oidc
  - cicd
Status: draft
---

# Negative Proof Plan

This note records the Phase 6 negative proof strategy for the public `docs-registry` endpoint.

## Local automated proof cases

The Go tests cover negative cases that do not require production credentials:

- unauthenticated upload returns `401 unauthorized`;
- wrong package for a valid publisher token returns `403 forbidden`;
- invalid SQLite returns `400 invalid_help_db`;
- duplicate version with different content returns `409 version_already_exists`;
- package byte quota rejection returns `507 quota_exceeded`;
- JWT package mismatch returns `ErrForbidden`;
- JWT wrong audience returns `ErrUnauthorized`;
- JWT wrong issuer returns `ErrUnauthorized`;
- JWT expired token returns `ErrUnauthorized`;
- JWT tampered token returns `ErrUnauthorized`;
- JWT missing `token_use` returns `ErrForbidden`;
- JWT missing `package` returns `ErrForbidden`.

## Production-safe proof cases

The script `scripts/01-production-safe-negative-probes.sh` performs checks that are safe to run against the public production registry without secrets:

- `GET /healthz` must return 200;
- unauthenticated `PUT /v1/packages/{package}/versions/{fresh-version}/sqlite` must return 401 before storage;
- `GET /metrics` may return 200 if public scraping is allowed, or another status if ingress rules later restrict it.

The unauthenticated upload sends only a tiny invalid body and no Authorization header. Because auth happens before upload validation/storage, this request should never publish data.

## Cases that require GitHub/Vault controlled proof

The following cases require a controlled GitHub Actions run or a deliberately constrained Vault test role. They should not be attempted by hand with long-lived credentials:

- wrong GitHub `repository_id` bound claim;
- wrong `workflow_ref`;
- wrong `job_workflow_ref`;
- wrong release tag/ref shape;
- wrong GitHub event name;
- valid JWT for one package attempting to publish another package in production;
- duplicate-version different-content rejection using a real publish JWT.

For these, prefer a temporary test workflow in a disposable branch/tag or a Vault role that can mint a token for a non-production package name. Capture only status codes, stable error codes, request IDs, and audit/metric evidence. Never paste raw GitHub OIDC tokens, Vault client tokens, or Vault Identity tokens into ticket docs.

## Evidence to capture

For each negative proof run, capture:

- command name and sanitized arguments;
- HTTP status;
- stable registry error code;
- request ID;
- relevant `docs registry publish` audit log line with token values absent;
- relevant metrics counter before/after if scraping is enabled.
