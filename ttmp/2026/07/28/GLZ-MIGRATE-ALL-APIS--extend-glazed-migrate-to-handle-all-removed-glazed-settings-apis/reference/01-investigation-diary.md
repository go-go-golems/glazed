---
Title: Investigation Diary
Ticket: GLZ-MIGRATE-ALL-APIS
Status: in-progress
Topics:
    - glazed
    - migration
    - diary
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: abs:///home/manuel/code/wesen/go-go-golems/devctl/cmd/devctl/cmds/lifecycle.go
      Note: Slug lookup + SetupTableProcessor/SetupProcessorOutput two-call pattern (R6)
    - Path: abs:///home/manuel/code/wesen/go-go-golems/dmeta/pkg/dmeta/cmds/component_inspect.go
      Note: Redundant NewGlazedSchema on a GlazeCommand; auto-injection makes it deletable (R1)
    - Path: ws://remarquee/cmd/remarquee/cmds/cloud/ls.go
      Note: Representative NewGlazedSection + WithOutputSectionOptions site (output=json default)
ExternalSources: []
Summary: Chronological record of the investigation that produced the GLZ-MIGRATE-ALL-APIS design doc.
LastUpdated: 2026-07-28T00:00:00-04:00
WhatFor: Understand how the migration-surface evidence was gathered and verified.
WhenToUse: Read before resuming or auditing this ticket.
---




# Investigation Diary

## Goal

Record how the full removed-API surface was discovered and verified, so the
design doc's claims are auditable and the work can be continued safely.

## Step 1: Confirm the tool's actual scope

### Prompt Context

**User prompt (verbatim):** "Run `./glazed/cmd/tools/glazed-migrate/` to migrate the remarquee codebase to the new Glazed API."

**Assistant interpretation:** Execute the migration tool against `remarquee` and inspect the resulting changes.

**Inferred user intent:** Migrate remarquee off the removed Glazed settings APIs automatically.

### What I did

- Read `glazed/cmd/tools/glazed-migrate/main.go`. It calls
  `singlechecker.Main(glazedmigration.Analyzer)`: exactly one analyzer, no
  multichecker.
- Read `glazed/pkg/analysis/glazedmigration/analyzer.go`. It defines two
  constants — `oldConstructor = "NewGlazedSchema"` and
  `newConstructor = "NewStructuredOutputSection"` — and one rule that rewrites
  the former to the latter.
- Built the tool locally and ran it against `remarquee` with both `GOWORK=off`
  and the workspace `go.work`. Both runs produced only Go type-checker errors
  (`undefined: settings.NewGlazedSection`,
  `undefined: settings.WithOutputSectionOptions`) and applied zero fixes.

### Why

The user expected the tool to migrate remarquee. It did not. I needed to
determine whether the tool was broken or simply out of scope.

### What worked

- The `singlechecker` driver and `RunDespiteErrors: true` flag meant the tool
  ran despite the removed symbols, surfacing the underlying type errors. This
  confirmed the tool was functioning; it simply had no rule for the symbols
  remarquee uses.

### What didn't work

- `go run ../glazed/cmd/tools/glazed-migrate` from inside `remarquee` failed
  with "directory ... outside main module or its selected dependencies". Fix:
  build the binary in the `glazed` module (`go build -o /tmp/glazed-migrate
  ./cmd/tools/glazed-migrate`) and invoke the binary from `remarquee`.

### What I learned

- The analyzer is deliberately conservative: every fix is gated on a
  type-checker proof that the qualifier resolves to the Glazed settings package,
  with a fallback AST heuristic for shadow detection. This safety-first design
  is the right model to extend.

### What was tricky to build

- N/A for this step (investigation only).

### What warrants a second pair of eyes

- None yet; this step only confirmed scope.

### What should be done in the future

- Document the `go build` + binary invocation pattern in the tool's `--help`,
  since `go run` from a sibling module does not work without a workspace.

### Code review instructions

- Start at `glazed/pkg/analysis/glazedmigration/analyzer.go:14-15` for the
  hardcoded constants.
- Validate by running the binary against `remarquee` and confirming zero fixes.

### Technical details

- Build: `cd glazed && go build -o /tmp/glazed-migrate ./cmd/tools/glazed-migrate`
- Run: `cd remarquee && /tmp/glazed-migrate -fix -diff ./...`

## Step 2: Enumerate the full removed-API surface

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Inspected the cleanup commit `cdb5537` ("Cleanup flags and middlewares") with
  `git show cdb5537 --stat -- pkg/settings/`. It removed ten files:
  `glazed_section.go`, `settings_fields-filters.go`, `settings_jq.go`,
  `settings_output.go`, `settings_rename.go`, `settings_replace.go`,
  `settings_select.go`, `settings_skip_limit.go`, `settings_sort.go`,
  `settings_template.go` (+ test).
- Extracted the removed public symbols from `git show
  cdb5537~1:pkg/settings/glazed_section.go` and each settings file. The
  aggregate `glazed_section.go` alone exposed ~30 public symbols: the
  `GlazedSection` type, `GlazeSectionOption`, both constructors, eight
  `With*SectionOptions` wrappers, five `Setup*` helpers, and `GlazedSlug`.
- Confirmed what survives: `ls pkg/settings/*.go` shows only
  `structured_output.go` and `logcopter.go`.

### Why

The design doc must map every removed symbol, not just the constructor the
tool already handles. Without the full list, migration rules would be
incomplete.

### What worked

- `git show cdb5537~1:<file>` reliably reconstructed the pre-removal source for
  symbol enumeration.
- The per-feature settings files each declared a `*Section` type, a
  `*Settings` type, a `New*Section` constructor, and a `New*Settings` /
  `New*SettingsFromValues` factory — a uniform pattern that made enumeration
  systematic.

### What didn't work

- Initially `git show cdb5537:pkg/settings/glazed_section.go` returned empty
  because the file was deleted in that commit. Using `cdb5537~1` (the parent)
  recovered it.

### What I learned

- The cleanup removed not just constructors but entire feature sections
  (select, rename, replace, template, jq, sort, skip-limit, fields-filters) and
  their runtime settings types. This means some call sites cannot be migrated
  mechanically — they rely on deleted behavior and must be redesigned.

### What was tricky to build

- N/A.

### What warrants a second pair of eyes

- The R9 rule (report-only for removed feature sections) must correctly
  distinguish a removed feature section from a legitimate application-defined
  section of the same name. The analyzer resolves by package path, so this is
  safe, but worth a review.

### What should be done in the future

- If any consuming repo still uses the removed middlewares, track a redesign
  ticket per repo.

### Code review instructions

- Reproduce: `cd glazed && git show cdb5537~1:pkg/settings/glazed_section.go | grep -E '^func |^type |^const '`

## Step 3: Census call sites across consuming repositories

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Scanned every repository under `~/code/wesen/go-go-golems` with `grep -rEn
  --include='*.go'`, excluding `*/ttmp/`, `glazed/pkg/settings/`,
  `glazed/pkg/analysis/`, `glazed/cmd/tools/glazed-migrate`, and `*_test.go`.
- Counted each removed symbol. Headline numbers: `NewGlazedSchema` 135,
  `NewGlazedSection` 89, `GlazedSlug` 35, `WithOutputSectionOptions` 25,
  `SetupTableProcessor` 5, `SetupProcessorOutput` 3, `NewOutputFormatterSettings` 2,
  `SetupTableOutputFormatter` 1, `WithFieldsFiltersSectionOptions` 1.
- Inspected representative sites for each pattern: `dmeta` (redundant section),
  `remarquee` (section + output default), `devctl` (slug + Setup helpers),
  `parka` (SetupTableOutputFormatter), `sqleton` (GetField slug+output),
  `clay` (removed feature section).

### Why

The design doc's rules must be prioritized by real call-site volume. The
census also distinguishes mechanical migrations from manual redesigns.

### What worked

- The exclusion filter (`grep -v` for ttmp/glazed-internals/tests) produced a
  clean consumer-only count.
- Reading full context (±6 lines) around each `WithOutputSectionOptions` site
  revealed that all 25 use the identical `{"output": "<format>"}` default-map
  pattern, which makes the R4 key-rename safe and deterministic in practice.

### What didn't work

- A first grep using an alternation with `\b` undercounted `GlazedSlug`
  (returned 0). Switching to a literal `GlazedSlug\b` pattern found 35 sites.
  Lesson: avoid complex alternations in the census; enumerate one symbol per
  grep pass.

### What I learned

- The `output` → `format` flag rename is the highest-risk mechanical migration:
  leaving `"output"` as a default key causes a runtime crash in
  `InitializeDefaultsFromMap` (verified in
  `glazed/pkg/cmds/fields/definitions.go:393`), not a compile error.
- The `GlazedSlug` → `StructuredOutputSlug` rename is cross-cutting: it appears
  in slug lookups (`parsedValues.Get`), schema sets (`Schema.Set`), help-section
  lists, and `GetParameter`/`GetField` calls.

### What was tricky to build

- Confirming that default-map keys are matched by field name (and that an
  unknown key errors) required reading `InitializeDefaultsFromFields` →
  `InitializeDefaultsFromMap`. The error path is easy to miss because the
  rename otherwise looks purely cosmetic.

### What warrants a second pair of eyes

- The R4 value-set allowlist (`table|json|jsonl|csv|tsv|yaml`) must be derived
  from `structuredOutputFormats`, not hardcoded, or it will drift if formats
  are added later.

### What should be done in the future

- Add a regression test that fails if `structuredOutputFormats` changes without
  the analyzer's allowlist being updated.

### Code review instructions

- Reproduce the census with the loop in design-doc §5.
- Verify the key-matching semantics at
  `glazed/pkg/cmds/fields/definitions.go:393-404`.

## Step 4: Confirm the auto-injection escape hatch

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Read `glazed/pkg/cli/cobra.go:225-240`. `BuildCobraCommand` auto-injects a
  `NewStructuredOutputSection()` into the schema of any `cmds.GlazeCommand` that
  does not already declare one.
- Verified a representative consumer (`dmeta.ListComponentsCommand`) declares
  `var _ cmds.GlazeCommand = (*ListComponentsCommand)(nil)` and implements
  `RunIntoGlazeProcessor`, confirming its explicit `NewGlazedSchema()` is
  redundant and deletable.

### Why

This single fact transforms the migration strategy: 224 sites that look like
"rename the constructor" are better handled as "delete the redundant block".
The design doc's R1/R2 decision record depends on it.

### What worked

- The auto-injection code is a small, self-contained block with a clear comment
  ("If the command implements GlazeCommand, ensure structured output settings
  are present").

### What didn't work

- N/A.

### What I learned

- The safe deletion precondition is "the enclosing command implements
  `GlazeCommand`", which the analyzer can prove from the type checker's method
  set. This is feasible and is the basis of R1.

### What was tricky to build

- N/A (read-only step).

### What warrants a second pair of eyes

- The R1 `GlazeCommand` proof must handle commands returned through interface
  types or helper functions, where the concrete type is not locally visible.
  The design doc specifies a fallback to rename in those cases.

### What should be done in the future

- A follow-up cleanup pass can delete the renamed-but-redundant survivors that
  R1 could not prove.

### Code review instructions

- Read `glazed/pkg/cli/cobra.go:225-240`.
- Confirm `dmeta/pkg/dmeta/cmds/component_inspect.go:33,142` for the
  `GlazeCommand` assertion and method.

## Step 5: Author the design doc and create the ticket

### Prompt Context

**User prompt (verbatim):** "Ok, make a new docmgr ticket in glazed/ to address all the mgirations the tool needs to do, not just the one we tagged. Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create ticket `GLZ-MIGRATE-ALL-APIS` in the `glazed` repo, write an intern-ready design doc covering all migrations, store it via docmgr, and upload to reMarkable.

**Inferred user intent:** A self-contained, onboarding-grade document that lets a new engineer extend the migration tool without further hand-holding.

### What I did

- Created the ticket with `docmgr ticket create-ticket --ticket GLZ-MIGRATE-ALL-APIS`.
- Added a design-doc and a reference (diary) document.
- Wrote the design doc (`design-doc/01-...md`) with: executive summary, problem
  statement, background on the analysis framework and current analyzer,
  current-state removal surface, call-site census, gap analysis, nine migration
  rules (R1–R9) with pseudocode, four decision records, analyzer architecture,
  end-to-end migration flows, phased implementation plan, test strategy, risks,
  and a references section.
- Wrote this diary.

### Why

The user asked for an intern-ready guide. The structure follows the
`ticket-research-docmgr-remarkable` skill's design-doc order and the
`writing-style.md` guidance (evidence-anchored, decision records inline,
pseudocode for flows).

### What worked

- All evidence was gathered before writing, so the doc could be authored in one
  pass with file-anchored claims.
- The docmgr ticket workspace scaffolding (index/tasks/changelog/design-doc/
  reference) was created automatically.

### What didn't work

- N/A.

### What I learned

- The migration problem decomposes cleanly into three tiers (redundant-section
  deletion, semantic rewrites, runtime-helper reporting), which maps directly
  onto analyzer rules of increasing automation risk.

### What was tricky to build

- Articulating the `output` → `format` key rename's runtime-failure mode
  precisely: it is not a compile error but a runtime "unknown field" error in
  `InitializeDefaultsFromMap`, which is harder to diagnose without the doc.

### What warrants a second pair of eyes

- The R1 `GlazeCommand` proof and the R3 unwrapping edit are the two rules most
  likely to misfire. Their pseudocode and preconditions should be reviewed
  before implementation.
- The 11 report-only sites (R6–R8) depend on a human applying the sketched
  collapse; the diagnostic messages must be unambiguous.

### What should be done in the future

- Implement Phases 1–6 of the design doc.
- After implementation, re-run the census to confirm the fixable count drops to
  zero and only the report-only sites remain.

### Code review instructions

- Read the design doc in full; cross-check every file reference against the
  current `glazed` tree.
- Validate the census by re-running the loop in design-doc §5.

### Technical details

- Ticket path: `glazed/ttmp/2026/07/28/GLZ-MIGRATE-ALL-APIS--extend-glazed-migrate-to-handle-all-removed-glazed-settings-apis/`
- Design doc: `design-doc/01-intern-guide-extending-glazed-migrate-to-all-removed-glazed-settings-apis.md`
- Diary: `reference/01-investigation-diary.md`
