---
Title: sqlite refactor index tool
Ticket: GL-005-SQLITE-INDEX-TOOL
Status: active
Topics:
    - refactoring
    - tooling
    - sqlite
    - go
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Design for a SQLite-backed refactor index that ingests diffs, symbols, and references for tooling."
LastUpdated: 2026-02-03T18:17:42.434226978-05:00
WhatFor: "Provide a queryable index of code changes, symbols, and references to drive refactor tooling."
WhenToUse: "When planning or executing large-scale API renames or schema migrations."
---

# sqlite refactor index tool

## Executive Summary

We will build a small, repeatable “refactor index” tool that ingests git diffs, symbol metadata, and code references into a local SQLite database. The database becomes a stable, queryable source of truth for refactor planning, rename verification, and documentation impact analysis. The tool prioritizes determinism (same inputs → same DB), reproducibility (all source artifacts stored), and low friction (single command per snapshot).

The index is designed to layer multiple data sources: git diff (structural file changes), AST/static analysis (symbols/types), and optional gopls-driven references (semantic graph). The output supports direct SQL queries and small helper scripts that automate cleanup checklists, migration docs, and “what changed” reports.

## Problem Statement

Large refactors (symbol renames, API redesigns, and package reorgs) are hard to audit and easy to miss. The current workflow relies on ad‑hoc ripgrep and manual tracking. We need:

- A reproducible way to catalog changes between two commits/branches.
- A queryable view of symbol usage and references.
- The ability to cross‑link diffs, docs, and symbols without ad‑hoc spreadsheets.
- A persistent data store that can be used by multiple automation scripts.

## Proposed Solution

Build a CLI tool (`refactor-index`) that generates a SQLite database from a fixed set of inputs:

1. **Git diff ingestion** (baseline)
   - Ingest `git diff --name-status` + `git diff --unified=0` between two refs.
   - Normalize to structured rows: files changed, hunks, added/removed lines.
2. **Go symbol inventory** (static)
   - Use a Go AST/type pass to enumerate package‑level symbols, methods, fields, and identifiers with positions.
   - Optionally run `gopls symbols`/`workspace_symbol` for a LSP‑flavored view (especially for non‑Go files when supported).
3. **Reference graph** (semantic)
   - Use `gopls references` to enumerate references for selected symbols (rename candidates).
   - Store both declaration and reference locations.
4. **Doc/string inventory** (textual)
   - Capture text matches for legacy terms in docs/README/config files (ripgrep backed), with file/line/column + matched text.
5. **Metadata snapshots**
   - Record git refs, timestamps, tool versions, and environment info for reproducibility.

The tool writes all captured raw outputs to a `sources/` folder and persists normalized data into SQLite. The schema is intentionally denormalized to keep queries simple.

### CLI shape

```
refactor-index init --db index.sqlite
refactor-index ingest diff --db index.sqlite --from origin/main --to HEAD
refactor-index ingest symbols --db index.sqlite --root ./glazed
refactor-index ingest references --db index.sqlite --symbols symbols.csv
refactor-index ingest docs --db index.sqlite --terms terms.txt
refactor-index report --db index.sqlite --out report.md
```

### Output structure

- `index.sqlite` (main DB)
- `sources/` (raw tool outputs: git diff, gopls output, rg output)
- `reports/` (optional generated reports)

## Design Decisions

1. **SQLite as storage**  
   Chosen for simplicity, portability, and ability to query via standard SQL. It supports multi‑table joins, full‑text search (FTS), and reproducible snapshots.

2. **Single‑shot ingestion commands**  
   Each subcommand creates an immutable snapshot table for the run, avoiding in‑place mutation that hides history.

3. **Store raw outputs alongside normalized data**  
   Ensures we can debug parser mistakes and recover from schema changes without re‑running expensive analyses.

4. **Schema versioning**  
   A `schema_versions` table tracks migrations. Each ingest run stores its schema version and tool version.

5. **Minimal dependencies**  
   Prefer standard library + `modernc.org/sqlite` or `mattn/go-sqlite3` (depending on CGO constraints). Parsing relies on `go/parser`, `go/types`, and optional gopls CLI.

## Data Model (Proposed)

### Core tables

```
meta_runs(id, started_at, finished_at, tool_version, git_from, git_to, root_path)
files(id, path, ext, exists, is_binary)

diff_files(id, run_id, file_id, status, old_path, new_path)
diff_hunks(id, diff_file_id, old_start, old_lines, new_start, new_lines)
diff_lines(id, hunk_id, kind, line_no_old, line_no_new, text)

symbols(id, run_id, file_id, pkg, name, kind, recv, signature, line, col)
symbol_refs(id, run_id, symbol_id, file_id, line, col, is_decl)

doc_hits(id, run_id, file_id, line, col, term, match_text)
```

### Suggested indexes

- `symbols(run_id, name, kind)`
- `symbol_refs(symbol_id)`
- `diff_files(run_id, status)`
- `doc_hits(run_id, term)`

### Optional FTS

Use SQLite FTS5 on `doc_hits.match_text` and `diff_lines.text` for faster phrase search.

## Alternatives Considered

1. **Ad‑hoc ripgrep + spreadsheets**  
   Rejected: not reproducible, hard to merge data sources, no strong audit trail.

2. **Postgres or DuckDB**  
   Rejected for now: higher operational overhead; SQLite is portable and adequate.

3. **Full LSP client integration only**  
   Rejected: LSP references alone miss docs/configs; we need a mixed pipeline.

4. **Custom AST‑only solution**  
   Rejected: ASTs alone don’t capture diffs or doc strings. We still need git and rg.

## Implementation Plan

1. **Scaffold CLI** (`refactor-index`) with subcommands: `init`, `ingest`, `report`.
2. **Define SQLite schema** with migrations and version tracking.
3. **Diff ingestion**
   - Parse `git diff --name-status` and `git diff -U0` to populate file/hunk/line tables.
4. **Symbol inventory**
   - Use `go/packages` or `go/parser` + `go/types` to list symbols and method receivers.
5. **Reference ingestion**
   - Use gopls CLI in batch (`gopls references`) for target symbols.
6. **Doc/string hits**
   - Use ripgrep to capture text matches and insert into `doc_hits`.
7. **Reports**
   - Provide SQL‑backed reports: “changed APIs”, “dangling symbols”, “docs requiring updates”.
8. **Validation**
   - Add a `check` command that verifies row counts and required indexes.

## Open Questions

- Should we store symbol references only for selected “rename targets” or for all symbols?
- How do we handle multi‑module workspaces (go.work) consistently?
- Do we want an explicit “snapshot” table or just `run_id` references?

## References

- Existing refactor scripts and gopls experiments (see GL‑004 ticket).

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
