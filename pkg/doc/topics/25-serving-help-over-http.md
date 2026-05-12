---
Title: Serve Help Over HTTP
Slug: serve-help-over-http
Short: Use `glaze serve` to browse help pages in a browser, and reuse the help API and SPA handlers in your own HTTP server.
Topics:
- help
- http
- serve
- api
- documentation
Commands:
- serve
- help
Flags:
- address
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why `glaze serve` exists

`glaze serve` turns Glazed help pages into a browsable web application. Instead of only reading help in the terminal, you can load markdown help entries into a help store, expose a JSON API for search and filtering, and serve the embedded SPA that renders those entries in a browser.

This matters in two situations:

- you want a quick local browser for an existing help tree such as `docs/` or `pkg/doc/`, or
- you are embedding Glazed help into a larger application and want to reuse the API and SPA handlers under your own HTTP routes.

The command is intentionally built on the same primitives you can use from Go code. That means the CLI path and the programmatic path stay aligned.

## Use `glaze serve` from the command line

The simplest way to use the feature is to run it with no arguments — it will
load the built-in Glazed documentation:

```bash
glaze serve
```

You can also point it at one or more markdown files or directories:

```bash
glaze serve ./pkg/doc
```

You can also serve multiple paths at once:

```bash
glaze serve ./pkg/doc ./extras/help
```

If you need a different port, use `--address`:

```bash
glaze serve --address :18100 ./pkg/doc
```

When the server starts:

- If no paths are given, the built-in Glazed documentation (loaded at startup via `doc.AddDocToHelpSystem`) is used.
- If paths are given, Glazed walks the supplied directories recursively.
- It loads `.md` files with Glazed help frontmatter into the help store.
- It exposes the help API under `/api/...`.
- It serves the embedded SPA for browser routes such as `/`.

## API endpoints exposed by `glaze serve`

The browser UI talks to a small HTTP API. You can call the same endpoints yourself.

### `GET /api/health`

Returns a simple health response with the number of loaded sections.

```bash
curl http://localhost:18100/api/health
```

Example response:

```json
{"ok":true,"sections":65}
```

### `GET /api/sections`

Returns section summaries.

```bash
curl http://localhost:18100/api/sections
```

You can filter using query parameters:

- `type=GeneralTopic|Example|Application|Tutorial`
- `topic=...`
- `command=...`
- `flag=...`
- `q=...` for text search
- `limit=...`
- `offset=...`

Example:

```bash
curl 'http://localhost:18100/api/sections?type=Example&topic=help'
```

### `GET /api/sections/{slug}`

Returns a full section, including markdown content.

```bash
curl http://localhost:18100/api/sections/help-system
```

## Embed the help API in your own server

If you already have an HTTP server, you can embed the Glazed help JSON API directly into your Go application. You do not need to shell out to `glaze serve`.

The building block is:

- `server.NewServeHandler(...)` or `server.NewMountedHandler(...)` for the REST API

### API-only mode (recommended for external consumers)

Pass `nil` as the SPA handler to serve only the JSON endpoints. This is the recommended approach for binaries that depend on glazed as a library, since the embedded browser UI assets are only available when building from within the glazed repository.

```go
package main

import (
    "net/http"

    "github.com/go-go-golems/glazed/pkg/help"
    helpserver "github.com/go-go-golems/glazed/pkg/help/server"
)

func main() {
    hs := help.NewHelpSystem()

    // Load your own help pages here.
    // err := hs.LoadSectionsFromFS(...)

    // nil = API-only mode, no browser UI.
    handler := helpserver.NewServeHandler(
        helpserver.HandlerDeps{Store: hs.Store},
        nil,
    )

    _ = http.ListenAndServe(":18100", handler)
}
```

This serves the full JSON API (`/api/health`, `/api/sections`, `/api/packages`) without the browser UI. You can build your own frontend, or use `glaze serve --from-glazed-cmd` to browse help from multiple tools in a single web interface.

### With the browser UI (in-repo only)

If you are building within the glazed repository, you can include the embedded React SPA:

```go
package main

import (
    "net/http"

    "github.com/go-go-golems/glazed/pkg/help"
    helpserver "github.com/go-go-golems/glazed/pkg/help/server"
    "github.com/go-go-golems/glazed/pkg/web"
)

func main() {
    hs := help.NewHelpSystem()

    spaHandler, err := web.NewSPAHandler()
    if err != nil {
        panic(err)
    }

    handler := helpserver.NewServeHandler(
        helpserver.HandlerDeps{Store: hs.Store},
        spaHandler,
    )

    _ = http.ListenAndServe(":18100", handler)
}
```

**Note:** The SPA assets are built by `go generate ./pkg/web` (using Dagger/pnpm) and embedded via `//go:embed`. This only works when building from the glazed repository with `-tags embed`. External binaries that depend on glazed as a Go module cannot access these assets — use API-only mode or `glaze serve --from-glazed-cmd` instead.

### Mount under a prefix such as `/help`

Use `NewMountedHandler` to serve the help API under a sub-route of an existing HTTP server:

```go
mux := http.NewServeMux()
deps := helpserver.HandlerDeps{Store: hs.Store}

// Mount API-only at /help
mounted := helpserver.NewMountedHandler("/help", deps, nil)
mux.Handle("/help", mounted)
mux.Handle("/help/", mounted)
```

This serves:
- `/help/api/health` — health check
- `/help/api/sections` — section listing
- `/help/api/packages` — package listing

### Browsing help from multiple tools

The recommended pattern for viewing help from multiple Glazed-based tools in a browser is:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton
```

This loads each tool's help data via `help export --output json` and serves it in a single browser UI. You don't need to embed the SPA in each individual binary.

## Package names and section visibility

Help sections are grouped by package name. When sections are loaded from embedded markdown files (via `LoadSectionsFromFS`), they get an empty package name. `NewServeHandler` automatically assigns these sections a default package name so that they appear correctly in the API responses.

If you are using the Store directly (without `NewServeHandler`), you can call `Store.SetDefaultPackage(ctx, "myapp", "")` after loading your sections to assign them a package name. This is required for the package filter in the help API to work correctly.

## Loading help pages correctly

The HTTP layer only serves what you load into the help system. In practice that means you should load markdown files before starting the server.

For embedded documentation packages, the usual pattern is:

```go
helpSystem := help.NewHelpSystem()
if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
    panic(err)
}
```

For ad hoc files and directories, use the CLI:

```bash
glaze serve ./pkg/doc ./more-docs
```

If a page does not appear in the browser, the most common cause is that the markdown file is missing valid Glazed help frontmatter.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Browser shows a blank page or error | The SPA assets were not generated or embedded | Run `GOWORK=off go generate ./pkg/web`, then rebuild `glaze` |
| SPA shows "assets have not been generated" | Building from outside the glazed repository | The SPA is only available when building from the glazed repo. Use API-only mode (pass `nil` as SPA handler) or `glaze serve --from-glazed-cmd` |
| SPA shows "0 sections" but `/api/sections` returns data | Sections have no package name; the SPA filters by package | `NewServeHandler` auto-assigns a default package. For direct Store usage, call `hs.Store.SetDefaultPackage(ctx, "myapp", "")` after loading docs |
| `/api/health` works but no sections appear | The supplied markdown files were not loaded or were skipped | Verify the path exists, the files end in `.md`, and the frontmatter has fields like `Title`, `Slug`, and `SectionType` |
| `glaze serve` exits immediately | Invalid path or startup error | Re-run with valid file/directory arguments and inspect the error output |
| Mounted `/help` route returns 404 | Prefix mounting was not wired correctly in the outer mux | Register both `/help` and `/help/`, and use `server.NewMountedHandler("/help", ...)` |
| API calls work but browser routes fail | SPA handler was not mounted | Use `web.NewSPAHandler()` and pass it into `server.NewServeHandler(...)` or `server.NewMountedHandler(...)` |

## See Also

- `glaze help serve-external-help-sources` — Serve help exported from other Glazed binaries and snapshots
- `glaze help export-help-entries` — Export help sections to files, JSON, CSV, or SQLite
- `glaze help export-help-static-website`
- `glaze help help-system`
- `glaze help writing-help-entries`
- `glaze help how-to-write-good-documentation-pages`
- `glaze help sections-guide`
