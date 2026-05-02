# Tasks

## Completed research and design setup

- [x] Add initial tasks.
- [x] Inspect current Glazed multi-package serving implementation and production deployment evidence.
- [x] Evaluate storage and reload architecture options for package-uploaded SQLite docs.
- [x] Write intern-facing design and implementation guide.
- [x] Update diary, changelog, and file relationships.
- [x] Validate ticket and upload bundle to reMarkable.
- [x] Phase 1: document static Vault-stored package publish token model in the design guide.
- [x] Phase 1: define docsctl validate/publish CLI contract for package SQLite uploads.
- [x] Phase 1: design registry authorization that maps one token hash to one package.
- [x] Phase 1: specify token creation and rotation runbooks using Vault.
- [x] Phase 1: update diary/changelog and re-upload current-ticket bundle after the addendum.

## Phase 1A — Local validation and publisher CLI foundation

- [ ] Add a `docsctl` CLI entrypoint under `cmd/docsctl` with root help and version wiring.
- [ ] Implement package/version/path validation helpers for docs publishing names.
- [ ] Implement a SQLite help DB validator that opens exports read-only and checks schema, section count, slugs, and duplicate slugs.
- [ ] Add unit tests for valid DBs, missing files, non-SQLite files, empty DBs, missing `sections`, empty slugs, duplicate slugs, and invalid package/version names.
- [ ] Add `docsctl validate --package --version --file --json` using the validator package.
- [ ] Add CLI smoke tests or command-level tests for `docsctl validate` success and failure output.
- [ ] Document `docsctl validate` usage in the ticket diary and commit the validation slice.

## Phase 1B — Static package-token authorization core

- [ ] Define a `PublisherAuth` interface and `PublishRequest` / `PublisherIdentity` structs for registry authorization.
- [ ] Implement token hashing with constant-time comparison and no raw-token logging.
- [ ] Implement `StaticTokenAuth` that maps token hashes to exactly one package.
- [ ] Add tests proving an unknown token is rejected, an empty token is rejected, a package A token cannot publish package B, and a package A token can publish package A.
- [ ] Add Vault token-hash record structs and parsing helpers for `kv/docs-yolo/publishers/<package>` payloads.
- [ ] Add a reloadable in-memory publisher-token catalog abstraction so Phase 1 can start with file/env fixtures and later read from Vault.
- [ ] Document Phase 1 token auth behavior and commit the auth slice.

## Phase 1C — Direct upload registry skeleton

- [ ] Add `cmd/docs-registry` with HTTP server configuration for listen address, package root, auth mode, and optional Vault settings.
- [ ] Add registry routes for `GET /healthz`, `GET /v1/packages`, and `PUT /v1/packages/{package}/versions/{version}/sqlite`.
- [ ] Wire publish endpoint authorization through `PublisherAuth` before reading or writing upload content.
- [ ] Reuse the SQLite validator on uploaded DBs before publication.
- [ ] Implement structured JSON error responses for unauthorized, forbidden, invalid upload, validation failure, and publish failure.
- [ ] Add request size limits and temporary upload file handling.
- [ ] Add registry handler tests with `httptest` for success, forbidden package, invalid DB, and oversized request.
- [ ] Document registry API behavior and commit the registry skeleton slice.

## Phase 1D — PVC directory publisher and atomic writes

- [ ] Implement a `PackageStore` / publisher interface for materializing validated DBs.
- [ ] Implement `DirectoryPackageStore` that writes `packages/<package>/<version>/<package>.db` under a configured root.
- [ ] Use write-to-temp + fsync + atomic rename so failed uploads do not replace existing docs.
- [ ] Reject path traversal and unsafe package/version values before building filesystem paths.
- [ ] Write or update a local catalog/manifest entry with package, version, sha256, section count, source, and publish timestamp.
- [ ] Add tests for atomic replacement, path traversal rejection, manifest update, and preserving old DB on failed validation.
- [ ] Document the on-disk layout and commit the directory-publisher slice.

## Phase 1E — `docsctl publish` client

- [ ] Add `docsctl publish --server --package --version --file --token --json --dry-run`.
- [ ] Make `docsctl publish --dry-run` run local validation and print the target upload URL without sending the DB.
- [ ] Implement HTTP upload with bearer token and useful error reporting.
- [ ] Support reading token from `--token`, `DOCS_YOLO_PUBLISH_TOKEN`, or a token file, with precedence documented.
- [ ] Add command tests for missing token, dry-run validation failure, server error reporting, and successful upload against `httptest`.
- [ ] Add an example GitHub Actions snippet to the ticket diary/design guide and commit the publish client slice.

## Phase 1F — Vault token-hash integration

- [ ] Decide whether the registry reads token hashes directly from Vault or starts with a file-backed catalog plus operator sync; record the decision in the design doc.
- [ ] If direct Vault read is chosen, add a `VaultPublisherCatalog` loader with configurable address, mount, prefix, and refresh interval.
- [ ] If file-backed catalog is chosen first, add a documented YAML/JSON catalog format that mirrors the Vault record shape.
- [ ] Add tests using fake Vault/client or fixture catalog data.
- [ ] Add operator runbook commands for creating, rotating, and revoking package tokens.
- [ ] Document the selected integration and commit the Vault/catalog slice.

## Phase 1G — docs-yolo GitOps deployment scaffold

- [ ] Add or plan a paired k3s GitOps ticket if cluster manifests will live outside this repository.
- [ ] Draft `docs-yolo` Kustomize manifests: namespace/application, deployment, service, ingress, PVC, and registry deployment/service.
- [ ] Configure `glaze serve --from-sqlite-dir /var/lib/glazed-docs/packages` in the docs-yolo deployment.
- [ ] Configure registry package root to the shared PVC path.
- [ ] Add readiness/liveness probes for both docs browser and registry.
- [ ] Add HTTP-01 TLS host `docs.yolo.scapegoat.dev` following the existing Glazed deployment pattern.
- [ ] Render manifests with `kubectl kustomize` and record validation output.
- [ ] Commit GitOps scaffold in the appropriate repository and relate it to this ticket.

## Phase 1H — End-to-end smoke validation and operational handoff

- [ ] Generate at least two package/version SQLite exports locally.
- [ ] Start docs registry and docs browser against a temporary package root.
- [ ] Publish package A and package B with distinct package tokens.
- [ ] Verify token A cannot publish package B.
- [ ] Verify `/api/packages` shows both packages and expected versions.
- [ ] Verify `/api/sections?package=<package>&version=<version>` returns expected sections.
- [ ] Verify bad uploads do not replace existing valid DBs.
- [ ] Document rollout restart or reload procedure for making newly published docs visible.
- [ ] Upload the Phase 1 implementation bundle to reMarkable.
- [ ] Run `docmgr doctor`, update diary/changelog, and commit final Phase 1 docs.

## Later phases tracked separately

- [ ] Phase 2: implement GitHub OIDC to Vault JWT auth and registry Vault capabilities checks under ticket `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`.
- [ ] Phase 3: implement Vault-signed docs publish JWTs under ticket `GG-20260502-VAULT-OIDC-DOCS-PUBLISH`.
- [ ] Phase 4: evaluate bucket or OCI/ORAS artifact storage after Phase 1 direct registry/PVC publishing is proven.
