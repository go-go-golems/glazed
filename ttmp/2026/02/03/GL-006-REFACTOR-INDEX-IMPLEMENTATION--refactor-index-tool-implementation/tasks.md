# Tasks

## TODO

- [x] Add tasks here

- [x] Deep analysis of AST symbol + code unit snapshot pass; update analysis doc
- [x] Extend SQLite schema/migrations for symbol_defs, symbol_occurrences, code_units, code_unit_snapshots
- [x] Implement AST symbol inventory ingest (go/packages + go/ast + go/types)
- [x] Implement code unit snapshot extraction (functions/types) with hashes
- [x] Add CLI commands: ingest symbols / ingest code-units (GlazeCommand + outputs)
- [x] Golden tests for AST symbols + code unit snapshots using temp git repo
- [x] Add commit lineage schema + ingest commits/commit_files/file_blobs (commit-aware runs/worktrees)
- [x] Implement gopls references ingestion (prepare_rename + references) with raw output capture
- [x] Implement tree-sitter ingestion (Oak query builder) into ts_captures for non-Go files
- [x] Implement doc/string scan ingestion (ripgrep) into doc_hits with commit-aware runs
- [x] Add report generation (SQL-backed queries + markdown templates) and CLI report command
- [x] Add tests/fixtures for gopls, tree-sitter, doc scans (golden outputs)
- [x] Add commit-aware worktree orchestration for range ingest (diff + AST + gopls + rg + treesitter)
- [x] Add CLI commands for commit lineage, tree-sitter, doc hits, gopls refs, and range ingest
- [x] Extend smoke tests for commit lineage, gopls refs (skip if missing), and range ingest
- [x] Run manual CLI sanity on refactorio repo after tests
- [x] Link symbol_occurrences to commit_id (schema change + ingest updates)
- [x] Link code_unit_snapshots to commit_id (schema change + ingest updates)
- [x] Add FTS indexes for doc_hits.match_text and diff_lines.text
- [x] Evaluate AST caching by blob SHA (cost/benefit + scope)
- [x] Add symbol inventory listing (symbols + occurrences + target spec output)
