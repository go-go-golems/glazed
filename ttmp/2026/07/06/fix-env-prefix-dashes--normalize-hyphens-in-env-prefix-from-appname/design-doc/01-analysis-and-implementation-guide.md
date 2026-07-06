---
Title: Analysis and Implementation Guide
Ticket: fix-env-prefix-dashes
Status: active
Topics:
    - env
    - bug
    - sources
    - parsing
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cli/cobra-parser.go
      Note: |-
        built-in path passes strings.ToUpper(AppName) as the env prefix
        built-in path passes strings.ToUpper(AppName) as env prefix
    - Path: pkg/cmds/sources/update.go
      Note: |-
        updateFromEnv computes the env key; prefix is not hyphen-normalized
        updateFromEnv is the bug site and the single shared env-key helper
    - Path: pkg/cmds/sources/update_test.go
      Note: |-
        existing env source tests (patterns to mirror)
        existing env source tests to mirror
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/596
Summary: Env var prefix derived from a hyphenated AppName keeps its hyphen, producing unexportable env var names. Fix by normalizing the prefix the same way the field name already is.
LastUpdated: 2026-07-06T00:00:00Z
WhatFor: 'Fixing the env-source prefix normalization bug reported in issue #596'
WhenToUse: When AppName contains hyphens and the built-in env source silently fails to match
---


# Analysis and Implementation Guide

## Executive Summary

When a command sets `AppName` to a hyphenated value (e.g. `"llm-proxy"`), the built-in
env source builds env-var keys like `LLM-PROXY_BYOK_SECRET`. Most shells cannot export a
variable whose name contains a hyphen, so the env source silently never matches for
hyphenated app names. The field-name portion of the env key is already hyphen→underscore
normalized; only the app-name prefix is not. The fix is to normalize the prefix the same
way the field name is already normalized, in the single shared `updateFromEnv` helper.

## Problem Statement

`updateFromEnv` in `pkg/cmds/sources/update.go` computes the env key for each field:

```go
base := sectionPrefix + p.Name
envKey := strings.ToUpper(strings.ReplaceAll(base, "-", "_"))   // field name: hyphens normalized
if prefix != "" {
    envKey = strings.ToUpper(prefix) + "_" + envKey              // prefix: hyphens NOT normalized
}
```

Line 157 normalizes hyphens in `base` (section prefix + field name); line 159 only
uppercases the app-name `prefix` and does **not** run `ReplaceAll(prefix, "-", "_")`.

The built-in cobra parser path computes the prefix as
`envPrefix := strings.ToUpper(cfgCopy.AppName)` (`pkg/cli/cobra-parser.go:162`), so a
hyphenated `AppName` flows straight through into a hyphenated env prefix.

### Reproduction (from issue #596)

```go
parserConfig := cli.CobraParserConfig{ AppName: "llm-proxy" } // hyphenated
```

```sh
# Does NOT load — shell can't export the hyphenated name:
export LLM_PROXY_BYOK_SECRET=sk-test
./myapp --print-parsed-fields   # byok-secret is empty

# DOES load (quoted, bypasses the shell parser):
env 'LLM-PROXY_BYOK_SECRET=sk-test' ./myapp --print-parsed-fields   # byok-secret = sk-test
```

## Current-State Architecture (evidence)

- `pkg/cmds/sources/update.go:143` `updateFromEnv(schema_, parsedValues, prefix, options...)`
  is the single shared helper invoked by the `FromEnv(prefix)` middleware
  (`pkg/cmds/sources/update.go:213`).
- Callers that supply the prefix:
  - `pkg/cli/cobra-parser.go:162` — built-in path: `strings.ToUpper(cfgCopy.AppName)`.
  - `pkg/cmds/runner/run.go:170` — `cmd_sources.FromEnv(opts.EnvPrefix, ...)`.
  - `pkg/cmds/helpers/test-helpers.go:215` — `sources.FromEnv(*m.Prefix, ...)`.
  - `pkg/cmds/sources/vault.go:83` — `FromEnv(prefix, ...)`.
  - Several `cmd/` binaries call `FromEnv` with literal, already-underscored prefixes
    (`DOCSCTL`, `WEB`, `BUILD_WEB`, `APP`).
- Only `updateFromEnv` assembles the final env key, so it is the natural single choke point.

## Proposed Solution

Normalize the prefix the same way `base` is normalized — in `updateFromEnv`, so the fix
applies to **every** caller of `FromEnv`, not only the cobra built-in path:

```go
if prefix != "" {
    envKey = strings.ToUpper(strings.ReplaceAll(prefix, "-", "_")) + "_" + envKey
}
```

After this change, `AppName: "llm-proxy"` yields `LLM_PROXY_BYOK_SECRET`, which is
shell-exportable and consistent with how the field-name portion is already normalized.

## Design Decisions

- **Fix at `updateFromEnv`, not at `cobra-parser.go`.** Normalizing in the helper covers
  all callers (runner, helpers, vault, direct `cmd/` usage) and keeps the env-key
  assembly logic in one place. The cobra path continues to pass `strings.ToUpper(AppName)`
  and the helper makes it shell-safe. This matches the fix suggested in the issue.
- **No behavior change for already-underscored / all-caps prefixes.** `ReplaceAll` is a
  no-op when there is no hyphen, so existing callers (`DOCSCTL`, `WEB`, `APP`, …) are
  unaffected.
- **No backwards-compat shim.** A hyphenated prefix previously produced unexportable env
  var names, so no real-world env vars could have been matching via that prefix; there is
  nothing to preserve. (Per project guidelines, no compat layer is added.)

## Alternatives Considered

- Normalize in `cobra-parser.go` only. Rejected: leaves direct `FromEnv(...)` callers
  (vault, runner, helpers, `cmd/` binaries) still vulnerable to hyphenated prefixes.
- Require callers to pass underscore-separated `AppName`. Rejected: pushes the burden onto
  every consumer and widens the config surface (see issue workaround).

## Implementation Plan

1. Edit `pkg/cmds/sources/update.go` `updateFromEnv`: normalize the `prefix` with
   `strings.ReplaceAll(prefix, "-", "_")` before uppercasing.
2. Add a regression test in `pkg/cmds/sources/update_test.go` mirroring
   `TestUpdateFromEnvParsesTypedValues`, but using a hyphenated prefix (`"llm-proxy"`)
   and asserting it matches `LLM_PROXY_*` env vars (and that no `LLM-PROXY_*` lookup is
   needed).
3. Run `gofmt`, `go test ./pkg/cmds/sources/... -count=1`, then the full
   `go test ./... -count=1`.
4. Commit the fix + test, then open a PR against `go-go-golems/glazed` referencing #596.

## Testing and Validation

- New unit test asserts `FromEnv("llm-proxy")` picks up `LLM_PROXY_CFG_USER` and records
  the `env_key` metadata as `LLM_PROXY_CFG_USER`.
- Existing `TestUpdateFromEnvParsesTypedValues` / `TestUpdateFromEnvInvalidChoice`
  continue to pass (they use the non-hyphenated `"APP"` prefix).
- `go test ./... -count=1` is the acceptance gate.

## Open Questions

None. The fix is localized and the suggested approach in the issue is correct.

## References

- Issue: https://github.com/go-go-golems/glazed/issues/596
- `pkg/cmds/sources/update.go` (`updateFromEnv`)
- `pkg/cli/cobra-parser.go` (built-in env prefix from `AppName`)
