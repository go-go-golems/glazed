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

## Step 5: Expand the ticket into phased implementation tasks

The ticket previously had design-completion tasks only. I expanded it into a full implementation backlog with phases for local validation, static token auth, registry upload, atomic PVC publication, `docsctl publish`, Vault/catalog integration, GitOps deployment, and end-to-end smoke validation.

This turns the ticket from a research artifact into an executable checklist. The later GitHub OIDC and signed JWT work remains explicitly tracked by `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`, while this ticket focuses on Phase 1 direct registry/PVC publishing with Vault-managed static package tokens.

### Prompt Context

**User prompt (verbatim):** "add detailed list of tasks for the whole ticket, in multiple phases. Then work on the first phase, task by task, committing at appropriate intervals, keeping a diary as you go"

**Assistant interpretation:** Expand the current ticket task list into detailed multi-phase implementation tasks, then start executing Phase 1 in small committed slices with diary updates.

**Inferred user intent:** Move from design into implementation without losing task granularity or progress traceability.

### What I did

- Rewrote `tasks.md` into sections:
  - completed research/design setup;
  - Phase 1A local validation and `docsctl` CLI foundation;
  - Phase 1B static package-token auth;
  - Phase 1C direct upload registry skeleton;
  - Phase 1D PVC directory publisher and atomic writes;
  - Phase 1E `docsctl publish` client;
  - Phase 1F Vault token-hash integration;
  - Phase 1G docs-yolo GitOps deployment scaffold;
  - Phase 1H end-to-end smoke validation and handoff;
  - later phases tracked separately.

### Why

The implementation needs enough task detail to support task-by-task work, focused commits, and diary entries. A single “implement Phase 1” task would be too coarse to review safely.

### What worked

The Phase 1A/1B split gives a clean starting point: first build local validation and CLI foundation, then add auth.

### What didn't work

N/A.

### What I learned

The best implementation order is to validate package DBs before introducing registry or Vault behavior. That keeps the first code slice independent and easy to test.

### What was tricky to build

The tricky part was keeping this ticket focused while still listing the whole path. The solution was to include Phase 2/3 only as references to the separate auth ticket and to keep all direct implementation tasks in Phase 1.

### What warrants a second pair of eyes

- Whether Phase 1F should choose direct Vault reads from the registry or a file-backed catalog first.
- Whether GitOps tasks should be executed in this ticket or split into a paired k3s ticket before manifests are changed.

### What should be done in the future

Start Phase 1A with `cmd/docsctl`, name validation helpers, SQLite DB validator, and `docsctl validate`.

### Code review instructions

Review `tasks.md` first. It is now the implementation map for the ticket.

### Technical details

The first code slice should not require Kubernetes, Vault, or a running registry. It should be testable with:

```bash
go test ./pkg/help/publish/...
go test ./cmd/docsctl/...
```

## Step 6: Add the `docsctl` CLI entrypoint

The first implementation task was intentionally small: create a dedicated `docsctl` command that can grow validation and publishing subcommands without changing the existing `glaze` CLI. This gives package maintainers and CI workflows a purpose-built tool surface for docs publishing.

The command currently exposes root help and version output. Later Phase 1A tasks will add `docsctl validate`, and Phase 1E will add `docsctl publish`.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Begin Phase 1A implementation task by task, starting with the CLI entrypoint.

**Inferred user intent:** Establish a concrete command surface before implementing validation and registry upload behavior.

### What I did

- Added `cmd/docsctl/main.go`.
- Created a Cobra root command with:
  - `Use: docsctl`
  - version wiring through a `version` variable;
  - root help text explaining validation and publishing intent.
- Ran focused validation commands.
- Checked off the first Phase 1A task.

Commands run:

```bash
go test ./cmd/docsctl
go run ./cmd/docsctl --version
```

### Why

A separate CLI avoids overloading `glaze serve` or `glaze help`. Package CI should call a tool with publishing semantics, not the general formatting CLI.

### What worked

- The new command package builds.
- `docsctl --version` prints `docsctl version dev`.

### What didn't work

N/A.

### What I learned

The standalone Cobra entrypoint is minimal enough that future command tests can call `newRootCommand()` directly without invoking a subprocess.

### What was tricky to build

N/A. The key decision was naming and scope, not implementation complexity.

### What warrants a second pair of eyes

- Whether `docsctl` should live permanently in this repo or eventually move into a smaller docs publishing repo/tool.

### What should be done in the future

- Add `docsctl validate --package --version --file --json` after implementing validation helpers.
- Add linker-based version injection if `docsctl` becomes a released binary.

### Code review instructions

Review:

- `cmd/docsctl/main.go`

Validate:

```bash
go test ./cmd/docsctl
go run ./cmd/docsctl --help
go run ./cmd/docsctl --version
```

### Technical details

The root command returns `cmd.Help()` when invoked without subcommands, so plain `docsctl` is informative rather than a no-op.

## Step 7: Add package/version validation helpers

The next Phase 1A task was to define the naming rules that every later publishing path will rely on. Package and version names are both identities and filesystem path segments, so validation has to reject path traversal, empty values, whitespace-padded values, separators, and unsafe leading characters before registry or PVC code starts constructing paths.

This is a small package, but it is security-relevant: every future upload endpoint and directory publisher should call these helpers before authorizing or writing files.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Continue Phase 1A task execution with the package/version/path validation helper task.

**Inferred user intent:** Build safe foundations before accepting package-uploaded SQLite files.

### What I did

- Added `pkg/help/publish/validation.go`.
- Implemented:
  - `ValidatePackageName`
  - `ValidateVersion`
  - `ValidatePackageVersion`
  - `DBFileName`
  - `PackageVersionDir`
  - `PackageVersionDBPath`
- Added tests in `pkg/help/publish/validation_test.go` for valid names and unsafe values.
- Ran focused tests.
- Checked off the Phase 1A validation-helper task.

Commands run:

```bash
gofmt -w pkg/help/publish/validation.go pkg/help/publish/validation_test.go
go test ./pkg/help/publish
```

### Why

The registry must not build filesystem paths from unvalidated request parameters. A malicious package like `../../glazed` or version `v1/v2` should be rejected before authorization and storage logic.

### What worked

- Focused tests pass.
- The helper functions produce the canonical relative DB path `pinocchio/v1.2.3/pinocchio.db`.

### What didn't work

N/A.

### What I learned

The package and version validation rules should be similar but not identical: versions allow `+` for build metadata, while package names do not.

### What was tricky to build

The main subtlety was making the helpers path-segment oriented rather than URL oriented. They reject both `/` and `\\` separators so the same validation is safe on Linux and on developer machines that may use Windows path conventions.

### What warrants a second pair of eyes

- Whether uppercase letters should remain allowed. They are currently allowed for compatibility and display flexibility, but the registry may later choose to canonicalize package names to lowercase.
- Whether mutable versions like `main` and `current` should be allowed in Phase 1. They are currently allowed by the generic version rules.

### What should be done in the future

- Reuse these helpers in `docsctl validate`, `docsctl publish`, registry authorization, and `DirectoryPackageStore` path construction.

### Code review instructions

Review:

- `pkg/help/publish/validation.go`
- `pkg/help/publish/validation_test.go`

Validate:

```bash
go test ./pkg/help/publish
```

### Technical details

Package names match:

```text
^[A-Za-z0-9][A-Za-z0-9._-]*$
```

Versions match:

```text
^[A-Za-z0-9][A-Za-z0-9._+-]*$
```

## Step 8: Add SQLite help DB validation

The next Phase 1A slice added the local validator that `docsctl validate`, `docsctl publish`, and the registry upload endpoint will share. It opens a candidate help export read-only, verifies that it has a usable `sections` table, rejects empty or duplicate slugs, rejects empty databases, and reports warnings when embedded package metadata differs from the requested publish identity.

This is the core safety gate for Phase 1. Publishing should not write anything into the package root until this validator succeeds.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Continue Phase 1A by implementing the SQLite help DB validator and its unit tests.

**Inferred user intent:** Ensure uploaded package docs are structurally valid before registry storage or server reload is built.

### What I did

- Added `pkg/help/publish/sqlite_validator.go`.
- Implemented `ValidateSQLiteHelpDB(ctx, path, opts)`.
- Added `SQLiteValidationOptions` and `SQLiteValidationResult`.
- The validator checks:
  - non-empty file path;
  - safe package/version options;
  - read-only SQLite open;
  - required `sections` table;
  - required `slug` and `title` columns;
  - non-zero section count;
  - no empty slugs;
  - no duplicate slugs;
  - package metadata mismatch warnings.
- Added `pkg/help/publish/sqlite_validator_test.go`.
- Covered valid DBs, missing files, non-SQLite files, missing `sections`, empty DBs, missing slug column, empty slugs, duplicate slugs, invalid publish names, and metadata warnings.
- Checked off the validator implementation and unit test tasks.

Commands run:

```bash
gofmt -w pkg/help/publish/sqlite_validator.go pkg/help/publish/sqlite_validator_test.go
go test ./pkg/help/publish
```

### Why

The publish registry must reject bad docs before atomic publication. This validator gives both local CLI and server upload paths the same behavior.

### What worked

- Focused publish package tests pass.
- Test fixtures can create intentionally malformed SQLite DBs that bypass the production store constraints, which lets the validator prove it catches duplicates and empty slugs.

### What didn't work

N/A.

### What I learned

Opening SQLite with a `file:` DSN plus `mode=ro` and `_query_only=true` is a good default for upload validation. The validator should inspect uploads as data, not mutate them.

### What was tricky to build

The production Glazed store schema has uniqueness constraints, but uploaded DBs are untrusted and may have older, hand-written, or corrupt schemas. The tests therefore create minimal SQLite fixtures directly instead of relying only on the production store helper.

### What warrants a second pair of eyes

- Whether `slug` and `title` are sufficient required columns, or whether Phase 1 should also require the full current Glazed schema.
- Whether package metadata mismatches should remain warnings or become hard errors for Phase 1.

### What should be done in the future

- Wire this validator into `docsctl validate` next.
- Reuse this validator in the registry upload handler before `DirectoryPackageStore` writes to the package root.

### Code review instructions

Review:

- `pkg/help/publish/sqlite_validator.go`
- `pkg/help/publish/sqlite_validator_test.go`

Validate:

```bash
go test ./pkg/help/publish
```

### Technical details

Read-only SQLite DSN shape:

```text
file:/absolute/path/help.db?mode=ro&_query_only=true
```

## Step 9: Wire `docsctl validate`

This step connected the validation library to the new `docsctl` CLI. Package maintainers can now validate a generated help SQLite export locally before any registry upload exists. The command supports both human-readable output and JSON output for CI.

This completes the Phase 1A validation slice: we have a CLI entrypoint, safe package/version path validation, a read-only SQLite validator, unit tests, command tests, and a smoke run against an existing Pinocchio export.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Finish the Phase 1A local validation command and document the validation slice.

**Inferred user intent:** Give package CI a concrete first tool command that can be tested independently from the future registry.

### What I did

- Added `cmd/docsctl/validate.go`.
- Wired `newValidateCommand()` into the docsctl root command.
- Added flags:
  - `--package`
  - `--version`
  - `--file`
  - `--json`
- Added command tests in `cmd/docsctl/validate_test.go` covering:
  - text success output;
  - JSON success output;
  - validation failure;
  - required flag failure.
- Ran focused tests.
- Smoke-tested against the existing Pinocchio smoke DB.
- Checked off the Phase 1A `docsctl validate`, command test, and validation-slice documentation tasks.

Commands run:

```bash
gofmt -w cmd/docsctl/main.go cmd/docsctl/validate.go cmd/docsctl/validate_test.go
go test ./cmd/docsctl ./pkg/help/publish
go run ./cmd/docsctl validate --package pinocchio --version vtest --file /tmp/glazed-multi-help-smoke/pinocchio/vtest/pinocchio.db --json
```

Smoke output included:

```json
{
  "path": "/tmp/glazed-multi-help-smoke/pinocchio/vtest/pinocchio.db",
  "packageName": "pinocchio",
  "version": "vtest",
  "sectionCount": 69,
  "slugCount": 69
}
```

### Why

Local validation should be available before registry upload. It gives package maintainers fast feedback and gives CI a cheap preflight step.

### What worked

- Command tests pass.
- The command validates a real exported Pinocchio help DB and reports 69 sections / 69 slugs.
- JSON output is suitable for CI parsing.

### What didn't work

N/A.

### What I learned

The existing smoke DBs are useful as realistic fixtures for command-level smoke testing, even though unit tests should keep using isolated temp DBs.

### What was tricky to build

The command needs to use Cobra's output streams (`cmd.OutOrStdout`) so tests can capture output without subprocesses. This also keeps the command reusable for future tests.

### What warrants a second pair of eyes

- Whether `--version` should remain required in Phase 1. The current publishing model is versioned by design, so the command requires it.
- Whether JSON output should include warnings as strings only or structured warning codes later.

### What should be done in the future

- Start Phase 1B with `PublisherAuth`, token hashing, and static package-token authorization.
- Later, `docsctl publish --dry-run` should call the same validation path.

### Code review instructions

Review:

- `cmd/docsctl/main.go`
- `cmd/docsctl/validate.go`
- `cmd/docsctl/validate_test.go`
- `pkg/help/publish/sqlite_validator.go`

Validate:

```bash
go test ./cmd/docsctl ./pkg/help/publish
go run ./cmd/docsctl validate --package pinocchio --version vtest --file /tmp/glazed-multi-help-smoke/pinocchio/vtest/pinocchio.db --json
```

### Technical details

Human output shape:

```text
OK: /path/help.db is a valid Glazed help database for pinocchio@vtest (69 sections, 69 slugs)
```

JSON output shape uses `SQLiteValidationResult`.

## Step 10: Add static package-token authorization core

With local DB validation in place, I started Phase 1B by adding the registry authorization abstraction and the first implementation: static package tokens represented as SHA-256 hashes. This lets the future registry authorize a publish request without storing or logging raw tokens.

The important invariant is now encoded in code and tests: one token hash maps to one package, and a token for package A cannot publish package B.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Continue into Phase 1B with the auth interface, token hashing, static package-token implementation, and tests.

**Inferred user intent:** Build the Phase 1 Vault-static-token model as a reusable auth core before the registry HTTP layer exists.

### What I did

- Added `pkg/help/publish/auth.go`.
- Defined:
  - `PublisherAuth`
  - `PublishRequest`
  - `PublisherIdentity`
  - `StaticPublisherToken`
  - `StaticTokenAuth`
- Implemented:
  - `HashPublishToken`
  - `NormalizeTokenHash`
  - `ConstantTimeTokenHashEqual`
  - `StaticTokenAuth.AuthorizePublish`
- Added sentinel errors:
  - `ErrUnauthorized`
  - `ErrForbidden`
- Added `pkg/help/publish/auth_test.go` covering:
  - matching token/package success;
  - unknown token rejection;
  - empty token rejection;
  - package A token rejected for package B;
  - invalid request rejection;
  - duplicate token hash rejection;
  - token hash normalization;
  - constant-time equality behavior.
- Checked off the first four Phase 1B auth tasks.

Commands run:

```bash
gofmt -w pkg/help/publish/auth.go pkg/help/publish/auth_test.go
go test ./pkg/help/publish
```

### Why

The registry upload handler should not own auth details directly. A `PublisherAuth` interface lets Phase 1 use static token hashes, Phase 2 use Vault capabilities, and Phase 3 use signed JWTs without rewriting upload handling.

### What worked

- Focused publish package tests pass.
- Package-scoped denial is covered directly by tests.
- Duplicate token hashes are rejected at construction time so one token cannot ambiguously map to multiple packages.

### What didn't work

N/A.

### What I learned

Iterating over token hashes instead of doing a direct map lookup makes it straightforward to use constant-time comparison for the presented token hash. The token catalog will be small in Phase 1, so this is acceptable.

### What was tricky to build

The subtle part is error semantics. Unknown or missing tokens return `ErrUnauthorized`; recognized tokens used for the wrong package return `ErrForbidden`. This distinction is useful internally and in tests, but registry HTTP responses should still avoid leaking too much information to callers.

### What warrants a second pair of eyes

- Whether the registry should collapse all auth failures into a generic 403 externally.
- Whether SHA-256 token hashes are sufficient or whether Phase 1 should use HMAC with a server-side pepper.

### What should be done in the future

- Add Vault token-hash record structs and catalog reload behavior next.
- Wire `StaticTokenAuth` into the registry upload endpoint when Phase 1C begins.

### Code review instructions

Review:

- `pkg/help/publish/auth.go`
- `pkg/help/publish/auth_test.go`

Validate:

```bash
go test ./pkg/help/publish
```

### Technical details

Stored token hashes use this format:

```text
sha256:<64 hex chars>
```
