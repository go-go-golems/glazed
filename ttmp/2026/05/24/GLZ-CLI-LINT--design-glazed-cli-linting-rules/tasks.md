# Tasks

## Phase 0: Ticket planning and tracking

- [x] Create docmgr ticket `GLZ-CLI-LINT`.
- [x] Add primary design document for the Glazed CLI linter.
- [x] Add investigation diary.
- [x] Add detailed phased implementation tasks to this task list.
- [x] Validate ticket with `docmgr doctor` before implementation starts.

## Phase 1: Analyzer scaffold and vettool commands

- [x] Create analyzer package `pkg/analysis/glazedclilint`.
- [x] Export `Analyzer *analysis.Analyzer` with `inspect.Analyzer` dependency.
- [x] Add `cmd/tools/glazedclilint` singlechecker wrapper.
- [x] Add `cmd/tools/glazed-lint` multichecker wrapper.
- [x] Add Makefile targets for building and running the custom vettool.

## Phase 2: Rule A — direct environment reads

- [x] Detect `os.Getenv(...)` by resolved function identity, not by import spelling.
- [x] Support aliased `os` imports in test fixtures.
- [x] Skip generated files and `_test.go` files by default.
- [x] Add analyzer tests for direct and aliased `os.Getenv`.

## Phase 3: Rule C — raw Cobra, pflag, and Go flag definitions

- [x] Detect standard-library `flag` package flag definition calls.
- [x] Detect package-level `github.com/spf13/pflag` flag definition calls.
- [x] Detect Cobra/pflag method-chain flag definitions such as `cmd.Flags().StringVar(...)`.
- [x] Add analyzer tests for raw Cobra and Go `flag` APIs.

## Phase 4: Rule B — Glazed output section on non-structured commands

- [x] Track local variables initialized from `settings.NewGlazedSection` and `settings.NewGlazedSchema`.
- [x] Detect `cmds.WithSections(...)` calls that receive a Glazed section variable or direct constructor call.
- [x] Infer the command type being constructed in common constructor patterns.
- [x] Check whether the command type implements `RunIntoGlazeProcessor`.
- [x] Report non-Glaze commands that expose Glazed output flags.
- [x] Add analyzer tests for allowed and rejected Glazed section usage.

## Phase 5: Help and contributor documentation

- [x] Add Glazed help entry for the custom linter topic under `pkg/doc/topics/`.
- [x] Include what the linter checks, how to run it, and how to fix findings.
- [x] Include troubleshooting and See Also sections.
- [x] Ensure the help entry is embedded by the existing `pkg/doc/doc.go` `go:embed *` wiring.

## Phase 6: Validation and commits

- [x] Run focused analyzer tests.
- [x] Build the singlechecker and multichecker tools.
- [x] Run `go test` on affected packages.
- [x] Run `gofmt` on new Go files.
- [x] Commit the planning/doc updates.
- [x] Commit the analyzer implementation and help entry.
- [x] Update diary and changelog with implementation details, commands, failures, and validation evidence.
- [x] Run `docmgr doctor --ticket GLZ-CLI-LINT --stale-after 30` after updates.

## Deferred rollout tasks

- [ ] Decide whether to wire `glazed-lint` into the default `make lint` target after triaging existing repository findings.
- [ ] Run the new vettool against Pinocchio and decide which findings are immediate fixes versus allowlisted framework/application exceptions.
- [ ] Add downstream Pinocchio Makefile integration once a tagged Glazed version contains the vettool.
