---
Title: Analysis and implementation plan for removing Logstash support
Slug: analysis-logstash-removal
Short: Complete inventory of Logstash code and step-by-step removal plan
Topics:
- backend
- logging
- refactor
- cleanup
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Analysis: Removing Logstash Logging Support from Glazed

## Overview

Logstash logging support was added to glazed as a centralized logging option, but it is no longer needed. This document inventories every file, struct field, flag, function, and documentation reference that must be removed or updated, and provides an ordered implementation plan.

## Files to Modify or Delete

### 1. `glazed/pkg/cmds/logging/logstash_writer.go` — DELETE
- Entire file: `LogstashWriter` struct, `NewLogstashWriter`, `Write`, `ensureConnection`, `Close`, `SetupLogstashLogger`
- No external references outside the `logging` package itself.

### 2. `glazed/pkg/cmds/logging/section.go` — MODIFY
- Remove fields from `LoggingSettings`:
  - `LogstashEnabled`
  - `LogstashHost`
  - `LogstashPort`
  - `LogstashProtocol`
  - `LogstashAppName`
  - `LogstashEnvironment`
- Remove 6 field definitions from `NewLoggingSection()`:
  - `logstash-enabled`
  - `logstash-host`
  - `logstash-port`
  - `logstash-protocol`
  - `logstash-app-name`
  - `logstash-environment`
- Remove 6 `PersistentFlags` from `AddLoggingSectionToRootCommand()`:
  - `--logstash-enabled`
  - `--logstash-host`
  - `--logstash-port`
  - `--logstash-protocol`
  - `--logstash-app-name`
  - `--logstash-environment`

### 3. `glazed/pkg/cmds/logging/init.go` — MODIFY
- Remove Logstash initialization block in `InitLoggerFromSettings` (~L70-90).
- Remove logstash flag reads in `InitLoggerFromCobra` (~L155-177).
- Remove `Logstash*` fields from the `settings` struct built in `InitLoggerFromCobra`.
- Remove `Bool("logstash", ...)` from the debug log at end of `InitLoggerFromSettings`.

### 4. `glazed/pkg/cmds/logging/init-early.go` — MODIFY
- Remove logstash entries from `allowedKV` and `allowedBool` maps in `filterEarlyLoggingArgs`.
- Remove logstash flag definitions in `InitEarlyLoggingFromArgs` (~L99-104).
- Remove logstash fields from the `LoggingSettings` constructed at the end.

### 5. `glazed/pkg/cmds/logging/README.md` — MODIFY
- Remove "Logstash Integration" bullet and any logstash references.

### 6. `glazed/pkg/doc/topics/logging-section.md` — MODIFY
- Remove all Logstash rows from the flags table.
- Remove the "Centralized Logging" section.
- Remove `Logstash*` fields from the `LoggingSettings` struct snippet.
- Remove "Logstash Connection Failures" from Common Issues.
- Remove logstash YAML from Configuration Examples.
- Remove logstash env vars.

### 7. `glazed/pkg/doc/tutorials/custom-section.md` — MODIFY
- Remove Logstash references from prose and code snippets.
- Remove `includeLogstash` field and `WithLogstash` option from tutorial code.

### 8. `glazed/pkg/doc/tutorials/05-build-first-command.md` — MODIFY
- Remove logstash flag comment reference.

### 9. `clay/examples/logstash/logstash_example.go` — DELETE
- Entire file and its parent directory `clay/examples/logstash/`.

### 10. `clay/examples/simple/logging_layer_example.go` — MODIFY
- Remove Logstash references from `Long` description.
- Remove logstash usage examples from the comment block at the bottom.

## Implementation Order

1. **Delete core implementation** — Remove `logstash_writer.go` and all logstash fields/flags from `section.go`, `init.go`, `init-early.go`.
2. **Update documentation** — Clean up `README.md`, `logging-section.md`, `custom-section.md`, `05-build-first-command.md`.
3. **Delete/update examples** — Remove `clay/examples/logstash/` and update `logging_layer_example.go`.
4. **Build + test** — Run `go build ./...` and `go test ./...` in both `glazed/` and `clay/` workspaces.
5. **Commit** — One or more focused commits.

## Risk Assessment

- **Low risk**: Logstash support is completely self-contained in the `logging` package. No other glazed packages import `logstash_writer.go` symbols.
- **Breaking change**: Any external users passing `--logstash-*` flags will get "unknown flag" errors. This is intentional and expected.
- **No API surface change**: `LoggingSettings` struct tag changes are backward-compatible for JSON/YAML decode (extra fields are ignored), but Go code using the struct fields will break at compile time — again, intentional.

## Validation Checklist

- [ ] `go build ./...` passes in `glazed/`
- [ ] `go test ./...` passes in `glazed/`
- [ ] `go build ./...` passes in `clay/`
- [ ] `go test ./...` passes in `clay/`
- [ ] No remaining `logstash` / `Logstash` references in `.go` files outside of `ttmp/` history
- [ ] No remaining `logstash` / `Logstash` references in `.md` files outside of `ttmp/` history
