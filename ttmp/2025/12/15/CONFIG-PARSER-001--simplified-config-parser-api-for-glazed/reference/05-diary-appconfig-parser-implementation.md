---
Title: 'Diary: appconfig.Parser implementation'
Ticket: CONFIG-PARSER-001
Status: active
Topics:
    - glazed
    - config
    - api-design
    - parsing
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/appconfig/doc.go
      Note: Package docs + scope statement (commit bf627f0)
    - Path: glazed/pkg/appconfig/options.go
      Note: ParserOption helpers (env/config files/middlewares) (commit bf627f0)
    - Path: glazed/pkg/appconfig/parser.go
      Note: Core appconfig.Parser[T] implementation (commit bf627f0)
    - Path: glazed/pkg/appconfig/parser_test.go
      Note: Unit tests for Register/Parse invariants, precedence, and hydration behavior (commit d452edc)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-16T00:00:00Z
---


# Diary: appconfig.Parser implementation

## Goal

Document the step-by-step implementation of `glazed/pkg/appconfig` (specifically `appconfig.Parser[T]`) for CONFIG-PARSER-001: what changed, why it changed, what worked/didn’t, and the decisions we made while translating the design into working code.

## Step 1: Reset package placement and start a v1 `pkg/appconfig` skeleton

This step starts the actual implementation work. The key change vs the design’s earlier assumptions is package placement: `glazed/pkg/config` is already scoped to config files, so we’re creating a dedicated `glazed/pkg/appconfig` package to host `appconfig.Parser[T]`. The goal for this step is to get a compiling skeleton in place (constructor + registration + parse pipeline wiring) without overreaching into “struct-first schema derivation” again.

**Commit (code):** bf627f0438d67ab5457de0ecf22882516a823cdd — "feat: add appconfig.Parser v1 skeleton"

### What I did

- Updated docs/tasks to reflect that `appconfig.Parser` will live in `glazed/pkg/appconfig`.
- Implemented `glazed/pkg/appconfig`:
  - `pkg/appconfig/doc.go`
  - `pkg/appconfig/options.go`
  - `pkg/appconfig/parser.go`
- Reused `runner.ParseCommandParameters` by creating a tiny `cmds.Command` stub (only `Description()` + `ToYAML()`).
- Ran:
  - `gofmt -w pkg/appconfig/*.go`
  - `go test ./... -count=1`
- Committed the v1 skeleton (`bf627f0`).

### Why

Keeping `appconfig.Parser` separate avoids mixing concerns (config-file tooling vs config parsing façade). It also allows us to evolve `appconfig.Parser` without worrying about `pkg/config` API expectations.

### What worked

- `cmds.Command` is minimal (`Description()` + `ToYAML()`), so a safe stub is easy.
- The v1 package compiles and full test suite passes (`go test ./... -count=1`).

### What didn't work

- A normal `git commit` was blocked by `lefthook` running `govulncheck ./...`, which exits non-zero due to Go stdlib vulnerabilities in the installed toolchain.
  - Fix is not in this repo; the tool output indicated the standard library needs a newer Go patchlevel.
  - For now (per instruction), I committed with `LEFTHOOK=0`.

### What I learned

- We can reuse `runner.ParseCommandParameters` without having to implement any `Run()` methods (it only needs `cmd.Description().Layers`).

### What was tricky to build

- Ensuring we keep the “incremental appconfig.Parser” scope: **explicit layer + tagged struct hydration** only; no schema generation.
- Avoiding slug mismatches: v1 `Register` requires `slug == layer.GetSlug()` so parsed layer keys line up with registration keys.

### What warrants a second pair of eyes

- Middleware ordering/precedence: we should follow runner’s exact ordering to avoid subtle precedence bugs.
- The v1 option composition: we currently translate `appconfig` options into runner `ParseOption`s; confirm we don’t accidentally invert precedence with additional middlewares.

### What should be done in the future

- Add a second diary step once the first v1 code lands, including tests and precedence verification.

### Code review instructions

- Start with `glazed/pkg/appconfig/parser.go` and verify the Parse pipeline matches runner behavior.
- Then review `glazed/pkg/appconfig/options.go` to confirm option mapping and invariants.


## Quick Reference

V1 usage sketch:

```go
const RedisSlug appconfig.LayerSlug = "redis"

type AppSettings struct {
	Redis RedisSettings
}

type RedisSettings struct {
	Host string `glazed.parameter:"host"`
}

parser, _ := appconfig.NewParser[AppSettings](
	appconfig.WithEnv("MYAPP"),
	appconfig.WithConfigFiles("base.yaml"),
)

_ = parser.Register(RedisSlug, redisLayer, func(t *AppSettings) any { return &t.Redis })
cfg, err := parser.Parse()
_ = cfg
_ = err
```

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->

## Step 2: Add P2 tests (registration invariants, precedence, hydration) and validate behavior

This step adds a first meaningful safety net around the new `appconfig.Parser` API. The tests focus on the intended v1 contracts: registration invariants, binder failure behavior, precedence across defaults/config/env, and the “tag-required” hydration semantics. This is important because the API is a façade over Glazed’s existing runner + `InitializeStruct` behavior; tests make that coupling explicit and prevent accidental regressions.

**Commit (code):** d452edccbb91da12228ccac1957389999cd0996c — "test: add appconfig.Parser unit tests"

### What I did

- Added `glazed/pkg/appconfig/parser_test.go` with:
  - Register validation tests (empty slug, nil layer, nil bind, slug mismatch, duplicate slug)
  - Binder failure tests (bind returns nil, non-pointer, nil pointer)
  - Hydration test demonstrating **tag-required** behavior (no `glazed.parameter` tags → zero values)
  - Precedence test: defaults < config file < env
- Ran:
  - `gofmt -w pkg/appconfig/*.go`
  - `go test ./... -count=1`
- Committed tests with `LEFTHOOK=0` (same rationale as Step 1).

### Why

These tests encode the v1 “contract” in executable form, so future work (examples, API polish, CLI adapter) doesn’t silently change:

- which registration patterns are allowed,
- how precedence is composed,
- and how hydration behaves when tags/params don’t line up.

### What worked

- All tests passed, including the env>config precedence case.

### What didn't work

- Same as Step 1: pre-commit `govulncheck` currently blocks go-file commits in this environment; using `LEFTHOOK=0` is the temporary workaround.

### What I learned

- The runner + default config-file loader expects the config shape:
  - `layer-slug: { param-name: value }`
  - This is compatible with our v1 direction (“explicit layers”), so we don’t need mapping machinery yet.

### What was tricky to build

- Getting precedence tests right without accidentally introducing layer prefixes (env naming uses layer prefix, not slug).

### What warrants a second pair of eyes

- Confirm the decision “require slug == layer.GetSlug()” is acceptable for real-world layers (some wrapper layers might have mismatched registration slugs).

### What should be done in the future

- Add table-driven tests for multi-layer env collisions (two layers with same param name + empty prefixes) to clarify expected behavior.

### Code review instructions

- Start with `glazed/pkg/appconfig/parser_test.go` to understand the v1 contract and invariants.

## Step 3: Introduce `appconfig.LayerSlug` to encourage `const` slugs in caller code

This step is a small API ergonomics tweak: we want callers to declare slugs as constants (and not pass ad-hoc string literals everywhere). To nudge that from day one, we introduced a dedicated `appconfig.LayerSlug` type and updated `Register` to accept it. Importantly, we verified that Glazed does not already define a shared `LayerSlug` type elsewhere, so we’re not duplicating an existing concept.

**Commit (code):** 91b10b2caa2ef92b94a13c9d6217727199d1e676 — "refactor: introduce appconfig.LayerSlug"

### What I did

- Added `type LayerSlug string` in `glazed/pkg/appconfig/parser.go`.
- Changed `Register` signature from `Register(slug string, ...)` to `Register(slug LayerSlug, ...)`.
- Updated internal usage to convert via `string(slug)` where Glazed APIs still expect plain strings.
- Updated tests to declare and use `const redisSlug LayerSlug = "redis"` and pass that into `Register`.
- Ran:
  - `gofmt -w pkg/appconfig/*.go`
  - `go test ./... -count=1`
- Committed with `LEFTHOOK=0` (same temporary workaround as Step 1/2).

### Why

Typed slugs make it slightly harder to accidentally pass the wrong string and make it easy to establish a codebase convention:

- `const RedisSlug appconfig.LayerSlug = "redis"`

### What worked

- Minimal change footprint: only `appconfig` and its tests needed updates.

### What didn't work

- N/A (no failures beyond the known `govulncheck` hook issue).

### What I learned

- We do not currently have a shared slug type in Glazed core packages; slugs are generally plain `string` values today (e.g. `layers.DefaultSlug`).

### What warrants a second pair of eyes

- Whether we should later migrate this type into a shared location (e.g. `pkg/cmds/layers`) if multiple packages want typed slugs.

## Step 4: Add a minimal glazed example using two registered layers

This step adds a tiny runnable example under `glazed/cmd/examples/` that demonstrates the “happy path” for v1: declare `const` slugs, build two layers, register each with a binder into a grouped `AppSettings` struct, call `Parse()`, and print the typed result. The purpose is to give future adopters a copy/paste starting point that doesn’t require reading the internals first.

**Commit (code):** 22c96597ad80833820fbe022dc797d9a49def7eb — "chore: add appconfig.Parser example"

### What I did

- Added: `glazed/cmd/examples/appconfig-parser/main.go`
  - Two layers: `redis` and `db`
  - `const` slugs using `appconfig.LayerSlug`
  - Uses `WithValuesForLayers` to keep it self-contained

### Why

Examples are the fastest path to adoption, and they also serve as an integration-style smoke test for API ergonomics (even without explicit assertions).

### What worked

- `go test ./... -count=1` includes compiling this example, so it stays build-verified.

### What didn't work

- N/A (no additional issues beyond the known `govulncheck` hook behavior for go-file commits).

### What should be done in the future

- Optionally add a second example that demonstrates `WithEnv` + `WithConfigFiles` once we decide the preferred user-facing config file shape for “app settings”.

## Step 5: Parse from Cobra flags/args (and env) via `WithCobra`

This step adds the missing piece for real CLI usage: the ability to have `appconfig.Parser` consume values from a Cobra command (flags + positional args) while still preserving the same precedence model we expect elsewhere (defaults < config < env < args < flags). The implementation is option-driven: `WithCobra(cmd, args)` flips `Parse()` into “cobra mode”, where we build an explicit middleware chain rather than going through `runner.ParseCommandParameters` (which is library-only and doesn’t know about cobra).

**Commit (code):** fbc05d407f07fe5bed57f345de8c900593859430 — "feat: add cobra parsing option to appconfig.Parser"

### What I did

- Added `WithCobra(cmd *cobra.Command, args []string)` option.
- Updated `Parse()`:
  - If cobra mode is enabled, it runs middlewares:
    - cobra flags (highest)
    - positional args
    - env
    - config files
    - provided-values
    - defaults (lowest)
  - Otherwise, it continues to use `runner.ParseCommandParameters`.
- Added a test proving precedence in cobra mode:
  - flags override env (`MYAPP_HOST` vs `--host`)
- Ran:
  - `gofmt -w pkg/appconfig/*.go`
  - `go test ./... -count=1`

### What worked

- The new cobra-mode middleware chain matches expected precedence (flags win).

### What didn't work

- N/A (aside from the known `govulncheck` hook behavior for go-file commits; still using `LEFTHOOK=0` as needed).

### What was tricky to build

- Cobra integration requires that the layer flags have already been added to the cobra command. The parser does not (yet) add flags automatically; it only parses.

### What warrants a second pair of eyes

- The exact placement/precedence of `WithMiddlewares(...)` in cobra mode (we aligned it with runner: “additional” is highest precedence).
