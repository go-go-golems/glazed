---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/12-rename-symbols.yaml
      Note: Symbol rename mapping referenced in appendix
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/26-build-postmortem-appendix.py
      Note: Generate appendices from git and rename map
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Postmortem: GL-002 refactor and tooling

## Executive summary

The GL-002 work completed a no-backward-compat refactor to remove legacy "layer" and "parameter" vocabulary from the public surface, replacing it with schema/section/field/values terminology. The refactor was technically successful (build/test/lint succeeded; public API aligned), but the execution exposed operational friction: tool scripts with brittle repo-root assumptions, a dependency on full pre-commit pipelines for validation, and a noisy document conversion path that required custom sanitization for reMarkable uploads.

The most effective aspect was the AST and scripted rename tooling, which allowed consistent changes across hundreds of files. The least elegant portions were the ad hoc fixes for doc snippets and the PDF conversion failures due to LaTeX parsing issues.

This postmortem aims to: (1) document the specific workflow used, (2) identify weaknesses and root causes, (3) derive a repeatable playbook for future large-scale API refactors, and (4) propose targeted tooling improvements.

## Objectives (original intent)

1) Remove all legacy "layer" and "parameter" vocabulary from exported API and docs.
2) Remove backwards compatibility aliases ("no return" refactor).
3) Update code, examples, docs, and supporting artifacts to new naming.
4) Maintain compile correctness across examples and packages.
5) Preserve traceability: detailed diary, task list, scripts stored in ticket.
6) Provide reMarkable-ready artifacts and close the ticket with audit trails.

## Scope

### In-scope
- Go code in `glazed/` (public API and internal references).
- Examples, docs, prompto, pinocchio, changelogs, and supporting YAML fixtures.
- Rename tooling (Go AST and helper scripts).
- Ticket artifacts (analysis, design doc, diary, scripts).
- Build/test validation (go test, lint, security scans in pre-commit).

### Out-of-scope
- Non-glazed repositories under the workspace root.
- ttmp historical docs (explicitly excluded from rename sweeps).

## Timeline (high-level)

- Early phase: exhaustive inventory, symbol discovery, and design plan.
- Mid phase: scripted renames (AST + targeted scripts) across code and docs.
- Late phase: cleanup, manual corrections, build fixes, and documentation polish.
- Finalization: tests/lint, diary + changelog updates, ticket closure, reMarkable upload.

## Tooling inventory and usage

### Core refactor tooling

1) Go AST rename tool
   - Location: `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/12-rename-symbols.go`
   - Mapping: `.../scripts/12-rename-symbols.yaml`
   - Purpose: rename Go identifiers safely across code.
   - Strengths: syntax-aware, compile-safe changes, repeatable.
   - Weaknesses: brittle repo-root selection; requires explicit mapping updates.

2) Targeted Python/SH scripts
   - Examples:
     - `13-rename-schema-tests.py`
     - `14-rename-sources-test-yaml.py`
     - `19-rename-fields-language.py`
     - `21-rename-field-types-example.py`
     - `22-rename-doc-terms.py`
     - `23-rename-initialize-struct-docs.py`
     - `24-update-docs-addfields-credentials.py`
   - Purpose: non-Go files and specific naming adjustments.
   - Strengths: quick corrections for fixtures and doc snippets.
   - Weaknesses: text-based replacements can miss semantic changes and can over/under-shoot if phrasing varies.

3) Auditing / verification
   - `rg -n -i "layer|parameter" glazed -g '!**/ttmp/**'`
   - Purpose: confirm vocabulary removal and remaining occurrences.
   - Strength: fast, reliable detection of residual naming.
   - Weakness: cannot detect semantic issues (e.g., example flags no longer valid).

### Validation tooling

1) `go test ./...`
   - Caught remaining alias usage in `cmd/examples/sources-example`.

2) Pre-commit hook pipeline (lefthook)
   - Runs `golangci-lint`, `gosec`, `govulncheck`.
   - Strength: comprehensive safety net.
   - Weakness: expensive; re-runs often due to gofmt or minor issues.

### Documentation / distribution tooling

1) reMarkable upload via `remarquee upload bundle`
   - Initial attempts failed due to pandoc LaTeX errors with inline `\n` and smart quotes.
   - Remediation: add sanitization script (`scripts/25-sanitize-remarkable-bundle.py`) and upload sanitized copies from `/tmp/remarkable-gl-002`.

## What worked well

1) AST tool as the primary rename engine.
   - High confidence for Go symbol changes without breaking syntax.
   - Enabled large-scale rename coverage quickly and repeatably.

2) Systematic inventory and audit.
   - Early symbol inventories gave a realistic scope.
   - `rg` scans provided continuous feedback and a clear stop condition.

3) Tight alignment of code and docs.
   - Doc updates were integrated into the same flow, preventing drift.

4) Build/test gates enforced correctness.
   - `go test` surfaced missing alias removal quickly.
   - Lint/security scans ensured no regressions were introduced.

## What was problematic or inelegant

1) Repo-root bugs in scripts.
   - Several scripts miscomputed repo root using parent offsets, causing file-not-found errors and retries.
   - This problem repeated across more than one script, indicating missing shared helpers.

2) Doc replacements still required manual fixes.
   - The general replacement scripts did not update example flags or semantic changes (`credentials-param`, `AddFlags`), so follow-up edits were needed.

3) Expensive validation loops.
   - Pre-commit automatically ran full lint/security scans, which is comprehensive but slow.
   - Gofmt errors in renamed files triggered re-runs.

4) reMarkable upload path required a workaround.
   - Pandoc LaTeX errors due to inline `\n` and smart quotes made the default bundle flow fail.
   - A custom sanitizer was required, which was not planned.

5) Late detection of alias usage.
   - `NewCommandDefinition` remained in `sources-example` until `go test` ran late in the process.

## Root cause analysis

### Problem: Script path errors
- Root cause: each script had its own repo-root logic, often using a hardcoded `parents[N]` offset.
- Evidence: initial run of the doc update script failed due to looking under `ttmp/.../pkg/doc/...`.
- Mitigation: corrected parent offsets ad hoc.
- Recommendation: introduce a shared helper or consistent path search strategy (walk up to find `go.mod` or `.git`).

### Problem: doc tooling did not fully update examples
- Root cause: text replacement is not schema-aware; it cannot infer example semantics or CLI flag conventions.
- Evidence: `credentials-param` references remained, and `AddFlags` doc snippets were still present.
- Mitigation: targeted script `24-update-docs-addfields-credentials.py` and manual updates.
- Recommendation: add a doc-aware linter that checks fenced Go code for forbidden symbols and mismatched APIs.

### Problem: validation loops due to formatting
- Root cause: automated renames created files not gofmt'd (esp. in tests and scripts).
- Evidence: golangci-lint reported gofmt errors in renamed files and scripts.
- Mitigation: run gofmt on the reported files.
- Recommendation: always run `gofmt -w` on all touched Go files immediately after tooling passes.

### Problem: reMarkable PDF generation failed
- Root cause: pandoc LaTeX fails on inline `\n` and smart quotes in markdown (not escaped).
- Evidence: error `Undefined control sequence ... fmt.Printf(“Parameters: \%v\n`.
- Mitigation: sanitize input markdown (escape `\n`, `\t`, convert smart quotes to ASCII).
- Recommendation: include a default sanitizer step for large bundles or add a preflight check that detects unescaped sequences.

## Tooling lessons learned

1) AST rename tools are necessary but not sufficient.
   - They solve the code renames reliably, but docs/tests/examples still need bespoke tooling.

2) Scripts should be modular and reusable.
   - Each script should import a shared `repo_root()` helper or a minimal utility module to avoid repeated bugs.

3) Refactor runs benefit from a single orchestrator.
   - A single "refactor runner" that runs AST rename, gofmt, tests, and audits would remove many manual loops.

4) Distribution pipelines need robust preflight.
   - The reMarkable upload path failed on a well-known LaTeX pitfall; a standardized sanitizer or preflight check should be part of the upload pipeline.

## Technical debt created (and mitigations)

1) Temporary sanitization artifacts
   - Sanitized copies exist only under `/tmp` and were not committed; the script was committed for reproducibility.

2) Tooling sprawl in ticket scripts
   - Many one-off scripts created; they are stored in ticket `scripts/` as required.
   - Improvement: deduplicate and consolidate into a small toolbox with parameters.

## Recommendations for next time (actionable)

### Process
- Establish a formal refactor playbook with the following steps:
  1) Inventory (rg + symbol extraction)
  2) AST rename pass
  3) gofmt on all touched Go files
  4) go test ./...
  5) Doc-lint for forbidden symbols
  6) Final rg audit

### Tooling
- Create a shared utility module for scripts:
  - `repo_root()` by walking up to `.git` or `go.mod`
  - Standard logging, dry-run, and report modes
- Introduce a doc-lint tool:
  - Scan fenced code blocks for forbidden symbols
  - Validate CLI flags and known API names in docs
- Standardize PDF sanitization:
  - Make sanitization a first-class step in reMarkable uploads
  - Add a preflight check for smart quotes and `\n` sequences

### Validation
- Move gofmt earlier in the pipeline to avoid lint failures late in commit.
- Use `go test ./...` after any API-alias removal to catch leftover call sites immediately.

## Refactor tooling evaluation (pros/cons)

### AST rename tool
- Pros: precise, reliable, safe for Go code, scalable to many files.
- Cons: needs careful mapping, not usable for docs or YAML, path assumptions must be correct.

### Targeted doc scripts
- Pros: cheap and effective for small, well-defined replacements.
- Cons: brittle, do not understand semantics, can leave stale examples.

### Ripgrep audits
- Pros: fast, clear, provides a hard stop condition.
- Cons: regex-based; can miss semantic mismatches or API usage that no longer compiles.

## Known constraints

- No backwards compatibility allowed; removal of aliases was mandatory.
- docmgr workflow required scripts to live under `ttmp/.../scripts`.
- reMarkable PDF toolchain is LaTeX-based; inline code needs careful escaping.

## Suggested future work (optional)

1) Build a reusable "refactor runner" script:
   - Accepts a YAML mapping and directory targets.
   - Runs AST renames + gofmt + go test + rg audit.

2) Build a doc example validator:
   - Parses fenced code blocks in `pkg/doc`.
   - Warns if forbidden identifiers are present.

3) Harden reMarkable workflows:
   - Provide a standard sanitized bundle generator.
   - Add a `--sanitize` flag to `remarquee upload bundle` (if possible).

## Concrete examples (code, locations, commits)

### Alias removal surfaced by build
- **Symptom:** `go test ./...` failed with undefined alias:\n
  `cmd/examples/sources-example/main.go:62:15: undefined: cmds.NewCommandDefinition`
- **Fix:** replaced `cmds.NewCommandDefinition` with `cmds.NewCommandDescription` in `cmd/examples/sources-example/main.go`.
- **Commit:** `6844cbf` (Refactor: finish section/field cleanup).

### Doc snippet cleanup (AddFlags → AddFields)
- **Locations updated:**\n
  - `pkg/doc/topics/sections-guide.md` (multiple code snippets calling `AddFlags`)\n
  - `pkg/doc/tutorials/migrating-to-facade-packages.md` (migration list)\n
- **Script used:** `ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/24-update-docs-addfields-credentials.py`
- **Commit:** `6844cbf`.

### Field type example rename
- **Symptom:** docs still referenced `credentials-param` (flags + struct tags) after field-type rename.\n
- **Fix:** `credentials-param` → `credentials-field` in `pkg/doc/topics/16-adding-field-types.md`.
- **Script used:** `.../scripts/24-update-docs-addfields-credentials.py`.
- **Commit:** `6844cbf`.

### Gofmt failures from large-scale renames
- **Symptom (golangci-lint):**\n
  `File is not properly formatted (gofmt)` in:\n
  - `pkg/cmds/fields/definitions_from_defaults_test.go`\n
  - `pkg/cmds/fields/definitions_test.go`\n
  - `pkg/cmds/fields/gather-fields_test.go`\n
  - `pkg/cmds/schema/section-impl.go`\n
  - `ttmp/.../scripts/12-rename-symbols.go`
- **Fix:** explicit `gofmt -w` on the files above.\n
- **Commit:** `6844cbf` (post-gofmt, pre-commit succeeded).

### reMarkable upload pipeline failure and mitigation
- **Symptom:** `remarquee upload bundle ...` failed:\n
  `Error producing PDF. ! Undefined control sequence. l.48779   fmt.Printf(“Parameters: \\%v\\n` (pandoc LaTeX error)\n
- **Root cause:** unescaped `\\n` sequences and smart quotes inside inline code in large markdown bundles.
- **Fix:** added sanitizer script `ttmp/.../scripts/25-sanitize-remarkable-bundle.py` to escape `\\n`/`\\t` and normalize quotes; upload sanitized copies from `/tmp/remarkable-gl-002`.\n
- **Commit:** `8a9c26e`.

### Example rename: parameter-types → field-types
- **Change:** renamed example command and assets from `cmd/examples/parameter-types` → `cmd/examples/field-types`, updated README and flags to field terminology.\n
- **Commit:** `6844cbf`.

### Ticket artifacts and traceability
- **Diary:** `ttmp/.../reference/01-diary.md` (Step 13 and Step 14 record tool usage, failed commands, and fixes).\n
- **Changelog entries:** `ttmp/.../changelog.md` (Steps 13–14).\n
- **Ticket closure:** `docmgr ticket close --ticket GL-002-FURTHER-CLEANUP` and commit `215ef84`.

## Final assessment

The refactor successfully achieved the no-compat, single-vocabulary API goal with a high degree of code correctness and documentation alignment. The remaining friction was primarily operational (script fragility, doc semantics, and PDF tooling). The strongest improvement area is in building a unified refactor playbook and shared tooling utilities so future renames are faster, safer, and less error-prone.

## Appendices

### Appendix A: Commit list (this branch vs upstream)

- 7874112 Refactor: rename Lua section/field conversion
- 50db9f3 Diary: record Lua rename step
- 6844cbf Refactor: finish section/field cleanup
- 46aa4e7 Diary: record step 13
- 215ef84 Ticket: close GL-002-FURTHER-CLEANUP
- 8a9c26e Docs: add remarkable bundle sanitizer
- 88d5a39 Diary: record step 14
- 37b787d Docs: expand GL-002 postmortem examples
- c474ac5 Diary: record step 15

### Appendix B: File renames (git diff --name-status --find-renames)

- Format: R<score> <old> <new>
- R100	cmd/examples/parameter-types/config.yaml	cmd/examples/field-types/config.yaml
- R100	cmd/examples/parameter-types/sample-lines.txt	cmd/examples/field-types/sample-lines.txt
- R100	cmd/examples/parameter-types/sample-list.json	cmd/examples/field-types/sample-list.json
- R067	cmd/examples/parameter-types/sample-text.txt	cmd/examples/field-types/sample-text.txt
- R100	cmd/examples/parameter-types/sample.json	cmd/examples/field-types/sample.json
- R100	cmd/examples/parameter-types/sample.yaml	cmd/examples/field-types/sample.yaml
- R100	cmd/examples/parameter-types/simple-config.yaml	cmd/examples/field-types/simple-config.yaml
- R100	misc/json-parameters-from-json.json	misc/json-fields-from-json.json
- R073	pkg/cmds/fields/parameters.go	pkg/cmds/fields/definitions.go
- R076	pkg/cmds/fields/parameters_from_defaults_test.go	pkg/cmds/fields/definitions_from_defaults_test.go
- R078	pkg/cmds/fields/parameters_test.go	pkg/cmds/fields/definitions_test.go
- R091	pkg/cmds/fields/parameter-type.go	pkg/cmds/fields/field-type.go
- R093	pkg/cmds/fields/parsed-parameter.go	pkg/cmds/fields/field-value.go
- R076	pkg/cmds/fields/gather-parameters.go	pkg/cmds/fields/gather-fields.go
- R064	pkg/cmds/fields/gather-parameters_test.go	pkg/cmds/fields/gather-fields_test.go
- R084	pkg/cmds/fields/test-data/parameters_test.yaml	pkg/cmds/fields/test-data/definitions_test.yaml
- R084	pkg/cmds/fields/test-data/parameters_validity_test.yaml	pkg/cmds/fields/test-data/definitions_validity_test.yaml
- R081	pkg/cmds/logging/layer.go	pkg/cmds/logging/section.go
- R082	pkg/cmds/schema/layer.go	pkg/cmds/schema/schema.go
- R079	pkg/cmds/schema/layer-impl.go	pkg/cmds/schema/section-impl.go
- R073	pkg/cmds/schema/layer-impl_test.go	pkg/cmds/schema/section-impl_test.go
- R074	pkg/cmds/sources/load-parameters-from-json.go	pkg/cmds/sources/load-fields-from-config.go
- R064	pkg/cmds/sources/tests/wrap-with-restricted-layers.yaml	pkg/cmds/sources/tests/wrap-with-restricted-sections.yaml
- R097	pkg/cmds/values/parsed-layer.go	pkg/cmds/values/section-values.go
- R057	pkg/cmds/values/parsed-layer_test.go	pkg/cmds/values/section-values_test.go
- R078	pkg/doc/topics/08-file-parameter-type.md	pkg/doc/topics/08-file-field-type.md
- R060	pkg/doc/topics/16-adding-parameter-types.md	pkg/doc/topics/16-adding-field-types.md
- R084	pkg/doc/topics/logging-layer.md	pkg/doc/topics/logging-section.md
- R058	pkg/doc/topics/layers-guide.md	pkg/doc/topics/sections-guide.md
- R073	pkg/doc/tutorials/custom-layer.md	pkg/doc/tutorials/custom-section.md

### Appendix C: Symbol rename map (AST tool YAML)

- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayersHandler -> WhitelistSectionsHandler
- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayerParametersHandler -> WhitelistSectionFieldsHandler
- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayers -> WhitelistSections
- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayersFirst -> WhitelistSectionsFirst
- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayerParameters -> WhitelistSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/sources: WhitelistLayerParametersFirst -> WhitelistSectionFieldsFirst
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayersHandler -> BlacklistSectionsHandler
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayerParametersHandler -> BlacklistSectionFieldsHandler
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayers -> BlacklistSections
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayersFirst -> BlacklistSectionsFirst
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayerParameters -> BlacklistSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/sources: BlacklistLayerParametersFirst -> BlacklistSectionFieldsFirst
- github.com/go-go-golems/glazed/pkg/cmds/sources: WrapWithLayerModifyingHandler -> WrapWithSectionModifyingHandler
- github.com/go-go-golems/glazed/pkg/cmds/sources: WrapWithWhitelistedLayers -> WrapWithWhitelistedSections
- github.com/go-go-golems/glazed/pkg/cmds/sources: WrapWithWhitelistedParameterLayers -> WrapWithWhitelistedSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/sources: WrapWithBlacklistedLayers -> WrapWithBlacklistedSections
- github.com/go-go-golems/glazed/pkg/cmds/sources: WrapWithBlacklistedParameterLayers -> WrapWithBlacklistedSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/sources: LoadParametersFromResolvedFilesForCobra -> LoadFieldsFromResolvedFilesForCobra
- github.com/go-go-golems/glazed/pkg/cmds/sources: readConfigFileToLayerMap -> readConfigFileToSectionMap
- github.com/go-go-golems/glazed/pkg/cmds/sources: layers_ -> schema_
- github.com/go-go-golems/glazed/pkg/cmds/sources: parsedLayers -> parsedValues
- github.com/go-go-golems/glazed/pkg/cmds/sources: layer -> section
- github.com/go-go-golems/glazed/pkg/cmds/sources: parsedLayer -> sectionValues
- github.com/go-go-golems/glazed/pkg/cmds/sources: layerPrefix -> sectionPrefix
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestParameterLayer -> TestSection
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestParsedParameter -> TestParsedField
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestExpectedLayer -> TestExpectedSection
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestWhitelistLayers -> TestWhitelistSections
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestWhitelistLayersFirst -> TestWhitelistSectionsFirst
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestWhitelistLayerParameters -> TestWhitelistSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestWhitelistLayerParametersFirst -> TestWhitelistSectionFieldsFirst
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestBlacklistLayers -> TestBlacklistSections
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestBlacklistLayersFirst -> TestBlacklistSectionsFirst
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestBlacklistLayerParameters -> TestBlacklistSectionFields
- github.com/go-go-golems/glazed/pkg/cmds/helpers: TestBlacklistLayerParametersFirst -> TestBlacklistSectionFieldsFirst
- github.com/go-go-golems/glazed/pkg/cmds/helpers: NewTestParameterLayer -> NewTestSection
- github.com/go-go-golems/glazed/pkg/cmds/helpers: NewTestParameterLayers -> NewTestSchema
- github.com/go-go-golems/glazed/pkg/cmds/schema: AppendLayers -> AppendSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: PrependLayers -> PrependSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: ParseLayerFromCobraCommand -> ParseSectionFromCobraCommand
- github.com/go-go-golems/glazed/pkg/cmds/schema: InitializeDefaultsFromParameters -> InitializeDefaultsFromFields
- github.com/go-go-golems/glazed/pkg/cmds/schema: InitializeStructFromParameterDefaults -> InitializeStructFromFieldDefaults
- github.com/go-go-golems/glazed/pkg/cmds/schema: createParameterLayer -> createSection
- github.com/go-go-golems/glazed/pkg/cmds/schema: createSimpleParameterLayer -> createSimpleSection
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestNewParameterLayers -> TestNewSchema
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersSubset -> TestSchemaSubset
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersForEach -> TestSchemaForEach
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersForEachE -> TestSchemaForEachE
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersAppendLayers -> TestSchemaAppendSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersPrependLayers -> TestSchemaPrependSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersMerge -> TestSchemaMerge
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersAsList -> TestSchemaAsList
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersClone -> TestSchemaClone
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersGetAllDefinitions -> TestSchemaGetAllDefinitions
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersWithLayers -> TestSchemaWithSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersWithDuplicateSlugs -> TestSchemaWithDuplicateSlugs
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersSubsetWithNonExistentLayers -> TestSchemaSubsetWithMissingSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersMergeWithOverlappingLayers -> TestSchemaMergeWithOverlappingSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersWithLargeNumberOfLayers -> TestSchemaWithLargeNumberOfSections
- github.com/go-go-golems/glazed/pkg/cmds/schema: TestParameterLayersWithUnicodeLayerNames -> TestSchemaWithUnicodeSectionNames
- github.com/go-go-golems/glazed/pkg/cmds/schema: ChildLayers -> ChildSections
- github.com/go-go-golems/glazed/pkg/cmds/sources_test: wrapWithRestrictedLayersTestsYAML -> wrapWithRestrictedSectionsTestsYAML
- github.com/go-go-golems/glazed/pkg/cmds/sources_test: wrapWithRestrictedLayersTest -> wrapWithRestrictedSectionsTest
- github.com/go-go-golems/glazed/pkg/cmds/sources_test: TestWrapWithRestrictedLayers -> TestWrapWithRestrictedSections
- github.com/go-go-golems/glazed/pkg/settings: GlazedParameterLayers -> GlazedSection
- github.com/go-go-golems/glazed/pkg/settings: GlazeParameterLayerOption -> GlazeSectionOption
- github.com/go-go-golems/glazed/pkg/settings: NewGlazedParameterLayers -> NewGlazedSection
- github.com/go-go-golems/glazed/pkg/settings: FieldsFiltersParameterLayer -> FieldsFiltersSection
- github.com/go-go-golems/glazed/pkg/settings: OutputParameterLayer -> OutputSection
- github.com/go-go-golems/glazed/pkg/settings: RenameParameterLayer -> RenameSection
- github.com/go-go-golems/glazed/pkg/settings: ReplaceParameterLayer -> ReplaceSection
- github.com/go-go-golems/glazed/pkg/settings: SelectParameterLayer -> SelectSection
- github.com/go-go-golems/glazed/pkg/settings: TemplateParameterLayer -> TemplateSection
- github.com/go-go-golems/glazed/pkg/settings: JqParameterLayer -> JqSection
- github.com/go-go-golems/glazed/pkg/settings: SortParameterLayer -> SortSection
- github.com/go-go-golems/glazed/pkg/settings: SkipLimitParameterLayer -> SkipLimitSection
- github.com/go-go-golems/glazed/pkg/settings: NewFieldsFiltersParameterLayer -> NewFieldsFiltersSection
- github.com/go-go-golems/glazed/pkg/settings: NewOutputParameterLayer -> NewOutputSection
- github.com/go-go-golems/glazed/pkg/settings: NewRenameParameterLayer -> NewRenameSection
- github.com/go-go-golems/glazed/pkg/settings: NewReplaceParameterLayer -> NewReplaceSection
- github.com/go-go-golems/glazed/pkg/settings: NewSelectParameterLayer -> NewSelectSection
- github.com/go-go-golems/glazed/pkg/settings: NewTemplateParameterLayer -> NewTemplateSection
- github.com/go-go-golems/glazed/pkg/settings: NewJqParameterLayer -> NewJqSection
- github.com/go-go-golems/glazed/pkg/settings: NewSortParameterLayer -> NewSortSection
- github.com/go-go-golems/glazed/pkg/settings: NewSkipLimitParameterLayer -> NewSkipLimitSection
- github.com/go-go-golems/glazed/pkg/settings: WithOutputParameterLayerOptions -> WithOutputSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithSelectParameterLayerOptions -> WithSelectSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithTemplateParameterLayerOptions -> WithTemplateSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithRenameParameterLayerOptions -> WithRenameSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithReplaceParameterLayerOptions -> WithReplaceSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithFieldsFiltersParameterLayerOptions -> WithFieldsFiltersSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithJqParameterLayerOptions -> WithJqSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithSortParameterLayerOptions -> WithSortSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: WithSkipLimitParameterLayerOptions -> WithSkipLimitSectionOptions
- github.com/go-go-golems/glazed/pkg/settings: GlazedTemplateLayerSlug -> GlazedTemplateSectionSlug
- github.com/go-go-golems/glazed/pkg/settings: NewSelectSettingsFromParameters -> NewSelectSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: NewRenameSettingsFromParameters -> NewRenameSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: NewReplaceSettingsFromParameters -> NewReplaceSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: NewJqSettingsFromParameters -> NewJqSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: NewSortSettingsFromParameters -> NewSortSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: NewSkipLimitSettingsFromParameters -> NewSkipLimitSettingsFromValues
- github.com/go-go-golems/glazed/pkg/settings: glazedLayer -> glazedValues
- github.com/go-go-golems/glazed/pkg/cmds: Layers -> Schema
- github.com/go-go-golems/glazed/pkg/cmds: WithLayersList -> WithSections
- github.com/go-go-golems/glazed/pkg/cmds: WithLayers -> WithSchema
- github.com/go-go-golems/glazed/pkg/cmds: WithLayersMap -> WithSectionsMap
- github.com/go-go-golems/glazed/pkg/cmds: WithReplaceLayers -> WithReplaceSections
- github.com/go-go-golems/glazed/pkg/cmds: GetDefaultLayer -> GetDefaultSection
- github.com/go-go-golems/glazed/pkg/cmds: GetLayer -> GetSection
- github.com/go-go-golems/glazed/pkg/cmds: SetLayers -> SetSections
- github.com/go-go-golems/glazed/pkg/cmds/logging: LoggingLayerSlug -> LoggingSectionSlug
- github.com/go-go-golems/glazed/pkg/cmds/logging: NewLoggingLayer -> NewLoggingSection
- github.com/go-go-golems/glazed/pkg/cmds/logging: AddLoggingLayerToCommand -> AddLoggingSectionToCommand
- github.com/go-go-golems/glazed/pkg/cmds/logging: AddLoggingLayerToRootCommand -> AddLoggingSectionToRootCommand
- github.com/go-go-golems/glazed/pkg/codegen: ParameterDefinitionToDict -> FieldDefinitionToDict
- github.com/go-go-golems/glazed/pkg/helpers/templating: toUrlParameter -> toUrlField
- github.com/go-go-golems/glazed/pkg/cli/cliopatra: Parameter -> Field
- github.com/go-go-golems/glazed/pkg/cli/cliopatra: getCliopatraParameters -> getCliopatraFields

### Appendix D: Ticket scripts inventory

- 01-exhaustive-parameter-layer-audit.py
- 02-symbol-inventory.go
- 03-render-symbol-inventory.py
- 04-rename-add-layer-to-section.py
- 05-rename-glazed-api.go
- 06-rename-parsed-fields.py
- 07-rename-decode-into.py
- 08-rename-dot-parameters.py
- 09-rename-values-decode.py
- 10-rename-yaml-parameters.py
- 11-layer-parameter-inventory.py
- 12-rename-symbols.go
- 12-rename-symbols.yaml
- 13-rename-schema-tests.py
- 14-rename-sources-test-yaml.py
- 15-rename-custom-profiles-test.py
- 16-rename-schema-tests.py
- 17-rename-settings-language.py
- 18-rename-fields-files.sh
- 19-rename-fields-language.py
- 20-rename-values-tests.py
- 21-rename-field-types-example.py
- 22-rename-doc-terms.py
- 23-rename-initialize-struct-docs.py
- 24-update-docs-addfields-credentials.py
- 25-sanitize-remarkable-bundle.py
- 26-build-postmortem-appendix.py

### Appendix E: Validation commands (as executed)

- `rg -n -i "layer|parameter" glazed -g '!**/ttmp/**'`
- `go test ./...`
- `golangci-lint run -v --max-same-issues=100`
- `gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=ttmp -exclude-dir=.history ./...`
- `govulncheck ./...`
