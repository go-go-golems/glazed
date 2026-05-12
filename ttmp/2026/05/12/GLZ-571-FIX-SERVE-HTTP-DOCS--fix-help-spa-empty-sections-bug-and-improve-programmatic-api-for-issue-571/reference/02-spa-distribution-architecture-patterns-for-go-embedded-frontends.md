---
title: SPA Distribution Architecture Patterns for Go Embedded Frontends
doc_type: reference
status: active
intent: long-term
topics:
  - help
  - serve
  - http
  - spa
  - api
  - bug
  - paper-cut
  - documentation
  - intern-guide
owners:
  - manuel
ticket: GLZ-571-FIX-SERVE-HTTP-DOCS
created: "2026-05-12"
---

# SPA Distribution Architecture Patterns for Go Embedded Frontends

## Goal

A decision-making reference for choosing how to distribute a built React/Vite SPA so that Go binaries can embed it via `//go:embed`. Covers five concrete patterns with pseudocode, CI pipeline sketches, tradeoff matrices, and real-world examples from open-source projects.

## Context

Glazed builds a React SPA (`web/`) via Dagger/pnpm, copies the output to `pkg/web/embed/public/`, and embeds it with `//go:embed`. This works for `glaze` itself because `go generate` + `go build -tags embed` run in the same repo. But any external Go binary that imports `glazed/pkg/web` cannot get the SPA assets — the `go:generate` pipeline doesn't cross module boundaries, the generated files are `.gitignore`d, and the Go module cache only contains committed source.

This is not unique to glazed. Every Go project that embeds a web frontend hits this problem. The patterns below are general-purpose.

---

## The Constraint

```
//go:embed only sees files that exist on disk relative to the .go file
       at the time go build runs.

go mod download only fetches files committed to git
       (respects .gitignore).

go generate only runs in your own module
       (not in dependencies in the module cache).
```

These three constraints define the design space. Every solution either:
- **commits files to git** (so `go mod download` gets them),
- **generates files at build time** (so `go generate` or a Makefile creates them), or
- **fetches files at runtime** (so the binary downloads them on first use).

---

## Pattern A: Commit Built Assets to a Separate Module

### Architecture

```
github.com/go-go-golems/glazed/           ← main repo, no blobs
github.com/go-go-golems/glazed-spa/       ← SPA dist only, blobs here
```

The SPA module is a tiny repo that contains nothing but the built frontend and an `embed.go`:

```
glazed-spa/
├── go.mod                module github.com/go-go-golems/glazed-spa
├── embed.go
├── dist/
│   ├── index.html
│   └── assets/
│       ├── index-abc123.js
│       └── index-def456.css
└── README.md             ← auto-generated: "Built assets for glazed help SPA"
```

### embed.go

```go
package spa

import "embed"

//go:embed dist
var Assets embed.FS
```

### go.mod

```
module github.com/go-go-golems/glazed-spa

go 1.22
```

That's it. No dependencies. No code. Just a container for `//go:embed`.

### CI Pipeline (GitHub Actions)

```yaml
# .github/workflows/release-spa.yml in the glazed repo
on:
  push:
    tags: ["v*"]

jobs:
  publish-spa:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build SPA
        run: GOWORK=off go generate ./pkg/web

      - name: Clone glazed-spa repo
        uses: actions/checkout@v4
        with:
          repository: go-go-golems/glazed-spa
          token: ${{ secrets.SPA_REPO_TOKEN }}
          path: glazed-spa-repo

      - name: Sync dist
        run: |
          rm -rf glazed-spa-repo/dist/*
          cp -r pkg/web/embed/public/ glazed-spa-repo/dist/
          cd glazed-spa-repo
          git config user.name "CI"
          git config user.email "ci@example.com"
          git add -A
          git commit -m "Update SPA for glazed ${{ github.ref_name }}"
          git tag "${{ github.ref_name }}"
          git push origin main --tags
```

### Consumer Code

```go
// cmd/glaze/main.go — in the glazed repo
import (
    "github.com/go-go-golems/glazed/pkg/help/server"
    "github.com/go-go-golems/glazed/pkg/web"
    spadist "github.com/go-go-golems/glazed-spa"
)

// Option 1: inject assets explicitly
spaHandler, err := web.NewSPAHandlerFromFS(spadist.Assets)

// Option 2: use the default (which checks for the spa module)
handler := server.NewServeHandler(deps, spaHandler)
```

```go
// external binary — API only, no SPA import needed
handler := server.NewServeHandler(deps, nil)
```

```go
// external binary — wants the SPA too
import spadist "github.com/go-go-golems/glazed-spa"

spaHandler, err := web.NewSPAHandlerFromFS(spadist.Assets)
handler := server.NewServeHandler(deps, spaHandler)
```

### Versioning Strategy

| Strategy | How | Tradeoff |
|----------|-----|----------|
| Lockstep tags | Both repos tagged `v0.38.0` simultaneously | Simple, but requires CI coordination |
| Same major, independent minor | `glazed v0.38.0` depends on `glazed-spa v0.38.x` | Allows SPA-only patches |
| SPA follows main | `glazed-spa` always tracks latest `glazed` | Simpler CI, but no pinning |

Recommendation: **Lockstep tags** with CI automation. The SPA is a build artifact of the main repo — there's no reason to version them independently.

### Size Impact

```
Typical Vite SPA build output:
  index.html     ~1 KB
  index-*.js     ~150-300 KB (minified, gzippable)
  index-*.css    ~30-80 KB

Per release: ~200-400 KB
After 50 releases: ~10-20 MB repo size
After 100 releases: ~20-40 MB repo size
```

Mitigation: Periodically squash history or use `git replace` to cap repo size. Or accept it — 40 MB is negligible for most teams.

### Real-World Examples

- **Hugo** — theme assets in separate repos (`hugo-github-markdown`, etc.)
- **Traefik** — web UI in a separate repo (`traefik-webui`)
- **Pulumi** — provider schemas in separate packages

---

## Pattern B: GitHub Release Artifact + go:generate Fetch

### Architecture

No extra repo. The SPA tarball is a release asset on the main repo.

```
github.com/go-go-golems/glazed/
├── releases/
│   └── v0.38.0/
│       └── glazed-spa.tar.gz    ← attached to GitHub Release
├── pkg/web/
│   ├── cmd/fetch-spa/main.go   ← go:generate helper
│   ├── embed.go                ← //go:embed embed/public
│   └── embed/public/           ← .gitignored, created by fetch
```

### fetch-spa tool

```go
// pkg/web/cmd/fetch-spa/main.go
package main

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func main() {
    version := flag.String("version", "", "glazed version to fetch")
    outDir := flag.String("out", "embed/public", "output directory")
    flag.Parse()

    if *version == "" {
        // Read from go.mod
        *version = detectVersion()
    }

    url := fmt.Sprintf(
        "https://github.com/go-go-golems/glazed/releases/download/%s/glazed-spa.tar.gz",
        *version,
    )

    resp, err := http.Get(url)
    if err != nil {
        log.Fatalf("fetch: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        log.Fatalf("fetch: HTTP %d", resp.StatusCode)
    }

    os.MkdirAll(*outDir, 0o755)
    extractTarGz(resp.Body, *outDir)
}
```

### gen.go

```go
// pkg/web/gen.go
//go:generate go run cmd/fetch-spa/main.go -version=$(shell cat VERSION) -out=embed/public
```

### CI Pipeline

```yaml
# In the glazed repo's release workflow
- name: Build SPA
  run: GOWORK=off go generate ./pkg/web

- name: Package SPA
  run: tar czf glazed-spa.tar.gz -C pkg/web/embed/public .

- name: Upload to Release
  uses: softprops/action-gh-release@v2
  with:
    files: glazed-spa.tar.gz
```

### Consumer Build

```bash
# Must run go generate explicitly (not automatic)
go generate github.com/go-go-golems/glazed/pkg/web@v0.38.0
go build -tags embed -o myapp ./cmd/myapp
```

### Why This Doesn't Work for External Consumers

`go generate` only processes source in the **current module**. When pinocchio runs `go generate ./...`, it does not traverse into glazed's module cache. The `//go:generate` directive in glazed's `gen.go` is invisible.

Workaround: The consumer would need to explicitly run:

```bash
go run github.com/go-go-golems/glazed/pkg/web/cmd/fetch-spa@v0.38.0 \
    -version=v0.38.0 \
    -out=$(go env GOMODCACHE)/github.com/go-go-golems/glazed@v0.38.0/pkg/web/embed/public
```

This is fragile and breaks with `go mod tidy` or version changes.

### Real-World Examples

- **Protocol Buffers** — `buf generate` fetches `.proto` from registries
- **sqlc** — fetches schema from migration directories
- **oapi-codegen** — generates from OpenAPI specs fetched at build time

---

## Pattern C: Runtime Fetch with Local Cache

### Architecture

No build-time embedding. The binary downloads the SPA on first use.

```go
// pkg/web/fetch.go
type CachedSPAHandler struct {
    cacheDir  string
    version   string
    upstream  string  // e.g. "https://releases.example.com/glazed-spa/"
    handler   http.Handler
    mu        sync.Once
}

func (h *CachedSPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.mu.Do(func() {
        if !h.hasCachedAssets() {
            h.fetchAssets()
        }
        h.handler = newSPAFromDir(h.cacheDir)
    })
    h.handler.ServeHTTP(w, r)
}
```

### Cache Layout

```
~/.cache/glazed-spa/
├── v0.38.0/
│   ├── index.html
│   └── assets/
└── v0.37.0/
    └── ...
```

### Startup Flow

```
Binary starts
  → check ~/.cache/glazed-spa/v0.38.0/
  → if exists: serve from cache
  → if not: fetch tarball, extract, serve
  → first request has ~200ms latency for extraction
```

### Pseudocode

```go
func NewCachedSPAHandler(version string) (http.Handler, error) {
    cacheDir := filepath.Join(os.UserCacheDir(), "glazed-spa", version)
    
    if _, err := os.Stat(filepath.Join(cacheDir, "index.html")); err == nil {
        // Cache hit — serve immediately
        return newSPAFromDir(cacheDir)
    }
    
    // Cache miss — download async, serve placeholder meanwhile
    return &LazyFetchHandler{
        cacheDir: cacheDir,
        version:  version,
        upstream: "https://github.com/go-go-golems/glazed/releases/download/" + version + "/glazed-spa.tar.gz",
    }, nil
}
```

### Pros

- No blobs anywhere in git
- No build-time requirements beyond standard `go build`
- Always serves the correct version for the binary
- Works for external consumers transparently

### Cons

- Requires network on first run
- Breaks air-gapped deployments
- `~/.cache` management is another concern
- More complex code (HTTP client, tar extraction, error handling, concurrent access)
- First-request latency

### Real-World Examples

- **Gravitational Teleport** — fetches web UI from CDN, caches in `~/.tsh/`
- **HashiCorp Vault** — downloads UI plugin at runtime
- **minikube** — downloads ISO images on first run

---

## Pattern D: npm Package as Distribution Channel

### Architecture

Publish the SPA as an npm package. Go fetches it at generate time.

```json
// @go-go-golems/glazed-spa package.json
{
  "name": "@go-go-golems/glazed-spa",
  "version": "0.38.0",
  "files": ["dist/"]
}
```

### Go Integration

```go
// pkg/web/gen.go
//go:generate sh -c "mkdir -p embed/public && cd embed/public && npm pack @go-go-golems/glazed-spa@$(cat ../../VERSION) && tar xzf *.tgz --strip-components=1 package/dist && rm *.tgz"
```

### CI Pipeline

```yaml
# In the glazed repo
- name: Build SPA
  run: GOWORK=off go generate ./pkg/web

- name: Publish to npm
  run: |
    cd pkg/web/embed/public
    npm init -y --scope go-go-golems
    npm publish --access public
  env:
    NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Why This Is Usually Wrong

- Requires Node.js at Go build time (defeats the point)
- npm registry is a new operational dependency
- npm authentication adds CI complexity
- The Go ecosystem doesn't naturally consume npm packages
- Version coordination between npm and Go modules is fragile

### When It Makes Sense

- Your team is already deep in the JS ecosystem
- You need the SPA to be independently consumable by other JS projects
- You have a monorepo with both Go and JS consumers

---

## Pattern E: Build the SPA in the Consumer's Build Pipeline

### Architecture

The SPA source lives in the glazed repo. The consumer's build pipeline clones glazed, builds the SPA, and then builds their binary.

```makefile
# Consumer's Makefile
GLAZED_VERSION := v0.38.0
GLAZED_DIR := .glazed-build

build: .glazed-build/pkg/web/embed/public
    go build -tags embed -o myapp ./cmd/myapp

.glazed-build/pkg/web/embed/public:
    git clone --depth 1 --branch $(GLAZED_VERSION) \
        https://github.com/go-go-golems/glazed.git $(GLAZED_DIR)
    cd $(GLAZED_DIR) && GOWORK=off go generate ./pkg/web
    cp -r $(GLAZED_DIR)/pkg/web/embed/public ./pkg/web/embed/public
```

### Why This Is Usually Wrong

- Requires Dagger/Node/pnpm in the consumer's build environment
- Requires git clone of the glazed repo (not just `go mod download`)
- Fragile — depends on glazed's internal directory structure
- Slow — full frontend build on every consumer build

### When It Makes Sense

- Monorepo where all Go binaries share the same build environment
- You already have a monorepo Dagger pipeline that builds everything

---

## Decision Matrix

| Pattern | Blobs in git? | External consumers work? | Network at build? | Network at runtime? | Complexity |
|---------|:---:|:---:|:---:|:---:|:---:|
| **A: Separate module** | Yes (separate repo) | Yes (import module) | No | No | Low |
| **B: Release artifact + fetch** | No | No (go:generate limitation) | Yes | No | Medium |
| **C: Runtime fetch + cache** | No | Yes | No | Yes (first run) | High |
| **D: npm package** | No (npm registry) | No (needs Node at build) | Yes | No | Medium |
| **F: Document API-only** | No | Yes (no SPA needed) | No | No | Minimal |

## Decision Framework

Answer these questions to pick a pattern:

### Q1: Do external consumers need the SPA at all?

- **No** → Use Pattern F (document API-only mode). Stop here.
- **Yes, sometimes** → Continue to Q2.
- **Yes, always** → Continue to Q2.

### Q2: Can you accept blobs in a git repo somewhere?

- **Yes, in a separate dedicated repo** → Pattern A (separate module). Clean, Go-idiomatic.
- **Yes, in the main repo** → Simplest option but you said you don't want this.
- **No, never** → Continue to Q3.

### Q3: Can you accept network access at runtime?

- **Yes** → Pattern C (runtime fetch). Transparent for consumers, but more code.
- **No (air-gapped deployments)** → Pattern B (release artifact fetch). Requires explicit `go generate` step in consumer's build. Fragile across module boundaries.

### Q4: Is your team already in the npm ecosystem?

- **Yes** → Pattern D (npm package) is viable but awkward.
- **No** → Avoid Pattern D.

---

## Recommendation for Glazed (Current State)

Based on the analysis:

1. **Immediate**: Pattern F — document API-only mode for external consumers. This is honest and requires zero code changes. The `glaze serve --from-glazed-cmd` pattern already works perfectly for multi-tool help browsing.

2. **If demand arises**: Pattern A — create `glazed-spa` module. This is the lowest-complexity option that actually works for external consumers. The separate repo is a pure artifact channel — you never read its git history, never review its diffs. It exists solely so `go mod download` can carry the files.

3. **Avoid**: Patterns B, D, and E — they fight the Go module system and add operational fragility.

---

## Appendix: How to Implement Pattern A (Separate Module) — Step by Step

### 1. Create the glazed-spa repo

```bash
mkdir glazed-spa && cd glazed-spa
git init
go mod init github.com/go-go-golems/glazed-spa
```

### 2. Write embed.go

```go
// embed.go
package spa

import "embed"

// Assets contains the built SPA frontend files.
// Import this package to make the assets available to web.NewSPAHandlerFromFS.
//
//go:embed dist
var Assets embed.FS
```

### 3. Add a README

```markdown
# glazed-spa

Built frontend assets for the Glazed Help Browser SPA.
This package is automatically updated by CI when a new glazed release is tagged.

## Usage

```go
import (
    "github.com/go-go-golems/glazed/pkg/web"
    spadist "github.com/go-go-golems/glazed-spa"
)

spaHandler, err := web.NewSPAHandlerFromFS(spadist.Assets)
```
```

### 4. Restructure glazed's pkg/web

```go
// pkg/web/static.go — add FromFS variant
func NewSPAHandler() (http.Handler, error) {
    return NewSPAHandlerFromFS(DefaultPublicFS)
}

func NewSPAHandlerFromFS(assets fs.FS) (http.Handler, error) {
    indexBytes, err := fs.ReadFile(assets, "index.html")
    if err != nil {
        return nil, fmt.Errorf("reading SPA index.html: %w", err)
    }

    fileServer := http.FileServer(http.FS(assets))
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... same SPA fallback logic as current, but using `assets` parameter
    }), nil
}
```

### 5. Wire in cmd/glaze/main.go

```go
import (
    spadist "github.com/go-go-golems/glazed-spa"
    "github.com/go-go-golems/glazed/pkg/web"
)

spaHandler, err := web.NewSPAHandlerFromFS(spadist.Assets)
// instead of: spaHandler, err := web.NewSPAHandler()
```

### 6. Add CI workflow

```yaml
# .github/workflows/publish-spa.yml
on:
  push:
    tags: ["v*"]

jobs:
  publish-spa:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }

      - name: Build SPA
        run: GOWORK=off go generate ./pkg/web

      - uses: actions/checkout@v4
        with:
          repository: go-go-golems/glazed-spa
          token: ${{ secrets.SPA_REPO_PAT }}
          path: spa-repo

      - name: Sync and push
        run: |
          rm -rf spa-repo/dist/*
          cp -r pkg/web/embed/public/* spa-repo/dist/
          cd spa-repo
          go mod tidy
          git add -A
          git diff --cached --quiet || git commit -m "SPA for glazed ${{ github.ref_name }}"
          git tag -f "${{ github.ref_name }}"
          git push origin main "${{ github.ref_name }}"
```

### 7. Update go.mod in glazed

```
require github.com/go-go-golems/glazed-spa v0.38.0
```

### 8. Update GoReleaser

```yaml
# .goreleaser.yaml — no changes needed
# before: hooks already runs `go generate ./...`
# The -tags embed is still needed, or remove it since the SPA
# now comes from the separate module, not from embed/public/
```

Actually, with Pattern A, the `embed` build tag becomes optional. The `glazed-spa` module handles embedding. The `pkg/web/embed.go` can be simplified or removed. The `!embed` fallback in `embed_none.go` still provides the placeholder for tests.
