# Tasks

## Phase 0: Planning and ticket structure

- [x] Replace the initial high-level task list with implementation phases and concrete tasks.
- [x] Record the implementation strategy in the diary before changing code.

## Phase 1: docs-registry JWT auth mode in Glazed

- [x] Add a `JWTPublisherAuth` implementation under `pkg/help/publish`.
- [x] Validate Vault-issued JWTs with OIDC discovery/JWKS rather than hand-rolled cryptography.
- [x] Require `token_use = docsctl-publish`.
- [x] Require the JWT `package` claim to match the requested `{package}` path parameter.
- [x] Add negative tests for wrong package, wrong audience, wrong issuer, expired token, tampered token, and missing claims.
- [x] Add `docs-registry --auth-mode static-catalog|vault-oidc-jwt`.
- [x] Keep current `--publisher-catalog` static mode as the default migration/rollback path.
- [x] Add `--jwt-issuer` and `--jwt-client-id` flags for JWT mode.
- [x] Run targeted Go tests for `pkg/help/publish`, `cmd/docs-registry`, and related packages.
- [x] Commit the Glazed implementation.

## Phase 2: Terraform Vault Identity/OIDC resources

- [x] Extend the Vault GitHub Actions Terraform environment with docsctl publisher locals.
- [x] Add a Vault Identity/OIDC issuer configuration if needed.
- [x] Add a `docs-registry-publish` OIDC signing key.
- [x] Add one Vault Identity OIDC role per package (`glazed`, `pinocchio`, `remarquee`, `sqleton`).
- [x] Add package-specific policies allowing each GitHub Actions role to mint only its own publish JWT.
- [x] Add GitHub Actions JWT roles with `repository`, `repository_id`, `ref`, `event_name`, and `job_workflow_ref` bound claims.
- [x] Add claim mappings for repository/workflow/run metadata.
- [x] Run `terraform fmt` and best-effort validation/plan checks without leaking credentials.
- [x] Commit the Terraform implementation.

## Phase 3: reusable GitHub Actions workflow integration

- [ ] Locate or create the shared workflow home for `publish-docsctl.yml`.
- [ ] Add workflow inputs for package name, version, export command, Vault role, Vault token role, Vault audience, and registry URL.
- [ ] Add a Vault login step using `hashicorp/vault-action@v3` and `id-token: write`.
- [ ] Add a publish JWT minting step using `GET /v1/identity/oidc/token/<role>`.
- [ ] Publish with `docsctl publish --token-file`.
- [ ] Verify the published package/version through `https://docs.yolo.scapegoat.dev/api/packages`.
- [ ] Commit the reusable workflow implementation in the appropriate repository.

## Phase 4: live proof for one package

- [ ] Run the workflow for `glazed` against a test version such as `vtest-jwt`.
- [ ] Decode and inspect non-sensitive publish JWT claims in a temporary proof log.
- [ ] Confirm wrong repository/branch/event/workflow cases are denied by Vault.
- [ ] Confirm docs-registry accepts a valid publish JWT.
- [ ] Confirm docs-registry rejects mismatched package JWTs.
- [ ] Record proof output and failures in the diary.

## Phase 5: k3s docs-yolo migration

- [ ] Build and publish a new Glazed image containing JWT auth mode.
- [ ] Update docs-yolo registry args to `--auth-mode vault-oidc-jwt`.
- [ ] Add `--jwt-issuer` and `--jwt-client-id` args.
- [ ] Remove the publisher catalog mount after rollback is no longer needed.
- [ ] Keep static-catalog rollback instructions in the ticket.
- [ ] Commit the GitOps deployment change.

## Phase 6: complete onboarding and delivery

- [ ] Onboard `pinocchio`, `remarquee`, and `sqleton` after `glazed` proves the path.
- [ ] Update the design document with any implementation corrections discovered during proof.
- [ ] Run final `docmgr doctor`.
- [ ] Upload the updated bundle to reMarkable.
- [ ] Close or mark remaining follow-ups explicitly.
