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
RelatedFiles:
    - Path: ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/02-symbol-inventory.go
      Note: AST symbol inventory reference
    - Path: ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/12-rename-symbols.go
      Note: Rename tooling reference
    - Path: ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/22-rename-doc-terms.py
      Note: Doc term rewrite reference
    - Path: ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/01-gopls-help.txt
      Note: gopls CLI reference output
ExternalSources: []
Summary: Design for a SQLite-backed refactor index that ingests diffs, symbols, and references for tooling.
LastUpdated: 2026-02-03T18:31:12-05:00
WhatFor: Provide a queryable index of code changes, symbols, and references to drive refactor tooling.
WhenToUse: When planning or executing large-scale API renames or schema migrations.
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

## Detailed Ingestion Passes and Pseudocode

This section expands GL-005 into concrete ingestion passes. Each pass is designed to be deterministic, auditable, and easy to re-run. The pseudocode is intentionally explicit about inputs, outputs, and error handling. The intent is to translate each pass into a standalone `ingest` subcommand with a shared run context.

### Pass 0: Run initialization and filesystem layout

Reasoning: every ingestion must be reproducible and traceable. We create a run record up front, ensure a stable `sources/` layout, and capture tool versions. This avoids "silent" provenance loss when data is re-ingested later.

Key decisions:
- Use a `run_id` row for each `ingest` invocation (even if multiple passes share a `run_id`).
- Persist the CLI args and tool versions in `meta_runs`.
- Store raw outputs under `sources/<run_id>/` and register them in `raw_outputs`.

Pseudocode:

```
function init_run(db, config):
    now = timestamp()
    run_id = insert meta_runs(
        started_at=now,
        tool_version=config.tool_version,
        git_from=config.git_from,
        git_to=config.git_to,
        root_path=config.root_path,
        args_json=config.args_json
    )
    run_dir = path.join(config.sources_dir, run_id)
    ensure_dir(run_dir)
    return run_id, run_dir
```

### Pass 1: Git diff ingestion

Reasoning: git provides ground truth for file-level changes and line-level deltas. Parsing `--name-status` captures renames and file status. Parsing `-U0` diff hunks yields precise added/removed lines without extra context. This forms the spine of the index because every other pass (AST, gopls, docs) will be joined to file paths that are known to change.

Key decisions:
- Use `git diff --name-status -z` to safely handle spaces.
- Use `git diff -U0 --no-color` to parse hunks.
- Record rename pairs (old_path, new_path).
- Normalize paths relative to repo root.

Pseudocode:

```
function ingest_git_diff(db, run_id, repo_root, from_ref, to_ref):
    status_output = run("git", "-C", repo_root, "diff", "--name-status", "-z", from_ref, to_ref)
    raw_status_path = write_raw(run_id, "git-name-status.txt", status_output)
    insert raw_outputs(run_id, source="git-name-status", path=raw_status_path)

    for record in parse_name_status_z(status_output):
        # record.status: A/M/D/R/C, record.paths: [old, new] for rename/copy
        file_id_old = ensure_file(db, record.old_path)
        file_id_new = ensure_file(db, record.new_path or record.old_path)
        insert diff_files(run_id, file_id_new, status=record.status, old_path=record.old_path, new_path=record.new_path)

    diff_output = run("git", "-C", repo_root, "diff", "-U0", "--no-color", from_ref, to_ref)
    raw_diff_path = write_raw(run_id, "git-diff-u0.patch", diff_output)
    insert raw_outputs(run_id, source="git-diff-u0", path=raw_diff_path)

    for file_patch in parse_unified_diff(diff_output):
        diff_file_id = find_diff_file(run_id, file_patch.path)
        for hunk in file_patch.hunks:
            hunk_id = insert diff_hunks(diff_file_id, hunk.old_start, hunk.old_lines, hunk.new_start, hunk.new_lines)
            old_line = hunk.old_start
            new_line = hunk.new_start
            for line in hunk.lines:
                if line.kind == " ":
                    old_line += 1; new_line += 1; continue
                if line.kind == "-":
                    insert diff_lines(hunk_id, "del", old_line, NULL, line.text)
                    old_line += 1
                if line.kind == "+":
                    insert diff_lines(hunk_id, "add", NULL, new_line, line.text)
                    new_line += 1
```

### Pass 2: AST symbol inventory

Reasoning: we need a stable, non-LSP symbol inventory that is independent of gopls runtime state. The AST pass anchors symbols to files and provides type signatures and receivers for later joins. This is also the place to capture exported vs unexported names and to categorize kinds (type, func, method, const, var, field).

Key decisions:
- Use `go/packages` with `NeedTypes`, `NeedSyntax`, `NeedTypesInfo`.
- Capture file positions as line/column (byte offset) plus file path.
- Record receiver types for methods.
- Treat test files separately (optional flag to include/exclude).

Pseudocode:

```
function ingest_ast_symbols(db, run_id, root, include_tests):
    cfg = packages.Config{Dir: root, Mode: NeedSyntax|NeedTypes|NeedTypesInfo|NeedName|NeedFiles}
    pkgs = packages.Load(cfg, "./...")
    for pkg in pkgs:
        for file in pkg.Syntax:
            file_path = pkg.Fset.File(file.Pos()).Name()
            file_id = ensure_file(db, file_path)
            for decl in file.Decls:
                for symbol in extract_symbols(pkg, decl, pkg.TypesInfo):
                    insert symbols(
                        run_id=run_id,
                        file_id=file_id,
                        pkg=pkg.PkgPath,
                        name=symbol.name,
                        kind=symbol.kind,
                        recv=symbol.receiver,
                        signature=symbol.signature,
                        line=symbol.line,
                        col=symbol.col,
                        is_exported=symbol.exported
                    )
```

Implementation notes:
- Use existing GL-002 symbol inventory tooling as a reference: `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/02-symbol-inventory.go`.
- The AST pass should mirror gopls symbol kinds where possible for future alignment.

### Pass 3: gopls reference ingestion

Reasoning: gopls provides semantic references across packages, including interface implementations and usages that are hard to infer from a pure AST walk. We use gopls CLI as a "semantic probe" for selected symbols, and store its results. The GL-004 ticket already captured gopls behavior and outputs that can be used to shape parsers and expectations.

Key decisions:
- Run gopls with `GOWORK` explicitly set to avoid accidental workspace mismatches.
- Use `prepare_rename` before `references` for validation.
- Only run references for a curated list of symbols (rename targets, public API).
- Store both declaration and reference rows with `is_decl`.

Pseudocode:

```
function ingest_gopls_refs(db, run_id, root, symbols):
    set_env("GOWORK", "off")
    for sym in symbols:
        pos = format_position(sym.file, sym.line, sym.col)
        prep = run("gopls", "prepare_rename", pos)
        if prep.failed:
            record_skip(sym, "prepare_rename failed")
            continue

        refs = run("gopls", "references", "-declaration", pos)
        raw_path = write_raw(run_id, "gopls-references-" + sym.name + ".txt", refs.stdout)
        insert raw_outputs(run_id, source="gopls-references", path=raw_path)

        for ref in parse_gopls_locations(refs.stdout):
            file_id = ensure_file(db, ref.file)
            insert symbol_refs(run_id, sym.id, file_id, ref.line, ref.col, ref.is_decl, source="gopls")
```

Implementation notes:
- Use GL-004 outputs for parser fixtures: `ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/`.
- For efficiency, allow `-remote=auto` to reuse the gopls daemon for batch runs.

### Pass 4: Doc and string scan (grep/ripgrep)

Reasoning: a refactor is incomplete if docs, README files, config templates, and string literals still contain old terms. A fast textual scan fills this gap and provides a "doc update checklist."

Key decisions:
- Use ripgrep with a term list (one term per line).
- Capture file path, line, column, and matched text.
- Exclude vendor/binary and large files by default.

Pseudocode:

```
function ingest_doc_hits(db, run_id, root, terms_file):
    for term in read_lines(terms_file):
        rg = run("rg", "--line-number", "--column", "--no-heading", term, root)
        raw_path = write_raw(run_id, "rg-" + slug(term) + ".txt", rg.stdout)
        insert raw_outputs(run_id, source="rg", path=raw_path)

        for match in parse_rg_lines(rg.stdout):
            file_id = ensure_file(db, match.file)
            insert doc_hits(run_id, file_id, match.line, match.col, term, match.text)
```

Implementation notes:
- Use GL-002 doc rename tool (`ttmp/.../scripts/22-rename-doc-terms.py`) to build a canonical terms list.
- Consider a "legacy terms" vocabulary with aliases and expected replacements.

### Pass 5: Data normalization and reporting

Reasoning: raw data becomes useful only after a normalization pass that detects changed APIs, impacted docs, and inconsistent rename outcomes. This pass materializes queries into reports (Markdown or JSON) that can be consumed by migration docs.

Key decisions:
- Keep report SQL in `reports/queries/` for reuse.
- Provide a `report` command that renders templates from SQL outputs.

Pseudocode:

```
function build_reports(db, run_id, out_dir):
    ensure_dir(out_dir)
    for report in list_reports():
        rows = db.query(report.sql, run_id)
        md = render_template(report.template, rows)
        write_file(path.join(out_dir, report.name + ".md"), md)
```

## Existing Scripts and Data to Reuse

The following prior tooling can inform or be reused in implementation:

- GL-002 symbol inventory: `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/02-symbol-inventory.go`
- GL-002 rename tooling: `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/12-rename-symbols.go`
- GL-002 doc term rewrite: `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/22-rename-doc-terms.py`
- GL-004 gopls experiments and outputs: `ttmp/2026/02/03/GL-004-GO-PLS-INVESTIGATION--gopls-investigation-for-refactoring-tooling/sources/`

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
