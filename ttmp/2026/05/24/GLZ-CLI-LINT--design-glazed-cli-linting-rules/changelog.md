# Changelog

## 2026-05-24

- Created ticket `GLZ-CLI-LINT` for a Glazed-specific CLI policy linter.
- Added design document `design-doc/01-glazed-cli-linting-rules-analysis-and-implementation-guide.md`.
- Added diary document `reference/01-investigation-diary.md`.
- Documented the Geppetto `go/analysis`/vettool packaging pattern from `turnsdatalint` and `geppetto-lint`.
- Documented Pinocchio downstream vettool usage from `pinocchio/Makefile`.
- Documented implementation designs for three rules: direct `os.Getenv`, Glazed output sections on non-structured commands, and raw Cobra/go flag usage in CLI verbs.
- Added follow-up implementation checklist in `tasks.md`.

## 2026-05-24

Created intern-oriented Glazed CLI linter design package and diary, with Geppetto analyzer precedent and Pinocchio downstream usage mapped.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/design-doc/01-glazed-cli-linting-rules-analysis-and-implementation-guide.md — Primary design deliverable
- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/reference/01-investigation-diary.md — Chronological investigation diary


## 2026-05-24

Validated ticket with docmgr doctor and uploaded final bundle to reMarkable at /ai/2026/05/24/GLZ-CLI-LINT.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/reference/01-investigation-diary.md — Diary records doctor validation and reMarkable upload evidence


## 2026-05-24

Implemented glazedclilint analyzer, singlechecker/multichecker vettools, Makefile targets, analysistest fixtures, and Glazed help topic (commit c657de352a8801dd8891ce863ef3063f4a000c63).

### Related Files

- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/Makefile — Vettool targets
- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/pkg/analysis/glazedclilint/analyzer.go — Main analyzer implementation
- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/pkg/doc/topics/31-glazed-cli-lint.md — Glazed help entry


## 2026-05-24

Updated task checklist and diary after implementation, docmgr doctor passed, and refreshed the reMarkable bundle at /ai/2026/05/24/GLZ-CLI-LINT.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/reference/01-investigation-diary.md — Implementation and delivery diary
- /home/manuel/workspaces/2026-05-24/add-glazed-linters/glazed/ttmp/2026/05/24/GLZ-CLI-LINT--design-glazed-cli-linting-rules/tasks.md — Completed implementation checklist with rollout deferrals

