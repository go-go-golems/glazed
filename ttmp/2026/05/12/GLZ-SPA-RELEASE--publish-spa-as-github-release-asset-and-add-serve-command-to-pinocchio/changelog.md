# Changelog

## 2026-05-12

- Initial workspace created


## 2026-05-12

Created implementation guide for publishing SPA as GitHub release asset and adding serve command to pinocchio. Two-phase plan: (1) glazed .goreleaser.yaml change to tar+attach SPA, (2) pinocchio pkg/spa/ + serve command + Makefile fetch-spa.


## 2026-05-12

All implementation tasks complete (4-9). Glazed GoReleaser configured for SPA tarball. Pinocchio has pkg/spa/, serve command, Makefile fetch-spa, GoReleaser embed tag. E2E test: 53 sections served, API works, #571 fix confirmed in pinocchio context.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/pinocchio/cmd/pinocchio/cmds/serve.go — New serve command for pinocchio help browser
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/pinocchio/pkg/spa/spa.go — SPA handler reusing glazed's pattern


## 2026-05-12

Continued after implementation: updated diary with actual commit hashes and pre-commit validation notes. Remaining work is Task 3 only: tag/release glazed, verify glazed-spa tarball, then bump pinocchio.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.goreleaser.yaml — Committed in d574dd4
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/pinocchio/cmd/pinocchio/cmds/serve.go — Committed in pinocchio 47da68e


## 2026-05-12

Fixed split/merge GoReleaser issue from review: removed SPA tar creation from .goreleaser before hooks and added Build SPA release asset step in .github/workflows/release.yaml goreleaser-merge job before continue --merge. Updated design guide and diary.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml — Builds glazed-spa.tar.gz in merge job before release.extra_files is evaluated
- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.goreleaser.yaml — release.extra_files remains


## 2026-05-12

Investigated failed glazed v1.2.10 release. Tag exists but no GitHub Release because goreleaser-darwin failed: Dagger unavailable on macOS and pnpm missing for local fallback. Added setup-node@v6 + corepack pnpm@10.15.0 to all release jobs before GoReleaser/go generate.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml — Installs pnpm in linux


## 2026-05-12

Addressed P1 review on PR #575: split setup-node so Node is installed first, pnpm is activated via Corepack, and pnpm cache restore runs only after pnpm exists.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml — Fixes release job pnpm cache ordering


## 2026-05-12

Diagnosed v1.2.11 release failure: linux GoReleaser completed and uploaded dist-linux, but setup-node failed in post-job pnpm cache save because the cache path did not exist. Removed release-job pnpm caching and kept Node 22 + Corepack pnpm activation only.

### Related Files

- /home/manuel/workspaces/2026-05-12/fix-serve-http-docs/glazed/.github/workflows/release.yaml — Removes fragile setup-node pnpm cache from release jobs

