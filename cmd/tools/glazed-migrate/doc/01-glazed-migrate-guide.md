---
Title: glazed-migrate guide
Slug: glazed-migrate-guide
Short: Find and fix usages of Glazed APIs removed by the structured-output cleanup.
Topics:
- commands
- migration
- refactoring
Commands:
- check
- fix
Flags:
- paths
- format
- output-fields
- max-output-rows
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

`glazed-migrate` is the command-framework front end for the `glazedmigration`
analyzer (`pkg/analysis/glazedmigration`). It scans Go source for usages of
Glazed APIs that were removed by the structured-output cleanup
(GLZ-OUTPUT-FLAGS-CLEANUP) and either reports them or rewrites them in place.

## The two commands

- `glazed-migrate check [paths...]` — report findings, one row per finding
  (`file`, `line`, `column`, `message`, `fixes_available`). No files are
  modified. Use `--format json` to feed the findings into other tooling.
- `glazed-migrate fix [paths...]` — apply every finding that has an automatic
  migration, editing files in place. One row per modified file
  (`edits_applied`), plus one row per finding that needs a manual redesign.

Both accept a variadic `paths` argument (default `.`): directories are walked
recursively (skipping hidden directories, `vendor`, `node_modules`, and
`testdata`), or you can pass individual `.go` files.

## What gets migrated automatically

| Rule | Removed API | Replacement |
|------|-------------|-------------|
| R1/R2 | `settings.NewGlazedSchema()` / `settings.NewGlazedSection()` | `settings.NewStructuredOutputSection()` |
| R3 | `settings.WithOutputSectionOptions(...)` wrapper | unwrapped into the section constructor |
| R4 | default-map key `"output"` | `"format"` |
| R5 | `settings.GlazedSlug` | `settings.StructuredOutputSlug` |

## What is only reported (manual redesign)

- `settings.Setup*` runtime helpers (R6/R7/R8) — replaced by explicit
  processor/formatter assembly; see `glaze help structured-output`.
- Removed feature sections (R9): select, rename, replace, template, jq, sort,
  skip-limit, fields-filters. Their options have no mechanical equivalent;
  move that logic into application fields.

## Why it works on broken code

The scanner parses source with `go/parser` and never type-checks. Code that
references removed APIs fails to compile against the current Glazed release —
that is exactly the code this tool must migrate. With no type information,
the rules fall back to import-aware AST matching (the same degraded mode the
analyzer's unit tests exercise). Local variables shadowing the `settings`
import can produce false positives in that mode, so always review the diff.

## Typical workflow

```bash
# 1. Survey the damage
glazed-migrate check ./... --format json

# 2. Apply automatic migrations
glazed-migrate fix ./...

# 3. Review, then rebuild and test
git diff
go build ./... && go test ./...
```

## CI usage

Fail the build when removed APIs linger:

```bash
test -z "$(glazed-migrate check . --format jsonl)"
```

## The vettool entry point

The analyzer also remains available as a `go/analysis` singlechecker for
`go vet`-style invocation:

```bash
go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate-analyzer@latest -fix ./...
```

Use the vettool when you want analysis-framework semantics (package-aware,
`//nolint`-style suppression via `go vet`); use `glazed-migrate` when you want
structured output, embedded help, and source-level operation on code that no
longer compiles.
