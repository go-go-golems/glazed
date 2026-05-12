---
title: Implementation Diary
doc_type: reference
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
created: "2026-05-12"
---

# Implementation Diary

## 2026-05-12 — Implementation Session

### Context
Implementing GLZ-SPA-RELEASE: publish SPA as GitHub release asset, add serve command to pinocchio.
Design doc at `design-doc/01-implementation-guide-spa-release-asset-and-help-serve-in-pinocchio.md`.

### Pre-flight
- GLZ-571 fix already committed (005ff53) — `NewServeHandler` auto-assigns default package
- Clean working tree in both glazed and pinocchio repos
- Workspace: `/home/manuel/workspaces/2026-05-12/fix-serve-http-docs/`

### Task 2: Modify glazed .goreleaser.yaml ✅

**What changed:** `.goreleaser.yaml` in the glazed repo.

1. Added `sh -c "if [ -d pkg/web/embed/public ] && [ -f pkg/web/embed/public/index.html ]; then tar czf glazed-spa.tar.gz -C pkg/web/embed/public .; fi"` to `before: hooks` (after `go generate ./...`).
2. Added `release.extra_files` pointing at `glazed-spa.tar.gz` with `name_template: glazed-spa-{{ .Version }}.tar.gz`.

**Why the guard:** Local dev builds without the SPA shouldn't fail. The `if [ -d ... ]` check ensures the tar only runs when `go generate` actually produced the SPA.

**Committed:** `d574dd4` (`goreleaser: publish SPA as glazed-spa.tar.gz release asset`).

### Task 3: Tag and release glazed — SKIPPED

Manual CI step. Needs a real tag push to trigger the release pipeline. Will be done after PR review.

### Task 4: Create pinocchio pkg/spa/ package ✅

Created three files:

- `pkg/spa/embed.go` — `//go:build embed`, `//go:embed dist`, `var Assets embed.FS`
- `pkg/spa/embed_none.go` — `//go:build !embed`, walks up from CWD looking for `pkg/spa/dist/index.html`, falls back to a placeholder HTML page with a stderr note.
- `pkg/spa/spa.go` — `NewHandler()` returns `http.Handler` with SPA fallback routing (~40 lines, mirrors glazed's `pkg/web/static.go`).

`go vet` passes clean.

### Task 5: Add make fetch-spa to pinocchio Makefile and .gitignore ✅

**Makefile changes:**
- Added `fetch-spa` target: parses glazed version from `go.mod` via `grep + awk`, downloads `glazed-spa.tar.gz` from GitHub Release, extracts to `pkg/spa/dist/`.
- Added `clean-spa` target.
- Added `build-with-spa` target (depends on fetch-spa, builds with `-tags embed`).
- Added all new targets to `.PHONY`.

**Key learning:** `go list -m` doesn't work in workspace mode — returns `(devel)` or empty for workspace modules. Had to parse `go.mod` directly with `grep 'go-go-golems/glazed ' go.mod | awk`.

**gitignore:** Added `pkg/spa/dist/`.

**Test:** `make fetch-spa` correctly detects v1.2.7 from go.mod, tries to fetch from GitHub (fails because v1.2.7 hasn't been released with the SPA yet), falls back gracefully.

### Task 6: Create serve command and help_loader ✅

Created two files:

- `cmd/pinocchio/cmds/help_loader.go` — `LoadAllHelpDocs(hs)` loads geppetto docs, pinocchio/pkg/doc, catter docs, pinocchio/cmd/doc, and optional sessionstream docs. Extracted from `initRootCmd()` in main.go.
- `cmd/pinocchio/cmds/serve.go` — `NewServeCommand()` Cobra command with `--address` flag, `runServe()` creates HelpSystem, calls `LoadAllHelpDocs`, creates SPA handler (falls back to API-only), creates `NewServeHandler` (auto-assigns default package), starts HTTP server with graceful shutdown.

`go vet` and `go build` pass clean.

### Task 7: Wire serve command into main.go ✅

Added `rootCmd.AddCommand(pinocchio_cmds.NewServeCommand())` in `initRootCmd()`, right after the JS command. One line.

`pinocchio --help` shows the `serve` command in the listing.

### Task 8: Update pinocchio .goreleaser.yaml ✅

1. Added `make fetch-spa` to `before: hooks` (runs before `go generate`).
2. Added `tags: - embed` to both `pinocchio-linux` and `pinocchio-darwin` build configs.

### Task 9: Test end-to-end ✅

Built pinocchio without `-tags embed` (no SPA assets), started `pinocchio serve --address :18888`:

```
/api/health   → {"ok":true,"sections":53}
/api/packages → {"packages":[{"name":"default","sectionCount":53}],"defaultPackage":"default"}
/api/sections?package=default → Total: 53
/             → Serves SPA placeholder (expected, no assets fetched)
```

**53 sections loaded.** The API works correctly. The SPA serves a placeholder (expected — glazed v1.2.7 hasn't been released with the SPA tarball yet). Once glazed is released with the SPA artifact and pinocchio bumps to that version, `make fetch-spa` will download the real SPA and it'll work end-to-end.

### Commit and validation notes

**Glazed commits:**
- `d574dd4` — `goreleaser: publish SPA as glazed-spa.tar.gz release asset`
- `d223255` — `Update GLZ-SPA-RELEASE ticket: tasks 2,4-9 complete, diary written`

**Pinocchio commit:**
- `47da68e` — `Add pinocchio serve command with embedded help browser SPA`

**Pre-commit validation:**
- First pinocchio commit attempt failed on `gofmt` for `cmd/pinocchio/cmds/help_loader.go`.
- Fixed with `gofmt -w cmd/pinocchio/cmds/help_loader.go`.
- Second commit attempt passed lefthook: `go generate ./...`, `go build ./...`, `golangci-lint`, geppetto vet, and `go test ./...`.

### 2026-05-12 — Split-release review fix

A review pointed out a real issue in the first GoReleaser implementation: this repository uses split/merge releases. The linux/darwin jobs run `goreleaser release --clean --split`, upload only `dist/`, and the merge job runs `goreleaser continue --merge` from a fresh checkout plus downloaded artifacts. `before.hooks` run in the split jobs, but `release.extra_files` is evaluated in the merge/publish job.

That means a root-level `glazed-spa.tar.gz` created in `.goreleaser.yaml` `before.hooks` would be missing in the merge job where GoReleaser tries to publish it.

**Fix applied:**
- Removed the `tar czf glazed-spa.tar.gz ...` hook from `.goreleaser.yaml`.
- Kept `release.extra_files` in `.goreleaser.yaml`, pointing at `./glazed-spa.tar.gz`.
- Added a `Build SPA release asset` step to `.github/workflows/release.yaml` inside `goreleaser-merge`, immediately before `goreleaser continue --merge`:

```yaml
- name: Build SPA release asset
  run: |
    go generate ./pkg/web
    tar czf glazed-spa.tar.gz -C pkg/web/embed/public .
```

This recreates the platform-independent SPA tarball in the exact job where release publishing happens.

Validation: ran `goreleaser check`; configuration is valid, but GoReleaser exits nonzero because existing unrelated deprecated properties are present (`snapshot.name_template`, `brews`). No new schema error was introduced.

### 2026-05-12 — v1.2.10 release failure: pnpm missing on macOS

After merging PR #574 and tagging `v1.2.10`, the release workflow failed before publishing a GitHub Release. `gh release view v1.2.10` returned `release not found`, while `git ls-remote --tags` showed the tag exists.

The failing job was `goreleaser-darwin`. The relevant log:

```text
Dagger build failed ... driver for scheme "image" was not available
falling back to local pnpm
local build also failed: pnpm not found in PATH
pkg/web/gen.go:1: running "go": exit status 1
```

Root cause: `go generate ./...` runs `cmd/build-web`, which tries Dagger first and then local `pnpm`. On GitHub macOS runners, Dagger's image driver is unavailable and `pnpm` was not installed. The linux job succeeded, but the darwin split job failed, so the merge/publish job was skipped and no release asset was created.

**Fix applied:** updated `.github/workflows/release.yaml` to set up Node 22 and activate `pnpm@10.15.0` in all three release jobs (`goreleaser-linux`, `goreleaser-darwin`, and `goreleaser-merge`) before any `go generate` / GoReleaser invocation:

```yaml
- uses: actions/setup-node@v6
  with:
    node-version: '22'
    cache: pnpm
    cache-dependency-path: web/pnpm-lock.yaml
- name: Enable pnpm
  run: corepack enable && corepack prepare pnpm@10.15.0 --activate
```

This makes the local fallback path in `cmd/build-web` work when Dagger is unavailable, and also supports the merge job's `go generate ./pkg/web` step.

### Summary

All implementation tasks complete except the release verification portion of Task 3. The first `v1.2.10` tag attempt failed before publishing because macOS release workers lacked pnpm. After this workflow fix is merged, cut a new glazed tag (or rerun with an updated tag such as `v1.2.11`), verify `glazed-spa-<version>.tar.gz` appears on the GitHub Release, then bump pinocchio's `github.com/go-go-golems/glazed` dependency to that released version and rerun `make fetch-spa`.
