---
Title: Investigation diary
Ticket: GG-20260502-VAULT-OIDC-DOCS-PUBLISH
Status: active
Topics:
    - glazed
    - docs
    - deploy
    - kubernetes
    - gitops
    - security
    - vault
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Diary for the Phase 2/3 Vault-backed docs publishing auth design."
LastUpdated: 2026-05-02T14:20:00-04:00
WhatFor: "Records how the Vault GitHub OIDC and signed JWT publishing auth guide was created."
WhenToUse: "Use when continuing the Phase 2/3 docs publishing auth implementation."
---

# Diary

## Goal

This diary records the creation of the Phase 2 and Phase 3 publishing-auth design for `docs.yolo.scapegoat.dev`: GitHub OIDC to Vault, registry-side Vault capabilities checks, and later Vault-signed docs publish JWTs.

## Step 1: Split advanced publishing auth into its own ticket

The current multi-package docs deployment ticket now owns Phase 1: static package publish tokens stored and rotated through Vault. Phase 2 and Phase 3 are larger and more security-sensitive, so they deserve a separate ticket and implementation guide.

This split keeps the first implementation achievable while preserving the desired long-term model: no long-lived GitHub repository secrets, per-repository package authorization, short-lived credentials, and auditable publisher identity.

### Prompt Context

**User prompt (verbatim):** "update the current ticket design and the tasks with phase 1.

but also store all the phase 2 and phase 3 stuff in a deteailed implementation guide in a new docmgr ticket.

 Create a new docmgr to write the tool, and  reate a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                             
  system needed to understand what it is, with prose paragraphs and bullet                                                                                                                                    
  point sand pseudocode and diagrams and api references and file                                                                                                                                             
    references.                                                                                                                                                                                              
    It should be very clear and technical. Store in the ticket and the nupload to remarkable.       

---

After that, work on the current ticket, task by task, committing at appropriate intervals, keeping a clean diary."

**Assistant interpretation:** Update the current docs-yolo design with Phase 1 static Vault-token publishing, then create a new ticket for the more advanced Phase 2/3 Vault/GitHub OIDC and signed JWT publishing-auth model. The new ticket should contain a detailed intern-facing implementation guide and be uploaded to reMarkable.

**Inferred user intent:** Keep the immediate implementation tractable while not losing the stronger future authentication design.

### What I did

- Created ticket `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`.
- Created design doc `design-doc/01-vault-github-oidc-and-signed-jwt-docs-publishing-auth-implementation-guide.md`.
- Created this diary.
- Wrote the Phase 2 design:
  - GitHub Actions OIDC token;
  - Vault JWT auth role;
  - short-lived Vault token;
  - registry calls `sys/capabilities-self` for `docs-yolo/publish/<package>`.
- Wrote the Phase 3 design:
  - Vault-backed token issuer;
  - purpose-built docs publish JWT;
  - local registry JWT validation;
  - package/version/permission claims.
- Added implementation phases, pseudocode, API sketches, security checklist, test strategy, and rollback plan.

Commands run:

```bash
cd /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed
docmgr ticket create-ticket --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --title "Design Vault GitHub OIDC docs publishing auth for Glazed packages" --topics glazed,docs,deploy,kubernetes,gitops,security,vault
docmgr doc add --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --doc-type design-doc --title "Vault GitHub OIDC and signed JWT docs publishing auth implementation guide"
docmgr doc add --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --doc-type reference --title "Investigation diary"
```

### Why

Phase 2/3 auth involves Vault JWT auth, GitHub Actions OIDC claims, registry authorization, token exchange boundaries, and possibly a token issuer service. That is enough complexity that it should not be embedded as a small note in the Phase 1 deployment ticket.

### What worked

- The split made the implementation sequence clearer:
  - current ticket: Phase 1 static package tokens and core publishing path;
  - new ticket: Phase 2/3 identity hardening.
- Existing k3s Vault diaries and playbooks provide useful patterns for least-privilege Vault policies and operational validation.

### What didn't work

N/A for this step. No command failed while creating the new ticket or writing the guide.

### What I learned

The cleanest Phase 2 registry check is to treat Vault policy paths like an authorization namespace:

```text
docs-yolo/publish/<package>
```

The registry does not need to read real data at that path. It only needs to ask Vault whether the caller token has `update` capability on that path.

### What was tricky to build

The subtle boundary is deciding whether the registry should accept Vault client tokens directly. That is acceptable in Phase 2 if tokens are short-lived and scoped, but Phase 3 should replace it with a dedicated docs publish JWT so the registry can validate a narrow application token locally.

### What warrants a second pair of eyes

- Whether Vault will be reachable from GitHub-hosted runners or whether self-hosted runners are required.
- Whether package publisher roles should be generated from a GitOps catalog.
- Whether Phase 3 should use Vault Transit directly or a small token issuer service.

### What should be done in the future

- Validate the Phase 2 flow against a test Vault JWT auth mount before adding production package roles.
- Create a reusable GitHub Actions workflow once the registry endpoint exists.
- Add a token issuer prototype only after Phase 2 has been proven.

### Code review instructions

Review:

- `design-doc/01-vault-github-oidc-and-signed-jwt-docs-publishing-auth-implementation-guide.md`
- `reference/01-investigation-diary.md`
- `tasks.md`

Validate:

```bash
docmgr doctor --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --stale-after 30
```

### Technical details

The main Phase 2 authorization check is:

```http
POST /v1/sys/capabilities-self
X-Vault-Token: <publisher-token>

{"paths":["docs-yolo/publish/pinocchio"]}
```

The main Phase 3 registry check is local JWT verification of claims:

```text
iss == vault://docs-yolo
aud contains docs-registry
package == requested package
permissions contains publish
exp is in the future
```

## Step 2: Validate and upload the Phase 2/3 auth guide

After writing the guide, I validated the docmgr ticket and uploaded the design/diary bundle to reMarkable. This preserves the advanced auth plan independently from the Phase 1 implementation ticket.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Deliver the Phase 2/3 guide to reMarkable after docmgr validation.

**Inferred user intent:** Make the future auth design reviewable as a standalone artifact.

### What I did

- Ran docmgr doctor for the new ticket.
- Ran a dry-run reMarkable bundle upload.
- Uploaded the design doc and diary as one PDF bundle.
- Verified the remote folder listing.

Commands run:

```bash
docmgr doctor --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --stale-after 30
```

```bash
remarquee upload bundle --dry-run <design> <diary> \
  --name "GG-20260502-VAULT-OIDC-DOCS-PUBLISH - Vault OIDC docs publishing auth" \
  --remote-dir "/ai/2026/05/02/GG-20260502-VAULT-OIDC-DOCS-PUBLISH" \
  --toc-depth 2
```

```bash
remarquee upload bundle <design> <diary> \
  --name "GG-20260502-VAULT-OIDC-DOCS-PUBLISH - Vault OIDC docs publishing auth" \
  --remote-dir "/ai/2026/05/02/GG-20260502-VAULT-OIDC-DOCS-PUBLISH" \
  --toc-depth 2

remarquee cloud ls /ai/2026/05/02/GG-20260502-VAULT-OIDC-DOCS-PUBLISH --long --non-interactive
```

### Why

The ticket is intended as a future implementation guide. Uploading it now makes it available for review while Phase 1 work continues separately.

### What worked

- `docmgr doctor` passed.
- Dry-run upload showed the correct bundle inputs and destination.
- Real upload succeeded.
- Remote listing showed the bundle.

### What didn't work

N/A.

### What I learned

The Phase 2/3 guide is now cleanly separated from the Phase 1 docs-yolo ticket, which should make task execution less confusing.

### What was tricky to build

The tricky part was avoiding duplicate scope. The Phase 1 ticket now has only the static Vault token addendum; this ticket owns GitHub OIDC, Vault capabilities, and signed JWTs.

### What warrants a second pair of eyes

- Whether the reMarkable guide should also be mirrored into the k3s docmgr workspace when implementation moves there.

### What should be done in the future

- Start with a Vault dev-server proof of concept for GitHub OIDC claims before changing production Vault.

### Code review instructions

Review the uploaded design and validate locally with:

```bash
docmgr doctor --ticket GG-20260502-VAULT-OIDC-DOCS-PUBLISH --stale-after 30
```

### Technical details

Remote destination:

```text
/ai/2026/05/02/GG-20260502-VAULT-OIDC-DOCS-PUBLISH
```

Bundle:

```text
GG-20260502-VAULT-OIDC-DOCS-PUBLISH - Vault OIDC docs publishing auth
```

## Step 2: Rewrite the docs publishing auth target as clean Vault OIDC Option B

The user clarified that the platform already has GitHub Actions OIDC login to Vault through HK3S-0028, so the docs publishing design should go straight to the clean long-term path instead of productionizing static package tokens. I wrote a new intern-facing implementation guide that treats static token publishing as a proven Phase 1 scaffold, not the desired production mechanism.

The new design centers on Vault-issued short-lived package-scoped publish credentials. GitHub Actions authenticates to Vault using OIDC, Vault roles map trusted repositories to allowed package names, `docsctl publish` obtains or accepts a short-lived publish JWT, and `docs-registry` verifies the JWT directly before validating and storing the uploaded SQLite DB.

### Prompt Context

**User prompt (verbatim):** "Create a full design / implementation document for the vault OIDC approach with option B (the clean long term solution). We want to basically go straight to the proper way of doing things to avoid accumulating backwards / step by step building complexity along the way that we would need to shed anyway. We've shown that the underlying mechanism works fine.

 Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                                                                      
  system needed to understand what it is, with prose paragraphs and                                                                                                                                         
bullet                                                                                                                                                                                                      
  point sand pseudocode and diagrams and api references and                                                                                                                                                 
file                                                                                                                                                                                                        
                                                                                                                                                                                                            
references.                                                                                                                                                                                                 
    It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Produce a new detailed ticket design for the long-term Vault/GitHub OIDC docs publishing architecture, explicitly avoiding a production dependency on static publisher tokens.

**Inferred user intent:** Give future implementers a clear, evidence-backed blueprint for replacing the Phase 1 static token catalog with Vault-issued short-lived package-scoped publish credentials.

### What I did

- Read the existing docs publishing auth ticket and created a new design document:
  - `design-doc/02-vault-oidc-docs-publishing-clean-long-term-implementation-guide.md`
- Re-read the HK3S-0028 Vault/GitHub Actions OIDC article in the Obsidian vault.
- Inspected current Glazed publisher auth and registry files:
  - `pkg/help/publish/auth.go`
  - `pkg/help/publish/registry.go`
  - `cmd/docsctl/publish.go`
- Inspected current k3s OIDC and docs-yolo files:
  - `scripts/bootstrap-vault-github-actions-oidc.sh`
  - `vault/roles/github-actions/bot-signup-gitops-pr.json`
  - `gitops/kustomize/docs-yolo/deployment.yaml`
  - `gitops/kustomize/docs-yolo/service.yaml`
  - `gitops/kustomize/docs-yolo/ingress.yaml`
  - `gitops/kustomize/docs-yolo/publisher-catalog.yaml`
- Wrote a full implementation guide covering:
  - current state with file evidence
  - target sequence diagram
  - publish JWT claim contract
  - Vault role and policy shape
  - registry JWT verification rules
  - `docsctl publish` UX
  - k3s deployment changes
  - rollout and test plan
  - risks and alternatives

### Why

The existing Phase 1 static token path works, but it would create operational complexity that the platform now has a better primitive to avoid. HK3S-0028 already proves that GitHub Actions can authenticate to Vault without GitHub-stored long-lived secrets, so docs publishing should use that identity path directly.

### What worked

- The current `PublisherAuth` interface is already a useful seam for a future `JWTPublisherAuth` implementation.
- The registry already authorizes before reading the upload body, which is the right shape for a public or semi-public registry endpoint.
- The k3s repo already has GitHub Actions OIDC role and policy conventions that docs publishing can copy.

### What didn't work

- No implementation was attempted in this step.
- The exact Vault mechanism for minting constrained JWTs still needs to be selected during implementation. The design calls out direct Vault issuance, a broker/helper, and transit signing tradeoffs.

### What I learned

The important distinction is that Vault OIDC removes the need for GitHub-stored static docs tokens, but the system still needs an authorization mapping. In the clean model, that mapping moves from `publishers.json` into Vault roles/policies and signed package claims.

### What was tricky to build

The tricky part of the design is avoiding an unsafe "raw transit signing" shortcut. If CI can ask Vault transit to sign arbitrary bytes with a generic trusted key, it could potentially sign a JWT that claims a different package. The guide therefore recommends direct claim-constrained Vault issuance or a controlled broker/helper, and only allows transit signing if keys or registry verification are package-scoped.

### What warrants a second pair of eyes

- The final Vault JWT issuance mechanism.
- The registry's JWKS/key rotation strategy.
- Whether the registry should be exposed publicly at `registry.docs.yolo.scapegoat.dev` or reached through an in-cluster/self-hosted runner path.

### What should be done in the future

- Implement `JWTPublisherAuth` in Glazed.
- Add Vault docs-publish roles/policies in the k3s repo.
- Add `docsctl publish --auth-mode vault-jwt` support.
- Update docs-yolo manifests to remove the static publisher catalog in production.
- Run one live package CI publish using Vault OIDC.

### Code review instructions

Start with the new design document and compare it to:

- `pkg/help/publish/auth.go` for the current auth seam.
- `pkg/help/publish/registry.go` for upload authorization order.
- `cmd/docsctl/publish.go` for current client credential UX.
- `scripts/bootstrap-vault-github-actions-oidc.sh` for the platform OIDC role/policy pattern.
- `gitops/kustomize/docs-yolo/deployment.yaml` for the production registry flags that will change.

### Technical details

The target invariant from the document is:

```text
GitHub repository X on trusted ref/event may publish only package Y,
and docs-registry must verify that the signed package claim equals the route package.
```
