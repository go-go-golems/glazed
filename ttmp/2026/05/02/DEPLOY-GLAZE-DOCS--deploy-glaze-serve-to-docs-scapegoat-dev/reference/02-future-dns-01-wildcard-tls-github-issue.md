---
Title: Future DNS-01 Wildcard TLS GitHub Issue
Ticket: DEPLOY-GLAZE-DOCS
Status: active
Topics:
    - deployment
    - kubernetes
    - dns
    - glazed
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/glaze-docs/ingress.yaml
      Note: Current concrete HTTP-01 TLS decision
    - Path: ../../../../../../../../../../code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/platform-cert-issuer/clusterissuer.yaml
      Note: Current HTTP-01-only letsencrypt-prod issuer and future DNS-01 target
    - Path: ../../../../../../../../../../code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf
      Note: Wildcard docs DNS record
ExternalSources: []
Summary: GitHub issue text for adding DNS-01 wildcard TLS support to letsencrypt-prod later.
LastUpdated: 2026-05-02T11:09:48.044749297-04:00
WhatFor: Copy/paste source for the future GitHub issue about cert-manager DNS-01 support.
WhenToUse: Use when revisiting wildcard certificates for *.docs.scapegoat.dev or other wildcard TLS needs.
---


# Future DNS-01 Wildcard TLS GitHub Issue

## Goal

Capture the future work needed to add DNS-01 support to the cluster-wide `letsencrypt-prod` cert-manager `ClusterIssuer`, so wildcard certificates such as `*.docs.scapegoat.dev` can be issued safely.

For the current Glaze deployment we intentionally keep the simpler HTTP-01 pattern: wildcard DNS exists, but the Ingress requests a concrete certificate only for `glaze.docs.scapegoat.dev`.

## Context

The current GitOps-managed issuer lives at:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/platform-cert-issuer/clusterissuer.yaml`

It currently supports HTTP-01 only:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: wesen@ruinwesen.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod-account-key
    solvers:
      - http01:
          ingress:
            ingressClassName: traefik
```

HTTP-01 is enough for concrete host certificates such as:

- `glaze.docs.scapegoat.dev`
- `artifacts.yolo.scapegoat.dev`
- `goja.yolo.scapegoat.dev`

HTTP-01 cannot issue wildcard certificates such as:

- `*.docs.scapegoat.dev`
- `*.yolo.scapegoat.dev`

Wildcard certificates require DNS-01 validation.

## GitHub issue title

Add DNS-01 support to cert-manager letsencrypt-prod for wildcard docs certificates

## GitHub issue body

## Summary

Add DNS-01 support to the GitOps-managed cert-manager `ClusterIssuer/letsencrypt-prod`, primarily so the cluster can issue wildcard certificates for the docs subtree, such as `*.docs.scapegoat.dev`.

For now, the Glaze docs deployment should keep using the existing HTTP-01 pattern with an explicit concrete TLS hostname:

```yaml
tls:
  - hosts:
      - glaze.docs.scapegoat.dev
    secretName: glaze-docs-tls
```

The wildcard DNS record can still exist:

```text
*.docs.scapegoat.dev -> 91.98.46.169
```

This mirrors the current `*.yolo.scapegoat.dev` pattern: wildcard DNS points at the cluster, while each app usually gets its own concrete cert through HTTP-01.

## Current state

The issuer is defined in:

```text
gitops/kustomize/platform-cert-issuer/clusterissuer.yaml
```

Current solver configuration:

```yaml
solvers:
  - http01:
      ingress:
        ingressClassName: traefik
```

This means:

- Concrete certs work, assuming DNS points at Traefik and HTTP routing succeeds.
- Wildcard certs do not work.
- A Certificate or Ingress TLS request containing `*.docs.scapegoat.dev` will fail unless DNS-01 is added.

## Desired future state

Keep HTTP-01 as the default/fallback for existing apps, but add a DigitalOcean DNS-01 solver for the `docs.scapegoat.dev` zone.

Proposed shape:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
  annotations:
    argocd.argoproj.io/sync-wave: "-2"
spec:
  acme:
    email: wesen@ruinwesen.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod-account-key
    solvers:
      - selector:
          dnsZones:
            - docs.scapegoat.dev
        dns01:
          digitalocean:
            tokenSecretRef:
              name: digitalocean-dns
              key: access-token
      - http01:
          ingress:
            ingressClassName: traefik
```

Why use `dnsZones` instead of only `dnsNames`:

- `dnsZones: [docs.scapegoat.dev]` makes the intent explicit for the entire docs subtree.
- It covers future hosts under `docs.scapegoat.dev` without listing every name.
- It avoids ambiguity when a certificate contains both a wildcard and concrete names.

## Required secret

cert-manager's DigitalOcean DNS-01 solver needs an API token secret in the `cert-manager` namespace.

Example manual bootstrap:

```bash
kubectl -n cert-manager create secret generic digitalocean-dns \
  --from-literal=access-token="$DIGITALOCEAN_TOKEN"
```

Do not commit a raw `dop_v1_...` token to Git.

Possible secret management options:

- Manual one-time bootstrap with `kubectl create secret`.
- Vault Secrets Operator, if we want the token to be sourced from Vault.
- SOPS/SealedSecrets, if we standardize on encrypted GitOps secrets later.

## Security notes

The DigitalOcean token is sensitive. Anyone with it can modify DNS records for managed zones.

Before implementing, decide:

- Can DigitalOcean provide a token scoped narrowly enough for DNS only?
- Should the secret be manually bootstrapped or Vault-managed?
- Who is allowed to rotate the token?
- Where is token rotation documented?

## Validation plan

After adding the DNS-01 solver and secret:

```bash
kubectl get clusterissuer letsencrypt-prod -o yaml
kubectl -n cert-manager get secret digitalocean-dns
```

Create a test Certificate in a non-critical namespace:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: docs-wildcard-test
  namespace: glaze-docs
spec:
  secretName: docs-wildcard-test-tls
  issuerRef:
    kind: ClusterIssuer
    name: letsencrypt-prod
  dnsNames:
    - "*.docs.scapegoat.dev"
```

Watch cert-manager resources:

```bash
kubectl -n glaze-docs get certificate,certificaterequest,order,challenge
kubectl -n cert-manager logs deploy/cert-manager --tail=200
```

Expected outcome:

- Certificate becomes `Ready=True`.
- A DNS-01 challenge is created.
- DigitalOcean temporarily receives the `_acme-challenge.docs.scapegoat.dev` TXT record.

## Important certificate semantics

A wildcard certificate for:

```text
*.docs.scapegoat.dev
```

covers:

```text
glaze.docs.scapegoat.dev
foo.docs.scapegoat.dev
bar.docs.scapegoat.dev
```

It does not cover:

```text
docs.scapegoat.dev
```

If the bare docs host is needed, request both:

```yaml
dnsNames:
  - docs.scapegoat.dev
  - "*.docs.scapegoat.dev"
```

## Interaction with Glaze docs deployment

Current desired Glaze Ingress for HTTP-01-only operation:

```yaml
spec:
  tls:
    - hosts:
        - glaze.docs.scapegoat.dev
      secretName: glaze-docs-tls
  rules:
    - host: glaze.docs.scapegoat.dev
```

After DNS-01 support exists, we can decide whether to switch to wildcard TLS:

```yaml
spec:
  tls:
    - hosts:
        - "*.docs.scapegoat.dev"
      secretName: docs-scapegoat-dev-wildcard-tls
```

Caveat: Kubernetes TLS secrets are namespace-scoped. If future docs apps live in separate namespaces, each namespace either needs its own Certificate/secret or we need a deliberate secret replication strategy.

## Acceptance criteria

- [ ] `letsencrypt-prod` keeps working for existing HTTP-01 app certificates.
- [ ] DigitalOcean DNS token is available to cert-manager without committing the raw token to Git.
- [ ] `letsencrypt-prod` has a DNS-01 solver for `docs.scapegoat.dev`.
- [ ] A test `Certificate` for `*.docs.scapegoat.dev` becomes `Ready=True`.
- [ ] Documentation explains how to rotate the DigitalOcean DNS token.
- [ ] Glaze docs deployment remains on explicit `glaze.docs.scapegoat.dev` TLS until wildcard TLS is deliberately enabled.

## Related files

- `gitops/kustomize/platform-cert-issuer/clusterissuer.yaml`
- `gitops/kustomize/glaze-docs/ingress.yaml`
- `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`
- docmgr ticket: `DEPLOY-GLAZE-DOCS`

## Usage Examples

Create the issue from the GitOps repo with:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
gh issue create \
  --title "Add DNS-01 support to cert-manager letsencrypt-prod for wildcard docs certificates" \
  --body-file /tmp/dns01-wildcard-docs-issue.md
```

## Related

- `design-doc/01-production-deployment-guide.md`
- `reference/01-diary.md`
