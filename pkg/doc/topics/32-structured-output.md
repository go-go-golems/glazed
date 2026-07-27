---
Title: Structured output
Slug: structured-output
Short: Serialize GlazeCommand rows with a small, predictable three-flag output surface.
Topics:
- commands
- output
- formatters
Commands:
- json
- yaml
- csv
Flags:
- format
- output-fields
- max-output-rows
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Structured commands emit `types.Row` values and let Glazed serialize them. The universal output surface is deliberately limited to three flags so application commands retain ordinary names such as `output`, `fields`, `filter`, and `limit` for their own business logic.

## Choosing a format

`--format` accepts six values:

| Value | Result | Typical use |
|---|---|---|
| `table` | Deterministic terminal table; the default | Interactive use |
| `json` | One JSON array | Batch tools and APIs |
| `jsonl` | One compact JSON object per line | Streaming and coding agents |
| `csv` | Comma-separated table with headers | Spreadsheets and tabular tools |
| `tsv` | Tab-separated table with headers | Shell pipelines |
| `yaml` | One YAML sequence | Human-readable structured data |

```bash
glaze json records.json --format json
glaze json records.json --format jsonl
glaze json records.json --format csv > records.csv
```

JSONL is the streaming contract. There is no separate stream switch or object-framing toggle.

## Projecting output fields

`--output-fields` keeps the named fields. Tabular formats preserve the requested column order; JSON object key order is not a wire-level contract. Missing fields are omitted, and an empty list preserves every field.

```bash
glaze json records.json \
  --format jsonl \
  --output-fields id,name,status
```

Projection happens after format-required normalization such as CSV flattening and before serialization. It changes only emitted rows; it does not ask an upstream API or database to fetch fewer fields. Define an application field when projection must affect source work.

## Capping serialized rows

`--max-output-rows` prevents more than the requested number of rows from reaching the formatter. Zero means unlimited.

```bash
glaze json large-records.json \
  --format jsonl \
  --max-output-rows 100
```

This is an output guard, not source pagination. A command may continue its underlying work after the cap is reached. Commands that can avoid remote or database work should expose their own domain-specific limit.

## Composing transformations

Glazed does not attach generic sorting, renaming, templating, jq, deduplication, or replacement flags to every command. Serialize a machine-readable format and use a focused caller-side tool:

```bash
glaze json records.json --format jsonl |
  jq -c 'select(.status == "active") | {id, name}'
```

Application flags remain appropriate when filtering, sorting, or limiting changes the operation itself rather than merely changing already-produced rows.

## Go API

`cli.BuildCobraCommand` automatically adds the section to `cmds.GlazeCommand` implementations. Raw Cobra integrations can mount it explicitly:

```go
section, err := settings.NewStructuredOutputSection()
if err != nil {
    return err
}
if err := section.AddSectionToCobraCommand(cmd); err != nil {
    return err
}
```

Programmatic execution uses `settings.SetupStructuredOutput`. Callers that need projected and capped rows without serialization can use `settings.SetupStructuredProcessor`.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `Flag 'format' already exists` | The application also declared the framework serializer name. | Rename the application mode flag; reserve `format` for serialization. |
| JSONL contains more data than expected | The command emitted wide rows. | Add `--output-fields` or transform with `jq`. |
| The command still performs work after the row cap | `--max-output-rows` caps serialization, not source execution. | Add or use a command-specific source limit. |
| CSV nested values are surprising | Tabular output requires scalar cells. | Prefer JSONL for nested data or normalize it before CSV output. |

## See also

- `glaze help commands-reference`
- `glaze help 05-build-first-command`
- `glaze help 07-dual-commands`
- `glaze help 31-glazed-cli-lint`
