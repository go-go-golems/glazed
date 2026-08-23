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
      Note: |-
        Parse-only driver Scan/ApplyFixes (commit 89fb3ee)
        Review fixes—recursive path normalization, dedupe, cancellation, partial results (commit 668999b)
    - Path: repo://pkg/analysis/glazedmigration/rules_slug.go
      Note: Dot-import fallback with parser object shadow guard (commit 668999b)
    - Path: repo://pkg/doc/topics/33-glazed-source-migrations.md
      Note: Public guide migrated to check/fix syntax (commit 668999b)
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

## Step 2: Address six PR review findings

PR #620 received two P1 and four P2 inline findings. The fixes harden both usability and destructive-operation semantics: documented `./...` paths now work, overlapping roots are deduplicated, dot-imported slug references survive parse-only scanning, cancellation stops writes, and partial application is always surfaced before an error returns.

I also migrated the older public help topic to the new `check`/`fix` subcommands. This closed a documentation split where the newly embedded guides were correct but `pkg/doc/topics/33-glazed-source-migrations.md` still taught the removed singlechecker CLI.

### Prompt Context

**User prompt (verbatim):** "Address code review issues: https://github.com/go-go-golems/glazed/pull/620"

**Assistant interpretation:** Fetch every review issue on glazed PR #620, implement and test each requested correction, update ticket records, push, reply, and resolve the threads.

**Inferred user intent:** Make the new migration CLI safe, internally consistent, and ready to merge rather than merely replying to automated review.

**Commit (code):** `668999b` — "fix(glazed-migrate): address PR review findings"

### What I did
- Made `Scan` context-aware and taught path collection to normalize Go-style trailing `/...` before `os.Stat`.
- Deduplicated collected files by cleaned absolute path, eliminating duplicate diagnostics and duplicate/conflicting fix edits from overlapping roots.
- Made `ApplyFixes` context-aware; it checks cancellation before work and immediately before every file write.
- Added `ApplyResult` (`AppliedPerFile`, `Skipped`) and return it even on later failure. `FixCommand` emits partial modifications before returning the error, using `context.WithoutCancel` only for final accounting after a cancellation that followed earlier writes.
- Added the explicit post-scan cancellation check so SIGINT cannot turn already-scanned diagnostics into writes.
- Restored dot-imported R5 detection in parse-only mode using `go/parser`'s `Ident.Obj`: imported bare `GlazedSlug` has nil Obj, while same-file local declarations/usages retain Obj and are not rewritten.
- Updated `pkg/doc/topics/33-glazed-source-migrations.md` from the removed no-subcommand/`-fix` syntax to `check` and `fix`.
- Added regression tests for literal `./...`, overlapping roots/explicit files, dot imports plus local shadowing, cancellation-before-write, partial results after a later file disappears, command cancellation, and partial-result row emission.

### Why
- The guide repeatedly used `./...`; accepting that familiar Go package pattern is less surprising than rewriting every example to `.`.
- Destructive tools need stronger cancellation and failure accounting than read-only commands. A non-transactional multi-file rewrite must at least report every file changed before a later failure.
- Full type checking is inappropriate for a migration tool targeting code broken by removed APIs; parser object resolution supplies the narrow shadowing guard needed for dot imports without reintroducing that dependency.

### What worked
- `/tmp/glazed-migrate check ./... --format json` completed successfully against glazed itself (empty JSON result).
- `GOWORK=off go test ./... -count=1`, targeted vet, `git diff --check`, and all focused driver/command tests passed.
- Commit hook passed tests, lintmax, gosec (0 issues), and govulncheck (0 called vulnerabilities).

### What didn't work
- The first focused run failed because the old degraded-mode R5 unit test intentionally expected zero dot-import diagnostics:
  `--- FAIL: TestR5SlugRename (0.00s)`
  `--- FAIL: TestR5SlugRename/dot-import (0.00s)`
  `rules_test.go:234: got 1 diagnostics, want 0`
  The expected behavior changed per review; I updated the test and added a local-shadow regression.
- My first `/...` normalization draft called `filepath.Clean` before suffix detection. During self-review I caught that `filepath.Clean("./...")` becomes literal `"..."`; detection now happens first, with a direct `normalizeScanPath("./...") == "."` assertion.

### What I learned
- Go package patterns are CLI conventions, not filesystem paths; tools accepting source roots must normalize them deliberately.
- `analysis.Pass.TypesInfo` is not the only source of name resolution: the parser's same-file `Ident.Obj` is sufficient to distinguish a dot-imported bare name from local shadowing in this narrow rule.

### What was tricky to build
- Partial writes plus canceled contexts conflict with structured reporting: a canceled context may reject `AddRow`, exactly when the user most needs to know what changed. The command uses `context.WithoutCancel` only when `ApplyFixes` returned an error after at least one successful write; normal operation and pre-write cancellation retain the original context.
- Cross-file transactional writes would require temp files, metadata preservation, rename orchestration, and rollback. The review explicitly allowed surfacing partial results, so `ApplyResult` provides a smaller honest contract without pretending atomicity.

### What warrants a second pair of eyes
- Dot-import fallback protects same-file local declarations; a package-level `GlazedSlug` declared in another file is not visible to per-file parser object resolution and could still look imported. Dot imports are rare and the documented migration requires recognition, but reviewers should confirm this tradeoff.
- Confirm partial-result rows are flushed by every configured Glazed formatter when command execution subsequently returns an error.

### What should be done in the future
- Consider a dry-run patch-plan mode if the migration tool grows broader destructive rewrites.
- Consider package-level parsing if future dot-import rules need cross-file shadow resolution.

### Code review instructions
- Start at `pkg/analysis/glazedmigration/driver.go` (`Scan`, `collectGoFiles`, `ApplyResult`, `ApplyFixes`) and `cmd/tools/glazed-migrate/cmds/migrate.go` (`FixCommand`, `emitApplyResult`).
- Review R5 fallback in `rules_slug.go` and regressions in `driver_test.go`/`migrate_test.go`.
- Validate: `GOWORK=off go test ./... -count=1 && make lintmax gosec govulncheck`.
- Dogfood: `go build -o /tmp/glazed-migrate ./cmd/tools/glazed-migrate && /tmp/glazed-migrate check ./... --format json`.

### Technical details
- Review comments addressed: `3737028788`, `3737028798`, `3737028799`, `3737028804`, `3737028812`, `3737028815`.
- New driver signatures: `Scan(context.Context, []string)` and `ApplyFixes(context.Context, []Diagnostic) (ApplyResult, error)`.
