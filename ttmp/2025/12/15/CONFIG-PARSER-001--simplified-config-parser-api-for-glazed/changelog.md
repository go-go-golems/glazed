# Changelog

## 2025-12-15

- Initial workspace created
- Completed initial codebase exploration and architecture mapping (Glazed parsing system)
- Wrote comprehensive architecture analysis document
- Maintained detailed lab-book research diary (including searches/results and Moments `appconfig` prior art)
- Wrote design doc proposing a struct-first ConfigParser API with multiple implementation designs + tradeoffs
- Updated ticket `tasks.md` to reflect current status and next implementation steps

## 2025-12-16

- Added colleague interview/survey quiz to surface rationale and constraints behind current config parsing architecture and identify cleanup/merge/removal opportunities
- Reset CONFIG-PARSER-001 direction: added new `appconfig.Parser` design doc (register layers + tagged structs, configurable Parse middlewares) + a detailed redesign diary; marked the prior struct-first design as superseded


## 2025-12-16

Step 1: Implement pkg/appconfig.Parser v1 skeleton (commit bf627f0)

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/appconfig/doc.go — Package-level docs for appconfig
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/appconfig/options.go — ParserOption helpers for env/config files/middlewares
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/appconfig/parser.go — New Parser[T] type with Register + Parse using runner.ParseCommandParameters


## 2025-12-16

Step 2: Add unit tests for appconfig.Parser v1 contracts (commit d452edc)

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/appconfig/parser_test.go — Tests for Register/Parse invariants


## 2025-12-16

Step 3: Introduce appconfig.LayerSlug to encourage const slugs (commit 91b10b2)

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/appconfig/parser.go — Added LayerSlug type + Register now takes LayerSlug


## 2025-12-16

Step 4: Add minimal glazed example for appconfig.Parser (commit 22c9659)

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/cmd/examples/appconfig-parser/main.go — Demonstrates const LayerSlug

