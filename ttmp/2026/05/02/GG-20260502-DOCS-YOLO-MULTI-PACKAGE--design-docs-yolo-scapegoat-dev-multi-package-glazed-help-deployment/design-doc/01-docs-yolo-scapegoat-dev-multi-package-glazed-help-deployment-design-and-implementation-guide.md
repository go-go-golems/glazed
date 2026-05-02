---
Title: docs.yolo.scapegoat.dev multi-package Glazed help deployment design and implementation guide
Ticket: GG-20260502-DOCS-YOLO-MULTI-PACKAGE
Status: active
Topics:
    - glazed
    - docs
    - deploy
    - kubernetes
    - gitops
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - /home/manuel/code/wesen/obsidian-vault/Projects/2026/05/02/ARTICLE - Deploying Glazed Help Browser to Argo CD - Production Deep Dive.md
Summary: "Design for a docs.yolo.scapegoat.dev Glazed documentation registry that can serve many packages and versions from uploaded SQLite help exports."
LastUpdated: 2026-05-02T13:45:00-04:00
WhatFor: "Use when implementing a production multi-package Glazed help browser with uploadable package docs and reloadable serving."
WhenToUse: "Before changing Glazed serve, k3s GitOps manifests, package CI workflows, or object/PVC storage for shared docs hosting."
---

# docs.yolo.scapegoat.dev multi-package Glazed help deployment design and implementation guide

## Executive summary

We already have a production Glazed help browser deployed at `https://glaze.docs.scapegoat.dev`. That deployment proves the base path: a `glaze serve` process inside a GHCR image is reconciled by Argo CD into the Hetzner k3s cluster, exposed through Traefik, and protected by cert-manager. The current deployment is intentionally simple: one pod, no persistent volume, no external content source, and built-in Glazed docs loaded at process startup.

The requested next system is different. `https://docs.yolo.scapegoat.dev` should be a shared documentation hub where many packages can publish their own Glazed help SQLite export under a package name and version. The help browser should then list packages and versions, let users switch between them, and eventually reload content without rebuilding or redeploying the browser image for every package documentation update.

The recommended architecture is a **documentation registry service plus reloadable Glazed help server**:

1. Packages generate `help.db` files in CI using `glaze help export --format sqlite` or an equivalent package-specific command.
2. CI uploads those SQLite files to a content registry endpoint or object store path using a stable layout such as `packages/<package>/<version>/<package>.db`.
3. A cluster-side syncer materializes the registry into a read-only directory on a persistent volume.
4. `glaze serve --from-sqlite-dir /var/lib/glazed-docs/packages` loads that directory using the existing multi-package loader.
5. A new reload mechanism rebuilds an in-memory help store in the background and atomically swaps the handler's store pointer after validation.
6. As a safe Phase 1, the cluster can use rollout restarts instead of hot reload; hot reload should be Phase 2.

The most important implementation choice is to keep the package/version identity in the Glazed data model and API, not in the deployment layer. That is already how the current code is shaped. `model.Section` has `PackageName` and `PackageVersion`; the store has a composite uniqueness constraint on `(package_name, package_version, slug)`; `/api/packages` lists package/version groups; `/api/sections` accepts `package` and `version` query parameters. The deployment should build on these contracts rather than inventing a separate routing scheme per package.

## Problem statement and scope

### User-facing goal

Create a production deployment at:

```text
https://docs.yolo.scapegoat.dev
```

The site should serve documentation for many packages. A package should be able to publish one or more versioned Glazed SQLite help exports. The browser should expose those packages through the existing package/version selector and documentation tree.

### Operator-facing goal

Operators should not rebuild the `glazed` container image every time a package publishes docs. Package docs should be external content. The operator should be able to answer these questions quickly:

- Which packages are currently published?
- Which versions exist for a package?
- Which uploaded file produced the currently served docs?
- Did the latest upload validate before becoming visible?
- Can we roll back a bad docs upload?
- Can the server reload without dropping traffic?

### Package maintainer-facing goal

A maintainer of another package should have a small publishing contract:

```bash
my-tool help export --format sqlite --output-path ./dist/help/my-tool.db

docsctl publish \
  --package my-tool \
  --version v1.4.2 \
  --file ./dist/help/my-tool.db
```

The maintainer should not need direct Kubernetes access. The publishing workflow should work from GitHub Actions, a local machine, or an internal CI runner.

### Scope

In scope:

- Multi-package Glazed help serving at `docs.yolo.scapegoat.dev`.
- Versioned SQLite help export storage.
- Upload/publish workflows.
- Runtime refresh strategies, including hot reload and restart-triggered reload.
- GitOps and Kubernetes manifest changes needed for the new deployment.
- Security and validation model for uploaded docs.
- Intern-friendly implementation plan.

Out of scope for the first production milestone:

- Multi-tenant arbitrary HTML hosting.
- User accounts in the browser UI.
- Full search indexing service outside SQLite/Glazed.
- Wildcard TLS platform migration unless needed for future subdomains. The previous deployment intentionally used concrete HTTP-01 TLS for `glaze.docs.scapegoat.dev`.

## Current-state architecture with evidence

### Existing production deployment

The current deployed app is `glaze-docs` in the k3s GitOps repository. Argo CD watches `gitops/kustomize/glaze-docs` and syncs it into the `glaze-docs` namespace. Evidence:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml:10-16` points Argo CD at namespace `glaze-docs` and path `gitops/kustomize/glaze-docs`.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml:17-23` enables automated sync, prune, self-heal, namespace creation, and server-side apply.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml:23-30` runs `ghcr.io/go-go-golems/glazed:sha-2bc01c9` with args `serve --address :8088`.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml:34-45` probes `/api/health` for readiness and liveness.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/service.yaml:15-18` maps Service port `80` to the named pod port `http`.
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml:13-18` currently issues TLS and routes host `glaze.docs.scapegoat.dev`.

The production deep dive in the Obsidian vault records the important operational decisions: immutable GHCR image tags, CGO-enabled Debian image because `go-sqlite3` is required, `.dockerignore` to keep local `node_modules` out of builds, and concrete TLS rather than wildcard TLS.

### Existing Glazed serving model

`glaze serve` already has the key primitives needed for a shared docs server:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go:36-44` defines `ServeSettings` with `FromJSON`, `FromSQLite`, `FromSQLiteDir`, `FromGlazedCmd`, and `WithEmbedded`.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go:55-75` documents the external source modes and the package/version directory layout.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go:140-163` builds loaders and loads all configured sources before starting HTTP serving.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go:171-188` maps `--from-sqlite-dir` to `SQLiteDirLoader`.

This means the current process is **startup-loaded**. It does not continuously watch a directory, poll a bucket, or reload a store after the HTTP handler has started.

### Existing package/version directory loader

`SQLiteDirLoader` already accepts exactly the kind of versioned directory layout needed for a package docs registry:

```text
X.db       -> package X, no version
X/X.db     -> package X, no version
X/Y/X.db   -> package X, version Y
```

Evidence:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/loader/sources.go:202-205` documents accepted layouts.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/loader/sources.go:209-222` discovers package DBs and loads each as a `PackageRef`.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/loader/sources.go:228-260` walks the directory and infers package/version from relative path shape.

This is a strong reason to use a directory materialization layer in Kubernetes. Any storage backend can be synced into this layout, and Glazed can consume it with the existing flag.

### Existing package/version API

The server API already exposes packages and version filtering:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go:64-68` registers `/api/health`, `/api/packages`, `/api/sections`, `/api/sections/search`, and `/api/sections/{slug}`.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go:115-152` groups store package rows into the `/api/packages` response and chooses defaults.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go:272-306` parses `package` and `version` query params and filters by `store.InPackageVersion`.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/types.go:127-140` defines `PackageSummary` and `ListPackagesResponse`.

The frontend package selector and tree navigation work should therefore carry over if the server sees many packages.

### Existing store identity model

The store already supports duplicate slugs across packages and versions:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/store/store.go:109-126` creates `sections` with `package_name`, `package_version`, and `UNIQUE(package_name, package_version, slug)`.
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/store/store.go:649-668` lists package/version groups and section counts.

This matters because a shared docs site will naturally contain repeated slugs like `intro`, `getting-started`, `configuration`, and `examples` across packages.

### Current gap

The current code can load multiple package DBs at startup, but it cannot ingest new DBs after startup without process restart. It also lacks:

- upload authentication and authorization;
- package ownership registry;
- upload validation and quarantine;
- atomic publication semantics;
- live reload endpoint or watcher;
- content provenance metadata in `/api/packages`;
- operational metrics for reloads, package counts, failed publishes, and last successful sync.

## Target architecture

### Recommended architecture: registry + materialized directory + reloadable Glazed server

```mermaid
flowchart TD
    Maintainer[Package maintainer CI] --> Export[package help export --format sqlite]
    Export --> Publish[docsctl publish or HTTP upload]
    Publish --> Registry[Docs registry API]
    Registry --> Validate[Validate SQLite schema and metadata]
    Validate --> ObjectStore[(Object store or OCI registry)]
    Validate --> Manifest[(Package/version manifest)]

    Syncer[Cluster syncer sidecar or CronJob] --> ObjectStore
    Syncer --> Manifest
    Syncer --> PVC[(docs PVC)]
    PVC --> Dir[/var/lib/glazed-docs/packages/pkg/version/pkg.db/]

    Dir --> Glaze[glaze serve --from-sqlite-dir]
    Glaze --> Reload[Reload manager]
    Reload --> Store[Atomic in-memory Store pointer]
    Store --> API[/api/packages and /api/sections]
    API --> Browser[docs.yolo.scapegoat.dev]
```

The architecture separates four responsibilities:

1. **Publication**: how packages upload docs.
2. **Storage**: where uploaded versioned DBs live durably.
3. **Materialization**: how the server sees a local directory matching `SQLiteDirLoader`.
4. **Serving**: how Glazed loads, validates, and swaps content.

This separation keeps the serving code simple and makes storage replaceable. For example, the first implementation can use a Git-backed or PVC-backed registry, and a later one can move to R2/S3 without changing the browser API.

## Storage and reload options

### Option A — GitOps-only docs repository

Package CI opens a PR against a repository like `go-go-golems-docs-registry` containing:

```text
packages/glazed/current/glazed.db
packages/pinocchio/v0.8.1/pinocchio.db
manifest.yaml
```

Argo CD syncs that repo into the cluster through a ConfigMap-like generator or an init container that clones the repo.

Pros:

- excellent audit trail;
- easy rollback by Git revert;
- no new upload API;
- fits existing Argo CD mental model.

Cons:

- Git is poor for large binary SQLite churn;
- PR workflow may be too slow for frequent package releases;
- repo size grows forever;
- content update still needs pod restart or sync trigger.

Use this if package docs are small and releases are infrequent. Do not use it as the final high-scale design.

### Option B — S3/R2/MinIO bucket plus syncer

Package CI uploads to a bucket path:

```text
s3://glazed-docs/packages/<package>/<version>/<package>.db
s3://glazed-docs/manifests/catalog.json
```

A cluster syncer periodically or event-driven downloads changed files into a PVC.

Pros:

- durable and cheap;
- good fit for binary artifacts;
- easy presigned upload URLs;
- supports object versioning and lifecycle policies;
- can work with Cloudflare R2, AWS S3, MinIO, Backblaze B2, or DigitalOcean Spaces.

Cons:

- needs credentials, bucket policy, and sync logic;
- event triggers differ by provider;
- local development needs a mock or MinIO.

This is the recommended durable storage model if we want provider-neutral package publication.

### Option C — OCI artifact registry

Each package publishes docs as an OCI artifact to GHCR:

```text
ghcr.io/go-go-golems/docs/glazed:v1.2.3
ghcr.io/go-go-golems/docs/pinocchio:v0.8.1
```

The artifact contains:

```text
/package.db
/metadata.json
```

A cluster syncer uses `oras pull` to materialize selected artifacts into the docs directory.

Pros:

- reuses GHCR and GitHub auth;
- versions/tags are native;
- immutable digests are excellent for provenance;
- package CI already knows how to publish to GHCR.

Cons:

- less obvious than a bucket for browsing raw files;
- requires ORAS tooling or Go registry client;
- catalog discovery is harder unless we maintain an index.

This is the most elegant option if package releases already publish container images and we want docs to be release artifacts.

### Option D — Direct upload API backed by PVC

Expose a registry service at an internal or protected endpoint:

```http
PUT /api/packages/{package}/versions/{version}/sqlite
Authorization: Bearer <token>
Content-Type: application/vnd.sqlite3
```

The registry writes to a PVC with atomic rename:

```text
/tmp/uploads/<uuid>.db
packages/<package>/<version>/<package>.db
```

Pros:

- simplest mental model for package maintainers;
- immediate validation;
- easy to trigger Glazed reload after successful publish;
- no external bucket required.

Cons:

- PVC becomes the durable source of truth unless backed up;
- upload API must handle auth, quotas, validation, rollback, and disaster recovery;
- harder to share across clusters.

This is attractive as a fast MVP but should be paired with Velero/restic backups or a mirrored object store.

### Option E — Release webhooks that create Kubernetes Jobs

Package CI sends a webhook with artifact URL and checksum. The cluster creates a short-lived Job that downloads, validates, writes to PVC, and triggers reload.

Pros:

- no long-running sync loop;
- clear per-publish logs;
- easy retry model;
- integrates with GitHub release events.

Cons:

- more Kubernetes machinery;
- webhook auth must be correct;
- jobs need controlled RBAC.

This pairs well with Option B or C.

### Option F — NATS/queue-driven publish pipeline

Package CI publishes a message like:

```json
{
  "package": "pinocchio",
  "version": "v0.8.1",
  "artifactUrl": "oci://ghcr.io/...",
  "sha256": "..."
}
```

A worker consumes, validates, stores, and triggers reload.

Pros:

- robust event-driven pipeline;
- can add retries, dead-letter queues, and notifications;
- good if a broader release automation platform exists.

Cons:

- overkill for the first shared docs deployment;
- introduces another stateful platform dependency.

Use later if package publishing becomes high-volume.

## Recommendation

Use a phased design:

### Phase 1: PVC materialization and rollout restart

Deploy `docs-yolo` as a separate Argo CD app. Mount a PVC at:

```text
/var/lib/glazed-docs/packages
```

Run Glazed with:

```yaml
args:
  - serve
  - --address
  - :8088
  - --from-sqlite-dir
  - /var/lib/glazed-docs/packages
```

Publish package DBs with an operator-controlled job or manual `kubectl cp`/sync script at first. After each publish, restart the deployment:

```bash
kubectl -n docs-yolo rollout restart deployment/docs-yolo
```

This uses existing `SQLiteDirLoader` and does not require risky live-reload code on day one.

### Phase 2: Registry service and validation

Add a small `docs-registry` service that accepts uploads, validates SQLite files, writes them atomically into PVC or object storage, and records metadata. The registry should be separate from the public help browser so upload auth and public read paths do not mix.

### Phase 3: Hot reload in Glazed

Modify Glazed server internals to support an atomic store swap. After a successful registry publish, call:

```http
POST http://docs-yolo.docs-yolo.svc.cluster.local/admin/reload
Authorization: Bearer <reload-token>
```

The reload handler should build a new store from `/var/lib/glazed-docs/packages`, validate it, then atomically replace the store used by request handlers.

### Phase 4: Object/OCI-backed source of truth

Move durable source of truth to R2/S3/MinIO or GHCR OCI artifacts. The PVC becomes a cache/materialized view. This enables disaster recovery and a cleaner package maintainer workflow.

## Detailed system design

### Content layout

Use the existing `SQLiteDirLoader` layout as the local serving contract:

```text
/var/lib/glazed-docs/packages/
  glazed/
    current/
      glazed.db
    v1.0.0/
      glazed.db
  pinocchio/
    v0.8.1/
      pinocchio.db
  sqleton/
    v0.3.0/
      sqleton.db
```

Rules:

- Directory name `packages/<package>` is the package identity.
- Directory name `packages/<package>/<version>` is the package version.
- DB basename should match package name: `<package>.db`.
- A version named `current` is allowed only if we decide it is a mutable channel; immutable semantic versions should remain the default.
- Uploads must be written to a temporary path and atomically renamed into place.

### Package manifest

Add a manifest next to the DB files:

```yaml
apiVersion: docs.yolo.scapegoat.dev/v1alpha1
kind: DocsCatalog
packages:
  - name: glazed
    displayName: Glazed
    owner: go-go-golems
    versions:
      - version: v1.0.0
        path: packages/glazed/v1.0.0/glazed.db
        sha256: "..."
        source:
          type: github-actions
          repo: go-go-golems/glazed
          commit: 2bc01c9
          workflowRun: "123456"
        publishedAt: "2026-05-02T17:00:00Z"
```

The first Glazed server does not need to read the manifest. The registry and operators do. Later, `/api/packages` can include provenance fields if the UI needs them.

### Upload API sketch

```http
POST /v1/packages/{package}/versions/{version}:prepare
Authorization: Bearer <publisher-token>
Content-Type: application/json

{
  "sha256": "...",
  "sizeBytes": 123456,
  "source": {
    "repo": "go-go-golems/pinocchio",
    "commit": "abc123",
    "workflowRun": "987654"
  }
}
```

Response for bucket-backed uploads:

```json
{
  "uploadUrl": "https://...presigned...",
  "objectKey": "packages/pinocchio/v0.8.1/pinocchio.db",
  "expiresAt": "2026-05-02T17:10:00Z"
}
```

Finalize:

```http
POST /v1/packages/{package}/versions/{version}:finalize
Authorization: Bearer <publisher-token>
Content-Type: application/json

{
  "objectKey": "packages/pinocchio/v0.8.1/pinocchio.db",
  "sha256": "..."
}
```

Direct upload variant:

```http
PUT /v1/packages/{package}/versions/{version}/sqlite
Authorization: Bearer <publisher-token>
Content-Type: application/vnd.sqlite3
Digest: sha-256=<base64>
```

### Validation contract

Every upload must pass validation before publication:

1. File opens as SQLite.
2. Expected `sections` table exists.
3. `package_name` and `package_version` are either empty or match the path identity after import.
4. At least one section exists.
5. Slugs are non-empty.
6. Composite identity `(package_name, package_version, slug)` has no duplicates.
7. Optional max size and max section count limits pass.
8. The file checksum matches the declared checksum.

Pseudocode:

```go
func ValidateUpload(path, packageName, version string) (*ValidationResult, error) {
    db := openSQLiteReadOnly(path)
    defer db.Close()

    if !tableExists(db, "sections") {
        return nil, error("missing sections table")
    }

    rows := querySections(db)
    if len(rows) == 0 {
        return nil, error("empty docs export")
    }

    seen := map[string]bool{}
    for _, row := range rows {
        if row.Slug == "" {
            return nil, error("empty slug")
        }
        key := packageName + "\x00" + version + "\x00" + row.Slug
        if seen[key] {
            return nil, error("duplicate slug")
        }
        seen[key] = true
    }

    return &ValidationResult{SectionCount: len(rows)}, nil
}
```

### Reload manager design

Current `HandlerDeps` effectively freezes a `Store` pointer when `NewHandler` is built. For hot reload, introduce a provider abstraction:

```go
type StoreProvider interface {
    CurrentStore() *store.Store
}

type AtomicStoreProvider struct {
    v atomic.Value // stores *store.Store
}

func (p *AtomicStoreProvider) CurrentStore() *store.Store {
    return p.v.Load().(*store.Store)
}

func (p *AtomicStoreProvider) Swap(next *store.Store) {
    p.v.Store(next)
}
```

Then handlers use `h.store()` instead of `h.deps.Store`:

```go
func (h *Handler) store() *store.Store {
    if h.deps.StoreProvider != nil {
        return h.deps.StoreProvider.CurrentStore()
    }
    return h.deps.Store
}
```

Reload flow:

```go
func (r *ReloadManager) Reload(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    nextStore := store.New(":memory:")
    nextHS := &help.HelpSystem{Store: nextStore}

    loader := &loader.SQLiteDirLoader{Roots: []string{r.Root}}
    if err := loader.Load(ctx, nextHS); err != nil {
        nextStore.Close()
        return err
    }

    count, err := nextStore.Count(ctx)
    if err != nil || count == 0 {
        nextStore.Close()
        return fmt.Errorf("refusing to publish empty docs catalog")
    }

    old := r.provider.CurrentStore()
    r.provider.Swap(nextStore)
    scheduleClose(old, 30*time.Second)
    r.metrics.LastReloadSuccess.Set(time.Now())
    return nil
}
```

Important invariants:

- Never mutate the active store during reload.
- Never swap to an empty or invalid store.
- Only one reload runs at a time.
- Keep the old store briefly so in-flight requests can finish.
- Admin reload endpoint must not be public.

### Admin API sketch

```http
GET /admin/catalog/status
Authorization: Bearer <admin-token>
```

```json
{
  "ok": true,
  "root": "/var/lib/glazed-docs/packages",
  "lastReloadAt": "2026-05-02T17:00:00Z",
  "lastReloadDurationMs": 431,
  "lastReloadError": "",
  "packageCount": 14,
  "versionCount": 38,
  "sectionCount": 1842
}
```

```http
POST /admin/reload
Authorization: Bearer <admin-token>
```

```json
{
  "ok": true,
  "sectionCount": 1842,
  "packageCount": 14,
  "versionCount": 38,
  "durationMs": 431
}
```

### Kubernetes design

Create a new app, not a mutation of the existing `glaze-docs`, so the current production site remains stable.

Suggested namespace:

```text
docs-yolo
```

Suggested objects:

```text
gitops/applications/docs-yolo.yaml
gitops/kustomize/docs-yolo/deployment.yaml
gitops/kustomize/docs-yolo/service.yaml
gitops/kustomize/docs-yolo/ingress.yaml
gitops/kustomize/docs-yolo/pvc.yaml
gitops/kustomize/docs-yolo/registry-deployment.yaml   # Phase 2
gitops/kustomize/docs-yolo/registry-service.yaml      # Phase 2
gitops/kustomize/docs-yolo/syncer-cronjob.yaml        # if bucket-backed
```

Deployment sketch for Phase 1:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docs-yolo
spec:
  replicas: 1
  template:
    spec:
      enableServiceLinks: false
      containers:
        - name: docs-yolo
          image: ghcr.io/go-go-golems/glazed:sha-<new-commit>
          args:
            - serve
            - --address
            - :8088
            - --from-sqlite-dir
            - /var/lib/glazed-docs/packages
          ports:
            - name: http
              containerPort: 8088
          volumeMounts:
            - name: docs-packages
              mountPath: /var/lib/glazed-docs/packages
              readOnly: true
          readinessProbe:
            httpGet:
              path: /api/health
              port: http
      volumes:
        - name: docs-packages
          persistentVolumeClaim:
            claimName: docs-yolo-packages
```

Ingress sketch:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: docs-yolo
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: traefik
  tls:
    - hosts:
        - docs.yolo.scapegoat.dev
      secretName: docs-yolo-tls
  rules:
    - host: docs.yolo.scapegoat.dev
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: docs-yolo
                port:
                  number: 80
```

The previous deployment used concrete TLS with HTTP-01. Do the same here unless the broader DNS-01 platform work is already complete.

## Security model

### Public surface

Public:

- `GET /`
- `GET /api/health`
- `GET /api/packages`
- `GET /api/sections`
- `GET /api/sections/{slug}`

Private:

- upload API;
- registry admin API;
- reload API;
- syncer status if it exposes credentials or source metadata.

### Publisher authorization

Use per-package publisher tokens or GitHub OIDC claims. The registry must enforce that `pinocchio` CI cannot overwrite `glazed` docs unless explicitly authorized.

Minimal token claims:

```json
{
  "sub": "repo:go-go-golems/pinocchio:ref:refs/tags/v0.8.1",
  "package": "pinocchio",
  "permissions": ["publish:pinocchio"],
  "exp": 1777740000
}
```

If using GitHub OIDC, the registry validates the token issuer and repository claim rather than storing long-lived CI secrets.

### SQLite safety

SQLite docs exports are data, but they are still untrusted files. Validation should:

- open uploads read-only;
- never execute arbitrary SQL from upload metadata;
- use prepared statements;
- impose size limits;
- run validation in a separate process or container if we start accepting third-party packages;
- never serve the raw uploaded DB over the public browser path unless intentionally exposed.

## Implementation guide for a new intern

### Phase 0 — Learn the current pieces

Read these files first:

1. `pkg/help/server/serve.go` — how `glaze serve` loads sources and starts HTTP.
2. `pkg/help/loader/sources.go` — how SQLite DBs are discovered and imported.
3. `pkg/help/store/store.go` — how package/version/slug identity is stored.
4. `pkg/help/server/handlers.go` — public API routes.
5. `web/src/App.tsx` — package/version selector and current browser state.
6. `gitops/kustomize/glaze-docs/*.yaml` in the k3s repo — production deployment pattern.
7. The Obsidian deep dive article — what went wrong in the first production deployment.

Checkpoint question: can you explain why `:memory:` SQLite required a single DB connection and why the production image needs CGO? If not, reread the deployment article before changing runtime code.

### Phase 1 — Add a docs-yolo GitOps app

1. Copy the existing `glaze-docs` Kustomize package to `docs-yolo`.
2. Change names, labels, namespace, TLS secret, and host to `docs.yolo.scapegoat.dev`.
3. Add a PVC for package DBs.
4. Mount the PVC at `/var/lib/glazed-docs/packages`.
5. Change args to include `--from-sqlite-dir /var/lib/glazed-docs/packages`.
6. Seed the PVC with at least two package/version DBs.
7. Validate:

```bash
kubectl kustomize gitops/kustomize/docs-yolo >/tmp/docs-yolo.yaml
kubectl apply --dry-run=server -f /tmp/docs-yolo.yaml
curl -fsS https://docs.yolo.scapegoat.dev/api/packages | jq .
```

Expected `/api/packages` shape:

```json
{
  "packages": [
    {"name": "glazed", "displayName": "Glazed", "versions": ["v1.0.0"], "sectionCount": 72},
    {"name": "pinocchio", "displayName": "Pinocchio", "versions": ["v0.8.1"], "sectionCount": 69}
  ],
  "defaultPackage": "glazed",
  "defaultVersion": "v1.0.0"
}
```

### Phase 2 — Add a local publisher tool

Create a small `docsctl` CLI or script with commands:

```bash
docsctl validate --package pinocchio --version v0.8.1 --file pinocchio.db
docsctl publish-local --root /var/lib/glazed-docs/packages --package pinocchio --version v0.8.1 --file pinocchio.db
docsctl publish --server https://registry.docs.yolo.scapegoat.dev --package pinocchio --version v0.8.1 --file pinocchio.db
```

Pseudocode for local publish:

```go
func PublishLocal(root, packageName, version, file string) error {
    if err := validateNames(packageName, version); err != nil { return err }
    if err := ValidateUpload(file, packageName, version); err != nil { return err }

    dir := filepath.Join(root, packageName, version)
    tmp := filepath.Join(root, ".incoming", uuid()+".db")
    final := filepath.Join(dir, packageName+".db")

    copyFile(file, tmp)
    fsync(tmp)
    os.MkdirAll(dir, 0755)
    os.Rename(tmp, final)
    writeManifestEntry(...)
    return nil
}
```

### Phase 3 — Add registry service

Start with direct upload to PVC if speed matters. Keep the public browser and private registry as separate Deployments.

Registry endpoints:

```text
PUT  /v1/packages/{package}/versions/{version}/sqlite
GET  /v1/packages
GET  /v1/packages/{package}/versions/{version}
POST /v1/reload-targets/docs-yolo/reload
```

Validation tests:

- rejects non-SQLite file;
- rejects empty DB;
- rejects DB without `sections`;
- rejects package name path traversal (`../../etc/passwd`);
- accepts valid Glazed help export;
- writes final file atomically;
- does not replace old version on failed validation.

### Phase 4 — Add hot reload

Implement `StoreProvider` and `ReloadManager` in `pkg/help/server`.

Files likely touched:

```text
pkg/help/server/handlers.go
pkg/help/server/serve.go
pkg/help/server/reload.go
pkg/help/server/reload_test.go
pkg/help/loader/sources.go
```

Test cases:

1. Server starts with package A.
2. Write package B DB into temp root.
3. Call reload.
4. `/api/packages` returns package A and B.
5. Corrupt DB does not replace current store.
6. Concurrent `/api/sections` calls during reload do not panic or return partial data.

### Phase 5 — Switch durable storage to object or OCI

If choosing S3/R2/MinIO:

- create bucket;
- create least-privilege upload role;
- add syncer with checksum validation;
- write sync state to PVC;
- trigger reload after successful sync.

If choosing OCI:

- define artifact media type, for example `application/vnd.go-go-golems.glazed.help.sqlite.v1`;
- publish with `oras push`;
- maintain catalog manifest mapping package/version to OCI digest;
- sync with `oras pull`.

## Testing and validation strategy

### Unit tests

Glazed:

```bash
go test ./pkg/help/server ./pkg/help/loader ./pkg/help/store
```

Registry service:

```bash
go test ./cmd/docs-registry ./pkg/docsregistry/...
```

Frontend:

```bash
cd web
pnpm test -- --run
pnpm exec tsc --noEmit
pnpm build
```

### Integration tests

Use a temp directory with two package DBs:

```text
/tmp/docs-test/glazed/v1/glazed.db
/tmp/docs-test/pinocchio/vtest/pinocchio.db
```

Run:

```bash
go run ./cmd/glaze serve --from-sqlite-dir /tmp/docs-test --address :8099
curl -fsS http://127.0.0.1:8099/api/packages | jq .
curl -fsS 'http://127.0.0.1:8099/api/sections?package=pinocchio&version=vtest' | jq '{total}'
```

### Cluster validation

```bash
kubectl -n argocd get application docs-yolo -o wide
kubectl -n docs-yolo get deploy,svc,ingress,pods,pvc,certificate
kubectl -n docs-yolo logs deploy/docs-yolo --tail=100
curl -fsS https://docs.yolo.scapegoat.dev/api/health
curl -fsS https://docs.yolo.scapegoat.dev/api/packages | jq .
```

### Failure drills

1. Upload corrupt DB; verify current docs remain available.
2. Upload empty DB; verify validation rejects it.
3. Delete one object from bucket; verify syncer does not remove live docs without explicit delete manifest.
4. Restart Glazed pod; verify it reloads from PVC.
5. Restore PVC from backup or resync from bucket/OCI.

## Risks and mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Bad upload replaces good docs | Public docs break | Validate before atomic rename; keep previous version; reload only after validation. |
| Package overwrites another package | Integrity issue | Per-package auth and package ownership registry. |
| PVC lost | Docs disappear | Use object/OCI source of truth or backup PVC. |
| Hot reload races with requests | 500s or data race | Atomic store pointer; build new store off-path; swap only after success. |
| Bucket credentials leak | Unauthorized writes | OIDC/presigned URLs; short-lived credentials; least privilege. |
| Large docs DBs slow reload | Longer stale windows | Metrics, max file size, background reload, old store stays active. |
| Ambiguous mutable versions like `latest` | Hard rollbacks | Prefer immutable semver versions; allow mutable channels only with provenance. |
| Public upload endpoint abused | Storage exhaustion | Keep registry private; enforce auth, quotas, max size, and rate limits. |

## Design decisions

### Decision 1: Use local directory layout as the server contract

Rationale: `SQLiteDirLoader` already accepts package/version directory layouts. A local directory lets us change storage backends without changing `glaze serve` semantics.

### Decision 2: Start with restart-based reload before hot reload

Rationale: restart-based reload uses existing code and is operationally safe. Hot reload touches handler/store concurrency and should be tested carefully.

### Decision 3: Keep upload API separate from public browser API

Rationale: public docs serving has simple unauthenticated reads. Uploads need auth, validation, quotas, and audit logs. Mixing them makes the public server more sensitive.

### Decision 4: Prefer immutable versions

Rationale: immutable versions make provenance and rollback understandable. Mutable aliases like `current` can be added later as channels.

### Decision 5: Plan for object/OCI source of truth

Rationale: a PVC alone is not a durable registry. It is acceptable for Phase 1 but should become a cache/materialized view.

## Alternatives considered

See the storage options section for GitOps-only, bucket, OCI artifact, direct PVC upload, webhook Job, and queue-driven designs. The recommended path combines the fastest safe MVP with a migration path to durable artifact storage.

The main rejected final-state design is “only rebuild the Glazed image with all docs embedded.” That would be simple for serving but bad for independent package publishing because every docs update becomes an application release.

## Open questions

1. Should `docs.yolo.scapegoat.dev` include the built-in Glazed docs by default, or only packages uploaded to the registry?
2. Should mutable channels like `latest`, `stable`, and `current` be allowed?
3. Which durable backend should be first-class: Cloudflare R2, DigitalOcean Spaces, MinIO in-cluster, or GHCR OCI artifacts?
4. Do we want public raw DB downloads for debugging, or should raw DBs stay private?
5. Should package docs deletion be allowed, or should versions be append-only with hidden/deprecated state?
6. Should reload be push-triggered by registry, pull-triggered by watcher, or both?

## References

Application repo:

- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/serve.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/handlers.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/server/types.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/loader/sources.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/pkg/help/store/store.go`
- `/home/manuel/workspaces/2026-05-02/multi-package-hosting-glazed/glazed/web/src/App.tsx`

GitOps repo:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/glaze-docs.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/service.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml`

External write-up:

- `/home/manuel/code/wesen/obsidian-vault/Projects/2026/05/02/ARTICLE - Deploying Glazed Help Browser to Argo CD - Production Deep Dive.md`

## Phase 1 addendum: static package publishing tokens stored in Vault

This addendum narrows the first implementation phase. The broader design above describes several storage and reload options. For the first shippable publishing path, use a deliberately small authentication model: **one static publish token per package, stored and rotated through Vault, enforced by the docs registry**.

This is not the final identity model. It is the simplest safe bridge between no upload service and the later GitHub OIDC/Vault flow. It gives every package a separate credential and lets the registry enforce package scoping immediately, while avoiding the operational work of exposing Vault JWT auth to GitHub-hosted runners on day one.

### Phase 1 responsibilities

Phase 1 has four pieces:

1. **Vault-held package token inventory**
   - Vault stores one generated token per package.
   - The raw token is only shown when created or rotated.
   - The registry stores or reads only token hashes where possible.

2. **Package-scoped registry authorization**
   - A token maps to exactly one package name.
   - A `pinocchio` token cannot publish `glazed`.
   - A `glazed` token cannot publish `pinocchio`.

3. **Direct upload to registry or operator-mediated local publish**
   - `docsctl publish --package X --version Y --file X.db --token ...` uploads the SQLite DB.
   - The registry validates the DB before writing it into the package/version directory.

4. **Restart-based reload**
   - After a successful publish, Phase 1 either prints the required rollout restart command or invokes an operator-controlled reload hook.
   - Hot reload is intentionally deferred.

### Phase 1 package maintainer workflow

A package maintainer gets a token through an operator-approved process. In GitHub Actions, that token is stored as a repository secret, for example:

```text
DOCS_YOLO_PUBLISH_TOKEN
```

A release workflow then runs:

```bash
mkdir -p dist/help

pinocchio help export \
  --format sqlite \
  --output-path dist/help/pinocchio.db

docsctl publish \
  --server https://registry.docs.yolo.scapegoat.dev \
  --package pinocchio \
  --version "${GITHUB_REF_NAME}" \
  --file dist/help/pinocchio.db \
  --token "${DOCS_YOLO_PUBLISH_TOKEN}"
```

The registry checks the token before accepting the file.

### Phase 1 Vault layout

Use a predictable Vault path per package. Exact mount names can change to match the existing Vault conventions, but the logical structure should be:

```text
kv/docs-yolo/publishers/<package>
```

Example secret payload:

```json
{
  "package": "pinocchio",
  "token_hash": "sha256:...",
  "created_at": "2026-05-02T18:00:00Z",
  "created_by": "manuel",
  "rotated_at": "",
  "notes": "GitHub Actions secret DOCS_YOLO_PUBLISH_TOKEN in go-go-golems/pinocchio"
}
```

The raw token should not be stored in plaintext if avoidable. A practical first implementation may store it temporarily while bootstrapping, but the desired model is:

```text
operator generates token -> stores hash in Vault -> gives raw token once to package owner
```

### Phase 1 registry token table

The registry needs a fast way to map tokens to packages. Two acceptable implementations exist:

#### Option A: registry reads Vault on startup

On startup, the registry reads all allowed package token hashes from Vault and keeps them in memory:

```go
type StaticPublisher struct {
    Package   string
    TokenHash string
}
```

Pros:

- Vault remains the source of truth.
- Registry does not need a separate database for token metadata.

Cons:

- Token rotation needs registry reload or periodic refresh.
- Registry needs read access to the Vault path containing token hashes.

#### Option B: registry has its own token table, populated by operator command

The registry stores package token hashes in its own small SQLite/Postgres table:

```sql
CREATE TABLE publisher_tokens (
  id TEXT PRIMARY KEY,
  package_name TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  revoked_at TEXT
);
```

Pros:

- Registry can authorize without Vault during request handling.
- Token rotation can be handled by registry admin API.

Cons:

- Vault is no longer the only source of truth unless the operator process writes both.

For the first implementation, prefer **Option A** if Vault access from the registry pod is already straightforward. Prefer **Option B** if registry simplicity matters more than Vault coupling.

### Phase 1 authorization pseudocode

```go
type StaticTokenAuth struct {
    publishers map[string]PublisherIdentity // token hash -> identity
}

func (a *StaticTokenAuth) AuthorizePublish(ctx context.Context, rawToken string, req PublishRequest) (*PublisherIdentity, error) {
    if rawToken == "" {
        return nil, ErrUnauthorized
    }

    hash := HashToken(rawToken)
    identity, ok := a.publishers[hash]
    if !ok {
        return nil, ErrUnauthorized
    }

    if identity.Package != req.Package {
        return nil, ErrForbidden
    }

    if !ValidVersion(req.Version) {
        return nil, ErrForbidden
    }

    return &identity, nil
}
```

### Phase 1 token creation runbook

Operator flow:

```bash
PACKAGE=pinocchio
TOKEN="$(openssl rand -base64 48)"
TOKEN_HASH="$(printf '%s' "$TOKEN" | sha256sum | awk '{print $1}')"

vault kv put kv/docs-yolo/publishers/${PACKAGE} \
  package="${PACKAGE}" \
  token_hash="sha256:${TOKEN_HASH}" \
  created_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  created_by="$(whoami)"

printf 'Raw token for %s, store once in GitHub Actions secret DOCS_YOLO_PUBLISH_TOKEN:\n%s\n' "$PACKAGE" "$TOKEN"
```

Package owner adds the raw token to the package repository secret. The raw token should not be committed, logged, or copied into the docmgr ticket.

### Phase 1 token rotation runbook

```bash
PACKAGE=pinocchio
NEW_TOKEN="$(openssl rand -base64 48)"
NEW_HASH="$(printf '%s' "$NEW_TOKEN" | sha256sum | awk '{print $1}')"

vault kv patch kv/docs-yolo/publishers/${PACKAGE} \
  token_hash="sha256:${NEW_HASH}" \
  rotated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

printf 'New raw token for %s:\n%s\n' "$PACKAGE" "$NEW_TOKEN"
```

Then update the GitHub Actions secret and restart or refresh the registry token cache.

### Phase 1 limitations

This model is intentionally limited:

- It uses long-lived repository secrets.
- It does not prove which GitHub workflow used the token.
- It does not automatically bind publishing to tags or protected branches.
- Token leakage requires rotation.
- It does not produce the same quality of audit trail as GitHub OIDC to Vault.

These limitations are acceptable only because Phase 2 and Phase 3 are explicitly tracked separately.

### Phase 1 implementation tasks

Add these tasks to the implementation backlog:

1. Define `docsctl validate` for SQLite DB validation.
2. Define `docsctl publish` with `--server`, `--package`, `--version`, `--file`, and `--token`.
3. Add registry token hashing and package-scope authorization.
4. Add Vault-backed token-hash loading or an operator token table.
5. Add package DB validation before publication.
6. Add atomic write into `packages/<package>/<version>/<package>.db`.
7. Add post-publish operator instructions for rollout restart.
8. Add tests proving a token for package A cannot publish package B.

Phase 2 and Phase 3 should not be squeezed into this ticket's implementation scope. They are important enough to deserve their own design ticket and implementation guide.
