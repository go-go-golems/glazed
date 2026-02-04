# Changelog

## 2026-02-03

- Initial workspace created


## 2026-02-03

Step 1: import gopls sources and draft implementation design (commit c76f16f)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/design-doc/01-refactor-index-tool-implementation.md — Implementation design

## 2026-02-03

Step 1: deep analysis + task breakdown (commit d7d53d0)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/analysis/01-pass-2-ast-symbols-and-code-unit-snapshots-analysis.md — Pass 2 analysis


## 2026-02-03

Step 2: extend schema for symbols/code units (commit 0d30b1d)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/schema.go — Schema additions


## 2026-02-03

Step 3: implement AST symbol ingestion (commit 9d5238b)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_symbols.go — Symbol ingestion


## 2026-02-03

Step 4: implement code unit snapshot ingestion (commit 98a6142)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_code_units.go — Code unit ingestion


## 2026-02-03

Step 5: add CLI commands for symbols/code-units (commit 99bd539)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/cmd/refactor-index/ingest_symbols.go — CLI wiring


## 2026-02-03

Step 6: add golden tests for symbols/code-units (commit e4fc92e)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_symbols_code_units_smoke_test.go — Golden test


## 2026-02-03

Step 7: add remaining pipeline tasks from GL-005 design

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/tasks.md — Backlog expansion


## 2026-02-03

Step 8: add commit lineage ingestion (commit b4aecb9)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_commits.go — Commit ingestion


## 2026-02-03

Step 9: implement gopls references ingestion (commit b38dc2b)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_gopls_refs.go — gopls references


## 2026-02-03

Step 10: add tree-sitter ingestion (commit 43be56f)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_tree_sitter.go — Tree-sitter pass


## 2026-02-03

Step 11: add doc hits ingestion (commit f824ea0)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_doc_hits.go — Doc scan ingestion


## 2026-02-03

Step 12: add report generation + CLI (commit e37404b)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/report.go — Report generation


## 2026-02-03

Step 13: add fixtures for gopls/tree-sitter/doc scans (commit 67a68d4)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_tree_sitter_smoke_test.go — Fixtures


## 2026-02-03

Step 14: add commit-range worktree orchestration (commit eb15201)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_range.go — Range ingestion


## 2026-02-03

Step 15: add CLI commands for commits/gopls/tree-sitter/doc hits/range ingest (commit a4afa3b)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/cmd/refactor-index/root.go — wire new ingest commands


## 2026-02-03

Step 16: add smoke tests for commit lineage, range, and gopls refs (commit 74b2b8c)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_commits_range_smoke_test.go — test coverage for commit lineage


## 2026-02-03

Step 17: manual CLI sanity; range ingest with symbols/code-units failed due to worktree package load errors

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/ingest_range.go — worktree ingestion behavior observed


## 2026-02-03

Step 18: add tasks for commit linkage, FTS indexes, and AST caching evaluation

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/tasks.md — added follow-up tasks


## 2026-02-03

Step 19: add symbol inventory listing and CLI (commit 094ff98)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/cmd/refactor-index/list_symbols.go — symbol inventory CLI


## 2026-02-03

Step 20: link symbol occurrences to commit IDs (commit 2980f06)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/schema.go — symbol_occurrences commit_id


## 2026-02-03

Step 21: link code unit snapshots to commit IDs (commit 7bcba24)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/schema.go — code_unit_snapshots commit_id


## 2026-02-03

Step 22: add FTS indexes for diff lines and doc hits (commit 4c03aa7)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/refactorio/pkg/refactorindex/store.go — FTS setup


## 2026-02-03

Step 23: evaluate AST caching by blob SHA (analysis update)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/analysis/01-pass-2-ast-symbols-and-code-unit-snapshots-analysis.md — added caching evaluation


## 2026-02-03

Step 24: add validation playbook and upload to reMarkable

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/reference/02-validation-playbook.md — validation steps


## 2026-02-03

Ticket closed


## 2026-02-03

Step 25: close ticket (commit 6a830d4)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/index.md — ticket status complete


## 2026-02-03

Ticket closed

