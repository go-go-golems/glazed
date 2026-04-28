---
Title: Serve External Help Sources
Slug: serve-external-help-sources
Short: Use `glaze serve` to browse help exported from other Glazed binaries, JSON files, SQLite snapshots, and markdown directories.
Topics:
- help
- serve
- export
- web
- documentation
- sqlite
Commands:
- serve
- help
- export
Flags:
- from-glazed-cmd
- from-json
- from-sqlite
- with-embedded
- address
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why external help sources exist

`glaze serve` can act as a browser for more than the documentation embedded in the `glaze` binary. Many Glazed-based tools expose their own help pages through `help export`; external source loading lets you collect those pages into one local web server.

This is useful when you maintain a family of command-line tools. Instead of running one help browser per binary, run one server that loads help from all of them:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

By default, explicit external sources replace the embedded Glazed documentation. This keeps the browser focused on the tools you asked for. If you also want the built-in Glazed docs, add `--with-embedded=true`.

## Quick start: serve another Glazed binary

Use `--from-glazed-cmd` when each value is the name or path of a binary that supports `help export`.

```bash
glaze serve --from-glazed-cmd pinocchio
```

For each binary, `glaze serve` runs:

```bash
<binary> help export --with-content=true --output json
```

Then it imports the JSON into the in-memory help store and starts the existing help browser.

You can load multiple binaries at once:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

Or use repeated flags:

```bash
glaze serve \
  --from-glazed-cmd pinocchio \
  --from-glazed-cmd sqleton \
  --from-glazed-cmd xxx
```

## Source types

`glaze serve` can load several kinds of sources. All string source flags are list-valued, so you can pass multiple sources together.

| Source | Flag | Example |
|--------|------|---------|
| Glazed binary shorthand | `--from-glazed-cmd` | `--from-glazed-cmd pinocchio,sqleton` |
| JSON export file | `--from-json` | `--from-json ./pinocchio-help.json` |
| SQLite export database | `--from-sqlite` | `--from-sqlite ./pinocchio-help.db` |
| Markdown files/directories | positional paths | `glaze serve ./docs ./more-docs` |

Use `--from-glazed-cmd` for live binaries and `--from-json` or `--from-sqlite` when you need a filtered or archived snapshot.

## Embedded documentation behavior

When you run `glaze serve` with no sources, it serves the built-in Glazed documentation:

```bash
glaze serve
```

When you provide any explicit source, embedded docs are cleared by default:

```bash
glaze serve --from-glazed-cmd pinocchio
```

This serves Pinocchio help only. To merge Pinocchio help with the embedded Glazed docs, set:

```bash
glaze serve --with-embedded=true --from-glazed-cmd pinocchio
```

## Serve from exported JSON

JSON files are useful when you want a cached snapshot or when the source binary is not available on the serving machine.

```bash
pinocchio help export --output json > /tmp/pinocchio-help.json
glaze serve --from-json /tmp/pinocchio-help.json
```

You can also read JSON from stdin:

```bash
pinocchio help export --output json | glaze serve --from-json -
```

Only one `--from-json -` source is allowed because stdin can only be read once.

## Serve from SQLite

SQLite files are useful for archived snapshots and tooling that wants a queryable database.

```bash
pinocchio help export --format sqlite --output-path ./pinocchio-help.db
glaze serve --from-sqlite ./pinocchio-help.db
```

You can combine several snapshots:

```bash
glaze serve --from-sqlite ./pinocchio.db,./sqleton.db
```

## Combine all source types

You can build a unified help browser by combining live binaries, snapshots, and local markdown overrides:

```bash
glaze serve \
  --from-glazed-cmd pinocchio,sqleton \
  --from-json ./team-overrides.json \
  --from-sqlite ./legacy-help.db \
  ./company-docs
```

If two sources contain the same slug, the later source wins. The loading order is markdown paths, JSON files, SQLite files, then `--from-glazed-cmd` binaries.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `executable file not found` | A `--from-glazed-cmd` binary is not on `PATH` | Use an absolute path or install the binary. |
| Browser shows only one tool's docs | Explicit sources clear embedded docs by default | Add more sources or use `--with-embedded=true` to include built-in Glazed docs. |
| JSON import fails with missing type | The JSON file is not a Glazed help export | Generate it with `<binary> help export --output json`. |
| Stdin source hangs | The process feeding stdin did not finish | Verify the upstream command exits and writes valid JSON. |
| Duplicate pages disappear | Later sources overwrite earlier slugs | Rename slugs or change source order. |

## See Also

- `glaze help export-help-entries` — Export help sections to JSON, files, and SQLite
- `glaze help serve-help-over-http` — Serve the built-in help browser over HTTP
- `glaze help export-help-static-website` — Export help as a static website
- `glaze help help-system` — Overview of the Glazed help system
