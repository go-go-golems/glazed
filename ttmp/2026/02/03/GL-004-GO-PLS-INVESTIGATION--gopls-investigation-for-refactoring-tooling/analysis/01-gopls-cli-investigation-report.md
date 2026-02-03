---
Title: gopls CLI investigation report
Ticket: GL-004-GO-PLS-INVESTIGATION
Status: active
Topics:
    - gopls
    - refactoring
    - tooling
    - go
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/scripts/gopls-sandbox/lib/lib.go
      Note: Sandbox symbol/rename target
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/01-gopls-help.txt
      Note: CLI command inventory
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/06-gopls-rename-widget-diff.txt
      Note: Sample rename diff
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/10-gopls-stats-anon.json
      Note: Workspace stats sample
ExternalSources: []
Summary: CLI capabilities and refactor implications from gopls experiments.
LastUpdated: 2026-02-03T19:05:00-05:00
WhatFor: Guide refactor tooling design using gopls primitives.
WhenToUse: When planning or validating gopls-based refactor automation.
---


# gopls CLI investigation report

## Goal

Assess how much structured information and refactor capability the gopls command-line interface can provide, and how we can integrate it into future refactoring tooling.

## Environment

- Host: Linux (repo: `/home/manuel/workspaces/2026-02-02/refactor-glazed-names`)
- `gopls version`: `v0.20.0`
- `go version`: `go1.25.5`
- Sandbox module: `ttmp/.../scripts/gopls-sandbox` (see `scripts/`)

## Primary references

- gopls CLI is experimental and intended mainly for debugging; positions are `file:line:column` (UTF‑8 bytes) or `file:#offset`. citeturn3view0
- Feature catalog (rename, references, symbols, code actions, etc.) and where they sit in the LSP feature set. citeturn5view0
- Implementation overview (protocol/server/cache/cmd layers, CLI subcommands as thin wrappers around LSP requests). citeturn4view0

## High‑level takeaway

The gopls CLI can already deliver the key primitives we need for automated refactors: symbol discovery (`symbols`, `workspace_symbol`), cross‑reference lookup (`references`), semantic navigation (`definition`, `implementation`), and safe rename checks + application (`prepare_rename`, `rename`). It also exposes code actions (`codeaction`) for refactor/quickfix passes. The CLI is explicitly experimental and should be treated as a debugging interface, but it is still a valuable substrate for a refactoring pipeline if we design around its constraints (position encoding, workspace loading, and stability). citeturn3view0turn5view0

## How gopls works (relevant internals)

- `gopls` is a language server built around LSP requests; the `server` layer implements handlers for LSP requests, and the `cmd` package provides the CLI subcommands by issuing one‑off LSP requests and exiting. citeturn4view0
- Core logic lives in the `golang` package for Go files, with other handlers for `go.mod`, `go.work`, and templates. State is managed by `cache` (sessions, views, snapshots, and file cache). citeturn4view0

Implication: CLI commands are simply a narrow LSP client, so all refactor automation must respect LSP semantics and workspace loading behavior.

## CLI inventory snapshot (local)

Captured raw help output in:
- `sources/01-gopls-help.txt`
- `sources/02-gopls-help-*.txt`

Notable subcommands relevant to refactoring work:
- `symbols`, `workspace_symbol`
- `references`, `definition`, `implementation`
- `prepare_rename`, `rename`
- `codeaction`, `codelens`
- `stats`, `api-json`, `remote`

## Sandbox experiments

All experiments were run in `scripts/gopls-sandbox` with `GOWORK=off`.

### Symbol discovery

- `gopls symbols lib/lib.go` → returns all symbols with ranges. Output: `sources/03-gopls-symbols-lib.txt`.
- `gopls workspace_symbol Widget` → returns a workspace‑wide symbol list. Output: `sources/07-gopls-workspace-symbol-widget.txt`.

Observation: CLI outputs are stable, line‑oriented, and already suitable for parsing into a refactor index (symbol name, kind, range).

### Cross‑reference & navigation

- `gopls references lib/lib.go:5:6` → returns all reference locations to `Widget`. Output: `sources/04-gopls-references-widget.txt`.
- `gopls definition main.go:10:16` → resolves a use to the `NewWidget` definition. Output: `sources/08-gopls-definition-newwidget.txt`.
- `gopls implementation lib/lib.go:9:6` → returns types implementing `Runner`. Output: `sources/09-gopls-implementation-runner.txt`.

Observation: references + definition/implementation is sufficient to build a call‑graph‑adjacent index and support safe rename planning.

### Rename safety and diffs

- `gopls prepare_rename lib/lib.go:5:6` → validates rename target. Output: `sources/05-gopls-prepare-rename-widget.txt`.
- `gopls rename -d lib/lib.go:5:6 Gizmo` → produces a diff across all impacted files. Output: `sources/06-gopls-rename-widget-diff.txt`.

Observation: rename edits are semantically aware (receiver types, embedded fields) but do not update string literals or comments. This reinforces that higher‑level refactor tooling still needs a non‑semantic text pass for docs / strings.

### Code actions

- `gopls codeaction main.go` → returns available actions in the sandbox. Output: `sources/12-gopls-codeaction-main.txt`.
- `gopls codelens main.go` → no lenses available by default. Output: `sources/13-gopls-codelens-main.txt`.

Observation: code actions are discoverable and executable, but require filtering by kind/title and may be workspace‑dependent.

### Workspace stats

- `gopls stats -anon` in repo root to inspect workspace load and cache data. Output: `sources/10-gopls-stats-anon.json`.

Observation: stats are useful for reporting workspace size, cache state, and potential performance impact when running batch refactors.

### API schema

- `gopls api-json` → JSON description of gopls API. Output: `sources/11-gopls-api.json`.

Observation: the API schema is a useful seed for generating a typed client or validating request/response structures for automation.

## Refactor tooling implications

1. **Symbol index**: `symbols` + `workspace_symbol` provides a cheap symbol list. References can be gathered incrementally for rename candidates.
2. **Rename flow**: `prepare_rename` ➜ `rename -d` ➜ `rename -w` is an auditable flow. Use `-list` or `-diff` to stage edits before writing.
3. **Position encoding**: gopls CLI uses UTF‑8 byte columns and `file:#offset` addressing. Tooling must normalize offsets (especially for UTF‑16 offsets reported by LSP clients). citeturn3view0
4. **Non‑semantic edits**: string literals, comments, README/docs need a separate pass.
5. **Stability**: CLI is experimental; a thin wrapper around LSP should expect changes. citeturn3view0
6. **Performance**: `gopls stats` indicates the cost of workspace loading; use `-remote` or daemon modes in batch workflows to amortize startup.

## Recommendations

- Build a thin “gopls driver” that:
  - enumerates symbols, references, and definitions into a local index;
  - supports a rename pipeline using `prepare_rename` + `rename -d` + `rename -w`;
  - logs raw gopls output to ticket `sources/` for reproducibility.
- Pair gopls rename with structured text rewrites for docs and string literals.
- Keep a parser for CLI output (line‑oriented ranges) + JSON parser for `api-json`.

## Open questions

- Which refactor actions are reliably surfaced as `codeaction` in a headless CLI workflow, and which require editor integration?
- Can we drive `gopls execute` or `remote` to reduce startup costs for large codebases, and is it stable enough for CI automation?
