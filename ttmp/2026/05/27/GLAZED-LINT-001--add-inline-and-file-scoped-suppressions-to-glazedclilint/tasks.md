---
Title: Tasks
Ticket: GLAZED-LINT-001
Status: active
Topics:
  - cli
  - automation
  - linting
DocType: tasks
Intent: short-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: Task checklist for adding inline and file scoped glazedclilint suppressions.
LastUpdated: 2026-05-27T15:10:00-04:00
---

# Tasks

- [x] Design suppression syntax and semantics.
- [x] Implement comment parsing for `glazedclilint:ignore` and `glazedclilint:file-ignore`.
- [x] Require non-empty reasons for suppressions.
- [x] Wire raw env, raw flag, and Glazed-without-rows diagnostics through suppression-aware reporting.
- [x] Add analysistest fixtures for inline, previous-line, file-scoped, and invalid suppressions.
- [x] Run `go test ./pkg/analysis/glazedclilint -count=1`.
- [x] Run `go test ./...` or the repo-appropriate full validation.
- [x] Update diary and changelog.
- [x] Commit implementation and docs.
- [x] Report readiness for downstream allow-scope reduction pass.
