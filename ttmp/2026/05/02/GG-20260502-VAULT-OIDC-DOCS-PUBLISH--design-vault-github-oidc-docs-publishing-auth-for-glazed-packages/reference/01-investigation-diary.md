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
