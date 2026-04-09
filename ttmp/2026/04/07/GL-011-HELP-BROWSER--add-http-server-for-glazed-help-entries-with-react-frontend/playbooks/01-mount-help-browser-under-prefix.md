---
Title: Mount Help Browser Under a Prefix
Ticket: GL-011-HELP-BROWSER
Status: active
Topics:
  - glazed
  - help
  - http
  - mount
  - prefix
DocType: playbook
Intent: how-to
Owners:
  - manuel
RelatedFiles:
  - Path: pkg/help/server/serve.go
    Note: `NewServeHandler`, `MountPrefix`, and `NewMountedHandler`
  - Path: pkg/web/static.go
    Note: `NewSPAHandler()` for the shared embedded SPA
Summary: How to mount the help browser under `/help` or another prefix in an existing HTTP server
LastUpdated: 2026-04-08T09:18:00-04:00
---

# Mount Help Browser Under a Prefix

This playbook shows how to reuse the help browser in an existing HTTP server without requiring it to live at `/`.

## Goal

Expose the help browser under a prefix such as:

- `/help`
- `/docs`
- `/internal/help`

while preserving:

- `/help/api/...` for the API
- `/help/...` for the SPA

## Building blocks

Use:

- `server.NewServeHandler(deps, spaHandler)` to build the root handler
- `server.MountPrefix("/help", h)` to adapt it for a prefixed mount
- `web.NewSPAHandler()` to create the shared SPA handler

## Example

```go
package main

import (
    "net/http"

    helpserver "github.com/go-go-golems/glazed/pkg/help/server"
    "github.com/go-go-golems/glazed/pkg/help/store"
    "github.com/go-go-golems/glazed/pkg/web"
)

func main() {
    st, err := store.NewInMemory()
    if err != nil {
        panic(err)
    }

    spaHandler, err := web.NewSPAHandler()
    if err != nil {
        panic(err)
    }

    deps := helpserver.HandlerDeps{Store: st}
    rootHelpHandler := helpserver.NewServeHandler(deps, spaHandler)

    mux := http.NewServeMux()
    mux.Handle("/help", helpserver.MountPrefix("/help", rootHelpHandler))
    mux.Handle("/help/", helpserver.MountPrefix("/help", rootHelpHandler))

    _ = http.ListenAndServe(":8080", mux)
}
```

## Behavior

With the example above:

- `GET /help/` serves the SPA
- `GET /help/api/health` serves the health endpoint
- `GET /help/api/sections` serves the sections endpoint
- `GET /api/health` does **not** match and should return 404 from the outer mux unless another handler owns it

## Notes

- `MountPrefix` expects a **root-mounted** help handler and rewrites the request path before delegation.
- `NewServeHandler` already uses `NewHandler`, which already includes CORS headers for the API.
- The SPA handler returned by `pkg/web.NewSPAHandler()` only serves `GET` and `HEAD`; other methods fall through to 404 behavior.

## Validation

After wiring the handler into your server:

```bash
curl http://localhost:8080/help/api/health
curl http://localhost:8080/help/
```

The first call should return JSON; the second should return the embedded `index.html`.
