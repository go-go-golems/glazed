---
Title: Diary
Ticket: DEPLOY-GLAZE-DOCS
Status: ""
Topics: []
DocType: reference
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml
      Note: Step 3 changed TLS to concrete HTTP-01 hostname
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/kustomization.yaml
      Note: Diary records Kustomize validation
    - Path: ../../../../../../../../../../code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf
      Note: Diary records DNS change
    - Path: Dockerfile
      Note: Diary records the CGO SQLite container fix
    - Path: ttmp/2026/05/02/DEPLOY-GLAZE-DOCS--deploy-glaze-serve-to-docs-scapegoat-dev/reference/02-future-dns-01-wildcard-tls-github-issue.md
      Note: Step 3 stores future DNS-01 issue source
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T10:57:55.499417474-04:00
WhatFor: ""
WhenToUse: ""
---






# Diary

## Goal

Record the productionization work for exposing `glaze serve` at `https://glaze.docs.scapegoat.dev` while reserving `*.docs.scapegoat.dev` for later documentation sites.

## Step 1: Map existing infrastructure and create the ticket

I first inspected the current workspace, the Hetzner k3s GitOps repository, and the Terraform DNS repository to understand how applications and DNS are normally declared. The important discovery was that the k3s repo uses per-application Argo CD `Application` manifests in `gitops/applications/` plus Kustomize application directories in `gitops/kustomize/`, while the Terraform DNS repo keeps `scapegoat.dev` records in one explicit `local.base_records` map.

I then created this docmgr ticket and added a diary plus a design/implementation guide so the deployment can be reviewed and continued without relying on transcript memory.

### Prompt Context

**User prompt (verbatim):** "I want to deploy glaze serve as glaze.docs.scapegoat.dev . I do want a wildcard *.docs.scapegoat.dev because later we will be serving more pages from that deployment.

Look at /home/manuel/code/wesen/2026-03-27--hetzner-k3s which is where we store our gitops argocd setup, and /home/manuel/code/wesen/terraform which is where we manage our dns iirc.

Create anew docmgr to deploy glazed into production, keep a diary as you work,  Create a new docmgr to write the tool, and  reate a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                             
  system needed to understand what it is, with prose paragraphs and bullet                                                                                                                                   
  point sand pseudocode and diagrams and api references and file                                                                                                                                             
    references.                                                                                                                                                                                              
    It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Deploy the Glazed help browser to the k3s cluster, add wildcard DNS for future docs hosts, maintain a docmgr ticket with a diary, and publish a detailed guide to reMarkable.

**Inferred user intent:** Make Glazed documentation production-accessible and leave enough written context that a new intern can safely understand and operate the GitOps/DNS workflow.

### What I did
- Read the docmgr, diary, and reMarkable upload skill instructions.
- Inspected `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications` and `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize`.
- Inspected `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`.
- Created ticket `DEPLOY-GLAZE-DOCS` in `glazed/ttmp/2026/05/02/DEPLOY-GLAZE-DOCS--deploy-glaze-serve-to-docs-scapegoat-dev/`.

### Why
- The deployment spans three concerns: application image, Kubernetes routing, and public DNS.
- The ticket keeps those concerns linked and reviewable.

### What worked
- Existing infrastructure conventions were clear enough to follow: Kustomize for Kubernetes objects, Argo CD `Application` for reconciliation, Terraform for DNS.

### What didn't work
- N/A in this step.

### What I learned
- The live cluster ingress IP is already represented in DNS by `k3s` and `*.yolo` records pointing to `91.98.46.169`.
- The deployment should use the same IP for `*.docs.scapegoat.dev`.

### What was tricky to build
- The prompt contains typos and repeated phrases, but the intent is clear: create ticket docs, implement deployment work, and upload the resulting guide.

### What warrants a second pair of eyes
- Confirm whether `*.docs.scapegoat.dev` should be a single wildcard A record or whether an explicit `docs.scapegoat.dev` apex-under-zone record is also desired later.

### What should be done in the future
- Apply the Argo CD application manifest after the GitOps repo is pushed.
- Apply the Terraform DNS plan after reviewing drift.

### Code review instructions
- Start with the guide in `design-doc/01-production-deployment-guide.md`.
- Review related infra files listed in this ticket.

### Technical details
- GitOps repo: `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`
- DNS repo: `/home/manuel/code/wesen/terraform`
- Glazed repo/worktree: `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed`

## Step 2: Add container build, Kubernetes GitOps manifests, and DNS

I added a production container definition for the `glaze` binary and a GitHub Actions workflow that publishes `ghcr.io/go-go-golems/glazed` tags on pushes to `main` and version tags. The Dockerfile builds the embedded React help browser via `go generate ./pkg/web` and then compiles `./cmd/glaze` with CGO enabled because the help store depends on `go-sqlite3` even for the in-memory store.

In the k3s GitOps repo I added a `glaze-docs` Kustomize application with a Deployment, Service, and Traefik Ingress. The Ingress routes `glaze.docs.scapegoat.dev` to the service and requests a TLS certificate that includes both `glaze.docs.scapegoat.dev` and `*.docs.scapegoat.dev`. In the Terraform DNS repo I added the wildcard A record `*.docs -> 91.98.46.169`.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the production deployment files and document how they fit together.

**Inferred user intent:** Have concrete, reviewable repo changes rather than only a written plan.

### What I did
- Added `glazed/Dockerfile`.
- Added `glazed/.github/workflows/container.yml`.
- Added `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/{kustomization.yaml,deployment.yaml,service.yaml,ingress.yaml}`.
- Added `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml`.
- Edited `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf` to add `wildcard_docs_a`.

### Why
- The cluster can only run `glaze serve` reliably if the image contains the embedded web UI and a CGO-capable SQLite build.
- Argo CD needs an `Application` object to reconcile the Kustomize directory into the `glaze-docs` namespace.
- Public clients need `*.docs.scapegoat.dev` to resolve to the k3s ingress address before Traefik can route HTTP(S).

### What worked
- `kubectl kustomize` rendered the new app after switching away from deprecated `commonLabels`.
- `terraform fmt -check` passed for the DNS environment.
- The package tests for `pkg/help/server` and `pkg/web` passed.
- The Docker image built and `curl http://127.0.0.1:18088/api/health` returned `{"ok":true,"sections":72}` from a local container.

### What didn't work
- First Docker runtime test failed with:
  - `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub`
- Fix: rebuild with `CGO_ENABLED=1` and use a Debian runtime image instead of a static distroless image.
- Dagger inside Docker was unavailable during `go generate ./pkg/web`:
  - `start engine: driver for scheme "image" was not available`
- This was non-blocking because `cmd/build-web` fell back to local pnpm successfully.

### What I learned
- `glaze serve` uses SQLite-backed help storage even when no on-disk database is provided.
- The image must not be a fully static CGO-disabled Go binary.
- The current build-web command already has a useful fallback path for environments without Dagger.

### What was tricky to build
- The Dockerfile initially looked like a normal static Go service build, but `go-sqlite3` changes that assumption. The failure only appeared at runtime when the help store initialized, so the validation had to include actually starting `glaze serve`, not just compiling the image.

### What warrants a second pair of eyes
- Confirm that `ghcr.io/go-go-golems/glazed:main` is the desired production tag policy; pinning to `sha-<commit>` would be safer for deterministic rollouts.
- Confirm whether cert-manager's HTTP-01 issuer can issue a wildcard certificate in this cluster. Wildcard certificates normally require DNS-01. If the existing `letsencrypt-prod` ClusterIssuer is HTTP-01 only, split TLS into `glaze.docs.scapegoat.dev` now and add wildcard TLS when DNS-01 is configured.

### What should be done in the future
- Push the Glazed container workflow to GitHub and verify the first GHCR image exists.
- Push/apply the GitOps Application after the image exists.
- Run a Terraform plan and apply for the wildcard DNS record.

### Code review instructions
- Review Dockerfile build/runtime choices first.
- Review Kustomize output with `kubectl kustomize gitops/kustomize/glaze-docs`.
- Review DNS with `terraform -chdir=dns/zones/scapegoat-dev/envs/prod plan`.

### Technical details
- Health endpoint: `GET /api/health`
- Service port: Kubernetes Service `80` -> container port `8088`
- Public host: `glaze.docs.scapegoat.dev`
- DNS wildcard: `*.docs.scapegoat.dev A 91.98.46.169`

## Step 3: Keep HTTP-01 TLS now and create a future DNS-01 issue

After reviewing the existing `letsencrypt-prod` `ClusterIssuer`, I confirmed that it is HTTP-01-only. That makes it suitable for concrete hostnames such as `glaze.docs.scapegoat.dev`, but not for wildcard certificates such as `*.docs.scapegoat.dev`. This is the same operational pattern as the existing `*.yolo.scapegoat.dev` setup: wildcard DNS points at the cluster, while each app requests a concrete HTTP-01 certificate.

I updated the Glaze Ingress to request only `glaze.docs.scapegoat.dev` for now and created a GitHub issue to preserve the future DNS-01 design. The issue is https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65.

### Prompt Context

**User prompt (verbatim):** "ok, store all that information in a github issue for the future, so that we can get a solid start. we'll keep the http-01 explicit domain name for now."

**Assistant interpretation:** Persist the DNS-01 wildcard TLS plan as a GitHub issue, and change the current deployment to use only the explicit HTTP-01 hostname.

**Inferred user intent:** Avoid blocking the current Glaze deployment on DNS-01 work while keeping a detailed future implementation plan.

### What I did
- Edited `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml` to remove `*.docs.scapegoat.dev` from TLS hosts.
- Changed the TLS secret from `docs-scapegoat-dev-tls` to `glaze-docs-tls` for the concrete Glaze hostname.
- Rendered the Kustomize app and confirmed the Ingress TLS hosts contain only `glaze.docs.scapegoat.dev`.
- Wrote `reference/02-future-dns-01-wildcard-tls-github-issue.md` as the durable issue source.
- Created GitHub issue https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65.
- Updated the production deployment guide to reflect the current HTTP-01 decision.

### Why
- The live GitOps issuer is HTTP-01-only.
- Let's Encrypt wildcard certificates require DNS-01.
- Keeping wildcard DNS but using concrete TLS is the safest immediate deployment shape.

### What worked
- `kubectl kustomize gitops/kustomize/glaze-docs` rendered successfully.
- The rendered Ingress now shows `spec.tls[0].hosts = [glaze.docs.scapegoat.dev]`.
- GitHub issue creation succeeded with `gh issue create`.

### What didn't work
- N/A in this step.

### What I learned
- The future DNS-01 work belongs in the k3s GitOps repo because it changes the platform-level `ClusterIssuer` and secret-management pattern.

### What was tricky to build
- The distinction between wildcard DNS and wildcard TLS is easy to blur. DNS wildcard support is already straightforward with DigitalOcean Terraform, but wildcard TLS introduces ACME DNS-01 and secret-management requirements.

### What warrants a second pair of eyes
- The future issue proposes a DigitalOcean token secret named `digitalocean-dns`; review whether this should be manually bootstrapped, Vault-managed, or handled by another encrypted GitOps mechanism.

### What should be done in the future
- Implement GitHub issue #65 when wildcard docs certificates are actually needed.
- Keep the current Glaze deployment on concrete HTTP-01 TLS until then.

### Code review instructions
- Review the Glaze Ingress first: it should request only `glaze.docs.scapegoat.dev` TLS.
- Review issue #65 for the future DNS-01 plan.
- Validate with `kubectl kustomize gitops/kustomize/glaze-docs`.

### Technical details
- GitHub issue: https://github.com/wesen/2026-03-27--hetzner-k3s/issues/65
- Current TLS secret: `glaze-docs-tls`
- Current DNS remains wildcard: `*.docs.scapegoat.dev A 91.98.46.169`
