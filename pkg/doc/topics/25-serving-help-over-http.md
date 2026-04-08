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

## Reuse the API and SPA in your own server

If you already have an HTTP server, you do not need to shell out to `glaze serve`. You can compose the help API and the shared SPA directly from Go.

The two building blocks are:

- `web.NewSPAHandler()` for the embedded frontend
- `server.NewServeHandler(...)` or `server.NewMountedHandler(...)` for API + SPA composition

### Mount at the server root

Use this when help is the whole server or when you want the help browser to own `/`.

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

    // Load your own help pages here.
    // err := hs.LoadSectionsFromFS(...)

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

### Mount under a prefix such as `/help`

Use this when you already have an HTTP server and want help to live under a sub-route.

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

    deps := helpserver.HandlerDeps{Store: hs.Store}

    mux := http.NewServeMux()
    mounted := helpserver.NewMountedHandler("/help", deps, spaHandler)
    mux.Handle("/help", mounted)
    mux.Handle("/help/", mounted)

    _ = http.ListenAndServe(":18100", mux)
}
```

With this setup:

- `/help/` serves the SPA
- `/help/api/health` serves the API
- `/help/api/sections` lists sections

## API-only usage

If you want to serve the JSON endpoints without the embedded browser UI, pass `nil` as the SPA handler to `NewServeHandler`.

```go
handler := helpserver.NewServeHandler(
    helpserver.HandlerDeps{Store: hs.Store},
    nil,
)
```

This is useful when:

- your frontend is separate,
- you want a custom UI,
- or you only need a programmatic documentation API.

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
| `/api/health` works but no sections appear | The supplied markdown files were not loaded or were skipped | Verify the path exists, the files end in `.md`, and the frontmatter has fields like `Title`, `Slug`, and `SectionType` |
| `glaze serve` exits immediately | Invalid path or startup error | Re-run with valid file/directory arguments and inspect the error output |
| Mounted `/help` route returns 404 | Prefix mounting was not wired correctly in the outer mux | Register both `/help` and `/help/`, and use `server.NewMountedHandler("/help", ...)` |
| API calls work but browser routes fail | SPA handler was not mounted | Use `web.NewSPAHandler()` and pass it into `server.NewServeHandler(...)` or `server.NewMountedHandler(...)` |

## See Also

- `glaze help help-system`
- `glaze help writing-help-entries`
- `glaze help how-to-write-good-documentation-pages`
- `glaze help sections-guide`
