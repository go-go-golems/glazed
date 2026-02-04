---
Title: Refactoring tool analysis
Ticket: GL-008-CREATE-REFACTORING-TOOL
Status: active
Topics:
    - refactoring
    - tooling
    - go
    - gopls
    - sqlite
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/04-postmortem-gl-002-refactor-and-tooling.md
      Note: Postmortem lessons and workflow pain points
    - Path: ../../../../../../../glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/sources/local/gopls CLI Complete Guide.md
      Note: gopls CLI capabilities and output shapes
    - Path: ../../../../../../../glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/sources/local/gopls_deep_dive_analysis.md
      Note: gopls internals, CLI architecture, and integration options
    - Path: ../../../../../../../refactorio/pkg/refactorindex/schema.go
      Note: Current refactor index schema
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_range.go
      Note: Commit-range ingestion orchestration
ExternalSources: []
Summary: "Analysis of a refactoring tool suite informed by GL-002 postmortem lessons and gopls research, defining required capabilities and gaps."
LastUpdated: 2026-02-04T00:00:00Z
WhatFor: "Guide which refactoring tools to build and how to integrate gopls, AST, git, and doc checks into a repeatable pipeline."
WhenToUse: "Before designing or implementing new refactor tooling for large-scale API changes."
---

# Refactoring tool analysis

## Executive summary

GL-002 proved that large refactors can be successful with AST tooling and scripted renames, but it exposed operational friction: brittle repo-root logic, doc snippet drift, slow validation loops, and a fragile reMarkable output pipeline. Since then, GL-006 established a SQLite-backed refactor index and gopls ingestion primitives. The next step is to build a cohesive, end-to-end refactoring tool suite that is deterministic, reproducible, and safe by default. This analysis consolidates the GL-002 lessons, gopls CLI capabilities, and refactor-index infrastructure into an actionable capability map and a prioritized tool roadmap.

## Problem statement

We need a refactoring tool suite that can:

- reliably rename APIs across Go code, configs, and docs,
- quantify impact via symbol inventories and references,
- provide reproducible snapshots and audits,
- avoid brittle scripts and ad hoc steps,
- and keep outputs consistent and reviewable (including reMarkable uploads).

The existing tools are necessary but fragmented. GL-002 used a mix of AST tools and targeted scripts; GL-006 added an index and gopls ingestion; neither provides a unified workflow with preflight checks, validation gates, and a single, declarative refactor plan.

## Key inputs and lessons

### GL-002 postmortem takeaways (condensed)

- **AST rename tools are essential** for safe Go refactors, but not sufficient for docs/configs.
- **Repo-root detection** was brittle across scripts; shared helpers are mandatory.
- **Doc replacements require semantic checks**, not just text search.
- **Validation loops were expensive** (gofmt + lint + tests); ordering and automation matter.
- **reMarkable uploads need sanitization**; inline code often breaks LaTeX.

### gopls CLI research takeaways

From the gopls CLI guide and deep-dive analysis:

- gopls exposes navigation and refactoring primitives directly (`definition`, `references`, `implementation`, `prepare_rename`, `rename`, `call_hierarchy`, `codeaction`, `codelens`).
- CLI is a thin wrapper over LSP: **one can call CLI or embed gopls libraries** for tighter integration.
- gopls uses a **session + snapshot model**; command latency is dominated by workspace loading and type-checking.
- `prepare_rename` provides **rename safety gating** and precise rename ranges.
- `references` is the best primitive for **impact analysis** of a rename target.
- `codeaction` and `codelens` provide structured transformations and diagnostics that could become optional advanced steps.

### Refactor index capabilities

The refactor-index tool already supports:

- git diff ingestion (files/hunks/lines)
- symbol inventory (AST)
- code unit snapshots (AST)
- gopls references
- doc hits (ripgrep)
- tree-sitter captures
- commit lineage and commit-aware ingestion (with range mode)
- FTS indexes for diff lines and doc hits

This provides a stable data plane for refactor audits and reports.

## Requirements (functional)

1. **Inventory and scope**
   - Enumerate symbols, code units, docs, configs.
   - Provide stable identifiers and change scope.

2. **Rename planning**
   - Accept a declarative mapping (YAML/CSV).
   - Validate rename targets (gopls `prepare_rename`).

3. **Rename execution**
   - Apply semantic renames in Go via gopls or AST.
   - Apply safe string replacements in docs/configs with guardrails.

4. **Validation gates**
   - Enforce gofmt and `go test` early.
   - Perform symbol and doc lint checks (forbidden names).

5. **Audit and reporting**
   - Query index for leftover names.
   - Produce diffs and checklists.

6. **Distribution and traceability**
   - Record runs (args, versions, commit ranges).
   - Provide reMarkable-ready outputs.

## Requirements (non-functional)

- **Determinism**: same input yields same output and DB state.
- **Safety by default**: dry-run defaults, explicit `--write` or `--apply`.
- **Repo-agnostic**: robust root detection, works with go.work.
- **Scalable**: handle large repos and commit ranges.
- **Extensible**: new passes and checks can be added without rewriting core.

## Capability map (gopls + AST + diff + docs)

### gopls-derived capabilities

- **Reference graph**: `gopls references` => impact set for a symbol
- **Rename feasibility**: `gopls prepare_rename` => rename preflight check
- **Semantic rename**: `gopls rename -d/-w` => authoritative Go rename
- **Call graph**: `gopls call_hierarchy` => optional influence scoring
- **Diagnostics**: `gopls diagnostics` => post-change health

### AST-derived capabilities

- **Symbol inventory**: stable symbol hashes for matching/lookup
- **Code-unit snapshots**: full spans for diffing and report context
- **Signature normalization**: package + receiver + signature ensures stable IDs

### Git/diff-derived capabilities

- **Structural file changes**: diff lines/hunks for evidence and reporting
- **Commit lineage**: commit-aware snapshots, per-commit analyses

### Doc/text-derived capabilities

- **Doc hits (FTS)**: fast queries for legacy terms in docs/README/configs
- **Tree-sitter**: structured captures for non-Go syntax (YAML/JSON/MD)

## Gaps identified

- A **unified refactor runner** that orchestrates inventory → plan → apply → validate → report.
- **Doc snippet linting** to catch API drift in fenced code blocks.
- A **single declarative mapping format** that drives all rename and doc updates.
- Integrated **reMarkable pipeline** with sanitization as a default step.

## Prioritization of tool work

### Phase 1 (core workflows)

1. Refactor runner (orchestrator) with dry-run, apply, and validation gates.
2. Rename planner that validates targets via gopls `prepare_rename`.
3. Doc-lint + RG/FTS audit passes, integrated into runner.

### Phase 2 (advanced usage)

1. Impact scoring using call hierarchy + references.
2. Tree-sitter driven refactors for non-Go file structures.
3. Optional gopls codeaction pass for fix-its and cleanups.

## Open questions

- Do we standardize on gopls rename or AST rename for Go code? (maybe both with a toggle)
- Should the runner always require a refactor index DB, or allow direct-only workflows?
- How do we handle large multi-module workspaces with mixed Go versions?
- What is the minimal doc-lint scope to be useful without overfitting?

## Summary of decisions (analysis-level)

- **We should build a unified refactor runner** rather than more one-off scripts.
- **gopls CLI should be the default semantic engine** for references and renames.
- **AST tools remain necessary** for inventory and fallback modes.
- **Refactor index is the canonical audit store**, with FTS for text-level checks.
- **ReMarkable output requires a sanctioned sanitizer step** (from GL-002).
