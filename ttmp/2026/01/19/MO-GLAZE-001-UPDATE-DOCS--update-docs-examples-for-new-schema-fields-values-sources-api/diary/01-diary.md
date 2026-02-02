---
Title: Diary
Ticket: MO-GLAZE-001-UPDATE-DOCS
Status: active
Topics:
    - docs
    - examples
    - glazed
    - schema
    - values
    - sources
    - migration
DocType: diary
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: .ttmp.yaml
      Note: Docmgr config normalized to relative paths
    - Path: cmd/examples/new-api-build-first-command/main.go
      Note: New minimal wrapper-API example (build-first-command)
    - Path: cmd/examples/new-api-dual-mode/main.go
      Note: New dual-mode wrapper-API example
    - Path: cmd/examples/register-cobra/main.go
      Note: Updated to use fields/schema/values vocabulary
    - Path: cmd/examples/sources-example/main.go
      Note: Updated to use values in BareCommand + sources helpers
    - Path: pkg/doc/tutorials/05-build-first-command.md
      Note: Updated tutorial to use schema/fields/values wrappers
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T11:09:28.893593848-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Update Glazed docs and example programs to use the newer façade packages (`schema`, `fields`, `values`, `sources`) wherever possible, and verify all updated examples compile.

## Step 1: Update “Build Your First Glazed Command” tutorial to wrapper API

The tutorial in `pkg/doc/tutorials/05-build-first-command.md` was still teaching the older vocabulary (`layers`, `parameters`, direct `settings.NewGlazedParameterLayers`). I updated it to use the wrapper packages introduced in ticket `001-REFACTOR-NEW-PACKAGES` so readers see the new API surface first.

The goal here was not to change behavior, but to update imports, decoding language, and code blocks so the tutorial aligns with the current recommended vocabulary and the repo’s Go version.

**Commit (code):** 9d2e3c07e7205fec39b4823229b22940a7fb5cc3 — "docs: update build-first-command for new schema/fields/values API"

### What I did
- Updated the tutorial code blocks to import and use `pkg/cmds/schema`, `pkg/cmds/fields`, `pkg/cmds/values`.
- Replaced `parsedLayers.InitializeStruct(schema.DefaultSlug, ...)` examples with `values.DecodeSectionInto(vals, schema.DefaultSlug, ...)`.
- Updated parameter definitions from `fields.New(...)` to `fields.New(...)`.
- Updated short-help layer references to `schema.DefaultSlug`.
- Matched the tutorial’s Go prerequisite to `go.mod` (Go 1.25+).
- Compile-checked the tutorial code by assembling it into a standalone temp module with a `go.mod` `replace` to the local checkout and running `go build ./...`.

### Why
- The new wrapper packages exist specifically to make onboarding vocabulary clearer without breaking existing code, so tutorials should prefer them.
- Keeping tutorial examples compiling is the fastest way to prevent drift during refactors.

### What worked
- The wrapper API covered the tutorial’s needs directly (`settings.NewGlazedSchema`, `fields.New`, `values.DecodeSectionInto`).
- The assembled quickstart program compiled cleanly against the local checkout.

### What didn't work
- Cleanup of the temp compile directory was blocked by policy:
  - `/usr/bin/zsh -lc 'rm -rf /tmp/glazed-build-first-command.QbuBrs' rejected: blocked by policy`

### What I learned
- `cli.BuildCobraCommand` auto-adds the glazed schema section for `GlazeCommand` implementations, so tutorials/examples can often omit the explicit `settings.NewGlazedSchema()` unless they want to show the schema composition explicitly.

### What was tricky to build
- Keeping examples “new API vocabulary” while still reflecting the underlying alias reality (e.g. some glue points still talk about `layers` in other parts of the codebase).

### What warrants a second pair of eyes
- The tutorial is long and includes multiple embedded examples (dual-mode, advanced parameter types). A reviewer should skim for any remaining stale wording (“layers/parsedLayers”) that might confuse readers even if code compiles.

### What should be done in the future
- N/A

### Code review instructions
- Start at `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/tutorials/05-build-first-command.md`.
- Validate by compiling the tutorial code blocks (or re-running the temp-module approach) and by comparing vocabulary against the wrapper package design in ticket `001-REFACTOR-NEW-PACKAGES`.

### Technical details
- Compile-check approach used a temp module with:
  - `go mod edit -replace github.com/go-go-golems/glazed=/home/manuel/code/wesen/corporate-headquarters/glazed`
  - `go mod tidy`
  - `go build ./...`

## Step 2: Add wrapper-API examples and modernize existing cmd/examples

After the tutorial was updated, the next gap was runnable examples. I added two new example programs focused on the wrapper API and then modernized most existing `cmd/examples/*` programs to use the new vocabulary where feasible (especially `schema`, `fields`, `values`, and `sources`).

This keeps the examples directory consistent with the tutorial and reduces confusion for newcomers who browse examples first.

**Commit (code):** 1291a92d2fdd2092e582a74406d53cfdcea31dc5 — "examples: add wrapper-API examples and modernize existing ones"

### What I did
- Added new runnable examples:
  - `cmd/examples/new-api-build-first-command` (minimal list-users with structured output)
  - `cmd/examples/new-api-dual-mode` (dual-mode command with `--with-glaze-output`)
- Updated existing examples to use wrapper vocabulary where possible:
  - `cmd/examples/register-cobra` now uses `fields/schema/values` and relies on the builder to inject glazed flags for `GlazeCommand`.
  - config/env/config-mapper examples now use `schema.NewSection`, `schema.WithFields`, `fields.New`, `values.DecodeSectionInto`, and `sources.*` helpers where applicable.
- Ran compile checks:
  - `go test ./cmd/examples/...`
  - `go test ./...`

### Why
- Examples are “copy/paste entry points”; they should teach the preferred import paths and naming.
- Updating existing examples prevents the examples directory from becoming a grab bag of old/new styles.

### What worked
- `sources` wrappers were sufficient for the common “defaults/config/env/flags” chains in examples without reaching into `cmds/middlewares` directly.
- Repository-wide compile/test passed:
  - `go test ./cmd/examples/...`
  - `go test ./...`
- Pre-commit checks passed after formatting fixes (gofmt + golangci-lint + gosec + govulncheck as configured by lefthook).

### What didn't work
- The first commit attempt failed due to gofmt issues flagged by golangci-lint:
  - `cmd/examples/config-pattern-mapper/main.go` (gofmt)
  - `cmd/examples/new-api-build-first-command/main.go` (gofmt)
  - `cmd/examples/new-api-dual-mode/main.go` (gofmt)
- Fixed by running:
  - `gofmt -w cmd/examples/config-pattern-mapper/main.go cmd/examples/new-api-build-first-command/main.go cmd/examples/new-api-dual-mode/main.go`

### What I learned
- The lefthook pre-commit pipeline runs a pretty thorough set of checks; formatting needs to be kept tight to avoid slow/failed commits.

### What was tricky to build
- Some examples still legitimately need the underlying packages (e.g. `cmds/parameters` for `FileData`, or `cmds/layers` for cobra-layer plumbing), so “new API everywhere” has to be applied pragmatically rather than dogmatically.

### What warrants a second pair of eyes
- The examples touched a wide surface area; a reviewer should spot-check that the updated examples still reflect the recommended practices (especially around config precedence and when to rely on `cli.BuildCobraCommand` auto-injecting glazed layers).

### What should be done in the future
- Consider consolidating overlapping config examples if the project wants a smaller “golden set” (e.g. pick one canonical config overlay example and link to it from others).

### Code review instructions
- Start at:
  - `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/examples/new-api-build-first-command/main.go`
  - `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/examples/new-api-dual-mode/main.go`
- Then review the updated examples:
  - `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/examples/register-cobra/main.go`
  - `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/examples/sources-example/main.go`
  - `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/examples/middlewares-config-env/main.go`
- Validate with:
  - `go test ./cmd/examples/...`
  - `go test ./...`

### Technical details
- Wrapper packages used:
  - `schema.NewSection`, `schema.WithFields`, `settings.NewGlazedSchema`
  - `fields.New`, `fields.Type*`
  - `values.DecodeSectionInto`
  - `sources.FromCobra`, `sources.FromEnv`, `sources.FromFile`, `sources.FromDefaults`, `sources.Execute`

