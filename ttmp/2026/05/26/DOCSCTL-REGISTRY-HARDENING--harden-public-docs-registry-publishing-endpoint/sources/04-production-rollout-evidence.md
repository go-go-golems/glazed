---
Title: Production Rollout Evidence
Type: source
Ticket: DOCSCTL-REGISTRY-HARDENING
Topics:
  - docs-yolo
  - backend
  - kubernetes
  - gitops
Status: final
---

# Production Rollout Evidence

This source records the Phase 7 rollout of the Phase 4-6 registry hardening work.

## Glazed image build

The Glazed hardening commits were pushed to `go-go-golems/glazed` `main`.

```text
Pushed range: 94af442..312fa79
Head commit: 312fa792e29b3ed5aa94b9615990d72c541ed41b
```

The container workflow succeeded:

```text
Workflow: Container image
Run ID: 26484320165
Job ID: 77988328251
Status: success
Duration: 4m58s
Images:
  ghcr.io/go-go-golems/glazed:sha-312fa79
  ghcr.io/go-go-golems/glazed-ssr:sha-312fa79
```

## k3s GitOps rollout

The k3s GitOps deployment was updated and pushed:

```text
Repo: /home/manuel/code/wesen/2026-03-27--hetzner-k3s
Commit: 05deb9c
Message: docs-yolo: deploy registry audit metrics proofs
```

Deployment images after rollout:

```text
docs-browser=ghcr.io/go-go-golems/glazed:sha-312fa79
docs-registry=ghcr.io/go-go-golems/glazed:sha-312fa79
docs-ssr=ghcr.io/go-go-golems/glazed-ssr:sha-312fa79
```

Argo CD status after rollout:

```text
Synced Healthy Succeeded
```

Kubernetes rollout status:

```text
deployment "docs-yolo" successfully rolled out
```

## Health validation

Registry health:

```bash
curl -fsS https://docs-registry.yolo.scapegoat.dev/healthz
```

```json
{"ok":true}
```

Docs browser health:

```bash
curl -fsS https://docs.yolo.scapegoat.dev/api/health
```

```json
{"ok":true,"sections":333}
```

## Metrics validation

Command:

```bash
curl -fsS https://docs-registry.yolo.scapegoat.dev/metrics | grep -E 'docs_registry_(http_requests|publish_attempts)'
```

Observed output after safe negative probes:

```text
# HELP docs_registry_http_requests_total Total docs-registry HTTP requests by route class, method, and status.
# TYPE docs_registry_http_requests_total counter
docs_registry_http_requests_total{route_class="health",method="GET",status="200"} 12
docs_registry_http_requests_total{route_class="other",method="GET",status="200"} 3
docs_registry_http_requests_total{route_class="publish",method="PUT",status="401"} 2
# HELP docs_registry_publish_attempts_total Total docs-registry publish attempts by package, outcome, and stable error code.
# TYPE docs_registry_publish_attempts_total counter
docs_registry_publish_attempts_total{package="glazed",outcome="rejected",error_code="unauthorized"} 2
```

## Production-safe negative probe

After fixing the script cleanup trap, the production-safe probe succeeded:

```bash
REGISTRY_URL=https://docs-registry.yolo.scapegoat.dev \
  ttmp/2026/05/26/DOCSCTL-REGISTRY-HARDENING--harden-public-docs-registry-publishing-endpoint/scripts/01-production-safe-negative-probes.sh
```

Output:

```text
OK: GET https://docs-registry.yolo.scapegoat.dev/healthz -> 200
OK: PUT https://docs-registry.yolo.scapegoat.dev/v1/packages/glazed/versions/negative-proof-20260527T010911Z/sqlite -> 401
OK: GET https://docs-registry.yolo.scapegoat.dev/metrics -> 200
```

The first script run also proved the endpoint behavior but exposed a shell cleanup bug:

```text
ttmp/.../scripts/01-production-safe-negative-probes.sh: line 1: tmp_body: unbound variable
```

Fix commit:

```text
Glazed 6a99c4c  DOCSCTL-REGISTRY-HARDENING: fix negative probe cleanup
```

## Audit log validation

Registry logs showed publish-specific audit events for the unauthenticated negative probes. No bearer token material was present.

```text
2026/05/27 01:08:49 INFO docs registry publish request_id=cb90f9bd3c2b12d99d4419805cd552c3 package=glazed version=negative-proof-20260527T010848Z status=401 outcome=rejected error_code=unauthorized duration_ms=0 content_length=22 upload_bytes=0 section_count=0 slug_count=0 sha256="" client_ip=10.42.0.1 remote_addr=10.42.0.205:37168 user_agent=curl/8.5.0
2026/05/27 01:09:12 INFO docs registry publish request_id=282f45c86a67dfca06b02e28e33f4e3b package=glazed version=negative-proof-20260527T010911Z status=401 outcome=rejected error_code=unauthorized duration_ms=0 content_length=22 upload_bytes=0 section_count=0 slug_count=0 sha256="" client_ip=10.42.0.1 remote_addr=10.42.0.205:37168 user_agent=curl/8.5.0
```

## Notes

- `/metrics` is currently reachable through the public registry ingress. This is acceptable for the immediate proof but should be revisited if metrics should be restricted to in-cluster scraping.
- The deployed runtime image is `sha-312fa79`, which includes the Phase 4-6 registry code. The later `6a99c4c` commit fixes only the ticket probe script cleanup and does not change registry runtime code.
