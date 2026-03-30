---
Title: gopls refactoring playbook
Ticket: GL-004-GO-PLS-INVESTIGATION
Status: active
Topics:
    - gopls
    - refactoring
    - tooling
    - go
DocType: playbook
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/02-gopls-help-codeaction.txt
      Note: Code action reference
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/02-gopls-help-rename.txt
      Note: Rename command reference
ExternalSources: []
Summary: Repeatable CLI workflow for gopls-driven refactors.
LastUpdated: 2026-02-03T19:05:00-05:00
WhatFor: Operational guide for gopls CLI refactor workflows.
WhenToUse: When running gopls-assisted refactors or experiments.
---


# gopls refactoring playbook

## Purpose

Provide a repeatable, CLI‑only workflow for using gopls to explore and apply Go refactors (symbol discovery, references, rename, and code actions) as inputs to a larger refactoring toolchain.

Note: the gopls CLI is experimental and intended primarily for debugging, so treat it as a best‑effort interface and validate outputs carefully. citeturn3view0

## Environment assumptions

- `gopls` installed and on `PATH`
- Go toolchain installed (matching or newer than project Go version)
- Workspace builds locally (`go list` succeeds)
- For isolated experiments, use a scratch module and `GOWORK=off`

## Commands

### 1) Baseline inventory

```bash
gopls version

gopls help
```

Optional: capture help output for later parsing.

```bash
gopls help > /abs/path/to/sources/01-gopls-help.txt
```

### 2) Workspace statistics

```bash
gopls stats -anon > /abs/path/to/sources/10-gopls-stats-anon.json
```

Use this to estimate workspace size and cache load cost.

### 3) Symbol discovery

```bash
# File‑level symbols
gopls symbols path/to/file.go

# Workspace‑wide search
gopls workspace_symbol MySymbol
```

### 4) References & navigation

```bash
# 1‑based line:column, columns are UTF‑8 bytes (or use #offset)
gopls references path/to/file.go:LINE:COL

gopls definition path/to/file.go:LINE:COL

gopls implementation path/to/file.go:LINE:COL
```

### 5) Rename workflow

```bash
# Validate rename target

gopls prepare_rename path/to/file.go:LINE:COL

# Preview diff

gopls rename -d path/to/file.go:LINE:COL NewName

# Apply changes (with backups if desired)

gopls rename -w -preserve path/to/file.go:LINE:COL NewName
```

Notes:
- Use `-list` to show affected files without diff output.
- gopls rename updates semantic references only; comments and strings need a separate pass.

### 6) Code actions

```bash
# List actions for a file or range

gopls codeaction path/to/file.go

# Execute a specific kind or title match

gopls codeaction -kind=refactor.rewrite -exec -diff path/to/file.go
```

### 7) API schema for integration

```bash
gopls api-json > /abs/path/to/sources/11-gopls-api.json
```

Use this schema as a seed for typed client generation or JSON validation.

### 8) Daemon/remote mode (optional)

```bash
# Use existing daemon if available

gopls -remote=auto references path/to/file.go:LINE:COL

# Inspect remote sessions

gopls remote sessions
```

## Exit criteria

- CLI commands above succeed on target modules without errors.
- Symbol listings and references are produced and are parseable.
- Rename preview (`-d`) yields the expected semantic edits.
- Code actions (if used) are discoverable and can be executed with `-exec`.
- Results captured in `sources/` for reproducibility.

## Failure modes & mitigations

- **No workspace loaded / zero packages**: ensure working directory is within a module, or use `GOWORK=off` for a scratch module.
- **Rename fails**: confirm position is correct (`file:line:col` or `file:#offset`) and run `prepare_rename` first. citeturn3view0
- **Missing code actions**: actions are context‑dependent; try a different file or include `-kind` filters.
