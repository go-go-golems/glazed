---
Title: Legacy Output Flag Cleanup Analysis Design and Implementation Guide
Ticket: GLZ-OUTPUT-FLAGS-CLEANUP
Status: complete
Topics:
    - glazed
    - cli
    - cobra
    - middleware
    - settings
    - formatters
    - api-design
    - intern-guide
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: repo://go.mod
      Note: Removes Excelize and gojq dependency families
    - Path: repo://pkg/cli/cobra.go
      Note: Injects minimal output settings and mounts command trees atomically
    - Path: repo://pkg/doc/topics/32-structured-output.md
      Note: Canonical user-facing documentation
    - Path: repo://pkg/formatters/json/json.go
      Note: Implements explicit JSON array and compact JSONL constructors
    - Path: repo://pkg/help/cmd/export.go
      Note: Separates export-mode behavior from format serialization
    - Path: repo://pkg/middlewares/row/output-fields.go
      Note: Implements exact output field projection
    - Path: repo://pkg/settings/structured_output.go
      Note: Defines the final three-flag structured-output contract
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/600
Summary: Implemented hard cut from 44 legacy output and transformation flags to format, output-fields, and max-output-rows, with six serializers and removal of Excelize and embedded jq.
LastUpdated: 2026-07-27T16:30:00-04:00
WhatFor: Explain the implemented structured-output architecture, behavior, source layout, tests, and review path.
WhenToUse: Use when maintaining Glazed output serialization or reviewing the GLZ-OUTPUT-FLAGS-CLEANUP implementation.
---


# Legacy Output Flag Cleanup Analysis, Design, and Implementation Guide

## Executive summary

Glazed previously attached 44 output, transformation, rendering, pagination, and file-routing flags to every `cmds.GlazeCommand`. The implementation replaced that aggregate with three output controls:

```text
--format table|json|jsonl|csv|tsv|yaml
--output-fields field1,field2,...
--max-output-rows N
```

The change is a hard cut. The old settings YAML files, aggregate section, settings adapters, Excel formatter, embedded jq middleware, old-flag documentation, and dedicated legacy examples were deleted. There are no aliases or compatibility shims.

The implementation keeps application business fields distinct from output behavior:

- `--format` chooses byte serialization.
- `--output-fields` chooses which emitted fields reach the formatter.
- `--max-output-rows` caps rows reaching the formatter; zero means unlimited.
- Application-specific filtering, pagination, sorting, and projection remain ordinary command fields when they change source work.
- Ad hoc transformations of already-produced rows belong in tools such as `jq`.

## Final public contract

### `--format`

Supported values are exactly:

| Format | Framing | Execution mode |
|---|---|---|
| `table` | deterministic ASCII table | buffered table |
| `json` | one JSON array | streaming row formatter with array close |
| `jsonl` | one compact JSON object per line | streaming row formatter |
| `csv` | comma-separated rows with headers | buffered table |
| `tsv` | tab-separated rows with headers | buffered table |
| `yaml` | one YAML sequence | buffered table |

The default is `table`. JSONL is the explicit streaming contract; the old `stream` and `output-as-objects` matrix no longer exists.

### `--output-fields`

The flag accepts a string list. Empty means all fields. Values are trimmed and deduplicated while preserving the first requested occurrence.

The middleware constructs a new row by looking up fields in requested order. Missing fields are omitted. CSV, TSV, and table output preserve that order. JSON object key ordering is not part of the wire contract.

Example:

```bash
glaze json records.json --format csv --output-fields name,id
```

### `--max-output-rows`

The flag accepts a non-negative integer. Zero means unlimited. A negative value returns:

```text
max-output-rows must be greater than or equal to zero
```

The cap is applied as row middleware before serialization. It limits emitted bytes but does not promise upstream cancellation. Commands that can avoid network or database work should expose their own domain limit.

## Architecture

### Construction and execution flow

```text
CommandDescription
      |
      v
BuildCobraCommandFromCommandAndFunc
      |
      +-- GlazeCommand? inject structured-output section if absent
      |
      v
Cobra parser resolves application and structured-output values
      |
      v
SetupStructuredOutput
      |
      +-- output-fields projection
      +-- max-output-rows cap
      +-- format-specific row or table output middleware
      |
      v
RunIntoGlazeProcessor emits types.Row values
      |
      v
TableProcessor.Close completes buffered output and formatter close
```

### API

The implementation lives in `pkg/settings/structured_output.go`:

```go
type OutputFormat string

const (
    OutputTable OutputFormat = "table"
    OutputJSON  OutputFormat = "json"
    OutputJSONL OutputFormat = "jsonl"
    OutputCSV   OutputFormat = "csv"
    OutputTSV   OutputFormat = "tsv"
    OutputYAML  OutputFormat = "yaml"
)

const (
    StructuredOutputSlug = "structured-output"
    StructuredOutputFlag = "format"
)

type StructuredOutputSettings struct {
    Format        OutputFormat `glazed:"format"`
    OutputFields  []string     `glazed:"output-fields"`
    MaxOutputRows int          `glazed:"max-output-rows"`
}
```

Primary functions:

```go
func NewStructuredOutputSection(
    options ...schema.SectionOption,
) (*schema.SectionImpl, error)

func DecodeStructuredOutputSettings(
    sectionValues *values.SectionValues,
) (*StructuredOutputSettings, error)

func SetupStructuredProcessor(
    sectionValues *values.SectionValues,
    options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, *StructuredOutputSettings, error)

func SetupStructuredOutput(
    sectionValues *values.SectionValues,
    writer io.Writer,
    options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, formatters.OutputFormatter, error)
```

`SetupStructuredProcessor` exists for programmatic callers such as Lua that need projection and row caps but want a `types.Table` rather than serialized bytes.

## Middleware ordering

The processor is assembled in this order:

```text
caller-provided row middleware
format-required front middleware (for example CSV flattening)
output-fields projection
max-output-rows cap
row serializer, or table collection followed by table serializer
```

The formatter may prepend normalization middleware. CSV and TSV flatten nested objects before `--output-fields` is applied, so users can select flattened field names.

`OutputFieldsMiddleware` is intentionally narrower than the deleted generic field-filter middleware. It supports exact names and order only. It does not support regular expressions, exclusions, null removal, deduplication, or implicit `all` markers.

## Cobra collision behavior

The automatic framework reservation is now three names instead of 44. Applications can define former generic names such as:

```text
output, output-file, fields, filter, template, select, stream, sort-by
```

`format`, `output-fields`, and `max-output-rows` are framework-owned on `GlazeCommand` implementations.

`AddCommandsToRootCommand` now builds every command and alias before mounting any of them. A schema or flag collision returns an error and leaves the root unchanged:

```text
build all commands -> build all aliases -> if all succeeded, mount all
```

Errors include command full path and source.

### Runtime error propagation

Issue [#611](https://github.com/go-go-golems/glazed/issues/611) identified that generated commands used Cobra's non-returning `Run` callback and called `cobra.CheckErr`. That let a library callback print an error and terminate the process with status 1 before the embedding application's `Execute()` could inspect it.

Generated commands and aliases now use `RunE`. Parsing, settings, command execution, structured-output setup, and processor-close failures return through Cobra to the application. This preserves typed errors for application-owned logging, telemetry, cleanup, and exit-code mapping. `ExitWithoutGlazeError` remains a successful early return, and context cancellation retains its prior non-error behavior.

Tests cover bare commands, `GlazeCommand`, explicit `BuildCobraCommandFromCommandAndFunc` callbacks, and aliases. They assert that the original typed error reaches `Execute()`.

## Removed code and dependencies

### Deleted settings surface

The implementation removed:

```text
pkg/settings/glazed_section.go
pkg/settings/settings_output.go
pkg/settings/settings_fields-filters.go
pkg/settings/settings_jq.go
pkg/settings/settings_rename.go
pkg/settings/settings_replace.go
pkg/settings/settings_select.go
pkg/settings/settings_skip_limit.go
pkg/settings/settings_sort.go
pkg/settings/settings_template.go
pkg/settings/flags/
```

The old template settings tests were removed with their implementation.

### Deleted features

```text
pkg/formatters/excel/
pkg/middlewares/jq.go
pkg/middlewares/jq_test.go
```

`go mod tidy` removed:

- `github.com/xuri/excelize/v2`;
- `github.com/itchyny/gojq`;
- Excelize and gojq-only transitive dependencies including `xuri/efp`, `xuri/nfp`, `richardlehane/mscfb`, `richardlehane/msoleps`, and `itchyny/timefmt-go`.

SQL, template, simple, and row transformation packages remain available as explicit Go libraries. They are not mounted as universal CLI behavior.

## Application integrations

### Automatic `GlazeCommand` setup

`pkg/cli/cobra.go` injects `NewStructuredOutputSection` when a `GlazeCommand` description does not already contain the section. Runtime execution looks up `StructuredOutputSlug` and calls `SetupStructuredOutput`.

### Raw Cobra helpers

The explicit helpers are:

```go
cli.CreateStructuredOutputProcessorFromCobra(cmd)
cli.AddStructuredOutputFlagsToCobraCommand(cmd, options...)
```

Raw Cobra commands in the `glaze` binary were migrated to these names.

### Help export

Help export previously used application `--format glazed|files|sqlite` while Glazed used `--output` for serialization. The implementation now separates these concepts:

```text
--export-mode glazed|files|sqlite
--format table|json|jsonl|csv|tsv|yaml
```

The external binary loader invokes:

```text
<binary> help export --with-content=true --format json
```

### Lua and programmatic runner

- Lua uses `SetupStructuredProcessor` plus `NullTableMiddleware` to return a table without writing output.
- `pkg/cmds/runner` uses `SetupStructuredOutput` when no processor is supplied.

## Documentation cleanup

Legacy-only documentation and fixtures were deleted rather than converted into a migration catalog. Removed material includes dedicated pages for:

- generic field filters and regex filters;
- rename and replacement flags;
- embedded jq;
- select shortcuts;
- skip/limit and sort flags;
- output file splitting and SQL output flags;
- template flags and table-format styling;
- the old VHS output demo;
- the old facade migration page and old-output prompt.

Current documentation was updated in place. The canonical new reference is:

```text
pkg/doc/topics/32-structured-output.md
```

README, command tutorials, dual-command docs, help export docs, CLI lint docs, sections docs, examples, and external-help instructions now use the three-flag contract.

## Tests

### Settings tests

`pkg/settings/structured_output_test.go` covers:

- default table format;
- parsing JSONL, output fields, and max rows;
- trimming and deduplicating output fields;
- rejecting negative row caps;
- end-to-end projection and capping with JSONL.

### Middleware tests

`pkg/middlewares/row/output-fields_test.go` covers:

- requested tabular field order;
- omission of missing and unrequested fields;
- empty-list pass-through.

### JSON tests

`pkg/formatters/json/json_test.go` covers compact one-object-per-line JSONL and empty-stream behavior.

### CLI tests

`pkg/cli/structured_output_test.go` covers:

- exactly the three minimal framework flags;
- absence of representative legacy flags;
- coexistence with former generic application flag names;
- collision on application `format`;
- atomic root mounting on build failure.

### Runtime smoke tests

Representative commands:

```bash
GOWORK=off go run ./cmd/glaze json misc/test-data/2.json \
  --format jsonl --output-fields b,a --max-output-rows 1

GOWORK=off go run ./cmd/glaze json misc/test-data/2.json \
  --format csv --output-fields b,a --max-output-rows 1
```

The CSV result preserves requested order:

```text
b,a
20,10
```

## Failures encountered during implementation

The implementation initially exposed three useful test failures:

1. `OutputFieldsMiddleware` tests called a nonexistent `Row.GetValue`; they were corrected to use `Row.Get`.
2. JSONL assertions expected requested key order, but JSON conversion intentionally uses a map and object key order is not a semantic contract. The assertion and docs were corrected; tabular order remains tested.
3. `cmd/glaze/cmds/docs.go` passed a partially populated `StructuredOutputSettings` struct to `schema.WithDefaults`. Its zero `Format` overwrote the section default and caused a startup panic. The code now supplies only `output-fields` through a defaults map.
4. Updating the help DSL's canonical example changed an example test's expected output. The expected query now uses `flag:--format`.

The enclosing workspace declares Go 1.25 while this module requires Go 1.26.1. Validation therefore uses `GOWORK=off`, allowing the module toolchain directive to select Go 1.26.5.

## Review guide

Review in this order:

1. `pkg/settings/structured_output.go` — public contract and processor assembly.
2. `pkg/middlewares/row/output-fields.go` — projection semantics.
3. `pkg/formatters/json/json.go` — JSON versus compact JSONL framing.
4. `pkg/cli/cobra.go` — injection, execution, and atomic mounting.
5. `pkg/help/cmd/export.go` — `export-mode` versus serializer format.
6. `go.mod` and `go.sum` — dependency removal.
7. Deleted settings, Excel, jq, and docs — verify no accidental unrelated deletion.
8. Tests and `pkg/doc/topics/32-structured-output.md`.

Validation commands:

```bash
GOWORK=off go test ./... -count=1
GOWORK=off go build ./...
GOWORK=off make govulncheck
GOWORK=off go run ./cmd/glaze json --help
GOWORK=off go run ./cmd/glaze help structured-output --short
```

The vulnerability audit required `golang.org/x/text` v0.39.0 or newer for GO-2026-5970 and the aligned `go.opentelemetry.io/otel`, `otel/metric`, and `otel/trace` v1.42.0 modules for GO-2026-5158. `make govulncheck` now reports zero reachable vulnerabilities.

The `json --help` structured-output group must contain only:

```text
--format
--max-output-rows
--output-fields
```

## Final decisions

- Keep three universal output flags, not one and not 44.
- Keep six serializers.
- Treat JSONL as the streaming format.
- Keep output projection exact and intentionally simple.
- Treat max rows as an output cap, not source cancellation.
- Delete old flags and their dedicated documentation rather than preserve migration machinery.
- Delete Excel and embedded jq dependencies.
- Keep reusable non-automatic formatter and middleware libraries.
- Add no compatibility aliases.
