---
Title: Pass 2 AST symbols and code unit snapshots analysis
Ticket: GL-006-REFACTOR-INDEX-IMPLEMENTATION
Status: active
Topics:
    - refactoring
    - tooling
    - sqlite
    - go
    - gopls
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Deep analysis for AST symbol inventory and code unit snapshot ingestion (pass 2)."
LastUpdated: 2026-02-03T19:24:23-05:00
WhatFor: "Guide implementation of symbol + code unit ingestion in refactor-index."
WhenToUse: "Before implementing pass 2 ingestion (AST symbols, code units, snapshots)."
---

# Pass 2 AST symbols and code unit snapshots analysis

## Goal

Implement deterministic ingestion of AST symbols and higher-level code units (functions, methods, types) into SQLite, with stable hashes to support rename tracking and diffing across runs.

## Scope

- **AST symbol inventory** (package-level + type-level symbols) using `go/packages`, `go/ast`, `go/types`.
- **Code unit snapshots** for functions/methods/types: capture span, body text, doc text, and hashes.
- **SQLite schema extensions** to store symbol definitions, symbol occurrences, code unit definitions, and snapshots per run.

## Non-goals (for this pass)

- gopls reference graph ingestion.
- Multi-commit historical indexing (commit lineage tables).
- Full-text search (FTS) for code bodies or docs.

## Schema additions (proposed)

Add to the existing schema (keeping `run_id` as the snapshot boundary):

```
CREATE TABLE symbol_defs (
  id INTEGER PRIMARY KEY,
  pkg TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  recv TEXT,
  signature TEXT,
  symbol_hash TEXT NOT NULL
);

CREATE TABLE symbol_occurrences (
  id INTEGER PRIMARY KEY,
  run_id INTEGER NOT NULL,
  file_id INTEGER NOT NULL,
  symbol_def_id INTEGER NOT NULL,
  line INTEGER NOT NULL,
  col INTEGER NOT NULL,
  is_exported INTEGER NOT NULL,
  FOREIGN KEY(run_id) REFERENCES meta_runs(id),
  FOREIGN KEY(file_id) REFERENCES files(id),
  FOREIGN KEY(symbol_def_id) REFERENCES symbol_defs(id)
);

CREATE TABLE code_units (
  id INTEGER PRIMARY KEY,
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  pkg TEXT NOT NULL,
  recv TEXT,
  signature TEXT,
  unit_hash TEXT NOT NULL
);

CREATE TABLE code_unit_snapshots (
  id INTEGER PRIMARY KEY,
  run_id INTEGER NOT NULL,
  file_id INTEGER NOT NULL,
  code_unit_id INTEGER NOT NULL,
  start_line INTEGER NOT NULL,
  start_col INTEGER NOT NULL,
  end_line INTEGER NOT NULL,
  end_col INTEGER NOT NULL,
  body_hash TEXT NOT NULL,
  body_text TEXT NOT NULL,
  doc_text TEXT,
  FOREIGN KEY(run_id) REFERENCES meta_runs(id),
  FOREIGN KEY(file_id) REFERENCES files(id),
  FOREIGN KEY(code_unit_id) REFERENCES code_units(id)
);
```

### Hashing strategy

- `symbol_hash`: hash of `pkg|name|kind|recv|signature` (normalized, lower-variance).
- `unit_hash`: hash of `pkg|name|kind|recv|signature` (same as symbols for now).
- `body_hash`: hash of a normalized body text (trim trailing whitespace, normalize line endings).

## Ingestion approach

### Package loading

- Use `go/packages` with `NeedName|NeedTypes|NeedSyntax|NeedTypesInfo|NeedFiles`.
- Respect `GO111MODULE` and `GOWORK` via environment if needed.
- For a first pass, index only packages under the provided `--root` directory.

### Symbol inventory rules

- Include: package-level funcs, types, consts, vars; type members (methods/fields) as occurrences.
- `kind` values: `func`, `method`, `type`, `const`, `var`, `field`, `interface_method` (if needed).
- `signature` uses `types.TypeString` with a stable qualifier (package path or short names).

### Code unit snapshot rules

- Code units: `FuncDecl`, `GenDecl` (type/const/var), `TypeSpec` with struct/interface bodies.
- Capture `doc_text` from `*ast.CommentGroup` when present.
- `body_text` is extracted from file bytes using `FileSet` positions; include the full node span.

### Determinism guarantees

- Always sort packages and files before insertion.
- Normalize paths to forward slashes and remove the repo root prefix.
- Store only UTF-8 text; skip binary files.

## Output + CLI shape

Add commands under the existing `refactor-index ingest` group:

- `ingest symbols --db index.sqlite --root ./refactorio`
- `ingest code-units --db index.sqlite --root ./refactorio`

Each should output a glazed row with counts (`run_id`, `symbols`, `occurrences`, `code_units`, `snapshots`).

## Risks and mitigations

- **go/packages load cost**: limit roots; allow `--packages` override later.
- **Signature normalization**: define a stable type qualifier in one place; test determinism.
- **Snapshot size**: consider truncation or externalization later if bodies are large.

## Validation plan

- Golden tests with a temp repo containing multiple packages and types.
- Assertions on symbol hashes, occurrence counts, and snapshot body hashes.
- SQL sanity checks for referential integrity.

## Open questions

- Should code unit snapshots include import blocks or package comments?
- Do we need a separate `symbol_occurrences` table for fields vs methods?
- Is `run_id` sufficient for versioning, or should we add explicit commit hashes early?

## AST caching by blob SHA (evaluation)

**Summary:** Useful for large commit ranges; moderate complexity. It reduces repeated parsing by reusing symbol/snapshot output for identical file blobs across commits.

**Benefits**
- Avoids re-running `go/packages` / AST parsing when file content is unchanged across commits.
- Makes commit-range indexing scale better for long histories with small diffs.

**Costs / complexity**
- Requires reading file content by blob SHA (or git show) per changed file to parse in isolation.
- Needs a cache keyed by blob SHA and per-file path (for receiver/position normalization).
- Must ensure cached symbol occurrences still map to the correct file path (renames) and commit.

**Recommendation**
- **Defer** unless we see ranges >100 commits or large monorepos.
- If implemented, scope to changed Go files only and cache only the extracted symbol/code-unit payloads, not `go/packages` metadata.

**Open questions**
- Is reuse safe when file path changes (renames) but blob identical? (likely yes; update file_id + positions only)
- Do we need a persistent cache table, or is in-memory per range run enough?
