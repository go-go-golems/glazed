# Diff Import Scripts

## import_git_diff_to_sqlite.py

Imports the git diff between a base ref (default: `origin/main`) and `HEAD` into a sqlite database and extracts exported Go symbols for both versions.

```bash
python3 import_git_diff_to_sqlite.py \
  --repo /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed \
  --base origin/main \
  --db /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main.sqlite \
  --summary-json /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/various/git-diff-origin-main-summary.json
```

Notes:
- Exported symbols are extracted heuristically (comment-stripping + regex) and are intended for quick analysis, not compiler-grade accuracy.
- If you need higher fidelity, replace the symbol extractor with a Go AST-based parser and re-run the import.

## analysis_queries.sql

Reference SQL queries used during analysis (symbol changes, high-churn files, signature shifts).

```bash
sqlite3 /path/to/git-diff-origin-main.sqlite < analysis_queries.sql
```
