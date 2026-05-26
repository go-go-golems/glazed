---
Title: Diary
Ticket: DOCSCTL-VAULT-OIDC-JWT
Status: active
Topics:
    - docsctl
    - docs-yolo
    - vault
    - oidc
    - github-actions
    - security
    - cicd
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml
      Note: Reusable GitHub Actions workflow for Vault OIDC docsctl publishing (commit a95c5d5a08539d6b691b0c1ebb4c086132707808)
    - Path: ../../../../../../../../../../code/wesen/terraform/vault/github-actions/envs/k3s/main.tf
      Note: |-
        Existing Terraform Vault GitHub Actions JWT role pattern
        Vault Identity/OIDC and GitHub Actions roles for package-scoped publish JWTs (commit 04451fe795314065d872f22c8710044682525963)
        Applied Vault Identity/OIDC docsctl publisher resources and fixed live template/issuer validation (commit 2e56b7eb4cefd543df7e789af26b2eaedaf39c2a)
    - Path: ../../../../../../../../../../code/wesen/terraform/vault/github-actions/envs/k3s/variables.tf
      Note: Reusable workflow ref variable for docsctl publish JWT auth (commit 04451fe795314065d872f22c8710044682525963)
    - Path: .github/workflows/publish-docs.yml
      Note: Glazed caller workflow for publish-docsctl reusable workflow (commit 209ee884288cf086a4751044040c244d98aa61d2)
    - Path: cmd/docs-registry/main.go
      Note: |-
        Registry command boot path used to design auth-mode flags
        docs-registry auth mode flags and wiring (commit aa6946a40f2156689e81a831a10e634398102261)
    - Path: pkg/help/publish/auth.go
      Note: PublisherAuth interface that should receive the JWT implementation
    - Path: pkg/help/publish/jwt_auth.go
      Note: Vault OIDC JWT publisher auth implementation (commit aa6946a40f2156689e81a831a10e634398102261)
    - Path: pkg/help/publish/jwt_auth_test.go
      Note: JWT auth unit tests covering issuer/audience/package/expiry/tamper failures (commit aa6946a40f2156689e81a831a10e634398102261)
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/design-doc/01-vault-identity-oidc-publish-jwt-implementation-guide.md
      Note: |-
        Primary design document created for this ticket
        Primary design deliverable recorded by this diary
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/01-vault-identity-tokens.defuddle.md
      Note: Vault Identity Token conceptual evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/02-vault-identity-token-api.defuddle.md
      Note: Vault Identity Token API evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/05-github-oidc-reference.defuddle.md
      Note: GitHub Actions OIDC claim evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/github-oidc-reference.defuddle.md
      Note: GitHub Actions OIDC claim evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/terraform-vault-oidc-resource-schema.txt
      Note: Local Terraform Vault provider schema evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/vault-identity-token-api.defuddle.md
      Note: Vault Identity Token API evidence
    - Path: ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/vault-identity-tokens.defuddle.md
      Note: Vault Identity Token conceptual evidence
ExternalSources:
    - https://developer.hashicorp.com/vault/docs/secrets/identity/identity-token
    - https://developer.hashicorp.com/vault/api-docs/secret/identity/tokens
    - https://docs.github.com/actions/reference/openid-connect-reference
Summary: Chronological notes for designing Vault Identity/OIDC publish JWTs for docs-registry.
LastUpdated: 2026-05-26T07:51:07.835369732-04:00
WhatFor: Use this to resume or review the investigation behind DOCSCTL-VAULT-OIDC-JWT.
WhenToUse: Read before implementing the Vault Identity Token Terraform, registry JWT auth mode, or reusable workflow changes.
---





# Diary

## Goal

This diary records the creation of the `DOCSCTL-VAULT-OIDC-JWT` ticket and the design work for the second implementation of docs-registry authentication: GitHub Actions authenticates to Vault, Vault mints a short-lived OIDC-compliant docs publish JWT, and docs-registry validates that JWT instead of reading `publishers.json`.

## Step 1: Create ticket, gather Vault/OIDC evidence, and write the implementation guide

I created a new docmgr ticket for the Vault Identity/OIDC publish JWT design, fetched the relevant Vault and GitHub OIDC documentation into the ticket sources, inspected the existing docs-registry and docsctl code paths, checked local Terraform provider support for Vault Identity/OIDC resources, and wrote the primary implementation guide. The guide uses a textbook style for the cryptography and OIDC material: it explains the token types, signature verification, issuer/audience checks, JWKS, Vault Identity Token roles, and the difference between Vault Identity/OIDC and Vault Transit.

The design intentionally chooses Vault Identity Tokens rather than direct Transit signing. Vault Identity Token roles let operators define a package-specific claims template and control access through Vault policy. Transit can sign bytes, but it does not itself decide whether a CI job should be allowed to sign a JWT with `package = glazed` or `package = pinocchio`. For this ticket, the token issuer should be Vault Identity/OIDC, with Terraform managing the key, role, GitHub claim bindings, and token-minting policies.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to add vault identity/oidc publish JWT. (second implementation).

Explain all the cryptography / OIDC part in a section for our intern.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Use a textbook writing style where appropriate (skip analogies, read the skill).

Keep a diary as you work."

**Assistant interpretation:** Create a structured docmgr ticket and write a detailed intern-ready design for Vault-issued OIDC/JWT publish credentials for docs-registry, including cryptographic foundations and implementation guidance, then upload the result to reMarkable.

**Inferred user intent:** The user wants to turn the prior discussion about Vault as an internal token issuer into an actionable implementation package that can be handed to an intern without losing the security reasoning.

**Commit (code):** N/A — documentation-only work in the ticket workspace.

### What I did

- Read the `textbook-authoring` skill to match the requested writing style.
- Created ticket `DOCSCTL-VAULT-OIDC-JWT` with topics:
  - `docsctl`
  - `docs-yolo`
  - `vault`
  - `oidc`
  - `github-actions`
  - `security`
  - `cicd`
- Created the design doc:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/design-doc/01-vault-identity-oidc-publish-jwt-implementation-guide.md`
- Created this diary:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/reference/01-diary.md`
- Added ticket tasks for evidence gathering, design, cryptography explanation, Terraform/API/test design, and reMarkable upload.
- Fetched source documentation with `defuddle`:
  - `sources/01-vault-identity-tokens.defuddle.md`
  - `sources/02-vault-identity-token-api.defuddle.md`
  - `sources/03-vault-oidc-provider-concepts.defuddle.md`
  - `sources/04-vault-jwt-auth.defuddle.md`
  - `sources/05-github-oidc-reference.defuddle.md`
- Tried to fetch Terraform registry pages with `defuddle`; those pages produced empty files because the registry docs are script-heavy.
- Used `terraform providers schema -json` in a temporary Terraform directory to confirm local Vault provider resource support and saved the relevant schema summary to:
  - `sources/terraform-vault-oidc-resource-schema.txt`
- Inspected implementation files:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docs-registry/main.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/registry.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go`
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/cmd/docsctl/publish.go`
  - `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`

### Why

- The current `publishers.json` model uses long-lived static package tokens. That is acceptable for internal-only operation, but less attractive for a public registry endpoint.
- Vault Identity Tokens can issue short-lived signed JWTs with operator-defined claims. This is the correct shape for an application-specific publish token.
- The intern implementing the feature needs to understand not only which files to edit, but also why issuer, audience, signature, expiration, JWKS, and package claims are security-critical.

### What worked

- Vault's Identity Token docs directly support the desired design: authenticated clients can request signed OIDC-compliant ID tokens against a role, roles define templates and TTL, keys publish public JWKS for verification, and tokens can be validated by standard OIDC/JWT libraries.
- GitHub's OIDC docs identify the exact claims needed for caller validation: `repository`, `repository_id`, `ref`, `event_name`, `workflow_ref`, `job_workflow_ref`, and run metadata.
- The local Terraform provider schema confirmed support for `vault_identity_oidc`, `vault_identity_oidc_key`, `vault_identity_oidc_role`, and `vault_jwt_auth_backend_role`.
- The existing Go code has a clean `PublisherAuth` interface, so the registry can add JWT auth mode without changing the upload route or storage code.

### What didn't work

- Terraform provider docs on `registry.terraform.io` fetched through `defuddle` produced empty files:
  - `terraform-vault-identity-oidc-role.defuddle.md`
  - `terraform-vault-identity-oidc-key.defuddle.md`
  - `terraform-vault-identity-oidc-provider.defuddle.md`
- First attempt to inspect the existing Terraform provider schema in `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s` failed because the S3 backend wanted credentials:

```text
Error: No valid credential sources found
Please see https://developer.hashicorp.com/terraform/language/backend/s3
for more information about providing credentials.
Error: failed to refresh cached credentials, no EC2 IMDS role found,
operation error ec2imds: GetMetadata, request canceled, context deadline exceeded
```

- A second attempt with `AWS_EC2_METADATA_DISABLED=true` still failed because the backend remained active:

```text
Error: failed to refresh cached credentials, no EC2 IMDS role found,
operation error ec2imds: GetMetadata, access disabled to EC2 IMDS via
client option, or "AWS_EC2_METADATA_DISABLED" environment variable
```

- I then created a temporary Terraform config with `terraform init -backend=false`. The first temporary config had invalid single-line provider syntax:

```text
Error: Invalid single-argument block definition
on main.tf line 6, in provider "vault":
A single-line block definition must end with a closing brace immediately
after its single argument definition.
```

- Rewriting the provider block across multiple lines fixed the Terraform schema extraction.

### What I learned

- Vault Identity Tokens are the relevant Vault feature for this ticket, not the interactive OIDC authorization-code provider flow.
- Vault Identity Token role templates can add custom claims to a signed JWT, but top-level template keys must not overwrite standard OIDC claims such as `iss`, `sub`, `aud`, `iat`, or `exp`.
- The Vault JWT auth role can copy incoming GitHub claims into token and alias metadata using `claim_mappings`, which can then be used by Identity Token templates.
- Terraform provider support exists locally for the needed Vault resources, so this should be implementable without imperative bootstrap scripts except for proof/debugging.

### What was tricky to build

- The main tricky part is the metadata path between GitHub's OIDC token and Vault's emitted publish JWT. GitHub claims enter Vault during `auth/github-actions` login. The JWT auth role must validate claims and map selected claims into Vault metadata. The Identity Token role template then reads those mapped values using the JWT auth mount accessor. This is powerful but precise: if a claim mapping is absent, or if the template references the wrong alias accessor, the emitted publish JWT will have missing values.
- The design avoids making CI choose the JWT payload. That is the reason for choosing Identity Token roles instead of raw Transit signing. The package claim is fixed in the Vault role template, and CI only requests the role it has policy permission to read.

### What warrants a second pair of eyes

- Confirm the exact Vault issuer URL before implementation. The registry verifier must use the same issuer string that appears in the token's `iss` claim.
- Confirm whether the deployed Vault version allows multiple Identity Token roles to share `client_id = "docs-registry"`. If not, the design should switch to package-specific audiences.
- Confirm the exact GitHub repository names and immutable repository IDs before writing Terraform roles.
- Review whether `github.com/coreos/go-oidc/v3/oidc` is the preferred Go dependency for OIDC validation in this repository.

### What should be done in the future

- Add a live proof job for `glazed` that logs into Vault with GitHub OIDC, mints a publish JWT, decodes non-sensitive claims, and verifies the token against Vault JWKS.
- Implement `pkg/help/publish/jwt_auth.go` with unit tests covering issuer, audience, expiration, tampering, token use, and package mismatch.
- Update k3s GitOps manifests to run `docs-registry --auth-mode vault-oidc-jwt` once the proof is complete.

### Code review instructions

- Start with the design doc's Sections 5-7:
  - cryptography/OIDC foundations;
  - Vault configuration;
  - docs-registry implementation design.
- Then inspect:
  - `cmd/docs-registry/main.go` for the auth selection point.
  - `pkg/help/publish/auth.go` for the interface.
  - `pkg/help/publish/registry.go` for where `AuthorizePublish` is called.
  - `cmd/docsctl/publish.go` for existing bearer-token behavior.
  - Terraform `main.tf` for GitHub Actions JWT auth role conventions.

### Technical details

Ticket setup commands:

```bash
docmgr ticket create-ticket \
  --ticket DOCSCTL-VAULT-OIDC-JWT \
  --title "Vault Identity OIDC publish JWTs for docs-registry" \
  --topics docsctl,docs-yolo,vault,oidc,github-actions,security,cicd

docmgr doc add \
  --ticket DOCSCTL-VAULT-OIDC-JWT \
  --doc-type design-doc \
  --title "Vault Identity OIDC publish JWT implementation guide"

docmgr doc add \
  --ticket DOCSCTL-VAULT-OIDC-JWT \
  --doc-type reference \
  --title "Diary"
```

Source fetching commands:

```bash
defuddle parse https://developer.hashicorp.com/vault/docs/secrets/identity/identity-token --md -o sources/vault-identity-tokens.defuddle.md
defuddle parse https://developer.hashicorp.com/vault/api-docs/secret/identity/tokens --md -o sources/vault-identity-token-api.defuddle.md
defuddle parse https://developer.hashicorp.com/vault/docs/concepts/oidc-provider --md -o sources/vault-oidc-provider-concepts.defuddle.md
defuddle parse https://developer.hashicorp.com/vault/docs/auth/jwt --md -o sources/vault-jwt-auth.defuddle.md
defuddle parse https://docs.github.com/actions/reference/openid-connect-reference --md -o sources/github-oidc-reference.defuddle.md

# Later normalized to docmgr-friendly numeric filenames with frontmatter:
# sources/01-vault-identity-tokens.defuddle.md
# sources/02-vault-identity-token-api.defuddle.md
# sources/03-vault-oidc-provider-concepts.defuddle.md
# sources/04-vault-jwt-auth.defuddle.md
# sources/05-github-oidc-reference.defuddle.md
```

Terraform schema proof used a temporary config with backend disabled:

```bash
terraform init -backend=false -input=false
terraform providers schema -json > /tmp/vault-provider-schema2.json
```

Key Terraform resources confirmed:

```text
vault_identity_oidc
vault_identity_oidc_key
vault_identity_oidc_role
vault_identity_oidc_provider
vault_identity_oidc_client
vault_identity_oidc_scope
vault_jwt_auth_backend_role
```

## Step 2: Normalize sources, validate ticket, and upload to reMarkable

I cleaned up the ticket source files so `docmgr doctor` could validate the workspace, uploaded the final bundle to reMarkable, and checked the final upload task. The final ticket now has a passing doctor report and a delivered PDF bundle.

The main cleanup was caused by external documentation captures. `defuddle` creates clean Markdown but not docmgr frontmatter, and Terraform Registry pages produced empty Markdown captures. I removed the empty Terraform captures, renamed the real source files with numeric prefixes, added minimal source frontmatter, refreshed diary relations, and reran validation.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish docmgr hygiene and deliver the Vault Identity/OIDC publish JWT design package to reMarkable.

**Inferred user intent:** The user wants a durable, searchable ticket plus a readable PDF bundle for review away from the terminal.

**Commit (code):** N/A — documentation-only work in the ticket workspace.

### What I did

- Removed empty Terraform Registry captures:
  - `terraform-vault-identity-oidc-key.defuddle.md`
  - `terraform-vault-identity-oidc-provider.defuddle.md`
  - `terraform-vault-identity-oidc-role.defuddle.md`
- Renamed source docs to numeric docmgr-friendly filenames.
- Added minimal frontmatter to source Markdown files.
- Refreshed diary file relations to the renamed source files.
- Ran `docmgr doctor --ticket DOCSCTL-VAULT-OIDC-JWT --stale-after 30` until it passed.
- Uploaded the bundle to reMarkable:
  - `/ai/2026/05/26/DOCSCTL-VAULT-OIDC-JWT/DOCSCTL Vault OIDC Publish JWT.pdf`
- Checked the final upload task.

### Why

- `docmgr doctor` enforces frontmatter and filename conventions. The ticket should pass validation before handoff.
- The user requested reMarkable upload, so the design doc, diary, tasks, and changelog were bundled into one PDF.

### What worked

- The final doctor report passed:

```text
## Doctor Report (1 findings)

### DOCSCTL-VAULT-OIDC-JWT

- ✅ All checks passed
```

- The reMarkable upload completed with:

```text
OK: uploaded DOCSCTL Vault OIDC Publish JWT.pdf -> /ai/2026/05/26/DOCSCTL-VAULT-OIDC-JWT
```

### What didn't work

- Initial `docmgr doctor` failed because the fetched source Markdown files had no YAML frontmatter:

```text
[error] YAML/frontmatter syntax error
Problem: frontmatter delimiters '---' not found
```

- The Terraform Registry defuddle captures were empty, so keeping them as `.md` files created invalid ticket sources. I removed those empty files and kept the Terraform provider schema output as a `.txt` source.

### What I learned

- When storing defuddle captures inside a docmgr ticket, add frontmatter and numeric prefixes before running `docmgr doctor`.
- For Terraform provider docs, `terraform providers schema -json` can be more reliable than scraping the dynamic Registry UI.

### What was tricky to build

- The source capture cleanup touched both filesystem names and docmgr relations. I refreshed the diary relations after renaming so the ticket would point at the actual source files rather than the pre-normalization paths.

### What warrants a second pair of eyes

- Check whether the design should include the source documents in the reMarkable bundle. I uploaded only the design, diary, tasks, and changelog to keep the PDF focused.

### What should be done in the future

- Consider a small helper script for ticket research that runs `defuddle`, adds source frontmatter, and prefixes filenames automatically.

### Code review instructions

- Validate ticket hygiene with:
  - `docmgr doctor --ticket DOCSCTL-VAULT-OIDC-JWT --stale-after 30`
- Review uploaded source list under:
  - `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/`

### Technical details

Successful upload command:

```bash
remarquee upload bundle \
  "$TICKET_DIR/design-doc/01-vault-identity-oidc-publish-jwt-implementation-guide.md" \
  "$TICKET_DIR/reference/01-diary.md" \
  "$TICKET_DIR/tasks.md" \
  "$TICKET_DIR/changelog.md" \
  --name "DOCSCTL Vault OIDC Publish JWT" \
  --remote-dir "/ai/2026/05/26/DOCSCTL-VAULT-OIDC-JWT" \
  --toc-depth 2 \
  --non-interactive 2>&1
```

## Step 3: Expand the ticket into implementation phases

I replaced the coarse task list with a phase-by-phase implementation checklist. The phases now separate planning, Glazed registry code, Terraform Vault resources, reusable workflow integration, live proof, k3s migration, and final delivery. This makes the ticket executable rather than only descriptive.

The ordering is deliberate. The registry can learn to validate Vault-issued JWTs before production uses that mode. Terraform can define the Vault token issuer before k3s switches the deployment. The reusable workflow can be proven against one package before all packages are onboarded.

### Prompt Context

**User prompt (verbatim):** "Alright, create a detailed set of phases + tasks in the ticket. Then implement them one by one, committing at appropriate intervals (in the terraform repo etc...) and keeping a detailed diary (in the customer format with why things worked / not worked etc...)"

**Assistant interpretation:** Turn the design into a sequenced implementation plan, then begin implementing the phases with focused commits and detailed diary updates.

**Inferred user intent:** The user wants this to move from research/design into actual implementation across Glazed, Terraform, and eventually GitOps, while preserving a reviewable audit trail.

**Commit (code):** Pending — planning update will be committed with the first implementation batch or as a focused docs commit if needed.

### What I did

- Rewrote `tasks.md` with phases 0 through 6.
- Broke the registry implementation into concrete testable tasks.
- Broke the Terraform implementation into concrete Vault resources and validation steps.
- Left workflow, live proof, k3s migration, and package onboarding as later phases because they depend on code and Terraform being ready first.

### Why

- The design document explained the system, but implementation needs a checklist that can be checked off as commits land.
- Splitting work by repository reduces the chance of mixing unrelated changes in one commit.
- The live k3s deployment should not be migrated before the image and Vault resources exist.

### What worked

- The existing ticket structure already had a `tasks.md`, so the implementation plan could be recorded directly in the ticket.
- The phase boundaries align with natural commit boundaries: Glazed code, Terraform resources, workflow integration, GitOps migration.

### What didn't work

- N/A for this planning step.

### What I learned

- The safest path is not to update the live GitOps deployment in the same step as the registry code. The registry image must exist, and Vault JWT issuance must be proven first.

### What was tricky to build

- The task list needed to be detailed enough to guide implementation but not pretend that later phases can be completed before their prerequisites exist. I kept phase 5 k3s migration pending for that reason.

### What warrants a second pair of eyes

- Review whether the shared reusable workflow should live in an existing automation repository or in a new repository. The task list intentionally leaves that discovery in Phase 3.

### What should be done in the future

- Check off tasks only after validation commands have been run and the relevant commit hashes are recorded.

### Code review instructions

- Start with `tasks.md` and confirm the ordering matches the desired rollout safety.
- Then review the next implementation commit against Phase 1.

### Technical details

Updated file:

```text
/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/tasks.md
```

## Step 4: Implement docs-registry Vault OIDC JWT auth mode in Glazed

I implemented the registry-side JWT validation path in Glazed. The new `JWTPublisherAuth` uses OIDC discovery and JWKS validation through `github.com/coreos/go-oidc/v3/oidc`, then applies docs-registry-specific authorization checks: the token must have `token_use = docsctl-publish` and its `package` claim must match the package in the upload URL.

This step keeps the existing static catalog auth mode as the default. That preserves the current deployment behavior and gives us a rollback path while the Vault Identity Token and CI pieces are implemented.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Begin implementing the phased plan, starting with the Glazed registry code because Terraform and k3s rollout depend on a registry image that can validate Vault-issued publish JWTs.

**Inferred user intent:** The user wants concrete code changes, tests, and commits rather than only a design document.

**Commit (code):** Pending at time of diary entry; intended commit message: `docs-registry: add Vault OIDC JWT auth mode`.

### What I did

- Added `pkg/help/publish/jwt_auth.go`.
- Added `pkg/help/publish/jwt_auth_test.go`.
- Added dependency `github.com/coreos/go-oidc/v3/oidc` via `go get`.
- Updated `cmd/docs-registry/main.go` with:
  - `--auth-mode` flag;
  - `--jwt-issuer` flag;
  - `--jwt-client-id` flag;
  - `buildPublisherAuth` helper;
  - static catalog mode as the default.
- Ran `gofmt`.
- Ran targeted tests:

```bash
cd /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed
go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl
```

### Why

- The registry already delegates authorization to `PublisherAuth`, so adding a new implementation is the smallest safe change.
- OIDC discovery/JWKS validation should be handled by a maintained library. The registry should not implement JWT signature verification by hand.
- Static mode should remain available until the Vault roles, reusable workflow, and k3s deployment have all been proven.

### What worked

- The OIDC/JWKS unit tests use a local `httptest.Server` as a miniature OIDC issuer, which lets the test suite cover signature, issuer, audience, expiry, tampering, token-use, and package-claim behavior without live Vault.
- Targeted Go tests passed:

```text
ok  	github.com/go-go-golems/glazed/pkg/help/publish	1.069s
?   	github.com/go-go-golems/glazed/cmd/docs-registry	[no test files]
ok  	github.com/go-go-golems/glazed/cmd/docsctl	0.065s
```

### What didn't work

- N/A in this code step. The implementation and targeted tests passed on the first run after formatting.

### What I learned

- The existing `PublisherAuth` interface was the correct extension point. The upload path in `registry.go` did not need to change.
- `docsctl publish` can carry the Vault-issued JWT without client changes because it already sends `Authorization: Bearer <token>` from `--token-file`.

### What was tricky to build

- The tests needed to exercise real cryptographic verification without depending on Vault. I generated an RSA key in the test, served OIDC discovery and JWKS endpoints, signed RS256 tokens, and verified them through the same OIDC library the registry will use in production.
- The tamper test intentionally rewrites the payload while keeping the original signature so the verifier rejects the token before any package claim is trusted.

### What warrants a second pair of eyes

- Review whether `github.com/coreos/go-oidc/v3/oidc` is the preferred dependency for this repository.
- Review whether `token_use = docsctl-publish` and `package` are sufficient application-specific checks for Phase 1, or whether repository ID should also be enforced locally by the registry.

### What should be done in the future

- Add live Vault integration proof after Terraform creates the Identity Token role.
- Consider first-class `docsctl publish --auth vault-identity-oidc` after the shell-based workflow proof works.

### Code review instructions

- Start with `pkg/help/publish/jwt_auth.go` for the authorization logic.
- Then read `pkg/help/publish/jwt_auth_test.go` to see the security expectations encoded as tests.
- Finally inspect `cmd/docs-registry/main.go` for migration behavior and flag validation.
- Validate with:

```bash
go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl
```

### Technical details

New production command shape:

```bash
docs-registry \
  --auth-mode vault-oidc-jwt \
  --jwt-issuer https://vault.yolo.scapegoat.dev/v1/identity/oidc \
  --jwt-client-id docs-registry \
  --address :8090 \
  --package-root /var/lib/glazed-docs/packages
```

Static mode remains:

```bash
docs-registry \
  --auth-mode static-catalog \
  --publisher-catalog /etc/docs-yolo/publishers.json \
  --address :8090 \
  --package-root /var/lib/glazed-docs/packages
```

## Step 5: Commit the Glazed JWT implementation after full pre-commit validation

I committed the Glazed JWT implementation after the repository pre-commit hook ran the broader test, lint, gosec, and govulncheck checks. The first commit attempt failed because an existing untracked `wellknown.go` file in the working tree had lint issues. I fixed those lint-only issues in the untracked file so the repository could pass validation, then committed only the staged JWT implementation files.

This is important because Go tooling sees untracked `.go` files when it loads packages. Even though `wellknown.go` was not staged for the JWT commit, the linter and tests still compiled/linted it because it sits in `pkg/help/server`.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Make a focused Glazed commit for the registry JWT implementation and record any validation friction.

**Inferred user intent:** The user wants focused commits and a diary that explains not only the happy path but also why failures happened.

**Commit (code):** `aa6946a40f2156689e81a831a10e634398102261` — `docs-registry: add Vault OIDC JWT auth mode`

### What I did

- Staged only:
  - `cmd/docs-registry/main.go`
  - `go.mod`
  - `go.sum`
  - `pkg/help/publish/jwt_auth.go`
  - `pkg/help/publish/jwt_auth_test.go`
- Attempted commit with the normal pre-commit hook.
- Fixed two lint issues in the existing untracked `pkg/help/server/wellknown.go`:
  - replaced `b.WriteString(fmt.Sprintf(...))` with `fmt.Fprintf(...)`;
  - removed named returns from `isMarkdownSuffixURL`.
- Re-ran the commit.

### Why

- The JWT implementation should be reviewable independently from the previously-existing server/A14Y working-tree changes.
- Skipping hooks would have hidden the fact that the current checkout could not pass lint with the untracked server file present.

### What worked

- The second commit passed the full hook:
  - `go test` package set passed;
  - `golangci-lint` reported `0 issues`;
  - `gosec` reported `Issues: 0`;
  - `govulncheck` reported no called vulnerabilities.

### What didn't work

The first commit attempt failed in the `lint` hook with:

```text
pkg/help/server/wellknown.go:385:1: named return "pkgName" with type "string" found (nonamedreturns)
func isMarkdownSuffixURL(path string) (pkgName, version, slug string, ok bool) {
^
pkg/help/server/wellknown.go:300:3: QF1012: Use fmt.Fprintf(...) instead of WriteString(fmt.Sprintf(...)) (staticcheck)
		b.WriteString(fmt.Sprintf("## %s\n\n", name))
		^
2 issues:
* nonamedreturns: 1
* staticcheck: 1
make: *** [Makefile:25: lintmax] Error 1
```

### What I learned

- A focused Git commit can still be blocked by unrelated untracked Go files because package-level tooling scans the whole package directory, not only staged files.
- The right fix here was not to bypass validation, but to make the checkout lint-clean while staging only the intended commit contents.

### What was tricky to build

- The tricky part was preserving commit hygiene. `wellknown.go` had to be lint-fixed in the working tree so hooks could pass, but it was intentionally not included in the JWT commit. I verified the staged file list before committing.

### What warrants a second pair of eyes

- Review the staged commit independently of the untracked `wellknown.go` state. The commit hash above should include only JWT auth files plus module dependency updates.

### What should be done in the future

- Finish or separately commit the pre-existing help-server/A14Y changes so future commits are not affected by untracked package files.

### Code review instructions

- Review commit `aa6946a40f2156689e81a831a10e634398102261` directly.
- Confirm it does not include the unrelated help-server file.
- Validate with `go test ./pkg/help/publish ./cmd/docs-registry ./cmd/docsctl` or the repository pre-commit hook.

### Technical details

The commit command that ultimately passed was:

```bash
git commit -m "docs-registry: add Vault OIDC JWT auth mode"
```

## Step 6: Add Terraform Vault roles for docsctl publish JWT issuance

I added Terraform resources to the existing Vault GitHub Actions auth environment for package-scoped docsctl publish JWTs. The configuration now defines the package list, the Vault Identity/OIDC issuer, a signing key for docs-registry tokens, one Vault Identity OIDC role per package, one policy per package, and one GitHub Actions JWT auth role per package.

The Terraform side is intentionally package-scoped. A `glazed` workflow can mint only the `docsctl-glazed-publisher` token, and the resulting publish JWT carries `package = "glazed"`. The registry code from Step 4 then rejects that JWT if it is used against another package URL.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Implement the Vault/Terraform phase after the Glazed registry can validate Vault-issued OIDC JWTs.

**Inferred user intent:** The user wants the infrastructure repository to contain the Vault resources needed for real CI publishing, not only application code.

**Commit (code):** `04451fe795314065d872f22c8710044682525963` — `vault: add docsctl publish JWT roles`

### What I did

- Modified `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`.
- Added `local.docsctl_publishers` for:
  - `glazed` / repository id `565461475`;
  - `pinocchio` / repository id `802670903`;
  - `remarquee` / repository id `1116463013`;
  - `sqleton` / repository id `579241534`.
- Added `vault_identity_oidc.docs_registry` with issuer `${var.vault_addr}/v1/identity/oidc`.
- Added `vault_identity_oidc_key.docs_registry_publish` with `RS256`, `allowed_client_ids = ["docs-registry"]`, daily rotation, and seven-day verification TTL.
- Added `vault_identity_oidc_role.docsctl_publish` for one role per package.
- Added `vault_policy.docsctl_publish` so each GitHub Actions role can read only its package token endpoint.
- Added `vault_jwt_auth_backend_role.docsctl_publish` with bound claims for repository owner, repository, repository id, ref, event name, workflow ref, and reusable workflow ref.
- Added `docsctl_publish_job_workflow_ref` variable.
- Ran `terraform fmt`.
- Ran validation in a temporary backend-free copy because the real environment requires S3 backend credentials.

### Why

- The GitHub Actions OIDC JWT proves which repository/workflow is calling Vault.
- The Vault auth role turns that proof into a short-lived Vault token with only one policy.
- The policy allows reading only one Vault Identity/OIDC token endpoint.
- The Vault Identity/OIDC role signs the docs-registry publish JWT with a package claim that the registry can verify.

### What worked

- `terraform fmt main.tf variables.tf` completed successfully.
- Backend-free validation in a temporary copy passed:

```text
Success! The configuration is valid.
```

- The Terraform commit was created in `/home/manuel/code/wesen/terraform`:

```text
04451fe795314065d872f22c8710044682525963 vault: add docsctl publish JWT roles
```

### What didn't work

Running Terraform initialization directly in the real environment failed because the S3 backend credentials are not available in this shell:

```text
Error: No valid credential sources found

Please see https://developer.hashicorp.com/terraform/language/backend/s3
for more information about providing credentials.

Error: failed to refresh cached credentials, no EC2 IMDS role found,
operation error ec2imds: GetMetadata, request canceled, context deadline
exceeded
```

To validate syntax/provider compatibility without touching the remote backend, I copied the environment to a temporary directory, removed the `backend "s3"` block only in the temporary copy, ran `terraform init -backend=false`, and then ran `terraform validate`.

### What I learned

- The local shell still cannot access the Terraform S3 backend, so apply/plan must be done from an environment with the correct AWS credentials.
- `gh api repos/<owner>/<repo>` provided the numeric GitHub repository ids that the GitHub OIDC `repository_id` claim should use.

### What was tricky to build

- Vault Identity Token templates need Vault's alias metadata path, which depends on the GitHub Actions auth mount accessor. The Terraform template uses `${vault_jwt_auth_backend.github_actions.accessor}` so Vault can later render values from the entity alias metadata populated by `claim_mappings`.
- Terraform validation was tricky because backend access failed before normal validation could proceed. The workaround validated a backend-free temporary copy without modifying the real checked-in backend configuration.

### What warrants a second pair of eyes

- Confirm that the deployed Vault version supports the Identity/OIDC resources exactly as modeled.
- Confirm that Vault template rendering accepts the alias metadata paths as written.
- Confirm that the future reusable workflow ref should be `go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@refs/heads/main`, or update `docsctl_publish_job_workflow_ref` before applying.
- Confirm that the package repositories will use `.github/workflows/publish-docs.yml@refs/heads/main` as the caller workflow file.

### What should be done in the future

- Run `terraform plan` and `terraform apply` from a credentialed environment.
- After apply, manually mint one token through Vault for `glazed` and inspect non-sensitive claims.
- If the shared reusable workflow repository/name changes, update `docsctl_publish_job_workflow_ref` before applying.

### Code review instructions

- Start in `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf` at `local.docsctl_publishers`.
- Review the relationship between `vault_jwt_auth_backend_role.docsctl_publish`, `vault_policy.docsctl_publish`, and `vault_identity_oidc_role.docsctl_publish`.
- Validate from a credentialed Terraform environment with:

```bash
terraform -chdir=/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s init
terraform -chdir=/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s plan
```

### Technical details

Backend-free validation command pattern used locally:

```bash
tmp=$(mktemp -d)
cp -a /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/. "$tmp"
# Remove only the backend "s3" block in the temporary copy.
cd "$tmp"
rm -rf .terraform
terraform init -backend=false -input=false
terraform validate
```

## Step 7: Try the k3s `.envrc` for Terraform credentials

I tried to use the k3s repository `.envrc` as the user suggested, then re-ran Terraform initialization from the Vault GitHub Actions Terraform environment. The `.envrc` does set cluster- and application-related variables such as `VAULT_ADDR`, `KUBECONFIG`, and docs/GitOps tokens, but it does not set AWS credentials or an AWS profile for the S3 Terraform backend.

As a result, the Terraform remote backend remains blocked from this shell. I did not attempt a Terraform apply with a backend-free local state, because that would be unsafe and could diverge from the real managed state.

### Prompt Context

**User prompt (verbatim):** "use the ~/code/wesen/2026-03-27--hetzner-k3s/.envrc for keys."

**Assistant interpretation:** Source the k3s `.envrc` before running Terraform so the missing credentials from the previous step may become available.

**Inferred user intent:** The user wants the Terraform plan/apply work to proceed using the environment variables already stored in the k3s repository.

**Commit (code):** N/A — no code changed in this step.

### What I did

- Sourced `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/.envrc`.
- Re-ran `terraform init` in `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s`.
- Repeated the source from inside the k3s directory so `$PWD`-based exports such as `KUBECONFIG` resolve correctly.
- Checked only whether credential-related environment variables were set, without recording secret values in the diary.

### Why

- The previous Terraform validation step could not initialize the real S3 backend because AWS credentials were unavailable.
- The user pointed at the k3s `.envrc` as the likely source for the required keys.

### What worked

- The `.envrc` sourced successfully.
- It set `VAULT_ADDR` and cluster/application token variables.
- Sourcing from inside the k3s directory correctly set the kubeconfig path.

### What didn't work

- The `.envrc` did not set these Terraform S3 backend credential variables:

```text
AWS_ACCESS_KEY_ID=unset
AWS_SECRET_ACCESS_KEY=unset
AWS_SESSION_TOKEN=unset
AWS_PROFILE=unset
AWS_DEFAULT_PROFILE=unset
AWS_REGION=unset
AWS_DEFAULT_REGION=unset
VAULT_TOKEN=unset
```

- `terraform init` still failed with:

```text
Error: No valid credential sources found

Please see https://developer.hashicorp.com/terraform/language/backend/s3
for more information about providing credentials.

Error: failed to refresh cached credentials, no EC2 IMDS role found,
operation error ec2imds: GetMetadata, access disabled to EC2 IMDS via
client option, or "AWS_EC2_METADATA_DISABLED" environment variable
```

### What I learned

- The k3s `.envrc` is useful for cluster and docs/GitOps operations, but it is not sufficient for Terraform S3 backend access in this environment.
- The real Terraform plan/apply still needs AWS backend credentials or an AWS profile supplied from elsewhere.

### What was tricky to build

- Sourcing `.envrc` from the wrong working directory can produce incorrect `$PWD`-based paths. I retried from inside the k3s repository so `KUBECONFIG=$PWD/.cache/...` would be accurate.
- Terraform fails during backend initialization before provider-level Vault credentials matter, so `VAULT_ADDR` alone does not unblock planning.

### What warrants a second pair of eyes

- Confirm where AWS credentials for the `go-go-golems-tf-state` S3 backend are expected to come from on this workstation.
- Review the `.envrc` storage pattern for plaintext long-lived tokens; those should ideally be moved to a secret manager or rotated if they were exposed in logs.

### What should be done in the future

- Provide or load an AWS profile/credentials for the Terraform backend, then run:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
source ./.envrc
cd /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s
terraform init -input=false
terraform plan -input=false -out=/tmp/docsctl-vault-oidc-jwt.tfplan
```

### Code review instructions

- No code review is needed for this step.
- Review the next Terraform plan once backend credentials are available.

### Technical details

I did not use `terraform init -backend=false` against the real checkout for apply. Backend-free validation remains acceptable for syntax checks only, not for changing live Vault resources.

## Step 8: Use the Terraform `.envrc` for the real backend plan

I checked `/home/manuel/code/wesen/terraform/.envrc` after the user asked whether that file contains the missing credentials. It does: it sets `AWS_PROFILE`, and that profile can authenticate to AWS. After sourcing it from the Terraform repository, `terraform init` succeeded against the real S3 backend and `terraform plan` succeeded for the Vault GitHub Actions environment.

The resulting plan proposes only the expected docsctl publish JWT resources: Vault Identity/OIDC issuer configuration, a docs-registry signing key, four package-specific Identity OIDC roles, four package-specific policies, and four GitHub Actions JWT auth roles. It reports `14 to add, 0 to change, 0 to destroy`.

### Prompt Context

**User prompt (verbatim):** "does ~/code/wesen/terraform/.envrc contain them?"

**Assistant interpretation:** Inspect the Terraform repository `.envrc` for AWS backend credentials/profile and, if present, retry Terraform initialization/planning with that environment.

**Inferred user intent:** The user is pointing to the likely correct credential source so we can unblock the real Terraform backend plan.

**Commit (code):** Pending diary-only commit at time of writing.

### What I did

- Checked that `/home/manuel/code/wesen/terraform/.envrc` exists.
- Inspected it with secret values redacted.
- Confirmed it exports `AWS_PROFILE`.
- Confirmed the AWS profile works with `aws sts get-caller-identity` without recording sensitive values.
- Sourced the Terraform `.envrc` from `/home/manuel/code/wesen/terraform`.
- Ran:

```bash
cd /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s
terraform init -input=false
terraform plan -input=false -out=/tmp/docsctl-vault-oidc-jwt.tfplan
```

### Why

- The k3s `.envrc` contains cluster and application tokens but not AWS S3 backend credentials.
- The Terraform `.envrc` is the correct place to load the AWS profile used by the Terraform backend.

### What worked

- `aws sts get-caller-identity` succeeded using the sourced `AWS_PROFILE`.
- `terraform init` succeeded against the real backend.
- `terraform plan` succeeded and wrote:

```text
/tmp/docsctl-vault-oidc-jwt.tfplan
```

- Plan summary:

```text
Plan: 14 to add, 0 to change, 0 to destroy.
```

### What didn't work

- N/A for this retry. The Terraform `.envrc` unblocked the backend.

### What I learned

- The correct credential source for this Terraform state is `/home/manuel/code/wesen/terraform/.envrc`, not the k3s `.envrc`.
- The plan does not modify existing GitHub Actions Vault roles; it only adds the docsctl publish JWT resources.

### What was tricky to build

- It was important to source `.envrc` from the Terraform repository because it also conditionally sources `.envrc.local` and may rely on repository-relative behavior.
- I avoided printing secret values while still proving that the relevant variables and AWS identity are available.

### What warrants a second pair of eyes

- Review the plan before apply, especially the `job_workflow_ref` and package `workflow_ref` values, because applying these roles will enforce those exact GitHub Actions workflow paths.
- Confirm that `go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@refs/heads/main` is the intended reusable workflow identity before applying.

### What should be done in the future

- If the workflow refs are correct, apply the saved plan:

```bash
cd /home/manuel/code/wesen/terraform
source ./.envrc
cd /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s
terraform apply /tmp/docsctl-vault-oidc-jwt.tfplan
```

### Code review instructions

- Review Terraform commit `04451fe795314065d872f22c8710044682525963` together with the generated plan.
- Pay special attention to the exact role names and bound claims.

### Technical details

The successful plan showed the following high-level resource count:

```text
1 x vault_identity_oidc.docs_registry
1 x vault_identity_oidc_key.docs_registry_publish
4 x vault_identity_oidc_role.docsctl_publish
4 x vault_policy.docsctl_publish
4 x vault_jwt_auth_backend_role.docsctl_publish
```

## Step 9: Apply Terraform, fix live Vault validation errors, and re-apply

I applied the saved Terraform plan after the user approved it. The first apply partially succeeded: Vault created the `docs-registry-publish` signing key, then rejected the Identity/OIDC issuer configuration and all package token role templates. I fixed both issues in Terraform, planned again, and the second apply succeeded.

The key correction is that Vault's Identity/OIDC configuration resource accepts only a scheme, host, and optional port for `issuer`, even though the Identity Token discovery document is later served from `/v1/identity/oidc/.well-known/openid-configuration` and reports the token issuer as `https://vault.yolo.scapegoat.dev/v1/identity/oidc`. The second correction is that Vault Identity Token templates expect templated values to be unquoted JSON fragments, because Vault inserts JSON-encoded strings during token rendering.

### Prompt Context

**User prompt (verbatim):** "alright, apply, then continue."

**Assistant interpretation:** Apply the approved Terraform plan, fix any live apply issues, then continue implementing the next project phase.

**Inferred user intent:** The user wants the Vault resources actually created, not only planned, and then wants implementation to proceed toward CI workflow integration.

**Commit (code):** `2e56b7eb4cefd543df7e789af26b2eaedaf39c2a` — `vault: fix docsctl publish token templates`

### What I did

- Sourced `/home/manuel/code/wesen/terraform/.envrc` for the AWS backend profile.
- Ran `terraform apply /tmp/docsctl-vault-oidc-jwt.tfplan`.
- Observed live Vault validation failures.
- Updated `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`:
  - changed `vault_identity_oidc.docs_registry.issuer` from `${var.vault_addr}/v1/identity/oidc` to `var.vault_addr`;
  - removed quotes around dynamic Vault Identity template expressions.
- Ran a new plan to `/tmp/docsctl-vault-oidc-jwt-v2.tfplan`.
- Applied the corrected plan.
- Ran a final `terraform plan -detailed-exitcode` and confirmed exit code `0` / no changes.
- Committed the Terraform correction.

### Why

- The first failures came from Vault server-side validation rules that are stricter/different than Terraform syntax validation.
- The partial first apply had already created the signing key, so the corrected second plan had `13 to add` rather than the original `14 to add`.

### What worked

- Corrected apply succeeded:

```text
Apply complete! Resources: 13 added, 0 changed, 0 destroyed.
```

- Final clean plan succeeded:

```text
No changes. Your infrastructure matches the configuration.
```

- Public Vault Identity Token discovery endpoints are reachable:

```text
https://vault.yolo.scapegoat.dev/v1/identity/oidc/.well-known/openid-configuration -> HTTP 200
https://vault.yolo.scapegoat.dev/v1/identity/oidc/.well-known/keys -> HTTP 200
```

### What didn't work

The first apply failed with:

```text
Error: error writing IdentityOidc https://vault.yolo.scapegoat.dev: Error making API request.
URL: PUT https://vault.yolo.scapegoat.dev/v1/identity/oidc/config
Code: 400. Errors:
* invalid issuer, which must include only a scheme, host, and optional port (e.g. https://example.com:8200)
```

It also failed for each package role with:

```text
Error: error writing IdentityOidcRole identity/oidc/role/docsctl-<package>-publisher: Error making API request.
URL: PUT https://vault.yolo.scapegoat.dev/v1/identity/oidc/role/docsctl-<package>-publisher
Code: 400. Errors:
* error parsing template JSON: invalid character '"' after object key:value pair
```

### What I learned

- `vault_identity_oidc.issuer` config is the public host base, while the Identity Token OIDC discovery URL remains `/v1/identity/oidc/.well-known/openid-configuration`.
- Vault Identity Token templates are JSON-like templates, not plain Terraform-interpolated JSON. Dynamic template placeholders should be unquoted when the referenced value is a JSON string.

### What was tricky to build

- The first apply partially succeeded. I checked Terraform state and confirmed `vault_identity_oidc_key.docs_registry_publish` was already managed before producing the second plan.
- The registry should still use JWT issuer `https://vault.yolo.scapegoat.dev/v1/identity/oidc`, because that is what Vault's discovery endpoint reports, even though Terraform's `vault_identity_oidc` config accepts only `https://vault.yolo.scapegoat.dev`.

### What warrants a second pair of eyes

- Review the distinction between the Vault Identity/OIDC config issuer and the OIDC discovery issuer before wiring k3s registry flags.
- Review the unquoted template placeholders in Terraform; they are intentional and follow Vault's Identity Token template examples.

### What should be done in the future

- Prove a real GitHub Actions login and token mint from the `glazed` repository after the workflow commits are pushed.
- Keep using the discovery issuer `https://vault.yolo.scapegoat.dev/v1/identity/oidc` for docs-registry JWT validation.

### Code review instructions

- Review Terraform commit `2e56b7eb4cefd543df7e789af26b2eaedaf39c2a`.
- Validate with:

```bash
cd /home/manuel/code/wesen/terraform
source ./.envrc
cd /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s
terraform plan -input=false -detailed-exitcode
```

### Technical details

The corrected token template uses this form:

```json
{
  "repository": {{identity.entity.aliases.<accessor>.metadata.repository}}
}
```

not this form:

```json
{
  "repository": "{{identity.entity.aliases.<accessor>.metadata.repository}}"
}
```

## Step 10: Add the reusable docsctl publish workflow and Glazed caller workflow

I added the reusable GitHub Actions workflow that callers can use to export a Glazed help SQLite database, authenticate to Vault with GitHub OIDC, mint a package-scoped docs-registry publish JWT, and run `docsctl publish --token-file`. I then added a Glazed caller workflow that uses that reusable workflow for the `glazed` package.

This completes the static implementation of Phase 3. The next proof step requires pushing the infra-tooling and Glazed workflow commits so GitHub Actions can produce a real OIDC token with the expected `workflow_ref` and `job_workflow_ref` claims.

### Prompt Context

**User prompt (verbatim):** (same as Step 9)

**Assistant interpretation:** Continue after Terraform apply by implementing the reusable CI workflow and a first package caller.

**Inferred user intent:** The user wants the Vault resources exercised by the intended GitHub Actions path rather than left as unused infrastructure.

**Commit (code):** `a95c5d5a08539d6b691b0c1ebb4c086132707808` — `actions: add docsctl publish workflow`

**Commit (code):** `209ee884288cf086a4751044040c244d98aa61d2` — `ci: add docsctl publish workflow`

### What I did

- Created `/home/manuel/code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml`.
- Created `/home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/.github/workflows/publish-docs.yml`.
- The reusable workflow accepts inputs for:
  - package name;
  - package version;
  - export command;
  - SQLite path;
  - docsctl install command;
  - Vault address/auth path/role/audience/token role;
  - docs-registry URL;
  - docs API verification URL.
- The workflow performs:
  - checkout;
  - Go setup;
  - docsctl install;
  - SQLite export;
  - Vault login using `hashicorp/vault-action@v3` and GitHub OIDC;
  - publish JWT mint via `GET /v1/identity/oidc/token/<role>`;
  - non-sensitive JWT claim decoding for logs;
  - `docsctl publish --token-file`;
  - post-publish package/version verification through the docs API.
- The Glazed caller workflow uses:

```text
go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@main
```

and exports with:

```bash
go run ./cmd/glaze help export --format sqlite --output-path .docsctl/help.sqlite
```

### Why

- This workflow is the bridge between the Terraform Vault roles and the docs-registry JWT auth implementation.
- Keeping publish logic in `infra-tooling` avoids duplicating the Vault/token/publish dance across all package repositories.

### What worked

- Both workflow files were created and committed in focused commits.
- The Glazed caller workflow path matches the Terraform-bound caller workflow path:

```text
go-go-golems/glazed/.github/workflows/publish-docs.yml@refs/heads/main
```

### What didn't work

- I have not run a live GitHub Actions proof yet. That requires the commits to be pushed to their corresponding GitHub repositories, and the reusable workflow must be addressable at the ref used by the caller.

### What I learned

- The live proof is the next point where the exact GitHub OIDC claim values will be confirmed. Static workflow files can be reviewed locally, but they cannot prove `job_workflow_ref` until GitHub runs the reusable workflow.

### What was tricky to build

- The reusable workflow needs to avoid exposing the publish JWT while still leaving enough evidence for debugging. It masks the JWT and prints only non-sensitive decoded claims.
- Defaulting the Vault role and token role from `package_name` keeps callers simple while still allowing overrides.

### What warrants a second pair of eyes

- Review whether `go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@main` will produce `job_workflow_ref` with `@refs/heads/main`, as Terraform expects.
- Review whether the public registry URL should be `https://docs-registry.yolo.scapegoat.dev`; if the registry is not exposed yet, Phase 4 will fail until ingress or a self-hosted runner path exists.
- Review whether `workflow_dispatch` with `package_version = manual` should publish to `manual` or a test version like `vtest-jwt`.

### What should be done in the future

- Push `infra-tooling` commit `a95c5d5a08539d6b691b0c1ebb4c086132707808` to `main`.
- Push Glazed commit `209ee884288cf086a4751044040c244d98aa61d2` to the branch/repository intended for proof.
- Run the Glazed workflow manually with a test package version.

### Code review instructions

- Start with the reusable workflow in `infra-tooling`.
- Then review the Glazed caller workflow and compare its path/ref to the Terraform `workflow_ref` bound claim.
- The live validation step is a GitHub Actions run, not a local unit test.

### Technical details

The reusable workflow mints the publish JWT with:

```bash
curl --fail-with-body --silent --show-error \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${VAULT_ADDR}/v1/identity/oidc/token/${VAULT_TOKEN_ROLE}"
```
