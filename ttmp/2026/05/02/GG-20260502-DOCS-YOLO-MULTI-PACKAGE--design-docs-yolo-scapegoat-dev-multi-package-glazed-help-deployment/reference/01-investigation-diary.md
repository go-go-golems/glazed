---
Title: Investigation diary
Ticket: GG-20260502-DOCS-YOLO-MULTI-PACKAGE
Status: active
Topics:
    - glazed
    - docs
    - deploy
    - kubernetes
    - gitops
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - /home/manuel/code/wesen/obsidian-vault/Projects/2026/05/02/ARTICLE - Deploying Glazed Help Browser to Argo CD - Production Deep Dive.md
Summary: "Chronological investigation diary for the docs.yolo.scapegoat.dev multi-package Glazed docs deployment design."
LastUpdated: 2026-05-02T13:45:00-04:00
WhatFor: "Records evidence, decisions, commands, failures, and validation for the multi-package docs deployment research ticket."
WhenToUse: "Use when continuing the docs.yolo.scapegoat.dev design or reviewing why the recommended registry/reload architecture was chosen."
---

# Diary

## Goal

This diary records the investigation and design work for a future `docs.yolo.scapegoat.dev` deployment that can serve many Glazed help packages and versions from uploaded SQLite documentation exports.

## Step 1: Create the research ticket and map the existing deployment

The first step was to create a dedicated docmgr ticket and gather evidence from the existing Glazed production deployment. The goal was not to design from scratch, but to understand the already-proven path from Glazed source to GHCR image to Argo CD to Traefik/cert-manager before extending it into a shared multi-package documentation hub.

The existing deployment is deliberately small: one `glaze serve` pod, no persistent docs volume, and no external package registry. That simplicity is useful evidence because it identifies exactly what must change for `docs.yolo.scapegoat.dev`: external content storage, upload validation, local package DB materialization, and reload/restart behavior.

### Prompt Context

**User prompt (verbatim):** "we are currently deploying glazed as part of /home/manuel/code/wesen/2026-03-27--hetzner-k3s (look at our recent ticket diary), but I now want a deploy to docs.yolo.scapegoat.dev that handles all kinds of packages. THe idea is that we have some storage where packages can upload their doc sqlite under their version number to be served, maybea bucket or some upload somewhere.

Create a new docmgr ticket where you investigate how to do this multi-package serving glazed (which will probably be able to reload the package dynamically, or maybe on a trigger) and the different options available to make it work (be creative, it doesn't need to be a bucket as I said). Here's a write up of the deploy work we did: /home/manuel/code/wesen/obsidian-vault/Projects/2026/05/02/ARTICLE - Deploying Glazed Help Browser to Argo CD - Production Deep Dive.md

 Create a new docmgr to write the tool, and  reate a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                              
  system needed to understand what it is, with prose paragraphs and bullet                                                                                                                                    
  point sand pseudocode and diagrams and api references and file                                                                                                                                             
    references.                                                                                                                                                                                              
    It should be very clear and technical. Store in the ticket and the nupload to remarkable.       

Keep a diary as you investigate."

**Assistant interpretation:** Create a new docmgr ticket in the Glazed docs workspace, investigate production options for a shared multi-package Glazed help deployment at `docs.yolo.scapegoat.dev`, write a detailed intern-facing design/implementation guide, keep a diary, and upload the final bundle to reMarkable.

**Inferred user intent:** The user wants a durable planning artifact that can guide a future implementation rather than a quick sketch. They also want enough operational context from the current Hetzner/Argo CD deployment that an intern can safely implement the next version.

### What I did

- Created docmgr ticket `GG-20260502-DOCS-YOLO-MULTI-PACKAGE`.
- Created design doc `design-doc/01-docs-yolo-scapegoat-dev-multi-package-glazed-help-deployment-design-and-implementation-guide.md`.
- Created diary doc `reference/01-investigation-diary.md`.
- Added initial ticket tasks for evidence gathering, architecture option analysis, guide writing, bookkeeping, validation, and reMarkable upload.
- Read the Obsidian production deep dive at `/home/manuel/code/wesen/obsidian-vault/Projects/2026/05/02/ARTICLE - Deploying Glazed Help Browser to Argo CD - Production Deep Dive.md`.
- Inspected the k3s GitOps manifests:
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/service.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml`
- Inspected Glazed server, loader, API, and store files:
  - `pkg/help/server/serve.go`
  - `pkg/help/server/handlers.go`
  - `pkg/help/server/types.go`
  - `pkg/help/loader/sources.go`
  - `pkg/help/store/store.go`

Commands run:

```bash
cd /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed
docmgr ticket create-ticket --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --title "Design docs.yolo.scapegoat.dev multi-package Glazed help deployment" --topics glazed,docs,deploy,kubernetes,gitops
docmgr doc add --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --doc-type design-doc --title "docs.yolo.scapegoat.dev multi-package Glazed help deployment design and implementation guide"
docmgr doc add --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --doc-type reference --title "Investigation diary"
```

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
rg -n "glaze|glazed|docs\.scapegoat|glaze\.docs" ttmp gitops docs -S --glob '!**/.git/**'
```

```bash
cd /home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed
rg -n "from-sqlite-dir|SQLiteDir|ListPackages|api/packages|PackageName|PackageVersion" pkg/help -S
```

### Why

The future deployment touches application code, GitOps manifests, storage architecture, package CI workflows, and operator runbooks. A safe design needs evidence from all of those layers.

### What worked

- The existing multi-package Glazed code already has most of the data model and API shape needed for a shared docs browser.
- The existing `--from-sqlite-dir` loader already uses a package/version directory convention that can become the local materialization contract for any storage backend.
- The existing k3s deployment provides a concrete copyable pattern for Argo CD, Deployment, Service, Ingress, probes, TLS annotations, and immutable image tags.

### What didn't work

- An initial `rg` command referenced non-existent `apps` and `manifests` directories in the k3s repo:

```text
rg: apps: No such file or directory (os error 2)
rg: manifests: No such file or directory (os error 2)
```

I corrected the search to use the actual `gitops` directory.

### What I learned

- The current production deployment is startup-loaded and image-centered. That is good for the single built-in Glazed docs site, but not sufficient for independent package docs publication.
- The `SQLiteDirLoader` accepted layout is the key seam. A bucket, OCI registry, upload API, Git repo, or webhook worker can all converge on the same local directory layout.
- Hot reload should not be the first production dependency. A restart-triggered Phase 1 can use existing loading behavior while the storage and publishing workflow is proven.

### What was tricky to build

The tricky design issue is separating durable source of truth from local serving shape. A bucket or OCI registry is good durable storage, but Glazed currently consumes local files. A PVC is a good local serving shape, but it is not an ideal long-term source of truth by itself. The design therefore uses a materialization layer: sync durable artifacts into `/var/lib/glazed-docs/packages`, then let Glazed load that directory.

The other tricky issue is reload safety. The current handler receives a store pointer at construction time. Hot reload should not mutate that active store in place. It should build a new store in the background, validate it, then atomically swap a store provider pointer.

### What warrants a second pair of eyes

- Whether the first durable backend should be S3/R2/MinIO or OCI artifacts in GHCR.
- Whether mutable version aliases like `current` should be allowed.
- Whether the upload registry should be implemented in Glazed itself or as a separate smaller service.
- How to authenticate package publishers: static tokens, Vault-issued tokens, or GitHub OIDC.

### What should be done in the future

- Prototype Phase 1 with a PVC and rollout restart before implementing hot reload.
- Add a concrete `docsctl validate/publish` command or CI reusable workflow after the storage decision is made.
- Add metrics for reload success/failure, last reload timestamp, package count, version count, and section count.

### Code review instructions

Start review with the design doc, then verify the evidence against these files:

- `pkg/help/server/serve.go`
- `pkg/help/loader/sources.go`
- `pkg/help/server/handlers.go`
- `pkg/help/store/store.go`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml`

Validation for this research step is doc validation, not code tests:

```bash
docmgr doctor --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --stale-after 30
```

### Technical details

The most important observed contracts are:

```text
GET /api/packages
GET /api/sections?package=<name>&version=<version>
GET /api/sections/<slug>?package=<name>&version=<version>
```

and the local DB layout:

```text
X.db       -> package X, no version
X/X.db     -> package X, no version
X/Y/X.db   -> package X, version Y
```

## Step 2: Validate and upload the research bundle to reMarkable

After the design guide and diary were written, I validated the docmgr ticket and uploaded the deliverable bundle to reMarkable. This makes the research available outside the repository as a single PDF with a table of contents.

The upload step used the safe workflow: first a dry run, then a real upload, then a remote listing to verify that the file exists in the expected folder.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Validate the docmgr ticket and deliver the completed design/diary bundle to reMarkable.

**Inferred user intent:** The user wants the planning document to be reviewable on the reMarkable device/cloud, not only as repo-local Markdown.

### What I did

- Ran `docmgr doctor --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --stale-after 30`.
- Added missing topic vocabulary entries for `deploy`, `gitops`, and `kubernetes`.
- Removed cross-repository files from docmgr `RelatedFiles` because this docmgr workspace validates related paths relative to the Glazed repository and reported the external paths as missing. The design doc still cites those external files explicitly as evidence and the docs keep the Obsidian article in `ExternalSources`.
- Ran `remarquee status` and `remarquee cloud account --non-interactive`.
- Ran a dry-run bundle upload.
- Uploaded the design doc and diary bundle to `/ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE`.
- Verified the remote listing.

Commands run:

```bash
remarquee status
remarquee cloud account --non-interactive
```

```bash
remarquee upload bundle --dry-run \
  <design-doc.md> <diary.md> \
  --name "GG-20260502-DOCS-YOLO-MULTI-PACKAGE - docs.yolo multi-package Glazed docs design" \
  --remote-dir "/ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE" \
  --toc-depth 2
```

```bash
remarquee upload bundle \
  <design-doc.md> <diary.md> \
  --name "GG-20260502-DOCS-YOLO-MULTI-PACKAGE - docs.yolo multi-package Glazed docs design" \
  --remote-dir "/ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE" \
  --toc-depth 2

remarquee cloud ls /ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE --long --non-interactive
```

### Why

The ticket-research workflow requires both repository-local validation and reMarkable delivery. The remote listing is important because it proves the upload reached the intended folder.

### What worked

- `docmgr doctor` passed after vocabulary and related-file cleanup.
- `remarquee status` reported `remarquee: ok`.
- The dry run showed the expected bundle inputs and destination.
- The real upload succeeded.
- Remote listing showed:

```text
[f]	GG-20260502-DOCS-YOLO-MULTI-PACKAGE - docs.yolo multi-package Glazed docs design
```

### What didn't work

`docmgr doctor` initially warned about unknown topic vocabulary and missing external related files. The exact topic warning was for:

```text
deploy, gitops, kubernetes
```

The missing related file warnings were for cross-repository paths under `/home/manuel/code/wesen/...`. I resolved this by adding the vocabulary entries and removing the external files from `RelatedFiles`; the design doc still cites them as absolute file references.

The real reMarkable upload emitted a non-fatal warning:

```text
WARNING: 2026/05/02 13:08:17 apictx.go:259: remote tree has changed, refresh the file tree
```

The upload still completed successfully and the remote listing verified the file.

### What I learned

For this docmgr workspace, external files outside the Glazed repo are better cited in prose or `ExternalSources` rather than ticket `RelatedFiles`, because `docmgr doctor` treats the generated relative paths as missing. Repo-local Glazed files remain suitable as related files.

### What was tricky to build

The ticket legitimately depends on evidence from three locations: Glazed, the k3s GitOps repo, and the Obsidian article. The docmgr relationship model worked cleanly for the Glazed repo-local files but not for external absolute paths. The workaround preserves validation while keeping the actual file references in the design doc.

### What warrants a second pair of eyes

- Whether the ticket should also be mirrored into the k3s docmgr workspace because part of the implementation will live there.
- Whether future cross-repo research tickets should keep external files in `ExternalSources` from the start to avoid doctor warnings.

### What should be done in the future

- If implementation starts, create a paired k3s ticket or explicitly decide that this Glazed ticket remains the parent design ticket.
- When the storage backend is chosen, add a focused implementation diary step before writing code.

### Code review instructions

Review the final artifacts:

- design doc: `design-doc/01-docs-yolo-scapegoat-dev-multi-package-glazed-help-deployment-design-and-implementation-guide.md`
- diary: `reference/01-investigation-diary.md`
- tasks: `tasks.md`
- changelog: `changelog.md`

Validate with:

```bash
docmgr doctor --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --stale-after 30
```

### Technical details

reMarkable destination:

```text
/ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE
```

Bundle name:

```text
GG-20260502-DOCS-YOLO-MULTI-PACKAGE - docs.yolo multi-package Glazed docs design
```

## Step 3: Add Phase 1 static Vault-token publishing scope

The current ticket now explicitly owns Phase 1 of publishing authentication. The goal is to keep the first implementation tractable: use one package-scoped static publish token per package, store and rotate token hashes through Vault, and enforce package scoping in the registry.

This does not replace the stronger GitHub OIDC design. Instead, the Phase 2/3 design moved into a separate ticket, `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`, so that the current ticket can proceed task by task on the static-token MVP.

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

**Assistant interpretation:** Update the current docs-yolo ticket with the Phase 1 static-token publishing model, create a separate Phase 2/3 auth ticket, upload it, and keep a clean diary while proceeding through tasks.

**Inferred user intent:** Avoid losing the ambitious Vault/GitHub OIDC plan while making the immediate implementation small enough to execute.

### What I did

- Appended a Phase 1 addendum to the current design doc.
- Added current-ticket tasks for:
  - static Vault-stored package publish token model;
  - `docsctl validate/publish` contract;
  - token-hash-to-package registry authorization;
  - token creation/rotation runbooks;
  - diary/changelog/re-upload.
- Checked off the completed documentation tasks for Phase 1.
- Created the separate Phase 2/3 ticket `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`.

### Why

The static-token Phase 1 is the fastest way to prove the docs registry and upload flow without introducing GitHub OIDC, Vault JWT auth mounts, token issuer services, or public Vault reachability decisions.

### What worked

- The Phase 1 addendum now gives enough implementation detail for an intern to build static package-token authorization.
- Phase 2/3 is no longer mixed into the Phase 1 task list.

### What didn't work

A task check command initially checked the wrong subset of Phase 1 task IDs. I read `tasks.md`, corrected the remaining unchecked task IDs, and checked the appropriate completed items.

### What I learned

The Phase 1 registry can still use Vault meaningfully without GitHub OIDC: Vault can be the operator source of truth for package token hashes and rotation metadata.

### What was tricky to build

The tricky part was making Phase 1 useful without pretending it is the final model. The design now explicitly lists Phase 1 limitations and points to the separate Phase 2/3 ticket for repository-scoped short-lived credentials.

### What warrants a second pair of eyes

- Whether the Phase 1 registry should read token hashes directly from Vault or maintain its own token table populated by an operator command.
- Whether static tokens should be accepted for all internal packages or only for the first one or two pilot packages.

### What should be done in the future

- Implement `docsctl validate` first.
- Then implement static-token registry authorization tests before wiring real upload storage.

### Code review instructions

Start with the Phase 1 addendum at the bottom of the current design doc. Then review `tasks.md` and this diary step.

Validation:

```bash
docmgr doctor --ticket GG-20260502-DOCS-YOLO-MULTI-PACKAGE --stale-after 30
```

### Technical details

The central Phase 1 invariant is:

```text
hash(token) -> exactly one package
```

Registry authorization should reject:

```text
pinocchio token publishing glazed
empty token
unknown token hash
invalid version string
path traversal package/version names
```

## Step 4: Upload the updated Phase 1 ticket bundle to reMarkable

After the Phase 1 static Vault-token addendum was added and committed, the user asked to upload the first ticket again. This upload includes the fuller ticket bundle, not just the design and diary, so the reMarkable copy has the index, task state, changelog, design guide, and diary together.

### Prompt Context

**User prompt (verbatim):** "upload the first ticket to remarkable, now that it's been updated with phase 1"

**Assistant interpretation:** Upload the updated `GG-20260502-DOCS-YOLO-MULTI-PACKAGE` ticket bundle after the Phase 1 addendum.

**Inferred user intent:** Ensure the reMarkable copy reflects the latest Phase 1 task and design state.

### What I did

- Prepared a bundle containing:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - the Phase 1-updated design doc
  - this diary
- Ran a dry-run upload.
- Uploaded the bundle to the existing ticket folder on reMarkable.
- Verified the remote listing.

### Why

The earlier upload proved the updated design could be delivered, but this step creates a complete ticket-level bundle for review.

### What worked

- Dry-run upload showed the expected five bundle inputs: index, tasks, changelog, design doc, and diary.
- Real upload succeeded.
- Remote listing verified the new complete-ticket bundle alongside the earlier uploads.

### What didn't work

N/A.

### What I learned

N/A.

### What was tricky to build

The only detail is naming: to avoid overwriting earlier PDFs, this upload uses a distinct bundle name ending in `complete ticket with Phase 1`.

### What warrants a second pair of eyes

N/A.

### What should be done in the future

N/A.

### Code review instructions

Verify the reMarkable folder:

```bash
remarquee cloud ls /ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE --long --non-interactive
```

### Technical details

Remote destination:

```text
/ai/2026/05/02/GG-20260502-DOCS-YOLO-MULTI-PACKAGE
```

Bundle name:

```text
GG-20260502-DOCS-YOLO-MULTI-PACKAGE - complete ticket with Phase 1
```
