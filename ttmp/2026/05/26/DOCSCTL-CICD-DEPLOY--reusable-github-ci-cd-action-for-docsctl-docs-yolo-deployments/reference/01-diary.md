---
Title: Diary
Ticket: DOCSCTL-CICD-DEPLOY
Status: active
Topics:
    - docsctl
    - github-actions
    - cicd
    - kubernetes
    - vault
    - oidc
    - docs-yolo
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml
      Note: Evidence for current docs-yolo pod shape and missing SSR sidecar
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Docsctl and Docs-Yolo Documentation Deployment.md
      Note: Source article requested by the user and used as baseline system context
    - Path: glazed/cmd/docsctl/publish.go
      Note: Evidence for docsctl publish URL construction and token handling
    - Path: glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md
      Note: |-
        Primary design deliverable created in Step 1
        Primary design deliverable recorded by this diary
    - Path: glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/sources/01-article-docsctl-and-docs-yolo-documentation-deployment.md
      Note: Copied source article for ticket-local reference
ExternalSources:
    - https://docs.github.com/actions/reference/openid-connect-reference
    - https://developer.hashicorp.com/vault/docs/auth/jwt
Summary: Chronological notes for DOCSCTL-CICD-DEPLOY research and design.
LastUpdated: 2026-05-26T06:22:01.183359942-04:00
WhatFor: Use this to resume or review how the reusable docsctl CI/CD design was produced.
WhenToUse: Read before continuing implementation, changing the design, or auditing the evidence trail.
---


# Diary

## Goal

This diary records the creation of the DOCSCTL-CICD-DEPLOY ticket and the evidence-backed design for reusable GitHub Actions publishing of docsctl packages to docs-yolo, including Vault OIDC caller validation and SSR sidecar deployment implications.

## Step 1: Create ticket, gather evidence, and write the implementation guide

I created the `DOCSCTL-CICD-DEPLOY` ticket, copied the requested Obsidian article into the ticket sources, inspected the existing `docsctl`, `docs-registry`, docs-browser SSR, k3s docs-yolo, and Vault GitHub Actions OIDC materials, and wrote the primary design document. The result is a long-form intern-ready guide that explains the system boundaries, the reusable workflow contract, the Vault role model, the registry reachability decision, and the SSR sidecar packaging/deployment work.

The most important design decision is that Vault should validate the GitHub Actions caller before releasing a package-scoped docsctl publish token. For reusable workflows, the Vault role should bind both the caller repository claim (`repository` or preferably `repository_id`) and the reusable workflow claim (`job_workflow_ref`). This directly addresses the unknown in the prompt: how to validate that the caller is the right repository while still centralizing the implementation in a reusable workflow.

### Prompt Context

**User prompt (verbatim):** "Let's create a new ticket to create a reusable github CICD action to deploy docs to docs.yolo.scapegoat.dev with docsctl. Read /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Docsctl and Docs-Yolo Documentation Deployment.md

We just worked on DOCSCTL-SSR and DOCSCTL-A14Y which also requires the packaging and deployment of the SSR sidecar in k3s as well (/home/manuel/code/wesen/2026-03-27--hetzner-k3s)

We also need to figure out how to add proper vault github oidc login and validate that the caller is the right repo, which I don't know how to do exactly.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a diary as you work (see format in skill)"

**Assistant interpretation:** Create a docmgr ticket and deliver a detailed design/implementation guide for reusable GitHub Actions docsctl publishing to docs-yolo, including Vault OIDC security, SSR sidecar deployment, and reMarkable delivery.

**Inferred user intent:** The user wants a handoff-quality technical plan that a new intern can implement safely without losing context from DOCSCTL-SSR, DOCSCTL-A14Y, the docs-yolo deployment, or prior Vault GitHub OIDC work.

**Commit (code):** N/A — documentation-only work in the ticket workspace.

### What I did

- Loaded the ticket-research/docmgr/reMarkable workflow instructions.
- Read the requested source article:
  - `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Docsctl and Docs-Yolo Documentation Deployment.md`
- Created ticket `DOCSCTL-CICD-DEPLOY` with topics:
  - `docsctl`
  - `github-actions`
  - `cicd`
  - `kubernetes`
  - `vault`
  - `oidc`
  - `docs-yolo`
- Added the design doc:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md`
- Added this diary:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/reference/01-diary.md`
- Added initial ticket tasks for system mapping, reusable workflow design, Vault OIDC design, SSR sidecar deployment, and reMarkable upload.
- Copied the requested source article into ticket sources:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/sources/01-article-docsctl-and-docs-yolo-documentation-deployment.md`
- Inspected prior tickets:
  - `DOCSCTL-SSR`
  - `DOCSCTL-A14Y`
- Inspected implementation files:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docsctl/publish.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/directory_store.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/server/serve.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/package.json`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/web/server.mjs`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/Dockerfile`
- Inspected k3s files:
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/service.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/vault-static-secret.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/vault-auth.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/vault/policies/kubernetes/docs-yolo.hcl`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/vault/roles/kubernetes/docs-yolo.json`
- Inspected existing Vault GitHub Actions OIDC documentation and Terraform:
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/github-actions-vault-oidc-playbook.md`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/source-app-deployment-infrastructure-playbook.md`
  - `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/ttmp/2026/05/02/HK3S-0028--enable-github-actions-oidc-access-to-vault/sources/github-oidc-reference.defuddle.md`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/ttmp/2026/05/02/HK3S-0028--enable-github-actions-oidc-access-to-vault/sources/vault-jwt-auth-docs.defuddle.md`

### Why

- The existing manual publishing path uses an operator port-forward and local package tokens; CI needs a repeatable, least-privilege equivalent.
- The existing Vault GitHub Actions OIDC pattern already solves a similar problem for GitOps PR credentials, so the docs publish design should reuse that model instead of inventing a new secret system.
- The SSR and a14y work changes what “done” means for docs-yolo: publishing databases alone is not enough if production does not also run the SSR sidecar and agent-readable route handling.

### What worked

- The requested article provided a clear baseline for docs-yolo's current read/write architecture.
- The existing `HK3S-0028` Vault OIDC docs and Terraform provided a concrete pattern for GitHub Actions claim binding.
- The code evidence was sufficient to identify a subtle `docsctl publish` issue: the help text mentions `DOCSCTL_TOKEN`, but the current resolver only checks `--token` and `--token-file`.
- The design now clearly separates first implementation from future hardening: use Vault-gated package tokens first; consider direct GitHub OIDC registry auth later.

### What didn't work

- No command failed during this step.
- One documentation/code mismatch was found, not caused by this step:
  - `cmd/docsctl/publish.go` says token precedence is `--token`, `DOCSCTL_TOKEN`, then `--token-file`.
  - `resolvePublishToken` currently checks only `--token` and `--token-file`.
  - The design works around this by recommending `--token-file` in CI until the mismatch is fixed.

### What I learned

- For reusable GitHub Actions workflows, `job_workflow_ref` is the key claim for proving that a trusted reusable workflow performed the sensitive Vault read.
- `repository` or `repository_id` is still needed because `job_workflow_ref` alone only proves which reusable workflow ran, not which repository invoked it.
- The current docs-yolo registry service is ClusterIP-only, so GitHub-hosted CI cannot publish unless we either expose a separate registry ingress or use a trusted self-hosted runner.
- The SSR source exists in `glazed`, but the live k3s deployment has not yet been changed to run the Node sidecar.

### What was tricky to build

- The hardest part was designing the security boundary for a reusable workflow. A reusable workflow centralizes implementation, but the sensitive authorization decision must remain caller-specific. The solution is to make Vault, not the workflow, enforce caller-specific policy through OIDC `bound_claims`.
- The second tricky part was registry reachability. The manual workflow uses a local port-forward to an internal service. A hosted GitHub runner cannot do that without broader cluster credentials. The design therefore forces an explicit choice between a public write ingress with strong package-token controls and a self-hosted runner with internal network reachability.
- The SSR sidecar is also a cross-repository integration point: code support lives in `glazed`, but production activation lives in the k3s GitOps repo. The guide calls out both image packaging and manifest changes so an intern does not only implement the Node code and forget the deployment.

### What warrants a second pair of eyes

- The exact GitHub owner/repository names in proposed Vault claims must be verified against the real GitHub repositories before implementation.
- The public registry ingress option should be security-reviewed before exposing the write API to the internet.
- The first live reusable workflow run should decode and inspect non-sensitive OIDC claims in a sandbox to confirm `repository`, `workflow_ref`, and `job_workflow_ref` values match the proposed Vault role.
- The SSR image/Dockerfile design should be reviewed for reproducibility, image size, and dependency pinning.

### What should be done in the future

- Implement the `docsctl publish` token environment fix or update its help text.
- Add a safe `docsctl token hash` operator command so updating `publishers.json` does not require ad-hoc Go snippets.
- Decide between public registry ingress and self-hosted runner before writing final CI workflow defaults.
- Implement periodic registry publisher-catalog reload or document rollout restart as mandatory after token hash rotation.

### Code review instructions

- Start with the design doc:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md`
- Then compare the claims against implementation files:
  - `cmd/docsctl/publish.go` for docsctl publish behavior.
  - `pkg/help/publish/registry.go` and `pkg/help/publish/auth.go` for registry auth.
  - `pkg/help/server/serve.go` and `web/server.mjs` for SSR behavior.
  - `gitops/kustomize/docs-yolo/deployment.yaml` for the live pod shape.
  - `terraform/vault/github-actions/envs/k3s/main.tf` for existing Vault JWT role conventions.
- Validate ticket hygiene with:
  - `docmgr doctor --ticket DOCSCTL-CICD-DEPLOY --stale-after 30`

### Technical details

Commands run or effectively performed in this step:

```bash
docmgr ticket create-ticket \
  --ticket DOCSCTL-CICD-DEPLOY \
  --title "Reusable GitHub CI/CD action for docsctl docs-yolo deployments" \
  --topics docsctl,github-actions,cicd,kubernetes,vault,oidc,docs-yolo

docmgr doc add \
  --ticket DOCSCTL-CICD-DEPLOY \
  --doc-type design-doc \
  --title "Reusable GitHub CI/CD docsctl deployment guide"

docmgr doc add \
  --ticket DOCSCTL-CICD-DEPLOY \
  --doc-type reference \
  --title "Diary"

cp \
  "/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Docsctl and Docs-Yolo Documentation Deployment.md" \
  "/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/sources/01-article-docsctl-and-docs-yolo-documentation-deployment.md"
```

Key design snippets:

```hcl
bound_claims = {
  repository       = "go-go-golems/glazed"
  ref              = "refs/heads/main"
  event_name       = "push"
  job_workflow_ref = "go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@refs/heads/main"
}
```

```yaml
- uses: hashicorp/vault-action@v3
  with:
    url: https://vault.yolo.scapegoat.dev
    method: jwt
    path: github-actions
    role: docsctl-glazed-publisher
    jwtGithubAudience: vault://docs-yolo/docsctl-publish/glazed
    secrets: |
      kv/data/apps/docs-yolo/publish/tokens/glazed token | DOCSCTL_PUBLISH_TOKEN
```

## Step 2: Validate, fix Mermaid rendering, and upload the bundle to reMarkable

I ran docmgr validation, resolved the new topic vocabulary warnings, uploaded the design bundle to reMarkable, fixed a Mermaid rendering problem, and re-uploaded the corrected bundle with `--force`. The final upload succeeded cleanly and the ticket task list is complete.

The first upload succeeded but reported Mermaid rendering warnings. The issue was caused by unquoted Mermaid node labels containing path-like text such as `/etc/docs-yolo/publishers.json`. I updated the design doc diagrams to quote labels and reran the upload so the reMarkable PDF is generated from the corrected source.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish ticket hygiene and deliver the documentation bundle to reMarkable.

**Inferred user intent:** The user wants the finished design and diary available both in the docmgr ticket and on the reMarkable device for review.

**Commit (code):** N/A — documentation-only work in the ticket workspace.

### What I did

- Ran `docmgr doctor --ticket DOCSCTL-CICD-DEPLOY --stale-after 30`.
- Added vocabulary entries for new topics:
  - `cicd`
  - `docs-yolo`
  - `github-actions`
  - `oidc`
- Fixed Mermaid diagram labels in the design doc after the first PDF upload reported Mermaid parse warnings.
- Uploaded the final bundle to:
  - `/ai/2026/05/26/DOCSCTL-CICD-DEPLOY/DOCSCTL CICD Deploy Guide.pdf`
- Checked the final reMarkable upload task.

### Why

- `docmgr doctor` should pass before delivery so the ticket is searchable and vocabulary-consistent.
- The Mermaid warnings meant the first generated PDF could be missing diagrams, so the source doc needed a small formatting correction before final delivery.
- The user explicitly requested reMarkable upload.

### What worked

- `docmgr doctor` passed after adding topic vocabulary entries.
- The second reMarkable upload completed with:
  - `OK: uploaded DOCSCTL CICD Deploy Guide.pdf -> /ai/2026/05/26/DOCSCTL-CICD-DEPLOY`

### What didn't work

- First upload command:

```bash
remarquee upload bundle \
  "$TICKET_DIR/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md" \
  "$TICKET_DIR/reference/01-diary.md" \
  "$TICKET_DIR/tasks.md" \
  "$TICKET_DIR/changelog.md" \
  --name "DOCSCTL CICD Deploy Guide" \
  --remote-dir "/ai/2026/05/26/DOCSCTL-CICD-DEPLOY" \
  --toc-depth 2 \
  --non-interactive 2>&1
```

reported warnings like:

```text
WARNING: failed to render Mermaid block 4: mmdc failed: Generating single mermaid chart

Error: Lexical error on line 9. Unrecognized text.
...yolo/publishers.json]
-----------------------^
```

The upload still printed `OK: uploaded`, but I treated the Mermaid warning as a quality issue and corrected the diagrams before final upload.

### What I learned

- Mermaid labels that include slash-heavy paths are safer when quoted, for example `catalog["/etc/docs-yolo/publishers.json"]` instead of `catalog[/etc/docs-yolo/publishers.json]`.
- `remarquee upload bundle --force` is the right follow-up when a PDF was already uploaded but source formatting needed correction.

### What was tricky to build

- The reMarkable upload could succeed even while diagram rendering emitted warnings. The symptom was an `OK: uploaded` success line mixed with Mermaid errors. I handled this by fixing the source rather than accepting a potentially degraded PDF.

### What warrants a second pair of eyes

- Review the uploaded PDF diagrams if the reMarkable rendering is critical, because the CLI success line confirms upload but not visual quality beyond the absence of rendering warnings.

### What should be done in the future

- Consider adding a local Markdown/PDF preflight for ticket docs that include Mermaid diagrams before upload.

### Code review instructions

- Review the corrected Mermaid blocks in:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-CICD-DEPLOY--reusable-github-ci-cd-action-for-docsctl-docs-yolo-deployments/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md`
- Validate with:
  - `docmgr doctor --ticket DOCSCTL-CICD-DEPLOY --stale-after 30`

### Technical details

Successful final upload command:

```bash
remarquee upload bundle \
  "$TICKET_DIR/design-doc/01-reusable-github-ci-cd-docsctl-deployment-guide.md" \
  "$TICKET_DIR/reference/01-diary.md" \
  "$TICKET_DIR/tasks.md" \
  "$TICKET_DIR/changelog.md" \
  --name "DOCSCTL CICD Deploy Guide" \
  --remote-dir "/ai/2026/05/26/DOCSCTL-CICD-DEPLOY" \
  --toc-depth 2 \
  --force \
  --non-interactive 2>&1
```
