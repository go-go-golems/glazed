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

- [ ] Design suppression syntax and semantics.
- [ ] Implement comment parsing for `glazedclilint:ignore` and `glazedclilint:file-ignore`.
- [ ] Require non-empty reasons for suppressions.
- [ ] Wire raw env, raw flag, and Glazed-without-rows diagnostics through suppression-aware reporting.
- [ ] Add analysistest fixtures for inline, previous-line, file-scoped, and invalid suppressions.
- [ ] Run `go test ./pkg/analysis/glazedclilint -count=1`.
- [ ] Run `go test ./...` or the repo-appropriate full validation.
- [ ] Update diary and changelog.
- [ ] Commit implementation and docs.
- [ ] Report readiness for downstream allow-scope reduction pass.
