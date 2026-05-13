---
Title: Implementation Diary
Ticket: GLZ-SPA-RELEASE
Status: active
Topics:
  - help
  - serve
  - http
  - spa
  - release
  - goreleaser
  - pinocchio
  - distribution
DocType: reference
Intent: long-term
Owners:
  - manuel
RelatedFiles:
  - Path: /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml
    Note: Release workflow Node/Corepack/pnpm cache ordering recorded in diary
Created: "2026-05-12"
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

### 2026-05-12 — P1 review fix: install pnpm before pnpm cache lookup

The first workflow fix installed pnpm, but it asked `actions/setup-node@v6` to restore the `pnpm` cache before Corepack activated pnpm. Codex correctly flagged this as still fragile because setup-node's pnpm cache lookup expects the package manager to be available before cache setup.

The workflow now uses a two-step setup in each release job: first install Node 22 without cache, then enable `pnpm@10.15.0` via Corepack, then invoke setup-node again with `cache: pnpm`. This preserves dependency caching while ensuring pnpm is installed before setup-node resolves the pnpm cache path.

#### Prompt Context

**User prompt (verbatim):** "P1 Badge Install pnpm before enabling setup-node caching

In each added release job, actions/setup-node is asked to restore the pnpm cache before the following corepack enable step makes pnpm available. The setup-node v6 docs state that the package manager for cache should be pre-installed, and its pnpm cache path lookup will fail on runners like the macOS runner that already lacked pnpm, so the workflow can still stop before reaching the fallback build. Move pnpm/corepack setup before the cached setup-node step, or split setup-node without cache from a later cache restore after pnpm is installed.

Useful? React with 👍 / 👎."

**Assistant interpretation:** Treat the review as valid, react with 👍, and update the release workflow so cache restoration happens only after pnpm exists.

**Inferred user intent:** Make PR #575 robust enough to fix the original macOS release failure instead of replacing it with an earlier setup-node cache failure.

**Commit (code):** 829ea9e — "Fix pnpm cache ordering in release workflow"

#### What I did
- Reacted 👍 to PR review comment `3230549754`.
- Updated `/home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml` in all release jobs.
- Replaced `setup-node(cache: pnpm) -> corepack enable` with `setup-node(no cache) -> corepack enable pnpm@10.15.0 -> setup-node(cache: pnpm)`.

#### Why
- `actions/setup-node` resolves the pnpm cache path during the action execution.
- On runners without preinstalled pnpm, requesting `cache: pnpm` before Corepack activation can fail before the workflow reaches `go generate`.

#### What worked
- The fix keeps the intended pnpm cache behavior and avoids needing a separate `actions/cache` shell step.
- The same pattern applies cleanly to linux, darwin, and merge jobs.

#### What didn't work
- The previous patch was incomplete because it assumed setup-node could configure pnpm caching before pnpm was available.
- My first validation command used a relative `ttmp/...` path from inside the repo, and docmgr resolved it as `.../glazed/ttmp/ttmp/...`, failing with `open ... no such file or directory`; rerunning with the absolute doc path worked.
- `docmgr doc relate` rewrote the diary frontmatter into empty schema keys (`Title: ""`, `Ticket: ""`, `DocType: ""`), so `docmgr validate frontmatter` failed with `missing required fields: Title, Ticket, DocType`; I restored the frontmatter using the docmgr-required uppercase keys and revalidated successfully.

#### What I learned
- For setup-node's package-manager cache modes, the package manager must already be installed or activated.
- Corepack activation requires Node first, so the ordering must be Node without cache, Corepack/pnpm, then cache-aware setup-node or equivalent cache restore.

#### What was tricky to build
- The ordering has a dependency cycle at first glance: Corepack needs Node, while pnpm caching needs pnpm. Splitting setup-node into a no-cache Node setup followed by Corepack activation and a second cache-aware setup-node call breaks that cycle without introducing a custom cache path script.

#### What warrants a second pair of eyes
- Confirm that running `actions/setup-node@v6` twice in one job is acceptable for this repository's release workflow and does not unexpectedly reset PATH or Node state.
- Confirm whether `pnpm@10.15.0` should remain pinned here or be derived from `web/packageManager` if that is later added to `package.json`.

#### What should be done in the future
- If setup-node double invocation proves noisy, replace the second setup-node call with `actions/cache` and `pnpm store path` after Corepack activation.

#### Code review instructions
- Start with `.github/workflows/release.yaml` and inspect the Node/Corepack/cache ordering in `goreleaser-linux`, `goreleaser-darwin`, and `goreleaser-merge`.
- Validate by checking PR #575 CI and, after merge, cutting a new release tag so the macOS split job reaches and completes `go generate ./...`.

#### Technical details

The intended per-job order is:

```yaml
- uses: actions/setup-node@v6
  with:
    node-version: '22'
- name: Enable pnpm
  run: corepack enable && corepack prepare pnpm@10.15.0 --activate
- uses: actions/setup-node@v6
  with:
    node-version: '22'
    cache: pnpm
    cache-dependency-path: web/pnpm-lock.yaml
```

### 2026-05-12 — Release attempt v1.2.11: setup-node pnpm cache post-job failure

A second release attempt progressed farther: the linux split job completed `go generate ./...`, built both linux targets, and uploaded the `dist-linux` artifact. The job then failed during setup-node post-job cache cleanup, not during the actual build.

The relevant log from run `25769721655`, job `75689991088`:

```text
goreleaser-linux Run goreleaser/goreleaser-action@v7 ... release succeeded after 2m14s
goreleaser-linux Run actions/upload-artifact@v7 ... Artifact dist-linux has been successfully uploaded!
goreleaser-linux Post Run actions/setup-node@v6 Post job cleanup.
Error: Path Validation Error: Path(s) specified in the action for caching do(es) not exist, hence no cache is being saved.
```

#### Prompt Context

**User prompt (verbatim):** "this time th elinux build has an issue in post run setup-node 
0s
Post job cleanup.
Error: Path Validation Error: Path(s) specified in the action for caching do(es) not exist, hence no cache is being saved.


https://github.com/go-go-golems/glazed/actions/runs/25769721655/job/75689991088"

**Assistant interpretation:** Diagnose and fix the new release workflow failure where setup-node's pnpm cache save fails after a successful linux build.

**Inferred user intent:** Remove the remaining CI release blocker so the next Glazed tag can publish the release and SPA asset.

**Commit (code):** a4f860d — "Remove pnpm cache from release workflow"

#### What I did
- Fetched the linux job log with `gh run view 25769721655 --repo go-go-golems/glazed --job 75689991088 --log`.
- Confirmed GoReleaser itself succeeded and uploaded artifacts before setup-node's post-job cache step failed.
- Removed `cache: pnpm` setup-node invocations from all release jobs and kept only Node 22 setup plus Corepack activation of `pnpm@10.15.0`.

#### Why
- The release workflow needs pnpm availability for macOS/merge fallback SPA generation, but dependency caching is not required for correctness.
- On linux, Dagger handled `go generate`, so local pnpm did not necessarily create a pnpm store path. setup-node then tried to save a cache for a path that did not exist and failed the otherwise-successful job.

#### What worked
- The log proved that the actual release build path was healthy up to artifact upload.
- Removing caching eliminates both ordering and empty-cache-path failure modes while preserving the pnpm fallback executable.

#### What didn't work
- The prior two-step setup-node cache pattern fixed pre-run pnpm availability but introduced a post-run failure when no pnpm store was created.

#### What I learned
- In this workflow, pnpm is a fallback dependency for release jobs, not a guaranteed dependency installer in every job.
- setup-node's built-in cache is too strict for a conditional pnpm usage path unless we also force creation/use of the store.

#### What was tricky to build
- The release jobs differ by platform behavior: linux can complete generation via Dagger without local pnpm usage, while macOS needs local pnpm after Dagger fails. A global setup-node pnpm cache therefore sees inconsistent filesystem state. Removing cache is safer than forcing fake cache directories because correctness depends only on pnpm availability, not cache persistence.

#### What warrants a second pair of eyes
- Confirm the release workflow's extra few seconds/minutes without pnpm caching are acceptable.
- Confirm macOS still reaches local pnpm fallback successfully after removing setup-node cache.

#### What should be done in the future
- If release runtime becomes a problem, add explicit `pnpm store path` + `mkdir -p` + `actions/cache` around jobs that actually run local pnpm, or change `cmd/build-web` to expose a deterministic cacheable path.

#### Code review instructions
- Review `.github/workflows/release.yaml`: each release job should now have exactly one setup-node step and one Corepack pnpm activation step, with no `cache: pnpm`.
- Validate by checking the next tagged release run; the linux job should no longer fail in setup-node post-job cleanup.

#### Technical details

The intended per-job order is now intentionally minimal:

```yaml
- uses: actions/setup-node@v6
  with:
    node-version: '22'
- name: Enable pnpm
  run: corepack enable && corepack prepare pnpm@10.15.0 --activate
```

### 2026-05-12 — Release attempt v1.2.12: GitHub Release published, Fury publisher tried SPA

The third release attempt got all the way through split builds, merge, SPA generation, GitHub release upload, and Homebrew formula publishing. The important success is that the GitHub Release exists and contains `glazed-spa-1.2.12.tar.gz`.

The remaining failure happened after GitHub publishing, in the custom Fury publisher. GoReleaser invoked the Fury `curl` publisher for `glazed-spa-1.2.12.tar.gz` even though the publisher is intended only for nfpm `.deb`/`.rpm` artifacts. The local file is `glazed-spa.tar.gz`; GoReleaser's release asset name is `glazed-spa-1.2.12.tar.gz`, so `curl -F package=@{{ .ArtifactName }}` could not open that local filename.

Relevant log from run `25770477640`, job `75692688757`:

```text
uploading to release name=glazed-spa-1.2.12.tar.gz
release published url=https://github.com/go-go-golems/glazed/releases/tag/v1.2.12
custom publisher
publishing cmd=curl artifact=glazed-spa-1.2.12.tar.gz
continue failed after 8s error=custom publisher: failed to publish artifacts: exit status 26 message=publishing failed cmd=curl
output=
│ curl: (26) Failed to open/read local data from file/application
```

#### Prompt Context

**User prompt (verbatim):** "• pushing                                      repository=go-go-golems/homebrew-go-go-go branch= file=glaze.rb
    • custom publisher
      • publishing                                   cmd=curl artifact=glaze_1.2.12_linux_amd64.deb
      • publishing                                   cmd=curl artifact=glaze_1.2.12_linux_arm64.deb
      • publishing                                   cmd=curl artifact=glaze_1.2.12_linux_arm64.rpm
      • publishing                                   cmd=curl artifact=glaze_1.2.12_linux_amd64.rpm
      • publishing                                   cmd=curl artifact=glazed-spa-1.2.12.tar.gz
  ⨯ continue failed after 8s                         error=custom publisher: failed to publish artifacts: exit status 26 message=publishing failed cmd=curl
    output=
    │ curl: (26) Failed to open/read local data from file/application
Error: The process '/opt/hostedtoolcache/goreleaser-action/2.15.4/x64/goreleaser' failed with exit code 1
 https://github.com/go-go-golems/glazed/actions/runs/25770477640/job/75692688757

getting further"

**Assistant interpretation:** Diagnose the new GoReleaser merge failure and patch the release configuration so Fury publishing does not fail on the SPA release asset.

**Inferred user intent:** Finish stabilizing the release pipeline while preserving the already-successful GitHub Release SPA asset publication.

**Commit (code):** 15ce479 — "Guard Fury publisher against SPA release asset"

#### What I did
- Inspected the merge job log with `gh run view 25770477640 --repo go-go-golems/glazed --job 75692688757 --log`.
- Verified `gh release view v1.2.12` succeeds and lists `glazed-spa-1.2.12.tar.gz` among release assets.
- Updated `.goreleaser.yaml` so the Fury custom publisher shell command only uploads `*.deb` and `*.rpm` artifacts, and explicitly skips non-package artifacts such as the SPA tarball.
- Ran `goreleaser check`; the configuration parses, and the only reported issues are pre-existing deprecated properties (`snapshot.name_template`, `brews`).

#### Why
- Fury is a package repository and should only receive nfpm package artifacts.
- The SPA tarball is only meant to be a GitHub Release asset for downstream Go projects to fetch/embed.
- Skipping non-package artifacts in the publisher command is robust even if GoReleaser still invokes the custom publisher for extra release files.

#### What worked
- v1.2.12 confirmed that the Node/pnpm, split job, merge job, and GitHub asset upload problems are solved.
- The release asset is present at `https://github.com/go-go-golems/glazed/releases/tag/v1.2.12`.

#### What didn't work
- GoReleaser still treated the SPA extra file as an artifact for the custom publisher path.
- The old publisher command used `{{ .ArtifactName }}` as a local filename. That breaks for assets whose published name differs from the local path, and it was also semantically wrong for the SPA because the SPA should not be pushed to Fury.

#### What I learned
- `release.extra_files.name_template` affects the GitHub Release asset name, but custom publishers may still see that artifact name.
- Publisher `ids` did not prevent the SPA extra file from reaching this custom publisher in the merge phase, so the command itself needs a defensive artifact-type guard.

#### What was tricky to build
- The failure happened after GitHub release publication, so the CI status is failed even though the primary SPA distribution objective succeeded. The fix must avoid breaking that successful path while stopping the later Fury step from treating the SPA as a package artifact.

#### What warrants a second pair of eyes
- Confirm the shell quoting in `.goreleaser.yaml` works under GoReleaser's publisher execution environment.
- Confirm Fury should never receive the SPA tarball; current assumption is yes, because it is not a deb/rpm package.

#### What should be done in the future
- Consider replacing the custom Fury publisher with a built-in package publisher configuration if GoReleaser provides a narrower artifact selector.
- Consider fixing unrelated GoReleaser deprecations (`snapshot.name_template`, `brews`) separately.

#### Code review instructions
- Review `.goreleaser.yaml` under `publishers:` and verify only `*.deb`/`*.rpm` invoke `curl`.
- Validate by cutting a follow-up tag such as `v1.2.13`; GitHub should publish the SPA asset and the custom publisher should skip it without failing.

#### Technical details

The publisher command now guards by artifact filename:

```yaml
cmd: >-
  sh -c 'case "$1" in
    *.deb|*.rpm) curl -F package=@"$1" "https://$FURY_TOKEN@push.fury.io/go-go-golems/" ;;
    *) echo "skipping non-package artifact $1 for fury.io" ;;
  esac' -- "{{ .ArtifactName }}"
```

### Summary

All implementation tasks are complete from the SPA distribution perspective: v1.2.12 published a GitHub Release containing `glazed-spa-1.2.12.tar.gz`. The release workflow still reported failure because the Fury custom publisher tried to upload the SPA tarball as if it were a package. The current fix guards the Fury command so only `.deb`/`.rpm` files are uploaded; a follow-up tag such as `v1.2.13` should validate the full release pipeline exits green.
