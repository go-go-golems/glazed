# Changelog

## 2026-05-12

- Initial workspace created


## 2026-05-12

Created issue 556 analysis package with evidence scripts, reproduction log, intern-oriented implementation guide, and diary.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/design-doc/01-required-fields-after-env-and-config-resolution-design.md — Primary design and implementation guide
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Chronological investigation diary
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/scripts/02-reproduce-required-env-parser.sh — Executable reproduction for required env failure


## 2026-05-12

Validated docs with docmgr doctor and uploaded bundled design package to reMarkable at /ai/2026/05/12/GLAZED-556-REQUIRED-ENV/GLAZED_556_Required_Env_Design.pdf.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/design-doc/01-required-fields-after-env-and-config-resolution-design.md — Uploaded in reMarkable bundle
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Uploaded in reMarkable bundle
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/sources/01-github-issue-556.md — Uploaded in reMarkable bundle


## 2026-05-12

Updated design guide with --print-parsed-fields and --help required-validation skip semantics plus detailed implementation tasks.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/design-doc/01-required-fields-after-env-and-config-resolution-design.md — Guide updated for control-path validation policy
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/tasks.md — Detailed implementation tasks added


## 2026-05-12

Implemented final required-value validation after source merging, with env/config regression tests and validation skips for print-parsed-fields/help.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra-parser.go — Wires conditional final required validation after source execution
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra_parser_config_test.go — Adds issue 556 regression coverage
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/schema/section-impl.go — Makes Cobra source collection ignore requiredness
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/validate_required.go — Adds final required-value validator
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records implementation step


## 2026-05-12

Ran full repository validation for required-env fix: go test ./... -count=1 passed.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records full-test validation


## 2026-05-12

Updated user-facing docs for final required-value validation and diagnostic skip behavior.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/doc/topics/16-parsing-fields.md — Clarifies low-level required parsing versus Cobra final-value validation
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/doc/topics/24-config-files.md — Documents config/env satisfying required fields and diagnostic skip behavior
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records docs update step


## 2026-05-12

Re-uploaded refreshed ticket bundle to reMarkable after implementation and docs updates.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records refreshed reMarkable upload


## 2026-05-12

Addressed PR 573 review comments: validate against middleware-filtered schema and skip required validation for all diagnostic command settings.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra-parser.go — Uses filtered schema for required validation and skips print-yaml/print-schema
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra_parser_config_test.go — Adds filtered-schema and diagnostic skip regression tests
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/middlewares.go — Adds ExecuteWithSchema for validation against the middleware-mutated schema clone
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records PR review response step


## 2026-05-12

Fixed golangci-lint exhaustive warnings in required-value emptiness checks.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/validate_required.go — Adds exhaustive ignore comments for intentional partial switches
- /home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/reference/01-investigation-diary.md — Records lint fix and validation

