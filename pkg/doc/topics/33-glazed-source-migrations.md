---
Title: Automated Glazed source migrations
Slug: glazed-source-migrations
Short: Apply syntax-aware Glazed API migrations with glazed-migrate.
Topics:
- glazed
- cli
- migration
Commands:
- glazed-migrate
Flags:
- fix
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

`glazed-migrate` is a `go/analysis` source migration command. It identifies removed Glazed APIs by syntax and import path and emits standard suggested fixes. The analyzer runs despite type-checking errors, which is necessary because code that still calls a removed symbol such as `settings.NewGlazedSchema` no longer compiles against Glazed v1.4.

## Migrating NewGlazedSchema

Build the migration command and preview findings:

```bash
make glazed-migrate-build
/tmp/glazed-migrate ./...
```

Apply safe fixes in place:

```bash
/tmp/glazed-migrate -fix ./...
gofmt -w .
go test ./...
```

You can also run the released command without cloning Glazed:

```bash
go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate@latest -fix ./...
```

A call without legacy options is rewritten as follows:

```go
// Before
outputSection, err := settings.NewGlazedSchema()

// After
outputSection, err := settings.NewStructuredOutputSection()
```

The migration recognizes default imports, renamed imports, and dot imports from `github.com/go-go-golems/glazed/pkg/settings`. It does not rewrite a function with the same name from another package.

Calls that pass legacy `GlazeSectionOption` values are diagnosed but not rewritten automatically:

```go
outputSection, err := settings.NewGlazedSchema(
    settings.WithFieldsFiltersSectionOptions(...),
)
```

Those options configured output features removed in v1.4 and are not type-compatible with `schema.SectionOption`. Replace the constructor manually, retain only the supported `--format`, `--output-fields`, and `--max-output-rows` behavior, and move domain-specific behavior into command fields or explicit middleware.

Always inspect the diff and run the target repository's formatter, tests, and build after applying fixes. The migration changes source syntax; it does not determine whether a command should continue mounting an output section or instead rely on `cli.BuildCobraCommandFromCommand` to inject it.

## See Also

- `glaze help structured-output` — current structured-output flags and runtime behavior.
- `glaze help glazed-cli-lint` — Glazed command-policy analysis.
- `glaze help 05-build-first-command` — current command construction pattern.
