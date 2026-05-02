---
Title: Production Deployment Guide
Ticket: DEPLOY-GLAZE-DOCS
Status: ""
Topics: []
DocType: design-doc
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml
      Note: Argo CD application for the docs deployment
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml
      Note: Kubernetes workload running glaze serve
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml
      Note: Traefik and cert-manager public routing
    - Path: ../../../../../../../../../../code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf
      Note: Wildcard docs DNS record
    - Path: .github/workflows/container.yml
      Note: GHCR image publishing workflow
    - Path: Dockerfile
      Note: Container build for glaze serve
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T10:57:55.276976486-04:00
WhatFor: ""
WhenToUse: ""
---





# Deploying `glaze serve` to `glaze.docs.scapegoat.dev`

## Executive summary

This document explains how the Glazed documentation browser is deployed to production at:

- `https://glaze.docs.scapegoat.dev`
- with DNS reserved for future sites under `*.docs.scapegoat.dev`

The system has three repositories/areas that work together:

1. **Glazed application repo**: builds a container image containing `glaze serve` and the embedded React help browser.
2. **Hetzner k3s GitOps repo**: tells Argo CD how to run that image in Kubernetes and expose it through Traefik.
3. **Terraform DNS repo**: manages DigitalOcean DNS records for `scapegoat.dev`, including the wildcard `*.docs.scapegoat.dev`.

The short path from source to production is:

```text
Developer pushes Glazed code
        |
        v
GitHub Actions builds ghcr.io/go-go-golems/glazed:main
        |
        v
Argo CD reconciles gitops/kustomize/glaze-docs
        |
        v
Kubernetes runs Deployment -> Service -> Traefik Ingress
        |
        v
DigitalOcean DNS maps *.docs.scapegoat.dev -> 91.98.46.169
        |
        v
Browser reaches https://glaze.docs.scapegoat.dev
```

## Problem statement

`glaze serve` is a local command that serves the Glazed help system over HTTP. We want it to become a production documentation site.

The deployment must satisfy these constraints:

- The public hostname must be `glaze.docs.scapegoat.dev`.
- A wildcard DNS record for `*.docs.scapegoat.dev` must exist because later documentation sites will live under the same `docs.scapegoat.dev` subtree.
- The deployment should follow the existing GitOps pattern in `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`.
- DNS should be managed through `/home/manuel/code/wesen/terraform`.
- The work should be documented in a docmgr ticket for review and onboarding.

## Current implementation files

### Glazed application repository

Base path:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed`

Files added or relevant:

- `Dockerfile`
  - Builds the embedded web UI with `go generate ./pkg/web`.
  - Builds `./cmd/glaze`.
  - Runs `glaze serve --address :8088` by default.
  - Uses `CGO_ENABLED=1` because `go-sqlite3` is required by the help store.
- `.github/workflows/container.yml`
  - Builds and pushes `ghcr.io/go-go-golems/glazed`.
  - Publishes tags for branch, version tag, and commit SHA.
- `cmd/glaze/main.go`
  - Wires the `serve` subcommand into the root command.
- `pkg/help/server/serve.go`
  - Defines the `serve` command and its flags.
- `pkg/help/server/handlers.go`
  - Defines the HTTP API handler.
- `pkg/web/static.go`
  - Serves the embedded React SPA.
- `cmd/build-web/main.go`
  - Builds the React frontend and copies it into `pkg/web/dist` for embedding.

### k3s GitOps repository

Base path:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`

Files added:

- `gitops/applications/glaze-docs.yaml`
  - Argo CD `Application` object.
  - Points Argo CD at `gitops/kustomize/glaze-docs`.
  - Deploys into namespace `glaze-docs`.
- `gitops/kustomize/glaze-docs/kustomization.yaml`
  - Kustomize entry point.
- `gitops/kustomize/glaze-docs/deployment.yaml`
  - Runs `ghcr.io/go-go-golems/glazed:main`.
  - Starts `glaze serve --address :8088`.
  - Defines readiness and liveness probes on `/api/health`.
- `gitops/kustomize/glaze-docs/service.yaml`
  - Exposes container port `8088` as service port `80` inside the cluster.
- `gitops/kustomize/glaze-docs/ingress.yaml`
  - Routes `glaze.docs.scapegoat.dev` to the service through Traefik.
  - Requests a concrete HTTP-01-compatible TLS certificate for `glaze.docs.scapegoat.dev` via cert-manager.
  - Does not request wildcard TLS yet; future DNS-01 work is tracked in GitHub issue https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65.

### Terraform DNS repository

Base path:

- `/home/manuel/code/wesen/terraform`

File changed:

- `dns/zones/scapegoat-dev/envs/prod/main.tf`
  - Adds `wildcard_docs_a`:

```hcl
wildcard_docs_a = {
  type  = "A"
  name  = "*.docs"
  value = "91.98.46.169"
  ttl   = 3600
}
```

This creates the public DNS record:

```text
*.docs.scapegoat.dev. 3600 IN A 91.98.46.169
```

## Architecture

### Runtime request flow

```text
User browser
  |
  | DNS query: glaze.docs.scapegoat.dev
  v
DigitalOcean DNS
  |
  | A answer via wildcard *.docs.scapegoat.dev = 91.98.46.169
  v
Hetzner k3s node public IP
  |
  | HTTPS request with Host: glaze.docs.scapegoat.dev
  v
Traefik ingress controller
  |
  | Kubernetes Ingress rule host=glaze.docs.scapegoat.dev path=/
  v
Service glaze-docs in namespace glaze-docs
  |
  | targetPort=http -> containerPort 8088
  v
Pod running ghcr.io/go-go-golems/glazed:main
  |
  | /api/* served by pkg/help/server
  | /* served by pkg/web embedded SPA
  v
Glazed help browser
```

### Kubernetes objects

```text
Argo CD Application (argocd namespace)
  owns/reconciles
    Kustomize app directory
      creates
        Deployment/glaze-docs
        Service/glaze-docs
        Ingress/glaze-docs
```

The objects have separate responsibilities:

- **Deployment**: desired running pods.
- **Pod/container**: actual process: `glaze serve --address :8088`.
- **Service**: stable in-cluster network endpoint for matching pods.
- **Ingress**: public HTTP(S) routing rule consumed by Traefik.
- **Argo CD Application**: controller instruction telling Argo what Git path to reconcile.

### DNS and TLS

DNS and TLS are related but separate:

- DNS answers: "which IP should this hostname connect to?"
- TLS answers: "is this server allowed to present a certificate for this hostname?"
- Ingress answers: "which Kubernetes service should receive this HTTP request?"

The DNS wildcard is broad:

```text
*.docs.scapegoat.dev -> 91.98.46.169
```

The current Ingress routing rule is narrow:

```yaml
rules:
  - host: glaze.docs.scapegoat.dev
```

That means future hosts such as `pinocchio.docs.scapegoat.dev` will resolve to the cluster, but Traefik will not route them to this service until a matching Ingress rule is added.

## API reference for `glaze serve`

### Command

Defined in:

- `pkg/help/server/serve.go`

Default command inside the container:

```bash
glaze serve --address :8088
```

Important flags:

- `--address`: bind address; default is `:8088`.
- `--from-json`: load JSON help exports.
- `--from-sqlite`: load SQLite help export databases.
- `--from-glazed-cmd`: load help by running another Glazed binary's `help export --output json`.
- `--with-embedded`: when external sources are used, keep embedded Glazed docs instead of replacing them.

### HTTP endpoints

The server exposes:

- `GET /api/health`
  - Used by Kubernetes readiness/liveness probes.
  - Example local validation response:

```json
{"ok":true,"sections":72}
```

- `GET /api/*`
  - Help-browser API routes for listing and retrieving sections.
- `GET /*`
  - Embedded React SPA fallback.

## Implementation guide for a new intern

### 1. Understand the deployment unit

The thing being deployed is not "a static website". It is a Go HTTP server with an embedded React frontend.

Conceptually:

```pseudocode
main():
    helpSystem = load embedded Glazed documentation
    spaHandler = load embedded web/dist files
    serveCommand = new ServeCommand(helpSystem, spaHandler)
    rootCommand.add(serveCommand)
    rootCommand.execute()
```

When the container starts:

```pseudocode
container ENTRYPOINT = /usr/local/bin/glaze
container CMD = ["serve", "--address", ":8088"]

process starts
    create help store
    load embedded docs
    start HTTP server on :8088
```

### 2. Build the image

The Dockerfile does four important things:

1. Copies the repository into a Go builder image.
2. Installs Node/npm/corepack so `go generate ./pkg/web` can build the Vite frontend if Dagger is unavailable.
3. Compiles `cmd/glaze` with `CGO_ENABLED=1`.
4. Copies the binary into a small Debian runtime image.

Why CGO matters:

- The project imports `github.com/mattn/go-sqlite3` indirectly through the help store.
- `go-sqlite3` requires CGO.
- A static `CGO_ENABLED=0` build compiles, but fails at runtime with:

```text
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

### 3. Publish the image

The workflow `.github/workflows/container.yml` publishes images to GitHub Container Registry.

Expected tags include:

- `ghcr.io/go-go-golems/glazed:main`
- `ghcr.io/go-go-golems/glazed:sha-<shortsha>`
- `ghcr.io/go-go-golems/glazed:<version-tag>` for `v*` tags

Intern rule of thumb:

- Use `:main` for fast iteration only if that is the established convention.
- Prefer `sha-<shortsha>` in production when you want a deterministic rollback target.

### 4. Deploy with Argo CD

The Argo CD Application is the bridge from Git to Kubernetes:

```yaml
source:
  repoURL: https://github.com/wesen/2026-03-27--hetzner-k3s.git
  targetRevision: main
  path: gitops/kustomize/glaze-docs
```

Pseudocode for the controller loop:

```pseudocode
loop forever:
    desired = render_kustomize(repoURL, targetRevision, path)
    live = read_kubernetes_objects(namespace="glaze-docs")
    diff = compare(desired, live)
    if diff exists and automated selfHeal enabled:
        apply(desired)
    if desired removed and prune enabled:
        delete_orphaned_live_objects()
```

The `syncOptions` include:

- `CreateNamespace=true`: Argo CD can create the `glaze-docs` namespace.
- `ServerSideApply=true`: Kubernetes server-side apply manages field ownership.

### 5. Route traffic through Kubernetes

The routing chain is:

```text
Ingress host rule -> Service port 80 -> Pod container port 8088
```

The Deployment labels must match the Service selector:

```yaml
selector:
  app.kubernetes.io/name: glaze-docs
  app.kubernetes.io/component: web
```

If labels do not match, the Service will have no endpoints and Ingress traffic will fail.

### 6. Manage DNS with Terraform

DNS is managed in:

```bash
/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod
```

The normal workflow is:

```bash
cd /home/manuel/code/wesen/terraform
direnv allow
terraform -chdir=dns/zones/scapegoat-dev/envs/prod init
terraform -chdir=dns/zones/scapegoat-dev/envs/prod plan
terraform -chdir=dns/zones/scapegoat-dev/envs/prod apply
```

The DigitalOcean token comes from the ignored local environment file according to `dns/README.md`:

```bash
export DIGITALOCEAN_TOKEN="dop_v1_..."
```

### 7. Validate before and after rollout

Local validation commands used during implementation:

```bash
cd /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed
go test ./pkg/help/server ./pkg/web

docker build -t glazed-docs:test .
docker run --rm --name glazed-docs-test -p 18088:8088 glazed-docs:test
curl -fsS http://127.0.0.1:18088/api/health
```

GitOps validation:

```bash
kubectl kustomize /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs >/tmp/glaze-docs.yaml
```

DNS formatting validation:

```bash
terraform -chdir=/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod fmt -check
```

Post-rollout checks:

```bash
# DNS resolves to the k3s public IP
dig +short glaze.docs.scapegoat.dev

# Certificate and HTTP route work
curl -fsS https://glaze.docs.scapegoat.dev/api/health

# Kubernetes health
kubectl -n glaze-docs get deploy,svc,ingress,pods
kubectl -n glaze-docs logs deploy/glaze-docs --tail=100

# Argo status
kubectl -n argocd get application glaze-docs
```

Expected health response:

```json
{"ok":true,"sections":72}
```

## Design decisions

### Decision: Use a containerized long-running `glaze serve`

Rationale:

- Matches existing cluster patterns.
- Keeps dynamic API and embedded SPA together.
- Allows future `--from-*` extensions if the docs deployment later aggregates several tools.

Alternative considered:

- Export a static site and serve it with nginx.

Why not now:

- The requested deployment is specifically `glaze serve`.
- The server mode is already implemented and has health checks.

### Decision: Use `*.docs.scapegoat.dev` DNS now, but only route `glaze.docs.scapegoat.dev`

Rationale:

- DNS wildcard reserves the namespace for future docs sites.
- Narrow Ingress routing prevents accidental routing of all future docs hosts to Glazed.

Alternative considered:

- Add an Ingress wildcard host rule immediately.

Why not now:

- Kubernetes Ingress wildcard host rules can be useful, but later docs hosts may need different services.
- The current requirement names `glaze.docs.scapegoat.dev` as the concrete first site.

### Decision: Keep image tag as `ghcr.io/go-go-golems/glazed:main` for initial wiring

Rationale:

- It matches branch-based publishing and is simple for first deployment.

Risk:

- `:main` is mutable.

Recommended production hardening:

- After the first image is built, update Deployment to `ghcr.io/go-go-golems/glazed:sha-<shortsha>`.

### Decision: Use wildcard DNS now, but concrete HTTP-01 TLS for Glaze

Rationale:

- The user explicitly wants wildcard `*.docs.scapegoat.dev` DNS for later pages.
- The existing `letsencrypt-prod` ClusterIssuer is HTTP-01-only, matching the established `*.yolo.scapegoat.dev` pattern: wildcard DNS points at the cluster, while each concrete app hostname gets its own HTTP-01 certificate.
- The Glaze Ingress therefore requests only `glaze.docs.scapegoat.dev` in `spec.tls.hosts` and stores it in `glaze-docs-tls`.

Future work:

- GitHub issue https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65 tracks adding DigitalOcean DNS-01 support to `letsencrypt-prod` so wildcard docs certificates can be issued later.

## Operations runbook

### First deployment checklist

1. Merge/push Glazed image changes.
2. Confirm GHCR image exists:

```bash
docker pull ghcr.io/go-go-golems/glazed:main
```

3. Merge/push DNS Terraform change.
4. Run Terraform plan/apply in the DNS repo.
5. Merge/push GitOps app changes.
6. Apply the Argo CD `Application` once if it is not managed by an app-of-apps:

```bash
kubectl apply -f /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml
```

7. Watch rollout:

```bash
kubectl -n argocd get application glaze-docs -w
kubectl -n glaze-docs rollout status deploy/glaze-docs
```

8. Verify public endpoint:

```bash
curl -fsS https://glaze.docs.scapegoat.dev/api/health
```

### Debugging guide

#### Symptom: DNS does not resolve

Check:

```bash
dig +short glaze.docs.scapegoat.dev
```

Likely causes:

- Terraform apply not run.
- DigitalOcean provider token missing.
- Record not yet propagated.

#### Symptom: Ingress returns 404

Check:

```bash
kubectl -n glaze-docs get ingress glaze-docs -o yaml
kubectl -n kube-system logs deploy/traefik --tail=100
```

Likely causes:

- Host rule does not match request Host header.
- Ingress class is wrong.
- Traefik has not reconciled the new Ingress.

#### Symptom: Service has no endpoints

Check:

```bash
kubectl -n glaze-docs get endpoints glaze-docs -o yaml
kubectl -n glaze-docs get pods --show-labels
```

Likely causes:

- Deployment labels do not match Service selector.
- Pods are not Ready.

#### Symptom: Pod crashes with SQLite CGO error

Error:

```text
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work
```

Fix:

- Rebuild image with `CGO_ENABLED=1`.
- Use a runtime image compatible with dynamically linked CGO binaries, such as Debian slim.

#### Symptom: Certificate does not issue

Check:

```bash
kubectl -n glaze-docs get certificate,certificaterequest,order,challenge
kubectl -n cert-manager logs deploy/cert-manager --tail=200
```

Likely causes:

- DNS has not propagated to the k3s public IP.
- The `letsencrypt-prod` HTTP-01 solver cannot reach the temporary challenge route through Traefik.
- The Ingress references the wrong ClusterIssuer name.

Current wildcard note:

- The Glaze Ingress intentionally does not request `*.docs.scapegoat.dev` TLS yet. Wildcard TLS requires DNS-01 and is tracked for later in https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65.

## Open questions and review notes

- Should the production Deployment pin `sha-<commit>` instead of `main` before applying?
- Future DNS-01 support for wildcard docs certificates is tracked in https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65.
- Should future docs hosts be routed by separate Ingress objects, or should this deployment become a multi-host documentation aggregator?
- Should `glaze.docs.scapegoat.dev` eventually aggregate Pinocchio, Sqleton, and other Glazed-based tools via `--from-glazed-cmd` or pre-exported JSON/SQLite snapshots?

## Related files

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/Dockerfile`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/.github/workflows/container.yml`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/web/static.go`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/service.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml`
- `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`
