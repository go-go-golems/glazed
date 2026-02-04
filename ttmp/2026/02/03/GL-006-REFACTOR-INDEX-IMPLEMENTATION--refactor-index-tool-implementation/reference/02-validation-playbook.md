---
Title: Validation Playbook
Ticket: GL-006-REFACTOR-INDEX-IMPLEMENTATION
Status: active
Topics:
    - refactoring
    - tooling
    - sqlite
    - go
    - gopls
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../refactorio/cmd/refactor-index/main.go
      Note: CLI entry point
    - Path: ../../../../../../../refactorio/pkg/refactorindex/schema.go
      Note: Schema + version
    - Path: ../../../../../../../refactorio/pkg/refactorindex/store.go
      Note: Schema init, inserts, FTS
    - Path: ../../../../../../../refactorio/pkg/refactorindex/query.go
      Note: Symbol inventory query
    - Path: ../../../../../../../refactorio/pkg/refactorindex/ingest_range.go
      Note: Range ingest orchestration
---

# Validation Playbook

## Goal
Verify that the refactor index tool ingests diffs, symbols, code units, docs, tree-sitter captures, gopls refs, and commit-linked snapshots into SQLite with FTS indexes, and that the CLI outputs expected structured rows.

## Context
This playbook runs the tool against (1) the refactorio repo itself and (2) a synthetic git repo that mirrors the smoke tests. Use temp directories to avoid dirtying the repo. Some steps are optional depending on whether `gopls` and `rg` are installed.

## Quick Reference

### Prereqs
- Go toolchain available
- Optional: `gopls` in PATH (for gopls references)
- Optional: `rg` in PATH (for doc hits)

Check quickly:
```
which gopls || true
which rg || true
```

### Track A — Run on refactorio repo
All commands below assume:
```
ROOT=/home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio
TMP=$(mktemp -d)
DB=$TMP/index.sqlite
SOURCES=$TMP/sources
TERMS=$TMP/terms.txt
QUERY=$TMP/queries.yaml
```
Create small inputs:
```
cd "$ROOT"

echo "refactor" > "$TERMS"
cat > "$QUERY" <<'QEOF'
language: go
queries:
  funcs: |
    (function_declaration name: (identifier) @name)
QEOF
```

1) Init DB
```
go run ./cmd/refactor-index init --db "$DB"
```
Expect: a row with `schema_version` printed.

2) Commit lineage
```
go run ./cmd/refactor-index ingest commits --db "$DB" --repo . --from HEAD~1 --to HEAD
```
Expect: `commits >= 1` and non-zero `files`/`blobs`.

3) Symbols + code units
```
go run ./cmd/refactor-index ingest symbols --db "$DB" --root .

go run ./cmd/refactor-index ingest code-units --db "$DB" --root .
```
Expect: non-zero `symbols`, `occurrences`, `code_units`, `snapshots`.

4) Symbol inventory + gopls target spec
```
go run ./cmd/refactor-index list symbols --db "$DB" --limit 5
```
Expect: rows with `target_spec` like `symbol_hash|file|line|col`.

5) Doc hits (requires rg)
```
go run ./cmd/refactor-index ingest doc-hits --db "$DB" --root . --terms "$TERMS" --sources-dir "$SOURCES"
```
Expect: non-zero `hits`.

6) Tree-sitter captures
```
go run ./cmd/refactor-index ingest tree-sitter --db "$DB" --root . --language go --queries "$QUERY" --file-glob "$ROOT/cmd/refactor-index/*.go"
```
Expect: non-zero `captures`.

7) gopls refs (requires gopls)
```
TARGET=$(DB="$DB" python3 - <<'PY'
import sqlite3
import os
import sys

conn = sqlite3.connect(os.environ["DB"])  # DB env
cur = conn.cursor()
cur.execute("""
SELECT d.symbol_hash, f.path, o.line, o.col
FROM symbol_occurrences o
JOIN symbol_defs d ON d.id = o.symbol_def_id
JOIN files f ON f.id = o.file_id
WHERE d.kind = 'func'
LIMIT 1
""")
row = cur.fetchone()
if row:
    print(f"{row[0]}|{row[1]}|{row[2]}|{row[3]}")
PY
)

go run ./cmd/refactor-index ingest gopls-refs --db "$DB" --repo . --target "$TARGET" --sources-dir "$SOURCES"
```
Expect: non-zero `references`.

8) Range ingest (diff-only is safest)
```
go run ./cmd/refactor-index ingest range --db "$DB" --repo . --from HEAD~1 --to HEAD --include-diff
```
Expect: a row with `commit_lineage_run_id` and `diff_run_id` populated.

Note: range ingest with `--include-symbols`/`--include-code-units` can fail if worktrees don’t see `go.work`. Use diff-only unless you’ve addressed the go.work/worktree issue.

### Track B — Synthetic repo (golden expectations)
This mirrors the smoke tests to verify exact counts.

```
TMP=$(mktemp -d)
REPO=$TMP/repo
DB=$TMP/index.sqlite
SOURCES=$TMP/sources

mkdir -p "$REPO"
cd "$REPO"

git init

git config user.email test@example.com

git config user.name "Refactor Index"

cat > fileA.txt <<'EOF_A'
alpha
beta
EOF_A

cat > fileB.txt <<'EOF_B'
one
EOF_B

cat > fileC.txt <<'EOF_C'
gone
EOF_C


git add -A
git commit -m "initial"
FROM=$(git rev-parse HEAD)

cat > fileA.txt <<'EOF_A2'
alpha
beta2
EOF_A2

git mv fileB.txt fileB_renamed.txt
rm fileC.txt
cat > fileD.txt <<'EOF_D'
new
EOF_D

git add -A
git commit -m "update"
TO=$(git rev-parse HEAD)

cd /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio

go run ./cmd/refactor-index init --db "$DB"

go run ./cmd/refactor-index ingest diff --db "$DB" --repo "$REPO" --from "$FROM" --to "$TO" --sources-dir "$SOURCES"
```
Expected from diff ingest:
- `files = 4` (A, B_renamed, C, D)
- `hunks > 0`, `lines > 0`

Commit lineage on the synthetic repo:
```
go run ./cmd/refactor-index ingest commits --db "$DB" --repo "$REPO" --from "$FROM" --to "$TO"
```
Expected:
- `commits = 1`
- `files = 5` (rename shows as add+delete in diff-tree)
- `blobs >= 3`

## Usage Examples

### Inspect DB counts
```
sqlite3 "$DB" "SELECT COUNT(*) FROM diff_files;"
sqlite3 "$DB" "SELECT COUNT(*) FROM symbol_defs;"
sqlite3 "$DB" "SELECT COUNT(*) FROM code_unit_snapshots;"
```

### Verify commit linkage (range ingest)
```
sqlite3 "$DB" "SELECT COUNT(*) FROM symbol_occurrences WHERE commit_id IS NOT NULL;"
sqlite3 "$DB" "SELECT COUNT(*) FROM code_unit_snapshots WHERE commit_id IS NOT NULL;"
```

### Verify FTS indexes
```
sqlite3 "$DB" "SELECT COUNT(*) FROM diff_lines_fts;"
sqlite3 "$DB" "SELECT COUNT(*) FROM doc_hits_fts;"

# Sample FTS query
sqlite3 "$DB" "SELECT rowid, match_text FROM doc_hits_fts WHERE doc_hits_fts MATCH 'refactor' LIMIT 5;"
```

### Generate gopls target spec from list symbols
```
go run ./cmd/refactor-index list symbols --db "$DB" --exported-only --kind func --limit 1
```
Copy the `target_spec` field and use it with `ingest gopls-refs`.

## Notes / Gotchas
- Range ingest with symbols/code-units may fail in worktrees due to missing `go.work` context. Diff-only range ingest is expected to work.
- For deterministic results, use temp DBs and repos.
- `sources/` is written under the given `--sources-dir`; keep it outside the repo to avoid dirtying the worktree.
