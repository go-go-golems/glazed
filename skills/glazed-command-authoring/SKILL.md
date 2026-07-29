---
name: glazed-command-authoring
description: "Create, refactor, and validate Glazed commands using the current v1.4 structured-output, Cobra RunE, schema, middleware, help, and logging conventions. Use when adding a GlazeCommand, wiring command groups, defining fields or sections, emitting rows, choosing output behavior, or migrating code away from removed legacy output APIs."
---

# Glazed Command Authoring

Use this skill when creating or refactoring commands built with `github.com/go-go-golems/glazed`. It describes the v1.4 command contract: command implementations emit rows, Cobra builders add the minimal structured-output section, and applications retain ownership of domain flags and process errors.

## Required workflow

1. Inspect the current APIs and nearby first-party examples before writing code.
2. Define a command description and a settings struct with `glazed` tags.
3. Implement the appropriate command interface.
4. Build the Cobra command through `pkg/cli` rather than duplicating parser setup.
5. Keep domain behavior in application fields and serialization behavior in structured output.
6. Add focused tests for parsing, emitted rows, errors, and help-visible flags.
7. Run the complete validation sequence before finishing.

## Current public output contract

Every `cmds.GlazeCommand` built with `cli.BuildCobraCommandFromCommand` receives exactly these universal output flags:

```text
--format table|json|jsonl|csv|tsv|yaml
--output-fields field1,field2,...
--max-output-rows N
```

Their meanings are intentionally narrow:

- `--format` selects stdout serialization. The default is `table`.
- `--output-fields` projects emitted rows. Tabular formats preserve requested field order.
- `--max-output-rows` caps serialized rows. Zero means unlimited. It does not cancel source work.

Do not reintroduce removed universal flags for jq, templates, rename, replacement, sorting, filtering, SQL, Excel, output files, table styles, stream toggles, or skip/limit behavior.

Use application fields when an operation changes source behavior. A database limit, API page size, server-side filter, or domain sort belongs to the command even if it also changes the output.

## Canonical imports

```go
import (
    "context"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/values"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)
```

Common invalid imports:

- `glazed/pkg/cmds/parameters/fields` — use `pkg/cmds/fields`.
- `glazed/pkg/cmds/middlewares` — use `pkg/middlewares`.
- `glazed/pkg/values` — use `pkg/cmds/values`.
- `glazed/pkg/settings/schema` — use `pkg/cmds/schema`.

## Minimal GlazeCommand

The Cobra builder injects structured output automatically. Do not manually add the structured-output section to ordinary `GlazeCommand` descriptions.

```go
type ListCommand struct {
    *cmds.CommandDescription
}

type ListSettings struct {
    Query string `glazed:"query"`
    Limit int    `glazed:"limit"`
}

func NewListCommand() *ListCommand {
    return &ListCommand{CommandDescription: cmds.NewCommandDescription(
        "list",
        cmds.WithShort("List matching records"),
        cmds.WithFlags(
            fields.New(
                "query",
                fields.TypeString,
                fields.WithHelp("Domain query applied by the source"),
            ),
            fields.New(
                "limit",
                fields.TypeInteger,
                fields.WithDefault(100),
                fields.WithHelp("Maximum records requested from the source"),
            ),
        ),
    )}
}

func (c *ListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsed *values.Values,
    processor middlewares.Processor,
) error {
    settings := &ListSettings{}
    if err := parsed.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
        return err
    }

    records, err := loadRecords(ctx, settings.Query, settings.Limit)
    if err != nil {
        return err
    }
    for _, record := range records {
        if err := processor.AddRow(ctx, types.NewRow(
            types.MRP("id", record.ID),
            types.MRP("name", record.Name),
        )); err != nil {
            return err
        }
    }
    return nil
}
```

`types.NewRow` returns a `types.Row` value, not `*types.Row`. `Processor.AddRow` accepts that value.

## Fields and positional arguments

Define flags with `cmds.WithFlags`. Decode them through `values.Values`; do not read Cobra flags inside the domain command.

```go
fields.New(
    "status",
    fields.TypeChoice,
    fields.WithChoices("active", "archived"),
    fields.WithDefault("active"),
    fields.WithHelp("Status requested from the service"),
)
```

Define positional arguments with `cmds.WithArguments` and `fields.WithIsArgument(true)`:

```go
cmds.WithArguments(
    fields.New(
        "path",
        fields.TypeString,
        fields.WithIsArgument(true),
        fields.WithHelp("Input path"),
    ),
)
```

A variadic positional argument must use a list type such as `fields.TypeStringList`, must be the final positional argument, and is decoded through the same `glazed` struct tag mechanism.

## Building and registering Cobra commands

Build one command with:

```go
cobraCommand, err := cli.BuildCobraCommandFromCommand(command,
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpSections: []string{schema.DefaultSlug},
    }),
)
if err != nil {
    return err
}
root.AddCommand(cobraCommand)
```

For a collection, prefer `cli.AddCommandsToRootCommand`. It builds all commands and aliases before mounting any of them, so a schema or flag collision does not leave a partially mutated command tree.

Generated commands use Cobra `RunE`. Command errors, parsing failures, output setup failures, and processor-close failures propagate to the application's `Execute()` call. Do not call `cobra.CheckErr` or `os.Exit` from reusable command implementations. The application root owns error rendering, telemetry, cleanup, and exit-code mapping.

## Raw Cobra integration

A raw Cobra command can mount the same three output flags explicitly:

```go
if err := cli.AddStructuredOutputFlagsToCobraCommand(command); err != nil {
    return err
}
```

Create its processor after Cobra has parsed flags:

```go
processor, formatter, err := cli.CreateStructuredOutputProcessorFromCobra(command)
if err != nil {
    return err
}
_ = formatter
```

The helper returns setup errors. Callers must propagate or handle them; they must not terminate inside the helper.

## Structured-output APIs

Use `pkg/settings` only when a programmatic caller needs direct control over output setup:

```go
section, err := settings.NewStructuredOutputSection()
processor, formatter, err := settings.SetupStructuredOutput(sectionValues, writer)
processor, outputSettings, err := settings.SetupStructuredProcessor(sectionValues)
```

- `SetupStructuredOutput` applies projection, row capping, and serialization.
- `SetupStructuredProcessor` applies projection and row capping without attaching a formatter. Use it when the caller needs a `types.Table` or its own output middleware.

Do not use removed APIs such as `NewGlazedSection`, `NewGlazedSchema`, `GlazedSlug`, `SetupTableProcessor`, `SetupProcessorOutput`, or `WithOutputSectionOptions`.

## Middleware composition and streaming

The middleware architecture remains available even though the old middleware flags were removed. `TableProcessor` still supports object, row, and table middleware.

```go
processor, formatter, err := settings.SetupStructuredOutput(
    sectionValues,
    writer,
    middlewares.WithRowMiddleware(customRowMiddleware),
)
```

Caller-provided middleware runs before structured-output projection, row capping, and serialization. Formatters may prepend required normalization middleware, such as object flattening for CSV and TSV.

Execution characteristics:

- JSON and JSONL serialize through row middleware and do not accumulate a table.
- Table, CSV, TSV, and YAML collect accepted rows and serialize during `Close`.
- JSONL writes one compact JSON object per line.
- JSON streams array elements and completes array framing during `Close`.

Processor ownership determines who closes it:

- A `GlazeCommand.RunIntoGlazeProcessor` implementation receives a builder-owned processor. It must emit rows and return without closing the processor; the Cobra `RunE` path closes it exactly once.
- Code that directly creates a processor with `SetupStructuredOutput`, `SetupStructuredProcessor`, or `CreateStructuredOutputProcessorFromCobra` owns that processor and must close it exactly once after emission.

Closing runs table middleware and finalizes formatter framing. Closing a processor twice is invalid: buffered formats can be emitted twice and streaming JSON can receive duplicate closing syntax.

## Custom sections

Custom sections remain appropriate for reusable domain configuration such as database, authentication, or logging settings:

```go
section, err := schema.NewSection(
    "database",
    "Database",
    schema.WithFields(
        fields.New("database-url", fields.TypeString),
    ),
)
```

Add custom sections through `cmds.WithSections`. Do not duplicate a section on both a parent and child when that would mount the same flags twice.

## Root initialization

A complete Glazed root initializes logging and embedded help:

```go
root := &cobra.Command{
    Use:   "myapp",
    Short: "Application description",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return logging.InitLoggerFromCobra(cmd)
    },
}

if err := logging.AddLoggingSectionToRootCommand(root, "myapp"); err != nil {
    return err
}

helpSystem := help.NewHelpSystem()
if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
    return err
}
help_cmd.SetupCobraRootCommand(helpSystem, root)
```

The root owns shared logging and help registration. Child packages should register command groups without creating independent help systems.

If a custom parser `MiddlewaresFunc` replaces the default source chain, re-add every required source explicitly. Setting `AppName` does not restore environment loading after replacing the chain.

## Command tree layout

For larger applications, mirror the CLI tree in folders:

```text
cmd/myapp/
  main.go
  cmds/
    records/
      root.go
      list.go
      get.go
    users/
      root.go
      create.go
```

If users type `myapp records list`, place `list.go` under `cmds/records`. Each group should expose one registration function or Cobra constructor.

## Testing requirements

At minimum, test:

1. Settings decode into the correct default section.
2. The command emits expected rows and propagates processor errors.
3. `BuildCobraCommandFromCommand` adds `format`, `output-fields`, and `max-output-rows` for a `GlazeCommand`.
4. Domain fields do not collide with framework fields.
5. Typed errors reach `root.Execute()` when exit-code mapping matters.
6. Sparse rows preserve requested tabular projection order.
7. Streaming commands stop under context cancellation.

For sparse projection, include a case equivalent to:

```text
requested: [a, missing, b]
row 1:     {a: 1}
row 2:     {b: 2}
columns:   [a, b]
```

## Validation

Run from the module root. This repository may be nested under a mismatched workspace, so use `GOWORK=off` when the module toolchain must take precedence.

```bash
gofmt -w <changed-go-files>
GOWORK=off go test ./... -count=1
GOWORK=off go build ./...
GOWORK=off go vet ./...
GOWORK=off make glazed-lint
GOWORK=off make govulncheck
git diff --check
```

Before committing, inspect help for representative commands and confirm the structured-output group contains only the three universal flags.

## Repository references

Read these files when the task needs more detail:

- `pkg/doc/tutorials/05-build-first-command.md` — build-first command tutorial.
- `pkg/doc/topics/32-structured-output.md` — current output contract.
- `pkg/doc/topics/commands-reference.md` — command lifecycle and builders.
- `pkg/doc/topics/sections-guide.md` — sections and values.
- `pkg/doc/topics/07-dual-commands.md` — dual writer/Glaze commands.
- `pkg/settings/structured_output.go` — processor and formatter assembly.
- `pkg/cli/cobra.go` — automatic injection, registration, and `RunE` behavior.
- `pkg/cli/helpers.go` — raw Cobra integration.
- `pkg/middlewares/processor.go` — middleware execution model.
- `pkg/cli/structured_output_test.go` — flag-surface characterization.
- `pkg/cli/cobra_error_test.go` — error-propagation contract.

Use `glaze help --all` to discover embedded documentation and `glaze help <slug>` to read a focused page.
