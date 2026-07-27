---
Title: Export help entries
Slug: export-help-entries
Short: Export loaded help sections as structured rows, Markdown files, or SQLite.
Topics:
- help
- export
- documentation
Commands:
- help
- export
Flags:
- format
- export-mode
- output-path
SectionType: GeneralTopic
IsTopLevel: false
ShowPerDefault: false
---

# Export help entries

`glaze help export` exports the sections currently loaded in the help system. It has two independent controls:

- `--export-mode glazed|files|sqlite` chooses what the command does.
- `--format table|json|jsonl|csv|tsv|yaml` chooses how rows are serialized when `--export-mode glazed` is active.

The default export mode is `glazed`.

## Structured export

```bash
# JSON with complete Markdown content
glaze help export --format json

# Metadata-only CSV
glaze help export --with-content=false --format csv

# A bounded JSONL sample
glaze help export --format jsonl --max-output-rows 25
```

The structured mode emits one row per section. Common fields include `slug`, `title`, `short`, `content`,
`section_type`, `topics`, `commands`, `flags`, `is_top_level`, `show_per_default`, and `order`.

Use `--output-fields` when only a subset is needed:

```bash
glaze help export --format jsonl \
  --output-fields slug,title,section_type \
  --with-content=false
```

## Markdown file export

```bash
# Preserve type-based subdirectories
glaze help export --export-mode files --output-path ./exported-help

# Put every file directly in the destination
glaze help export --export-mode files --output-path ./flat --flatten-dirs
```

Each result is a valid help-section Markdown file with frontmatter and can be loaded again by Glazed.

## SQLite export

```bash
glaze help export --export-mode sqlite --output-path ./help.db
sqlite3 ./help.db "SELECT slug, title, section_type FROM sections LIMIT 5;"
```

The resulting database is self-contained and uses the same section schema as the in-process help store.

## Filtering

Filters are combined with AND semantics:

```bash
# JSON examples only
glaze help export --type Example --topic json --format json

# Export one command's docs as files
glaze help export --command serve \
  --export-mode files --output-path ./serve-docs

# Select one exact slug
glaze help export --slug help-system --format yaml
```

Available filters are `--type`, `--topic`, `--command`, `--flag`, and `--slug`.

## Implementation notes

`pkg/help/cmd/export.go` implements the command as a `BareCommand` because file and SQLite modes own their
output artifacts. It mounts `settings.NewStructuredOutputSection()` explicitly for the default row-emitting
mode. The external help loader invokes binaries with `help export --format json`.

## Troubleshooting

| Problem | Resolution |
|---|---|
| No sections matched | Remove filters and run `glaze help export --format json`. |
| Exported files omit bodies | Ensure `--with-content=true`, which is the default. |
| SQLite reports `database is locked` | Close the other database user or select another `--output-path`. |
| Flat export has filename collisions | Omit `--flatten-dirs` or make section slugs unique. |

## See also

- `glaze help serve-external-help-sources`
- `glaze help help-system`
- `glaze help serve-help-over-http`
- `glaze help export-help-static-website`
