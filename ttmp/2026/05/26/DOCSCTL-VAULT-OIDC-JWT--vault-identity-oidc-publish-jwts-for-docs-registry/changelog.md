# Changelog

## 2026-05-26

- Initial workspace created


## 2026-05-26

Created Vault Identity/OIDC publish JWT design, cryptography primer, Terraform/API sketches, registry implementation plan, and diary.

### Related Files

- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Terraform pattern to extend
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/auth.go — PublisherAuth extension point
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/design-doc/01-vault-identity-oidc-publish-jwt-implementation-guide.md — Primary implementation guide
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/reference/01-diary.md — Investigation diary


## 2026-05-26

Normalized source files, validated ticket, uploaded bundle to reMarkable, and completed tasks.

### Related Files

- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/reference/01-diary.md — Recorded validation and upload step
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/sources/terraform-vault-oidc-resource-schema.txt — Terraform provider schema retained as source evidence
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/ttmp/2026/05/26/DOCSCTL-VAULT-OIDC-JWT--vault-identity-oidc-publish-jwts-for-docs-registry/tasks.md — All ticket tasks completed


## 2026-05-26

Implemented Phase 1 docs-registry Vault OIDC JWT auth and Phase 2 Terraform Vault publish roles (Glazed commit aa6946a40f2156689e81a831a10e634398102261; Terraform commit 04451fe795314065d872f22c8710044682525963).

### Related Files

- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Vault docsctl publish roles
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/publish/jwt_auth.go — Registry JWT auth implementation


## 2026-05-26

Applied Vault docsctl publish JWT resources, fixed live Vault issuer/template validation, and added reusable/caller GitHub Actions workflows (Terraform commit 2e56b7eb4cefd543df7e789af26b2eaedaf39c2a; infra-tooling commit a95c5d5a08539d6b691b0c1ebb4c086132707808; Glazed commit 209ee884288cf086a4751044040c244d98aa61d2).

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml — Reusable publish workflow
- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Applied Vault docsctl publish resources
- /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/.github/workflows/publish-docs.yml — Glazed caller workflow

