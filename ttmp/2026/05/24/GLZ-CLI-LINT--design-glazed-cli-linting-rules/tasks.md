# Tasks

## Phase 0: Ticket planning and tracking

- [x] Create docmgr ticket `GLZ-CLI-LINT`.
- [x] Add primary design document for the Glazed CLI linter.
- [x] Add investigation diary.
- [x] Add detailed phased implementation tasks to this task list.
- [x] Validate ticket with `docmgr doctor` before implementation starts.

## Phase 1: Analyzer scaffold and vettool commands

- [ ] Create analyzer package `pkg/analysis/glazedclilint`.
- [ ] Export `Analyzer *analysis.Analyzer` with `inspect.Analyzer` dependency.
- [ ] Add `cmd/tools/glazedclilint` singlechecker wrapper.
- [ ] Add `cmd/tools/glazed-lint` multichecker wrapper.
- [ ] Add Makefile targets for building and running the custom vettool.

## Phase 2: Rule A — direct environment reads

- [ ] Detect `os.Getenv(...)` by resolved function identity, not by import spelling.
- [ ] Support aliased `os` imports in test fixtures.
- [ ] Skip generated files and `_test.go` files by default.
- [ ] Add analyzer tests for direct and aliased `os.Getenv`.

## Phase 3: Rule C — raw Cobra, pflag, and Go flag definitions

- [ ] Detect standard-library `flag` package flag definition calls.
- [ ] Detect package-level `github.com/spf13/pflag` flag definition calls.
- [ ] Detect Cobra/pflag method-chain flag definitions such as `cmd.Flags().StringVar(...)`.
- [ ] Add analyzer tests for raw Cobra and Go `flag` APIs.

## Phase 4: Rule B — Glazed output section on non-structured commands

- [ ] Track local variables initialized from `settings.NewGlazedSection` and `settings.NewGlazedSchema`.
- [ ] Detect `cmds.WithSections(...)` calls that receive a Glazed section variable or direct constructor call.
- [ ] Infer the command type being constructed in common constructor patterns.
- [ ] Check whether the command type implements `RunIntoGlazeProcessor`.
- [ ] Report non-Glaze commands that expose Glazed output flags.
- [ ] Add analyzer tests for allowed and rejected Glazed section usage.

## Phase 5: Help and contributor documentation

- [ ] Add Glazed help entry for the custom linter topic under `pkg/doc/topics/`.
- [ ] Include what the linter checks, how to run it, and how to fix findings.
- [ ] Include troubleshooting and See Also sections.
- [ ] Ensure the help entry is embedded by the existing `pkg/doc/doc.go` `go:embed *` wiring.

## Phase 6: Validation and commits

- [ ] Run focused analyzer tests.
- [ ] Build the singlechecker and multichecker tools.
- [ ] Run `go test` on affected packages.
- [ ] Run `gofmt` on new Go files.
- [ ] Commit the planning/doc updates.
- [ ] Commit the analyzer implementation and help entry.
- [ ] Update diary and changelog with implementation details, commands, failures, and validation evidence.
- [ ] Run `docmgr doctor --ticket GLZ-CLI-LINT --stale-after 30` after updates.

## Deferred rollout tasks

- [ ] Decide whether to wire `glazed-lint` into the default `make lint` target after triaging existing repository findings.
- [ ] Run the new vettool against Pinocchio and decide which findings are immediate fixes versus allowlisted framework/application exceptions.
- [ ] Add downstream Pinocchio Makefile integration once a tagged Glazed version contains the vettool.
