# Tasks

## TODO

- [ ] **Audit current pattern**: Extract the repeated “bootstrap parse profile-settings” logic from `geppetto/pkg/layers/layers.go` into a minimal checklist of required steps.
- [ ] **Design helper API (Glazed)**: Propose an API that supports:
  - [ ] resolving `cli.ProfileSettings` from defaults + config + env + flags
  - [ ] reusing/plugging in a `ConfigFilesFunc` and optional config mapper
  - [ ] producing either: resolved settings, or a ready-to-append `middlewares.Middleware` for profiles
  - [ ] consistent parse metadata (`profiles` parse step should include `{profile, profileFile}`)
- [ ] **Choose package location**:
  - [ ] `glazed/pkg/cli` (Cobra-oriented helper) vs `glazed/pkg/cmds/middlewares` (middleware primitive) vs `glazed/pkg/config` (path + file list resolution)
- [ ] **Refactor sketch (Geppetto)**: Show how `geppetto/pkg/layers/layers.go` would shrink after adopting the helper.
- [ ] **Docs update plan**: Update `glazed/pkg/doc/topics/15-profiles.md` and other docs to recommend the helper and show a short example.

