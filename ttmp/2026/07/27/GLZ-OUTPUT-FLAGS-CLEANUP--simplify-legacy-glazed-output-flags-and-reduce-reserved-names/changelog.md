# Changelog

## 2026-07-27

- Initial workspace created


## 2026-07-27

Completed evidence-backed audit and intern implementation design: reduce 44 automatic flags to --glazed-output, retain six stdout formats, and remove Excel and embedded jq

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/cli/cobra.go — Automatic injection and collision boundary
- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/settings/glazed_section.go — Primary legacy aggregation and processor wiring under analysis


## 2026-07-27

Validated all ticket frontmatter and passed docmgr doctor with all checks clean


## 2026-07-27

Uploaded validated ticket bundle GLZ OUTPUT FLAGS CLEANUP.pdf to /ai/2026/07/27/GLZ-OUTPUT-FLAGS-CLEANUP after a successful dry run

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/ttmp/2026/07/27/GLZ-OUTPUT-FLAGS-CLEANUP--simplify-legacy-glazed-output-flags-and-reduce-reserved-names/design-doc/01-legacy-output-flag-cleanup-analysis-design-and-implementation-guide.md — Primary document delivered in the reMarkable bundle


## 2026-07-27

Added open implementation phase tasks; ticket remains active after completion of the research and delivery milestone


## 2026-07-27

Superseded the proposed --glazed-output name with user-approved --format; updated the API contract, tests, migration examples, and renamed help export business selection to --export-mode

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/help/cmd/export.go — Known application format collision that must migrate to export-mode


## 2026-07-27

Refreshed the reMarkable bundle with the accepted --format design at /ai/2026/07/27/GLZ-OUTPUT-FLAGS-CLEANUP


## 2026-07-27

Implemented three-flag structured output (--format, --output-fields, --max-output-rows), atomic command mounting, JSONL, docs cleanup, and removal of Excelize and embedded jq; full tests/build/vet/lint pass

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/cli/cobra.go — Automatic injection and atomic registration
- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/doc/topics/32-structured-output.md — Canonical documentation
- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/settings/structured_output.go — Final structured-output implementation


## 2026-07-27

Ticket closed


## 2026-07-27

Addressed GitHub issue #611: generated Cobra commands and aliases now return typed errors through RunE instead of terminating through cobra.CheckErr

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/cli/cobra.go — RunE-based application-owned error handling
- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/pkg/cli/cobra_error_test.go — Regression coverage


## 2026-07-27

Resolved govulncheck findings GO-2026-5970 and GO-2026-5158 by upgrading x/text to v0.39.0 and aligned OpenTelemetry modules to v1.42.0

### Related Files

- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/go.mod — Fixed dependency versions
- /home/manuel/workspaces/2026-07-27/glazed-cleanup/glazed/go.sum — Updated dependency checksums

