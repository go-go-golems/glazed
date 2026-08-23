---
Title: glazed-migrate examples
Slug: glazed-migrate-examples
Short: Copy-paste recipes for surveying and fixing removed Glazed API usages.
Topics:
- commands
- migration
Commands:
- check
- fix
Flags:
- paths
- format
- output-fields
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

## Survey a whole repository

```bash
glazed-migrate check ./...
```

Rows: one per finding with `file`, `line`, `column`, `message`,
`fixes_available`.

## Survey as JSON lines for scripting

```bash
glazed-migrate check ./... --format jsonl
```

## Count how many findings have automatic fixes

```bash
glazed-migrate check . --format json --output-fields fixes_available
```

## Check a single package or file

```bash
glazed-migrate check ./pkg/cli
glazed-migrate check ./pkg/cli/cobra.go
```

## Fix everything in place

```bash
glazed-migrate fix ./...
git diff   # always review; the scanner does not type-check
```

Rows: one per modified file with `edits_applied`, then one row per finding
that requires manual redesign (`manual migration required: ...`).

## Fix only one directory first

```bash
glazed-migrate fix ./pkg/settings
go build ./pkg/settings/... && go test ./pkg/settings/...
```

## Use in CI to prevent regressions

```bash
test -z "$(glazed-migrate check . --format jsonl)"
```

## Run the vettool variant instead

```bash
go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate-analyzer@latest ./...
go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate-analyzer@latest -fix ./...
```
