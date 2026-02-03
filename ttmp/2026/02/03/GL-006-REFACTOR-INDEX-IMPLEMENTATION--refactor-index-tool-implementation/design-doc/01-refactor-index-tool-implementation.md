---
Title: refactor index tool implementation
Ticket: GL-006-REFACTOR-INDEX-IMPLEMENTATION
Status: active
Topics:
    - refactoring
    - tooling
    - sqlite
    - go
    - gopls
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/sources/local/gopls CLI Complete Guide.md
      Note: CLI usage examples
    - Path: ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/sources/local/gopls_deep_dive_analysis.md
      Note: gopls architecture and CLI details
ExternalSources: []
Summary: Implementation design for a SQLite-backed refactor index tool, integrating git, AST, and gopls data.
LastUpdated: 2026-02-03T19:40:00-05:00
WhatFor: Guide implementation of a refactor indexing tool using gopls and static analysis.
WhenToUse: When building or extending refactor automation that needs structured symbol + diff data.
---


# refactor index tool implementation

## Executive Summary

This document specifies the implementation design for a `refactor-index` tool that ingests git diffs, Go symbol data, and gopls‑derived references into a SQLite database. The tool is intended to drive large refactors by producing a durable, queryable index of changes and dependencies. It emphasizes reproducible ingestion runs, deterministic parsing, and traceability back to raw source artifacts.

The implementation couples multiple data sources:

- `git diff` for file‑level change structure and line‑level edits.
- AST/type analysis for symbol inventory and signatures.
- gopls CLI for semantic references, rename validation, and optional code actions.
- ripgrep (or plain text scan) for doc/string occurrences.

The result is a consistent snapshot that powers reports (migration docs, rename checklists) and automation (batch renames, doc updates).

## Problem Statement

We need a reliable, automated way to answer refactor questions:

- Which APIs changed between two refs?
- Where are renamed symbols referenced?
- Which docs/config files mention old terms?
- What structural changes should be reflected in migration guides?

Ad‑hoc scripts and grep calls are brittle and lack reproducibility. The refactor tooling needs a stable, queryable database that can be regenerated for any revision range with a clear audit trail.

## Proposed Solution

Implement `refactor-index` as a Go CLI with modular ingest stages. The CLI orchestrates collection steps, normalizes outputs into SQLite, and stores raw outputs to support audit/debug. It supports both “single run” and “incremental run” modes.

### Architecture overview

```
refactor-index
├── cmd/ (CLI)
├── ingest/
│   ├── gitdiff/        # git diff parsing
│   ├── astsymbols/     # go/parser + go/types
│   ├── gopls/          # gopls CLI driver
│   ├── textscan/       # docs/strings (rg)
├── store/
│   ├── sqlite/         # schema + inserts
├── report/
│   ├── templates/      # markdown reports
└── internal/
    ├── config/
    ├── paths/
    ├── logging/
```

### Ingest stages

1. **Git diff ingestion**
   - Parse `git diff --name-status` (files changed, rename status).
   - Parse `git diff -U0` to capture hunks/line edits.
2. **Symbol inventory**
   - Use `go/packages` to load packages and `go/types` + `go/ast` for symbols.
   - Capture `kind`, `signature`, receiver, and file position.
3. **Reference graph**
   - Use gopls CLI: `prepare_rename`, `references`, `definition`, `implementation`.
   - Focus on explicit symbol sets (e.g., rename candidates) to limit runtime.
4. **Textual/doc hits**
   - Use ripgrep with a terms list to capture doc/README/config occurrences.
   - Store matches with file path, line, column, and match text.
5. **Metadata snapshot**
   - Tool version, git refs, timestamps, command arguments, OS/Go version.

### gopls integration design

Based on the gopls deep dive and CLI guide, the CLI is a thin wrapper around LSP requests, using file positions and spans (`file:line:column` or `file:#offset`) to issue requests. The tool should treat gopls as an external dependency whose outputs are captured and normalized. The integration layer will:

- Normalize positions and file paths.
- Support batch requests for `references` via a queue.
- Keep raw gopls output for debugging.
- Optionally use `-remote=auto` to reuse a gopls daemon for performance.

### Data model (implementation view)

The SQLite schema follows the GL‑005 design but adds explicit staging tables for raw outputs and a normalized view layer for queries.

```
meta_runs(id, started_at, finished_at, tool_version, git_from, git_to, root_path)
raw_outputs(id, run_id, source, path, sha256)

files(id, path, ext, exists, is_binary)
diff_files(id, run_id, file_id, status, old_path, new_path)
diff_hunks(id, diff_file_id, old_start, old_lines, new_start, new_lines)
diff_lines(id, hunk_id, kind, line_no_old, line_no_new, text)

symbols(id, run_id, file_id, pkg, name, kind, recv, signature, line, col)
symbol_refs(id, run_id, symbol_id, file_id, line, col, is_decl, source)

doc_hits(id, run_id, file_id, line, col, term, match_text)
```

### CLI commands

```
refactor-index init --db index.sqlite
refactor-index ingest diff --from origin/main --to HEAD --db index.sqlite
refactor-index ingest symbols --root ./glazed --db index.sqlite
refactor-index ingest refs --symbols rename_targets.csv --db index.sqlite
refactor-index ingest docs --terms terms.txt --db index.sqlite
refactor-index report --db index.sqlite --out report.md
```

## Design Decisions

1. **gopls CLI vs library integration**  
   Use CLI for initial implementation to reduce dependency on internal gopls APIs. The deep‑dive doc clarifies that the CLI is a thin LSP client, which is sufficient for references and rename planning. We will keep a driver abstraction so we can later switch to library‑level calls if necessary.

2. **Batch ingestion with `run_id`**  
   Each ingestion run writes to a new `run_id`, preserving history and enabling diffing between runs.

3. **Raw output preservation**  
   All external command outputs (gopls, git diff, rg) are stored in `sources/` and registered in `raw_outputs`. This enables re‑parsing and audit without re‑running.

4. **Position encoding standardization**  
   We store both line/col and `#offset` where available. gopls CLI uses UTF‑8 byte offsets; internal representation will store the raw gopls values and a normalized row for display.

5. **Explicit symbol selection for references**  
   Full graph extraction is expensive. The system will allow a curated symbol list to drive `references` requests.

## Alternatives Considered

1. **Direct LSP client + gopls server**  
   Higher fidelity but more complex to implement and maintain. CLI approach is faster to ship.

2. **AST‑only indexing**  
   Misses runtime refs and rename logic; gopls provides more semantic context.

3. **No database, only JSON**  
   JSON is easier to generate but harder to query at scale. SQLite offers flexible SQL queries.

## Implementation Plan

1. **Scaffold**
   - Create Go module, Cobra CLI, config loader.
   - Implement `init` command (schema creation).
2. **Schema + migrations**
   - Create `schema_versions` table.
   - Provide migrations in `store/sqlite/migrations`.
3. **Git diff ingestion**
   - Implement `gitdiff` parser for name‑status and zero‑context hunks.
   - Insert into `diff_*` tables.
4. **Symbol inventory**
   - Use `go/packages` + `go/types`.
   - Record `pkg`, `name`, `kind`, `signature`, `recv`, `line/col`.
5. **gopls driver**
   - Implement CLI wrapper with `prepare_rename`, `references`, `definition`.
   - Capture raw outputs and normalize to `symbol_refs`.
6. **Doc scanning**
   - Implement ripgrep driver or use `bufio` to scan text.
   - Store matches in `doc_hits`.
7. **Reports**
   - Generate Markdown reports from SQL queries.
8. **Validation + tests**
   - Golden tests for diff parsing.
   - Fixture tests for gopls output parsing.
   - `sqlite` integration tests in temp DB.

## Open Questions

- Should we support incremental updates (append new run_id) vs overwrite?
- Do we want a DAG of symbol renames over time?
- How do we handle multiple modules and `go.work`? (Dedicated root per run?)

## References

- GL‑005: sqlite refactor index tool design (overall architecture)
- Imported sources:
  - `sources/local/gopls_deep_dive_analysis.md`
  - `sources/local/gopls CLI Complete Guide.md`
