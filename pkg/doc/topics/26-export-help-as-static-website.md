---
Title: Export Help as a Static Website
Slug: export-help-static-website
Short: Use `glaze render-site` to export the Glazed help browser as a static site that can be previewed locally or hosted without a Go server.
Topics:
- help
- http
- web
- static
- documentation
Commands:
- render-site
- help
Flags:
- output-dir
- overwrite
- base-path
- data-dir
- site-title
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why `glaze render-site` exists

`glaze render-site` exports the same help browser used by `glaze serve`, but writes everything to disk instead of starting a live HTTP server. The exported site contains the SPA assets, a runtime config file, and a static JSON snapshot of the loaded help sections.

This matters when you want browser-based help without keeping a Go process running. It is useful for publishing documentation to a static host, attaching generated docs to a release artifact, or previewing a frozen documentation snapshot during review.

## Use `glaze render-site` from the command line

The simplest invocation exports the built-in Glazed documentation into `./glaze-site`:

```bash
glaze render-site
```

You can also export a custom help tree from one or more markdown files or directories:

```bash
glaze render-site ./pkg/doc
```

```bash
glaze render-site ./pkg/doc ./more-docs
```

If you want to choose the destination directory explicitly, use `--output-dir`:

```bash
glaze render-site ./pkg/doc --output-dir /tmp/glaze-doc-site
```

When the command runs:

- If no paths are given, it exports the help pages already loaded into the help system.
- If paths are given, it clears the preloaded sections and exports only the sections discovered from those explicit paths.
- It copies the embedded frontend into the output directory.
- It writes a `site-config.js` file and a `site-data/` JSON tree the SPA can browse without `/api/...`.

## Important flags

These flags control where the site is written and how the generated URLs behave:

| Flag | What it does | When to use it |
| --- | --- | --- |
| `--output-dir` | Chooses the export directory | Use it when you do not want the default `./glaze-site` path |
| `--overwrite` | Allows reusing a non-empty output directory | Use it when regenerating an existing export |
| `--base-path` | Writes a base path into the runtime config | Use it when the site will be hosted under a prefix such as `/docs` |
| `--data-dir` | Renames the relative directory that stores exported JSON | Use it only if you have a hosting constraint or want a different layout |
| `--site-title` | Sets the title written into the runtime config | Use it when exporting docs for another application or branded site |

## What the export contains

The exported site is a normal directory tree. The exact asset filenames under `assets/` are content-hashed and may change between builds, but the stable files and directories look like this:

| Path | Purpose |
| --- | --- |
| `index.html` | SPA entrypoint |
| `site-config.js` | Runtime config that tells the frontend to run in static mode |
| `assets/` | Bundled JS and CSS assets for the embedded frontend |
| `site-data/health.json` | Health metadata with the number of exported sections |
| `site-data/sections.json` | Full list of section summaries used for the sidebar |
| `site-data/sections/<slug>.json` | Per-section detail payloads |
| `site-data/indexes/topics.json` | Topic-to-slug lookup index |
| `site-data/indexes/commands.json` | Command-to-slug lookup index |
| `site-data/indexes/flags.json` | Flag-to-slug lookup index |
| `site-data/indexes/top-level.json` | Slugs marked as top-level |
| `site-data/indexes/defaults.json` | Slugs shown by default |
| `site-data/manifest.json` | Build metadata for the exported snapshot |

## Preview the exported site locally

The exported output is static, so you can preview it with any simple file server. A quick local check looks like this:

```bash
glaze render-site ./pkg/doc --output-dir /tmp/glaze-doc-site --overwrite
python3 -m http.server 8123 --directory /tmp/glaze-doc-site
```

Then open:

```text
http://127.0.0.1:8123/
```

Because the frontend uses hash routes, links such as `#/sections/help-system` work on simple static hosting without server-side route rewrites.

## Host the exported site

The generated output can be copied to any static host that serves plain files. In practice that includes:

- local file servers used during development,
- Nginx or Caddy serving a directory,
- static hosting services such as GitHub Pages,
- artifact previews attached to CI or release pipelines.

If you are hosting under a sub-path such as `/docs`, export with `--base-path /docs` so the runtime config points the SPA at the right static JSON directory.

```bash
glaze render-site ./pkg/doc --output-dir ./dist/docs --base-path /docs --overwrite
```

## How this differs from `glaze serve`

`glaze serve` and `glaze render-site` use the same help content and the same frontend, but they solve different delivery problems:

| Command | Best for | Runtime model |
| --- | --- | --- |
| `glaze serve` | Local exploration, embedding into a live Go server, dynamic API access | Starts an HTTP server and serves `/api/...` plus the SPA |
| `glaze render-site` | Publishing or sharing a frozen documentation snapshot | Writes files to disk and serves no live API |

Choose `serve` when you want a running server process. Choose `render-site` when you want an artifact you can copy, host, or review later.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `output directory "...\" is not empty` | The destination already contains files and `--overwrite` was not set | Re-run with `--overwrite` or choose a fresh output directory |
| The export completes but expected pages are missing | The supplied markdown files were not loaded or did not contain valid help frontmatter | Verify the paths exist, files end in `.md`, and the frontmatter includes fields like `Title`, `Slug`, and `SectionType` |
| The browser loads `index.html` but sections do not appear | The site is being opened incorrectly or the exported JSON tree is missing | Serve the directory over HTTP and confirm `site-config.js` and `site-data/sections.json` exist |
| Links work at `/` but not under `/docs` or another prefix | The static host path and exported base path do not match | Re-export with `--base-path` set to the hosted prefix |
| A regenerated site still shows old content | The old output directory contents were left in place | Re-run with `--overwrite` to clear the export directory before writing the new snapshot |

## See Also

- `glaze help serve-help-over-http`
- `glaze help writing-help-entries`
- `glaze help how-to-write-good-documentation-pages`
- `glaze help sections-guide`
