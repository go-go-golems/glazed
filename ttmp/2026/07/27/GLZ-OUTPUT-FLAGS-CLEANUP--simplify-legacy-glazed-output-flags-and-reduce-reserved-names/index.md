---
Title: Simplify Legacy Glazed Output Flags and Reduce Reserved Names
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
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/600
Summary: Implemented replacement of 44 legacy output and transformation flags with format, output-fields, and max-output-rows; removed Excelize, embedded jq, and old-flag documentation.
LastUpdated: 2026-07-27T12:45:59.605794944-04:00
WhatFor: Track the implementation, evidence, documentation, and review path for the structured-output cleanup.
WhenToUse: Start here when reviewing or maintaining GLZ-OUTPUT-FLAGS-CLEANUP.
---


# Simplify Legacy Glazed Output Flags and Reduce Reserved Names

## Result

Every `cmds.GlazeCommand` now receives three structured-output flags:

```text
--format table|json|jsonl|csv|tsv|yaml
--output-fields field1,field2,...
--max-output-rows N
```

The old 44-flag aggregate, settings YAML, settings adapters, Excel formatter, embedded jq middleware, and dedicated legacy documentation were deleted. No compatibility aliases were added.

## Key documents

- [Implementation guide and review report](design-doc/01-legacy-output-flag-cleanup-analysis-design-and-implementation-guide.md)
- [Investigation and implementation diary](reference/01-investigation-diary.md)
- [Tasks](tasks.md)
- [Changelog](changelog.md)

## Primary implementation files

- `pkg/settings/structured_output.go`
- `pkg/middlewares/row/output-fields.go`
- `pkg/formatters/json/json.go`
- `pkg/cli/cobra.go`
- `pkg/cli/helpers.go`
- `pkg/doc/topics/32-structured-output.md`

## Behavior

- `--format` chooses one of six serializers.
- `--output-fields` projects exact fields and preserves requested order for tabular formats.
- `--max-output-rows` caps rows reaching serialization; zero is unlimited.
- Application-specific source filters, pagination, sorting, and projections remain application fields.
- Caller-side tools such as `jq` handle ad hoc transformations.

## Validation

Use the module toolchain outside the enclosing Go 1.25 workspace:

```bash
GOWORK=off go test ./... -count=1
GOWORK=off go build ./...
GOWORK=off go run ./cmd/glaze json --help
GOWORK=off go run ./cmd/glaze help structured-output --short
```
