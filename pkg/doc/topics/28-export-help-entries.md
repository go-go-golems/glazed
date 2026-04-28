---
Title: Export Help Entries
Slug: export-help-entries
Short: Use `glaze help export` to export help section metadata and content to JSON, CSV, files, or SQLite for backup, indexing, and external tooling.
Topics:
- help
- export
- cli
- sqlite
- json
- documentation
Commands:
- help
- export
Flags:
- with-content
- format
- output-path
- flatten-dirs
- type
- topic
- command
- flag
- slug
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why `glaze help export` exists

The Glazed help system stores documentation in an SQLite-backed store at runtime. That makes querying fast, but it also means the data is trapped inside a running process. `glaze help export` solves this by letting you dump the entire help tree—or a filtered slice of it—to formats you can use elsewhere.

This matters when you want to:

- back up or version-control your help content as plain files,
- feed help metadata into an external search index or CMS,
- generate reports on documentation coverage (which commands have examples, which topics are missing),
- or ship a SQLite database of your app's documentation alongside the binary.

The command is built on the same `HelpSystem` and `Store` primitives that `glaze serve` and `glaze render-site` use, so the exported data is always consistent with what the CLI and the web browser show.

## Basic usage

The simplest invocation exports every loaded section as JSON:

```bash
glaze help export --output json
```

By default, `--with-content` is `true`, so the JSON includes the full markdown body of each section. If you only need metadata, disable content:

```bash
glaze help export --with-content=false --output csv
```

## Export formats

The `--format` flag controls the export target:

| Format | What it produces | Best for |
|--------|-----------------|----------|
| `glazed` (default) | Tabular output via the GlazeProcessor | JSON, CSV, YAML, or table dumps to stdout |
| `files` | One `.md` file per section on disk | Backup, version control, or editing |
| `sqlite` | A standalone SQLite database | Querying with SQL, shipping with apps |

### Tabular export (`--format glazed`)

This is the default. It behaves like any other Glazed command: you can choose the output serializer with `--output`.

```bash
# Pretty-printed JSON with full content
glaze help export --output json

# Compact CSV for spreadsheet import
glaze help export --with-content=false --output csv

# YAML for human review
glaze help export --output yaml
```

The tabular mode emits one row per section with these columns:

| Column | Description |
|--------|-------------|
| `slug` | Unique section identifier |
| `title` | Display title |
| `short` | One-line description |
| `content` | Full markdown body (omitted when `--with-content=false`) |
| `section_type` | `GeneralTopic`, `Example`, `Application`, or `Tutorial` |
| `topics` | Comma-separated topic tags |
| `commands` | Comma-linked command associations |
| `flags` | Comma-linked flag associations |
| `is_top_level` | Whether the section appears in top-level listings |
| `show_per_default` | Whether the section is shown without `--all` |
| `order` | Sort order within its type group |

### File export (`--format files`)

File mode writes each section to a separate `.md` file, reconstructing the original frontmatter so the files can be re-loaded into Glazed later.

```bash
# Default: typed subdirectories
glaze help export --format files --output-path ./exported-help
```

This produces a directory tree like:

```
exported-help/
  general-topic/
    help-system.md
    sections-guide.md
  example/
    json-output.md
  application/
    real-world-pipeline.md
  tutorial/
    getting-started.md
```

If you prefer a flat layout, use `--flatten-dirs`:

```bash
glaze help export --format files --output-path ./flat --flatten-dirs
```

Which produces:

```
flat/
  help-system.md
  sections-guide.md
  json-output.md
  ...
```

Each exported file is a valid Glazed help section. You can load it back with:

```bash
glaze serve ./exported-help
```

### SQLite export (`--format sqlite`)

SQLite mode creates a new database file containing all exported sections in the same schema that Glazed uses internally.

```bash
glaze help export --format sqlite --output-path ./help.db
```

You can then query it directly:

```bash
sqlite3 ./help.db "SELECT slug, title, section_type FROM sections LIMIT 5;"
```

The exported database is fully self-contained. It does not depend on the original application binary, so you can ship it to tools that understand Glazed's schema or use it as a snapshot for offline browsing.

## Filtering exports

You rarely need to export everything. The command accepts the same metadata filters that `glaze help` uses for querying:

```bash
# Export only examples about JSON
glaze help export --type Example --topic json --output yaml

# Export sections related to a specific command
glaze help export --command serve --format files --output-path ./serve-docs

# Export by exact slug (useful for one-off backups)
glaze help export --slug help-system --output json

# Combine multiple filters (AND logic)
glaze help export --type Example --topic advanced --command json --output csv
```

Available filters:

| Flag | Matches |
|------|---------|
| `--type` | Section type: `GeneralTopic`, `Example`, `Application`, `Tutorial` |
| `--topic` | Topic tag in the section's `Topics` list |
| `--command` | Command name in the section's `Commands` list |
| `--flag` | Flag name in the section's `Flags` list |
| `--slug` | Exact slug match (repeatable for multiple slugs) |

All filters are combined with AND logic. If no filters are given, every loaded section is exported.

## Practical examples

### Back up documentation before a refactor

Before restructuring your help files, snapshot the current state:

```bash
glaze help export --format files --output-path ./docs-backup-$(date +%Y%m%d)
```

### Generate a CSV inventory for a spreadsheet

Create a lightweight inventory of all sections for editorial review:

```bash
glaze help export --with-content=false --output csv > section-inventory.csv
```

### Ship a documentation database with your app

Build a release artifact that contains both the binary and a queryable help DB:

```bash
glaze help export --format sqlite --output-path ./dist/myapp-help.db
```

### Reconstruct markdown from the internal store

If you loaded sections programmatically and want to recover the original `.md` files:

```bash
glaze help export --format files --flatten-dirs --output-path ./recovered
```

## How the export command is wired

`glaze help export` is implemented as a `BareCommand` in `pkg/help/cmd/export.go`. It adds a glazed section to its schema so that `--output json/csv/table/yaml` flags are available, but bypasses the GlazeProcessor when `--format` is `files` or `sqlite` so it can write to disk instead of stdout.

The command lives under the `help` subcommand tree. It is registered automatically when a binary calls `help_cmd.SetupCobraRootCommand(...)`. The `glaze` binary does this in `cmd/glaze/main.go`, so the export verb is available out of the box. Advanced integrations can still call `help_cmd.AddExportCommand(...)` directly when constructing a custom help command tree; the helper is idempotent and will not add a duplicate `export` child.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `Flag 'output' already exists` error | Using an older version that conflicts with glazed flags | Upgrade to the latest glazed version; the export command uses `--output-path` for file destinations. |
| Exported files are missing content | `--with-content=false` was set, or the source sections had empty `Content` fields | Re-run with `--with-content=true` (the default). |
| `no sections matched the given filters` | Filters are too restrictive, or no sections are loaded | Run `glaze help export --output json` without filters to verify sections exist. |
| SQLite export fails with `database is locked` | Another process is holding the target database open | Close the other process or choose a different `--output-path`. |
| Reconstructed markdown has reordered frontmatter fields | `yaml.v3` encodes maps in alphabetical order | This is cosmetic; the fields are semantically identical and parse correctly. |
| `flatten-dirs` causes filename collisions | Two sections have the same slug in different types | Avoid `--flatten-dirs` when exporting mixed section types, or rename slugs before export. |

## See Also

- [`glaze help serve-external-help-sources`](glaze help serve-external-help-sources) — Serve exported help from other Glazed binaries and snapshots
- [`glaze help help-system`](glaze help help-system) — Overview of the Glazed help system
- [`glaze help serve-help-over-http`](glaze help serve-help-over-http) — Browse help in a web browser
- [`glaze help export-help-static-website`](glaze help export-help-static-website) — Export help as a static site
- [`glaze help writing-help-entries`](glaze help writing-help-entries) — How to write help sections
- [`glaze help sections-guide`](glaze help sections-guide) — Deep dive into section metadata and the query DSL
