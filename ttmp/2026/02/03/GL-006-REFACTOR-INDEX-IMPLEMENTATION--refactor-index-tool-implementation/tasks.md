# Tasks

## TODO

- [ ] Add tasks here

- [x] Deep analysis of AST symbol + code unit snapshot pass; update analysis doc
- [x] Extend SQLite schema/migrations for symbol_defs, symbol_occurrences, code_units, code_unit_snapshots
- [x] Implement AST symbol inventory ingest (go/packages + go/ast + go/types)
- [x] Implement code unit snapshot extraction (functions/types) with hashes
- [x] Add CLI commands: ingest symbols / ingest code-units (GlazeCommand + outputs)
- [x] Golden tests for AST symbols + code unit snapshots using temp git repo
- [x] Add commit lineage schema + ingest commits/commit_files/file_blobs (commit-aware runs/worktrees)
- [x] Implement gopls references ingestion (prepare_rename + references) with raw output capture
- [ ] Implement tree-sitter ingestion (Oak query builder) into ts_captures for non-Go files
- [ ] Implement doc/string scan ingestion (ripgrep) into doc_hits with commit-aware runs
- [ ] Add report generation (SQL-backed queries + markdown templates) and CLI report command
- [ ] Add tests/fixtures for gopls, tree-sitter, doc scans (golden outputs)
- [ ] Add commit-aware worktree orchestration for range ingest (diff + AST + gopls + rg + treesitter)
