# Tasks

## TODO

- [ ] 1. Phase 1: extend analyzer with `NewGlazedSection` rule (R2) and `GlazeCommand`-proven deletion (R1)
- [ ] 2. Phase 2: implement wrapper unwrapping (R3) and `output`→`format` key rename (R4) with value-set validation
- [ ] 3. Phase 3: implement `GlazedSlug`→`StructuredOutputSlug` rename (R5)
- [ ] 4. Phase 4: implement report-only diagnostics for `SetupTableProcessor`/`SetupProcessorOutput` (R6), `SetupTableOutputFormatter` (R7), `NewOutputFormatterSettings` (R8)
- [ ] 5. Phase 5: implement report-only diagnostics for removed feature sections (R9)
- [ ] 6. Phase 6: integration tests against remarquee, devctl, dmeta, go-minitrace, parka
- [ ] 7. Add `make migrate-check` target running the analyzer without `-fix` on workspace consuming repos
- [ ] 8. Derive R4 format allowlist from `structuredOutputFormats` (not hardcoded)

## DONE

- [x] A. Confirm the tool's single-rule scope (Step 1)
- [x] B. Enumerate the full removed-API surface from commit `cdb5537` (Step 2)
- [x] C. Census call sites across all consuming repositories (Step 3)
- [x] D. Confirm `BuildCobraCommand` auto-injection escape hatch (Step 4)
- [x] E. Create ticket `GLZ-MIGRATE-ALL-APIS` and write intern-ready design doc + diary (Step 5)
