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

_ = parser.Register("redis", redisLayer, func(t *AppSettings) any { return &t.Redis })
cfg, err := parser.Parse()
_ = cfg
_ = err
```

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
