# Changelog

## 2026-04-08

- Initial workspace created


## 2026-04-08

Completed pre-PR review: analyzed 21h session with go-minitrace, found 2 critical bugs, 3 dead code patterns, 5 significant issues

### Related Files

- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/dsl_bridge.go — O(N) performance bug
- /home/manuel/workspaces/2026-04-07/glaze-help-browser/glazed/pkg/help/store/compat.go — Dead code


## 2026-04-08

Added 20 detailed cleanup tasks across 8 phases: delete dead code (T1-T4), fix DSL bridge bug (T5), consolidate markdown parsing (T6-T8), eliminate Section wrapper (T9-T11), eliminate SectionQuery builder (T12-T15), fix server FTS5 bypass (T16), frontend cleanup (T17-T18), final verification (T19-T20)


## 2026-04-08

Implemented cleanup plan end-to-end: removed dead wrappers, unified markdown parsing, eliminated help.Section + SectionQuery, switched server search to store.TextSearch, cleaned frontend types/API, and validated with full Go/web builds/tests (commits d97240c..b4d879a)

