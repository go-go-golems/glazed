---
title: Implementation Guide: SPA Release Asset and Help Serve in Pinocchio
doc_type: design-doc
status: active
intent: long-term
topics:
  - help
  - serve
  - http
  - spa
  - release
  - goreleaser
  - pinocchio
  - distribution
owners:
  - manuel
ticket: GLZ-SPA-RELEASE
related_tickets:
  - GLZ-571-FIX-SERVE-HTTP-DOCS
created: "2026-05-12"
---

# Implementation Guide: SPA Release Asset and Help Serve in Pinocchio

## Executive Summary

This guide covers two related changes:

1. **In glazed**: Modify the release pipeline to publish the built SPA frontend as a tarball attached to the GitHub Release, so other Go binaries can download and embed it.
2. **In pinocchio**: Add a `pinocchio serve` command that embeds the glazed SPA from the release asset, loads pinocchio's help docs, and serves a browsable documentation web UI.

The key insight is that `glaze serve --from-glazed-cmd pinocchio` already works for multi-tool help browsing, but `pinocchio serve` would let pinocchio be self-contained — users don't need `glaze` installed to browse pinocchio's docs in a browser.

---

## Part 1: How the Current SPA Build Works

### The build chain in glazed

```
pkg/web/gen.go
  //go:generate go run ../../cmd/build-web

cmd/build-web/main.go
  1. Find repo root (walk up to go.mod)
  2. Use Dagger (or local pnpm) to build web/
  3. Copy web/dist/ → pkg/web/embed/public/

pkg/web/embed.go  (build tag: embed)
  //go:embed embed/public
  var embeddedFS embed.FS
  var PublicFS = mustSubFS(embeddedFS, "embed/public")

pkg/web/embed_none.go  (build tag: !embed)
  var PublicFS = findPublicFS()  // disk fallback → placeholder

pkg/web/static.go
  NewSPAHandler() → serves from PublicFS
```

### The release pipeline in glazed

**File:** `.github/workflows/release.yaml`

```
3 jobs: goreleaser-linux → goreleaser-darwin → goreleaser-merge

goreleaser-linux/goreleaser-darwin:
  - checkout
  - setup go
  - goreleaser release --clean --split
    → runs before: hooks from .goreleaser.yaml
      - go mod tidy
      - go generate ./...  ← builds the SPA via Dagger/pnpm
    → builds glaze binary with -tags embed,fts5

goreleaser-merge:
  - download artifacts from linux + darwin
  - merge dist/
  - sign checksums
  - goreleaser continue --merge
    → creates GitHub Release with:
      - glaze binary (linux amd64, linux arm64, darwin amd64, darwin arm64)
      - checksums.txt
      - .deb, .rpm packages
```

The SPA is already built during `go generate ./...` in the `before: hooks` of GoReleaser. The built files sit in `pkg/web/embed/public/` during the CI run. We just need to tar them up and attach them to the release.

---

## Part 2: Changes to the Glazed Repo

### 2.1 Attach the SPA tarball in a split/merge GoReleaser release

**Important split-release detail:** This repository uses GoReleaser's split/merge flow. The linux and darwin jobs run `goreleaser release --clean --split`, upload only `dist/`, and the merge job runs `goreleaser continue --merge` from a fresh checkout plus downloaded `dist` artifacts. GoReleaser `before.hooks` do **not** rerun during `continue --merge`, while `release.extra_files` is evaluated in the merge/publish phase.

Therefore, do **not** create `./glazed-spa.tar.gz` in `.goreleaser.yaml` `before.hooks`: it would exist only in the linux/darwin split-job filesystem and would be missing in the merge job. Instead:

1. Keep `go generate ./...` in `.goreleaser.yaml` so the split jobs can embed SPA assets into the `glaze` binaries.
2. Add `release.extra_files` in `.goreleaser.yaml` to publish `./glazed-spa.tar.gz`.
3. Rebuild/package the SPA tarball in `.github/workflows/release.yaml` in the `goreleaser-merge` job immediately before `goreleaser continue --merge`.

**`.goreleaser.yaml`:**

```yaml
before:
  hooks:
    - go mod tidy
    - go generate ./...

release:
  extra_files:
    - glob: ./glazed-spa.tar.gz
      name_template: glazed-spa-{{ .Version }}.tar.gz
```

**`.github/workflows/release.yaml` (inside `goreleaser-merge`, before `goreleaser continue --merge`):**

```yaml
      - name: Build SPA release asset
        run: |
          go generate ./pkg/web
          tar czf glazed-spa.tar.gz -C pkg/web/embed/public .
```

This guarantees the tarball exists in the exact job where GoReleaser evaluates `release.extra_files` and publishes release assets.

### 2.2 What the release will look like after the change

```
GitHub Release v1.2.8:
  glaze_1.2.8_linux_amd64.tar.gz
  glaze_1.2.8_linux_arm64.tar.gz
  glaze_1.2.8_darwin_amd64.tar.gz
  glaze_1.2.8_darwin_arm64.tar.gz
  checksums.txt
  glazed-spa.tar.gz          ← NEW: SPA frontend bundle (~300KB)
  glaze_1.2.8_amd64.deb
  glaze_1.2.8_arm64.deb
  ...
```

### 2.3 The download URL pattern

```
https://github.com/go-go-golems/glazed/releases/download/v1.2.8/glazed-spa.tar.gz
```

This is stable, versioned, and fetchable with curl in any CI environment.

---

## Part 3: Changes to the Pinocchio Repo

### 3.1 Add a Makefile target to fetch the SPA

**File:** `Makefile`

```makefile
# Add to existing targets
GLAZED_VERSION := $(shell go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null | sed 's/^v//')
GLAZED_SPA_DIR := pkg/spa/dist

.PHONY: fetch-spa clean-spa

fetch-spa:
	@if [ -z "$(GLAZED_VERSION)" ]; then echo "Error: cannot detect glazed version from go.mod"; exit 1; fi
	mkdir -p $(GLAZED_SPA_DIR)
	curl -sL https://github.com/go-go-golems/glazed/releases/download/v$(GLAZED_VERSION)/glazed-spa.tar.gz \
		| tar xz -C $(GLAZED_SPA_DIR)
	@echo "SPA assets fetched for glazed v$(GLAZED_VERSION) → $(GLAZED_SPA_DIR)"

clean-spa:
	rm -rf $(GLAZED_SPA_DIR)
```

**How it works:**

- `go list -m -f '{{.Version}}' github.com/go-go-golems/glazed` reads the glazed version from `go.mod`. This returns something like `v1.2.8`.
- `curl` downloads the tarball from the GitHub release.
- `tar xz` extracts into `pkg/spa/dist/`.
- The `pkg/spa/dist/` directory is added to `.gitignore`.

**Add to `.gitignore`:**

```
pkg/spa/dist/
```

### 3.2 Create the SPA embed package

**File:** `pkg/spa/embed.go`

```go
//go:build embed

package spa

import "embed"

// Assets contains the Glazed help browser SPA frontend files.
// Built from the glazed-spa release artifact, fetched by `make fetch-spa`.
//
//go:embed dist
var Assets embed.FS
```

**File:** `pkg/spa/embed_none.go`

```go
//go:build !embed

package spa

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing/fstest"
)

// Assets serves the SPA frontend when the binary was built with -tags embed.
// Without the tag, it falls back to a placeholder.
var Assets fs.FS = findAssets()

func findAssets() fs.FS {
	// Try to find assets on disk (dev builds)
	wd, _ := os.Getwd()
	for dir := wd; dir != ""; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "pkg", "spa", "dist", "index.html")
		if _, err := os.Stat(candidate); err == nil {
			return os.DirFS(filepath.Join(dir, "pkg", "spa", "dist"))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	// Placeholder for builds without SPA assets
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head><meta charset="UTF-8" /><title>Pinocchio Help Browser</title></head>
  <body><div id="root">
    Pinocchio help browser assets not found.
    Run <code>make fetch-spa</code> and rebuild with <code>-tags embed</code>.
  </div></body>
</html>`))},
	}
}
```

**File:** `pkg/spa/static.go`

```go
package spa

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewSPAHandler returns an http.Handler that serves the Glazed help browser SPA
// from the embedded assets.
func NewSPAHandler() (http.Handler, error) {
	indexBytes, err := fs.ReadFile(Assets, "index.html")
	if err != nil {
		return nil, fmt.Errorf("reading SPA index.html: %w", err)
	}

	fileServer := http.FileServer(http.FS(Assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		cleanPath := path.Clean("/" + r.URL.Path)
		if cleanPath == "/" {
			serveSPAIndex(w, r, indexBytes)
			return
		}

		assetPath := strings.TrimPrefix(cleanPath, "/")
		if _, err := fs.Stat(Assets, assetPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for all unknown paths (client-side routing)
		serveSPAIndex(w, r, indexBytes)
	}), nil
}

func serveSPAIndex(w http.ResponseWriter, r *http.Request, indexBytes []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(indexBytes)
	}
}
```

Wait — actually this is duplicating the SPA serving logic from `glazed/pkg/web/static.go`. Better to reuse glazed's `NewSPAHandler` but with a different `fs.FS`:

Actually, `glazed/pkg/web/static.go` currently hardcodes `PublicFS` as a package-level variable. We'd need to either:
1. Add a `NewSPAHandlerFromFS(fs.FS)` function to glazed (clean, reusable)
2. Duplicate the small SPA handler in pinocchio (quick, no glazed changes needed now)

For now, option 2 is simpler and the handler is ~30 lines. We can refactor into glazed later.

**Simplified `pkg/spa/spa.go`:**

```go
package spa

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewHandler returns an http.Handler that serves the Glazed help browser SPA.
func NewHandler() (http.Handler, error) {
	indexBytes, err := fs.ReadFile(Assets, "index.html")
	if err != nil {
		return nil, fmt.Errorf("reading SPA assets: %w (run 'make fetch-spa' and rebuild with -tags embed)", err)
	}

	fileServer := http.FileServer(http.FS(Assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		cleanPath := path.Clean("/" + r.URL.Path)
		if cleanPath == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexBytes)
			return
		}
		assetPath := strings.TrimPrefix(cleanPath, "/")
		if _, err := fs.Stat(Assets, assetPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback for client-side routing
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexBytes)
	}), nil
}
```

### 3.3 Add the serve command to pinocchio

**File:** `cmd/pinocchio/cmds/serve.go`

```go
package cmds

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/help"
	helpserver "github.com/go-go-golems/glazed/pkg/help/server"
	"github.com/go-go-golems/pinocchio/pkg/spa"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve pinocchio help documentation as a web application",
		Long: `Start an HTTP server that serves pinocchio's help documentation
as a browsable web application with a React SPA frontend.

The server exposes:
  GET /api/*   — JSON API for section listing and retrieval
  GET /*       — React SPA (browser UI)

Use --address to change the listen address (default :8088).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(address)
		},
	}

	cmd.Flags().StringVar(&address, "address", ":8088", "Address to listen on")

	return cmd
}

func runServe(address string) error {
	// Create help system and load pinocchio's documentation.
	// This mirrors what cmd/pinocchio/main.go does in initRootCmd().
	hs := help.NewHelpSystem()

	// Load docs from all the same sources as main.go.
	// We import the doc packages and call their AddDocToHelpSystem.
	// These are already wired in main.go; we duplicate the calls here
	// because the serve command runs as a standalone command.
	if err := loadAllDocs(hs); err != nil {
		return fmt.Errorf("loading help docs: %w", err)
	}

	// Create the SPA handler from the embedded (or fetched) assets.
	spaHandler, err := spa.NewHandler()
	if err != nil {
		log.Warn().Err(err).Msg("SPA handler not available, serving API only")
		spaHandler = nil
	}

	// Create the combined handler (API + SPA).
	// NewServeHandler auto-assigns a default package name to sections
	// loaded without one (fixes issue #571).
	deps := helpserver.HandlerDeps{Store: hs.Store}
	handler := helpserver.NewServeHandler(deps, spaHandler)

	// Start the server with graceful shutdown.
	httpSrv := &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info().Str("address", address).Msg("Pinocchio help browser listening")

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigCh:
		log.Info().Str("signal", sig.String()).Msg("Shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(ctx)
	}
}
```

Wait, we need to load the same docs as `main.go`. Let me check what gets loaded:

```go
// From cmd/pinocchio/main.go initRootCmd():
doc.AddDocToHelpSystem(helpSystem)              // geppetto docs
pkg_doc.AddDocToHelpSystem(helpSystem)          // pinocchio/pkg/doc
catter_doc.AddDocToHelpSystem(helpSystem)       // catter docs
pinocchio_docs.AddDocToHelpSystem(helpSystem)   // cmd/pinocchio/doc
addWorkspaceSessionstreamDocs(helpSystem)       // sessionstream docs (optional)
```

We should extract a helper function that both `main.go` and `serve.go` can call:

**File:** `cmd/pinocchio/doc/docs.go` (already exists, check it): 

Actually, the simplest approach: create a shared function in the `cmds` package that both `main.go` and `serve.go` call:

**File:** `cmd/pinocchio/cmds/help_loader.go`

```go
package cmds

import (
	"os"

	"github.com/go-go-golems/geppetto/pkg/doc"
	pinocchio_docs "github.com/go-go-golems/pinocchio/cmd/pinocchio/doc"
	"github.com/go-go-golems/pinocchio/cmd/pinocchio/cmds/catter/pkg/doc as catter_doc"
	pkg_doc "github.com/go-go-golems/pinocchio/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
)

// LoadAllHelpDocs loads all help documentation into the given HelpSystem.
// This is the same set of docs loaded by initRootCmd() in main.go.
func LoadAllHelpDocs(hs *help.HelpSystem) error {
	if err := doc.AddDocToHelpSystem(hs); err != nil {
		return err
	}
	if err := pkg_doc.AddDocToHelpSystem(hs); err != nil {
		return err
	}
	if err := catter_doc.AddDocToHelpSystem(hs); err != nil {
		return err
	}
	if err := pinocchio_docs.AddDocToHelpSystem(hs); err != nil {
		return err
	}

	// Optional: load sessionstream docs if available
	for _, candidate := range []string{
		"../sessionstream/pkg/doc",
		"../../sessionstream/pkg/doc",
	} {
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}
		if err := hs.LoadSectionsFromFS(os.DirFS(candidate), "."); err != nil {
			return err
		}
		break
	}

	return nil
}
```

Then the serve command becomes cleaner:

```go
func runServe(address string) error {
	hs := help.NewHelpSystem()
	if err := cmds.LoadAllHelpDocs(hs); err != nil {
		return fmt.Errorf("loading help docs: %w", err)
	}

	spaHandler, err := spa.NewHandler()
	if err != nil {
		log.Warn().Err(err).Msg("SPA unavailable, serving API only")
		spaHandler = nil
	}

	handler := helpserver.NewServeHandler(
		helpserver.HandlerDeps{Store: hs.Store},
		spaHandler,
	)

	// ... server startup (same as above)
}
```

### 3.4 Wire the serve command into main.go

**File:** `cmd/pinocchio/main.go` — add after `rootCmd.AddCommand(runCommandCmd)`:

```go
// Add the serve command (help browser with embedded SPA)
rootCmd.AddCommand(pinocchio_cmds.NewServeCommand())
```

This should go in the `else` branch of the `if len(os.Args) >= 3 && os.Args[1] == "run-command"` check, or better yet, register it in `initAllCommands`.

Actually, looking at the code more carefully, the simplest place is in `initAllCommands`:

```go
func initAllCommands(helpSystem *help.HelpSystem) error {
	// ... existing code ...

	// Add serve command
	rootCmd.AddCommand(pinocchio_cmds.NewServeCommand())

	return nil
}
```

### 3.5 Update GoReleaser and Makefile for pinocchio

**File:** `Makefile` — add fetch-spa to the build dependency:

```makefile
build:
	make fetch-spa
	go generate ./...
	go build -tags embed ./...
```

Wait, `go generate ./...` in pinocchio doesn't build a frontend — it generates other things (proto, etc.). The SPA fetch is separate. Let me be more precise:

```makefile
build:
	go generate ./...
	go build ./...

build-with-spa: fetch-spa
	go generate ./...
	go build -tags embed ./...
```

**File:** `.goreleaser.yaml` — add the SPA fetch step and embed tag:

```yaml
before:
  hooks:
    - make fetch-spa
    - go generate ./...

builds:
  - id: pinocchio-linux
    # ... existing config ...
    tags:
      - embed
  - id: pinocchio-darwin
    # ... existing config ...
    tags:
      - embed
```

But wait — `make fetch-spa` needs to know the glazed version, and in GoReleaser CI, `go list -m` might not work before `go mod download`. Let me check...

Actually, `go list -m` reads from `go.mod` which is in the repo, so it works fine in CI before the build. The GoReleaser CI already runs `go generate ./...` which implies Go is set up and `go.mod` is available.

But there's a subtlety: `GLAZED_VERSION` in the Makefile uses `go list -m` which returns the version from `go.mod`. For a release build, this is the correct version — it's the version pinocchio was built and tested against.

### 3.6 Handle the case where fetch-spa fails

If the glazed release doesn't have the SPA asset yet (e.g., pinocchio pins to an older glazed version), the `make fetch-spa` will fail. We should make it non-fatal:

```makefile
fetch-spa:
	@if [ -z "$(GLAZED_VERSION)" ]; then \
		echo "Warning: cannot detect glazed version, skipping SPA fetch"; \
		exit 0; \
	fi
	mkdir -p $(GLAZED_SPA_DIR)
	curl -sfL https://github.com/go-go-golems/glazed/releases/download/v$(GLAZED_VERSION)/glazed-spa.tar.gz \
		| tar xz -C $(GLAZED_SPA_DIR) \
	|| (echo "Warning: SPA assets not found for glazed v$(GLAZED_VERSION), building without browser UI"; rm -rf $(GLAZED_SPA_DIR))
```

The `curl -sfL` flags: `-s` silent, `-f` fail on HTTP errors (don't write error HTML to tar), `-L` follow redirects. The `|| (echo ...; rm -rf ...)` makes it non-fatal.

### 3.7 Complete file listing for pinocchio

```
New files:
  pkg/spa/
  ├── embed.go          ← //go:build embed, //go:embed dist
  ├── embed_none.go     ← //go:build !embed, fallback
  └── spa.go            ← NewHandler() → http.Handler

  cmd/pinocchio/cmds/
  ├── serve.go          ← NewServeCommand() cobra command
  └── help_loader.go    ← LoadAllHelpDocs() shared helper

Modified files:
  cmd/pinocchio/main.go          ← add rootCmd.AddCommand(NewServeCommand())
  Makefile                       ← add fetch-spa, build-with-spa targets
  .goreleaser.yaml               ← add before hook for fetch-spa, add -tags embed
  .gitignore                     ← add pkg/spa/dist/
```

---

## Part 4: End-to-End Flow

### For the glazed release

```
1. Push tag v1.2.8 to go-go-golems/glazed
2. GitHub Actions triggers release.yaml
3. goreleaser-linux / goreleaser-darwin jobs:
   a. go generate ./...          ← Dagger builds SPA → embed/public/
   b. tar czf glazed-spa.tar.gz  ← packages SPA
   c. go build -tags embed       ← bakes SPA into glaze binary
4. goreleaser-merge job:
   a. Merge artifacts
   b. Sign checksums
   c. Create GitHub Release with:
      - glaze binaries
      - glazed-spa.tar.gz         ← NEW
      - checksums, debs, rpms
```

### For the pinocchio release

```
1. Push tag v0.1.15 to go-go-golems/pinocchio
2. GitHub Actions triggers release.yml
3. goreleaser-linux / goreleaser-darwin jobs:
   a. make fetch-spa
      → reads glazed v1.2.8 from go.mod
      → curl https://github.com/.../glazed-spa.tar.gz
      → extract to pkg/spa/dist/
   b. go generate ./...
   c. go build -tags embed
      → pkg/spa/embed.go picks up dist/ via //go:embed
      → pinocchio binary now has the SPA baked in
4. goreleaser-merge job:
   a. Create GitHub Release with pinocchio binaries
```

### For the end user

```bash
# Option A: pinocchio serves its own help browser
pinocchio serve
# → opens http://localhost:8088 with pinocchio's docs in a browser

# Option B: glaze serves multiple tools' docs
glaze serve --from-glazed-cmd pinocchio,sqleton
# → all three tools' docs in one browser

# Option C: API only (no browser, for scripts)
pinocchio serve --address :9999
# (if built without -tags embed, serves API only)
```

---

## Part 5: Implementation Checklist

### Phase 1: Glazed repo changes

- [ ] Modify `.goreleaser.yaml`: keep `go generate ./...` in before hooks and add `release.extra_files` for `./glazed-spa.tar.gz`
- [ ] Modify `.github/workflows/release.yaml`: in `goreleaser-merge`, run `go generate ./pkg/web && tar czf glazed-spa.tar.gz -C pkg/web/embed/public .` before `goreleaser continue --merge`
- [ ] Tag and release: verify the tarball appears on the GitHub Release
- [ ] Verify the download URL works: `curl -L https://github.com/go-go-golems/glazed/releases/download/vX.Y.Z/glazed-spa.tar.gz`

### Phase 2: Pinocchio repo changes

- [ ] Create `pkg/spa/embed.go`, `pkg/spa/embed_none.go`, `pkg/spa/spa.go`
- [ ] Create `cmd/pinocchio/cmds/serve.go`
- [ ] Create `cmd/pinocchio/cmds/help_loader.go`
- [ ] Add `fetch-spa` target to `Makefile`
- [ ] Add `pkg/spa/dist/` to `.gitignore`
- [ ] Wire `NewServeCommand()` into `cmd/pinocchio/main.go`
- [ ] Test locally: `make fetch-spa && go build -tags embed -o pinocchio ./cmd/pinocchio && ./pinocchio serve`
- [ ] Update `.goreleaser.yaml`: add `make fetch-spa` to before hooks, add `-tags embed`

### Phase 3: Update other repos (sqleton, etc.)

Follow the same pattern as pinocchio:
- Create `pkg/spa/` with embed/embed_none/spa.go
- Add `make fetch-spa` to Makefile
- Add serve command
- Wire into main.go

---

## Part 6: Key File References

### Glazed repo (files to modify)

| File | Change |
|------|--------|
| `.goreleaser.yaml` | Add `release.extra_files` for `./glazed-spa.tar.gz`; keep `go generate ./...` for binary embedding |
| `.github/workflows/release.yaml` | Build/package `glazed-spa.tar.gz` in the merge job before `goreleaser continue --merge` |

### Pinocchio repo (new files)

| File | Purpose |
|------|---------|
| `pkg/spa/embed.go` | `//go:embed dist` with build tag `embed` |
| `pkg/spa/embed_none.go` | Fallback for builds without `embed` tag |
| `pkg/spa/spa.go` | `NewHandler()` — SPA serving logic (~40 lines) |
| `cmd/pinocchio/cmds/serve.go` | `NewServeCommand()` — Cobra command |
| `cmd/pinocchio/cmds/help_loader.go` | `LoadAllHelpDocs()` — shared doc loading |

### Pinocchio repo (files to modify)

| File | Change |
|------|--------|
| `cmd/pinocchio/main.go` | Add `rootCmd.AddCommand(pinocchio_cmds.NewServeCommand())` |
| `Makefile` | Add `fetch-spa`, `build-with-spa` targets |
| `.gitignore` | Add `pkg/spa/dist/` |
| `.goreleaser.yaml` | Add `make fetch-spa` to before hooks, add `-tags embed` |

### Glazed API surface used by pinocchio

| Type | Import | Purpose |
|------|--------|---------|
| `help.NewHelpSystem()` | `pkg/help` | Create in-memory SQLite help store |
| `helpserver.HandlerDeps` | `pkg/help/server` | Dependencies for HTTP handlers |
| `helpserver.NewServeHandler(deps, spa)` | `pkg/help/server` | Create API + SPA handler (auto-assigns default package) |

---

## Appendix: Dependency Diagram

```
                    ┌──────────────────┐
                    │  glazed repo     │
                    │  release v1.2.8  │
                    ├──────────────────┤
                    │ glaze binary     │
                    │ glazed-spa.tar.gz│──── curl download ────┐
                    └──────────────────┘                        │
                                                                ▼
┌──────────────────┐                             ┌──────────────────┐
│  pinocchio repo  │                             │  pinocchio build │
│                  │                             │                  │
│  go.mod:         │  go list -m ──► version ──► │  make fetch-spa  │
│  glazed v1.2.8   │                             │    ↓             │
│                  │                             │  pkg/spa/dist/   │
│  pkg/spa/        │                             │    ↓             │
│  embed.go ───────┼─ //go:embed dist ─────────► │  go build -tags  │
│                  │                             │    embed         │
└──────────────────┘                             │    ↓             │
                                                 │  pinocchio binary│
                                                 │  (with SPA)      │
                                                 └──────────────────┘
```
