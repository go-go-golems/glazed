---
Title: Refactoring tool design
Ticket: GL-008-CREATE-REFACTORING-TOOL
Status: active
Topics:
    - refactoring
    - tooling
    - go
    - gopls
    - sqlite
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_diff.go
      Note: Diff ingestion pass
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_symbols.go
      Note: AST symbol inventory
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_gopls_refs.go
      Note: gopls reference ingestion
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_doc_hits.go
      Note: Doc hits ingestion
    - Path: ../../../../../../../refactorio/pkg/refactorindex/report.go
      Note: Report rendering system
ExternalSources: []
Summary: "Design for a cohesive refactoring tool suite (runner, planner, audits) built on refactor-index, gopls CLI, and doc linting."
LastUpdated: 2026-02-04T00:00:00Z
WhatFor: "Define the architecture, commands, data flow, and implementation plan for the new refactoring tool suite."
WhenToUse: "When implementing or iterating on the refactor tool suite after GL-006." 
---

# Refactoring tool design

## Executive summary

We will build a cohesive refactoring tool suite that orchestrates inventory, rename planning, semantic renames, doc/config updates, validation, and reporting. The suite will extend the existing refactor-index data plane and use gopls CLI as the semantic engine. A single `refactor-runner` command will provide end-to-end workflows from a declarative mapping file, ensuring safe defaults, deterministic outputs, and auditability.

## Problem statement

Current refactor workflows are fragmented across AST scripts, manual ripgrep audits, and ad hoc documentation fixes. This is error-prone and difficult to scale. We need a structured tool suite that offers a single entry point for complex refactors while remaining modular and scriptable.

## Proposed solution

### Overview

Add a new `refactorio` tool suite with the following top-level commands:

- `refactor-runner` (or `refactorio run`): orchestrates the full pipeline
- `refactor-plan`: validates rename targets and produces a plan
- `refactor-apply`: applies the plan (gopls rename + doc updates)
- `refactor-audit`: audits index data, FTS hits, and doc lints
- `refactor-report`: generates SQL-backed reports (via refactor-index)

The runner wraps these steps and records a run in the refactor index.

### Core data flow

```
            +------------------+
            | mapping.yaml     |
            | (rename plan)    |
            +--------+---------+
                     |
                     v
+---------+  +-----------------+   +------------------+
| git     |->| refactor-index  |-->| refactor-report  |
+---------+  | (diff/symbols)  |   +------------------+
     |       +-----------------+              |
     v                  |                     v
+---------+             v               +-------------+
| gopls   |--------> gopls refs         | reports/    |
+---------+                             +-------------+
     |
     v
+---------------------+
| refactor-runner     |
| (plan/apply/verify) |
+---------------------+
```

### Mapping format (input)

YAML mapping drives the refactor:

```yaml
refactor:
  id: "glazed-layer-to-section"
  scope:
    root: /path/to/repo
    include:
      - "pkg/**"
      - "cmd/**"
    exclude:
      - "**/ttmp/**"
  symbols:
    - from: "github.com/go-go-golems/glazed/pkg/cmds.Schema"
      to:   "github.com/go-go-golems/glazed/pkg/cmds.Sections"
      kind: "type"
    - from: "WithLayers"
      to:   "WithSchema"
      kind: "func"
  docs:
    terms:
      - from: "layer"
        to: "section"
  config:
    files:
      - "**/*.yaml"
      - "**/*.md"
```

### Pipeline stages (runner)

1. **Inventory**
   - Run refactor-index: diff + symbols + code units + doc hits + tree-sitter

2. **Plan**
   - For each symbol: locate via symbol inventory + `gopls prepare_rename`
   - Produce a plan (JSON) with targets, rename ranges, and risk flags

3. **Apply**
   - Execute `gopls rename` for each plan entry (dry-run or write)
   - Apply doc/config updates (safe replacements + linted code blocks)

4. **Validate**
   - gofmt + go test
   - doc-lint for forbidden symbols and invalid flag usage
   - FTS/ripgrep audit for remaining legacy terms

5. **Report**
   - Generate SQL-backed reports + markdown summary
   - Optionally bundle and upload to reMarkable

## Design decisions

1. **gopls CLI as the default semantic engine**
   - Best signal for rename safety and reference impact
   - Aligns with GL-006 ingestion flow

2. **Index-first workflow**
   - refactor-index is the canonical data store
   - Supports audits and reporting before/after changes

3. **Runner orchestration with dry-run default**
   - All mutations gated behind `--apply`
   - All steps produce structured logs and artifacts

4. **Doc updates are linted and validated**
   - Fenced code blocks are checked for legacy symbols
   - doc replacements are tracked and reversible

5. **Pluggable passes**
   - Each stage is modular (`run --steps inventory,plan,apply`)

## Architecture details

### CLI surface

```
refactorio run --config mapping.yaml --db index.sqlite --repo . --from <ref> --to <ref>
refactorio plan --config mapping.yaml --db index.sqlite --out plan.json
refactorio apply --plan plan.json --write
refactorio audit --db index.sqlite --terms terms.txt --fail-on-hit
refactorio report --db index.sqlite --run-id <id> --out reports/
```

### Data structures

```go
// plan.json

type RenameTarget struct {
  SymbolHash string
  FilePath   string
  Line       int
  Col        int
  OldName    string
  NewName    string
  Kind       string
  Package    string
  Risk       []string
}

type RefactorPlan struct {
  ID         string
  CreatedAt  string
  RepoRoot   string
  Targets    []RenameTarget
  DocTerms   []DocTermChange
}
```

### Pseudocode: plan generation

```
func BuildPlan(config, db) Plan:
  inventory = ListSymbolInventory(db, filters=config.symbols)
  plan = new Plan

  for entry in inventory:
    if !match(entry, config.symbols): continue
    if gopls.prepare_rename(entry.file, entry.line, entry.col) fails:
       plan.addRisk(entry, "prepare_rename_failed")
       continue
    plan.addTarget(entry, newName)

  return plan
```

### Pseudocode: apply

```
func ApplyPlan(plan, write):
  for target in plan.targets:
    if write:
       gopls rename -w <pos> <new>
    else:
       gopls rename -d <pos> <new>

  applyDocChanges(plan.docTerms, write)
  runGofmt()
  runGoTest()
  runDocLint()
```

### Doc lint rules (initial)

- Reject fenced Go code blocks containing any old symbol names.
- Validate `glazed` example flags against current command schemas.
- Require at least one `refactor-audit` pass after doc updates.

## Alternatives considered

1. **AST-only refactor tool**
   - Rejected: lacks semantic references and rename safety.

2. **gopls-only tool**
   - Rejected: no doc/config updates and no inventory persistence.

3. **Use LSP library directly**
   - Deferred: gopls CLI is simpler and sufficient; library integration can come later for performance.

## Implementation plan

1. **Define mapping schema** (YAML + validation) and a plan JSON format.
2. **Implement plan builder** using symbol inventory + gopls prepare_rename.
3. **Implement apply step** using gopls rename + doc updates + validations.
4. **Integrate refactor-index** as a required dependency for audits and reports.
5. **Add doc-lint** and `rg`/FTS checks for legacy terms.
6. **Add reMarkable export** as an optional final step (sanitized bundle).

## Open questions

- Should rename plans include automatic dependency updates (module import paths)?
- How do we safely handle cross-repo renames or monorepo boundaries?
- Should the runner auto-start a long-lived gopls daemon for speed?

## Appendix: Command-to-gopls mapping

- `refactor-plan` uses `gopls prepare_rename` and `gopls references`
- `refactor-apply` uses `gopls rename -d/-w`
- `refactor-audit` uses `rg` + FTS queries in SQLite
