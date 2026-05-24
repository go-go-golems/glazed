---
Title: Glazed CLI linting with glazedclilint
Slug: glazed-cli-lint
Short: Enforce Glazed CLI command conventions with the custom go vet analyzer.
Topics:
- glazed
- cli
- linting
- cobra
Commands:
- glazed-lint
- glazedclilint
Flags:
- allow-tests
- allow-generated
- allow-paths
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

`glazedclilint` is Glazed's custom `go/analysis` linter for CLI command policy. It catches patterns that compile successfully but bypass the Glazed command model: direct environment reads, raw Cobra or Go flag definitions, and Glazed output flags on commands that do not emit structured rows.

The analyzer is packaged as a Go vettool. Use the focused `glazedclilint` command when developing the analyzer itself, and use the bundled `glazed-lint` command when running Glazed's custom analyzer suite in local checks or downstream repositories.

## What the linter checks

The linter enforces three conventions that keep Glazed CLIs predictable, inspectable, and easy to configure.

### Direct environment reads

Command code should not call `os.Getenv` directly for user-facing configuration. Direct reads do not appear in the command schema, do not show up in generated help, and do not participate in Glazed value-source precedence.

Prefer declaring a field and resolving the environment through Glazed parser sources:

```go
cmd_sources.FromEnv("MYAPP", fields.WithSource("env"))
```

or use the built-in parser path with an application name when no custom middleware chain is needed:

```go
cli.WithParserConfig(cli.CobraParserConfig{
    AppName: "myapp",
})
```

Use `os.UserHomeDir` for home-directory lookup when the value is not command configuration.

### Raw Cobra, pflag, or Go flags

Normal Glazed CLI verbs should not define user-facing flags with raw Cobra or Go flag APIs:

```go
cmd.Flags().StringVar(&address, "address", ":8088", "Address to listen on")
flag.String("config", "", "Config file")
```

Declare flags in the command description instead:

```go
cmds.NewCommandDescription(
    "serve",
    cmds.WithFlags(
        fields.New(
            "address",
            fields.TypeString,
            fields.WithDefault(":8088"),
            fields.WithHelp("Address to listen on"),
        ),
    ),
)
```

This lets Glazed expose the flag in schemas, help pages, aliases, config/env resolution, defaults, and command generation tools.

### Glazed output flags on non-row commands

The Glazed output section adds flags such as `--output`, `--fields`, `--jq`, sorting, templating, and skip/limit controls. Those flags only make sense when a command emits rows through `RunIntoGlazeProcessor`.

Do not attach `settings.NewGlazedSection` or `settings.NewGlazedSchema` to a command that only implements `cmds.BareCommand` or `cmds.WriterCommand`:

```go
// Bad: this exposes --output and --fields, but the command writes text itself.
type TextCommand struct { *cmds.CommandDescription }
var _ cmds.WriterCommand = (*TextCommand)(nil)

func NewTextCommand() (*TextCommand, error) {
    glazedSection, _ := settings.NewGlazedSection()
    return &TextCommand{CommandDescription: cmds.NewCommandDescription(
        "text",
        cmds.WithSections(glazedSection),
    )}, nil
}
```

Either remove the Glazed output section or implement structured output:

```go
type RowsCommand struct { *cmds.CommandDescription }
var _ cmds.GlazeCommand = (*RowsCommand)(nil)

func (c *RowsCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsed *values.Values,
    gp middlewares.Processor,
) error {
    return gp.AddRow(ctx, types.NewRow(types.MRP("status", "ok")))
}
```

## Running the linter

Build and run the bundled vettool from the Glazed repository:

```bash
make glazed-lint-build
go vet -vettool=/tmp/glazed-lint ./cmd/... ./pkg/...
```

or use the convenience target:

```bash
make glazed-lint
```

For focused analyzer development, build the single-analyzer command:

```bash
go build -o /tmp/glazedclilint ./cmd/tools/glazedclilint
go vet -vettool=/tmp/glazedclilint ./cmd/... ./pkg/...
```

Run the analyzer tests with:

```bash
go test ./pkg/analysis/glazedclilint -count=1
```

## Analyzer flags

The vettool exposes analyzer flags through `go vet` using the analyzer name as a prefix.

```bash
go vet -vettool=/tmp/glazed-lint \
  -glazedclilint.allow-tests=true \
  -glazedclilint.allow-generated=true \
  -glazedclilint.allow-paths='pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/' \
  ./cmd/... ./pkg/...
```

| Flag | Default | Meaning |
|---|---:|---|
| `allow-tests` | `true` | Skip `_test.go` files. |
| `allow-generated` | `true` | Skip files with standard `Code generated ... DO NOT EDIT` comments. |
| `allow-paths` | framework bridge paths | Comma-separated path fragments where raw Cobra/env usage is allowed. |

The default allowlist exists because Glazed's framework bridge packages must use Cobra, pflag, analyzer flags, environment-backed source adapters, and rendering-library environment conventions internally. Application verbs and examples under `cmd/` should still use `cmds.WithFlags` and Glazed sections.

## How to fix findings

When the linter reports a finding, first identify which convention was violated.

| Finding | Cause | Fix |
|---|---|---|
| `use Glazed config/env middleware...` | Command reads process env directly. | Add a Glazed field and resolve env through `cmd_sources.FromEnv`, `CobraParserConfig.AppName`, or config plans. |
| `define CLI flags with cmds.WithFlags...` | Command defines user-facing flags via Cobra, pflag, or `flag`. | Convert the command to `cmds.BareCommand`, `cmds.WriterCommand`, or `cmds.GlazeCommand` and declare flags with `fields.New`. |
| `exposes Glazed output flags but does not implement RunIntoGlazeProcessor` | Command adds `settings.NewGlazedSection` but does not emit structured rows. | Remove the Glazed output section or implement `RunIntoGlazeProcessor`. |

Prefer changing the command shape over suppressing the diagnostic. A suppression or allowlist should mean "this file is framework bridge code", not "this command is hard to migrate".

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| The linter flags framework code in `pkg/cli` or help internals | The file path is not covered by `allow-paths`. | Add a narrow path fragment to `-glazedclilint.allow-paths` and document why that package is framework bridge code. |
| A command with `--output` is flagged | The command does not implement `RunIntoGlazeProcessor`. | Either remove `settings.NewGlazedSection` or convert the command to `cmds.GlazeCommand`. |
| A raw Cobra command is flagged but it only starts a server | Server commands can still be Glazed `BareCommand`s. | Put flags in `cmds.WithFlags`, decode with `values.DecodeSectionInto`, and call the server function from `Run`. |
| Downstream CI installs a stale vettool | The Makefile uses `@latest` instead of the module version. | Install `github.com/go-go-golems/glazed/cmd/tools/glazed-lint@$(GLAZED_VERSION)` or fall back to workspace install for `(devel)`. |
| The analyzer misses a complex command constructor | The v1 inference handles common constructor return patterns. | Add a small regression test in `pkg/analysis/glazedclilint/testdata` and extend command type inference. |

## See Also

- `glaze help 05-build-first-command` — canonical Glazed command structure and `RunIntoGlazeProcessor`.
- `glaze help 13-sections-and-values` — how sections and parsed values fit together.
- `glaze help 21-cmds-middlewares` — how Glazed value sources replace ad-hoc flag and env reads.
- `glaze help 07-dual-commands` — how to design commands that can support both classic and structured output.
- `glaze help writing-help-entries` — how to add or update Glazed help pages.
