---
Title: Distribute the Help Browser SPA to Other Go Binaries
Slug: distribute-help-browser-spa
Short: Fetch the versioned `glazed-spa` GitHub Release asset and embed it in another Glazed-based binary.
Topics:
- help
- http
- serve
- spa
- release
- embed
- documentation
Commands:
- serve
- help
Flags:
- address
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

## Why distribute the SPA separately

Glazed includes a browser-based help UI, but most downstream tools should not commit the built JavaScript bundle into their Go repositories. The recommended pattern is to publish the Glazed Help SPA as a GitHub Release asset and have downstream binaries fetch that tarball during their release build.

This keeps the Glazed repository responsible for building the frontend, while tools such as Pinocchio can embed the exact browser UI that matches their pinned `github.com/go-go-golems/glazed` module version. The downstream repository stays clean: fetched files live in an ignored `pkg/spa/dist/` directory and are only baked into binaries built with an explicit `embed` tag.

## Release asset naming

Glazed release tags use Go module versions with a leading `v`:

```text
v1.2.13
```

The SPA release asset is versioned without the leading `v`:

```text
glazed-spa-1.2.13.tar.gz
```

The download URL therefore has both forms:

```text
https://github.com/go-go-golems/glazed/releases/download/v1.2.13/glazed-spa-1.2.13.tar.gz
```

When deriving the URL from `go.mod`, keep the leading `v` for the release tag and strip it only for the filename.

## Minimal downstream layout

A downstream Go binary that wants its own `serve` command can keep the fetched SPA in a small package:

```text
pkg/spa/
  embed.go       # build tag: embed; embeds dist/ and exposes it as fs.FS root
  embed_none.go  # build tag: !embed; optional dev fallback or placeholder
  spa.go         # NewHandler() serving index.html and static assets
  dist/          # fetched release asset contents; ignored by git
```

Add the fetched directory to `.gitignore`:

```gitignore
pkg/spa/dist/
```

The important embed detail is that `//go:embed dist` preserves the `dist/` prefix. If your HTTP handler reads `index.html` at the filesystem root, expose a sub-filesystem:

```go
//go:build embed

package spa

import (
    "embed"
    "io/fs"
)

//go:embed dist
var embeddedAssets embed.FS

var Assets fs.FS = mustSub(embeddedAssets, "dist")

func mustSub(fsys fs.FS, dir string) fs.FS {
    sub, err := fs.Sub(fsys, dir)
    if err != nil {
        panic(err)
    }
    return sub
}
```

This makes embedded builds behave like `os.DirFS("pkg/spa/dist")`: `index.html` is available at the filesystem root.

## Fetch the SPA in a Makefile

Parse the Glazed dependency directly from `go.mod`. This avoids `go list -m` surprises in workspace mode.

```makefile
GLAZED_VERSION := $(shell grep 'go-go-golems/glazed ' go.mod | head -1 | awk '{print $$2}')
GLAZED_VERSION_NO_V := $(patsubst v%,%,$(GLAZED_VERSION))
GLAZED_SPA_DIR := pkg/spa/dist

fetch-spa:
	@if [ -z "$(GLAZED_VERSION)" ]; then echo "Warning: cannot detect glazed version from go.mod, skipping SPA fetch"; exit 0; fi
	@mkdir -p $(GLAZED_SPA_DIR)
	@echo "Fetching SPA assets for glazed $(GLAZED_VERSION)..."
	@curl -sfL https://github.com/go-go-golems/glazed/releases/download/$(GLAZED_VERSION)/glazed-spa-$(GLAZED_VERSION_NO_V).tar.gz \
		| tar xz -C $(GLAZED_SPA_DIR) \
	|| (echo "Warning: SPA assets not found for glazed $(GLAZED_VERSION), building without browser UI"; rm -rf $(GLAZED_SPA_DIR))

clean-spa:
	rm -rf $(GLAZED_SPA_DIR)

build-with-spa: fetch-spa
	go build -tags embed -o ./mytool ./cmd/mytool
```

For release builds, prefer making missing SPA assets fail fast before `go build -tags embed`, because `//go:embed dist` requires the directory to exist.

## Serve the SPA with the Glazed help API

Load your own help docs, create a SPA handler from the fetched assets, and pass both to `helpserver.NewServeHandler`:

```go
package main

import (
    "net/http"

    "github.com/go-go-golems/glazed/pkg/help"
    helpserver "github.com/go-go-golems/glazed/pkg/help/server"
    "github.com/my-org/mytool/pkg/doc"
    "github.com/my-org/mytool/pkg/spa"
)

func main() {
    hs := help.NewHelpSystem()
    if err := doc.AddDocToHelpSystem(hs); err != nil {
        panic(err)
    }

    spaHandler, err := spa.NewHandler()
    if err != nil {
        // You may choose to fail here for release binaries. During development,
        // API-only fallback can be useful.
        spaHandler = nil
    }

    handler := helpserver.NewServeHandler(
        helpserver.HandlerDeps{Store: hs.Store},
        spaHandler,
    )

    _ = http.ListenAndServe(":8088", handler)
}
```

`NewServeHandler` assigns a default package to loaded sections that do not already have one. This is important because the browser UI groups sections by package and otherwise may appear to show no sections.

## End-to-end smoke test

After bumping Glazed and fetching assets, the practical smoke test is:

```bash
go get github.com/go-go-golems/glazed@v1.2.13
go mod tidy
make fetch-spa build-with-spa
./mytool serve --address :18888
```

In another terminal:

```bash
curl -I http://localhost:18888/
curl -s http://localhost:18888/api/health
```

Expected results:

- `/` returns `200 OK` with `Content-Type: text/html; charset=utf-8`.
- `/api/health` returns `{"ok":true,"sections":N}` with a nonzero section count.
- The server log does not contain `SPA handler not available`.

The Pinocchio integration validated this path with Glazed `v1.2.13`: `pinocchio serve` served the browser UI and reported 53 help sections.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Download returns 404 | Asset filename used `v1.2.13` instead of `1.2.13` | Keep `v` in the tag path, strip it only for `glazed-spa-<version>.tar.gz` |
| `go build -tags embed` fails with `pattern dist: no matching files found` | `pkg/spa/dist/` was not fetched before build | Run `make fetch-spa`; make release builds fail fast if fetching fails |
| Server logs `open index.html: file does not exist` | Embedded FS still has the `dist/` prefix | Use `fs.Sub(embeddedAssets, "dist")` before serving assets |
| API works but `/` returns 404 | SPA handler was nil or failed to initialize | Check `spa.NewHandler()` errors and confirm `index.html` exists in `pkg/spa/dist` before building |
| Browser loads but shows no sections | Help docs were not loaded or sections lack a package | Load docs before starting the server; use `NewServeHandler`, which assigns a default package |
| Build unexpectedly runs frontend generation | Downstream `build-with-spa` still calls `go generate ./...` | Fetch the release asset and build the command directly with `go build -tags embed -o ./mytool ./cmd/mytool` |

## See Also

- `glaze help serve-help-over-http` — Serve the help API and browser UI over HTTP
- `glaze help serve-external-help-sources` — Browse help exported from other Glazed binaries
- `glaze help export-help-entries` — Export help sections for external tools and snapshots
- `glaze help writing-help-entries` — Author embedded help docs for your own binary
