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

