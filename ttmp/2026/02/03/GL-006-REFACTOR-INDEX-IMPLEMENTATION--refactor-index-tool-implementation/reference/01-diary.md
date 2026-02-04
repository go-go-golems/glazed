---
Title: Diary
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
    - Path: glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/analysis/01-pass-2-ast-symbols-and-code-unit-snapshots-analysis.md
      Note: Pass 2 analysis
    - Path: glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/design-doc/01-refactor-index-tool-implementation.md
      Note: GL-006 design doc
    - Path: glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/tasks.md
      Note: GL-006 task breakdown
    - Path: refactorio/cmd/refactor-index/ingest_code_units.go
      Note: CLI command for code unit snapshots
    - Path: refactorio/cmd/refactor-index/ingest_symbols.go
      Note: CLI command for symbols ingestion
    - Path: refactorio/cmd/refactor-index/root.go
      Note: Wired new ingest subcommands
    - Path: refactorio/pkg/refactorindex/ingest_code_units.go
      Note: Code unit snapshot ingestion
    - Path: refactorio/pkg/refactorindex/ingest_commits.go
      Note: Commit lineage ingestion
    - Path: refactorio/pkg/refactorindex/ingest_gopls_refs.go
      Note: gopls references ingestion
    - Path: refactorio/pkg/refactorindex/ingest_symbols.go
      Note: AST symbol ingestion
    - Path: refactorio/pkg/refactorindex/ingest_symbols_code_units_smoke_test.go
      Note: Golden tests for symbols and code units
    - Path: refactorio/pkg/refactorindex/schema.go
      Note: |-
        Pass 2 schema additions
        Commit lineage tables
        symbol_refs table
    - Path: refactorio/pkg/refactorindex/store.go
      Note: Symbol insert helpers
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-03T19:56:44-05:00
WhatFor: ""
WhenToUse: ""
---









# Diary

## Goal

Track implementation of GL-006 pass 2 (AST symbols + code unit snapshots), including analysis, tasks, and incremental code changes.

## Step 1: Deep analysis and task breakdown for pass 2

I reviewed the GL-006 design doc and produced a focused analysis for the pass 2 ingestion work (AST symbols + code unit snapshots). I also created the detailed task list that will drive incremental implementation work, matching the same task/commit/diary loop used in GL-005.

This step sets the scope and data model for the next round of code changes while keeping the work structured and auditable.

### Prompt Context

**User prompt (verbatim):** "ok, then run the test suite, do some manual CLI sanity in the refactorio repo itself, then assuming there are no issues, close the ticket and move on to GL-006, doing a deep analysis and creating the necessary tasks, then continuing in the same way with task, build, check off, commit, diary, etc..."

**Assistant interpretation:** After validating GL-005, start GL-006 by doing a deep analysis and creating detailed tasks for the AST symbol + code unit snapshot pass.

**Inferred user intent:** Ensure GL-006 starts with clear scope, tasks, and documentation before coding begins.

**Commit (code):** d7d53d0 — "docmgr: start GL-006 analysis and tasks"

### What I did
- Drafted a deep analysis document for AST symbols + code unit snapshots.
- Created detailed GL-006 tasks for schema, ingestion, CLI, and tests.
- Created the GL-006 diary document for ongoing updates.

### Why
- This keeps GL-006 aligned with the structured workflow requested by the user.

### What worked
- The analysis doc captures schema additions, hashing rules, and ingestion steps.

### What didn't work
- N/A.

### What I learned
- N/A (analysis focused on consolidating existing design details).

### What was tricky to build
- N/A (no code changes yet).

### What warrants a second pair of eyes
- Review the proposed schema additions and hashing strategy before implementation.

### What should be done in the future
- Execute tasks in order: schema extension, symbol ingestion, code unit snapshots, CLI wiring, golden tests.

### Code review instructions
- Start at the pass 2 analysis doc for data model and ingestion details.
- Review GL-006 task list for planned sequence.

### Technical details
- `symbol_defs` + `symbol_occurrences` tables for stable symbol hashing and run-specific occurrences.
- `code_units` + `code_unit_snapshots` for body/doc text + hashes.

## Step 2: Extend schema for symbols and code units

I extended the SQLite schema to include symbol definitions/occurrences and code unit snapshots, bumping the schema version and adding indexes for hash lookups. This creates the storage foundation needed before wiring AST ingestion.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the first concrete GL-006 task by extending the schema for pass 2 data.

**Inferred user intent:** Ensure the DB can persist symbol and code unit data before we build ingestion logic.

**Commit (code):** 0d30b1d — "Extend schema for symbols and code units"

### What I did
- Added `symbol_defs`, `symbol_occurrences`, `code_units`, and `code_unit_snapshots` tables.
- Added hash/run-based indexes for symbol and code-unit lookups.
- Bumped `SchemaVersion` to 2.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- The AST ingestion and snapshot work needs stable tables and indexes first.

### What worked
- Schema updates compile and the existing tests still pass.

### What didn't work
- N/A.

### What I learned
- N/A.

### What was tricky to build
- N/A (straightforward schema additions).

### What warrants a second pair of eyes
- Verify schema naming and indexes align with the pass 2 analysis doc.

### What should be done in the future
- Implement AST symbol ingestion and code unit snapshot extraction.

### Code review instructions
- Review `refactorio/pkg/refactorindex/schema.go` for new tables and indexes.

### Technical details
- New tables include hash columns (`symbol_hash`, `unit_hash`, `body_hash`) for stable identity and diffing.

## Step 3: Implement AST symbol ingestion

I implemented the pass 2 AST symbol ingestion pipeline using `go/packages` and added store helpers for symbol definitions and occurrences. This includes stable hashing for symbol definitions, path normalization, and run tracking, with schema updates to enforce unique symbol hashes.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build the AST symbol inventory ingestion logic after the schema is in place.

**Inferred user intent:** Begin real pass 2 ingestion work with deterministic symbol data in SQLite.

**Commit (code):** 9d5238b — "Implement AST symbol ingestion"

### What I did
- Added `IngestSymbols` to load packages and extract top-level symbols.
- Added store helpers for `symbol_defs` and `symbol_occurrences`.
- Enforced uniqueness on `symbol_hash` and bumped schema version to 3.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- Symbol ingestion is the core of pass 2 and a prerequisite for code unit snapshots and reports.

### What worked
- The ingestion function compiles and integrates with the existing store/run pipeline.

### What didn't work
- N/A.

### What I learned
- `go/packages` can supply enough type info for stable signatures when paired with `types.RelativeTo`.

### What was tricky to build
- Ensuring deterministic symbol hashing required consistent qualifier usage and path normalization.

### What warrants a second pair of eyes
- Review symbol coverage (currently top-level declarations + methods) and ensure it matches expectations.

### What should be done in the future
- Add code unit snapshot extraction and CLI wiring for `ingest symbols`.

### Code review instructions
- Start at `refactorio/pkg/refactorindex/ingest_symbols.go`.
- Review `refactorio/pkg/refactorindex/store.go` for symbol insert helpers.
- Check `refactorio/pkg/refactorindex/schema.go` for unique hash constraints.

### Technical details
- Symbol hashes are SHA-256 of `pkg|name|kind|recv|signature`.

## Step 4: Implement code unit snapshot ingestion

I added the code unit snapshot ingestion pipeline, capturing function/method/type spans, body text, doc text, and hashes. The implementation computes stable unit hashes, stores snapshots with start/end line/column, and tracks counts for code units and snapshots.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build the code unit snapshot pass using AST nodes and type info.

**Inferred user intent:** Capture stable, queryable code unit snapshots to support refactor diffing.

**Commit (code):** 98a6142 — "Implement code unit snapshot ingestion"

### What I did
- Added `IngestCodeUnits` to walk AST declarations and record code unit snapshots.
- Added store helpers for `code_units` and `code_unit_snapshots` inserts.
- Normalized body text for hashing and kept raw body text for snapshots.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- Code unit snapshots enable diffing and historical tracking for functions and types.

### What worked
- Snapshot extraction uses node spans to capture accurate start/end positions.

### What didn't work
- N/A.

### What I learned
- Using node spans (not symbol positions) is necessary for accurate snapshot ranges.

### What was tricky to build
- Handling multi-spec `type (...)` blocks required choosing between GenDecl and TypeSpec spans.

### What warrants a second pair of eyes
- Review span selection logic for `type` declarations with multiple specs.

### What should be done in the future
- Add CLI wiring for `ingest code-units` and golden tests.

### Code review instructions
- Start at `refactorio/pkg/refactorindex/ingest_code_units.go`.
- Review `refactorio/pkg/refactorindex/store.go` for code unit inserts.

### Technical details
- Body hash uses SHA-256 of normalized text (CRLF → LF, trim trailing whitespace).

## Step 5: Add CLI commands for symbol and code unit ingestion

I added `ingest symbols` and `ingest code-units` GlazeCommands and wired them into the `ingest` command group. Each command calls the new ingestion functions and emits structured rows with counts.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Expose the new ingestion passes via CLI commands that follow the Glazed patterns.

**Inferred user intent:** Make pass 2 ingestion usable from the CLI for manual runs and tests.

**Commit (code):** 99bd539 — "Add CLI commands for symbol and code-unit ingestion"

### What I did
- Added `cmd/refactor-index/ingest_symbols.go` and `cmd/refactor-index/ingest_code_units.go`.
- Wired both commands under the `ingest` group in `root.go`.
- Ran `go test ./... -count=1`.

### Why
- CLI wiring is required before golden tests can run through the command surface.

### What worked
- The commands compile and return structured output rows.

### What didn't work
- N/A.

### What I learned
- N/A.

### What was tricky to build
- N/A (straightforward wiring).

### What warrants a second pair of eyes
- Ensure command naming (`code-units`) aligns with expected UX.

### What should be done in the future
- Add golden tests for symbols and code unit snapshots.

### Code review instructions
- Start at `refactorio/cmd/refactor-index/ingest_symbols.go` and `refactorio/cmd/refactor-index/ingest_code_units.go`.
- Check `refactorio/cmd/refactor-index/root.go` for wiring.

### Technical details
- Output rows include counts for symbols/occurrences and code-units/snapshots.

## Step 6: Add golden smoke tests for symbols and code units

I added a golden smoke test that creates a minimal Go module in a temp directory, runs both symbol and code-unit ingestion, and asserts expected symbols, code units, and snapshot bodies in SQLite. This provides a full pass 2 regression check on the AST and snapshot pipelines.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Create golden tests that validate the new pass 2 ingestion flows end-to-end.

**Inferred user intent:** Ensure symbol and code unit ingestion are stable and verifiable.

**Commit (code):** e4fc92e — "Add golden smoke test for symbols and code units"

### What I did
- Added `ingest_symbols_code_units_smoke_test.go` with a temp Go module and package.
- Asserted symbols (`Person`, `Greet`, `Add`, `Answer`) and code units in SQLite.
- Validated snapshot bodies contain expected text.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- The smoke test guards against regressions in symbol and code unit ingestion.

### What worked
- The test passes and validates both ingestion paths in a single run.

### What didn't work
- N/A.

### What I learned
- N/A.

### What was tricky to build
- Ensuring the test module is self-contained and loads correctly via `go/packages`.

### What warrants a second pair of eyes
- Confirm that the assertions are strict enough but not overly brittle across Go versions.

### What should be done in the future
- Consider adding multi-package fixtures if we need broader coverage.

### Code review instructions
- Start at `refactorio/pkg/refactorindex/ingest_symbols_code_units_smoke_test.go`.

### Technical details
- The test checks snapshot bodies using SQL `LIKE` for `type Person` and `func Add`.

## Step 7: Add remaining index pipeline tasks from GL-005 design

I reviewed the GL-005 design doc and added the remaining pipeline tasks to GL-006: commit lineage ingestion, gopls references, tree-sitter captures, doc scan ingestion, report generation, worktree orchestration, and golden fixtures. This aligns the implementation backlog with the full indexing roadmap described in GL-005.

### Prompt Context

**User prompt (verbatim):** "Look at the glazed/ttmp/2026/02/03/GL-005-SQLITE-INDEX-TOOL--sqlite-refactor-index-tool/design-doc/01-sqlite-refactor-index-tool.md and the tasks for that index part (treesitter, etc...) and add those to the ticket"

**Assistant interpretation:** Extract the remaining ingestion/reporting tasks from GL-005 and add them as GL-006 tasks.

**Inferred user intent:** Ensure the GL-006 backlog captures the full index pipeline beyond pass 2.

**Commit (code):** N/A

### What I did
- Added GL-006 tasks for commit lineage ingestion, gopls references, tree-sitter captures, doc scan ingestion, report generation, worktree orchestration, and test fixtures.

### Why
- The GL-005 design doc defines a broader ingestion roadmap that should be tracked in GL-006.

### What worked
- Task list now reflects all remaining index pipeline passes.

### What didn't work
- N/A.

### What I learned
- N/A.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- Review task ordering/prioritization for commit-aware worktree support.

### What should be done in the future
- Implement the newly added tasks in order of dependency (commit lineage → gopls/tree-sitter/doc scans → reports → golden fixtures).

### Code review instructions
- Review `glazed/ttmp/2026/02/03/GL-006-REFACTOR-INDEX-IMPLEMENTATION--refactor-index-tool-implementation/tasks.md` for the updated backlog.

### Technical details
- New tasks target `ts_captures`, `doc_hits`, `symbol_refs`, `commits/commit_files/file_blobs`, and reporting SQL/templating.

## Step 8: Add commit lineage ingestion

I added commit lineage schema tables and an ingestion pipeline that walks commits in a range, captures commit metadata, and records file changes with blob SHAs and basic blob stats. This builds the commit-aware backbone needed for per-commit AST, gopls, tree-sitter, and doc scan runs.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement commit lineage ingestion to support commit-aware indexing.

**Inferred user intent:** Make the index capable of tracking history across commit ranges.

**Commit (code):** b4aecb9 — "Add commit lineage ingestion"

### What I did
- Added `commits`, `commit_files`, and `file_blobs` tables (schema version 4).
- Implemented `IngestCommits` to parse commit metadata and diff-tree name-status entries.
- Captured blob SHAs and basic size/line-count stats for new blobs.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- Commit lineage is required to drive commit-aware AST/gopls/tree-sitter ingestion.

### What worked
- The ingestion pipeline compiles and integrates with the existing store/run flow.

### What didn't work
- N/A.

### What I learned
- Using `git show -s --format` with a unit separator keeps commit parsing robust.

### What was tricky to build
- Handling rename/copy entries required mapping old/new paths and resolving blobs on parent commits.

### What warrants a second pair of eyes
- Review blob stats computation for large/binary files and the use of parent commit for old blobs.

### What should be done in the future
- Wire commit ingestion into the CLI and add tests/fixtures.

### Code review instructions
- Start at `refactorio/pkg/refactorindex/ingest_commits.go`.
- Review `refactorio/pkg/refactorindex/schema.go` and `refactorio/pkg/refactorindex/store.go` for new tables/helpers.

### Technical details
- Commit parsing format: `%H%x1f%an%x1f%ae%x1f%ad%x1f%cd%x1f%s%x1f%b`.

## Step 9: Implement gopls references ingestion

I added gopls reference ingestion with raw output capture and parsing into `symbol_refs`. The new pipeline runs `gopls prepare_rename` and `gopls references -declaration`, stores outputs, and normalizes reference locations to file IDs.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the gopls references ingestion pass described in GL-005.

**Inferred user intent:** Capture semantic reference data into SQLite for refactor checks.

**Commit (code):** b38dc2b — "Add gopls references ingestion"

### What I did
- Added `symbol_refs` table and indexes (schema version 5).
- Implemented `IngestGoplsReferences` with prepare_rename + references calls.
- Added parsing for multiple gopls reference output formats.
- Added store helpers to insert symbol refs and resolve symbol IDs by hash.
- Ran `go test ./pkg/refactorindex -count=1`.

### Why
- gopls references provide semantic cross-package usage data that AST alone can’t capture.

### What worked
- The ingestion pipeline compiles and captures raw outputs per run.

### What didn't work
- N/A.

### What I learned
- gopls reference output may include ranges, so parsing must handle multiple span formats.

### What was tricky to build
- Mapping references back to symbol definitions requires a stable symbol hash from the target list.

### What warrants a second pair of eyes
- Validate the output parser against real gopls reference output variants.

### What should be done in the future
- Add fixtures/tests for gopls parsing (Task 13).

### Code review instructions
- Start at `refactorio/pkg/refactorindex/ingest_gopls_refs.go`.
- Review `refactorio/pkg/refactorindex/schema.go` and `refactorio/pkg/refactorindex/store.go` for symbol_refs support.

### Technical details
- Raw outputs are stored under `sources/<run_id>/gopls/`.
