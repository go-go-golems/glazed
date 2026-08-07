---
Title: Diary
Ticket: GLAZED-MIGRATE-CLI-001
Status: active
Topics:
    - glazed
    - migration
    - architecture
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: repo://Makefile
      Note: govulncheck GOWORK=off (commit 89fb3ee)
    - Path: repo://cmd/tools/glazed-migrate/cmds/migrate.go
      Note: CheckCommand and FixCommand GlazeCommands (commit 89fb3ee)
    - Path: repo://cmd/tools/glazed-migrate/doc/01-glazed-migrate-guide.md
      Note: GeneralTopic help page (commit 89fb3ee)
    - Path: repo://cmd/tools/glazed-migrate/main.go
      Note: Binary root with logging + help system (commit 89fb3ee)
    - Path: repo://pkg/analysis/glazedmigration/driver.go
      Note: Parse-only driver Scan/ApplyFixes (commit 89fb3ee)
ExternalSources: []
Summary: Implementation diary for porting glazed-migrate to the Glazed command framework.
LastUpdated: 2026-08-07T00:00:00-04:00
WhatFor: Record implementation steps, failures, and review guidance for GLAZED-MIGRATE-CLI-001.
WhenToUse: Read before resuming or reviewing work on this ticket.
---


# Diary

## Goal

Port `glazed-migrate` from a bare `singlechecker` main to a proper Glazed command-framework binary (`check`/`fix` GlazeCommands, logging, embedded help pages), keep the vettool entry point alive, and verify downstream (geppetto/pinocchio) adaptation against glazed v1.4.2.

## Step 1: Downstream adaptation probe, driver, commands, help pages, validation

I first verified the actual adaptation need by running the analyzer over geppetto and pinocchio against glazed v1.4.2 — zero diagnostics on both, so the "adapt" half of the task was already done by the earlier sections/values hard-cut migration. The real work was the framework port, which I built as a parse-only driver plus two GlazeCommands with embedded help, validated end to end, and committed after clearing three distinct hook failures.

### Prompt Context

**User prompt (verbatim):** "I bumped glazed, you will probably have to adapt according to use ./glazed/cmd/tools/glazed-migrate. Which while you're at it you should actually port to the glazed command framework and add help entries on how to use it. Create a new docmgr ticket in glazed/ttmp (docmgr --root ...) and then keep a diary as you work and commit at appropriate intervals."

**Assistant interpretation:** Run the migration analyzer against the workspace repos, port `cmd/tools/glazed-migrate` to the Glazed command framework with help entries, and do it under a docmgr ticket in glazed/ttmp with diary and incremental commits.

**Inferred user intent:** The migration tool itself should eat its own dogfood — it ships inside glazed, so it should demonstrate the current command conventions (GlazeCommand, universal output flags, embedded help) instead of being a bare singlechecker.

**Commit (code):** `89fb3ee` — "feat(glazed-migrate): port to Glazed command framework with check/fix commands and help entries (GLAZED-MIGRATE-CLI-001)"

### What I did
- Probed adaptation: `go run ../glazed/cmd/tools/glazed-migrate ./...` in geppetto and pinocchio → 0 diagnostics on both (glazed v1.4.2). Nothing to adapt.
- Created ticket GLAZED-MIGRATE-CLI-001 in glazed/ttmp (`docmgr ticket create-ticket` from the glazed root; docmgr picks up glazed/.ttmp.yaml).
- Wrote `pkg/analysis/glazedmigration/driver.go`: `Scan(paths)` (walk dirs or accept .go files; skip hidden/vendor/node_modules/testdata; parse with go/parser; run the analyzer per file with an empty TypesInfo, matching the analyzer's own degraded test mode; resolve suggested-fix text edits to byte offsets at scan time) and `ApplyFixes(diagnostics)` (per-file grouping, back-to-front edit application, overlap/range skipping, per-file applied counts). Plus `writeMigratedFile` (filepath.Clean + .go-suffix guard + #nosec G703 justification).
- Verified with `driver_test.go`: scan finds 3 diagnostics on a fixture, fixes rewrite correctly, rescan-after-fix is idempotent, missing path errors.
- Built `cmd/tools/glazed-migrate/cmds/migrate.go`: `CheckCommand` (rows: file/line/column/message/fixes_available) and `FixCommand` (rows per modified file with edits_applied + rows for report-only findings as "manual migration required"), sharing a variadic `paths` positional argument (TypeStringList, default `.`).
- Built `main.go`: cobra root + logging section + help system with `//go:embed doc` + `cli.AddCommandsToRootCommand`.
- Wrote two help pages: `doc/01-glazed-migrate-guide.md` (GeneralTopic: rules table R1–R9, manual-vs-auto split, why it works on broken code, workflow, CI usage, vettool variant) and `doc/02-glazed-migrate-examples.md` (Example: copy-paste recipes).
- Preserved the vettool entry point as `cmd/tools/glazed-migrate-analyzer/main.go` (the old singlechecker main, verbatim with updated doc comment).
- End-to-end verified: check/fix on a /tmp fixture (4 findings, 3 applied, 1 manual), `help glazed-migrate-guide` renders, `--format json` works, vettool `-help` works.
- Validation: 4 command tests (rows, fix+manual split, scan-error propagation, universal output flags only), full test/build/vet, gosec 0 issues, govulncheck 0, `git diff --check`.

### Why
- A parse-only driver (no go/packages, no type-checking) matches the analyzer's reason to exist: it must migrate code that no longer compiles. The rules guard every TypesInfo use with `obj != nil`, so empty type info degrades to the import-aware AST matching the analyzer's own unit tests exercise.
- `AddCommandsToRootCommand` builds all commands before mounting, so a schema collision can't leave a half-mutated tree (per the glazed-command-authoring skill).

### What worked
- Driver tests and command tests passed; the /tmp end-to-end demo rewrote the fixture exactly right.
- The hook failures were all real findings, not noise.

### What didn't work (three hook iterations)
1. `nonamedreturns`: named returns in `ApplyFixes` — refactored to plain returns.
2. `gosec G703` (file-write path traversal) on the in-place write — added `writeMigratedFile` with Clean + suffix guard + justified `#nosec`.
3. `govulncheck`: otel v1.43.0→v1.44.0 bump (GO-2026-5158), plus stdlib vulns that only appear in workspace mode — glazed's `govulncheck` target ran in workspace mode (go.work pins go1.26.3), so I mirrored geppetto's `GOWORK=off` in the Makefile target (glazed go.mod already declares `toolchain go1.26.5`).
4. Bonus: `ineffassign` — a leftover re-initialization of `appliedPerFile` after the nonamedreturns refactor (my own sloppy edit; one-line fix).

### What I learned
- The glazed Cobra builder hardwires `os.Stdout` for the formatter (pkg/cli/cobra.go:167), so in-process command tests must call `RunIntoGlazeProcessor` directly with hand-built `values.Values` instead of capturing `Execute()` output; `parsedWithPaths` builds the default section values exactly like the parser would.
- `go.work` toolchain selection overrides module `toolchain` directives; `GOWORK=off` is the escape hatch, and geppetto's Makefile already used it for exactly this reason.
- The lefthook lint gate (`make lintmax gosec govulncheck`) catches real policy layers: style (nonamedreturns), security (gosec), and supply chain (govulncheck).

### What was tricky to build
- Converting `analysis.TextEdit` token positions to byte offsets: positions only make sense against the scan-time FileSet, so the driver resolves offsets inside `scanFile`'s Report callback while the FileSet is alive, keeping the public `Diagnostic` shape self-contained (my first draft used a package-level registry — rewritten immediately as racy/ugly).
- The `FixCommand` row semantics: applied counts are per-file-edit, manual findings are per-diagnostic; the driver API was reshaped (per-file counts) so the command stays trivial.

### What warrants a second pair of eyes
- The `#nosec G703` justification: paths are constrained to `.go` files discovered under user-supplied roots, but a reviewer should confirm the threat model (the tool intentionally edits arbitrary scanned files).
- Empty-TypesInfo matching: false-positive risk on locally shadowed `settings` identifiers is documented in the guide; confirm that's acceptable for a human-in-the-loop migration tool.
- The Makefile `GOWORK=off` change — behavior-neutral outside workspaces but worth an explicit nod.

### What should be done in the future
- Push branch `task/glazed-migrate-cli` and open a glazed PR.
- Consider wiring `glazed-migrate check` into geppetto/pinocchio CI as a guard against reintroducing removed APIs (one-liner in the guide).
- The `--format jsonl` CI one-liner from the examples page could become a playbook.

### Code review instructions
- Start at `pkg/analysis/glazedmigration/driver.go` (Scan/ApplyFixes), then `cmd/tools/glazed-migrate/cmds/migrate.go`, `main.go`, and `doc/*.md`.
- Validate: `GOWORK=off go test ./pkg/analysis/glazedmigration/ ./cmd/tools/glazed-migrate/... -count=1 && make lintmax gosec govulncheck`.
- Demo: `go run ./cmd/tools/glazed-migrate check <dir-with-removed-APIs>`.

### Technical details
- Rules covered (analyzer.go): R1/R2 constructors, R3 wrapper unwrap, R4 "output"→"format" key, R5 slug rename, R6–R9 report-only.
- Validation evidence: 8 new tests total (4 driver + 4 command), all passing; hooks green on commit 89fb3ee.
